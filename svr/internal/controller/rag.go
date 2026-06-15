package controller

import (
	"net/http"

	"wiki/internal/model"
	"wiki/pkg/response"

	"github.com/gin-gonic/gin"
)

// NewRAGController 创建 RAG 控制器并注入工作流服务。
func NewRAGController(svc *model.RAGService) *model.RAGController {
	return &model.RAGController{
		IngestFunc: func(c *gin.Context) { ingest(c, svc) },
		SearchFunc: func(c *gin.Context) { search(c, svc) },
	}
}

// ingest 处理文档入库请求：文本 → 切块 → 向量化 → Redis 存储。
func ingest(c *gin.Context, svc *model.RAGService) {
	var req model.RAGIngestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	result, err := svc.Ingest(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "ingest failed: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, result)
}

// search 处理语义检索请求：查询 → 向量化 → Redis 向量搜索。
func search(c *gin.Context, svc *model.RAGService) {
	var req model.RAGSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	result, err := svc.Search(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "search failed: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, result)
}
