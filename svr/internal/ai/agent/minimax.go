package agent

import (
	"context"
	"log"

	internalmodel "wiki/internal/model"
	"wiki/pkg/database"
	"wiki/pkg/utils"

	"github.com/cloudwego/eino-ext/components/model/claude"
)

type MinimaxAgent = internalmodel.MinimaxAgent

// NewMinimaxAgent 创建并初始化 Minimax Anthropic 协议模型实例。
func NewMinimaxAgent(ctx context.Context) *MinimaxAgent {
	// 从 ai_model 表读取 Minimax 模型配置。
	aimodel, err := internalmodel.GetAIModelByName(ctx, database.DB, "minimax")
	if err != nil {
		log.Fatalf("failed to find ai_model 'minimax': %v", err)
	}

	apiKey := aimodel.APIKeyValue()
	if apiKey == "" {
		log.Fatal("api_key for minimax is not configured")
	}

	modelID := aimodel.ModelId
	if modelID == "" {
		log.Fatal("model_id for minimax is not configured")
	}

	baseURL := aimodel.BaseURLValue()
	if baseURL == "" {
		baseURL = "https://api.minimax.chat/v1"
	}

	m, err := claude.NewChatModel(ctx, &claude.Config{
		BaseURL: utils.Ptr(baseURL),
		APIKey:  apiKey,
		Model:   modelID,
	})
	if err != nil {
		log.Fatalf("failed to create minimax agent: %v", err)
	}

	return &internalmodel.MinimaxAgent{Model: m}
}
