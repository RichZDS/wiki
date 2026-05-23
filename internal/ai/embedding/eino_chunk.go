package embedding

import (
	"context"

	"github.com/cloudwego/eino/schema"
)

// einoChunker 通过 Eino EmbeddingModel 进行语义切块。
//
// 与固定长度切块不同，语义切块在内容"语义转折点"处切分，使每块内容主题更凝聚。
//
// 算法思路：
//  1. 将文档按句号/换行拆分为句子序列
//  2. 用 Embedder.EmbedStrings 计算每个句子的向量
//  3. 计算相邻句子的余弦相似度
//  4. 在相似度"谷值"处（低于阈值或相对于相邻窗口明显下降）执行切分
//  5. 合并句子片段至接近 ChunkSize，超长再 fallback 到 freeChunker
//
// 依赖：
//   - embedding.Embedder 实例（OpenAI text-embedding-3-small 或同家族模型）
//   - 索引和检索必须使用同一模型，否则向量空间不匹配
//
// 典型配置：
//
//	cfg := ChunkConfig{
//	    ChunkSize:    500,  // 每块最大字符数
//	    ChunkOverlap: 30,   // 语义边界处的延伸字符数
//	}
//
// 当前状态：待实现，直接返回 (nil, nil)。
type einoChunker struct{}

func (c *einoChunker) Chunk(ctx context.Context, content string, cfg ChunkConfig) ([]*schema.Document, error) {
	// TODO: 使用 eino-ext 的 embedding 组件进行语义分割
	// 如按语义相似度边界切分，而非固定长度
	return nil, nil
}
