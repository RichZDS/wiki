// Package rag 整合 RAG 流水线的初始化和业务逻辑。
//
// 设计参考 eino-ext 官方 README：
//   - https://github.com/cloudwego/eino-ext/blob/main/components/indexer/redis/README_zh.md
//   - https://github.com/cloudwego/eino-ext/blob/main/components/retriever/redis/README_zh.md
//
// 调用方在启动时调用 Init 获取 *model.RAGService，
// 退出时调用 Cleanup 释放 Redis 连接。
package rag

import (
	"context"
	"fmt"
	"strings"

	"wiki/internal/ai/chunk"
	"wiki/internal/ai/embedding"
	"wiki/internal/config"
	"wiki/internal/model"
	"wiki/internal/model/consts"
	"wiki/pkg/logger"

	einoindexer "github.com/cloudwego/eino-ext/components/indexer/redis"
	einoretriever "github.com/cloudwego/eino-ext/components/retriever/redis"
	"github.com/redis/go-redis/v9"
)

// 默认参数；与 eino-ext indexer 的默认 DocumentToHashes 行为对齐。
const (
	defaultIndexName   = "wiki_idx"
	defaultKeyPrefix   = "doc:"
	defaultVectorField = "content_vector"
	defaultBatchSize   = 10
	defaultTopK        = 5
	defaultVectorDim   = 768
	defaultStrategy    = string(consts.StrategyFree)
)

// vectorRedis 是 RAG 专用的 Redis 客户端，由 Init 创建，由 Cleanup 关闭。
// FT.SEARCH 需要 RESP2 协议且开启 UnstableResp3。
var vectorRedis *redis.Client

// Init 初始化 Embedder / RediSearch 索引 / Indexer / Retriever，
// 返回组合好的 *model.RAGService 供 controller 调用。
func Init(ctx context.Context, cfg config.Config) (*model.RAGService, error) {
	log := logger.GetLogger()

	rdb, err := newVectorRedis(ctx, cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("new vector redis: %w", err)
	}
	vectorRedis = rdb
	log.Printf("RAG: vector redis 连接成功 %s:%s", cfg.Redis.Host, cfg.Redis.Port)

	emb, err := embedding.NewGeminiEmbedderFromDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("create embedder: %w", err)
	}
	log.Printf("RAG: embedder 初始化成功")

	if err := ensureIndex(ctx, rdb, defaultIndexName, defaultKeyPrefix, defaultVectorDim); err != nil {
		return nil, fmt.Errorf("ensure index: %w", err)
	}
	log.Printf("RAG: 索引 %s 已就绪", defaultIndexName)

	idx, err := einoindexer.NewIndexer(ctx, &einoindexer.IndexerConfig{
		Client:    rdb,
		KeyPrefix: defaultKeyPrefix,
		BatchSize: defaultBatchSize,
		Embedding: emb,
	})
	if err != nil {
		return nil, fmt.Errorf("new indexer: %w", err)
	}

	ret, err := einoretriever.NewRetriever(ctx, &einoretriever.RetrieverConfig{
		Client:      rdb,
		Index:       defaultIndexName,
		VectorField: defaultVectorField,
		TopK:        defaultTopK,
		Embedding:   emb,
	})
	if err != nil {
		return nil, fmt.Errorf("new retriever: %w", err)
	}
	log.Printf("RAG: indexer/retriever 初始化成功")

	return &model.RAGService{
		IngestFunc: func(c context.Context, req model.RAGIngestRequest) (*model.RAGIngestResult, error) {
			return ingest(c, idx, emb, req)
		},
		SearchFunc: func(c context.Context, req model.RAGSearchRequest) (*model.RAGSearchResult, error) {
			return search(c, ret, req)
		},
	}, nil
}

// Cleanup 释放 RAG 持有的 Redis 连接。
func Cleanup() {
	if vectorRedis != nil {
		_ = vectorRedis.Close()
		vectorRedis = nil
		logger.GetLogger().Printf("RAG: vector redis 已释放")
	}
}

// newVectorRedis 创建带 RESP2 + UnstableResp3 的 Redis 客户端。
// FT.SEARCH/FT.CREATE 都依赖该配置。
func newVectorRedis(ctx context.Context, cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		Protocol: 2,
	})
	client.Options().UnstableResp3 = true

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}
	return client, nil
}

