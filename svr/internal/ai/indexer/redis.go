// Package indexer 封装 eino-ext 的 Redis 索引器，将文档及其向量嵌入存储到 Redis。
package indexer

import (
	"context"
	"fmt"

	"wiki/pkg/database"

	einoembedding "github.com/cloudwego/eino/components/embedding"
	einoindexer "github.com/cloudwego/eino-ext/components/indexer/redis"
	"github.com/redis/go-redis/v9"
)

// RedisIndexerConfig 是创建 Redis Indexer 所需的配置。
type RedisIndexerConfig struct {
	Client    *redis.Client
	KeyPrefix string
	BatchSize int
	Embedding einoembedding.Embedder
}

// NewRedisIndexer 创建并初始化 Redis Indexer 实例。
// 如果 Embedding 为 nil 或 Client 为 nil，返回错误。
func NewRedisIndexer(ctx context.Context, cfg RedisIndexerConfig) (*einoindexer.Indexer, error) {
	if cfg.Embedding == nil {
		return nil, fmt.Errorf("embedding is required for redis indexer")
	}
	if cfg.Client == nil {
		return nil, fmt.Errorf("redis client is required for redis indexer")
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 10
	}

	return einoindexer.NewIndexer(ctx, &einoindexer.IndexerConfig{
		Client:    cfg.Client,
		KeyPrefix: cfg.KeyPrefix,
		BatchSize: cfg.BatchSize,
		Embedding: cfg.Embedding,
	})
}

// NewDefaultIndexer 使用项目默认 Redis 客户端创建 Indexer。
// keyPrefix 为 Redis key 前缀，batchSize 为批量 embedding 大小。
func NewDefaultIndexer(ctx context.Context, emb einoembedding.Embedder, keyPrefix string, batchSize int) (*einoindexer.Indexer, error) {
	if database.RDB == nil {
		return nil, fmt.Errorf("redis client is not initialized")
	}
	return NewRedisIndexer(ctx, RedisIndexerConfig{
		Client:    database.RDB,
		KeyPrefix: keyPrefix,
		BatchSize: batchSize,
		Embedding: emb,
	})
}
