package model

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// =============================================================================
// Code 到 Message 的映射
// =============================================================================

// Helper functions

// generateTraceID 生成追踪 ID
func generateTraceID() string {
	return fmt.Sprintf("%s-%d", uuid.New().String()[:8], time.Now().UnixMilli()%1000000)
}

// Success 返回成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, ApiResponse{
		Code:    CodeSuccess,
		Message: GetMessage(CodeSuccess),
		TraceID: generateTraceID(),
		Data:    data,
	})
}

// SuccessWithMessage 返回成功响应（自定义消息）
func SuccessWithMessage(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusOK, ApiResponse{
		Code:    CodeSuccess,
		Message: message,
		TraceID: generateTraceID(),
		Data:    data,
	})
}

// Created 返回创建成功响应
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, ApiResponse{
		Code:    CodeCreated,
		Message: GetMessage(CodeCreated),
		TraceID: generateTraceID(),
		Data:    data,
	})
}

// Accepted 返回异步接受响应
func Accepted(c *gin.Context, data interface{}) {
	c.JSON(http.StatusAccepted, ApiResponse{
		Code:    CodeAccepted,
		Message: GetMessage(CodeAccepted),
		TraceID: generateTraceID(),
		Data:    data,
	})
}

// Fail 返回失败响应（使用业务 code）
func Fail(c *gin.Context, code int, errMsg string) {
	traceID := generateTraceID()
	message := errMsg
	if errMsg == "" {
		message = GetMessage(code)
	}
	c.JSON(http.StatusOK, ApiResponse{
		Code:    code,
		Message: message,
		TraceID: traceID,
		Error:   errMsg,
	})
}

// FailWithHTTPStatus 返回失败响应（使用 HTTP status + 业务 code）
func FailWithHTTPStatus(c *gin.Context, httpStatus int, bizCode int, errMsg string) {
	traceID := generateTraceID()
	message := errMsg
	if errMsg == "" {
		message = GetMessage(bizCode)
	}
	c.JSON(httpStatus, ApiResponse{
		Code:    bizCode,
		Message: message,
		TraceID: traceID,
		Error:   errMsg,
	})
}

// BadRequest 返回参数错误
func BadRequest(c *gin.Context, errMsg string) {
	FailWithHTTPStatus(c, http.StatusBadRequest, CodeParamInvalid, errMsg)
}

// Unauthorized 返回未认证
func Unauthorized(c *gin.Context, errMsg string) {
	FailWithHTTPStatus(c, http.StatusUnauthorized, CodeUnauthorized, errMsg)
}

// Forbidden 返回无权限
func Forbidden(c *gin.Context, errMsg string) {
	FailWithHTTPStatus(c, http.StatusForbidden, CodeForbidden, errMsg)
}

// NotFound 返回资源不存在
func NotFound(c *gin.Context, errMsg string) {
	FailWithHTTPStatus(c, http.StatusNotFound, CodeResourceNotFound, errMsg)
}

// InternalError 返回系统错误
func InternalError(c *gin.Context, errMsg string) {
	FailWithHTTPStatus(c, http.StatusInternalServerError, CodeSystemError, errMsg)
}

// =============================================================================
// Model types
// =============================================================================

type User struct {
	ID        string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	Email     string         `gorm:"uniqueIndex;type:varchar(255);not null" json:"email"`
	Name      string         `gorm:"type:varchar(100);not null" json:"name"`
	Password  string         `gorm:"type:varchar(255);not null" json:"-"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string {
	return "users"
}

type ApiKey struct {
	ID        string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	UserID    string         `gorm:"type:varchar(64);not null;index" json:"user_id"`
	Name      string         `gorm:"type:varchar(255);not null" json:"name"`
	Key       string         `gorm:"type:varchar(255);not null" json:"key"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ApiKey) TableName() string {
	return "api_keys"
}

