// Package job 模型健康检查：通过 eino 组件接口（embedding.Embedder.EmbedStrings、
// model.BaseChatModel.Generate）发起 1-token 探测，统一 chat 与 embedding 的检查方式。
package job

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"wiki/internal/ai/embedding"
	"wiki/internal/model"
	"wiki/internal/model/consts"
	"wiki/pkg/database"
	"wiki/pkg/logger"
	"wiki/pkg/utils"

	"github.com/cloudwego/eino-ext/components/model/openai"
	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"gorm.io/gorm"
)

// ModelChecker 是模型健康探测的最小接口。
// 实现者通过 eino 的组件 API（Generate / EmbedStrings）发起一次极小请求验证可用性。
type ModelChecker interface {
	Check(context.Context) error
}

type (
	// providerSpec 描述兼容 OpenAI 协议的 chat 端点。
	providerSpec struct {
		baseURL string
		envKey  string
	}
	// chatModelFactory 在每次探测时按当前 api_key 构造 eino ChatModel。
	chatModelFactory func(ctx context.Context) (einomodel.BaseChatModel, error)
	// embedderFactory 在每次探测时按当前 api_key 构造 eino Embedder。
	embedderFactory func(ctx context.Context) (embedding.Embedder, error)
)

var defaultProviders = map[string]providerSpec{
	"openai":   {baseURL: "https://api.openai.com/v1", envKey: "OPENAI_API_KEY"},
	"deepseek": {baseURL: "https://api.deepseek.com", envKey: "DEEPSEEK_API_KEY"},
	"minimax":  {baseURL: "https://api.minimaxi.com/v1", envKey: "MINIMAX_API_KEY"},
	"moonshot": {baseURL: "https://api.moonshot.cn/v1", envKey: "MOONSHOT_API_KEY"},
	"zhipu":    {baseURL: "https://open.bigmodel.cn/api/paas/v4", envKey: "ZHIPU_API_KEY"},
	"qwen":     {baseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1", envKey: "QWEN_API_KEY"},
}

// ModelHealthTask 是 model 包对外暴露的任务句柄类型，便于其他包引用。
type ModelHealthTask = model.ModelHealthTask

// NewModelHealthTask 创建模型健康检查周期任务。
func NewModelHealthTask(db *gorm.DB, checkers map[string]ModelChecker) *ModelHealthTask {
	return &model.ModelHealthTask{
		RunFunc: func(ctx context.Context) error {
			return runModelHealth(ctx, db, checkers, consts.ModelProbeTimeout)
		},
	}
}

// DefaultModelCheckers 从 ai_model 表读取所有需要监控的模型并构建探测器映射。
// 当前支持：
//   - "embedding"  -> 通过 embedding.Embedder.EmbedStrings 探测（Gemini）
//   - "openai/deepseek/minimax" -> 通过 model.BaseChatModel.Generate 探测
func DefaultModelCheckers() map[string]ModelChecker {
	ctx := context.Background()
	log := logger.GetLogger()
	checkers := make(map[string]ModelChecker)

	models, err := model.ListAllAIModelsInNeed(ctx, database.DB)
	if err != nil {
		log.Printf("[JOB] load ai_model failed: %v", err)
		return checkers
	}

	for _, m := range models {
		modelID := strings.TrimSpace(m.ModelId)
		if modelID == "" {
			markUnavailable(ctx, m.ID, consts.ModelFailReasonNotFound)
			continue
		}

		if m.ModelName == "embedding" {
			checkers[m.ModelName] = newEmbeddingChecker(m.ID, modelID, m.Provider)
			continue
		}

		spec, ok := defaultProviders[m.ModelName]
		if !ok {
			// 回退到 provider 字段（兼容通过 provider 字段路由的场景）
			if p := strings.ToLower(strings.TrimSpace(m.Provider)); p != "" {
				spec, ok = defaultProviders[p]
			}
		}
		if !ok {
			markUnavailable(ctx, m.ID, consts.ModelFailReasonNotFound)
			continue
		}
		checkers[m.ModelName] = newChatChecker(spec, m.ID, modelID)
	}
	return checkers
}

// newChatChecker 返回基于 eino BaseChatModel.Generate 的探测器。
func newChatChecker(spec providerSpec, recordID int64, modelID string) ModelChecker {
	factory := func(ctx context.Context) (einomodel.BaseChatModel, error) {
		apiKey := resolveAPIKey(ctx, recordID, spec.envKey)
		if apiKey == "" {
			return nil, errors.New("API key is not configured")
		}
		return openai.NewChatModel(ctx, &openai.ChatModelConfig{
			BaseURL:     resolveBaseURL(spec),
			APIKey:      apiKey,
			Model:       modelID,
			MaxTokens:   utils.Ptr(1),
			Temperature: utils.Ptr(float32(0)),
		})
	}
	return &model.CompatibleModelChecker{
		CheckFunc: func(ctx context.Context) error { return probeChat(ctx, factory) },
	}
}

// newEmbeddingChecker 返回基于 eino embedding.Embedder.EmbedStrings 的探测器。
// provider 参数从 ai_model.provider 字段读取，空值时默认 "gemini"。
func newEmbeddingChecker(recordID int64, modelID, provider string) ModelChecker {
	if provider == "" {
		provider = "gemini"
	}
	factory := func(ctx context.Context) (embedding.Embedder, error) {
		apiKey := resolveAPIKey(ctx, recordID, "GEMINI_API_KEY")
		if apiKey == "" {
			return nil, errors.New("API key is not configured")
		}
		return embedding.NewEmbedderByProvider(ctx, provider, apiKey, modelID)
	}
	return &model.CompatibleModelChecker{
		CheckFunc: func(ctx context.Context) error { return probeEmbed(ctx, factory) },
	}
}

// probeChat 用 1-token Generate 验证 chat 模型可达。
func probeChat(ctx context.Context, factory chatModelFactory) error {
	cm, err := factory(ctx)
	if err != nil {
		return fmt.Errorf("create chat model: %w", err)
	}
	if _, err := cm.Generate(ctx, []*schema.Message{schema.UserMessage("OK")}); err != nil {
		return fmt.Errorf("generate probe: %w", err)
	}
	return nil
}

// probeEmbed 用单条文本 EmbedStrings 验证 embedding 模型可达。
func probeEmbed(ctx context.Context, factory embedderFactory) error {
	emb, err := factory(ctx)
	if err != nil {
		return fmt.Errorf("create embedder: %w", err)
	}
	if _, err := emb.EmbedStrings(ctx, []string{"probe"}); err != nil {
		return fmt.Errorf("embed probe: %w", err)
	}
	return nil
}

// runModelHealth 遍历需要监控的模型并同步可用状态到 ai_model 表。
func runModelHealth(ctx context.Context, db *gorm.DB, checkers map[string]ModelChecker, timeout time.Duration) error {
	if db == nil {
		return errors.New("model health database is nil")
	}

	models, err := model.ListAllAIModelsInNeed(ctx, db)
	if err != nil {
		return fmt.Errorf("list models: %w", err)
	}

	var errs []error
	log := logger.GetLogger()
	for _, current := range models {
		if err := ctx.Err(); err != nil {
			return errors.Join(append(errs, err)...)
		}

		ok, reason := probeOne(ctx, current.ModelName, checkers, timeout)

		wantUsed := int8(0)
		wantReason := reason
		if ok {
			wantUsed, wantReason = 1, ""
		}
		if current.IsUsed == wantUsed && current.FailReason == wantReason {
			continue
		}

		if err := model.UpdateAIModelStatus(ctx, db, current.ID, wantUsed, wantReason); err != nil {
			errs = append(errs, fmt.Errorf("update %q: %w", current.ModelName, err))
			continue
		}
		log.Printf("[JOB] model %s availability=%d", current.ModelName, wantUsed)
	}
	return errors.Join(errs...)
}

// probeOne 调用指定模型的 checker 并返回（是否可用，失败原因）。
func probeOne(ctx context.Context, name string, checkers map[string]ModelChecker, timeout time.Duration) (bool, string) {
	checker, ok := checkers[name]
	if !ok {
		logger.GetLogger().Printf("[JOB] model %s has no registered checker", name)
		return false, consts.ModelFailReasonNotFound
	}
	probeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if err := checker.Check(probeCtx); err != nil {
		logger.GetLogger().Printf("[JOB] model %s health check failed: %v", name, err)
		return false, err.Error()
	}
	return true, ""
}

// resolveAPIKey 优先读取 ai_model.api_key，缺省回退到环境变量。
func resolveAPIKey(ctx context.Context, id int64, envKey string) string {
	if database.DB != nil {
		if key, err := model.GetAIModelAPIKey(ctx, database.DB, id); err == nil {
			if k := strings.TrimSpace(key); k != "" {
				return k
			}
		}
	}
	return strings.TrimSpace(os.Getenv(envKey))
}

// resolveBaseURL 返回 provider 的 Base URL，特殊 provider 支持环境变量覆盖。
func resolveBaseURL(spec providerSpec) string {
	if spec.envKey == "MINIMAX_API_KEY" {
		if u := strings.TrimSpace(os.Getenv("MINIMAX_BASE_URL")); u != "" {
			return u
		}
	}
	return spec.baseURL
}

// markUnavailable 把模型置为不可用并记录失败原因。
func markUnavailable(ctx context.Context, id int64, reason string) {
	if err := model.UpdateAIModelStatus(ctx, database.DB, id, 0, reason); err != nil {
		logger.GetLogger().Printf("[JOB] mark model %d unavailable: %v", id, err)
	}
}
