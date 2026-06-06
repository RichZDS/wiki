package model

import "context"

type MockEmbedder struct {
	Vectors [][]float64
}

// EmbedStrings 为切块测试生成稳定的模拟向量。
func (m *MockEmbedder) EmbedStrings(_ context.Context, texts []string) ([][]float64, error) {
	if m.Vectors != nil {
		return m.Vectors, nil
	}
	result := make([][]float64, len(texts))
	for i, text := range texts {
		vector := make([]float64, 3)
		for j := range vector {
			vector[j] = float64(len(text)+i+j) / 100.0
		}
		result[i] = vector
	}
	return result, nil
}
