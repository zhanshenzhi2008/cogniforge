package knowledge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"cogniforge/internal/config"
	"cogniforge/internal/trace"
)

type PythonServiceClient struct {
	baseURL string
	client  *http.Client
}

type ProcessRequest struct {
	FilePath       string                 `json:"file_path"`
	DocumentID     string                 `json:"document_id"`
	CollectionName string                 `json:"collection_name"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

type ProcessResponse struct {
	Success       bool   `json:"success"`
	DocumentID    string `json:"document_id"`
	ChunksCreated int    `json:"chunks_created"`
	Error         string `json:"error,omitempty"`
}

type PythonSearchRequest struct {
	Query          string  `json:"query"`
	CollectionName string  `json:"collection_name"`
	TopK           int     `json:"top_k"`
	MinScore       float64 `json:"min_score"`
}

type PythonSearchResult struct {
	ChunkID  string                 `json:"chunk_id"`
	Content  string                 `json:"content"`
	Score    float64                `json:"score"`
	Metadata map[string]interface{} `json:"metadata"`
}

type PythonSearchResponse struct {
	Query   string               `json:"query"`
	Results []PythonSearchResult `json:"results"`
	Total   int                  `json:"total"`
}

func NewPythonServiceClient(cfg *config.Config) *PythonServiceClient {
	return &PythonServiceClient{
		baseURL: cfg.RAG.PythonServiceURL,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *PythonServiceClient) ProcessDocument(ctx context.Context, req *ProcessRequest) (*ProcessResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/api/rag/process", c.baseURL),
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Pass trace_id to Python service (统一使用 X-Trace-ID)
	if traceID := trace.GetTraceIDFromContext(ctx); traceID != "" {
		httpReq.Header.Set("X-Trace-ID", traceID)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call python service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("python service error (status %d): %s", resp.StatusCode, string(body))
	}

	var result ProcessResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *PythonServiceClient) Search(ctx context.Context, req *PythonSearchRequest) (*PythonSearchResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/api/rag/search", c.baseURL),
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Pass trace_id to Python service (统一使用 X-Trace-ID)
	if traceID := trace.GetTraceIDFromContext(ctx); traceID != "" {
		httpReq.Header.Set("X-Trace-ID", traceID)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call python service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("python service error (status %d): %s", resp.StatusCode, string(body))
	}

	var result PythonSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *PythonServiceClient) Health() bool {
	resp, err := c.client.Get(fmt.Sprintf("%s/health", c.baseURL))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

type DeleteDocumentResponse struct {
	Success       bool   `json:"success"`
	DocumentID    string `json:"document_id"`
	ChunksDeleted int    `json:"chunks_deleted"`
}

// DeleteDocument 删除文档的所有向量
func (c *PythonServiceClient) DeleteDocument(ctx context.Context, documentID, collectionName string) error {
	url := fmt.Sprintf("%s/api/rag/%s/%s", c.baseURL, collectionName, documentID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if traceID := trace.GetTraceIDFromContext(ctx); traceID != "" {
		httpReq.Header.Set("X-Trace-ID", traceID)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to call python service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("python service error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}
