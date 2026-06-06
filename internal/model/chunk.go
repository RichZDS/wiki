package model

import (
	"context"

	"github.com/cloudwego/eino/schema"
)

type ChunkConfig struct {
	ChunkSize    int
	ChunkOverlap int
	Separators   []string
}

type FreeChunker struct {
	ChunkFunc func(context.Context, string, ChunkConfig) ([]*schema.Document, error)
}

// Chunk 执行自由文本切块。
func (c *FreeChunker) Chunk(ctx context.Context, content string, cfg ChunkConfig) ([]*schema.Document, error) {
	return c.ChunkFunc(ctx, content, cfg)
}

type MarkdownChunker struct {
	ChunkFunc func(context.Context, string, ChunkConfig) ([]*schema.Document, error)
}

// Chunk 执行 Markdown 文档切块。
func (c *MarkdownChunker) Chunk(ctx context.Context, content string, cfg ChunkConfig) ([]*schema.Document, error) {
	return c.ChunkFunc(ctx, content, cfg)
}

type EinoChunker struct {
	ChunkFunc func(context.Context, string, ChunkConfig) ([]*schema.Document, error)
}

// Chunk 执行基于向量语义的文档切块。
func (c *EinoChunker) Chunk(ctx context.Context, content string, cfg ChunkConfig) ([]*schema.Document, error) {
	return c.ChunkFunc(ctx, content, cfg)
}

type HierarchicalChunker struct {
	ChunkFunc func(context.Context, string, ChunkConfig) ([]*schema.Document, error)
}

// Chunk 执行父子层级文档切块。
func (c *HierarchicalChunker) Chunk(ctx context.Context, content string, cfg ChunkConfig) ([]*schema.Document, error) {
	return c.ChunkFunc(ctx, content, cfg)
}

type MarkdownElement struct {
	Type        string
	HeadingPath string
	Content     string
	Level       int
}

type MarkdownChunk struct {
	Content      string
	HeadingPath  string
	ElementTypes []string
	Level        int
}

type HeadingEntry struct {
	Text  string
	Level int
}

type HeadingStack struct {
	Stack []HeadingEntry
}
