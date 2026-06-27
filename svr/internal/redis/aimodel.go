// Package redis 统一管理项目中的 Redis 缓存操作。
// 所有 Redis 读写均通过本包提供的函数完成，避免业务代码直接调用 database.RDB。
package redis

import (
	"context"
	"time"

	"wiki/internal/model/consts"
	"wiki/pkg/database"
	"wiki/pkg/logger"
)

// GetCachedAIModelConfig 从 Redis 缓存中读取指定模型的 API Key、Model ID 和 Base URL。
// 返回值 hit 表示缓存命中。
func GetCachedAIModelConfig(ctx context.Context, modelName string) (apiKey, modelID, baseURL string, hit bool) {
	if database.RDB == nil {
		return "", "", "", false
	}
	key := consts.RedisKeyAIModelDeepSeek // 当前仅 DeepSeek 使用缓存，后续可扩展为 key 模板
	fields, err := database.RDB.HGetAll(ctx, key).Result()
	if err != nil || len(fields) == 0 {
		return "", "", "", false
	}
	apiKey = fields[consts.RedisFieldAPIKey]
	modelID = fields[consts.RedisFieldModelID]
	baseURL = fields[consts.RedisFieldBaseURL]
	if apiKey != "" && modelID != "" {
		logger.GetLogger().Printf("[REDIS] 模型 %s 配置缓存命中", modelName)
		return apiKey, modelID, baseURL, true
	}
	return "", "", "", false
}

// SetCachedAIModelConfig 将指定模型的 API Key、Model ID 和 Base URL 写入 Redis 缓存。
func SetCachedAIModelConfig(ctx context.Context, modelName, apiKey, modelID, baseURL string) {
	if database.RDB == nil {
		return
	}
	key := consts.RedisKeyAIModelDeepSeek
	if err := database.RDB.HSet(ctx, key,
		consts.RedisFieldAPIKey, apiKey,
		consts.RedisFieldModelID, modelID,
		consts.RedisFieldBaseURL, baseURL,
	).Err(); err != nil {
		logger.GetLogger().Printf("[REDIS] 模型 %s 缓存写入失败: %v", modelName, err)
		return
	}
	database.RDB.Expire(ctx, key, consts.RedisTTLAIModelDeepSeek)
	logger.GetLogger().Printf("[REDIS] 模型 %s 配置已缓存（%v）", modelName, consts.RedisTTLAIModelDeepSeek)
}

// 确保导入 time 包（用于 consts 中引用的 time.Duration 常量编译校验）。
var _ = time.Second
