package embedding

import "context"

// Embedder 计算文本序列的稠密向量表示。
// 实现者（如 OpenAI text-embedding-3-small 等）应保证同一模型的一致性。
type Embedder interface {
	EmbedStrings(ctx context.Context, texts []string) ([][]float64, error)
}
