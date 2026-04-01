package logger

import (
	"log/slog"
	"os"
	"strings"
)

func Init() {
	opts := &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	}

	if isTerminal(os.Stdout) && os.Getenv("NO_COLOR") == "" {
		opts.ReplaceAttr = func(groups []string, a slog.Attr) slog.Attr {
			if a.Key != slog.LevelKey {
				return a
			}
			level, ok := a.Value.Any().(slog.Level)
			if !ok {
				return a
			}
			return slog.String(slog.LevelKey, colorLevel(level))
		}
	}

	handler := slog.NewTextHandler(os.Stdout, opts)
	slog.SetDefault(slog.New(handler))
}

func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

func colorLevel(level slog.Level) string {
	s := strings.ToUpper(level.String())
	switch {
	case level >= slog.LevelError:
		return "\x1b[31m" + s + "\x1b[0m"
	case level >= slog.LevelWarn:
		return "\x1b[33m" + s + "\x1b[0m"
	case level >= slog.LevelInfo:
		return "\x1b[32m" + s + "\x1b[0m"
	default:
		return "\x1b[90m" + s + "\x1b[0m"
	}
}
