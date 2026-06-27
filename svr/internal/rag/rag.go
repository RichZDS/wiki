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
	"strconv"
	"strings"
	"time"

	"wiki/internal/ai/chunk"
	"wiki/internal/ai/embedding"
	"wiki/internal/config"
	"wiki/internal/model"
	"wiki/internal/model/consts"
	"wiki/pkg/logger"

	einoindexer "github.com/cloudwego/eino-ext/components/indexer/redis"
	einoretriever "github.com/cloudwego/eino-ext/components/retriever/redis"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	"github.com/redis/go-redis/v9"
)

// 默认参数；与 eino-ext indexer 的默认 DocumentToHashes 行为对齐。
const (
	defaultIndexName   = "wiki_vector_idx"
	defaultKeyPrefix   = "doc:"
	defaultVectorField = "vector_content"
	defaultBatchSize   = 10
	defaultTopK        = 5
	defaultMaxTopK     = 50
	defaultVectorDim   = 768
	defaultIndexType   = "hnsw"
	defaultHNSWM       = 16
	defaultHNSWEFBuild = 200
	defaultHNSWEFQuery = 10
	defaultStrategy    = string(consts.StrategyFree)
)

// vectorRedis 是 RAG 专用的 Redis 客户端，由 Init 创建，由 Cleanup 关闭。
// FT.SEARCH 需要 RESP2 协议且开启 UnstableResp3。
var vectorRedis *redis.Client

