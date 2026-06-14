package job

import (
	"time"

	"wiki/internal/model"
	"wiki/internal/model/consts"

	"gorm.io/gorm"
)

// JobDef 定义了一个可注册到调度器的周期任务。
type JobDef struct {
	Name     string
	Interval time.Duration
	Handler  model.JobHandler
}

// JobGroup 包含一组默认周期任务，新增任务只需在此添加即可自动注册。
type JobGroup struct {
	jobs []JobDef
}

// NewDefaultJobGroup 创建包含所有默认周期任务的 JobGroup。
// 新增周期任务时，只需在此构造函数中添加对应的 JobDef 条目。
func NewDefaultJobGroup(db *gorm.DB) *JobGroup {
	return &JobGroup{
		jobs: []JobDef{
			{
				Name:     "model-health-check",
				Interval: consts.ModelHealthInterval,
				Handler:  NewModelHealthTask(db, DefaultModelCheckers()).Run,
			},
			// 添加新的任务到这里
		},
	}
}

// RegisterAll 将 JobGroup 中的所有任务注册到指定的 JobManager。
func (g *JobGroup) RegisterAll(m *model.JobManager) error {
	for _, job := range g.jobs {
		if err := m.Register(job.Name, job.Interval, job.Handler); err != nil {
			return err
		}
	}
	return nil
}