type Agent struct {
	ID           string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	UserID       string         `gorm:"type:varchar(64);not null;index" json:"user_id"`
	Name         string         `gorm:"type:varchar(255);not null" json:"name"`
	Description  string         `gorm:"type:text" json:"description"`
	Model        string         `gorm:"type:varchar(100)" json:"model"`
	SystemPrompt string         `gorm:"type:text" json:"system_prompt"`
	Tools        JSONBArray     `gorm:"type:jsonb" json:"tools"`
	MemoryType   string         `gorm:"type:varchar(50)" json:"memory_type"`
	MemoryTurns  int            `gorm:"default:10" json:"memory_turns"`
	InputFilter  bool           `gorm:"default:true" json:"input_filter"`
	OutputFilter bool           `gorm:"default:true" json:"output_filter"`
	Status       string         `gorm:"type:varchar(50);default:'active'" json:"status"`
	Metadata     JSONBMap       `gorm:"type:jsonb" json:"metadata"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Agent) TableName() string {
	return "agents"
}

type Workflow struct {
	ID          string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	UserID      string         `gorm:"type:varchar(64);not null;index" json:"user_id"`
	Name        string         `gorm:"type:varchar(255);not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Definition  string         `gorm:"type:text" json:"definition"`
	Status      string         `gorm:"type:varchar(50);default:'draft'" json:"status"`
	Version     int            `gorm:"default:1" json:"version"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Workflow) TableName() string {
	return "workflows"
}

type WorkflowNode struct {
	ID         string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	WorkflowID string         `gorm:"type:varchar(64);not null;index" json:"workflow_id"`
	Type       string         `gorm:"type:varchar(50)" json:"type"`
	Name       string         `gorm:"type:varchar(255)" json:"name"`
	Config     string         `gorm:"type:text" json:"config"`
	PositionX  float64        `json:"position_x"`
	PositionY  float64        `json:"position_y"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (WorkflowNode) TableName() string {
	return "workflow_nodes"
}

type WorkflowEdge struct {
	ID         string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	WorkflowID string         `gorm:"type:varchar(64);not null;index" json:"workflow_id"`
	SourceID   string         `gorm:"type:varchar(64)" json:"source_id"`
	TargetID   string         `gorm:"type:varchar(64)" json:"target_id"`
	Config     string         `gorm:"type:text" json:"config"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (WorkflowEdge) TableName() string {
	return "workflow_edges"
}

type WorkflowExecution struct {
	ID          string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	WorkflowID  string         `gorm:"type:varchar(64);not null;index" json:"workflow_id"`
	UserID      string         `gorm:"type:varchar(64);not null;index" json:"user_id"`
	Status      string         `gorm:"type:varchar(50);default:'pending'" json:"status"`
	Input       JSONBMap       `gorm:"type:jsonb" json:"input"`
	Output      string         `gorm:"type:text" json:"output"`
	Error       string         `gorm:"type:text" json:"error"`
	StartedAt   *time.Time     `json:"started_at"`
	CompletedAt *time.Time     `json:"completed_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (WorkflowExecution) TableName() string {
	return "workflow_executions"
}

// =============================================================================
// Knowledge Base Models
// =============================================================================

type KnowledgeBase struct {
	ID          string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	UserID      string         `gorm:"type:varchar(64);not null;index" json:"user_id"`
	Name        string         `gorm:"type:varchar(255);not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	VectorDB    string         `gorm:"type:varchar(50);default:'chroma'" json:"vector_db"` // chroma, qdrant, weaviate
	EmbeddingModel string     `gorm:"type:varchar(100);default:'text-embedding-ada-002'" json:"embedding_model"`
	Status      string         `gorm:"type:varchar(50);default:'active'" json:"status"`
	Metadata    JSONBMap       `gorm:"type:jsonb" json:"metadata"`
	DocCount    int            `gorm:"default:0" json:"doc_count"`     // 文档数量
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (KnowledgeBase) TableName() string {
	return "knowledge_bases"
}

type Document struct {
	ID             string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	KnowledgeBaseID string        `gorm:"type:varchar(64);not null;index" json:"knowledge_base_id"`
	UserID         string         `gorm:"type:varchar(64);not null;index" json:"user_id"`
	Name           string         `gorm:"type:varchar(255);not null" json:"name"`
	FileName       string         `gorm:"type:varchar(255)" json:"file_name"` // 原始文件名
	FileSize       int64         `gorm:"type:bigint" json:"file_size"`      // 文件大小(bytes)
	FileType       string        `gorm:"type:varchar(50)" json:"file_type"` // pdf, txt, md, docx
	FilePath       string         `gorm:"type:text" json:"file_path"`       // 存储路径
	Status         string         `gorm:"type:varchar(50);default:'pending'" json:"status"` // pending, processing, completed, failed
	Error          string         `gorm:"type:text" json:"error"`            // 错误信息
	ChunkCount     int            `gorm:"default:0" json:"chunk_count"`     // 分块数量
	VectorCount    int            `gorm:"default:0" json:"vector_count"`    // 向量数量
	Metadata       JSONBMap       `gorm:"type:jsonb" json:"metadata"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Document) TableName() string {
	return "documents"
}

// =============================================================================
// JSONB types
// =============================================================================

type JSONBArray []string

func (j *JSONBArray) Scan(value any) error {
	if value == nil {
		*j = JSONBArray{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan JSONBArray: not a byte slice")
	}
	return json.Unmarshal(bytes, j)
}

func (j JSONBArray) Value() (any, error) {
	if j == nil {
		return "[]", nil
	}
	return json.Marshal(j)
}

type JSONBMap map[string]any

func (j *JSONBMap) Scan(value any) error {
	if value == nil {
		*j = JSONBMap{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan JSONBMap: not a byte slice")
	}
	return json.Unmarshal(bytes, j)
}

func (j JSONBMap) Value() (any, error) {
	if j == nil {
		return "{}", nil
	}
	return json.Marshal(j)
}
