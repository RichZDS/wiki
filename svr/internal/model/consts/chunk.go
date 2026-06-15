package consts

// Strategy 切块策略枚举，决定底层使用哪种分块算法。
type Strategy string

const (
	// StrategyFree 自由切块 — 按分隔符优先级递归分割，适合纯文本与日志。
	StrategyFree Strategy = "free"
	// StrategyMD Markdown 切块 — 解析 AST 按标题层级分割，保留 heading 路径元数据。
	StrategyMD Strategy = "md"
	// StrategyEino 语义切块 — 通过 Eino EmbeddingModel 在低相似度边界处切分。
	StrategyEino Strategy = "eino"
	// StrategyHierarchical 上下文感知分层切块 — 两层父子结构，小子块精确检索，大父块提供上下文。
	StrategyHierarchical Strategy = "hierarchical"
)

// 写入 schema.Document.MetaData 时使用的键名。
const (
	MetaKeyChunkIndex    = "chunk_index"     // 当前块序号，0-based
	MetaKeyTotalChunks   = "chunk_total"     // 该文档被切分的总块数
	MetaKeyHeadingPath   = "heading_path"    // 标题路径，如 "Chapter 1 > Section 1.1"
	MetaKeyElementTypes  = "element_types"   // 块内包含的元素类型列表
	MetaKeyChunkStrategy = "chunk_strategy"  // 生成该块的策略名
	MetaKeyChunkRole     = "chunk_role"      // 分层切块中的角色: "parent" / "child"
	MetaKeyParentContent = "parent_content"  // 父块完整文本（子块用）
	MetaKeyParentChunkID = "parent_chunk_id" // 父块 ID（子块用）
	MetaKeyChildChunkIDs = "child_chunk_ids" // 子块 ID 列表（父块用）
)

// ParentSizeMultiplier 分层切块中父块相对于子块的大小倍数。
const ParentSizeMultiplier = 3
