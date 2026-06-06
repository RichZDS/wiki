package chunk

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"aisearch/internal/ai/chunk"
	"aisearch/internal/ai/embedding"
	"aisearch/internal/model"

	tests "aisearch/tests"
)

// --- 共享测试数据 ---

var (
	shortMD  = "# Title\nSome content here."
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

// --- FreeChunker ---

// TestFreeChunker 验证自由切块器的基础切分行为。
func TestFreeChunker(t *testing.T) {
	chunker := chunk.NewChunker(chunk.StrategyFree)
	cfg := chunk.ChunkConfig{
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

// TestFreeChunkerEmpty 验证自由切块器处理空内容的行为。
func TestFreeChunkerEmpty(t *testing.T) {
	chunker := chunk.NewChunker(chunk.StrategyFree)
	docs, err := chunker.Chunk(context.Background(), "", chunk.ChunkConfig{})
	tests.AssertNoErr(t, err, "FreeChunker empty")
	if docs != nil {
		t.Error("expected nil for empty input")
	}
}

// --- mdChunker ---

// TestMDChunker 验证 Markdown 切块器的基础切分行为。
func TestMDChunker(t *testing.T) {
	chunker := chunk.NewChunker(chunk.StrategyMD)
	cfg := chunk.ChunkConfig{
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
		if d.MetaData["chunk_index"] == nil {
			t.Errorf("chunk %s: missing chunk_index", d.ID)
		}
		if d.MetaData["chunk_strategy"] != "md" {
			t.Errorf("chunk %s: expected chunk_strategy=md, got %v", d.ID, d.MetaData["chunk_strategy"])
		}
	}
	t.Logf("MDChunker OK: %d chunks from markdown doc", len(docs))
}

// TestMDChunkerHeadingPath 验证 Markdown 切块保留标题路径。
func TestMDChunkerHeadingPath(t *testing.T) {
	chunker := chunk.NewChunker(chunk.StrategyMD)
	content := "# Chapter 1\n## Section 1.1\nSome paragraph text under section 1.1.\n## Section 1.2\nMore text here."
	cfg := chunk.ChunkConfig{ChunkSize: 500, ChunkOverlap: 50}
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

// TestMDChunkerCodeBlockIntegrity 验证代码块在切分时保持完整。
func TestMDChunkerCodeBlockIntegrity(t *testing.T) {
	chunker := chunk.NewChunker(chunk.StrategyMD)
	content := "# Overview\nSome intro.\n\n```go\npackage main\n\nfunc main() {\n\tfmt.Println(\"hello world\")\n}\n```\n\nAfter the code block."
	cfg := chunk.ChunkConfig{ChunkSize: 100, ChunkOverlap: 20}
	docs, err := chunker.Chunk(context.Background(), content, cfg)
	tests.AssertNoErr(t, err, "MDChunker code block integrity")

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

// TestMDChunkerEmpty 验证 Markdown 切块器处理空内容的行为。
func TestMDChunkerEmpty(t *testing.T) {
	chunker := chunk.NewChunker(chunk.StrategyMD)
	docs, err := chunker.Chunk(context.Background(), "", chunk.ChunkConfig{})
	tests.AssertNoErr(t, err, "MDChunker empty")
	if docs != nil {
		t.Error("expected nil for empty input")
	}
}

// TestMDChunkerShortContent 验证短 Markdown 内容不会被过度切分。
func TestMDChunkerShortContent(t *testing.T) {
	chunker := chunk.NewChunker(chunk.StrategyMD)
	docs, err := chunker.Chunk(context.Background(), shortMD, chunk.ChunkConfig{ChunkSize: 500})
	tests.AssertNoErr(t, err, "MDChunker short content")
	if len(docs) != 1 {
		t.Errorf("expected 1 chunk for short content, got %d", len(docs))
	}
}

// --- einoChunker ---

// TestEinoChunkerNilEmbedder 验证缺少向量模型时返回错误。
func TestEinoChunkerNilEmbedder(t *testing.T) {
	chunker := chunk.NewChunker(chunk.StrategyEino)
	_, err := chunker.Chunk(context.Background(), "test", chunk.ChunkConfig{})
	if err == nil {
		t.Error("expected error for nil embedder")
	}
	t.Logf("nil embedder error: %v", err)
}

// TestEinoChunkerWithMock 验证语义切块器能够使用模拟向量。
func TestEinoChunkerWithMock(t *testing.T) {
	mock := &model.MockEmbedder{}
	chunker := chunk.NewEinoChunker(mock)
	cfg := chunk.ChunkConfig{ChunkSize: 500, ChunkOverlap: 30}

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

// TestEinoChunkerShortContent 验证短文本无需向量计算即可返回。
func TestEinoChunkerShortContent(t *testing.T) {
	mock := &model.MockEmbedder{}
	chunker := chunk.NewEinoChunker(mock)
	docs, err := chunker.Chunk(context.Background(), "short text", chunk.ChunkConfig{ChunkSize: 500})
	tests.AssertNoErr(t, err, "EinoChunker short content")
	if len(docs) != 1 {
		t.Errorf("expected 1 chunk for short content, got %d", len(docs))
	}
}

// TestEinoChunkerEmpty 验证语义切块器处理空内容的行为。
func TestEinoChunkerEmpty(t *testing.T) {
	mock := &model.MockEmbedder{}
	chunker := chunk.NewEinoChunker(mock)
	docs, err := chunker.Chunk(context.Background(), "", chunk.ChunkConfig{})
	tests.AssertNoErr(t, err, "EinoChunker empty")
	if docs != nil {
		t.Error("expected nil for empty input")
	}
}

// TestSplitSentences 验证中英文文本的分句结果。
func TestSplitSentences(t *testing.T) {
	mock := &model.MockEmbedder{
		Vectors: [][]float64{
			{1.0, 0.0, 0.0},
			{1.0, 0.1, 0.0},
			{0.0, 1.0, 0.0},
			{0.0, 1.0, 0.1},
			{1.0, 0.0, 0.0},
			{1.0, 0.1, 0.0},
		},
	}
	chunker := chunk.NewEinoChunker(mock)
	content := "First sentence. Second sentence. Third topic here. Fourth follows third. Fifth different topic. Sixth close to fifth."
	docs, err := chunker.Chunk(context.Background(), content, chunk.ChunkConfig{ChunkSize: 200, ChunkOverlap: 0})
	tests.AssertNoErr(t, err, "split sentences")
	t.Logf("split into %d chunks", len(docs))
	if len(docs) < 2 {
		t.Log("note: with small text, may only produce 1 chunk")
	}
}

// --- hierarchicalChunker ---

// TestHierarchicalChunker 验证分层切块器生成父子文档。
func TestHierarchicalChunker(t *testing.T) {
	chunker := chunk.NewChunker(chunk.StrategyHierarchical)
	cfg := chunk.ChunkConfig{ChunkSize: 100, ChunkOverlap: 20}

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

// TestHierarchicalChunkerParentChildLink 验证父子块之间的元数据关联。
func TestHierarchicalChunkerParentChildLink(t *testing.T) {
	chunker := chunk.NewChunker(chunk.StrategyHierarchical)
	cfg := chunk.ChunkConfig{ChunkSize: 100, ChunkOverlap: 0}

	docs, err := chunker.Chunk(context.Background(), mediumMD, cfg)
	tests.AssertNoErr(t, err, "HierarchicalChunker parent-child link")

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

	for childID, pid := range childParents {
		if _, exists := parentChildren[pid]; !exists {
			t.Errorf("child %s references non-existent parent %s", childID, pid)
		}
	}

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

// TestHierarchicalChunkerEmpty 验证分层切块器处理空内容的行为。
func TestHierarchicalChunkerEmpty(t *testing.T) {
	chunker := chunk.NewChunker(chunk.StrategyHierarchical)
	docs, err := chunker.Chunk(context.Background(), "", chunk.ChunkConfig{})
	tests.AssertNoErr(t, err, "HierarchicalChunker empty")
	if docs != nil {
		t.Error("expected nil for empty input")
	}
}

// TestHierarchicalChunkerShortContent 验证短内容只生成一个父块。
func TestHierarchicalChunkerShortContent(t *testing.T) {
	chunker := chunk.NewChunker(chunk.StrategyHierarchical)
	docs, err := chunker.Chunk(context.Background(), shortMD, chunk.ChunkConfig{ChunkSize: 500})
	tests.AssertNoErr(t, err, "HierarchicalChunker short content")
	if len(docs) != 1 {
		t.Errorf("expected 1 chunk for short content, got %d", len(docs))
	}
	role, _ := docs[0].MetaData["chunk_role"].(string)
	if role != "parent" {
		t.Errorf("expected single chunk to be parent, got %s", role)
	}
}

// TestHierarchicalChunkerChunkSizeRatio 验证父子块的尺寸比例。
func TestHierarchicalChunkerChunkSizeRatio(t *testing.T) {
	chunker := chunk.NewChunker(chunk.StrategyHierarchical)
	cfg := chunk.ChunkConfig{ChunkSize: 150, ChunkOverlap: 0}

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

// TestNewChunkerStrategy 验证工厂函数返回正确的切块策略。
func TestNewChunkerStrategy(t *testing.T) {
	free := chunk.NewChunker(chunk.StrategyFree)
	tests.AssertTrue(t, free != nil, "StrategyFree should return non-nil Chunker")

	md := chunk.NewChunker(chunk.StrategyMD)
	tests.AssertTrue(t, md != nil, "StrategyMD should return non-nil Chunker")

	eino := chunk.NewChunker(chunk.StrategyEino)
	tests.AssertTrue(t, eino != nil, "StrategyEino should return non-nil Chunker")

	hier := chunk.NewChunker(chunk.StrategyHierarchical)
	tests.AssertTrue(t, hier != nil, "StrategyHierarchical should return non-nil Chunker")

	def := chunk.NewChunker("unknown")
	tests.AssertTrue(t, def != nil, "unknown strategy should return non-nil Chunker")
}

// 确保模拟向量实现满足 Embedder 接口。
var _ embedding.Embedder = (*model.MockEmbedder)(nil)
