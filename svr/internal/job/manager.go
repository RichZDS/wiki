package job

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"wiki/internal/model"
	"wiki/internal/model/consts"
	"wiki/pkg/logger"
	"wiki/pkg/snowflake"
)

type Handler = model.JobHandler
type Manager = model.JobManager

const (
	jobStatusIdle      = "idle"
	jobStatusRunning   = "running"
	jobStatusSucceeded = "succeeded"
	jobStatusFailed    = "failed"
	jobStatusCanceled  = "canceled"
)

const (
	jobLogDebug = "debug"
	jobLogInfo  = "info"
	jobLogWarn  = "warn"
	jobLogError = "error"
)

// NewManager 创建并初始化作业调度器。
func NewManager(options ...model.JobManagerOptions) *Manager {
	opt := model.JobManagerOptions{LogDBLevel: jobLogInfo}
	if len(options) > 0 {
		opt = options[0]
	}
	opt.LogDBLevel = normalizeLogLevel(opt.LogDBLevel)

	state := &model.JobManagerState{
		Tasks:      make(map[string]*model.JobTask),
		DB:         opt.DB,
		LogDBLevel: opt.LogDBLevel,
	}
	return &model.JobManager{
		RegisterFunc: func(name string, interval time.Duration, handler model.JobHandler) error {
			return register(state, name, interval, handler)
		},
		StartFunc: func(ctx context.Context) {
			start(state, ctx)
		},
		WaitFunc: func() {
			state.Wg.Wait()
		},
		ListFunc: func() []model.JobSnapshot {
			return list(state)
		},
		GetFunc: func(name string) (*model.JobSnapshot, error) {
			return get(state, name)
		},
		EnableFunc: func(name string) (*model.JobSnapshot, error) {
			return enable(state, name)
		},
		DisableFunc: func(name string) (*model.JobSnapshot, error) {
			return disable(state, name)
		},
		RunNowFunc: func(name string) (*model.JobSnapshot, error) {
			return runNow(state, name)
		},
		ListLogsFunc: func(filter model.JobLogFilter) (*model.JobLogListResult, error) {
			return listLogs(state, filter)
		},
	}
}

// register 校验并注册周期任务。
func register(state *model.JobManagerState, name string, interval time.Duration, handler Handler) error {
	if name == "" {
		return errors.New("job name must not be empty")
	}
	if interval <= 0 {
		return errors.New("job interval must be greater than zero")
	}
	if handler == nil {
		return errors.New("job handler must not be nil")
	}

	state.Mu.Lock()
	defer state.Mu.Unlock()

	if _, exists := state.Tasks[name]; exists {
		return fmt.Errorf("job %q is already registered", name)
	}
	state.Tasks[name] = &model.JobTask{
		Name:       name,
		Interval:   interval,
		Handler:    handler,
		Enabled:    true,
		LastStatus: jobStatusIdle,
	}
	return nil
}

// start 启动调度循环，重复调用不会再次启动。
func start(state *model.JobManagerState, ctx context.Context) {
	state.Mu.Lock()
	if state.Started {
		state.Mu.Unlock()
		return
	}
	state.Started = true
	state.RootCtx = ctx
	state.Mu.Unlock()

	state.Wg.Add(1)
	go func() {
		defer state.Wg.Done()
		schedule(state, ctx)
	}()
}

// schedule 按固定扫描间隔调度到期任务。
func schedule(state *model.JobManagerState, ctx context.Context) {
	if ctx.Err() != nil {
		return
	}
	runDue(state, ctx, time.Now())

	ticker := time.NewTicker(consts.ScanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			runDue(state, ctx, now)
		}
	}
}

// runDue 启动当前时间已经到期且未运行的任务。
func runDue(state *model.JobManagerState, ctx context.Context, now time.Time) {
	if ctx.Err() != nil {
		return
	}

	state.Mu.Lock()
	defer state.Mu.Unlock()

	for _, registered := range state.Tasks {
		if !registered.Enabled || registered.Running || (!registered.NextRun.IsZero() && now.Before(registered.NextRun)) {
			continue
		}
		startTaskLocked(state, ctx, registered, "scheduled")
	}
}

