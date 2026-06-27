package controller

import (
	"net/http"

	"wiki/internal/model"
	"wiki/internal/service"
	"wiki/pkg/response"

	"github.com/gin-gonic/gin"
)

// Chunk 处理单策略切块请求。
// 从请求体解析 ChunkRequest 并调用 service 层执行切块。
func Chunk(c *gin.Context) {
	var req model.ChunkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	svc := service.NewChunkService()
	result, err := svc.Chunk(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "chunk failed: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, result)
}

// CompareChunk 处理多策略对比切块请求。
// 从请求体解析 ChunkCompareRequest 并调用 service 层执行对比。
func CompareChunk(c *gin.Context) {
	var req model.ChunkCompareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	svc := service.NewChunkService()
	result, err := svc.Compare(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "compare failed: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, result)
}
