package model

import (
	"context"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type JobHandler func(context.Context) error

type JobTask struct {
	Name           string
	Interval       time.Duration
	Handler        JobHandler
	NextRun        time.Time
	Running        bool
	Enabled        bool
	Cancel         context.CancelFunc
	LastStartedAt  time.Time
	LastFinishedAt time.Time
	LastStatus     string
	LastError      string
}

type JobManagerState struct {
	Mu         sync.Mutex
	Tasks      map[string]*JobTask
	Started    bool
	RootCtx    context.Context
	DB         *gorm.DB
	LogDBLevel string
	Wg         sync.WaitGroup
}

type JobManager struct {
	RegisterFunc func(string, time.Duration, JobHandler) error
	StartFunc    func(context.Context)
	WaitFunc     func()
	ListFunc     func() []JobSnapshot
	GetFunc      func(string) (*JobSnapshot, error)
	EnableFunc   func(string) (*JobSnapshot, error)
	DisableFunc  func(string) (*JobSnapshot, error)
	RunNowFunc   func(string) (*JobSnapshot, error)
	ListLogsFunc func(JobLogFilter) (*JobLogListResult, error)
}

type JobManagerOptions struct {
	DB         *gorm.DB
	LogDBLevel string
}

type JobSnapshot struct {
	Name           string     `json:"name"`
	IntervalMS     int64      `json:"interval_ms"`
	Enabled        bool       `json:"enabled"`
	Running        bool       `json:"running"`
	NextRun        *time.Time `json:"next_run,omitempty"`
	LastStartedAt  *time.Time `json:"last_started_at,omitempty"`
	LastFinishedAt *time.Time `json:"last_finished_at,omitempty"`
	LastStatus     string     `json:"last_status"`
	LastError      string     `json:"last_error"`
}

type JobLog struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	JobName   string    `gorm:"column:job_name;type:varchar(128);index;not null" json:"job_name"`
	RunID     string    `gorm:"column:run_id;type:varchar(64);index;not null" json:"run_id"`
	Level     string    `gorm:"column:level;type:varchar(16);index;not null" json:"level"`
	Message   string    `gorm:"column:message;type:varchar(1024);not null" json:"message"`
	CreatedAt time.Time `gorm:"column:created_at;not null;autoCreateTime" json:"created_at"`
}

// TableName 返回当前模型对应的数据库表名。
func (JobLog) TableName() string {
	return "job_log"
}

type JobLogFilter struct {
	JobName string
	Level   string
	Page    int
	Size    int
}

type JobLogListResult struct {
	Total int64    `json:"total"`
	List  []JobLog `json:"list"`
}

type JobService struct {
	ListFunc     func() []JobSnapshot
	GetFunc      func(string) (*JobSnapshot, error)
	EnableFunc   func(string) (*JobSnapshot, error)
	DisableFunc  func(string) (*JobSnapshot, error)
	RunNowFunc   func(string) (*JobSnapshot, error)
	ListLogsFunc func(JobLogFilter) (*JobLogListResult, error)
}

type JobController struct {
	ListFunc   func(*gin.Context)
	GetFunc    func(*gin.Context)
	StartFunc  func(*gin.Context)
	StopFunc   func(*gin.Context)
	RunNowFunc func(*gin.Context)
	LogsFunc   func(*gin.Context)
}

// Register 注册周期任务。
func (m *JobManager) Register(name string, interval time.Duration, handler JobHandler) error {
	return m.RegisterFunc(name, interval, handler)
}

// Start 启动作业调度器。
func (m *JobManager) Start(ctx context.Context) {
	m.StartFunc(ctx)
}

// Wait 等待调度器及其任务结束。
func (m *JobManager) Wait() {
	m.WaitFunc()
}

// List 返回所有注册任务的运行快照。
func (m *JobManager) List() []JobSnapshot {
	return m.ListFunc()
}

// Get 返回指定任务的运行快照。
func (m *JobManager) Get(name string) (*JobSnapshot, error) {
	return m.GetFunc(name)
}

// Enable 启用指定任务的周期调度。
func (m *JobManager) Enable(name string) (*JobSnapshot, error) {
	return m.EnableFunc(name)
}

// Disable 停用指定任务并取消正在运行的执行。
func (m *JobManager) Disable(name string) (*JobSnapshot, error) {
	return m.DisableFunc(name)
}

// RunNow 立即触发指定任务执行一次。
func (m *JobManager) RunNow(name string) (*JobSnapshot, error) {
	return m.RunNowFunc(name)
}

// ListLogs 查询任务持久化日志。
func (m *JobManager) ListLogs(filter JobLogFilter) (*JobLogListResult, error) {
	return m.ListLogsFunc(filter)
}

// List 查询所有任务快照。
func (s *JobService) List() []JobSnapshot {
	return s.ListFunc()
}

// Get 查询指定任务快照。
func (s *JobService) Get(name string) (*JobSnapshot, error) {
	return s.GetFunc(name)
}

// Enable 启用指定任务。
func (s *JobService) Enable(name string) (*JobSnapshot, error) {
	return s.EnableFunc(name)
}

// Disable 停用指定任务。
func (s *JobService) Disable(name string) (*JobSnapshot, error) {
	return s.DisableFunc(name)
}

// RunNow 立即执行指定任务。
func (s *JobService) RunNow(name string) (*JobSnapshot, error) {
	return s.RunNowFunc(name)
}

// ListLogs 查询任务日志。
func (s *JobService) ListLogs(filter JobLogFilter) (*JobLogListResult, error) {
	return s.ListLogsFunc(filter)
}

// List 处理任务列表请求。
func (c *JobController) List(ctx *gin.Context) {
	c.ListFunc(ctx)
}

// Get 处理任务详情请求。
func (c *JobController) Get(ctx *gin.Context) {
	c.GetFunc(ctx)
}

// Start 处理启用任务请求。
func (c *JobController) Start(ctx *gin.Context) {
	c.StartFunc(ctx)
}

// Stop 处理停用任务请求。
func (c *JobController) Stop(ctx *gin.Context) {
	c.StopFunc(ctx)
}

// RunNow 处理立即执行任务请求。
func (c *JobController) RunNow(ctx *gin.Context) {
	c.RunNowFunc(ctx)
}

// Logs 处理任务日志查询请求。
func (c *JobController) Logs(ctx *gin.Context) {
	c.LogsFunc(ctx)
}

type ModelHealthTask struct {
	RunFunc func(context.Context) error
}

// Run 执行模型健康检查任务。
func (t *ModelHealthTask) Run(ctx context.Context) error {
	return t.RunFunc(ctx)
}

type CompatibleModelChecker struct {
	CheckFunc func(context.Context) error
}

// Check 调用兼容 OpenAI 协议的模型执行健康探测。
func (c *CompatibleModelChecker) Check(ctx context.Context) error {
	return c.CheckFunc(ctx)
}
