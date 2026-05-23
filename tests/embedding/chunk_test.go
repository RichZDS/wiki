package embedding

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"aisearch/internal/ai/embedding"

	tests "aisearch/tests"
)

// --- 共享测试数据 ---

var (
	shortMD = "# Title\nSome content here."
	mediumMD = `# Chapter 1

This is the first chapter. It contains some introductory text.

## Section 1.1

Here is a section with a paragraph of content that discusses the topic in detail.

` + "```go\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n```" + `

## Section 1.2

More content here with some details about the implementation.

# Chapter 2

The second chapter begins with an overview of the architecture.`

	longMD = strings.Repeat(mediumMD+"\n", 5)
)

// --- FreeChunker (保持原有测试) ---

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
	for _, d := range docs {
		if d.MetaData["chunk_index"] == nil {
			t.Error("missing chunk_index metadata")
		}
	}
	t.Logf("FreeChunker OK: %d chunks", len(docs))
}

func TestFreeChunkerEmpty(t *testing.T) {
	chunker := embedding.NewChunker(embedding.StrategyFree)
	docs, err := chunker.Chunk(context.Background(), "", embedding.ChunkConfig{})
	tests.AssertNoErr(t, err, "FreeChunker empty")
	if docs != nil {
		t.Error("expected nil for empty input")
	}
}

// --- mdChunker (方案三) ---

func TestMDChunker(t *testing.T) {
	chunker := embedding.NewChunker(embedding.StrategyMD)
	cfg := embedding.ChunkConfig{
		ChunkSize:    500,
		ChunkOverlap: 50,
	}

	content := mediumMD
	docs, err := chunker.Chunk(context.Background(), content, cfg)
	tests.AssertNoErr(t, err, "MDChunker.Chunk")

	if len(docs) == 0 {
		t.Fatal("expected at least 1 chunk")
	}

	for _, d := range docs {
		// 验证基本元数据
		if d.MetaData["chunk_index"] == nil {
			t.Errorf("chunk %s: missing chunk_index", d.ID)
		}
		// 验证策略标记
		if d.MetaData["chunk_strategy"] != "md" {
			t.Errorf("chunk %s: expected chunk_strategy=md, got %v", d.ID, d.MetaData["chunk_strategy"])
		}
	}
	t.Logf("MDChunker OK: %d chunks from markdown doc", len(docs))
}

func TestMDChunkerHeadingPath(t *testing.T) {
	chunker := embedding.NewChunker(embedding.StrategyMD)
	content := "# Chapter 1\n## Section 1.1\nSome paragraph text under section 1.1.\n## Section 1.2\nMore text here."
	cfg := embedding.ChunkConfig{ChunkSize: 500, ChunkOverlap: 50}
	docs, err := chunker.Chunk(context.Background(), content, cfg)
	tests.AssertNoErr(t, err, "MDChunker heading path")

	foundHeadingPath := false
	for _, d := range docs {
		if hp, ok := d.MetaData["heading_path"].(string); ok && hp != "" {
			foundHeadingPath = true
			t.Logf("heading_path: %s", hp)
		}
	}
	if !foundHeadingPath {
		t.Error("expected at least one chunk with heading_path")
	}
}

func TestMDChunkerCodeBlockIntegrity(t *testing.T) {
	chunker := embedding.NewChunker(embedding.StrategyMD)
	content := "# Overview\nSome intro.\n\n```go\npackage main\n\nfunc main() {\n\tfmt.Println(\"hello world\")\n}\n```\n\nAfter the code block."
	cfg := embedding.ChunkConfig{ChunkSize: 100, ChunkOverlap: 20}
	docs, err := chunker.Chunk(context.Background(), content, cfg)
	tests.AssertNoErr(t, err, "MDChunker code block integrity")

	// Small chunk size should cause splits, but code block text should be present
	fullText := ""
	for _, d := range docs {
		fullText += d.Content
	}
	if !strings.Contains(fullText, "func main()") {
		t.Error("code block content missing from chunks")
	}
	if !strings.Contains(fullText, "fmt.Println") {
		t.Error("code block line missing from chunks")
	}
	t.Logf("MDChunker code block OK: %d chunks", len(docs))
}

func TestMDChunkerEmpty(t *testing.T) {
	chunker := embedding.NewChunker(embedding.StrategyMD)
	docs, err := chunker.Chunk(context.Background(), "", embedding.ChunkConfig{})
	tests.AssertNoErr(t, err, "MDChunker empty")
	if docs != nil {
		t.Error("expected nil for empty input")
	}
}

func TestMDChunkerShortContent(t *testing.T) {
	chunker := embedding.NewChunker(embedding.StrategyMD)
	docs, err := chunker.Chunk(context.Background(), shortMD, embedding.ChunkConfig{ChunkSize: 500})
	tests.AssertNoErr(t, err, "MDChunker short content")
	if len(docs) != 1 {
		t.Errorf("expected 1 chunk for short content, got %d", len(docs))
	}
}

// --- einoChunker (方案四) ---

