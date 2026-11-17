package trace

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Key traceId 在 context 中的 key
type Key string

const (
	// TraceIDKey traceId 在 Gin context 中的 key
	TraceIDKey Key = "trace_id"
	// TraceIDHeader 外部传入 traceId 的 header 名称
	TraceIDHeader = "X-Trace-ID"
	// ResponseTraceIDHeader 响应中返回 traceId 的 header 名称
	ResponseTraceIDHeader = "X-Trace-ID"
)

// GetTraceIDFromContext 从标准 context 获取 traceId（用于 slog 等工具）
func GetTraceIDFromContext(ctx context.Context) string {
	if traceID := ctx.Value(TraceIDKey); traceID != nil {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return ""
}

// GetTraceIDFromGin 从 Gin context 获取 traceId
func GetTraceIDFromGin(c *gin.Context) string {
	if traceID, exists := c.Get(string(TraceIDKey)); exists {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return ""
}

// GenerateTraceID 生成新的 traceId（唯一性更好：UUID前8位+时间戳后6位）
func GenerateTraceID() string {
	return fmt.Sprintf("%s%d", uuid.New().String()[:8], time.Now().UnixMilli()%1000000)
}

// GetTraceIDOrGenerate 从 Gin context 获取 traceId，优先从 context 获取，否则生成新的并打印警告
func GetTraceIDOrGenerate(c *gin.Context) string {
	if traceID, exists := c.Get(string(TraceIDKey)); exists {
		if id, ok := traceID.(string); ok && id != "" {
			return id
		}
	}
	slog.Warn("traceId not found in context, generating new one")
	return GenerateTraceID()
}

// Transport 是 http.RoundTripper 实现，自动从 context 提取 trace ID 并注入到请求 header 中。
// 所有通过它发出的 HTTP 请求都会自动携带 X-Trace-ID，无需在每个调用点手动处理。
type Transport struct {
	Base http.RoundTripper
}

// RoundTrip 实现 http.RoundTripper 接口，在每个请求发出前自动注入 trace header。
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if traceID := GetTraceIDFromContext(req.Context()); traceID != "" {
		req.Header.Set(TraceIDHeader, traceID)
	}

	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(req)
}

// NewHTTPClient 创建一个自动携带 trace ID 的 HTTP Client。
// 当 req.Context() 中包含 trace ID 时，所有请求会自动注入 X-Trace-ID header。
// 用法：
//
//	client := trace.NewHTTPClient()
//	req, _ := http.NewRequestWithContext(ctx, "GET", "http://example.com", nil)
//	resp, err := client.Do(req)  // 自动携带 trace ID
func NewHTTPClient() *http.Client {
	return &http.Client{
		Transport: &Transport{},
		Timeout:   60 * time.Second,
	}
}

// InjectFromGin 将 Gin context 中的 trace ID 注入到标准 context 中。
// 用于从 Gin Handler 传入 service 层时，确保 trace 信息在 context 中传递。
func InjectFromGin(c *gin.Context) context.Context {
	if traceID := GetTraceIDFromGin(c); traceID != "" {
		return context.WithValue(c.Request.Context(), TraceIDKey, traceID)
	}
	return c.Request.Context()
}

// InjectTrace 将指定 traceID 注入到 context 并返回新的 context。
func InjectTrace(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// GetTraceIDOrGenerateFromContext 从 context 获取 trace ID，没有则生成新的。
func GetTraceIDOrGenerateFromContext(ctx context.Context) string {
	if traceID := GetTraceIDFromContext(ctx); traceID != "" {
		return traceID
	}
	return GenerateTraceID()
}

// LogRequest 记录 HTTP 请求日志（包含 trace ID）。
func LogRequest(req *http.Request, body []byte) {
	traceID := GetTraceIDFromContext(req.Context())
	if body != nil {
		slog.Info("outbound request",
			"method", req.Method,
			"url", req.URL.String(),
			"trace_id", traceID,
			"body_size", len(body),
		)
	} else {
		slog.Info("outbound request",
			"method", req.Method,
			"url", req.URL.String(),
			"trace_id", traceID,
		)
	}
}

// LogResponse 记录 HTTP 响应日志（包含 trace ID）。
func LogResponse(req *http.Request, statusCode int, body []byte) {
	traceID := GetTraceIDFromContext(req.Context())
	if body != nil {
		slog.Info("outbound response",
			"method", req.Method,
			"url", req.URL.String(),
			"status", statusCode,
			"trace_id", traceID,
			"body_size", len(body),
		)
	} else {
		slog.Info("outbound response",
			"method", req.Method,
			"url", req.URL.String(),
			"status", statusCode,
			"trace_id", traceID,
		)
	}
}

// ReadBody 读取 HTTP 响应 body，恢复 body 后返回内容。
func ReadBody(resp *http.Response) []byte {
	if resp == nil || resp.Body == nil {
		return nil
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	resp.Body = io.NopCloser(
		bytes.NewBuffer(body),
	)
	return body
}
