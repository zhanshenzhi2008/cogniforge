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
	AvatarURL string         `gorm:"type:varchar(500)" json:"avatar_url"`             // 头像地址
	Status    string         `gorm:"type:varchar(50);default:'active'" json:"status"` // active, disabled, locked
	Role      string         `gorm:"type:varchar(100);default:'user'" json:"role"`    // 用户角色（简单方案）
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string {
	return "users"
}

// =============================================================================
// RBAC Models - 用户管理与权限系统
// =============================================================================

// UserSettings 用户设置表
type UserSettings struct {
	ID        string    `gorm:"primaryKey;type:varchar(64)" json:"id"`
	UserID    string    `gorm:"type:varchar(64);not null;uniqueIndex" json:"user_id"`
	AvatarURL string    `gorm:"type:varchar(500)" json:"avatar_url"`           // 头像地址
	Theme     string    `gorm:"type:varchar(50);default:'light'" json:"theme"` // 主题：light/dark
	Language  string    `gorm:"type:varchar(20);default:'zh-CN'" json:"language"`
	Timezone  string    `gorm:"type:varchar(50);default:'Asia/Shanghai'" json:"timezone"`
	Metadata  JSONBMap  `gorm:"type:jsonb" json:"metadata"` // 其他设置（通知偏好等）
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (UserSettings) TableName() string {
	return "user_settings"
}

// UserSession 用户会话表（记录登录设备）
type UserSession struct {
	ID        string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	UserID    string         `gorm:"type:varchar(64);not null;index" json:"user_id"`
	TokenID   string         `gorm:"type:varchar(255);not null;uniqueIndex" json:"token_id"` // JWT ID (jti)
	UserAgent string         `gorm:"type:varchar(500)" json:"user_agent"`                    // 用户代理
	IPAddress string         `gorm:"type:varchar(50)" json:"ip_address"`                     // IP 地址
	Device    string         `gorm:"type:varchar(100)" json:"device"`                        // 设备信息
	Location  string         `gorm:"type:varchar(255)" json:"location"`                      // 登录地点
	ExpiresAt time.Time      `json:"expires_at"`                                             // 过期时间
	LastUsed  time.Time      `json:"last_used"`                                              // 最后使用时间
	IsActive  bool           `gorm:"default:true" json:"is_active"`                          // 是否活跃
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (UserSession) TableName() string {
	return "user_sessions"
}

// Permission 权限点表
type Permission struct {
	ID        string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	Code      string         `gorm:"type:varchar(100);not null;uniqueIndex" json:"code"` // 权限代码（如：user:create, role:edit）
	Name      string         `gorm:"type:varchar(255);not null" json:"name"`             // 权限名称
	Group     string         `gorm:"type:varchar(100)" json:"group"`                     // 分组（如：用户管理、角色管理）
	Desc      string         `gorm:"type:text" json:"description"`                       // 描述
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Permission) TableName() string {
	return "permissions"
}

// Role 角色表
type Role struct {
	ID          string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`             // 角色名称
	Code        string         `gorm:"type:varchar(100);not null;uniqueIndex" json:"code"` // 角色代码（如：admin, user）
	Description string         `gorm:"type:text" json:"description"`                       // 描述
	IsSystem    bool           `gorm:"default:false" json:"is_system"`                     // 是否系统预置角色（不可删除）
	IsDefault   bool           `gorm:"default:false" json:"is_default"`                    // 是否默认角色
	Permissions string         `gorm:"type:text" json:"-"`                                 // 权限列表（缓存用，不直接查询）
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Role) TableName() string {
	return "roles"
}

// RolePermission 角色权限关联表（多对多）
type RolePermission struct {
	ID           string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	RoleID       string         `gorm:"type:varchar(64);not null;index" json:"role_id"`
	PermissionID string         `gorm:"type:varchar(64);not null;index" json:"permission_id"`
	CreatedAt    time.Time      `json:"created_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (RolePermission) TableName() string {
	return "role_permissions"
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
	CurrentNode string         `gorm:"type:varchar(64)" json:"current_node"` // 当前正在执行的节点（用于调试）
	StartedAt   *time.Time     `json:"started_at"`
	CompletedAt *time.Time     `json:"completed_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (WorkflowExecution) TableName() string {
	return "workflow_executions"
}

type WorkflowSchedule struct {
	ID             string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	WorkflowID     string         `gorm:"type:varchar(64);not null;index" json:"workflow_id"`
	UserID         string         `gorm:"type:varchar(64);not null;index" json:"user_id"`
	Name           string         `gorm:"type:varchar(255)" json:"name"`
	CronExpression string         `gorm:"type:varchar(100)" json:"cron_expression"`
	IsActive       bool           `gorm:"default:true" json:"is_active"`
	DefaultInput   JSONBMap       `gorm:"type:jsonb" json:"default_input"`
	LastRun        *time.Time     `json:"last_run"`
	LastError      string         `gorm:"type:text" json:"last_error"`
	NextRun        *time.Time     `json:"next_run"`
	RunCount       int            `gorm:"default:0" json:"run_count"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (WorkflowSchedule) TableName() string {
	return "workflow_schedules"
}

// =============================================================================
// Knowledge Base Models
// =============================================================================

type KnowledgeBase struct {
	ID             string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	UserID         string         `gorm:"type:varchar(64);not null;index" json:"user_id"`
	Name           string         `gorm:"type:varchar(255);not null" json:"name"`
	Description    string         `gorm:"type:text" json:"description"`
	VectorDB       string         `gorm:"type:varchar(50);default:'chroma'" json:"vector_db"` // chroma, qdrant, weaviate
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

// =============================================================================
// Monitor Models - 请求日志
// =============================================================================

type RequestLog struct {
	ID           string    `gorm:"primaryKey;type:varchar(64)" json:"id"`
	TraceID      string    `gorm:"type:varchar(64);index" json:"trace_id"`
	UserID       string    `gorm:"type:varchar(64);index" json:"user_id"`
	Method       string    `gorm:"type:varchar(10);not null" json:"method"`    // GET, POST, PUT, DELETE
	Path         string    `gorm:"type:varchar(500);not null" json:"path"`     // 请求路径
	Query        string    `gorm:"type:text" json:"query"`                     // 查询参数
	RequestBody  string    `gorm:"type:text" json:"request_body"`              // 请求体
	StatusCode   int       `gorm:"type:smallint;default:0" json:"status_code"` // 响应状态码
	ResponseBody string    `gorm:"type:text" json:"response_body"`             // 响应体
	Duration     int64     `gorm:"type:bigint;default:0" json:"duration"`      // 耗时(毫秒)
	UserAgent    string    `gorm:"type:varchar(500)" json:"user_agent"`        // 用户代理
	IP           string    `gorm:"type:varchar(50)" json:"ip"`                 // IP 地址
	Error        string    `gorm:"type:text" json:"error"`                     // 错误信息
	CreatedAt    time.Time `json:"created_at"`
}

func (RequestLog) TableName() string {
	return "request_logs"
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
