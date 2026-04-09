package logx

import (
	"log/slog"
	"os"
	"strings"
)

func New(level, format string) *slog.Logger {
	handlerOptions := &slog.HandlerOptions{
		Level: parseLevel(level),
	}

	switch strings.ToLower(format) {
	case "json":
		return slog.New(slog.NewJSONHandler(os.Stdout, handlerOptions))
	default:
		return slog.New(slog.NewTextHandler(os.Stdout, handlerOptions))
	}
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
