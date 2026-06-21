package controller

import (
	"net/http"
	"strconv"

	"wiki/internal/model"
	"wiki/internal/service"
	"wiki/pkg/response"

	"github.com/gin-gonic/gin"
)

type JobController = model.JobController

// NewJobController 创建任务控制器并注入任务服务。
func NewJobController(manager *model.JobManager) *JobController {
	svc := service.NewJobService(manager)
	return &model.JobController{
		ListFunc:   func(ctx *gin.Context) { listJobs(ctx, svc) },
		GetFunc:    func(ctx *gin.Context) { getJob(ctx, svc) },
		StartFunc:  func(ctx *gin.Context) { startJob(ctx, svc) },
		StopFunc:   func(ctx *gin.Context) { stopJob(ctx, svc) },
		RunNowFunc: func(ctx *gin.Context) { runJobNow(ctx, svc) },
		LogsFunc:   func(ctx *gin.Context) { listJobLogs(ctx, svc) },
	}
}

// listJobs 处理任务列表请求。
func listJobs(c *gin.Context, svc *service.JobService) {
	response.Success(c, http.StatusOK, svc.List())
}

// getJob 处理任务详情请求。
func getJob(c *gin.Context, svc *service.JobService) {
	item, err := svc.Get(c.Param("name"))
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}
	response.Success(c, http.StatusOK, item)
}

// startJob 处理启用任务请求。
func startJob(c *gin.Context, svc *service.JobService) {
	item, err := svc.Enable(c.Param("name"))
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}
	response.Success(c, http.StatusOK, item)
}

// stopJob 处理停用任务请求。
func stopJob(c *gin.Context, svc *service.JobService) {
	item, err := svc.Disable(c.Param("name"))
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}
	response.Success(c, http.StatusOK, item)
}

// runJobNow 处理立即执行任务请求。
func runJobNow(c *gin.Context, svc *service.JobService) {
	item, err := svc.RunNow(c.Param("name"))
	if err != nil {
		response.Error(c, http.StatusConflict, err.Error())
		return
	}
	response.Success(c, http.StatusOK, item)
}

// listJobLogs 处理任务日志列表请求。
func listJobLogs(c *gin.Context, svc *service.JobService) {
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
