package agent

import (
	"context"
	"fmt"
	"log"

	internalmodel "wiki/internal/model"
	"wiki/internal/redis"
	"wiki/pkg/database"
	"wiki/pkg/logger"
	"wiki/pkg/utils"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// DeepSeekAgent 是一个支持 Function Calling 的 AI Agent。
//
// 内部实现了 ReAct（Reasoning + Acting）循环：
//  1. 将用户消息发送给 LLM
//  2. 如果 LLM 返回 ToolCalls，通过 ToolsNode 执行工具
//  3. 将工具结果追加到对话历史
//  4. 再次调用 LLM（此时 LLM 可以基于工具结果生成最终回复）
//  5. 重复直到 LLM 不再请求工具调用，或达到最大步数
//
// 使用方式：
//
//	tools := tools.AllTools()
//	agent := NewDeepSeekAgent(ctx, tools)
//	messages := []*schema.Message{
//	    schema.UserMessage("深圳明天天气怎么样？"),
//	}
//	reply, err := agent.Generate(ctx, messages)
type DeepSeekAgent = internalmodel.DeepSeekAgent

// DeepSeekConfig 是 DeepSeekAgent 的配置。
type DeepSeekConfig = internalmodel.DeepSeekConfig

// NewDeepSeekAgent 创建一个带有工具调用能力的 DeepSeek Agent。
func NewDeepSeekAgent(ctx context.Context, tools []tool.BaseTool) *DeepSeekAgent {
	return NewDeepSeekAgentWithConfig(ctx, &DeepSeekConfig{Tools: tools})
}

// NewDeepSeekAgentWithConfig 使用完整配置创建 DeepSeek Agent。
func NewDeepSeekAgentWithConfig(ctx context.Context, cfg *DeepSeekConfig) *DeepSeekAgent {
	if cfg == nil {
		cfg = &DeepSeekConfig{}
	}

	// 读取 DeepSeek 模型配置（优先 Redis 缓存，回退 MySQL 并写入缓存）
	apiKey, modelID, baseURL := loadDeepSeekConfig(ctx)

	// MaxSteps
	maxSteps := cfg.MaxSteps
	if maxSteps <= 0 {
		maxSteps = 10
	}

	// MaxTokens
	maxTokens := cfg.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 2048
	}

	// Temperature
	temperature := cfg.Temperature
	if temperature <= 0 {
		temperature = 0.7
	}

	// 创建 ChatModel
	m, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL:     baseURL,
		APIKey:      apiKey,
		Model:       modelID,
		MaxTokens:   utils.Ptr(maxTokens),
		Temperature: utils.Ptr(temperature),
		TopP:        utils.Ptr(float32(0.7)),
	})
	if err != nil {
		log.Fatalf("failed to create deepseek chat model: %v", err)
	}

	// 创建 ToolsNode
	var toolsNode *compose.ToolsNode
	tools := cfg.Tools
	if len(tools) > 0 {
		toolsNode, err = compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
			Tools: tools,
		})
		if err != nil {
			log.Fatalf("failed to create tools node: %v", err)
		}
	}

	state := &internalmodel.DeepSeekAgentState{
		Model:     m,
		ToolsNode: toolsNode,
		Tools:     tools,
		MaxSteps:  maxSteps,
	}
	return &internalmodel.DeepSeekAgent{
		HasToolsFunc: func() bool {
			return hasTools(state)
		},
		GenerateFunc: func(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
			return generate(state, ctx, messages)
		},
		StreamFunc: func(ctx context.Context, messages []*schema.Message) (*schema.StreamReader[*schema.Message], error) {
			return stream(state, ctx, messages)
		},
		GetToolInfosFunc: func(ctx context.Context) ([]*schema.ToolInfo, error) {
			return getToolInfos(state, ctx)
		},
		ModelFunc: func() model.ToolCallingChatModel {
			return state.Model
		},
	}
}

// hasTools 判断代理是否具备可执行的工具。
func hasTools(state *internalmodel.DeepSeekAgentState) bool {
	return len(state.Tools) > 0 && state.ToolsNode != nil
}

// Generate 执行对话生成（阻塞式，支持多轮工具调用）。
//
// 如果 Agent 注册了工具，内部自动运行 ReAct 循环：
//
//	用户消息 → LLM → (有 ToolCall?) → 执行工具 → 追加结果 → LLM → ...
//
// 如果 Agent 没有注册工具，等同于直接调用 ChatModel.Generate。
func generate(state *internalmodel.DeepSeekAgentState, ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	if !hasTools(state) {
		return state.Model.Generate(ctx, messages)
	}

	// 将工具绑定到模型。WithTools 返回新实例，不会修改原始模型。
	toolInfos, err := getToolInfos(state, ctx)
	if err != nil {
		return nil, fmt.Errorf("get tool infos: %w", err)
	}

	toolCallingModel, err := state.Model.WithTools(toolInfos)
	if err != nil {
		return nil, fmt.Errorf("bind tools to model: %w", err)
	}

	// -------- ReAct 循环 --------
	history := make([]*schema.Message, len(messages))
	copy(history, messages)

	for step := 0; step < state.MaxSteps; step++ {
		// 1. 调用 LLM
		resp, err := toolCallingModel.Generate(ctx, history)
		if err != nil {
			return nil, fmt.Errorf("model generate (step %d): %w", step, err)
		}

		// 2. 没有工具调用 → 最终回复
		if len(resp.ToolCalls) == 0 {
			return resp, nil
		}

		logger.GetLogger().Printf("[REACT] 第 %d 步: LLM 请求调用 %d 个工具", step+1, len(resp.ToolCalls))
		for _, tc := range resp.ToolCalls {
			logger.GetLogger().Printf("[REACT]   → %s(%s)", tc.Function.Name, tc.Function.Arguments)
		}

		// 3. 将 LLM 返回的 Assistant Message 加入历史
		history = append(history, resp)

		// 4. 执行工具
		toolMessages, err := state.ToolsNode.Invoke(ctx, resp)
		if err != nil {
			return nil, fmt.Errorf("tool invoke (step %d): %w", step, err)
		}

		// 5. 将工具结果（Tool Messages）加入历史
		history = append(history, toolMessages...)
	}

	return nil, fmt.Errorf("reached max steps (%d) without final answer", state.MaxSteps)
}

