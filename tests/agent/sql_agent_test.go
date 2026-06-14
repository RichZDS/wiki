package agent

import (
	"context"
	"fmt"
	"testing"

	"wiki/internal/ai/agent"
	"wiki/internal/config"
	"wiki/internal/tools"
	"wiki/pkg/database"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// TestSQLAgent 测试 SQL Agent 的基本对话生成能力。
func TestSQLAgent(t *testing.T) {
	cfg := config.Load()
	database.InitMySQL(cfg.MySQL)

	a := agent.NewDeepSeekAgent(context.Background(), []tool.BaseTool{tools.NewSQLTool()})
	reply, err := a.Generate(context.Background(), []*schema.Message{
		{
			Role:    "user",
			Content: "请查询ai_model表中所有model信息",
		},
	})
	if err != nil {
		t.Fatalf("failed to generate reply: %v", err)
	}
	fmt.Println("reply:", reply.Content)
}
