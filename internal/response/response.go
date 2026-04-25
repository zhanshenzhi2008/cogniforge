package response

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	CodeModelNotSupport    = 4009 // 不支持的模型
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
	CodeUsernameExists    = 5012 // 用户名已被使用
	CodeVerifyCodeInvalid = 5013 // 验证码无效或已过期
	CodeRateLimitExceeded = 5014 // 请求频率超限
	CodeRequestTooLarge   = 5015 // 请求数据过大
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
}

// GetMessage 根据 code 获取默认消息
func GetMessage(code int) string {
	if msg, ok := codeMessages[code]; ok {
		return msg
	}
	return "未知错误"
}

// IsSuccess 判断是否为成功响应
func IsSuccess(code int) bool {
	return code >= 2000 && code < 3000
}

// IsBizError 判断是否为业务校验异常
func IsBizError(code int) bool {
	return code >= 5000 && code < 6000
}

// IsSysError 判断是否为系统运行异常
func IsSysError(code int) bool {
	return code >= 4000 && code < 5000
}

// =============================================================================
// Helper functions
// =============================================================================

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

// Conflict 返回资源冲突
func Conflict(c *gin.Context, errMsg string) {
	FailWithHTTPStatus(c, http.StatusConflict, CodeResourceConflict, errMsg)
}

// ToJSON 将 ApiResponse 转换为 JSON 字符串
func (r ApiResponse) ToJSON() (string, error) {
	bytes, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
