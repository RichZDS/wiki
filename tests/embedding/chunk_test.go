package embedding

import (
	"context"
	"strings"
	"testing"

	"aisearch/internal/ai/embedding"

	tests "aisearch/tests"
)

func TestFreeChunker(t *testing.T) {
	chunker := embedding.NewChunker(embedding.StrategyFree)
	cfg := embedding.ChunkConfig{
		ChunkSize:    100,
		ChunkOverlap: 20,
		Separators:   []string{"\n\n", "\n", "。", ".", " "},
	}

	content := strings.Repeat("hello world\n", 50)
	docs, err := chunker.Chunk(context.Background(), content, cfg)
	tests.AssertNoErr(t, err, "FreeChunker.Chunk")

	if len(docs) == 0 {
		t.Fatal("expected at least 1 chunk")
	}
	// Each chunk should have metadata
	for _, d := range docs {
		if d.MetaData["chunk_index"] == nil {
			t.Error("missing chunk_index metadata")
		}
	}
	t.Logf("FreeChunker OK: %d chunks", len(docs))
}

func TestMDChunker(t *testing.T) {
	t.Skip("mdChunker.Chunk not yet implemented")
	chunker := embedding.NewChunker(embedding.StrategyMD)
	cfg := embedding.ChunkConfig{
		ChunkSize:    500,
		ChunkOverlap: 50,
	}

	content := "# Title\n## Section 1\nSome content here.\n## Section 2\nMore content."
	chunks, err := chunker.Chunk(context.Background(), content, cfg)
	tests.AssertNoErr(t, err, "MDChunker.Chunk")

	if len(chunks) == 0 {
		t.Fatal("expected at least 1 chunk")
	}
	t.Logf("MDChunker OK: %d chunks", len(chunks))
}

func TestEinoChunker(t *testing.T) {
	t.Skip("einoChunker.Chunk not yet implemented")
	chunker := embedding.NewChunker(embedding.StrategyEino)
	cfg := embedding.ChunkConfig{
		ChunkSize:    500,
		ChunkOverlap: 50,
	}

	content := "This content will be split semantically."
	chunks, err := chunker.Chunk(context.Background(), content, cfg)
	tests.AssertNoErr(t, err, "EinoChunker.Chunk")

	t.Logf("EinoChunker OK: %d chunks", len(chunks))
}

func TestNewChunkerStrategy(t *testing.T) {
	free := embedding.NewChunker(embedding.StrategyFree)
	tests.AssertTrue(t, free != nil, "StrategyFree should return non-nil Chunker")

	md := embedding.NewChunker(embedding.StrategyMD)
	tests.AssertTrue(t, md != nil, "StrategyMD should return non-nil Chunker")

	eino := embedding.NewChunker(embedding.StrategyEino)
	tests.AssertTrue(t, eino != nil, "StrategyEino should return non-nil Chunker")

	// Unknown strategy defaults to free
	def := embedding.NewChunker("unknown")
	tests.AssertTrue(t, def != nil, "unknown strategy should return non-nil Chunker")
}
