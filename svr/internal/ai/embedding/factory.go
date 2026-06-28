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

// abilityKeywordEmbedding 是 ai_model.ability 字段中标识 embedding 能力的关键词。
const abilityKeywordEmbedding = "embedding"

// NewEmbedderFromDB 从 ai_model 表读取 ability 包含 "embedding" 且 is_used=1 的第一条配置，
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
	aimodel, err := model.GetFirstAIModelByAbility(ctx, database.DB, abilityKeywordEmbedding)
	if err != nil {
		return nil, fmt.Errorf("query ai_model with ability 'embedding': %w", err)
	}
	if aimodel.APIKey == nil || *aimodel.APIKey == "" {
		return nil, fmt.Errorf("api_key for embedding model %q is not configured", aimodel.ModelName)
	}
	if aimodel.ModelId == "" {
		return nil, fmt.Errorf("model_id for embedding model %q is not configured", aimodel.ModelName)
	}

	provider := strings.ToLower(strings.TrimSpace(aimodel.Provider))
	if provider == "" {
		provider = ProviderGemini // 向后兼容：旧数据未设置 provider 时默认 Gemini
	}

	return NewEmbedderByProvider(ctx, provider, *aimodel.APIKey, aimodel.ModelId, aimodel.BaseURLValue())
}

// NewEmbedderByProvider 根据 provider 名称创建对应的 Embedder 实例。
func NewEmbedderByProvider(ctx context.Context, provider, apiKey, modelID, baseURL string) (Embedder, error) {
	switch provider {
	case ProviderGemini:
		return NewGeminiEmbedder(ctx, apiKey, modelID)
	case ProviderOpenAI:
		return NewOpenAIEmbedder(ctx, apiKey, modelID, baseURL, "")
	case ProviderArk:
		return NewArkEmbedder(ctx, apiKey, modelID, baseURL, "")
	default:
		return nil, fmt.Errorf("unknown embedding provider: %s (supported: gemini, openai, ark)", provider)
	}
}
