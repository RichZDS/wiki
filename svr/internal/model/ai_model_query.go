package model

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// GetAIModelByName 根据模型名称查询 AI 模型配置。
func GetAIModelByName(ctx context.Context, db *gorm.DB, modelName string) (*AIModel, error) {
	var aimodel AIModel
	if err := db.WithContext(ctx).Where("model_name = ?", modelName).First(&aimodel).Error; err != nil {
		return nil, fmt.Errorf("get ai_model %q: %w", modelName, err)
	}
	return &aimodel, nil
}

// ListAllAIModels 查询所有 AI 模型配置。
func ListAllAIModels(ctx context.Context, db *gorm.DB) ([]AIModel, error) {
	var models []AIModel
	if err := db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list ai_models: %w", err)
	}
	return models, nil
}

// GetAIModelAPIKey 根据模型 ID 查询其 API Key。
func GetAIModelAPIKey(ctx context.Context, db *gorm.DB, id int64) (string, error) {
	var record AIModel
	if err := db.WithContext(ctx).Select("api_key").First(&record, id).Error; err != nil {
		return "", fmt.Errorf("get ai_model api_key (id=%d): %w", id, err)
	}
	return record.APIKey, nil
}

// UpdateAIModelStatus 更新指定模型的可用状态和失败原因。
func UpdateAIModelStatus(ctx context.Context, db *gorm.DB, id int64, isUsed int8, failReason string) error {
	result := db.WithContext(ctx).
		Model(&AIModel{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"is_used":     isUsed,
			"fail_reason": failReason,
		})
	if result.Error != nil {
		return fmt.Errorf("update ai_model status (id=%d): %w", id, result.Error)
	}
	return nil
}
