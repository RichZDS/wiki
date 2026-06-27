package embedding

import (
	"context"
	"fmt"
	"strings"

	"wiki/internal/model"
	"wiki/pkg/database"
)

// provider 常量定义，对应 ai_model 表的 provider 字段。
const (
	ProviderGemini = "gemini"
	ProviderOpenAI = "openai"
	ProviderArk    = "ark"
)

// NewEmbedderFromDB 从 ai_model 表读取 model_name="embedding" 的配置，
// 根据 provider 字段创建对应的 Embedder 实例。
//
// provider 取值与实现映射：
//   - "gemini"（或空，向后兼容）→ GeminiEmbedder
//   - "openai" → OpenAIEmbedder
//   - "ark" → ArkEmbedder（火山引擎）
//
// 返回的 Embedder 实现了 eino 的 embedding.Embedder 接口，
// 可直接用于 Indexer 和 Retriever。
func NewEmbedderFromDB(ctx context.Context) (Embedder, error) {
	aimodel, err := model.GetAIModelByName(ctx, database.DB, "embedding")
	if err != nil {
		return nil, fmt.Errorf("query ai_model 'embedding': %w", err)
	}
	if aimodel.APIKey == nil || *aimodel.APIKey == "" {
		return nil, fmt.Errorf("api_key for 'embedding' model is not configured")
	}
	if aimodel.ModelId == "" {
		return nil, fmt.Errorf("model_id for 'embedding' model is not configured")
	}

	provider := strings.ToLower(strings.TrimSpace(aimodel.Provider))
	if provider == "" {
		provider = ProviderGemini // 向后兼容：旧数据未设置 provider 时默认 Gemini
	}

	return NewEmbedderByProvider(ctx, provider, *aimodel.APIKey, aimodel.ModelId)
}

// NewEmbedderByProvider 根据 provider 名称创建对应的 Embedder 实例。
func NewEmbedderByProvider(ctx context.Context, provider, apiKey, modelID string) (Embedder, error) {
	switch provider {
	case ProviderGemini:
		return NewGeminiEmbedder(ctx, apiKey, modelID)
	case ProviderOpenAI:
		return NewOpenAIEmbedder(ctx, apiKey, modelID, "", "")
	case ProviderArk:
		return NewArkEmbedder(ctx, apiKey, modelID, "", "")
	default:
		return nil, fmt.Errorf("unknown embedding provider: %s (supported: gemini, openai, ark)", provider)
	}
}
