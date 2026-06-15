package chunk

import (
	"context"
	"fmt"
	"strings"

	"wiki/internal/model"

	"github.com/cloudwego/eino/schema"
)

// mdChunker 按 Markdown 标题层级切分文档（参照 Eino Header Splitter 思路）。
//
// 与 goldmark AST 方案相比，本方案采用逐行扫描，更简单直接：
//  1. 逐行扫描，识别标题边界（#、##、### 等）
//  2. 代码块内（``` 或 ~~~ 围起）不切分
//  3. 维护标题层级栈，为每个块记录 heading_path 元数据
//  4. 超长段落用 freeChunker 兜底二次切分
type mdChunker = model.MarkdownChunker

// NewMDChunker 返回 Markdown 标题切块器。
func NewMDChunker() *mdChunker {
	return &model.MarkdownChunker{ChunkFunc: mdChunk}
}

type headingEntry = model.HeadingEntry
type headingStack = model.HeadingStack

// mdSection 表示按标题切出的一个段落，包含内容和当前标题路径。
type mdSection struct {
	content     string
	headingPath string
}

// mdChunk 执行 Markdown 标题切块。
func mdChunk(ctx context.Context, content string, cfg ChunkConfig) ([]*schema.Document, error) {
	sanitizeConfig(&cfg)
	if len(content) == 0 {
		return nil, nil
	}

	sections := splitByHeaders(content)
	if len(sections) == 0 {
		return []*schema.Document{{
			ID:      "chunk_0",
			Content: content,
			MetaData: map[string]any{
				metaKeyChunkIndex:    0,
				metaKeyTotalChunks:   1,
				metaKeyChunkStrategy: "md",
			},
		}}, nil
	}

	// 收集所有文档块，超长段落用 freeChunker 二次切分
	var docs []*schema.Document
	for _, sec := range sections {
		if runeLen(sec.content) <= cfg.ChunkSize {
			docs = append(docs, sectionToDoc(sec, len(docs)))
		} else {
			subDocs := splitOversizedSection(sec, cfg, len(docs))
			docs = append(docs, subDocs...)
		}
	}

	// 重新编号
	for i, d := range docs {
		d.ID = fmt.Sprintf("chunk_%d", i)
		d.MetaData[metaKeyChunkIndex] = i
		d.MetaData[metaKeyTotalChunks] = len(docs)
	}

	return docs, nil
}

// splitByHeaders 逐行扫描，在标题边界处切分文档。
//
// 参照 Eino Header Splitter 的 splitText 算法：
//  - 维护 currentLines 累积当前段落的所有行
//  - 遇到标题时 flush 当前段落，更新标题层级栈
//  - 代码块内部不触发 flush
func splitByHeaders(text string) []mdSection {
	var sections []mdSection
	stack := newHeadingStack()
	var currentLines []string
	inCodeBlock := false
	var fence string

	lines := strings.Split(text, "\n")

	for _, rawLine := range lines {
		line := strings.TrimRight(rawLine, "\r")
		trimmed := strings.TrimSpace(line)

		// 空行直接追加
		if trimmed == "" {
			currentLines = append(currentLines, line)
			continue
		}

		// 代码块边界检测
		if !inCodeBlock {
			if isFenceStart(trimmed) {
				inCodeBlock = true
				fence = detectFence(trimmed)
			}
		} else {
			if strings.HasPrefix(trimmed, fence) {
				inCodeBlock = false
				fence = ""
			}
		}

		// 代码块内直接追加，不做切分
		if inCodeBlock {
			currentLines = append(currentLines, line)
			continue
		}

		// 检查是否为新标题
		level, headingText := parseHeading(trimmed)
		if level > 0 {
			// flush 当前累积的段落
			if len(currentLines) > 0 {
				sections = append(sections, mdSection{
					content:     strings.Join(currentLines, "\n"),
					headingPath: headingStackPath(stack),
				})
				currentLines = nil
			}

			// 更新标题层级栈：弹出同级或更低级标题
			pushHeading(stack, headingEntry{Text: headingText, Level: level})
			currentLines = append(currentLines, line)
			continue
		}

		currentLines = append(currentLines, line)
	}

	// flush 末尾未切分的段落
	if len(currentLines) > 0 {
		sections = append(sections, mdSection{
			content:     strings.Join(currentLines, "\n"),
			headingPath: headingStackPath(stack),
		})
	}

	return sections
}

// parseHeading 解析标题行，返回标题级别和纯文本。
// 仅当行以 # 开头且后面跟随空格时才视为标题。
// 例如 "# Title" → (1, "Title")，"## Section" → (2, "Section")。
func parseHeading(line string) (level int, text string) {
	if !strings.HasPrefix(line, "#") {
		return 0, ""
	}
	for _, ch := range line {
		if ch == '#' {
			level++
		} else {
			break
		}
	}
	if level == 0 || level > 6 {
		return 0, ""
	}
	// # 后必须跟空格或行结束
	if len(line) > level && line[level] != ' ' {
		return 0, ""
	}
	text = strings.TrimSpace(line[level:])
	return level, text
}

// isFenceStart 判断行是否为代码块围栏起始（``` 或 ~~~）。
func isFenceStart(line string) bool {
	return strings.HasPrefix(line, "```") || strings.HasPrefix(line, "~~~")
}

// detectFence 返回围栏字符串（"```" 或 "~~~"）。
func detectFence(line string) string {
	if strings.HasPrefix(line, "```") {
		return "```"
	}
	return "~~~"
}

// sectionToDoc 将标题段落转换为标准 Document。
func sectionToDoc(sec mdSection, index int) *schema.Document {
	return &schema.Document{
		ID:      fmt.Sprintf("chunk_%d", index),
		Content: sec.content,
		MetaData: map[string]any{
			metaKeyChunkIndex:    index,
			metaKeyHeadingPath:   sec.headingPath,
			metaKeyChunkStrategy: "md",
		},
	}
}

// splitOversizedSection 将超长段落用 freeChunker 二次切分，
// 并继承原段落的 heading_path 元数据。
func splitOversizedSection(sec mdSection, cfg ChunkConfig, startIndex int) []*schema.Document {
	fc := NewFreeChunker()
	subDocs, _ := fc.Chunk(context.Background(), sec.content, cfg)
	for _, d := range subDocs {
		if sec.headingPath != "" {
			d.MetaData[metaKeyHeadingPath] = sec.headingPath
		}
		d.MetaData[metaKeyChunkStrategy] = "md"
	}
	return subDocs
}

// runeLen 返回字符串的字符数（而非字节数）。
func runeLen(s string) int {
	return len([]rune(s))
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

// headingStackPath 返回当前标题层级的完整路径（"> " 分隔）。
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
