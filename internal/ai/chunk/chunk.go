// Package chunk 提供文档切块能力，是 RAG 流水线中 Transformer 角色的实现。
//
// 本包定义了三层核心概念：
//   - Strategy（策略枚举）：控制切块行为的分发开关
//   - ChunkConfig（配置结构体）：承载调用方传入的切块参数
//   - Chunker（接口）：统一切块入口，对齐 Eino 的 document.Transformer 语义
//
// 调用方通过 NewChunker(strategy) 获取对应的 Chunker 实现：
//
//	free  := chunk.NewChunker(chunk.StrategyFree)         // 递归字符分割
//	md    := chunk.NewChunker(chunk.StrategyMD)           // 元素分类+标题聚合
//	eino  := chunk.NewChunker(chunk.StrategyEino)         // 语义向量分割
//	hier  := chunk.NewChunker(chunk.StrategyHierarchical) // 上下文感知分层切分
//
// 返回类型 []*schema.Document 可直接对接 Eino 的 Indexer.Store。
package chunk

import (
	"context"
	"fmt"

	"wiki/internal/model"
	"wiki/internal/model/consts"

	"github.com/cloudwego/eino/schema"
)

// Strategy 切块策略枚举，决定底层使用哪种分块算法。
type Strategy = consts.Strategy

// 导出切块策略枚举值。
const (
	StrategyFree         = consts.StrategyFree
	StrategyMD           = consts.StrategyMD
	StrategyEino         = consts.StrategyEino
	StrategyHierarchical = consts.StrategyHierarchical
)

// ChunkConfig 切块参数，由调用方根据文档类型和下游模型窗口大小进行调参。
//
// 典型配置：
//
//	// 通用中文文本
//	cfg := ChunkConfig{ChunkSize: 500, ChunkOverlap: 50}
//
//	// 带自定义分隔符
//	cfg := ChunkConfig{
//	    ChunkSize:    300,
//	    ChunkOverlap: 30,
//	    Separators:   []string{"\n\n", "\n", "。"},
//	}
type ChunkConfig = model.ChunkConfig

// Chunker 统一切块接口，对齐 Eino document.Transformer 返回签名。
// 所有切块策略必须实现此接口。
//
// 返回值含义：
//   - []*schema.Document：切分后的文档切片，每个 Document.ID 为 "chunk_N"
//   - Document.MetaData 包含 chunk_index（序号）和 chunk_total（总数）
//   - 空输入返回 (nil, nil)
type Chunker interface {
	Chunk(ctx context.Context, content string, cfg ChunkConfig) ([]*schema.Document, error)
}

// NewChunker 根据策略枚举返回对应的 Chunker 实现。
// 未能识别的策略默认回退到 freeChunker。
func NewChunker(strategy Strategy) Chunker {
	switch strategy {
	case StrategyMD:
		return NewMDChunker()
	case StrategyEino:
		return NewEinoChunker(nil)
	case StrategyHierarchical:
		return NewHierarchicalChunker()
	default:
		return NewFreeChunker()
	}
}

// 以下为 schema.Document.MetaData 写入时的键名引用。
var (
	metaKeyChunkIndex    = consts.MetaKeyChunkIndex
	metaKeyTotalChunks   = consts.MetaKeyTotalChunks
	metaKeyHeadingPath   = consts.MetaKeyHeadingPath
	metaKeyElementTypes  = consts.MetaKeyElementTypes
	metaKeyChunkStrategy = consts.MetaKeyChunkStrategy
	metaKeyChunkRole     = consts.MetaKeyChunkRole
	metaKeyParentContent = consts.MetaKeyParentContent
	metaKeyParentChunkID = consts.MetaKeyParentChunkID
	metaKeyChildChunkIDs = consts.MetaKeyChildChunkIDs
)

// newDocument 构造一个带完整元数据的 Document 切片。
// index 为块的 0-based 序号，total 为切分后的总块数。
func newDocument(content string, index, total int) *schema.Document {
	return &schema.Document{
		ID:      fmt.Sprintf("chunk_%d", index),
		Content: content,
		MetaData: map[string]any{
			metaKeyChunkIndex:  index,
			metaKeyTotalChunks: total,
		},
	}
}

// sanitizeConfig 统一参数兜底，避免各 chunker 重复边界检查。
func sanitizeConfig(cfg *ChunkConfig) {
	if cfg.ChunkSize <= 0 {
		cfg.ChunkSize = 500
	}
	if cfg.ChunkOverlap >= cfg.ChunkSize {
		cfg.ChunkOverlap = cfg.ChunkSize - 1
	}
	if cfg.ChunkOverlap < 0 {
		cfg.ChunkOverlap = 0
	}
}