// list 返回所有任务的快照。
func list(state *model.JobManagerState) []model.JobSnapshot {
	state.Mu.Lock()
	defer state.Mu.Unlock()

	items := make([]model.JobSnapshot, 0, len(state.Tasks))
	for _, task := range state.Tasks {
		items = append(items, snapshotLocked(task))
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})
	return items
}

// get 返回指定任务的快照。
func get(state *model.JobManagerState, name string) (*model.JobSnapshot, error) {
	state.Mu.Lock()
	defer state.Mu.Unlock()

	task, err := getTaskLocked(state, name)
	if err != nil {
		return nil, err
	}
	snapshot := snapshotLocked(task)
	return &snapshot, nil
}

// enable 启用指定任务，并将下一次执行时间推到当前。
func enable(state *model.JobManagerState, name string) (*model.JobSnapshot, error) {
	state.Mu.Lock()
	task, err := getTaskLocked(state, name)
	if err != nil {
		state.Mu.Unlock()
		return nil, err
	}
	task.Enabled = true
	task.NextRun = time.Now()
	snapshot := snapshotLocked(task)
	state.Mu.Unlock()

	persistJobLog(state, name, "", jobLogInfo, "job enabled")
	return &snapshot, nil
}

// disable 停用指定任务，并取消当前正在运行的执行。
func disable(state *model.JobManagerState, name string) (*model.JobSnapshot, error) {
	state.Mu.Lock()
	task, err := getTaskLocked(state, name)
	if err != nil {
		state.Mu.Unlock()
		return nil, err
	}
	task.Enabled = false
	task.NextRun = time.Time{}
	if task.Cancel != nil {
		task.Cancel()
	}
	if task.Running {
		task.LastStatus = jobStatusCanceled
	}
	snapshot := snapshotLocked(task)
	state.Mu.Unlock()

	persistJobLog(state, name, "", jobLogWarn, "job disabled")
	return &snapshot, nil
}

// runNow 立即启动指定任务执行一次。
func runNow(state *model.JobManagerState, name string) (*model.JobSnapshot, error) {
	state.Mu.Lock()
	task, err := getTaskLocked(state, name)
	if err != nil {
		state.Mu.Unlock()
		return nil, err
	}
	if task.Running {
		snapshot := snapshotLocked(task)
		state.Mu.Unlock()
		return &snapshot, fmt.Errorf("job %q is already running", name)
	}

	parentCtx := state.RootCtx
	if parentCtx == nil {
		parentCtx = context.Background()
	}
	startTaskLocked(state, parentCtx, task, "manual")
	snapshot := snapshotLocked(task)
	state.Mu.Unlock()
	return &snapshot, nil
}

// listLogs 查询任务持久化日志。
func listLogs(state *model.JobManagerState, filter model.JobLogFilter) (*model.JobLogListResult, error) {
	if state.DB == nil {
		return &model.JobLogListResult{List: []model.JobLog{}}, nil
	}
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Size < 1 || filter.Size > 100 {
		filter.Size = 20
	}
	if strings.TrimSpace(filter.Level) != "" {
		filter.Level = normalizeLogLevel(filter.Level)
	}

	total, err := model.CountJobLogs(state.DB, filter)
	if err != nil {
		return nil, err
	}
	logs, err := model.ListJobLogs(state.DB, filter)
	if err != nil {
		return nil, err
	}
	return &model.JobLogListResult{Total: total, List: logs}, nil
}

// startTaskLocked 在持有锁时标记任务运行状态并启动 goroutine。
func startTaskLocked(state *model.JobManagerState, parentCtx context.Context, task *model.JobTask, trigger string) {
	runCtx, cancel := context.WithCancel(parentCtx)
	now := time.Now()
	runID := fmt.Sprint(snowflake.Next())

	task.Running = true
	task.Cancel = cancel
	task.LastStartedAt = now
	task.LastFinishedAt = time.Time{}
	task.LastStatus = jobStatusRunning
	task.LastError = ""

	state.Wg.Add(1)
	go run(state, runCtx, task, runID, trigger)
}

