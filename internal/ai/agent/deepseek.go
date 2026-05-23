package agent

import (
	"context"
	"log"
	"os"

	"aisearch/pkg/utils"

	"github.com/cloudwego/eino-ext/components/model/openai"
)

type DeepSeekAgent struct {
	Model *openai.ChatModel
}

func NewDeepSeekAgent(ctx context.Context) *DeepSeekAgent {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		log.Fatal("DEEPSEEK_API_KEY is not set")
	}

	m, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL:     "https://api.deepseek.com",
		APIKey:      apiKey,
		Model:       "deepseek-v4-pro",
		MaxTokens:   utils.Ptr(2048),
		Temperature: utils.Ptr(float32(0.7)),
		TopP:        utils.Ptr(float32(0.7)),
	})
	if err != nil {
		log.Fatalf("failed to create deepseek agent: %v", err)
	}

	return &DeepSeekAgent{Model: m}
}
