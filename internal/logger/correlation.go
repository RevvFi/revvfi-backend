package logger

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

// CorrelationIDKey is the context key for correlation ID
const CorrelationIDKey contextKey = "correlation_id"

// NewCorrelationID generates a new UUID v4 correlation ID
func NewCorrelationID() string {
	return uuid.New().String()
}

// WithCorrelationID adds a correlation ID to the context
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, CorrelationIDKey, correlationID)
}

// GetCorrelationID extracts correlation ID from context
func GetCorrelationID(ctx context.Context) string {
	if correlationID, ok := ctx.Value(CorrelationIDKey).(string); ok {
		return correlationID
	}
	return ""
}

// WithCorrelationIDAttr creates a correlation ID attribute for logging (not used - use slog.String directly)
// Deprecated: Use slog.String("correlation_id", correlationID) instead
func WithCorrelationIDAttr(correlationID string) any {
	return correlationID
}
