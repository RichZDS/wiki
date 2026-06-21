package rag

import (
	"testing"

	"wiki/internal/model"

	"github.com/cloudwego/eino/schema"
)

// TestNormalizeRAGConfig 验证 RAG 配置默认值会完整补齐。
func TestNormalizeRAGConfig(t *testing.T) {
	cfg := normalizeRAGConfig(model.RAGConfig{})
	if cfg.IndexName != defaultIndexName {
		t.Fatalf("IndexName got %q, want %q", cfg.IndexName, defaultIndexName)
	}
	if cfg.VectorField != defaultVectorField {
		t.Fatalf("VectorField got %q, want %q", cfg.VectorField, defaultVectorField)
	}
	if cfg.VectorIndexType != defaultIndexType {
		t.Fatalf("VectorIndexType got %q, want %q", cfg.VectorIndexType, defaultIndexType)
	}
	if cfg.DefaultTopK != defaultTopK || cfg.MaxTopK != defaultMaxTopK {
		t.Fatalf("topK defaults got default=%d max=%d", cfg.DefaultTopK, cfg.MaxTopK)
	}
}

// TestNormalizeTopK 验证请求级 top_k 会落在配置允许范围内。
func TestNormalizeTopK(t *testing.T) {
	cfg := model.RAGConfig{DefaultTopK: 5, MaxTopK: 20}
	if got := normalizeTopK(0, cfg); got != 5 {
		t.Fatalf("zero topK got %d, want 5", got)
	}
	if got := normalizeTopK(12, cfg); got != 12 {
		t.Fatalf("valid topK got %d, want 12", got)
	}
	if got := normalizeTopK(99, cfg); got != 20 {
		t.Fatalf("oversized topK got %d, want 20", got)
	}
}

// TestRelevanceScoreFromDistance 验证 Redis distance 能转换为相似度分数。
func TestRelevanceScoreFromDistance(t *testing.T) {
	doc := &schema.Document{MetaData: map[string]any{"distance": "0.25"}}
	score, ok := relevanceScore(doc)
	if !ok {
		t.Fatal("expected score")
	}
	if score != 0.75 {
		t.Fatalf("score got %v, want 0.75", score)
	}
}
