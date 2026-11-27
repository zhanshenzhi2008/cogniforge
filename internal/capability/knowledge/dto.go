package knowledge

import (
	"context"
	"time"

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
	RagTopK              *int     `json:"rag_top_k"`
	RagMinScore          *float64 `json:"rag_min_score"`
	RagTopKSemantic      *int     `json:"rag_top_k_semantic"`
	RagMinScoreSemantic  *float64 `json:"rag_min_score_semantic"`
	RagTopKKeyword       *int     `json:"rag_top_k_keyword"`
	RagMinScoreKeyword   *float64 `json:"rag_min_score_keyword"`
	RagTopKHybrid        *int     `json:"rag_top_k_hybrid"`
	RagMinScoreHybrid    *float64 `json:"rag_min_score_hybrid"`
	HybridSemanticWeight *float64 `json:"hybrid_semantic_weight"`
	SearchMode               string   `json:"search_mode"`
	RerankSemantic           *bool    `json:"rerank_semantic"`
	RerankKeyword            *bool    `json:"rerank_keyword"`
	RerankHybrid             *bool    `json:"rerank_hybrid"`
	RerankProviderSemantic   *string  `json:"rerank_provider_semantic"`
	RerankProviderKeyword    *string  `json:"rerank_provider_keyword"`
	RerankProviderHybrid     *string  `json:"rerank_provider_hybrid"`
	Visibility               string   `json:"visibility"`
}

type UpdateKBRequest struct {
	Name                     string   `json:"name"`
	Description              string   `json:"description"`
	FolderID                 *string  `json:"folder_id"` // nil=不改；""=移到未分类
	VectorDB                 string   `json:"vector_db"`
	EmbeddingModel           string   `json:"embedding_model"`
	ChunkSize                *int     `json:"chunk_size"`
	ChunkOverlap             *int     `json:"chunk_overlap"`
	SplitMode                string   `json:"split_mode"`
	SplitSeparator           *string  `json:"split_separator"`
	RagTopK                  *int     `json:"rag_top_k"`
	RagMinScore              *float64 `json:"rag_min_score"`
	RagTopKSemantic          *int     `json:"rag_top_k_semantic"`
	RagMinScoreSemantic      *float64 `json:"rag_min_score_semantic"`
	RagTopKKeyword           *int     `json:"rag_top_k_keyword"`
	RagMinScoreKeyword       *float64 `json:"rag_min_score_keyword"`
	RagTopKHybrid            *int     `json:"rag_top_k_hybrid"`
	RagMinScoreHybrid        *float64 `json:"rag_min_score_hybrid"`
	HybridSemanticWeight     *float64 `json:"hybrid_semantic_weight"`
	SearchMode               string   `json:"search_mode"`
	RerankSemantic           *bool    `json:"rerank_semantic"`
	RerankKeyword            *bool    `json:"rerank_keyword"`
	RerankHybrid             *bool    `json:"rerank_hybrid"`
	RerankProviderSemantic   *string  `json:"rerank_provider_semantic"`
	RerankProviderKeyword    *string  `json:"rerank_provider_keyword"`
	RerankProviderHybrid     *string  `json:"rerank_provider_hybrid"`
	Visibility               string   `json:"visibility"`
	Status                   string   `json:"status"`
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

// CreateTextDocumentRequest 自定义文本入库（md/txt）
type CreateTextDocumentRequest struct {
	Name           string `json:"name" binding:"required"`
	Content        string `json:"content" binding:"required"`
	FileType       string `json:"file_type"` // md | txt，默认 md
	ChunkSize      int    `json:"chunk_size,omitempty"`
	ChunkOverlap   int    `json:"chunk_overlap,omitempty"`
	SplitMode      string `json:"split_mode,omitempty"`
	SplitSeparator string `json:"split_separator,omitempty"`
}

// CreateURLDocumentRequest 抓取网页正文后入库
type CreateURLDocumentRequest struct {
	URL            string `json:"url" binding:"required"`
	Name           string `json:"name"`
	// Engine: auto | jina | firecrawl | basic；空=auto（优先 Firecrawl→Jina→简易 GET）
	Engine         string `json:"engine,omitempty"`
	ChunkSize      int    `json:"chunk_size,omitempty"`
	ChunkOverlap   int    `json:"chunk_overlap,omitempty"`
	SplitMode      string `json:"split_mode,omitempty"`
	SplitSeparator string `json:"split_separator,omitempty"`
}

// DocumentPreview 文档详情 + 可预览正文（文本类）
type DocumentPreview struct {
	model.Document
	Content     string `json:"content"`
	Previewable bool   `json:"previewable"`
	Editable    bool   `json:"editable"`
	PreviewNote string `json:"preview_note,omitempty"`
	Truncated   bool   `json:"truncated,omitempty"`
}

// UpdateDocumentContentRequest 更新文本类文档正文（md/txt/html）
type UpdateDocumentContentRequest struct {
	Content string `json:"content"`
	// Reparse 默认 true：保存后重新分块向量，保证检索跟上正文
	Reparse *bool `json:"reparse,omitempty"`
}

type SearchRequest struct {
	Context        context.Context `json:"-"`
	Query          string          `json:"query" binding:"required"`
	CollectionName string          `json:"collection_name"`
	TopK           int             `json:"top_k"`
	MinScore       float64         `json:"min_score"`
	// SearchMode / Rerank：测试与对话可覆盖库配置；关键字/混合/Rerank 引擎后续接入
	SearchMode           string   `json:"search_mode,omitempty"`
	Rerank               *bool    `json:"rerank,omitempty"`
	RerankProviderID     string   `json:"rerank_provider_id,omitempty"`
	HybridSemanticWeight *float64 `json:"hybrid_semantic_weight,omitempty"`
	// IncludeDrafts：管理排查用，默认 false（只召回 published）
	IncludeDrafts bool `json:"include_drafts,omitempty"`
	// IgnoreValidity：忽略生效/失效时间窗，默认 false
	IgnoreValidity bool `json:"ignore_validity,omitempty"`
}

// UpdateDocumentLifecycleRequest 更新业务状态、有效期与密级
type UpdateDocumentLifecycleRequest struct {
	LifecycleStatus *string    `json:"lifecycle_status,omitempty"` // draft|review|published|abolished
	SecurityLevel   *string    `json:"security_level,omitempty"`   // public|internal|confidential|secret
	EffectiveAt     *time.Time `json:"effective_at,omitempty"`
	ExpireAt        *time.Time `json:"expire_at,omitempty"`
	ClearEffective  bool       `json:"clear_effective,omitempty"`
	ClearExpire     bool       `json:"clear_expire,omitempty"`
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
