package embedding

import (
	"context"
	"fmt"
	"os"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
)

// ArkEmbedder 是基于火山引擎 Ark 平台的文本向量化实现，支持文本嵌入和多模态嵌入 API。
type ArkEmbedder = ark.Embedder

// NewArkEmbedder 创建 Ark Embedder 实例。
// 环境变量：
//   - ARK_API_KEY（必填）— API 密钥
//   - ARK_EMBEDDING_MODEL（必填）— Ark 平台上的端点 ID
//   - ARK_BASE_URL（可选，默认 https://ark.cn-beijing.volces.com/api/v3）— 服务基础 URL
//   - ARK_REGION（可选，默认 cn-beijing）— 服务区域
// 返回值：成功时返回 Embedder，失败时返回错误。
func NewArkEmbedder(ctx context.Context, apiKey string, modelID string, baseURL string, region string) (*ArkEmbedder, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("ARK_API_KEY is not set")
	}
	if modelID == "" {
		return nil, fmt.Errorf("ARK_EMBEDDING_MODEL is not set")
	}
	if baseURL == "" {
		baseURL = "https://ark.cn-beijing.volces.com/api/v3"
	}

	if region == "" {
		region = os.Getenv("ARK_REGION")
	}
	if region == "" {
		region = "cn-beijing"
	}

	emb, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		APIKey:  apiKey,
		Model:   modelID,
		BaseURL: baseURL,
		Region:  region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ark embedder: %w", err)
	}
	return emb, nil
}
