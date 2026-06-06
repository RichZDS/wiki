package model

import (
	"context"
	"sync"
	"time"
)

type JobHandler func(context.Context) error

type JobTask struct {
	Name     string
	Interval time.Duration
	Handler  JobHandler
	NextRun  time.Time
	Running  bool
}

type JobManagerState struct {
	Mu      sync.Mutex
	Tasks   map[string]*JobTask
	Started bool
	Wg      sync.WaitGroup
}

type JobManager struct {
	RegisterFunc func(string, time.Duration, JobHandler) error
	StartFunc    func(context.Context)
	WaitFunc     func()
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
