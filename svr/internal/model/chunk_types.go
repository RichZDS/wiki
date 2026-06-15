package model

import (
	"context"

	"github.com/gin-gonic/gin"
)

// ChunkRequest 单策略切块请求体。
type ChunkRequest struct {
	Content      string   `json:"content" binding:"required"`
	Strategy     string   `json:"strategy" binding:"required"`
	ChunkSize    int      `json:"chunk_size"`
	ChunkOverlap int      `json:"chunk_overlap"`
	Separators   []string `json:"separators"`
}

// ChunkConfigItem 多策略对比中单个策略的配置。
type ChunkConfigItem struct {
	Strategy     string   `json:"strategy"`
	ChunkSize    int      `json:"chunk_size"`
	ChunkOverlap int      `json:"chunk_overlap"`
	Separators   []string `json:"separators"`
}

// ChunkCompareRequest 多策略对比请求体。
type ChunkCompareRequest struct {
	Content string            `json:"content" binding:"required"`
	Configs []ChunkConfigItem `json:"configs" binding:"required,min=1,max=4"`
}

// ChunkItem 单个切块结果。
type ChunkItem struct {
	ID       string         `json:"id"`
	Content  string         `json:"content"`
	Length   int            `json:"length"`
	MetaData map[string]any `json:"metadata"`
}

// ChunkStats 切块统计信息。
type ChunkStats struct {
	TotalChunks     int     `json:"total_chunks"`
	TotalCharacters int     `json:"total_characters"`
	AvgChunkLength  float64 `json:"avg_chunk_length"`
	MinChunkLength  int     `json:"min_chunk_length"`
	MaxChunkLength  int     `json:"max_chunk_length"`
}

// ChunkResult 单策略切块响应。
type ChunkResult struct {
	Chunks []ChunkItem `json:"chunks"`
	Stats  ChunkStats  `json:"stats"`
}

// ChunkStrategyResult 多策略对比中单个策略的结果。
type ChunkStrategyResult struct {
	Strategy string      `json:"strategy"`
	Chunks   []ChunkItem `json:"chunks"`
	Stats    ChunkStats  `json:"stats"`
}

// ChunkCompareResult 多策略对比响应。
type ChunkCompareResult struct {
	Results []ChunkStrategyResult `json:"results"`
	Errors  map[string]string     `json:"errors,omitempty"`
}

// ChunkController 切块控制器，函数字段由 controller 层注入。
type ChunkController struct {
	ChunkFunc   func(*gin.Context)
	CompareFunc func(*gin.Context)
}

// Chunk 执行单策略切块并返回 JSON 结果。
func (c *ChunkController) Chunk(ctx *gin.Context) { c.ChunkFunc(ctx) }

// Compare 执行多策略对比切块并返回 JSON 结果。
func (c *ChunkController) Compare(ctx *gin.Context) { c.CompareFunc(ctx) }

// ChunkService 切块服务，函数字段由 service 层注入。
type ChunkService struct {
	ChunkFunc   func(context.Context, ChunkRequest) (*ChunkResult, error)
	CompareFunc func(context.Context, ChunkCompareRequest) (*ChunkCompareResult, error)
}

// Chunk 执行单策略切块。
func (s *ChunkService) Chunk(ctx context.Context, req ChunkRequest) (*ChunkResult, error) {
	return s.ChunkFunc(ctx, req)
}

// Compare 执行多策略对比切块。
func (s *ChunkService) Compare(ctx context.Context, req ChunkCompareRequest) (*ChunkCompareResult, error) {
	return s.CompareFunc(ctx, req)
}
