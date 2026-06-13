package chunk

import (
	"context"
	"fmt"
	"strings"

	"wiki/internal/model"

	"github.com/cloudwego/eino/schema"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// mdChunker 按元素分类 + 标题聚合方式切分 Markdown（对应 Unstructured.io chunk_by_title 思路）。
//
// 核心步骤：
//  1. 用 goldmark 解析 AST，收集元素列表（类型、标题路径、原始文本）
//  2. 以标题为 section 边界，同 section 内贪心聚合至接近 chunkSize
//  3. 代码块、列表、引用块等原子元素不被内部切分
//  4. 超长 section 用 freeChunker 兜底
type mdChunker = model.MarkdownChunker

// NewMDChunker 返回 Markdown 元素分类切块器。
func NewMDChunker() *mdChunker {
	return &model.MarkdownChunker{ChunkFunc: mdChunk}
}

type mdElement = model.MarkdownElement
type markdownChunk = model.MarkdownChunk
type headingEntry = model.HeadingEntry
type headingStack = model.HeadingStack

// Chunk 执行 Markdown 元素分类切块。
func mdChunk(ctx context.Context, content string, cfg ChunkConfig) ([]*schema.Document, error) {
	sanitizeConfig(&cfg)
	if len(content) == 0 {
		return nil, nil
	}

	elements := parseElements(content)
	if len(elements) == 0 {
		return []*schema.Document{{
			ID:      fmt.Sprintf("chunk_0"),
			Content: content,
			MetaData: map[string]any{
				metaKeyChunkIndex:    0,
				metaKeyTotalChunks:   1,
				metaKeyChunkStrategy: "md",
			},
		}}, nil
	}

	chunks := aggregateElements(elements, cfg)
	docs := make([]*schema.Document, len(chunks))
	for i, c := range chunks {
		docs[i] = markdownChunkToDocument(&c, i, len(chunks))
	}
	return docs, nil
}

// parseElements 用 goldmark 解析 Markdown，收集所有可切块元素。
func parseElements(content string) []mdElement {
	source := []byte(content)
	parser := goldmark.DefaultParser()
	root := parser.Parse(text.NewReader(source))

	var elements []mdElement
	headingStack := newHeadingStack()

	ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch n.Kind() {
		case ast.KindHeading:
			h := n.(*ast.Heading)
			pushHeading(headingStack, headingEntry{
				Text:  extractText(source, n),
				Level: h.Level,
			})

		case ast.KindParagraph:
			if n.Parent() != nil && n.Parent().Kind() == ast.KindListItem {
				return ast.WalkContinue, nil
			}
			elements = append(elements, mdElement{
				Type:        "paragraph",
				HeadingPath: headingStackPath(headingStack),
				Content:     linesToString(source, n),
				Level:       headingStackCurrentLevel(headingStack),
			})

		case ast.KindFencedCodeBlock:
			elements = append(elements, mdElement{
				Type:        "code_block",
				HeadingPath: headingStackPath(headingStack),
				Content:     rawSegment(source, n),
				Level:       headingStackCurrentLevel(headingStack),
			})

		case ast.KindList:
			elements = append(elements, mdElement{
				Type:        "list",
				HeadingPath: headingStackPath(headingStack),
				Content:     rawSegment(source, n),
				Level:       headingStackCurrentLevel(headingStack),
			})

		case ast.KindBlockquote:
			elements = append(elements, mdElement{
				Type:        "blockquote",
				HeadingPath: headingStackPath(headingStack),
				Content:     rawSegment(source, n),
				Level:       headingStackCurrentLevel(headingStack),
			})

		case ast.KindThematicBreak:
			elements = append(elements, mdElement{
				Type:        "thematic_break",
				HeadingPath: headingStackPath(headingStack),
				Content:     rawSegment(source, n),
				Level:       headingStackCurrentLevel(headingStack),
			})

		case ast.KindHTMLBlock:
			elements = append(elements, mdElement{
				Type:        "html_block",
				HeadingPath: headingStackPath(headingStack),
				Content:     rawSegment(source, n),
				Level:       headingStackCurrentLevel(headingStack),
			})
		}
		return ast.WalkContinue, nil
	})

	return elements
}

// markdownChunkToDocument 将 Markdown 块转换为标准文档。
func markdownChunkToDocument(c *markdownChunk, index, total int) *schema.Document {
	return &schema.Document{
		ID:      fmt.Sprintf("chunk_%d", index),
		Content: c.Content,
		MetaData: map[string]any{
			metaKeyChunkIndex:    index,
			metaKeyTotalChunks:   total,
			metaKeyHeadingPath:   c.HeadingPath,
			metaKeyElementTypes:  dedupeTypes(c.ElementTypes),
			metaKeyChunkStrategy: "md",
		},
	}
}

