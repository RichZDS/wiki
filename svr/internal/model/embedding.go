package model

import "context"

// MockEmbedder 是 Embedder 的模拟实现，用于测试语义切块等依赖向量的功能。
//
// 如果 Vectors 为 nil，则自动为每个输入文本生成固定维度的随机向量；
// 如果 Vectors 已赋值，则直接返回预设的向量序列。
type MockEmbedder struct {
	Vectors [][]float64
}

// EmbedStrings 返回模拟的文本向量表示。
func (m *MockEmbedder) EmbedStrings(ctx context.Context, texts []string) ([][]float64, error) {
	if m.Vectors != nil {
		return m.Vectors, nil
	}
	// 未设置预设向量时，生成简单的固定模式向量（用于测试中不关心具体值的场景）
	result := make([][]float64, len(texts))
	for i := range texts {
		// 每个文本生成 3 维模拟向量，值随索引递增以保证相邻句子有较高相似度
		result[i] = []float64{float64(i) * 0.1, float64(i)*0.1 + 0.5, float64(i)*0.1 + 1.0}
	}
	return result, nil
}
