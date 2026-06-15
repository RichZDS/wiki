package database

import (
	"context"
	"fmt"
	"time"

	"wiki/internal/config"
	"wiki/pkg/logger"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

// VectorRDB 是用于向量搜索（FT.SEARCH）的 Redis 客户端，
// 配置了 Protocol=2 和 UnstableResp3=true 以满足 RediSearch 要求。
var VectorRDB *redis.Client

// InitRedis 初始化对应的基础设施组件。
func InitRedis(cfg config.RedisConfig) {
	RDB = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := RDB.Ping(ctx).Err(); err != nil {
		panic(fmt.Errorf("【启动检查失败】Redis 连接不可用: %w\n程序无法启动，请检查 Redis 配置和网络连接", err))
	}

	logger.GetLogger().Printf("Redis 连接成功 %s:%s (db=%d)", cfg.Host, cfg.Port, cfg.DB)
}

// CloseRedis 关闭对应资源并释放连接。
func CloseRedis() {
	if RDB != nil {
		RDB.Close()
	}
}
