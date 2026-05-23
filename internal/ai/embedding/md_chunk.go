package embedding

import (
	"context"

	"github.com/cloudwego/eino/schema"
)

// mdChunker 按 Markdown 标题层级切块。
//
// 算法思路：
//  1. 解析 Markdown AST（可使用 github.com/yuin/goldmark）
//  2. 以 H1-H6 标题为边界将文档拆分为逻辑 section
//  3. 每个 section 标题路径写入 Document.MetaData["heading_path"]
//     形如 "第一章 > 1.1 概述 > 1.1.1 细节"
//  4. 若某个 section 内容超过 ChunkSize，交由 freeChunker 二次细分
//
// 典型配置：
//
//	cfg := ChunkConfig{
//	    ChunkSize:    500,  // 每块最大字符数
//	    ChunkOverlap: 50,   // 块间重叠，保持上下文连续性
//	}
//
// 当前状态：待实现，直接返回 (nil, nil)。
type mdChunker struct{}

func (c *mdChunker) Chunk(ctx context.Context, content string, cfg ChunkConfig) ([]*schema.Document, error) {
	// TODO: 解析 Markdown AST，按 H1-H6 标题拆分为 section，超长再细分
	// Metadata 中附带 heading 路径信息
	return nil, nil
}
