package logger

import (
	"context"
	"log/slog"
	"os"
	"time"
)

var (
	defaultLogger *slog.Logger
)

type Service string

const (
	ServiceBackend Service = "backend"
	ServiceIndexer Service = "indexer"
)

type Environment string

const (
	EnvDevelopment Environment = "development"
	EnvStaging     Environment = "staging"
	EnvProduction  Environment = "production"
)

// Init initializes the global logger (call once at startup)
func Init(service Service, env Environment) {
	level := slog.LevelInfo
	if env == EnvDevelopment {
		level = slog.LevelDebug
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Rename "time" to "timestamp" for consistency
			if a.Key == slog.TimeKey {
				a.Key = "timestamp"
				// Format as ISO8601
				if t, ok := a.Value.Any().(time.Time); ok {
					a.Value = slog.StringValue(t.Format(time.RFC3339Nano))
				}
			}
			// Rename "msg" to "message" for consistency
			if a.Key == slog.MessageKey {
				a.Key = "message"
			}
			return a
		},
	})

	defaultLogger = slog.New(handler).With(
		slog.String("service", string(service)),
		slog.String("environment", string(env)),
	)
}

// Get returns the global logger instance
func Get() *slog.Logger {
	if defaultLogger == nil {
		// Initialize with defaults if not initialized
		Init(ServiceBackend, EnvDevelopment)
	}
	return defaultLogger
}

// FromContext returns a logger with correlation ID from context if present
func FromContext(ctx context.Context) *slog.Logger {
	logger := Get()

	// Extract correlation ID from context
	if correlationID, ok := ctx.Value(CorrelationIDKey).(string); ok && correlationID != "" {
		logger = logger.With(slog.String("correlation_id", correlationID))
	}

	return logger
}

// Helper functions for common log levels

// Debug logs a debug-level message
func Debug(msg string, args ...any) {
	Get().Debug(msg, args...)
}

// Info logs an info-level message
func Info(msg string, args ...any) {
	Get().Info(msg, args...)
}

// Warn logs a warning-level message
func Warn(msg string, args ...any) {
	Get().Warn(msg, args...)
}

// Error logs an error-level message
func Error(msg string, args ...any) {
	Get().Error(msg, args...)
}

// DebugContext logs a debug message with context
func DebugContext(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).Debug(msg, args...)
}

// InfoContext logs an info message with context
func InfoContext(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).Info(msg, args...)
}

// WarnContext logs a warning message with context
func WarnContext(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).Warn(msg, args...)
}

// ErrorContext logs an error message with context
func ErrorContext(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).Error(msg, args...)
}
