package chunk

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"unicode/utf8"

	"aisearch/internal/ai/embedding"
	"aisearch/internal/model"

	"github.com/cloudwego/eino/schema"
)

// einoChunker 语义切块器，基于 embedding 向量相似度在语义边界处切分。
//
// 算法步骤：
//  1. 按中英文标点分句
//  2. 批量计算句向量
//  3. 计算相邻句余弦相似度
//  4. 在相似度谷值处切分
//  5. 贪心合并句子至接近 chunkSize
type einoChunker = model.EinoChunker

// NewEinoChunker 创建语义切块器，需注入 Embedder 实现。
func NewEinoChunker(emb embedding.Embedder) *einoChunker {
	return &model.EinoChunker{
		ChunkFunc: func(ctx context.Context, content string, cfg model.ChunkConfig) ([]*schema.Document, error) {
			return einoChunk(ctx, content, cfg, emb)
		},
	}
}

// Chunk 执行语义切块。
func einoChunk(ctx context.Context, content string, cfg ChunkConfig, embedder embedding.Embedder) ([]*schema.Document, error) {
	if embedder == nil {
		return nil, fmt.Errorf("einoChunker: Embedder is required, use NewEinoChunker(embedder) to construct")
	}
	sanitizeConfig(&cfg)
	if len(content) == 0 {
		return nil, nil
	}
	runes := []rune(content)
	if len(runes) <= cfg.ChunkSize {
		return []*schema.Document{newDocument(content, 0, 1)}, nil
	}

	// 1. 分句
	sentences := splitSentences(content)
	if len(sentences) <= 1 {
		return []*schema.Document{newDocument(content, 0, 1)}, nil
	}

	// 2. 计算句向量
	vectors, err := embedder.EmbedStrings(ctx, sentences)
	if err != nil {
		return nil, fmt.Errorf("einoChunker: embed failed: %w", err)
	}
	if len(vectors) != len(sentences) {
		return nil, fmt.Errorf("einoChunker: embed returned %d vectors for %d sentences", len(vectors), len(sentences))
	}

	// 3. 计算相邻相似度
	similarities := make([]float64, len(vectors)-1)
	for i := 0; i < len(vectors)-1; i++ {
		similarities[i] = cosineSimilarity(vectors[i], vectors[i+1])
	}

	// 4. 检测断点
	breakpoints := detectBreakpoints(similarities)

	// 5. 合并句子
	chunks := mergeSentences(sentences, breakpoints, cfg.ChunkSize, cfg.ChunkOverlap)

	// 6. 构造 Document
	docs := make([]*schema.Document, len(chunks))
	for i, c := range chunks {
		docs[i] = &schema.Document{
			ID:      fmt.Sprintf("chunk_%d", i),
			Content: c,
			MetaData: map[string]any{
				metaKeyChunkIndex:    i,
				metaKeyTotalChunks:   len(chunks),
				metaKeyChunkStrategy: "eino",
			},
		}
	}
	return docs, nil
}

// splitSentences 按中英文标点将文本拆分为句子序列。
func splitSentences(text string) []string {
	var sentences []string
	var current strings.Builder
	runes := []rune(text)

	for i := 0; i < len(runes); i++ {
		r := runes[i]
		current.WriteRune(r)

		// 句子结束标点
		isEndPunct := r == '.' || r == '。' || r == '!' || r == '！' || r == '?' || r == '？'
		// 换行符（单独一行视为句子边界）
		isNewline := r == '\n'

		if isEndPunct {
			// 检查后续字符：空格、换行或下一句开头
			if i+1 < len(runes) {
				next := runes[i+1]
				if next == ' ' || next == '\n' || next == '\t' {
					sentences = append(sentences, strings.TrimSpace(current.String()))
					current.Reset()
				}
			} else {
				sentences = append(sentences, strings.TrimSpace(current.String()))
				current.Reset()
			}
		} else if isNewline && current.Len() > 1 {
			// 单独的换行符作为句子边界（前面已有内容）
			prev := runes[i-1]
			if prev == '\n' {
				s := strings.TrimSpace(current.String())
				if len(s) > 0 {
					sentences = append(sentences, s)
				}
				current.Reset()
			}
		}
	}

	// 处理剩余文本
	remaining := strings.TrimSpace(current.String())
	if len(remaining) > 0 {
		sentences = append(sentences, remaining)
	}

	// 过滤掉过短的碎片（< 2 字符），合并到前一句
	if len(sentences) > 1 {
		var filtered []string
		for _, s := range sentences {
			if utf8.RuneCountInString(s) < 2 && len(filtered) > 0 {
				filtered[len(filtered)-1] += s
			} else if utf8.RuneCountInString(s) > 0 {
				filtered = append(filtered, s)
			}
		}
		sentences = filtered
	}

	return sentences
}

// cosineSimilarity 计算两个向量的余弦相似度。
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

// detectBreakpoints 基于相似度序列检测语义断点。
// 使用百分位阈值法：相似度低于第 20 百分位且低于绝对阈值 0.5 的位置作为断点。
func detectBreakpoints(similarities []float64) []int {
	if len(similarities) == 0 {
		return nil
	}

	// 计算第 20 百分位
	sorted := make([]float64, len(similarities))
	copy(sorted, similarities)
	sort.Float64s(sorted)
	p20Idx := int(float64(len(sorted)) * 0.2)
	percentile := sorted[p20Idx]

	threshold := math.Min(percentile, 0.5)

	var breakpoints []int
	for i, sim := range similarities {
		if sim < threshold {
			breakpoints = append(breakpoints, i) // 在句子 i 和 i+1 之间断开
		}
	}
	return breakpoints
}

// mergeSentences 按断点分组，贪心合并句子至接近 chunkSize。
func mergeSentences(sentences []string, breakpoints []int, chunkSize, overlap int) []string {
	if len(sentences) == 0 {
		return nil
	}

	// 构建断点集合
	bpSet := make(map[int]bool, len(breakpoints))
	for _, bp := range breakpoints {
		bpSet[bp] = true
	}

	var chunks []string
	current := make([]string, 0, len(sentences))
	currentLen := 0

	flush := func() {
		if len(current) > 0 {
			chunks = append(chunks, strings.Join(current, ""))
			current = nil
			currentLen = 0
		}
	}

	for i, s := range sentences {
		sl := len([]rune(s))

		// 当前句子位于断点处，且缓冲区已有内容 → 刷新
		if bpSet[i-1] && len(current) > 0 {
			flush()
			// 添加 overlap
			if overlap > 0 && len(chunks) > 0 {
				prev := []rune(chunks[len(chunks)-1])
				if len(prev) > overlap {
					overlapText := string(prev[len(prev)-overlap:])
					current = append(current, overlapText)
					currentLen = len([]rune(overlapText))
				}
			}
		}

		// 加入当前句子会超过 chunkSize → 先 flush
		if currentLen+sl > chunkSize && len(current) > 0 {
			flush()
			if overlap > 0 && len(chunks) > 0 {
				prev := []rune(chunks[len(chunks)-1])
				if len(prev) > overlap {
					overlapText := string(prev[len(prev)-overlap:])
					current = append(current, overlapText)
					currentLen = len([]rune(overlapText))
				}
			}
		}

		// 单个句子超长：强制截断
		if sl > chunkSize {
			flush()
			chunks = append(chunks, forceChunk(s, chunkSize)...)
			continue
		}

		current = append(current, s)
		currentLen += sl
	}

	flush()
	return chunks
}
