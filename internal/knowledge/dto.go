package knowledge

import "context"

// ============ 请求结构 ============

type CreateKBRequest struct {
	Name           string `json:"name" binding:"required"`
	Description    string `json:"description"`
	VectorDB       string `json:"vector_db"`
	EmbeddingModel string `json:"embedding_model"`
}

type UpdateKBRequest struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	VectorDB       string `json:"vector_db"`
	EmbeddingModel string `json:"embedding_model"`
	Status         string `json:"status"`
}

type ReparseDocumentRequest struct {
	DocumentID string `json:"document_id" binding:"required"`
}

type SearchRequest struct {
	Context        context.Context `json:"-"`
	Query          string          `json:"query" binding:"required"`
	CollectionName string          `json:"collection_name"`
	TopK           int             `json:"top_k"`
	MinScore       float64         `json:"min_score"`
}

// ============ 响应结构 ============

type SearchResult struct {
	DocumentID   string  `json:"document_id"`
	DocumentName string  `json:"document_name"`
	ChunkID      string  `json:"chunk_id"`
	Content      string  `json:"content"`
	Score        float64 `json:"score"`
}

type SearchResponse struct {
	Results  []SearchResult `json:"results"`
	Total    int            `json:"total"`
	Query    string         `json:"query"`
	Duration int64          `json:"duration_ms"`
}

// ============ Python RAG 服务 DTO ============

type ProcessRequest struct {
	Context        context.Context        `json:"-"`
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

type RAGSearchResult struct {
	ChunkID  string                 `json:"chunk_id"`
	Content  string                 `json:"content"`
	Score    float64                `json:"score"`
	Metadata map[string]interface{} `json:"metadata"`
}

type RAGSearchResponse struct {
	Query   string            `json:"query"`
	Results []RAGSearchResult `json:"results"`
	Total   int               `json:"total"`
}

type DeleteDocumentRequest struct {
	Context        context.Context `json:"-"`
	DocumentID     string
	CollectionName string
}

type DeleteDocumentResponse struct {
	Success       bool   `json:"success"`
	DocumentID    string `json:"document_id"`
	ChunksDeleted int    `json:"chunks_deleted"`
}