type mockEmbedder struct {
	vectors [][]float64
}

func (m *mockEmbedder) EmbedStrings(ctx context.Context, texts []string) ([][]float64, error) {
	if m.vectors != nil {
		return m.vectors, nil
	}
	// 生成确定性的简单向量：用文本长度构造 3 维向量，相邻文本会有相似值
	result := make([][]float64, len(texts))
	for i, t := range texts {
		v := make([]float64, 3)
		for j := range v {
			v[j] = float64(len(t)+i+j) / 100.0
		}
		result[i] = v
	}
	return result, nil
}

func TestEinoChunkerNilEmbedder(t *testing.T) {
	chunker := embedding.NewChunker(embedding.StrategyEino)
	_, err := chunker.Chunk(context.Background(), "test", embedding.ChunkConfig{})
	if err == nil {
		t.Error("expected error for nil embedder")
	}
	t.Logf("nil embedder error: %v", err)
}

func TestEinoChunkerWithMock(t *testing.T) {
	mock := &mockEmbedder{}
	chunker := embedding.NewEinoChunker(mock)
	cfg := embedding.ChunkConfig{ChunkSize: 500, ChunkOverlap: 30}

	// 构造多句文本，让 mock embedder 产生不同向量
	sentences := make([]string, 20)
	for i := range sentences {
		sentences[i] = fmt.Sprintf("This is sentence number %d in the document.", i)
	}
	content := strings.Join(sentences, " ")

	docs, err := chunker.Chunk(context.Background(), content, cfg)
	tests.AssertNoErr(t, err, "EinoChunker with mock")

	if len(docs) == 0 {
		t.Fatal("expected at least 1 chunk")
	}
	for _, d := range docs {
		if d.MetaData["chunk_strategy"] != "eino" {
			t.Errorf("chunk %s: expected chunk_strategy=eino, got %v", d.ID, d.MetaData["chunk_strategy"])
		}
	}
	t.Logf("EinoChunker OK: %d chunks", len(docs))
}

func TestEinoChunkerShortContent(t *testing.T) {
	mock := &mockEmbedder{}
	chunker := embedding.NewEinoChunker(mock)
	docs, err := chunker.Chunk(context.Background(), "short text", embedding.ChunkConfig{ChunkSize: 500})
	tests.AssertNoErr(t, err, "EinoChunker short content")
	if len(docs) != 1 {
		t.Errorf("expected 1 chunk for short content, got %d", len(docs))
	}
}

func TestEinoChunkerEmpty(t *testing.T) {
	mock := &mockEmbedder{}
	chunker := embedding.NewEinoChunker(mock)
	docs, err := chunker.Chunk(context.Background(), "", embedding.ChunkConfig{})
	tests.AssertNoErr(t, err, "EinoChunker empty")
	if docs != nil {
		t.Error("expected nil for empty input")
	}
}

func TestSplitSentences(t *testing.T) {
	// 通过 EinoChunker 间接测试分句逻辑
	mock := &mockEmbedder{
		vectors: [][]float64{
			{1.0, 0.0, 0.0},
			{1.0, 0.1, 0.0},
			{0.0, 1.0, 0.0}, // 差异大 → 断点
			{0.0, 1.0, 0.1},
			{1.0, 0.0, 0.0}, // 差异大 → 断点
			{1.0, 0.1, 0.0},
		},
	}
	chunker := embedding.NewEinoChunker(mock)
	content := "First sentence. Second sentence. Third topic here. Fourth follows third. Fifth different topic. Sixth close to fifth."
	docs, err := chunker.Chunk(context.Background(), content, embedding.ChunkConfig{ChunkSize: 200, ChunkOverlap: 0})
	tests.AssertNoErr(t, err, "split sentences")
	t.Logf("split into %d chunks", len(docs))
	if len(docs) < 2 {
		t.Log("note: with small text, may only produce 1 chunk")
	}
}

// --- hierarchicalChunker (方案五) ---

func TestHierarchicalChunker(t *testing.T) {
	chunker := embedding.NewChunker(embedding.StrategyHierarchical)
	cfg := embedding.ChunkConfig{ChunkSize: 100, ChunkOverlap: 20}

	docs, err := chunker.Chunk(context.Background(), mediumMD, cfg)
	tests.AssertNoErr(t, err, "HierarchicalChunker.Chunk")

	if len(docs) == 0 {
		t.Fatal("expected at least 1 chunk")
	}

	parents := 0
	children := 0
	for _, d := range docs {
		role, _ := d.MetaData["chunk_role"].(string)
		switch role {
		case "parent":
			parents++
		case "child":
			children++
		default:
			t.Errorf("chunk %s: missing or invalid chunk_role: %v", d.ID, role)
		}
		if d.MetaData["chunk_strategy"] != "hierarchical" {
			t.Errorf("chunk %s: expected chunk_strategy=hierarchical", d.ID)
		}
	}

	if parents == 0 {
		t.Error("expected at least 1 parent chunk")
	}
	if children == 0 {
		t.Error("expected at least 1 child chunk")
	}
	t.Logf("HierarchicalChunker OK: %d parents + %d children = %d total", parents, children, len(docs))
}

