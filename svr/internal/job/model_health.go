package job

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"wiki/internal/model"
	"wiki/internal/model/consts"
	"wiki/pkg/database"
	"wiki/pkg/logger"
	"wiki/pkg/utils"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
	"gorm.io/gorm"
)

type modelProvider struct {
	baseURL   string
	apiKeyEnv string
}

var defaultModelProviders = map[string]modelProvider{
	"openai":   {baseURL: "https://api.openai.com/v1", apiKeyEnv: "OPENAI_API_KEY"},
	"deepseek": {baseURL: "https://api.deepseek.com", apiKeyEnv: "DEEPSEEK_API_KEY"},
	"minimax":  {baseURL: "https://api.minimax.chat", apiKeyEnv: "MINIMAX_API_KEY"},
}

type ModelChecker interface {
	Check(context.Context) error
}

type ModelHealthTask = model.ModelHealthTask
type compatibleModelChecker = model.CompatibleModelChecker

// NewModelHealthTask 创建并初始化对应的实例。
func NewModelHealthTask(db *gorm.DB, checkers map[string]ModelChecker) *ModelHealthTask {
	return &model.ModelHealthTask{
		RunFunc: func(ctx context.Context) error {
			return runModelHealth(ctx, db, checkers, consts.ModelProbeTimeout)
		},
	}
}

// runModelHealth 检查全部模型并同步其可用状态。
func runModelHealth(ctx context.Context, db *gorm.DB, checkers map[string]ModelChecker, timeout time.Duration) error {
	if db == nil {
		return errors.New("model health database is nil")
	}

	models, err := model.ListAllAIModels(ctx, db)
	if err != nil {
		return fmt.Errorf("list models: %w", err)
	}

	var updateErrors []error
	for _, current := range models {
		if err := ctx.Err(); err != nil {
			return errors.Join(append(updateErrors, err)...)
		}

		available, failReason := checkModel(ctx, current.ModelName, checkers, timeout)
		if err := ctx.Err(); err != nil {
			return errors.Join(append(updateErrors, err)...)
		}

		wanted := int8(0)
		nextFailReason := failReason
		if available {
			wanted = 1
			nextFailReason = ""
		}

		if current.IsUsed == wanted && current.FailReason == nextFailReason {
			continue
		}

		if err := model.UpdateAIModelStatus(ctx, db, current.ID, wanted, nextFailReason); err != nil {
			updateErrors = append(updateErrors, fmt.Errorf("update model %q: %w", current.ModelName, err))
			continue
		}

		logger.GetLogger().Printf("[JOB] model %s availability changed to %d", current.ModelName, wanted)
	}

	return errors.Join(updateErrors...)
}

// checkModel 调用指定模型的健康检查器。
func checkModel(ctx context.Context, modelName string, checkers map[string]ModelChecker, timeout time.Duration) (bool, string) {
	checker, ok := checkers[modelName]
	if !ok {
		reason := consts.ModelFailReasonNotFound
		logger.GetLogger().Printf("[JOB] model %s has no registered health checker", modelName)
		return false, reason
	}

	probeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := checker.Check(probeCtx); err != nil {
		reason := err.Error()
		logger.GetLogger().Printf("[JOB] model %s health check failed: %v", modelName, err)
		return false, reason
	}
	return true, ""
}

// checkCompatibleModel 调用兼容 OpenAI 协议的接口进行探测。
func checkCompatibleModel(ctx context.Context, baseURL, apiKey, modelName string) error {
	if strings.TrimSpace(apiKey) == "" {
		return errors.New("API key is not configured")
	}

	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL:     baseURL,
		APIKey:      apiKey,
		Model:       modelName,
		MaxTokens:   utils.Ptr(1),
		Temperature: utils.Ptr(float32(0)),
	})
	if err != nil {
		return fmt.Errorf("create model client: %w", err)
	}

	if _, err := chatModel.Generate(ctx, []*schema.Message{schema.UserMessage("Reply OK.")}); err != nil {
		return fmt.Errorf("generate probe: %w", err)
	}
	return nil
}

// DefaultModelCheckers 从 ai_model 表读取模型并构建健康检查器映射。
func DefaultModelCheckers() map[string]ModelChecker {
	checkers := make(map[string]ModelChecker)

	ctx := context.Background()
	models, err := model.ListAllAIModels(ctx, database.DB)
	if err != nil {
		logger.GetLogger().Printf("[JOB] load ai_model failed: %v", err)
		return checkers
	}

	for _, current := range models {
		provider, ok := defaultModelProviders[current.ModelName]
		if !ok {
			markModelUnavailable(ctx, current.ID, consts.ModelFailReasonNotFound)
			continue
		}

		modelID := strings.TrimSpace(current.ModelId)
		if modelID == "" {
			markModelUnavailable(ctx, current.ID, consts.ModelFailReasonNotFound)
			continue
		}

		checkers[current.ModelName] = newCompatibleModelChecker(database.DB, provider, current.ID, modelID)
	}

	return checkers
}

// resolveModelAPIKey 优先使用 ai_model.api_key，为空时回退到环境变量。
func resolveModelAPIKey(ctx context.Context, db *gorm.DB, id int64, envKey string) string {
	if db != nil {
		if key, err := model.GetAIModelAPIKey(ctx, db, id); err == nil {
			if k := strings.TrimSpace(key); k != "" {
				return k
			}
		}
	}
	return strings.TrimSpace(os.Getenv(envKey))
}

// newCompatibleModelChecker 为指定 provider 创建基于 model_id 的健康检查器。
func newCompatibleModelChecker(db *gorm.DB, provider modelProvider, recordID int64, modelID string) *model.CompatibleModelChecker {
	return &model.CompatibleModelChecker{
		CheckFunc: func(ctx context.Context) error {
			apiKey := resolveModelAPIKey(ctx, db, recordID, provider.apiKeyEnv)
			return checkCompatibleModel(ctx, provider.baseURL, apiKey, modelID)
		},
	}
}

// markModelUnavailable 将模型标记为不可用并记录失败原因。
func markModelUnavailable(ctx context.Context, id int64, failReason string) {
	if err := model.UpdateAIModelStatus(ctx, database.DB, id, 0, failReason); err != nil {
		logger.GetLogger().Printf("[JOB] mark model %d unavailable: %v", id, err)
	}
}
