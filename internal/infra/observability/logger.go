package observability

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	JobIDKey     contextKey = "job_id"
	TraceIDKey   contextKey = "trace_id"
)

var Logger *slog.Logger

func InitLogger(levelStr string, format string) *slog.Logger {
	var level slog.Level
	switch strings.ToLower(levelStr) {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if strings.ToLower(format) == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	Logger = slog.New(handler)
	slog.SetDefault(Logger)
	return Logger
}

// WithContext returns a logger decorated with context identifiers (request_id, job_id, trace_id)
func WithContext(ctx context.Context) *slog.Logger {
	if Logger == nil {
		InitLogger("info", "json")
	}

	logger := Logger
	if ctx == nil {
		return logger
	}

	if reqID, ok := ctx.Value(RequestIDKey).(string); ok && reqID != "" {
		logger = logger.With("request_id", reqID)
	}
	if jobID, ok := ctx.Value(JobIDKey).(string); ok && jobID != "" {
		logger = logger.With("job_id", jobID)
	}
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok && traceID != "" {
		logger = logger.With("trace_id", traceID)
	}

	return logger
}
