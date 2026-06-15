package controller

import (
	"net/http"

	"wiki/internal/model"
	"wiki/internal/service"
	"wiki/pkg/response"

	"github.com/gin-gonic/gin"
)

// 类型别名，与 user controller 模式保持一致。
type ChunkController = model.ChunkController

// NewChunkController 创建切块控制器并注入切块服务。
func NewChunkController() *ChunkController {
	svc := service.NewChunkService()
	return &ChunkController{
		ChunkFunc:   func(ctx *gin.Context) { chunk(ctx, svc) },
		CompareFunc: func(ctx *gin.Context) { compare(ctx, svc) },
	}
}

// chunk 处理单策略切块请求。
// 从请求体解析 ChunkRequest 并调用 service 层执行切块。
func chunk(c *gin.Context, svc *service.ChunkService) {
	var req model.ChunkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	result, err := svc.Chunk(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "chunk failed: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, result)
}

// compare 处理多策略对比切块请求。
// 从请求体解析 ChunkCompareRequest 并调用 service 层执行对比。
func compare(c *gin.Context, svc *service.ChunkService) {
	var req model.ChunkCompareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	result, err := svc.Compare(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "compare failed: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, result)
}
