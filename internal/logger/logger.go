package logger

import (
	"context"
	"io"
	"log/slog"
	"os"

	"cogniforge/internal/trace"
)

func Init() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	slog.SetDefault(slog.New(handler))
}

// InitWithTraceID 初始化日志，使用自定义 handler 支持 traceId
// 当 context 中包含 traceId 时，日志会自动包含该字段
func InitWithTraceID() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := NewTraceIDHandler(os.Stdout, opts)
	slog.SetDefault(slog.New(handler))
}

// TraceIDHandler 带 traceId 支持的 slog handler
type TraceIDHandler struct {
	*slog.JSONHandler
}

// NewTraceIDHandler 创建支持 traceId 的 handler
func NewTraceIDHandler(w io.Writer, opts *slog.HandlerOptions) *TraceIDHandler {
	return &TraceIDHandler{
		JSONHandler: slog.NewJSONHandler(w, opts),
	}
}

// Handle 重写 Handle 方法，自动从 context 中获取 traceId 并添加到日志
func (h *TraceIDHandler) Handle(ctx context.Context, r slog.Record) error {
	// 从 context 中获取 traceId
	if traceID := trace.GetTraceIDFromContext(ctx); traceID != "" {
		r.Add("trace_id", traceID)
	}
	return h.JSONHandler.Handle(ctx, r)
}
