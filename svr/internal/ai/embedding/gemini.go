package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"wiki/internal/model"
	"wiki/pkg/database"

	einoembedding "github.com/cloudwego/eino/components/embedding"
)

// GeminiEmbedder 是基于 Google Gemini API 的文本向量化实现，
// 使用 gemini-embedding-2-preview 等模型生成稠密向量。
type GeminiEmbedder struct {
	apiKey  string
	modelID string
	client  *http.Client
}

// geminiBatchEmbedRequest 是 Gemini batchEmbedContents API 的请求体。
type geminiBatchEmbedRequest struct {
	Requests []geminiEmbedRequest `json:"requests"`
}

// geminiEmbedRequest 是单个文本的 embedding 请求。
type geminiEmbedRequest struct {
	Model   string               `json:"model"`
	Content geminiEmbedContent   `json:"content"`
}

// geminiEmbedContent 是 embedding 请求中的内容部分。
type geminiEmbedContent struct {
	Parts []geminiEmbedPart `json:"parts"`
}

// geminiEmbedPart 是内容片段。
type geminiEmbedPart struct {
	Text string `json:"text"`
}

// geminiBatchEmbedResponse 是 Gemini batchEmbedContents API 的响应体。
type geminiBatchEmbedResponse struct {
	Embeddings []geminiEmbedding `json:"embeddings"`
}

// geminiEmbedding 是单个文本的向量表示。
type geminiEmbedding struct {
	Values []float64 `json:"values"`
}

// NewGeminiEmbedder 创建 Gemini Embedder 实例。
// apiKey 为 Gemini API 密钥，modelID 为模型 ID（如 gemini-embedding-2-preview）。
func NewGeminiEmbedder(ctx context.Context, apiKey, modelID string) (*GeminiEmbedder, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("gemini api key is required")
	}
	if modelID == "" {
		return nil, fmt.Errorf("gemini model id is required")
	}

	return &GeminiEmbedder{
		apiKey:  apiKey,
		modelID: modelID,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

// NewGeminiEmbedderFromDB 从数据库 ai_model 表中读取 model_name="embedding" 的配置，
// 并创建 Gemini Embedder 实例。
// 返回的 Embedder 实现了 eino 的 embedding.Embedder 接口，可用于 Indexer 和 Retriever。
func NewGeminiEmbedderFromDB(ctx context.Context) (*GeminiEmbedder, error) {
	aimodel, err := model.GetAIModelByName(ctx, database.DB, "embedding")
	if err != nil {
		return nil, fmt.Errorf("query ai_model 'embedding': %w", err)
	}
	if aimodel.APIKey == "" {
		return nil, fmt.Errorf("api_key for 'embedding' model is not configured")
	}
	if aimodel.ModelId == "" {
		return nil, fmt.Errorf("model_id for 'embedding' model is not configured")
	}

	return NewGeminiEmbedder(ctx, aimodel.APIKey, aimodel.ModelId)
}

// EmbedStrings 实现 eino embedding.Embedder 接口，将文本序列转换为向量序列。
func (e *GeminiEmbedder) EmbedStrings(ctx context.Context, texts []string, opts ...einoembedding.Option) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	requests := make([]geminiEmbedRequest, len(texts))
	for i, text := range texts {
		requests[i] = geminiEmbedRequest{
			Model: fmt.Sprintf("models/%s", e.modelID),
			Content: geminiEmbedContent{
				Parts: []geminiEmbedPart{{Text: text}},
			},
		}
	}

	body := geminiBatchEmbedRequest{Requests: requests}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal gemini request: %w", err)
	}

	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:batchEmbedContents?key=%s",
		e.modelID, e.apiKey,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create gemini request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini embed request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read gemini response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini embed request returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result geminiBatchEmbedResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal gemini response: %w", err)
	}

	if len(result.Embeddings) != len(texts) {
		return nil, fmt.Errorf("gemini returned %d embeddings for %d texts", len(result.Embeddings), len(texts))
	}

	vectors := make([][]float64, len(result.Embeddings))
	for i, emb := range result.Embeddings {
		vectors[i] = emb.Values
	}

	return vectors, nil
}
