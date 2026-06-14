package embedding

import (
	"context"
	"log"
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
func NewOpenAIEmbedder(ctx context.Context) *OpenAIEmbedder {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY is not set")
	}

	modelID := os.Getenv("OPENAI_EMBEDDING_MODEL")
	if modelID == "" {
		modelID = "text-embedding-3-small"
	}

	emb, err := openai.NewEmbedder(ctx, &openai.EmbeddingConfig{
		APIKey: apiKey,
		Model:  modelID,
		// text-embedding-3-small 默认 1536 维，可通过 OPENAI_EMBEDDING_DIMENSIONS 调整
		Dimensions: extractEmbeddingDimensions(),
	})
	if err != nil {
		log.Fatalf("failed to create openai embedder: %v", err)
	}

	return emb
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
