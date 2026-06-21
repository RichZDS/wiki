package service

import "wiki/internal/model"

type JobService = model.JobService

// NewJobService 创建任务服务并注入调度器。
func NewJobService(manager *model.JobManager) *JobService {
	return &model.JobService{
		ListFunc: func() []model.JobSnapshot {
			return manager.List()
		},
		GetFunc: func(name string) (*model.JobSnapshot, error) {
			return manager.Get(name)
		},
		EnableFunc: func(name string) (*model.JobSnapshot, error) {
			return manager.Enable(name)
		},
		DisableFunc: func(name string) (*model.JobSnapshot, error) {
			return manager.Disable(name)
		},
		RunNowFunc: func(name string) (*model.JobSnapshot, error) {
			return manager.RunNow(name)
		},
		ListLogsFunc: func(filter model.JobLogFilter) (*model.JobLogListResult, error) {
			return manager.ListLogs(filter)
		},
	}
}
