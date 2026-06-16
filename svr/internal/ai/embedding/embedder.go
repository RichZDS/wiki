package embedding

import einoembedding "github.com/cloudwego/eino/components/embedding"

// Embedder 是项目内部使用的文本向量化接口，与 Eino 官方接口保持一致。
//
// 通过类型别名统一接口签名，避免在 chunker / indexer / retriever 之间编写适配层。
type Embedder = einoembedding.Embedder
