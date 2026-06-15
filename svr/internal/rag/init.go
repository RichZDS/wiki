// Package rag 统一管理 RAG 相关组件的初始化：向量搜索 Redis、Embedder、Indexer、Retriever，
// 以及对外的入库/检索工作流。
package rag

import (
	"context"
	"fmt"

	"wiki/internal/ai/embedding"
	"wiki/internal/ai/indexer"
	"wiki/internal/ai/retriever"
	"wiki/internal/config"
	"wiki/internal/model"
	"wiki/pkg/logger"
)

// Init 初始化 RAG 流水线所需的全部组件并返回可用的工作流服务。
// 依次初始化向量搜索 Redis 客户端、Gemini Embedder、Redis Indexer 和 Redis Retriever。
// 调用方应在程序退出时调用 Cleanup() 释放资源。
func Init(ctx context.Context, cfg config.Config) (*model.RAGService, error) {
	log := logger.GetLogger()

	// 1. 向量搜索 Redis 客户端（FT.SEARCH 需要 RESP2 协议）
	retriever.InitVectorRedis(cfg.Redis)
	log.Printf("向量搜索 Redis 连接成功")

	// 2. Gemini Embedder（从数据库 ai_model 表读取 API Key）
	emb, err := embedding.NewGeminiEmbedderFromDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("create gemini embedder: %w", err)
	}
	log.Printf("Gemini Embedder 初始化成功")

	// 3. Redis Indexer
	idx, err := indexer.NewDefaultIndexer(ctx, emb, "doc:", 10)
	if err != nil {
		return nil, fmt.Errorf("create redis indexer: %w", err)
	}
	log.Printf("Redis Indexer 初始化成功 (prefix=doc:, batch=10)")

	// 4. Redis Retriever
	ret, err := retriever.NewDefaultRetriever(ctx, emb, "wiki_idx", "vector_content", 5)
	if err != nil {
		return nil, fmt.Errorf("create redis retriever: %w", err)
	}
	log.Printf("Redis Retriever 初始化成功 (index=wiki_idx, topK=5)")

	ragComp := &model.RAGComponent{
		Indexer:   idx,
		Retriever: ret,
	}

	return NewWorkflow(ragComp, emb), nil
}

// Cleanup 释放 RAG 组件占用的资源。
func Cleanup() {
	retriever.CloseVectorRedis()
	logger.GetLogger().Printf("RAG 组件资源已释放")
}
