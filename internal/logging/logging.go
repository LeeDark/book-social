package logging

import (
	"log/slog"
	"os"
	"strings"
)

func New(env, level, format string) *slog.Logger {
	lvl := parseLevel(level)

	opts := &slog.HandlerOptions{
		Level:     lvl,
		AddSource: env == "dev",
	}

	var handler slog.Handler

	switch strings.ToLower(format) {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
