package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// =============================================================================
// 统一 API 响应结构
// =============================================================================

type ApiResponse struct {
	Code    int         `json:"code"`            // 业务状态码
	Message string      `json:"message"`         // 状态描述
	TraceID string      `json:"trace_id"`        // 请求追踪 ID
	Data    interface{} `json:"data,omitempty"`  // 业务数据
	Error   string      `json:"error,omitempty"` // 错误信息（失败时）
}

// =============================================================================
// 业务状态码定义
// =============================================================================

// 2xxx: 成功
const (
	CodeSuccess  = 2000 // 成功
	CodeCreated  = 2001 // 创建成功
	CodeUpdated  = 2002 // 更新成功
	CodeDeleted  = 2003 // 删除成功
	CodeAccepted = 2004 // 请求已接受（异步）
)

// 5xxx: 业务校验异常（参数、数据、权限等）
const (
	CodeParamInvalid      = 5001 // 参数无效
	CodeResourceNotFound  = 5002 // 资源不存在
	CodeResourceConflict  = 5003 // 资源冲突（如重复创建）
	CodeResourceDeleted   = 5004 // 资源已删除
	CodeUnauthorized      = 5005 // 未认证（需登录）
	CodeForbidden         = 5006 // 无权限访问
	CodeTokenInvalid      = 5007 // Token 无效
	CodeTokenExpired      = 5008 // Token 已过期
	CodeEmailExists       = 5009 // 邮箱已被注册
	CodeEmailNotExists    = 5010 // 邮箱不存在
	CodePasswordIncorrect = 5011 // 密码错误
	CodeUsernameExists    = 5012 // 用���名已被使用
	CodeVerifyCodeInvalid = 5013 // 验证码无效或已过期
	CodeRateLimitExceeded = 5014 // 请求频率超限
	CodeRequestTooLarge   = 5015 // 请求数据过大
)

// 4xxx: 系统运行异常（数据库、网络、服务不可用等）
const (
	CodeSystemError        = 4001 // 系统内部错误
	CodeDatabaseError      = 4002 // 数据库错误
	CodeCacheError         = 4003 // 缓存错误
	CodeNetworkError       = 4004 // 网络错误
	CodeServiceUnavailable = 4005 // 服务不可用
	CodeAIProviderError    = 4006 // AI 服务商错误
	CodeAIRequestTimeout   = 4007 // AI 请求超时
	CodeAIQuotaExhausted   = 4008 // AI 配额已用尽
	CodeModelNotSupport    = 4009 // 模型不支持
)

// =============================================================================
// Code 到 Message 的映射
// =============================================================================

var codeMessages = map[int]string{
	// 2xxx: 成功
	CodeSuccess:  "成功",
	CodeCreated:  "创建成功",
	CodeUpdated:  "更新成功",
	CodeDeleted:  "删除成功",
	CodeAccepted: "请求已接受",

	// 5xxx: 业务校验异常
	CodeParamInvalid:      "参数无效",
	CodeResourceNotFound:  "资源不存在",
	CodeResourceConflict:  "资源冲突",
	CodeResourceDeleted:   "资源已删除",
	CodeUnauthorized:      "请先登录",
	CodeForbidden:         "无权限访问",
	CodeTokenInvalid:      "Token 无效",
	CodeTokenExpired:      "Token 已过期",
	CodeEmailExists:       "该邮箱已被注册",
	CodeEmailNotExists:    "邮箱不存在",
	CodePasswordIncorrect: "密码错误",
	CodeUsernameExists:    "用户名已被使用",
	CodeVerifyCodeInvalid: "验证码无效或已过期",
	CodeRateLimitExceeded: "请求频率超限，请稍后重试",
	CodeRequestTooLarge:   "请求数据过大",

	// 4xxx: 系统运行异常
	CodeSystemError:        "系统内部错误",
	CodeDatabaseError:      "数据库错误",
	CodeCacheError:         "缓存错误",
	CodeNetworkError:       "网络错误",
	CodeServiceUnavailable: "服务暂时不可用",
	CodeAIProviderError:    "AI 服务暂时不可用",
	CodeAIRequestTimeout:   "AI 请求超时",
	CodeAIQuotaExhausted:   "AI 配额已用尽",
	CodeModelNotSupport:    "不支持的模型",
}

// GetMessage 根据 code 获取默认消息
func GetMessage(code int) string {
	if msg, ok := codeMessages[code]; ok {
		return msg
	}
	return "未知错误"
}

// =============================================================================
// 响应辅助函数
// =============================================================================

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
		Code:    CodeSuccess,
		Message: "创建成功",
		TraceID: generateTraceID(),
		Data:    data,
	})
}

// Accepted 返回异步接受响应
func Accepted(c *gin.Context, data interface{}) {
	c.JSON(http.StatusAccepted, ApiResponse{
		Code:    CodeSuccess,
		Message: "请求已接受",
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

// FailWithCode 返回失败响应（使用 HTTP status + 业务 code）
func FailWithCode(c *gin.Context, httpStatus int, bizCode int, errMsg string) {
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
func FailBadRequest(c *gin.Context, errMsg string) {
	FailWithCode(c, http.StatusBadRequest, CodeParamInvalid, errMsg)
}

// Unauthorized 返回未认证
func FailUnauthorized(c *gin.Context, errMsg string) {
	FailWithCode(c, http.StatusUnauthorized, CodeUnauthorized, errMsg)
}

// Forbidden 返回无权限
func FailForbidden(c *gin.Context, errMsg string) {
	FailWithCode(c, http.StatusForbidden, CodeForbidden, errMsg)
}

// NotFound 返回资源不存在
func FailNotFound(c *gin.Context, errMsg string) {
	FailWithCode(c, http.StatusNotFound, CodeResourceNotFound, errMsg)
}

// InternalError 返回系统错误
func FailInternal(c *gin.Context, errMsg string) {
	FailWithCode(c, http.StatusInternalServerError, CodeSystemError, errMsg)
}

// --- Model types ---

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

type JSONBArray []string

func (j *JSONBArray) Scan(value any) error {
	if value == nil {
		*j = JSONBArray{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan JSONBArray: not a byte slice")
	}
	return json.Unmarshal(bytes, j)
}

func (j JSONBArray) Value() (driver.Value, error) {
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
		return errors.New("failed to scan JSONBMap: not a byte slice")
	}
	return json.Unmarshal(bytes, j)
}

func (j JSONBMap) Value() (driver.Value, error) {
	if j == nil {
		return "{}", nil
	}
	return json.Marshal(j)
}
