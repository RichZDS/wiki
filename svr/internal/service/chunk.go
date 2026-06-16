package service

import (
	"context"
	"fmt"
	"os"

	"wiki/internal/ai/chunk"
	"wiki/internal/ai/embedding"
	"wiki/internal/model"
	"wiki/internal/model/consts"

	"github.com/cloudwego/eino/schema"
)

// 类型别名，与 user service 模式保持一致。
type (
	ChunkService        = model.ChunkService
	ChunkRequest        = model.ChunkRequest
	ChunkConfigItem     = model.ChunkConfigItem
	ChunkCompareRequest = model.ChunkCompareRequest
	ChunkResult         = model.ChunkResult
	ChunkItem           = model.ChunkItem
	ChunkStats          = model.ChunkStats
	ChunkStrategyResult = model.ChunkStrategyResult
	ChunkCompareResult  = model.ChunkCompareResult
)

// NewChunkService 创建切块服务并注入业务逻辑函数。
func NewChunkService() *ChunkService {
	return &ChunkService{
		ChunkFunc:   doChunk,
		CompareFunc: doCompare,
	}
}

// doChunk 执行单策略切块的业务逻辑。
func doChunk(ctx context.Context, req ChunkRequest) (*ChunkResult, error) {
	chunker, err := buildChunker(ctx, req.Strategy)
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

	return buildChunkResult(docs), nil
}

// doCompare 执行多策略对比切块的业务逻辑。
func doCompare(ctx context.Context, req ChunkCompareRequest) (*ChunkCompareResult, error) {
	result := &ChunkCompareResult{
		Results: make([]ChunkStrategyResult, 0, len(req.Configs)),
		Errors:  make(map[string]string),
	}

	for _, config := range req.Configs {
		chunker, err := buildChunker(ctx, config.Strategy)
		if err != nil {
			result.Errors[config.Strategy] = err.Error()
			continue
		}

		cfg := model.ChunkConfig{
			ChunkSize:    config.ChunkSize,
			ChunkOverlap: config.ChunkOverlap,
			Separators:   config.Separators,
		}

		docs, err := chunker.Chunk(ctx, req.Content, cfg)
		if err != nil {
			result.Errors[config.Strategy] = err.Error()
			continue
		}

		chunkResult := buildChunkResult(docs)
		result.Results = append(result.Results, ChunkStrategyResult{
			Strategy: config.Strategy,
			Chunks:   chunkResult.Chunks,
			Stats:    chunkResult.Stats,
		})
	}

	return result, nil
}

// buildChunker 根据策略字符串构建对应的 Chunker 实现。
func buildChunker(ctx context.Context, strategy string) (chunk.Chunker, error) {
	switch consts.Strategy(strategy) {
	case consts.StrategyEino:
		emb, err := tryCreateEmbedder(ctx)
		if err != nil {
			return nil, fmt.Errorf("eino strategy requires embedding service: %w", err)
		}
		return chunk.NewEinoChunker(emb), nil
	case consts.StrategyMD:
		return chunk.NewMDChunker(), nil
	case consts.StrategyHierarchical:
		return chunk.NewHierarchicalChunker(), nil
	case consts.StrategyFree:
		return chunk.NewFreeChunker(), nil
	default:
		return nil, fmt.Errorf("unknown chunk strategy: %s", strategy)
	}
}

// tryCreateEmbedder 尝试根据环境变量创建 Embedder 实例。
// 优先使用 ARK，其次 OpenAI。
func tryCreateEmbedder(ctx context.Context) (embedding.Embedder, error) {
	if apiKey := os.Getenv("ARK_API_KEY"); apiKey != "" {
		modelID := os.Getenv("ARK_EMBEDDING_MODEL")
		return embedding.NewArkEmbedder(ctx, apiKey, modelID, "", "")
	}

	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		modelID := os.Getenv("OPENAI_EMBEDDING_MODEL")
		if modelID == "" {
			modelID = "text-embedding-3-small"
		}
		return embedding.NewOpenAIEmbedder(ctx, apiKey, modelID, "", "")
	}

	return nil, fmt.Errorf("no embedding service configured: set ARK_API_KEY or OPENAI_API_KEY")
}

// buildChunkResult 将 Eino Document 切片转换为 API 响应格式并计算统计信息。
func buildChunkResult(docs []*schema.Document) *ChunkResult {
	items := make([]ChunkItem, 0, len(docs))
	totalChars := 0
	minLen := int(^uint(0) >> 1) // max int
	maxLen := 0

	for _, doc := range docs {
		content := doc.Content
		length := len([]rune(content))
		items = append(items, ChunkItem{
			ID:       doc.ID,
			Content:  content,
			Length:   length,
			MetaData: doc.MetaData,
		})
		totalChars += length
		if length < minLen {
			minLen = length
		}
		if length > maxLen {
			maxLen = length
		}
	}

	if len(items) == 0 {
		return &ChunkResult{Chunks: items, Stats: ChunkStats{}}
	}

	return &ChunkResult{
		Chunks: items,
		Stats: ChunkStats{
			TotalChunks:     len(items),
			TotalCharacters: totalChars,
			AvgChunkLength:  float64(totalChars) / float64(len(items)),
			MinChunkLength:  minLen,
			MaxChunkLength:  maxLen,
		},
	}
}
