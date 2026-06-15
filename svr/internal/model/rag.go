package model

import (
	einoindexer "github.com/cloudwego/eino-ext/components/indexer/redis"
	einoretriever "github.com/cloudwego/eino-ext/components/retriever/redis"
)

// RAGComponent 聚合 RAG 流水线的 Indexer 和 Retriever 组件。
type RAGComponent struct {
	Indexer   *einoindexer.Indexer
	Retriever *einoretriever.Retriever
}