// Init 初始化 Embedder / RediSearch 索引 / Indexer / Retriever，
// 返回组合好的 *model.RAGService 供 controller 调用。
func Init(ctx context.Context, cfg config.Config) (*model.RAGService, error) {
	log := logger.GetLogger()
	ragCfg := normalizeRAGConfig(cfg.RAG)

	rdb, err := newVectorRedis(ctx, cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("new vector redis: %w", err)
	}
	vectorRedis = rdb
	log.Printf("RAG: vector redis 连接成功 %s:%s", cfg.Redis.Host, cfg.Redis.Port)

	emb, err := embedding.NewEmbedderFromDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("create embedder: %w", err)
	}
	log.Printf("RAG: embedder 初始化成功")

	if err := ensureIndex(ctx, rdb, ragCfg); err != nil {
		return nil, fmt.Errorf("ensure index: %w", err)
	}
	log.Printf("RAG: 索引 %s 已就绪 (type=%s, field=%s)", ragCfg.IndexName, ragCfg.VectorIndexType, ragCfg.VectorField)

	idx, err := einoindexer.NewIndexer(ctx, &einoindexer.IndexerConfig{
		Client:    rdb,
		KeyPrefix: ragCfg.KeyPrefix,
		BatchSize: ragCfg.BatchSize,
		Embedding: emb,
	})
	if err != nil {
		return nil, fmt.Errorf("new indexer: %w", err)
	}

	ret, err := einoretriever.NewRetriever(ctx, &einoretriever.RetrieverConfig{
		Client:      rdb,
		Index:       ragCfg.IndexName,
		VectorField: ragCfg.VectorField,
		TopK:        ragCfg.DefaultTopK,
		ReturnFields: []string{
			"content",
			"distance",
		},
		Embedding: emb,
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
			return search(c, ret, ragCfg, req)
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
func ensureIndex(ctx context.Context, rdb *redis.Client, cfg model.RAGConfig) error {
	if _, err := rdb.FTInfo(ctx, cfg.IndexName).Result(); err == nil {
		return nil
	}

	return rdb.FTCreate(ctx,
		cfg.IndexName,
		&redis.FTCreateOptions{
			OnHash: true,
			Prefix: []interface{}{cfg.KeyPrefix},
		},
		&redis.FieldSchema{
			FieldName: "content",
			FieldType: redis.SearchFieldTypeText,
		},
		&redis.FieldSchema{
			FieldName:  cfg.VectorField,
			FieldType:  redis.SearchFieldTypeVector,
			VectorArgs: buildVectorArgs(cfg),
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
func search(ctx context.Context, ret *einoretriever.Retriever, cfg model.RAGConfig, req model.RAGSearchRequest) (*model.RAGSearchResult, error) {
	if strings.TrimSpace(req.Query) == "" {
		return nil, fmt.Errorf("query is empty")
	}

	topK := normalizeTopK(req.TopK, cfg)
	startedAt := time.Now()
	docs, err := ret.Retrieve(ctx, req.Query, retriever.WithTopK(topK))
	if err != nil {
		return nil, fmt.Errorf("retriever: %w", err)
	}

	results := make([]model.RAGSearchItem, 0, len(docs))
	for _, doc := range docs {
		item := model.RAGSearchItem{
			ID:       doc.ID,
			Content:  doc.Content,
			MetaData: sanitizeMetaData(doc.MetaData),
		}
		if score, ok := relevanceScore(doc); ok {
			item.Score = &score
		}
		results = append(results, item)
	}
	return &model.RAGSearchResult{
		Results:    results,
		TopK:       topK,
		DurationMS: time.Since(startedAt).Milliseconds(),
	}, nil
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

// normalizeRAGConfig 合并 YAML 配置和 RAG 默认值。
func normalizeRAGConfig(cfg model.RAGConfig) model.RAGConfig {
	if cfg.IndexName == "" {
		cfg.IndexName = defaultIndexName
	}
	if cfg.KeyPrefix == "" {
		cfg.KeyPrefix = defaultKeyPrefix
	}
	if cfg.VectorField == "" {
		cfg.VectorField = defaultVectorField
	}
	if cfg.VectorDim <= 0 {
		cfg.VectorDim = defaultVectorDim
	}
	if cfg.VectorIndexType == "" {
		cfg.VectorIndexType = defaultIndexType
	}
	cfg.VectorIndexType = strings.ToLower(cfg.VectorIndexType)
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = defaultBatchSize
	}
	if cfg.DefaultTopK <= 0 {
		cfg.DefaultTopK = defaultTopK
	}
	if cfg.MaxTopK <= 0 {
		cfg.MaxTopK = defaultMaxTopK
	}
	if cfg.HNSWMaxEdgesPerNode <= 0 {
		cfg.HNSWMaxEdgesPerNode = defaultHNSWM
	}
	if cfg.HNSWEFConstruction <= 0 {
		cfg.HNSWEFConstruction = defaultHNSWEFBuild
	}
	if cfg.HNSWEFRuntime <= 0 {
		cfg.HNSWEFRuntime = defaultHNSWEFQuery
	}
	return cfg
}

// buildVectorArgs 根据配置构造 Redis 向量索引参数。
func buildVectorArgs(cfg model.RAGConfig) *redis.FTVectorArgs {
	if cfg.VectorIndexType == "flat" {
		return &redis.FTVectorArgs{
			FlatOptions: &redis.FTFlatOptions{
				Type:           "FLOAT32",
				Dim:            cfg.VectorDim,
				DistanceMetric: "COSINE",
			},
		}
	}

	return &redis.FTVectorArgs{
		HNSWOptions: &redis.FTHNSWOptions{
			Type:                   "FLOAT32",
			Dim:                    cfg.VectorDim,
			DistanceMetric:         "COSINE",
			MaxEdgesPerNode:        cfg.HNSWMaxEdgesPerNode,
			MaxAllowedEdgesPerNode: cfg.HNSWEFConstruction,
			EFRunTime:              cfg.HNSWEFRuntime,
		},
	}
}

// normalizeTopK 将请求中的 top_k 限制在配置允许范围内。
func normalizeTopK(requestTopK int, cfg model.RAGConfig) int {
	if requestTopK <= 0 {
		return cfg.DefaultTopK
	}
	if requestTopK > cfg.MaxTopK {
		return cfg.MaxTopK
	}
	return requestTopK
}

// sanitizeMetaData 去掉不需要返回给前端的大字段。
func sanitizeMetaData(meta map[string]any) map[string]any {
	if len(meta) == 0 {
		return meta
	}
	clean := make(map[string]any, len(meta))
	for key, value := range meta {
		if key == "_dense_vector" || key == defaultVectorField {
			continue
		}
		clean[key] = value
	}
	return clean
}

// relevanceScore 将 Redis 的 cosine distance 转为越大越相关的分数。
func relevanceScore(doc *schema.Document) (float64, bool) {
	if score := doc.Score(); score != 0 {
		return score, true
	}
	raw, ok := doc.MetaData["distance"]
	if !ok {
		return 0, false
	}
	distance, ok := parseFloat(raw)
	if !ok {
		return 0, false
	}
	score := 1 - distance
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}
	return score, true
}

// parseFloat 兼容 Redis 返回的字符串或数值类型。
func parseFloat(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case string:
		f, err := strconv.ParseFloat(v, 64)
		return f, err == nil
	default:
		return 0, false
	}
}
