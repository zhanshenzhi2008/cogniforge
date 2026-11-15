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
	UserID         string         `gorm:"type:varchar(64);not null;index" json:"user_id"`
	FolderID       string         `gorm:"type:varchar(64);index:idx_kb_folder_id" json:"folder_id"` // 空=未分类
	Name           string         `gorm:"type:varchar(255);not null" json:"name"`
	Description    string         `gorm:"type:text" json:"description"`
	VectorDB       string         `gorm:"type:varchar(50);default:'pgvector'" json:"vector_db"` // pgvector, chroma, qdrant, weaviate
	EmbeddingModel string         `gorm:"type:varchar(100);default:'text-embedding-ada-002'" json:"embedding_model"`
	Status         string         `gorm:"type:varchar(50);default:'active'" json:"status"`
	Metadata       JSONBMap       `gorm:"type:jsonb" json:"metadata"`
	DocCount       int            `gorm:"default:0" json:"doc_count"` // 文档数量
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (KnowledgeBase) TableName() string {
	return "knowledge_bases"
}

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
	Status          string         `gorm:"type:varchar(50);default:'pending'" json:"status"` // pending, processing, completed, failed
	Error           string         `gorm:"type:text" json:"error"`                           // 错误信息
	ChunkCount      int            `gorm:"default:0" json:"chunk_count"`                     // 分块数量
	VectorCount     int            `gorm:"default:0" json:"vector_count"`                    // 向量数量
	Metadata        JSONBMap       `gorm:"type:jsonb" json:"metadata"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Document) TableName() string {
	return "documents"
}
