package agent

import (
	"context"
	"log"

	internalmodel "wiki/internal/model"
	"wiki/pkg/database"
	"wiki/pkg/utils"

	"github.com/cloudwego/eino-ext/components/model/openai"
)

type OpenAIAgent = internalmodel.OpenAIAgent

// NewOpenAIAgent 创建并初始化对应的实例。
func NewOpenAIAgent(ctx context.Context) *OpenAIAgent {
	// 从 ai_model 表读取 OpenAI 模型配置
	aimodel, err := internalmodel.GetAIModelByName(ctx, database.DB, "openai")
	if err != nil {
		log.Fatalf("failed to find ai_model 'openai': %v", err)
	}

	apiKey := aimodel.APIKeyValue()
	if apiKey == "" {
		log.Fatal("api_key for openai is not configured")
	}

	modelID := aimodel.ModelId
	if modelID == "" {
		log.Fatal("model_id for openai is not configured")
	}

	baseURL := aimodel.BaseURLValue()
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	m, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL:         baseURL,
		APIKey:          apiKey,
		Model:           modelID,
		MaxTokens:       utils.Ptr(2048),
		Temperature:     utils.Ptr(float32(0.7)),
		TopP:            utils.Ptr(float32(0.7)),
		ReasoningEffort: openai.ReasoningEffortLevelLow,
	})
	if err != nil {
		log.Fatalf("failed to create openai agent: %v", err)
	}

	return &internalmodel.OpenAIAgent{Model: m}
}
