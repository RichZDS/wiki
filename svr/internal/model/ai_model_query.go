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

// GetFirstAIModelByAbility 查询 ability 字段包含指定关键词且已启用的第一条模型记录。
// abilityKeyword 使用 LIKE 模糊匹配（如 "embedding" 匹配 "embedding,text"）。
func GetFirstAIModelByAbility(ctx context.Context, db *gorm.DB, abilityKeyword string) (*AIModel, error) {
	var aimodel AIModel
	if err := db.WithContext(ctx).
		Where("is_used = ?", 1).
		Where("ability LIKE ?", "%"+abilityKeyword+"%").
		First(&aimodel).Error; err != nil {
		return nil, fmt.Errorf("get ai_model by ability %q: %w", abilityKeyword, err)
	}
	return &aimodel, nil
}

// ListAllAIModelsInNeed 查询所有 AI 模型配置。
func ListAllAIModelsInNeed(ctx context.Context, db *gorm.DB) ([]AIModel, error) {
	var models []AIModel
	if err := db.WithContext(ctx).Where("is_check = ?", 0).Find(&models).Error; err != nil {
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
	return record.APIKeyValue(), nil
}

// UpdateAIModelStatus 更新指定模型的可用状态和失败原因。
func UpdateAIModelStatus(ctx context.Context, db *gorm.DB, id int64, isUsed int8, failReason string) error {
	result := db.WithContext(ctx).
		Model(&AIModel{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"is_used":     isUsed,
			"is_check":    isUsed,
			"fail_reason": failReason,
		})
	if result.Error != nil {
		return fmt.Errorf("update ai_model status (id=%d): %w", id, result.Error)
	}
	return nil
}
