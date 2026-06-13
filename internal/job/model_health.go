package job

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"wiki/internal/model"
	"wiki/pkg/logger"
	"wiki/pkg/utils"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
	"gorm.io/gorm"
)

const (
	ModelHealthInterval = 5 * time.Minute
	ModelProbeTimeout   = 15 * time.Second
)

type ModelChecker interface {
	Check(context.Context) error
}

type ModelHealthTask = model.ModelHealthTask
type compatibleModelChecker = model.CompatibleModelChecker

// NewModelHealthTask 创建并初始化对应的实例。
func NewModelHealthTask(db *gorm.DB, checkers map[string]ModelChecker) *ModelHealthTask {
	return &model.ModelHealthTask{
		RunFunc: func(ctx context.Context) error {
			return runModelHealth(ctx, db, checkers, ModelProbeTimeout)
		},
	}
}

// runModelHealth 检查全部模型并同步其可用状态。
func runModelHealth(ctx context.Context, db *gorm.DB, checkers map[string]ModelChecker, timeout time.Duration) error {
	if db == nil {
		return errors.New("model health database is nil")
	}

	var models []model.AIModel
	if err := db.WithContext(ctx).Find(&models).Error; err != nil {
		return fmt.Errorf("list models: %w", err)
	}

	var updateErrors []error
	for _, current := range models {
		if err := ctx.Err(); err != nil {
			return errors.Join(append(updateErrors, err)...)
		}

		available := checkModel(ctx, current.ModelName, checkers, timeout)
		if err := ctx.Err(); err != nil {
			return errors.Join(append(updateErrors, err)...)
		}

		wanted := int8(0)
		if available {
			wanted = 1
		}

		if current.IsUsed == wanted {
			continue
		}

		result := db.WithContext(ctx).
			Model(&model.AIModel{}).
			Where("id = ?", current.ID).
			Update("is_used", wanted)
		if result.Error != nil {
			updateErrors = append(updateErrors, fmt.Errorf("update model %q: %w", current.ModelName, result.Error))
			continue
		}

		logger.GetLogger().Printf("[JOB] model %s availability changed to %d", current.ModelName, wanted)
	}

	return errors.Join(updateErrors...)
}

// checkModel 调用指定模型的健康检查器。
func checkModel(ctx context.Context, modelName string, checkers map[string]ModelChecker, timeout time.Duration) bool {
	checker, ok := checkers[modelName]
	if !ok {
		logger.GetLogger().Printf("[JOB] model %s has no registered health checker", modelName)
		return false
	}

	probeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := checker.Check(probeCtx); err != nil {
		logger.GetLogger().Printf("[JOB] model %s health check failed: %v", modelName, err)
		return false
	}
	return true
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

// DefaultModelCheckers 负责处理当前模块中的对应业务逻辑。
func DefaultModelCheckers() map[string]ModelChecker {
	openAIModel := envOrDefault("OPENAI_MODEL_ID", "gpt-4o")
	deepSeekModel := envOrDefault("DEEPSEEK_MODEL_ID", "deepseek-v4-pro")

	return map[string]ModelChecker{
		openAIModel: &model.CompatibleModelChecker{
			CheckFunc: func(ctx context.Context) error {
				return checkCompatibleModel(ctx, "https://api.openai.com/v1", os.Getenv("OPENAI_API_KEY"), openAIModel)
			},
		},
		deepSeekModel: &model.CompatibleModelChecker{
			CheckFunc: func(ctx context.Context) error {
				return checkCompatibleModel(ctx, "https://api.deepseek.com", os.Getenv("DEEPSEEK_API_KEY"), deepSeekModel)
			},
		},
	}
}

// envOrDefault 负责处理当前模块中的对应业务逻辑。
func envOrDefault(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}
