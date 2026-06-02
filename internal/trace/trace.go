package trace

import (
	"context"
	"fmt"
	"log/slog"
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
