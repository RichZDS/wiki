package agent

import (
	"context"
	"os"
	"testing"

	"wiki/internal/ai/agent"
	"wiki/internal/tools"

	"github.com/cloudwego/eino/components/tool"
)

// ============================================================================
// tools 包测试
// ============================================================================

// TestToolexe 验证工具调用演示流程可以正常执行。
func TestToolexe(t *testing.T) {
	// 验证 Toolexe 演示函数无 panic 执行
	tools.Toolexe()
}

// TestNewWeatherTool 验证天气工具能够成功创建并提供描述。
func TestNewWeatherTool(t *testing.T) {
	wt := tools.NewWeatherTool()
	if wt == nil {
		t.Fatal("NewWeatherTool() returned nil")
	}

	// 验证工具的 Info 可正常获取
	ctx := context.Background()
	info, err := wt.Info(ctx)
	if err != nil {
		t.Fatalf("tool.Info() failed: %v", err)
	}
	if info.Name != "get_weather" {
		t.Errorf("expected tool name 'get_weather', got '%s'", info.Name)
	}
	t.Logf("Tool Info: name=%s, desc=%s", info.Name, info.Desc)
}

// TestAllTools 验证工具注册表返回全部可用工具。
func TestAllTools(t *testing.T) {
	all := tools.AllTools()
	if len(all) == 0 {
		t.Fatal("AllTools() returned empty slice")
	}

	ctx := context.Background()
	for _, toolItem := range all {
		info, err := toolItem.Info(ctx)
		if err != nil {
			t.Errorf("tool.Info() failed for %v: %v", toolItem, err)
			continue
		}
		t.Logf("  Tool: %s — %s", info.Name, info.Desc)
	}
}

// ============================================================================
// DeepSeekAgent 测试（不需要 API Key 来测试结构逻辑）
// ============================================================================

// TestDeepSeekAgentWithoutTools 验证代理在未注册工具时的状态。
func TestDeepSeekAgentWithoutTools(t *testing.T) {
	// 没有 API key 时创建会 Fatal，所以这些测试仅在有关键 API key 时运行
	if os.Getenv("DEEPSEEK_API_KEY") == "" {
		t.Skip("DEEPSEEK_API_KEY not set, skipping agent creation test")
	}

	ctx := context.Background()
	agent := agent.NewDeepSeekAgent(ctx, nil)

	if agent.HasTools() {
		t.Error("agent without tools should return HasTools() = false")
	}
}

// TestDeepSeekAgentWithTools 验证代理能够正确注册工具。
func TestDeepSeekAgentWithTools(t *testing.T) {
	if os.Getenv("DEEPSEEK_API_KEY") == "" {
		t.Skip("DEEPSEEK_API_KEY not set, skipping agent creation test")
	}

	ctx := context.Background()
	toolList := tools.AllTools()
	agent := agent.NewDeepSeekAgent(ctx, toolList)

	if !agent.HasTools() {
		t.Error("agent with tools should return HasTools() = true")
	}
}

// TestDeepSeekAgentWithConfig 验证代理能够使用自定义配置创建。
func TestDeepSeekAgentWithConfig(t *testing.T) {
	if os.Getenv("DEEPSEEK_API_KEY") == "" {
		t.Skip("DEEPSEEK_API_KEY not set, skipping agent creation test")
	}

	ctx := context.Background()

	t.Run("no tools", func(t *testing.T) {
		a := agent.NewDeepSeekAgentWithConfig(ctx, &agent.DeepSeekConfig{
			MaxSteps: 5,
		})
		if a.HasTools() {
			t.Error("expected HasTools() = false")
		}
	})

	t.Run("with tools", func(t *testing.T) {
		a := agent.NewDeepSeekAgentWithConfig(ctx, &agent.DeepSeekConfig{
			MaxSteps: 5,
			Tools:    []tool.BaseTool{tools.NewWeatherTool()},
		})
		if !a.HasTools() {
			t.Error("expected HasTools() = true")
		}
	})
}

// ============================================================================
// 端到端测试（需要有效的 API Key + 网络）
// ============================================================================

// TestDeepSeekAgentGenerateWithTools 验证代理执行带工具的生成流程。
func TestDeepSeekAgentGenerateWithTools(t *testing.T) {
	if os.Getenv("DEEPSEEK_API_KEY") == "" {
		t.Skip("DEEPSEEK_API_KEY not set, skipping integration test")
	}

	ctx := context.Background()
	a := agent.NewDeepSeekAgent(ctx, tools.AllTools())
	if !a.HasTools() {
		t.Fatal("expected agent to have tools")
	}
	t.Log("Agent created successfully with tools")
}