// run 执行单个任务并更新下一次运行时间。
func run(state *model.JobManagerState, ctx context.Context, task *model.JobTask, runID, trigger string) {
	name := task.Name
	interval := task.Interval
	handler := task.Handler

	defer state.Wg.Done()
	defer func() {
		if recovered := recover(); recovered != nil {
			finishJob(state, name, runID, interval, jobStatusFailed, fmt.Sprintf("panic: %v", recovered))
		}
	}()

	persistJobLog(state, name, runID, jobLogInfo, fmt.Sprintf("job started by %s", trigger))
	if err := handler(ctx); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
			finishJob(state, name, runID, interval, jobStatusCanceled, err.Error())
			return
		}
		finishJob(state, name, runID, interval, jobStatusFailed, err.Error())
		return
	}
	finishJob(state, name, runID, interval, jobStatusSucceeded, "")
}

// finishJob 记录任务结束状态并写入持久化日志。
func finishJob(state *model.JobManagerState, name, runID string, interval time.Duration, status, errText string) {
	now := time.Now()
	level := jobLogInfo
	message := "job completed"
	if status == jobStatusFailed {
		level = jobLogError
		message = "job failed: " + errText
	} else if status == jobStatusCanceled {
		level = jobLogWarn
		message = "job canceled"
		if errText != "" {
			message += ": " + errText
		}
	}

	state.Mu.Lock()
	if task, ok := state.Tasks[name]; ok {
		if task.Cancel != nil {
			task.Cancel()
		}
		task.Running = false
		task.Cancel = nil
		task.LastFinishedAt = now
		task.LastStatus = status
		task.LastError = errText
		if task.Enabled {
			task.NextRun = now.Add(interval)
		} else {
			task.NextRun = time.Time{}
		}
	}
	state.Mu.Unlock()

	if status == jobStatusFailed {
		logger.GetLogger().Printf("[JOB] %s failed: %s", name, errText)
	}
	persistJobLog(state, name, runID, level, message)
}

// getTaskLocked 在持有锁时查找任务。
func getTaskLocked(state *model.JobManagerState, name string) (*model.JobTask, error) {
	task, ok := state.Tasks[name]
	if !ok {
		return nil, fmt.Errorf("job %q not found", name)
	}
	return task, nil
}

// snapshotLocked 在持有锁时生成任务快照。
func snapshotLocked(task *model.JobTask) model.JobSnapshot {
	return model.JobSnapshot{
		Name:           task.Name,
		IntervalMS:     task.Interval.Milliseconds(),
		Enabled:        task.Enabled,
		Running:        task.Running,
		NextRun:        timePtr(task.NextRun),
		LastStartedAt:  timePtr(task.LastStartedAt),
		LastFinishedAt: timePtr(task.LastFinishedAt),
		LastStatus:     task.LastStatus,
		LastError:      task.LastError,
	}
}

// timePtr 将零值时间转换为空指针，便于 JSON 省略。
func timePtr(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	return &value
}

// persistJobLog 按配置的日志级别阈值写入数据库。
func persistJobLog(state *model.JobManagerState, jobName, runID, level, message string) {
	level = normalizeLogLevel(level)
	if !shouldPersistLevel(level, state.LogDBLevel) || state.DB == nil {
		return
	}
	item := &model.JobLog{
		ID:      snowflake.Next(),
		JobName: jobName,
		RunID:   runID,
		Level:   level,
		Message: message,
	}
	if err := model.InsertJobLog(state.DB, item); err != nil {
		logger.GetLogger().Printf("[JOB] persist log failed: %v", err)
	}
}

// shouldPersistLevel 判断当前日志级别是否达到入库阈值。
func shouldPersistLevel(level, minLevel string) bool {
	return logLevelRank(level) >= logLevelRank(minLevel)
}

// normalizeLogLevel 归一化日志级别名称。
func normalizeLogLevel(level string) string {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case jobLogDebug:
		return jobLogDebug
	case jobLogWarn, "warning":
		return jobLogWarn
	case jobLogError:
		return jobLogError
	default:
		return jobLogInfo
	}
}

// logLevelRank 返回日志级别的排序权重。
func logLevelRank(level string) int {
	switch normalizeLogLevel(level) {
	case jobLogDebug:
		return 10
	case jobLogInfo:
		return 20
	case jobLogWarn:
		return 30
	case jobLogError:
		return 40
	default:
		return 20
	}
}