// Stream 执行对话生成（流式，支持多轮工具调用）。
//
// 返回的 StreamReader 推送最终回复的各个 chunk。
// 内部的工具调用过程对调用方透明——调用方只会看到最终的自然语言输出。
//
// 注意：DeepSeek 的流式 ToolCalls 与 OpenAI 行为一致，
// 使用默认的 firstChunkStreamToolCallChecker 即可。
func stream(state *internalmodel.DeepSeekAgentState, ctx context.Context, messages []*schema.Message) (*schema.StreamReader[*schema.Message], error) {
	if !hasTools(state) {
		return state.Model.Stream(ctx, messages)
	}

	toolInfos, err := getToolInfos(state, ctx)
	if err != nil {
		return nil, fmt.Errorf("get tool infos: %w", err)
	}

	toolCallingModel, err := state.Model.WithTools(toolInfos)
	if err != nil {
		return nil, fmt.Errorf("bind tools to model: %w", err)
	}

	// 使用 Pipe 创建流式通道。工具调用在 goroutine 中完成，
	// 对外只输出最终的自然语言回复。
	sr, sw := schema.Pipe[*schema.Message](1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.GetLogger().Printf("[REACT] panic in stream: %v", r)
				sw.Send(nil, fmt.Errorf("panic: %v", r))
			}
		}()

		defer sw.Close()

		history := make([]*schema.Message, len(messages))
		copy(history, messages)

		for step := 0; step < state.MaxSteps; step++ {
			// 流式调用 LLM
			streamReader, err := toolCallingModel.Stream(ctx, history)
			if err != nil {
				sw.Send(nil, fmt.Errorf("model stream (step %d): %w", step, err))
				return
			}

			// 收集流式输出，判断是否包含 ToolCalls
			var fullMsg *schema.Message
			for {
				chunk, err := streamReader.Recv()
				if err != nil {
					break // EOF
				}
				if fullMsg == nil {
					fullMsg = chunk
				} else {
					fullMsg.Content += chunk.Content
					fullMsg.ToolCalls = append(fullMsg.ToolCalls, chunk.ToolCalls...)
				}
			}
			streamReader.Close()

			if fullMsg == nil {
				sw.Send(nil, fmt.Errorf("empty response at step %d", step))
				return
			}

			// 没有工具调用 → 输出最终回复
			if len(fullMsg.ToolCalls) == 0 {
				// 将完整消息拆分为 chunk 输出（模拟流式）
				sw.Send(fullMsg, nil)
				return
			}

			logger.GetLogger().Printf("[REACT] 第 %d 步: LLM 请求调用 %d 个工具", step+1, len(fullMsg.ToolCalls))

			// 执行工具
			history = append(history, fullMsg)
			toolMessages, err := state.ToolsNode.Invoke(ctx, fullMsg)
			if err != nil {
				sw.Send(nil, fmt.Errorf("tool invoke (step %d): %w", step, err))
				return
			}
			history = append(history, toolMessages...)
		}

		sw.Send(nil, fmt.Errorf("reached max steps (%d) without final answer", state.MaxSteps))
	}()

	return sr, nil
}

// getToolInfos 从注册的工具列表中提取工具描述。
func getToolInfos(state *internalmodel.DeepSeekAgentState, ctx context.Context) ([]*schema.ToolInfo, error) {
	toolInfos := make([]*schema.ToolInfo, 0, len(state.Tools))
	for _, t := range state.Tools {
		info, err := t.Info(ctx)
		if err != nil {
			return nil, err
		}
		toolInfos = append(toolInfos, info)
	}
	return toolInfos, nil
}

// loadDeepSeekConfig 从 Redis 缓存或 MySQL 数据库加载 DeepSeek 模型配置。
// 优先读取 Redis 缓存，未命中时查询 MySQL 并将结果写入 Redis（1 小时过期）。
func loadDeepSeekConfig(ctx context.Context) (apiKey, modelID, baseURL string) {
	// 1. 尝试从 Redis 读取
	if key, id, u, hit := redis.GetCachedAIModelConfig(ctx, "deepseek"); hit {
		if u == "" {
			u = "https://api.deepseek.com"
		}
		return key, id, u
	}

	// 2. Redis 未命中或不可用，从 MySQL 读取
	aimodel, err := internalmodel.GetAIModelByName(ctx, database.DB, "deepseek")
	if err != nil {
		log.Fatalf("failed to find ai_model 'deepseek': %v", err)
	}
	apiKey = aimodel.APIKeyValue()
	if apiKey == "" {
		log.Fatal("api_key for deepseek is not configured")
	}
	modelID = aimodel.ModelId
	if modelID == "" {
		log.Fatal("model_id for deepseek is not configured")
	}
	baseURL = aimodel.BaseURLValue()
	if baseURL == "" {
		baseURL = "https://api.deepseek.com"
	}

	// 3. 写入 Redis 缓存
	redis.SetCachedAIModelConfig(ctx, "deepseek", apiKey, modelID, baseURL)

	return apiKey, modelID, baseURL
}
