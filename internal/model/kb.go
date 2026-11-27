package model

import (
	"time"

	"gorm.io/gorm"
)

// =============================================================================
// Knowledge Base Models - 知识库
// =============================================================================

// KnowledgeFolder 知识库目录（用户侧分组）
type KnowledgeFolder struct {
	ID        string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	UserID    string         `gorm:"type:varchar(64);not null;uniqueIndex:uk_kb_folder_user_name;index:idx_kb_folder_user" json:"user_id"`
	Name      string         `gorm:"type:varchar(100);not null;uniqueIndex:uk_kb_folder_user_name" json:"name"`
	SortOrder int            `gorm:"default:0" json:"sort_order"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (KnowledgeFolder) TableName() string {
	return "knowledge_folders"
}

// KnowledgeBase 知识库表
type KnowledgeBase struct {
	ID             string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	UserID         string         `gorm:"type:varchar(64);not null;index;index:idx_kb_user_storage,priority:1" json:"user_id"`
	FolderID       string         `gorm:"type:varchar(64);index:idx_kb_folder_id" json:"folder_id"` // 空=未分类
	Name           string         `gorm:"type:varchar(255);not null" json:"name"`
	StorageDir     string         `gorm:"type:varchar(120);index:idx_kb_user_storage,priority:2" json:"storage_dir"` // 文件目录名，取知识库名；检索仍用 ID
	Description    string         `gorm:"type:text" json:"description"`
	VectorDB       string         `gorm:"type:varchar(50);default:'pgvector'" json:"vector_db"` // pgvector, chroma, qdrant, weaviate
	EmbeddingModel string         `gorm:"type:varchar(100);default:'text-embedding-ada-002'" json:"embedding_model"`
	// 默认分段（上传时可覆盖）
	ChunkSize      int            `gorm:"default:512" json:"chunk_size"`
	ChunkOverlap   int            `gorm:"default:50" json:"chunk_overlap"`
	SplitMode      string         `gorm:"type:varchar(32);default:'recursive'" json:"split_mode"` // recursive | fixed | separator
	SplitSeparator string         `gorm:"type:varchar(64);default:''" json:"split_separator"`
	// 检索参数（兼容字段：与当前 search_mode 对应的 TopK / 阈值同步）
	RagTopK     int     `gorm:"default:5" json:"rag_top_k"`
	RagMinScore float64 `gorm:"type:double precision;default:0.3" json:"rag_min_score"`
	// 三种检索各自的 TopK / 最低分
	RagTopKSemantic     int     `gorm:"default:5" json:"rag_top_k_semantic"`
	RagMinScoreSemantic float64 `gorm:"type:double precision;default:0.3" json:"rag_min_score_semantic"`
	RagTopKKeyword      int     `gorm:"default:5" json:"rag_top_k_keyword"`
	RagMinScoreKeyword  float64 `gorm:"type:double precision;default:0.3" json:"rag_min_score_keyword"`
	RagTopKHybrid       int     `gorm:"default:5" json:"rag_top_k_hybrid"`
	RagMinScoreHybrid   float64 `gorm:"type:double precision;default:0.3" json:"rag_min_score_hybrid"`
	// 混合检索：语义权重（关键词权重 = 1 - 本值）
	HybridSemanticWeight float64 `gorm:"type:double precision;default:0.7" json:"hybrid_semantic_weight"`
	// 检索方式：semantic | keyword | hybrid（实际混合/关键字引擎后续接入，先落配置）
	SearchMode string `gorm:"type:varchar(32);default:'semantic'" json:"search_mode"`
	// 各检索方式是否开启 Rerank，以及选用的 Rerank 供应商（须具备 rerank 能力）
	RerankSemantic           bool   `gorm:"default:false" json:"rerank_semantic"`
	RerankKeyword            bool   `gorm:"default:false" json:"rerank_keyword"`
	RerankHybrid             bool   `gorm:"default:false" json:"rerank_hybrid"`
	RerankProviderSemantic   string `gorm:"type:varchar(64);default:''" json:"rerank_provider_semantic"`
	RerankProviderKeyword    string `gorm:"type:varchar(64);default:''" json:"rerank_provider_keyword"`
	RerankProviderHybrid     string `gorm:"type:varchar(64);default:''" json:"rerank_provider_hybrid"`
	// 可见性占位（阶段三）：private | org（暂仅 private）
	Visibility string `gorm:"type:varchar(32);default:'private'" json:"visibility"`
	Status     string `gorm:"type:varchar(50);default:'active'" json:"status"`
	Metadata   JSONBMap       `gorm:"type:jsonb" json:"metadata"`
	DocCount   int            `gorm:"default:0" json:"doc_count"` // 文档数量
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (KnowledgeBase) TableName() string {
	return "knowledge_bases"
}

// 文档业务生命周期（与解析流水 status 分开）
const (
	DocLifecycleDraft      = "draft"
	DocLifecycleReview     = "review"
	DocLifecyclePublished  = "published"
	DocLifecycleAbolished  = "abolished"
)

// Document 文档表
type Document struct {
	ID              string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	KnowledgeBaseID string         `gorm:"type:varchar(64);not null;index" json:"knowledge_base_id"`
	UserID          string         `gorm:"type:varchar(64);not null;index" json:"user_id"`
	Name            string         `gorm:"type:varchar(255);not null" json:"name"`
	FileName        string         `gorm:"type:varchar(255)" json:"file_name"`               // 原始文件名
	FileSize        int64          `gorm:"type:bigint" json:"file_size"`                     // 文件大小(bytes)
	FileType        string         `gorm:"type:varchar(50)" json:"file_type"`                // pdf, txt, md, docx
	FilePath        string         `gorm:"type:text" json:"file_path"`                       // 存储路径
	Status          string         `gorm:"type:varchar(50);default:'pending'" json:"status"` // 解析流水：pending, processing, completed, failed
	LifecycleStatus string         `gorm:"type:varchar(32);not null;default:'published';index:idx_doc_lifecycle" json:"lifecycle_status"` // draft/review/published/abolished
	EffectiveAt     *time.Time     `json:"effective_at,omitempty"`                           // 空=立即生效
	ExpireAt        *time.Time     `json:"expire_at,omitempty"`                              // 空=不过期
	Error           string         `gorm:"type:text" json:"error"`                           // 错误信息
	ChunkCount      int            `gorm:"default:0" json:"chunk_count"`                     // 分块数量
	VectorCount     int            `gorm:"default:0" json:"vector_count"`                    // 向量数量
	Metadata        JSONBMap       `gorm:"type:jsonb" json:"metadata"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// IsSearchableNow 默认检索规则：发布且在有效期内
func (d *Document) IsSearchableNow(now time.Time) bool {
	if d == nil {
		return false
	}
	st := d.LifecycleStatus
	if st == "" {
		st = DocLifecyclePublished
	}
	if st != DocLifecyclePublished {
		return false
	}
	if d.EffectiveAt != nil && now.Before(*d.EffectiveAt) {
		return false
	}
	if d.ExpireAt != nil && !now.Before(*d.ExpireAt) {
		return false
	}
	return true
}

func (Document) TableName() string {
	return "documents"
}