// ensureIndex 不存在时创建 RediSearch 向量索引。
// eino-ext 的 Indexer 只写数据不建索引，必须在检索前手动创建。
func ensureIndex(ctx context.Context, rdb *redis.Client, indexName, keyPrefix string, vectorDim int) error {
	if _, err := rdb.FTInfo(ctx, indexName).Result(); err == nil {
		return nil
	}

	prefix := strings.TrimSuffix(keyPrefix, ":")
	return rdb.FTCreate(ctx,
		indexName,
		&redis.FTCreateOptions{
			OnHash: true,
			Prefix: []interface{}{prefix},
		},
		&redis.FieldSchema{
			FieldName: "content",
			FieldType: redis.SearchFieldTypeText,
		},
		&redis.FieldSchema{
			FieldName: defaultVectorField,
			FieldType: redis.SearchFieldTypeVector,
			VectorArgs: &redis.FTVectorArgs{
				FlatOptions: &redis.FTFlatOptions{
					Type:           "FLOAT32",
					Dim:            vectorDim,
					DistanceMetric: "COSINE",
				},
			},
		},
	).Err()
}

// ingest 执行：切块 → 向量化 → Redis 存储。
func ingest(ctx context.Context, idx *einoindexer.Indexer, emb embedding.Embedder, req model.RAGIngestRequest) (*model.RAGIngestResult, error) {
	if strings.TrimSpace(req.Content) == "" {
		return nil, fmt.Errorf("content is empty")
	}

	strategy := req.Strategy
	if strategy == "" {
		strategy = defaultStrategy
	}

	chunker, err := newChunker(consts.Strategy(strategy), emb)
	if err != nil {
		return nil, err
	}

	docs, err := chunker.Chunk(ctx, req.Content, model.ChunkConfig{
		ChunkSize:    req.ChunkSize,
		ChunkOverlap: req.ChunkOverlap,
		Separators:   req.Separators,
	})
	if err != nil {
		return nil, fmt.Errorf("chunk: %w", err)
	}
	if len(docs) == 0 {
		return nil, fmt.Errorf("no chunks produced")
	}

	prefix := req.DocIDPrefix
	if prefix == "" {
		prefix = "doc"
	}
	totalChars := 0
	for _, doc := range docs {
		doc.ID = fmt.Sprintf("%s_%s", prefix, doc.ID)
		totalChars += len([]rune(doc.Content))
	}

	storedIDs, err := idx.Store(ctx, docs)
	if err != nil {
		return nil, fmt.Errorf("indexer store: %w", err)
	}
	return &model.RAGIngestResult{
		StoredIDs:  storedIDs,
		ChunkCount: len(storedIDs),
		TotalChars: totalChars,
	}, nil
}

// search 执行：查询 → 向量化 → Redis 向量检索。
func search(ctx context.Context, ret *einoretriever.Retriever, req model.RAGSearchRequest) (*model.RAGSearchResult, error) {
	if strings.TrimSpace(req.Query) == "" {
		return nil, fmt.Errorf("query is empty")
	}

	docs, err := ret.Retrieve(ctx, req.Query)
	if err != nil {
		return nil, fmt.Errorf("retriever: %w", err)
	}

	results := make([]model.RAGSearchItem, 0, len(docs))
	for _, doc := range docs {
		item := model.RAGSearchItem{
			ID:       doc.ID,
			Content:  doc.Content,
			MetaData: doc.MetaData,
		}
		if s := doc.Score(); s != 0 {
			item.Score = &s
		}
		results = append(results, item)
	}
	return &model.RAGSearchResult{Results: results}, nil
}

// newChunker 按策略名构造 Chunker；语义切块复用同一个 embedder。
func newChunker(strategy consts.Strategy, emb embedding.Embedder) (chunk.Chunker, error) {
	switch strategy {
	case consts.StrategyFree:
		return chunk.NewFreeChunker(), nil
	case consts.StrategyMD:
		return chunk.NewMDChunker(), nil
	case consts.StrategyHierarchical:
		return chunk.NewHierarchicalChunker(), nil
	case consts.StrategyEino:
		if emb == nil {
			return nil, fmt.Errorf("embedder is required for semantic chunking")
		}
		return chunk.NewEinoChunker(emb), nil
	default:
		return nil, fmt.Errorf("unknown chunk strategy: %s", strategy)
	}
}
