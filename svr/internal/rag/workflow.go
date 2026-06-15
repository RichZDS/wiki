package rag

import (
	"context"
	"fmt"
	"strings"

	"wiki/internal/ai/chunk"
	"wiki/internal/ai/embedding"
	"wiki/internal/model"
	"wiki/internal/model/consts"
)

// 流水线常量。
const (
	defaultStrategy = "free"
)

// NewWorkflow 为 RAGComponent 创建完整的入库+检索工作流服务。
// emb 同时用于语义切块和 Indexer/Retriever 的向量化（二者已在 Init 时注入）。
func NewWorkflow(rag *model.RAGComponent, emb *embedding.GeminiEmbedder) *model.RAGService {
	return &model.RAGService{
		IngestFunc: func(ctx context.Context, req model.RAGIngestRequest) (*model.RAGIngestResult, error) {
			return ingest(ctx, rag, emb, req)
		},
		SearchFunc: func(ctx context.Context, req model.RAGSearchRequest) (*model.RAGSearchResult, error) {
			return search(ctx, rag, req)
		},
	}
}

// ingest 执行完整入库流程：切块 → 向量化 → Redis 存储。
func ingest(ctx context.Context, rag *model.RAGComponent, emb *embedding.GeminiEmbedder, req model.RAGIngestRequest) (*model.RAGIngestResult, error) {
	if strings.TrimSpace(req.Content) == "" {
		return nil, fmt.Errorf("content is empty")
	}

	strategy := req.Strategy
	if strategy == "" {
		strategy = defaultStrategy
	}

	chunker, err := buildChunker(strategy, emb)
	if err != nil {
		return nil, err
	}

	cfg := model.ChunkConfig{
		ChunkSize:    req.ChunkSize,
		ChunkOverlap: req.ChunkOverlap,
		Separators:   req.Separators,
	}

	docs, err := chunker.Chunk(ctx, req.Content, cfg)
	if err != nil {
		return nil, fmt.Errorf("chunk failed: %w", err)
	}
	if len(docs) == 0 {
		return nil, fmt.Errorf("no chunks produced")
	}

	// 为每个 chunk 分配业务 ID，便于按文档溯源
	prefix := req.DocIDPrefix
	if prefix == "" {
		prefix = "doc"
	}
	totalChars := 0
	for _, doc := range docs {
		doc.ID = fmt.Sprintf("%s_%s", prefix, doc.ID)
		totalChars += len([]rune(doc.Content))
	}

	storedIDs, err := rag.Indexer.Store(ctx, docs)
	if err != nil {
		return nil, fmt.Errorf("indexer store failed: %w", err)
	}

	return &model.RAGIngestResult{
		StoredIDs:  storedIDs,
		ChunkCount: len(storedIDs),
		TotalChars: totalChars,
	}, nil
}

// search 执行语义检索流程：查询 → 向量化 → Redis FT.SEARCH。
func search(ctx context.Context, rag *model.RAGComponent, req model.RAGSearchRequest) (*model.RAGSearchResult, error) {
	if strings.TrimSpace(req.Query) == "" {
		return nil, fmt.Errorf("query is empty")
	}

	docs, err := rag.Retriever.Retrieve(ctx, req.Query)
	if err != nil {
		return nil, fmt.Errorf("retriever search failed: %w", err)
	}

	results := make([]model.RAGSearchItem, 0, len(docs))
	for _, doc := range docs {
		item := model.RAGSearchItem{
			ID:       doc.ID,
			Content:  doc.Content,
			MetaData: doc.MetaData,
		}
		// Score 由 Retriever 写入 MetaData["_score"]
		if s := doc.Score(); s != 0 {
			item.Score = &s
		}
		results = append(results, item)
	}

	return &model.RAGSearchResult{Results: results}, nil
}

// buildChunker 根据策略名构造对应的 Chunker 实现。
// 语义切块（eino）需要 embedder，其他策略不需要。
func buildChunker(strategy string, emb *embedding.GeminiEmbedder) (chunk.Chunker, error) {
	switch consts.Strategy(strategy) {
	case consts.StrategyFree:
		return chunk.NewFreeChunker(), nil
	case consts.StrategyMD:
		return chunk.NewMDChunker(), nil
	case consts.StrategyEino:
		if emb == nil {
			return nil, fmt.Errorf("embedder is required for semantic chunking")
		}
		// GeminiEmbedder 实现了 eino 的 embedding.Embedder 接口（带 ...Option），
		// 本地 Chunker 接口不支持 Option 参数，通过适配器转换。
		return chunk.NewEinoChunker(&embedderAdapter{emb: emb}), nil
	case consts.StrategyHierarchical:
		return chunk.NewHierarchicalChunker(), nil
	default:
		return nil, fmt.Errorf("unknown chunk strategy: %s", strategy)
	}
}

// embedderAdapter 将 eino embedding.Embedder（带 ...Option）适配到本地 Embedder 接口（不带 Option）。
type embedderAdapter struct {
	emb *embedding.GeminiEmbedder
}

// EmbedStrings 实现本地 embedding.Embedder 接口，丢弃 eino 的 Option 参数。
func (a *embedderAdapter) EmbedStrings(ctx context.Context, texts []string) ([][]float64, error) {
	return a.emb.EmbedStrings(ctx, texts)
}
