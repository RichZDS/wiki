package model

import (
	"context"

	"github.com/gin-gonic/gin"
)

// RAGIngestRequest 文档入库请求体。
// Content 为原始文档文本，Strategy 指定切块策略。
type RAGIngestRequest struct {
	Content      string   `json:"content" binding:"required"`
	Strategy     string   `json:"strategy"`
	ChunkSize    int      `json:"chunk_size"`
	ChunkOverlap int      `json:"chunk_overlap"`
	Separators   []string `json:"separators"`
	DocIDPrefix  string   `json:"doc_id_prefix"`
}

// RAGIngestResult 文档入库响应体。
type RAGIngestResult struct {
	StoredIDs  []string `json:"stored_ids"`
	ChunkCount int      `json:"chunk_count"`
	TotalChars int      `json:"total_chars"`
}

// RAGSearchRequest 语义检索请求体。
type RAGSearchRequest struct {
	Query string `json:"query" binding:"required"`
	TopK  int    `json:"top_k"`
}

// RAGSearchItem 单条检索命中的结果。
type RAGSearchItem struct {
	ID       string         `json:"id"`
	Content  string         `json:"content"`
	Score    *float64       `json:"score,omitempty"`
	MetaData map[string]any `json:"metadata"`
}

// RAGSearchResult 语义检索响应体。
type RAGSearchResult struct {
	Results []RAGSearchItem `json:"results"`
}

// RAGController RAG 控制器，函数字段由 controller 层注入。
type RAGController struct {
	IngestFunc func(*gin.Context)
	SearchFunc func(*gin.Context)
}

// Ingest 执行文档入库（切块 + 向量化 + 存储）并返回 JSON 结果。
func (c *RAGController) Ingest(ctx *gin.Context) { c.IngestFunc(ctx) }

// Search 执行语义检索并返回 JSON 结果。
func (c *RAGController) Search(ctx *gin.Context) { c.SearchFunc(ctx) }

// RAGService RAG 工作流服务，函数字段由 rag 包注入。
type RAGService struct {
	IngestFunc func(context.Context, RAGIngestRequest) (*RAGIngestResult, error)
	SearchFunc func(context.Context, RAGSearchRequest) (*RAGSearchResult, error)
}

// Ingest 执行文档入库流程：文本 → 切块 → embedding → Redis 存储。
func (s *RAGService) Ingest(ctx context.Context, req RAGIngestRequest) (*RAGIngestResult, error) {
	return s.IngestFunc(ctx, req)
}

// Search 执行语义检索流程：查询 → embedding → Redis 向量搜索。
func (s *RAGService) Search(ctx context.Context, req RAGSearchRequest) (*RAGSearchResult, error) {
	return s.SearchFunc(ctx, req)
}
