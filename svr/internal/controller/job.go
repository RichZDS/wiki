package controller

import (
	"net/http"
	"strconv"

	"wiki/internal/model"
	"wiki/internal/service"
	"wiki/pkg/response"

	"github.com/gin-gonic/gin"
)

// jobManager 是任务管理器的全局实例，由 main 启动时通过 SetJobManager 注入。
var jobManager *model.JobManager

// SetJobManager 注入任务管理器实例。
func SetJobManager(mgr *model.JobManager) {
	jobManager = mgr
}

// ListJobs 处理任务列表请求。
func ListJobs(c *gin.Context) {
	svc := service.NewJobService(jobManager)
	response.Success(c, http.StatusOK, svc.List())
}

// GetJob 处理任务详情请求。
func GetJob(c *gin.Context) {
	svc := service.NewJobService(jobManager)
	item, err := svc.Get(c.Param("name"))
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}
	response.Success(c, http.StatusOK, item)
}

// StartJob 处理启用任务请求。
func StartJob(c *gin.Context) {
	svc := service.NewJobService(jobManager)
	item, err := svc.Enable(c.Param("name"))
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}
	response.Success(c, http.StatusOK, item)
}

// StopJob 处理停用任务请求。
func StopJob(c *gin.Context) {
	svc := service.NewJobService(jobManager)
	item, err := svc.Disable(c.Param("name"))
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}
	response.Success(c, http.StatusOK, item)
}

// RunJobNow 处理立即执行任务请求。
func RunJobNow(c *gin.Context) {
	svc := service.NewJobService(jobManager)
	item, err := svc.RunNow(c.Param("name"))
	if err != nil {
		response.Error(c, http.StatusConflict, err.Error())
		return
	}
	response.Success(c, http.StatusOK, item)
}

// ListJobLogs 处理任务日志列表请求。
func ListJobLogs(c *gin.Context) {
	svc := service.NewJobService(jobManager)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	result, err := svc.ListLogs(model.JobLogFilter{
		JobName: c.Param("name"),
		Level:   c.Query("level"),
		Page:    page,
		Size:    size,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, http.StatusOK, result)
}
