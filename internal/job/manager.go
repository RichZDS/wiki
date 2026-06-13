package job

import (
	"context"
	"errors"
	"fmt"
	"time"

	"wiki/internal/model"
	"wiki/pkg/logger"
)

// 时间轮
const ScanInterval = 100 * time.Millisecond

type Handler = model.JobHandler
type Manager = model.JobManager

// NewManager 创建并初始化作业调度器。
func NewManager() *Manager {
	state := &model.JobManagerState{Tasks: make(map[string]*model.JobTask)}
	return &model.JobManager{
		RegisterFunc: func(name string, interval time.Duration, handler model.JobHandler) error {
			return register(state, name, interval, handler)
		},
		StartFunc: func(ctx context.Context) {
			start(state, ctx)
		},
		WaitFunc: state.Wg.Wait,
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
		Name:     name,
		Interval: interval,
		Handler:  handler,
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

	ticker := time.NewTicker(ScanInterval)
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
		if registered.Running || (!registered.NextRun.IsZero() && now.Before(registered.NextRun)) {
			continue
		}

		registered.Running = true
		state.Wg.Add(1)
		go run(state, ctx, registered)
	}
}

// run 执行单个任务并更新下次运行时间。
func run(state *model.JobManagerState, ctx context.Context, registered *model.JobTask) {
	defer state.Wg.Done()
	defer func() {
		if recovered := recover(); recovered != nil {
			logger.GetLogger().Printf("[JOB] %s panic: %v", registered.Name, recovered)
		}

		state.Mu.Lock()
		registered.Running = false
		registered.NextRun = time.Now().Add(registered.Interval)
		state.Mu.Unlock()
	}()

	if err := registered.Handler(ctx); err != nil {
		logger.GetLogger().Printf("[JOB] %s failed: %v", registered.Name, err)
	}
}