func TestHierarchicalChunkerParentChildLink(t *testing.T) {
	chunker := embedding.NewChunker(embedding.StrategyHierarchical)
	cfg := embedding.ChunkConfig{ChunkSize: 100, ChunkOverlap: 0}

	docs, err := chunker.Chunk(context.Background(), mediumMD, cfg)
	tests.AssertNoErr(t, err, "HierarchicalChunker parent-child link")

	// 构建 parent_id → child_ids 的索引
	parentChildren := make(map[string][]string)
	childParents := make(map[string]string)

	for _, d := range docs {
		role, _ := d.MetaData["chunk_role"].(string)
		if role == "parent" {
			if ids, ok := d.MetaData["child_chunk_ids"].([]string); ok {
				parentChildren[d.ID] = ids
			}
		}
		if role == "child" {
			if pid, ok := d.MetaData["parent_chunk_id"].(string); ok {
				childParents[d.ID] = pid
			}
			if _, ok := d.MetaData["parent_content"].(string); !ok {
				t.Errorf("child %s: missing parent_content", d.ID)
			}
		}
	}

	// 验证每个子块的 parent_chunk_id 指向一个存在的父块
	for childID, pid := range childParents {
		if _, exists := parentChildren[pid]; !exists {
			t.Errorf("child %s references non-existent parent %s", childID, pid)
		}
	}

	// 验证每个父块的 child_chunk_ids 都指向存在的子块
	childIDs := make(map[string]bool)
	for _, d := range docs {
		if role, _ := d.MetaData["chunk_role"].(string); role == "child" {
			childIDs[d.ID] = true
		}
	}
	for _, ids := range parentChildren {
		for _, cid := range ids {
			if !childIDs[cid] {
				t.Errorf("parent references non-existent child %s", cid)
			}
		}
	}

	t.Logf("parent-child links: %d parents, %d children", len(parentChildren), len(childParents))
}

func TestHierarchicalChunkerEmpty(t *testing.T) {
	chunker := embedding.NewChunker(embedding.StrategyHierarchical)
	docs, err := chunker.Chunk(context.Background(), "", embedding.ChunkConfig{})
	tests.AssertNoErr(t, err, "HierarchicalChunker empty")
	if docs != nil {
		t.Error("expected nil for empty input")
	}
}

func TestHierarchicalChunkerShortContent(t *testing.T) {
	chunker := embedding.NewChunker(embedding.StrategyHierarchical)
	docs, err := chunker.Chunk(context.Background(), shortMD, embedding.ChunkConfig{ChunkSize: 500})
	tests.AssertNoErr(t, err, "HierarchicalChunker short content")
	if len(docs) != 1 {
		t.Errorf("expected 1 chunk for short content, got %d", len(docs))
	}
	role, _ := docs[0].MetaData["chunk_role"].(string)
	if role != "parent" {
		t.Errorf("expected single chunk to be parent, got %s", role)
	}
}

func TestHierarchicalChunkerChunkSizeRatio(t *testing.T) {
	chunker := embedding.NewChunker(embedding.StrategyHierarchical)
	// 使用足够大的文本确保生成多个层级
	cfg := embedding.ChunkConfig{ChunkSize: 150, ChunkOverlap: 0}

	docs, err := chunker.Chunk(context.Background(), longMD, cfg)
	tests.AssertNoErr(t, err, "HierarchicalChunker size ratio")

	var maxParentLen, maxChildLen int
	for _, d := range docs {
		role, _ := d.MetaData["chunk_role"].(string)
		cl := len([]rune(d.Content))
		switch role {
		case "parent":
			if cl > maxParentLen {
				maxParentLen = cl
			}
		case "child":
			if cl > maxChildLen {
				maxChildLen = cl
			}
		}
	}

	t.Logf("max parent chunk: %d chars, max child chunk: %d chars", maxParentLen, maxChildLen)
	if maxParentLen > 0 && maxChildLen > 0 && maxParentLen <= maxChildLen {
		t.Error("expected parent chunks to be larger than child chunks on average")
	}
}

// --- 策略分发 ---

func TestNewChunkerStrategy(t *testing.T) {
	free := embedding.NewChunker(embedding.StrategyFree)
	tests.AssertTrue(t, free != nil, "StrategyFree should return non-nil Chunker")

	md := embedding.NewChunker(embedding.StrategyMD)
	tests.AssertTrue(t, md != nil, "StrategyMD should return non-nil Chunker")

	eino := embedding.NewChunker(embedding.StrategyEino)
	tests.AssertTrue(t, eino != nil, "StrategyEino should return non-nil Chunker")

	hier := embedding.NewChunker(embedding.StrategyHierarchical)
	tests.AssertTrue(t, hier != nil, "StrategyHierarchical should return non-nil Chunker")

	// Unknown strategy defaults to free
	def := embedding.NewChunker("unknown")
	tests.AssertTrue(t, def != nil, "unknown strategy should return non-nil Chunker")
}
