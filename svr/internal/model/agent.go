package model

import (
	"context"

	"github.com/cloudwego/eino-ext/components/model/claude"
	"github.com/cloudwego/eino-ext/components/model/openai"
	chatmodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type OpenAIAgent struct {
	Model *openai.ChatModel
}

type MinimaxAgent struct {
	Model *claude.ChatModel
}

type DeepSeekConfig struct {
	APIKey      string
	Model       string
	MaxSteps    int
	MaxTokens   int
	Temperature float32
	Tools       []tool.BaseTool
}

type DeepSeekAgentState struct {
	Model     *openai.ChatModel
	ToolsNode *compose.ToolsNode
	Tools     []tool.BaseTool
	MaxSteps  int
}

type DeepSeekAgent struct {
	HasToolsFunc     func() bool
	GenerateFunc     func(context.Context, []*schema.Message) (*schema.Message, error)
	StreamFunc       func(context.Context, []*schema.Message) (*schema.StreamReader[*schema.Message], error)
	GetToolInfosFunc func(context.Context) ([]*schema.ToolInfo, error)
	ModelFunc        func() chatmodel.ToolCallingChatModel
}

// HasTools 返回当前代理是否注册了可调用工具。
func (a *DeepSeekAgent) HasTools() bool {
	return a.HasToolsFunc()
}

// Generate 以阻塞方式生成模型回复。
func (a *DeepSeekAgent) Generate(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	return a.GenerateFunc(ctx, messages)
}

// Stream 以流式方式生成模型回复。
func (a *DeepSeekAgent) Stream(ctx context.Context, messages []*schema.Message) (*schema.StreamReader[*schema.Message], error) {
	return a.StreamFunc(ctx, messages)
}

// GetToolInfos 获取当前代理注册的工具描述。
func (a *DeepSeekAgent) GetToolInfos(ctx context.Context) ([]*schema.ToolInfo, error) {
	return a.GetToolInfosFunc(ctx)
}

// Model 返回代理使用的底层工具调用模型。
func (a *DeepSeekAgent) Model() chatmodel.ToolCallingChatModel {
	return a.ModelFunc()
}
