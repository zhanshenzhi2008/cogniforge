package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"cogniforge/internal/trace"
)

type consoleHandler struct {
	opts slog.HandlerOptions
	w    io.Writer
}

func newHandler(w io.Writer) *consoleHandler {
	return &consoleHandler{
		opts: slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: true,
		},
		w: w,
	}
}

func (h *consoleHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

func (h *consoleHandler) Handle(ctx context.Context, r slog.Record) error {
	var buf []byte

	// 时间
	buf = fmt.Appendf(buf, "%s ", r.Time.Format("2006-01-02 15:04:05"))

	// 级别（彩色）
	buf = append(buf, colorLevel(r.Level)...)
	buf = append(buf, ' ')

	// 来源
	if r.PC != 0 {
		src := r.Source()
		buf = fmt.Appendf(buf, "%s:%d ", src.File, src.Line)
	}

	// traceId
	if traceID := trace.GetTraceIDFromContext(ctx); traceID != "" {
		buf = fmt.Appendf(buf, "[%s] ", traceID)
	}

	// 消息
	buf = append(buf, r.Message...)

	// attrs
	r.Attrs(func(a slog.Attr) bool {
		buf = fmt.Appendf(buf, " %s=%v", a.Key, a.Value)
		return true
	})

	buf = append(buf, '\n')
	_, err := h.w.Write(buf)
	return err
}

func (h *consoleHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *consoleHandler) WithGroup(string) slog.Handler      { return h }

func colorLevel(level slog.Level) string {
	switch level {
	case slog.LevelDebug:
		return "\033[36mDEBUG\033[0m"
	case slog.LevelInfo:
		return "\033[32mINFO \033[0m"
	case slog.LevelWarn:
		return "\033[33mWARN \033[0m"
	case slog.LevelError:
		return "\033[31mERROR\033[0m"
	default:
		return level.String()
	}
}

// Init initializes the logger
func Init() {
	slog.SetDefault(slog.New(newHandler(os.Stdout)))
}
