package bootstrap

import (
	"log/slog"

	"ai-for-oj/internal/config"
	"ai-for-oj/pkg/logx"
)

func NewLogger(cfg config.LogConfig) *slog.Logger {
	return logx.New(cfg.Level, cfg.Format)
}
