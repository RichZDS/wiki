package controller

import (
	"net/http"

	"wiki/internal/model"
	"wiki/pkg/response"

	"github.com/gin-gonic/gin"
)

// ragService 是 RAG 工作流服务的全局实例，由 main 启动时通过 SetRAGService 注入。
var ragService *model.RAGService

// SetRAGService 注入 RAG 工作流服务实例。
func SetRAGService(svc *model.RAGService) {
	ragService = svc
}

// IngestRAG 处理文档入库请求：文本 → 切块 → 向量化 → Redis 存储。
func IngestRAG(c *gin.Context) {
	var req model.RAGIngestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	result, err := ragService.Ingest(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "ingest failed: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, result)
}

// SearchRAG 处理语义检索请求：查询 → 向量化 → Redis 向量搜索。
func SearchRAG(c *gin.Context) {
	var req model.RAGSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	result, err := ragService.Search(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "search failed: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, result)
}
