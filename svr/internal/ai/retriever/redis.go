// Package retriever 封装 eino-ext 的 Redis 检索器，基于语义相似度从 Redis 中检索文档。
package retriever

import (
	"context"
	"fmt"

	"wiki/internal/config"
	"wiki/pkg/database"
	"wiki/pkg/logger"

	einoretriever "github.com/cloudwego/eino-ext/components/retriever/redis"
	einoembedding "github.com/cloudwego/eino/components/embedding"
	"github.com/redis/go-redis/v9"
)

// InitVectorRedis 初始化用于向量搜索的 Redis 客户端。
// 向量搜索 FT.SEARCH 要求 Protocol=2 且 UnstableResp3=true。
// 该函数从项目统一 Redis 配置读取连接信息并创建专用客户端，存入 database.VectorRDB。
func InitVectorRedis(cfg config.RedisConfig) {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		Protocol: 2, // FT.SEARCH 必须使用 RESP2 协议
	})
	client.Options().UnstableResp3 = true

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		logger.GetLogger().Fatalf("【启动检查失败】Redis 向量搜索连接不可用 (%s): %v", addr, err)
	}

	database.VectorRDB = client
	logger.GetLogger().Printf("Redis 向量搜索连接成功 %s:%s (db=%d)", cfg.Host, cfg.Port, cfg.DB)
}

// CloseVectorRedis 关闭向量搜索 Redis 客户端。
func CloseVectorRedis() {
	if database.VectorRDB != nil {
		database.VectorRDB.Close()
	}
}

// RedisRetrieverConfig 是创建 Redis Retriever 所需的配置。
type RedisRetrieverConfig struct {
	Client            *redis.Client
	Index             string
	VectorField       string
	TopK              int
	DistanceThreshold *float64
	Embedding         einoembedding.Embedder
}

// NewRedisRetriever 创建并初始化 Redis Retriever 实例。
func NewRedisRetriever(ctx context.Context, cfg RedisRetrieverConfig) (*einoretriever.Retriever, error) {
	if cfg.Embedding == nil {
		return nil, fmt.Errorf("embedding is required for redis retriever")
	}
	if cfg.Index == "" {
		return nil, fmt.Errorf("index name is required for redis retriever")
	}
	if cfg.Client == nil {
		return nil, fmt.Errorf("redis client is required for redis retriever")
	}
	if cfg.TopK <= 0 {
		cfg.TopK = 5
	}
	if cfg.VectorField == "" {
		cfg.VectorField = "vector_content"
	}

	return einoretriever.NewRetriever(ctx, &einoretriever.RetrieverConfig{
		Client:            cfg.Client,
		Index:             cfg.Index,
		VectorField:       cfg.VectorField,
		TopK:              cfg.TopK,
		DistanceThreshold: cfg.DistanceThreshold,
		Embedding:         cfg.Embedding,
	})
}

// NewDefaultRetriever 使用项目默认向量搜索 Redis 客户端创建 Retriever。
func NewDefaultRetriever(ctx context.Context, emb einoembedding.Embedder, indexName, vectorField string, topK int) (*einoretriever.Retriever, error) {
	if database.VectorRDB == nil {
		return nil, fmt.Errorf("vector redis client is not initialized")
	}
	return NewRedisRetriever(ctx, RedisRetrieverConfig{
		Client:      database.VectorRDB,
		Index:       indexName,
		VectorField: vectorField,
		TopK:        topK,
		Embedding:   emb,
	})
}
