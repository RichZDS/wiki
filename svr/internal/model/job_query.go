package model

import (
	"fmt"

	"gorm.io/gorm"
)

// InsertJobLog 向数据库写入一条任务日志。
func InsertJobLog(db *gorm.DB, item *JobLog) error {
	if db == nil {
		return nil
	}
	if err := db.Create(item).Error; err != nil {
		return fmt.Errorf("insert job log: %w", err)
	}
	return nil
}

// CountJobLogs 根据筛选条件统计任务日志数量。
func CountJobLogs(db *gorm.DB, filter JobLogFilter) (int64, error) {
	var total int64
	query := applyJobLogFilters(db.Model(&JobLog{}), filter)
	if err := query.Count(&total).Error; err != nil {
		return 0, fmt.Errorf("count job logs: %w", err)
	}
	return total, nil
}

// ListJobLogs 根据筛选条件分页查询任务日志。
func ListJobLogs(db *gorm.DB, filter JobLogFilter) ([]JobLog, error) {
	query := applyJobLogFilters(db.Model(&JobLog{}), filter)
	query = query.Order("created_at desc").Order("id desc")

	var logs []JobLog
	if err := query.Offset((filter.Page - 1) * filter.Size).Limit(filter.Size).Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("list job logs: %w", err)
	}
	return logs, nil
}

// applyJobLogFilters 将任务日志筛选条件应用到 GORM 查询。
func applyJobLogFilters(db *gorm.DB, filter JobLogFilter) *gorm.DB {
	if filter.JobName != "" {
		db = db.Where("job_name = ?", filter.JobName)
	}
	if filter.Level != "" {
		db = db.Where("level = ?", filter.Level)
	}
	return db
}
