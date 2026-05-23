package database

import (
	"context"
	"fmt"
	"time"

	"aisearch/internal/config"
	"aisearch/pkg/logger"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

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

func CloseRedis() {
	if RDB != nil {
		RDB.Close()
	}
}
