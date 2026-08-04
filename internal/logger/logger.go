package logger

import (
	"io"
	"log/slog"
	"os"

	"github.com/oni1997/gatewayx/internal/config"
)

func New(cfg config.LoggingConfig) *slog.Logger {
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: level,
	}

	var output io.Writer
	switch cfg.Output {
	case "file":
		if cfg.File == "" {
			cfg.File = "gatewayx.log"
		}
		f, err := os.OpenFile(cfg.File, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			output = os.Stdout
		} else {
			output = f
		}
	default:
		output = os.Stdout
	}

	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(output, opts)
	} else {
		handler = slog.NewTextHandler(output, opts)
	}

	return slog.New(handler)
}
