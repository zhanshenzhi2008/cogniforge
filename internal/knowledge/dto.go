package knowledge

import (
	"context"

	"cogniforge/internal/model"
)

// ============ 请求结构 ============

type CreateKBRequest struct {
	Name           string  `json:"name" binding:"required"`
	Description    string  `json:"description"`
	FolderID       string  `json:"folder_id"`
	VectorDB       string  `json:"vector_db"`
	EmbeddingModel string  `json:"embedding_model"`
	ChunkSize      *int    `json:"chunk_size"`
	ChunkOverlap   *int    `json:"chunk_overlap"`
	SplitMode      string  `json:"split_mode"`
	SplitSeparator string  `json:"split_separator"`
	RagTopK        *int    `json:"rag_top_k"`
	RagMinScore    *float64 `json:"rag_min_score"`
	Visibility     string  `json:"visibility"`
}

type UpdateKBRequest struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	FolderID       *string  `json:"folder_id"` // nil=不改；""=移到未分类
	VectorDB       string   `json:"vector_db"`
	EmbeddingModel string   `json:"embedding_model"`
	ChunkSize      *int     `json:"chunk_size"`
	ChunkOverlap   *int     `json:"chunk_overlap"`
	SplitMode      string   `json:"split_mode"`
	SplitSeparator *string  `json:"split_separator"`
	RagTopK        *int     `json:"rag_top_k"`
	RagMinScore    *float64 `json:"rag_min_score"`
	Visibility     string   `json:"visibility"`
	Status         string   `json:"status"`
}

// ParseOptions 上传/预览分段参数
type ParseOptions struct {
	ChunkSize      int    `json:"chunk_size"`
	ChunkOverlap   int    `json:"chunk_overlap"`
	SplitMode      string `json:"split_mode"`
	SplitSeparator string `json:"split_separator"`
}

type ChunkPreviewItem struct {
	Index     int    `json:"index"`
	Content   string `json:"content"`
	CharCount int    `json:"char_count"`
}

type ChunkPreviewResponse struct {
	TotalChunks  int                `json:"total_chunks"`
	PreviewCount int                `json:"preview_count"`
	CharCount    int                `json:"char_count"`
	Chunks       []ChunkPreviewItem `json:"chunks"`
}

type CreateFolderRequest struct {
	Name string `json:"name" binding:"required"`
}

type ReparseDocumentRequest struct {
	DocumentID string `json:"document_id" binding:"required"`
}

// DocumentPreview 文档详情 + 可预览正文（文本类）
type DocumentPreview struct {
	model.Document
	Content     string `json:"content"`
	Previewable bool   `json:"previewable"`
	PreviewNote string `json:"preview_note,omitempty"`
	Truncated   bool   `json:"truncated,omitempty"`
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
	ChunkSize      int                    `json:"chunk_size,omitempty"`
	ChunkOverlap   int                    `json:"chunk_overlap,omitempty"`
	SplitMode      string                 `json:"split_mode,omitempty"`
	Separator      string                 `json:"separator,omitempty"`
}

type PreviewChunksRequest struct {
	Context        context.Context `json:"-"`
	FilePath       string          `json:"file_path"`
	ChunkSize      int             `json:"chunk_size,omitempty"`
	ChunkOverlap   int             `json:"chunk_overlap,omitempty"`
	SplitMode      string          `json:"split_mode,omitempty"`
	Separator      string          `json:"separator,omitempty"`
	MaxChunks      int             `json:"max_chunks,omitempty"`
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
