package chunk

import (
	"context"
	"fmt"
	"strings"

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
type mdChunker struct{}

// NewMDChunker 返回 Markdown 元素分类切块器。
func NewMDChunker() *mdChunker {
	return &mdChunker{}
}

// mdElement 表示 Markdown AST 中的一个可切块元素。
type mdElement struct {
	typ         string // paragraph / code_block / list / table / blockquote / thematic_break / html_block
	headingPath string // 当前标题层级路径，如 "Chapter 1 > Section 1.1"
	content     string // 元素原始 Markdown 文本
	level       int    // 最近的标题层级 (1-6)，0 表示无标题
}

// Chunk 执行 Markdown 元素分类切块。
func (c *mdChunker) Chunk(ctx context.Context, content string, cfg ChunkConfig) ([]*schema.Document, error) {
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
		docs[i] = c.toDocument(i, len(chunks))
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
			headingStack.push(headingEntry{
				text:  extractText(source, n),
				level: h.Level,
			})

		case ast.KindParagraph:
			if n.Parent() != nil && n.Parent().Kind() == ast.KindListItem {
				return ast.WalkContinue, nil
			}
			elements = append(elements, mdElement{
				typ:         "paragraph",
				headingPath: headingStack.path(),
				content:     linesToString(source, n),
				level:       headingStack.currentLevel(),
			})

		case ast.KindFencedCodeBlock:
			elements = append(elements, mdElement{
				typ:         "code_block",
				headingPath: headingStack.path(),
				content:     rawSegment(source, n),
				level:       headingStack.currentLevel(),
			})

		case ast.KindList:
			elements = append(elements, mdElement{
				typ:         "list",
				headingPath: headingStack.path(),
				content:     rawSegment(source, n),
				level:       headingStack.currentLevel(),
			})

		case ast.KindBlockquote:
			elements = append(elements, mdElement{
				typ:         "blockquote",
				headingPath: headingStack.path(),
				content:     rawSegment(source, n),
				level:       headingStack.currentLevel(),
			})

		case ast.KindThematicBreak:
			elements = append(elements, mdElement{
				typ:         "thematic_break",
				headingPath: headingStack.path(),
				content:     rawSegment(source, n),
				level:       headingStack.currentLevel(),
			})

		case ast.KindHTMLBlock:
			elements = append(elements, mdElement{
				typ:         "html_block",
				headingPath: headingStack.path(),
				content:     rawSegment(source, n),
				level:       headingStack.currentLevel(),
			})
		}
		return ast.WalkContinue, nil
	})

	return elements
}

// mdChunk 聚合后的 Markdown 块。
type mdChunk struct {
	content      string
	headingPath  string
	elementTypes []string
	level        int
}

func (c *mdChunk) toDocument(index, total int) *schema.Document {
	return &schema.Document{
		ID:      fmt.Sprintf("chunk_%d", index),
		Content: c.content,
		MetaData: map[string]any{
			metaKeyChunkIndex:    index,
			metaKeyTotalChunks:   total,
			metaKeyHeadingPath:   c.headingPath,
			metaKeyElementTypes:  dedupeTypes(c.elementTypes),
			metaKeyChunkStrategy: "md",
		},
	}
}

// aggregateElements 将元素列表聚合为不超过 chunkSize 的块。
func aggregateElements(elements []mdElement, cfg ChunkConfig) []mdChunk {
	if len(elements) == 0 {
		return nil
	}

	var chunks []mdChunk
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
			sb.WriteString(e.content)
			types = append(types, e.typ)
		}
		chunks = append(chunks, mdChunk{
			content:      sb.String(),
			headingPath:  bufPath,
			elementTypes: types,
			level:        bufLevel,
		})
		buf = nil
		bufLen = 0
	}

	for _, e := range elements {
		el := len([]rune(e.content))

		// 标题路径变化作为 section 边界
		if e.headingPath != bufPath && len(buf) > 0 {
			flush()
		}
		if bufPath == "" || (e.headingPath != "" && e.headingPath != bufPath) {
			bufPath = e.headingPath
			bufLevel = e.level
		}

		// 原子元素如果单元素就超长，强制截断
		if isAtomic(e.typ) && el > cfg.ChunkSize {
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
func splitAtomicElement(e mdElement, chunkSize int) mdChunk {
	runes := []rune(e.content)
	if len(runes) <= chunkSize {
		return mdChunk{
			content:      e.content,
			headingPath:  e.headingPath,
			elementTypes: []string{e.typ},
			level:        e.level,
		}
	}
	return mdChunk{
		content:      string(runes[:chunkSize]),
		headingPath:  e.headingPath,
		elementTypes: []string{e.typ, "split_" + e.typ},
		level:        e.level,
	}
}

// flattenOversized 将聚合后仍然超长的块用 freeChunker 二次细分。
func flattenOversized(chunks []mdChunk, cfg ChunkConfig) []mdChunk {
	fc := &freeChunker{}
	var result []mdChunk
	for _, ch := range chunks {
		if len([]rune(ch.content)) <= cfg.ChunkSize {
			result = append(result, ch)
			continue
		}
		docs, _ := fc.Chunk(context.Background(), ch.content, cfg)
		for _, d := range docs {
			result = append(result, mdChunk{
				content:      d.Content,
				headingPath:  ch.headingPath,
				elementTypes: ch.elementTypes,
				level:        ch.level,
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

type headingEntry struct {
	text  string
	level int
}

type headingStack struct {
	stack []headingEntry
}

func newHeadingStack() *headingStack {
	return &headingStack{}
}

func (s *headingStack) push(h headingEntry) {
	for len(s.stack) > 0 && s.stack[len(s.stack)-1].level >= h.level {
		s.stack = s.stack[:len(s.stack)-1]
	}
	s.stack = append(s.stack, h)
}

func (s *headingStack) pop() {
	if len(s.stack) > 0 {
		s.stack = s.stack[:len(s.stack)-1]
	}
}

func (s *headingStack) path() string {
	if len(s.stack) == 0 {
		return ""
	}
	parts := make([]string, len(s.stack))
	for i, h := range s.stack {
		parts[i] = strings.TrimSpace(h.text)
	}
	return strings.Join(parts, " > ")
}

func (s *headingStack) currentLevel() int {
	if len(s.stack) == 0 {
		return 0
	}
	return s.stack[len(s.stack)-1].level
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