// aggregateElements 将元素列表聚合为不超过 chunkSize 的块。
func aggregateElements(elements []mdElement, cfg ChunkConfig) []markdownChunk {
	if len(elements) == 0 {
		return nil
	}

	var chunks []markdownChunk
	var buf []mdElement
	bufLen := 0
	bufPath := ""
	bufLevel := 0

	flush := func() {
		if len(buf) == 0 {
			return
		}
		var sb strings.Builder
		types := make([]string, 0, len(buf))
		for _, e := range buf {
			if sb.Len() > 0 {
				sb.WriteString("\n\n")
			}
			sb.WriteString(e.Content)
			types = append(types, e.Type)
		}
		chunks = append(chunks, markdownChunk{
			Content:      sb.String(),
			HeadingPath:  bufPath,
			ElementTypes: types,
			Level:        bufLevel,
		})
		buf = nil
		bufLen = 0
	}

	for _, e := range elements {
		el := len([]rune(e.Content))

		// 标题路径变化作为 section 边界
		if e.HeadingPath != bufPath && len(buf) > 0 {
			flush()
		}
		if bufPath == "" || (e.HeadingPath != "" && e.HeadingPath != bufPath) {
			bufPath = e.HeadingPath
			bufLevel = e.Level
		}

		// 原子元素如果单元素就超长，强制截断
		if isAtomic(e.Type) && el > cfg.ChunkSize {
			flush()
			chunks = append(chunks, splitAtomicElement(e, cfg.ChunkSize))
			continue
		}

		// 超过容量时 flush
		if bufLen+el > cfg.ChunkSize && len(buf) > 0 {
			flush()
		}

		buf = append(buf, e)
		bufLen += el
	}

	flush()

	// 对仍然超长的块，用 freeChunker 兜底再切
	return flattenOversized(chunks, cfg)
}

// isAtomic 判断元素类型是否应保持完整性、不被内部切分。
func isAtomic(typ string) bool {
	switch typ {
	case "code_block", "list", "blockquote", "table", "html_block", "thematic_break":
		return true
	}
	return false
}

// splitAtomicElement 将超长原子元素强制按 chunkSize 截断。
func splitAtomicElement(e mdElement, chunkSize int) markdownChunk {
	runes := []rune(e.Content)
	if len(runes) <= chunkSize {
		return markdownChunk{
			Content:      e.Content,
			HeadingPath:  e.HeadingPath,
			ElementTypes: []string{e.Type},
			Level:        e.Level,
		}
	}
	return markdownChunk{
		Content:      string(runes[:chunkSize]),
		HeadingPath:  e.HeadingPath,
		ElementTypes: []string{e.Type, "split_" + e.Type},
		Level:        e.Level,
	}
}

// flattenOversized 将聚合后仍然超长的块用 freeChunker 二次细分。
func flattenOversized(chunks []markdownChunk, cfg ChunkConfig) []markdownChunk {
	fc := NewFreeChunker()
	var result []markdownChunk
	for _, ch := range chunks {
		if len([]rune(ch.Content)) <= cfg.ChunkSize {
			result = append(result, ch)
			continue
		}
		docs, _ := fc.Chunk(context.Background(), ch.Content, cfg)
		for _, d := range docs {
			result = append(result, markdownChunk{
				Content:      d.Content,
				HeadingPath:  ch.HeadingPath,
				ElementTypes: ch.ElementTypes,
				Level:        ch.Level,
			})
		}
	}
	return result
}

// dedupeTypes 去重并保持顺序。
func dedupeTypes(types []string) []string {
	seen := make(map[string]bool, len(types))
	var out []string
	for _, t := range types {
		if !seen[t] {
			seen[t] = true
			out = append(out, t)
		}
	}
	return out
}

// --- heading stack ---

// newHeadingStack 创建空的 Markdown 标题层级栈。
func newHeadingStack() *headingStack {
	return &model.HeadingStack{}
}

// pushHeading 将标题压入层级栈并移除同级或更低级标题。
func pushHeading(s *headingStack, h headingEntry) {
	for len(s.Stack) > 0 && s.Stack[len(s.Stack)-1].Level >= h.Level {
		s.Stack = s.Stack[:len(s.Stack)-1]
	}
	s.Stack = append(s.Stack, h)
}

// headingStackPath 返回当前标题层级的完整路径。
func headingStackPath(s *headingStack) string {
	if len(s.Stack) == 0 {
		return ""
	}
	parts := make([]string, len(s.Stack))
	for i, h := range s.Stack {
		parts[i] = strings.TrimSpace(h.Text)
	}
	return strings.Join(parts, " > ")
}

// headingStackCurrentLevel 返回当前标题的层级。
func headingStackCurrentLevel(s *headingStack) int {
	if len(s.Stack) == 0 {
		return 0
	}
	return s.Stack[len(s.Stack)-1].Level
}

// --- goldmark helpers ---

// extractText 从 heading 节点的行中提取纯文本（去掉 # 标记）。
func extractText(source []byte, n ast.Node) string {
	lines := n.Lines()
	if lines == nil || lines.Len() == 0 {
		return ""
	}
	var buf strings.Builder
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		buf.Write(source[seg.Start:seg.Stop])
	}
	raw := buf.String()
	raw = strings.TrimLeft(raw, "#")
	raw = strings.TrimSpace(raw)
	return raw
}

// linesToString 从节点的 Lines() 中拼接完整文本。
func linesToString(source []byte, n ast.Node) string {
	lines := n.Lines()
	if lines == nil || lines.Len() == 0 {
		return ""
	}
	var buf strings.Builder
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		buf.Write(source[seg.Start:seg.Stop])
	}
	return buf.String()
}

// rawSegment 从原始 source 中截取节点对应的原始 Markdown 文本。
func rawSegment(source []byte, n ast.Node) string {
	lines := n.Lines()
	if lines == nil || lines.Len() == 0 {
		return ""
	}
	first := lines.At(0)
	last := lines.At(lines.Len() - 1)
	return string(source[first.Start:last.Stop])
}
