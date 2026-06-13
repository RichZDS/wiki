package agent

import (
	"context"
	"log"
	"os"

	internalmodel "wiki/internal/model"
	"wiki/pkg/utils"

	"github.com/cloudwego/eino-ext/components/model/openai"
)

type OpenAIAgent = internalmodel.OpenAIAgent

// NewOpenAIAgent 创建并初始化对应的实例。
func NewOpenAIAgent(ctx context.Context) *OpenAIAgent {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY is not set")
	}

	modelID := os.Getenv("OPENAI_MODEL_ID")
	if modelID == "" {
		modelID = "gpt-4o"
	}

	m, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL:         "https://api.openai.com/v1",
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
