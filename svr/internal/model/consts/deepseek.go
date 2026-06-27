package consts

import "time"

const (
	// RedisKeyAIModelDeepSeek DeepSeek 模型配置在 Redis 中的缓存键。
	RedisKeyAIModelDeepSeek = "aimodel:deepseek"
	// RedisTTLAIModelDeepSeek DeepSeek 模型配置的 Redis 缓存过期时间。
	RedisTTLAIModelDeepSeek = 1 * time.Hour
	// RedisFieldAPIKey AI 模型配置 Redis Hash 中 api_key 字段名。
	RedisFieldAPIKey = "api_key"
	// RedisFieldModelID AI 模型配置 Redis Hash 中 model_id 字段名。
	RedisFieldModelID = "model_id"
	// RedisFieldBaseURL AI 模型配置 Redis Hash 中 base_url 字段名。
	RedisFieldBaseURL = "base_url"
)
