package knowledge

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

type SearchRequest struct {
	Query    string  `json:"query" binding:"required"`
	TopK     int     `json:"top_k"`
	MinScore float64 `json:"min_score"`
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
