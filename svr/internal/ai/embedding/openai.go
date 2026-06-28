package embedding

import (
	"context"
	"fmt"
	"os"

	"wiki/pkg/utils"

	"github.com/cloudwego/eino-ext/components/embedding/openai"
)

// OpenAIEmbedder 是基于 OpenAI API 的文本向量化实现，支持 text-embedding-3-small、
// text-embedding-3-large、text-embedding-ada-002 等模型。
type OpenAIEmbedder = openai.Embedder

// NewOpenAIEmbedder 创建 OpenAI Embedder 实例。
// 环境变量：
//   - OPENAI_API_KEY（必填）— API 密钥
//   - OPENAI_EMBEDDING_MODEL（可选，默认 text-embedding-3-small）— 模型 ID
//
// 返回值：成功时返回 Embedder，失败时返回错误。
func NewOpenAIEmbedder(ctx context.Context, apiKey string, modelID string, baseURL string, region string) (*OpenAIEmbedder, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is not set")
	}
	if modelID == "" {
		modelID = "text-embedding-3-small"
	}
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	if region == "" {
		region = "us-east-1"
	}
	emb, err := openai.NewEmbedder(ctx, &openai.EmbeddingConfig{
		APIKey: apiKey,
		Model:  modelID,
		// text-embedding-3-small 默认 1536 维，可通过 OPENAI_EMBEDDING_DIMENSIONS 调整
		Dimensions: extractEmbeddingDimensions(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create openai embedder: %w", err)
	}

	return emb, nil
}

// extractEmbeddingDimensions 从环境变量中解析向量维度。
func extractEmbeddingDimensions() *int {
	if dims := os.Getenv("OPENAI_EMBEDDING_DIMENSIONS"); dims != "" {
		// 简单解析，由调用方保证合法
		var n int
		for _, c := range dims {
			if c >= '0' && c <= '9' {
				n = n*10 + int(c-'0')
			}
		}
		if n > 0 {
			return utils.Ptr(n)
		}
	}
	return nil
}
