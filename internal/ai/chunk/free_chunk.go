package chunk

import (
	"context"
	"strings"

	"aisearch/internal/model"

	"github.com/cloudwego/eino/schema"
)

// freeChunker 自由切块器，实现"递归字符分割"算法（Recursive Character Text Splitter）。
//
// 算法源自 LangChain 社区标准，分三步执行：
//  1. 找最佳分隔符 — 按优先级遍历分隔符列表，选一个能把文本切成全部 ≤ chunkSize 片段的
//  2. 递归分割 — 超长片段降级到下一级分隔符继续切，兜底按字符硬截断
//  3. 合并 + 重叠 — 贪心合并小片段至接近 chunkSize，相邻块间用 chunkOverlap 接续
//
// 当 cfg.Separators 为空时，使用默认分隔符优先级：
//
//	"\n\n" → "\n" → "。" → "." → "，" → "," → " " → ""（按字符兜底）
type freeChunker = model.FreeChunker

// 默认分隔符优先级列表。末尾 "" 表示逐字符拆分——任何文本都能被切分。
var defaultSeparators = []string{"\n\n", "\n", "。", ".", "，", ",", " ", ""}

// NewFreeChunker 创建自由文本切块器。
func NewFreeChunker() *freeChunker {
	return &model.FreeChunker{ChunkFunc: freeChunk}
}

// Chunk 执行自由切块。
// 空内容直接返回 (nil, nil)；单块内容直接返回一个 Document。
func freeChunk(ctx context.Context, content string, cfg ChunkConfig) ([]*schema.Document, error) {
	// --- 参数兜底 ---
	seps := cfg.Separators
	if len(seps) == 0 {
		seps = defaultSeparators
	}
	sanitizeConfig(&cfg)

	if len(content) == 0 {
		return nil, nil
	}

	// 短文本直接返回，避免空转分割逻辑
	runes := []rune(content)
	if len(runes) <= cfg.ChunkSize {
		return []*schema.Document{newDocument(content, 0, 1)}, nil
	}

	// --- 核心流程：分割 + 合并 ---
	chunks := splitText(string(runes), seps, cfg.ChunkSize, cfg.ChunkOverlap)

	// --- 封装为 Document ---
	docs := make([]*schema.Document, len(chunks))
	for i, c := range chunks {
		docs[i] = newDocument(c, i, len(chunks))
	}
	return docs, nil
}

// splitText 递归分割入口：找到合适分隔符 → 递归处理超长段 → 合并段为最终块。
// separators: 分隔符优先级列表 chunkSize: 每块最大字符数 overlap: 块间重叠字符数
func splitText(text string, separators []string, chunkSize, overlap int) []string {
	// --- 第一步：找到能切出全部 ≤ chunkSize 片段的分隔符 ---
	var goodSep string      // 最终选定的分隔符
	var goodSplits []string // 按该分隔符切分得到的所有片段

	for i, sep := range separators {
		if sep == "" {
			// 兜底策略：按 rune 逐一拆分，任何字符串都能被切分
			runes := []rune(text)
			splits := make([]string, len(runes))
			for j, r := range runes {
				splits[j] = string(r)
			}
			goodSep = separators[i]
			goodSplits = splits
			break // 兜底策略不再继续循环
		}
		if !strings.Contains(text, sep) {
			continue // 当前分隔符在文本中不存在，跳到下一级
		}
		splits := strings.Split(text, sep)
		goodSep = separators[i]
		goodSplits = splits
		if allFit(splits, chunkSize) {
			break // 找到合适分隔符，所有片段都不超过 chunkSize
		}
	}

	// --- 第二步：递归处理超长片段 ---
	nextSeps := nextSeparators(separators, goodSep)
	var finalSplits []string
	for _, split := range goodSplits {
		if len(split) == 0 {
			continue
		}
		if len([]rune(split)) <= chunkSize {
			finalSplits = append(finalSplits, split)
		} else if len(nextSeps) > 0 {
			// 当前片段仍然超长，用下一级分隔符递归切分
			finalSplits = append(finalSplits, splitText(split, nextSeps, chunkSize, overlap)...)
		} else {
			// 无更多分隔符可用，强制按 chunkSize 截断
			finalSplits = append(finalSplits, forceChunk(split, chunkSize)...)
		}
	}

	// --- 第三步：合并片段并添加块间重叠 ---
	return mergeSplits(finalSplits, chunkSize, overlap)
}

// allFit 判断所有片段长度是否都不超过 chunkSize（以 rune 计数）。
func allFit(splits []string, chunkSize int) bool {
	for _, s := range splits {
		if len([]rune(s)) > chunkSize {
			return false
		}
	}
	return true
}

// nextSeparators 返回当前分隔符之后的剩余分隔符列表（不包含当前）。
// 当分隔符为 "" 时返回 nil，表示没有更细粒度的分隔符可用了。
func nextSeparators(all []string, current string) []string {
	for i, s := range all {
		if s == current && i < len(all)-1 {
			return all[i+1:]
		}
	}
	return nil
}

// forceChunk 强制按 chunkSize 切分文本，适用于无法用分隔符拆分的超长无结构文本。
func forceChunk(s string, chunkSize int) []string {
	runes := []rune(s)
	var parts []string
	for i := 0; i < len(runes); i += chunkSize {
		end := i + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		parts = append(parts, string(runes[i:end]))
	}
	return parts
}

// mergeSplits 将碎片合并为 ≤ chunkSize 的大块，并在相邻块之间添加 overlap。
// 合并策略：
//   - 贪心拼接相邻片段，尽量使每块长度接近 chunkSize
//   - 单片段超过 chunkSize 的，直接强制切分
//   - 每生成一个完整块后，取出上一块尾部的 overlap 个字符拼到当前块头部
func mergeSplits(splits []string, chunkSize, overlap int) []string {
	if len(splits) == 0 {
		return nil
	}

	var chunks []string  // 最终输出
	var current []string // 当前正在拼接的缓冲区
	currentLen := 0      // current 的 rune 总长度

	// flushChunk 将当前缓冲区内容合并为一个块并清空缓冲区。
	flushChunk := func() {
		if len(current) > 0 {
			chunks = append(chunks, strings.Join(current, ""))
			current = nil
			currentLen = 0
		}
	}

	for _, split := range splits {
		sl := len([]rune(split))
		if sl > chunkSize {
			// 单片段超长：先 flush 已累积内容，再强制切这个超长片段
			flushChunk()
			chunks = append(chunks, forceChunk(split, chunkSize)...)
			continue
		}

		if currentLen+sl > chunkSize {
			// 缓冲区已满，生成一个完整的 chunk
			flushChunk()
			// 从上一个 chunk 尾部取 overlap 大小的文本作为当前块的起始
			if overlap > 0 && len(chunks) > 0 {
				prev := []rune(chunks[len(chunks)-1])
				if len(prev) > overlap {
					overlapText := string(prev[len(prev)-overlap:])
					current = append(current, overlapText)
					currentLen = len([]rune(overlapText))
				}
			}
		}

		current = append(current, split)
		currentLen += sl
	}

	flushChunk() // 处理末尾剩余片段
	return chunks
}
