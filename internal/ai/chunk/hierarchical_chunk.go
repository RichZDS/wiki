package chunk

import (
	"context"
	"fmt"

	"aisearch/internal/model"

	"github.com/cloudwego/eino/schema"
)

// hierarchicalChunker 上下文感知分层切块器。
//
// 两层结构：
//   - 父块（parent）：较大粒度（chunkSize * 3），保留完整上下文
//   - 子块（child）：较小粒度（原始 chunkSize），用于精确检索
//
// 父子通过 metadata 关联：子块记录 parent_content / parent_chunk_id，
// 父块记录 child_chunk_ids。返回切片中父块在前、子块在后。
type hierarchicalChunker = model.HierarchicalChunker

// NewHierarchicalChunker 创建上下文感知分层切块器。
func NewHierarchicalChunker() *hierarchicalChunker {
	return &model.HierarchicalChunker{ChunkFunc: hierarchicalChunk}
}

const parentSizeMultiplier = 3

// Chunk 执行分层切块。
func hierarchicalChunk(ctx context.Context, content string, cfg ChunkConfig) ([]*schema.Document, error) {
	sanitizeConfig(&cfg)
	if len(content) == 0 {
		return nil, nil
	}
	runes := []rune(content)
	if len(runes) <= cfg.ChunkSize {
		doc := newDocument(content, 0, 1)
		doc.MetaData[metaKeyChunkStrategy] = "hierarchical"
		doc.MetaData[metaKeyChunkRole] = "parent"
		return []*schema.Document{doc}, nil
	}

	// 第一层：父级切块（大粒度）
	parentCfg := cfg
	parentCfg.ChunkSize = cfg.ChunkSize * parentSizeMultiplier
	parentCfg.ChunkOverlap = 0 // 父级不需要 overlap
	parentDocs, err := mdChunk(ctx, content, parentCfg)
	if err != nil {
		return nil, fmt.Errorf("hierarchicalChunker: parent chunking failed: %w", err)
	}

	// 第二层：子级切块（小粒度）
	childCfg := cfg
	childCfg.ChunkOverlap = 0 // 子级通过 parent_content 提供上下文，不需要 overlap

	var result []*schema.Document
	childIdx := 0

	for pi, pDoc := range parentDocs {
		parentID := fmt.Sprintf("parent_%d", pi)
		parentHeadingPath, _ := pDoc.MetaData[metaKeyHeadingPath].(string)

		// 为每个父块创建 Document
		parentDoc := &schema.Document{
			ID:      parentID,
			Content: pDoc.Content,
			MetaData: map[string]any{
				metaKeyChunkIndex:    pi,
				metaKeyTotalChunks:   len(parentDocs),
				metaKeyChunkStrategy: "hierarchical",
				metaKeyChunkRole:     "parent",
				metaKeyHeadingPath:   parentHeadingPath,
			},
		}

		// 对父块内容进行子级切块
		childDocs, err := mdChunk(ctx, pDoc.Content, childCfg)
		if err != nil {
			return nil, fmt.Errorf("hierarchicalChunker: child chunking failed: %w", err)
		}

		var childIDs []string
		for _, cDoc := range childDocs {
			childID := fmt.Sprintf("child_%d", childIdx)
			childIDs = append(childIDs, childID)

			childHeadingPath := parentHeadingPath
			if cp, ok := cDoc.MetaData[metaKeyHeadingPath].(string); ok && cp != "" {
				childHeadingPath = cp
			}

			result = append(result, &schema.Document{
				ID:      childID,
				Content: cDoc.Content,
				MetaData: map[string]any{
					metaKeyChunkIndex:    childIdx,
					metaKeyTotalChunks:   0, // 会在最后统一设置
					metaKeyChunkStrategy: "hierarchical",
					metaKeyChunkRole:     "child",
					metaKeyHeadingPath:   childHeadingPath,
					metaKeyParentContent: pDoc.Content,
					metaKeyParentChunkID: parentID,
				},
			})
			childIdx++
		}

		parentDoc.MetaData[metaKeyChildChunkIDs] = childIDs
		parentDoc.MetaData[metaKeyTotalChunks] = len(childDocs)
		result = append(result, parentDoc)
	}

	// 将父块移到最前面：先父块，后子块
	parents := result[len(result)-len(parentDocs):]
	children := result[:len(result)-len(parentDocs)]
	final := make([]*schema.Document, 0, len(result))
	final = append(final, parents...)
	final = append(final, children...)

	// 修正子块的 total count
	totalChildren := len(children)
	for _, d := range children {
		d.MetaData[metaKeyTotalChunks] = totalChildren
	}

	return final, nil
}
