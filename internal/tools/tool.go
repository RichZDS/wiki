// Package tools 演示 Eino 框架中 AI 函数调用（Function Calling）的完整流程。
//
// 函数调用是 LLM 与外部系统交互的核心机制。大模型本身无法查询实时数据，
// 但可以通过"工具调用"让外部代码代为执行，再将结果返回给模型。
//
// 完整流程（Round-Trip）：
//
//	用户提问 → LLM（决定调用哪个工具 + 参数）
//	          → Assistant Message（含 ToolCalls）
//	          → ToolsNode.Invoke（执行工具，得到结果）
//	          → Tool Message（含工具返回内容）
//	          → LLM（基于工具结果生成最终回复）
//	          → 用户看到最终答案
//
// 本文件聚焦中间环节：从 "Assistant Message 携带 ToolCalls"
// 到 "ToolsNode 执行工具并返回 Tool Message" 的过程。
package tools

import (
	"context"
	"fmt"

	"wiki/internal/model"
	"wiki/pkg/logger"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// ============================================================================
// 第一步：定义工具的输入/输出结构体
// ============================================================================

// WeatherInput 是天气查询工具的输入参数。
// InferTool 会自动根据字段和 json tag 生成 JSON Schema 传给 LLM。
type WeatherInput = model.WeatherInput
type WeatherOutput = model.WeatherOutput

// ============================================================================
// 第二步：实现工具的核心逻辑
// ============================================================================

// getWeather 是天气查询的实际实现。
// 签名必须符合 utils.InvokeFunc[T, D] 泛型约束：
//
//	type InvokeFunc[T, D any] func(ctx context.Context, input T) (output D, err error)
//
// 在实际项目中，这里应调用第三方天气 API。本例用 Mock 数据演示。
func getWeather(ctx context.Context, input WeatherInput) (WeatherOutput, error) {
	logger.GetLogger().Printf("[TOOL] 正在查询 %s 的天气，日期：%s", input.City, input.Date)

	// Mock 返回数据
	result := WeatherOutput{
		City:        input.City,
		Date:        input.Date,
		Temperature: 26,
		Condition:   "晴",
		Humidity:    65,
	}

	return result, nil
}

// ============================================================================
// 第三步：用 InferTool 将 Go 函数包装为 Eino Tool
// ============================================================================

// NewWeatherTool 创建一个符合 Eino tool.InvokableTool 接口的天气工具。
//
// InferTool 做了三件事：
//  1. 从 WeatherInput 的 json tag 推断 JSON Schema → 传给 LLM
//  2. 在 ToolsNode 调用时，自动将 LLM 传来的 JSON 参数反序列化为 WeatherInput
//  3. 将 getWeather 的返回值 JSON 序列化后写入 Tool Message 的 Content
func NewWeatherTool() tool.InvokableTool {
	weatherTool, err := utils.InferTool(
		"get_weather", // 工具名称（LLM 通过这个名称决定调用哪个工具）
		`查询指定城市在指定日期的天气情况。
返回温度（摄氏度）、天气状况（晴/多云/雨/雪）和湿度百分比。
适用场景：用户询问"今天热不热"、"明天会下雨吗"等问题时调用。`,
		getWeather, // 工具的实际执行函数
	)
	if err != nil {
		// InferTool 仅在输入类型不合法时才会出错（如不支持的字段类型）
		logger.GetLogger().Printf("[ERROR] 创建天气工具失败: %v", err)
		return nil
	}
	return weatherTool
}

// ============================================================================
// 完整示例：模拟一次 AI 函数调用
// ============================================================================

// Toolexe 演示 ToolsNode 的完整执行流程。
//
// 时序（注释中以 ① ② ③ 标记）：
//
//	① 用户向 LLM 提问："深圳明天天气怎么样？"
//	② LLM 分析后决定调用 get_weather({"city":"深圳","date":"tomorrow"})
//	   并以 Assistant Message（含 ToolCalls）输出
//	③ ToolsNode 收到 Assistant Message，根据 ToolCalls 匹配工具并执行
//	④ ToolsNode 返回 Tool Message（role="tool"，content 为工具返回的 JSON）
//	⑤ 将 Tool Message 回传给 LLM
//	⑥ LLM 基于工具结果生成自然语言回复："深圳明天晴，26°C..."
func Toolexe() {
	ctx := context.Background()

	// -----------------------------------------------------------------------
	// 步骤 A：注册工具，创建 ToolsNode
	// -----------------------------------------------------------------------
	weatherTool := NewWeatherTool()
	if weatherTool == nil {
		return
	}

	toolsNode, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
		Tools: []tool.BaseTool{weatherTool},
		// ExecuteSequentially: true,  // 默认 false，多个工具并行执行
	})
	if err != nil {
		logger.GetLogger().Printf("[ERROR] 创建 ToolsNode 失败: %v", err)
		return
	}

	// -----------------------------------------------------------------------
	// 步骤 B：模拟 LLM 返回的 Assistant Message（包含 ToolCalls）
	//
	// 在实际项目中，这来自 ChatModel.Generate() 的返回值。
	// 当 LLM 认为需要调用工具时，返回的 Message.Role = Assistant，
	// 且 ToolCalls 字段非空。
	// -----------------------------------------------------------------------
	input := &schema.Message{
		Role:    schema.Assistant,
		Content: "", // 有 ToolCalls 时，Content 通常为空
		ToolCalls: []schema.ToolCall{
			{
				ID:   "call_abc123", // 工具调用唯一 ID，结果需要关联此 ID
				Type: "function",
				Function: schema.FunctionCall{
					Name:      "get_weather",                     // 必须与注册的工具名称一致
					Arguments: `{"city":"深圳","date":"tomorrow"}`, // JSON 字符串
				},
			},
		},
	}

	fmt.Println("========== 输入 Message（来自 LLM 的 Assistant Message）==========")
	fmt.Printf("  Role:       %s\n", input.Role)
	fmt.Printf("  ToolCalls:  %d 个\n", len(input.ToolCalls))
	fmt.Printf("  → Name:     %s\n", input.ToolCalls[0].Function.Name)
	fmt.Printf("  → Arguments: %s\n", input.ToolCalls[0].Function.Arguments)
	fmt.Println()

	// -----------------------------------------------------------------------
	// 步骤 C：ToolsNode.Invoke 执行工具
	//
	// ToolsNode 内部做了：
	//   1. 解析 ToolCalls，按 name 匹配注册的工具
	//   2. 将 Arguments（JSON）反序列化为 WeatherInput
	//   3. 调用 getWeather(ctx, WeatherInput{City:"深圳", Date:"tomorrow"})
	//   4. 将返回的 WeatherOutput 序列化为 JSON 字符串
	//   5. 封装为 Tool Message（Role=Tool, Content=JSON结果）
	//   6. 返回 []*schema.Message（每个 ToolCall 对应一个 Tool Message）
	// -----------------------------------------------------------------------
	toolMessages, err := toolsNode.Invoke(ctx, input)
	if err != nil {
		logger.GetLogger().Printf("[ERROR] 工具调用失败: %v", err)
		return
	}

	// -----------------------------------------------------------------------
	// 步骤 D：查看工具返回的 Tool Message
	//
	// 每个 Tool Message 对应输入中的一个 ToolCall，
	// 通过 ToolCallID 关联。
	// -----------------------------------------------------------------------
	fmt.Println("========== 输出 Messages（ToolsNode 返回的 Tool Messages）==========")
	for i, msg := range toolMessages {
		fmt.Printf("  [%d] Role:        %s\n", i, msg.Role)       // schema.Tool
		fmt.Printf("  [%d] ToolCallID:  %s\n", i, msg.ToolCallID) // 关联输入的 ToolCall.ID
		fmt.Printf("  [%d] ToolName:    %s\n", i, msg.ToolName)   // 工具名称
		fmt.Printf("  [%d] Content:     %s\n", i, msg.Content)    // JSON 格式的工具返回值
	}
	fmt.Println()

	// -----------------------------------------------------------------------
	// 步骤 E（概念展示）：将 Tool Messages 回传给 LLM
	//
	// 在实际项目中，下一步是将这些 Tool Messages 追加到对话历史中，
	// 再次调用 ChatModel.Generate()，让 LLM 基于工具结果生成用户可读的回答：
	//
	//   messages = append(messages, input)           // Assistant Message
	//   messages = append(messages, toolMessages...) // Tool Messages
	//   reply, _ := chatModel.Generate(ctx, messages)
	//   // reply.Content: "深圳明天（5月31日）天气晴朗，气温26°C，湿度65%，适合出行。"
	//
	// 至此，一次完整的 AI 函数调用 Round-Trip 就完成了。
	// -----------------------------------------------------------------------
	fmt.Println("========== 流程说明 ==========")
	fmt.Println("① LLM 决定调用工具 → Assistant Message（含 ToolCalls）")
	fmt.Println("② ToolsNode 执行工具 → Tool Messages（含执行结果）")
	fmt.Println("③ Tool Messages 回传 LLM → 生成自然语言回答")
	fmt.Println()
	fmt.Println("以上 ② 的步骤已在本次演示中完成。")
}

// ============================================================================
// 工具注册表：所有可用工具的集中管理
// ============================================================================

// AllTools 返回所有已注册的工具列表。
// 新增工具后，在这里追加即可自动对 Agent 生效。
func AllTools() []tool.BaseTool {
	return []tool.BaseTool{
		NewWeatherTool(),
		NewSQLTool(),
	}
}
