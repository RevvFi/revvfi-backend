package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/Revvfi/revvfi-backend/internal/logger"
)

// CorrelationID middleware extracts or generates correlation ID for request tracing
// This allows tracking requests across frontend -> backend -> indexer -> database
func CorrelationID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get correlation ID from request header
		correlationID := c.GetHeader("X-Correlation-ID")

		// Generate new correlation ID if not present
		if correlationID == "" {
			correlationID = logger.NewCorrelationID()
		}

		// Store in Gin context for use in handlers
		c.Set("correlation_id", correlationID)

		// Add correlation ID to context for logger
		ctx := logger.WithCorrelationID(c.Request.Context(), correlationID)
		c.Request = c.Request.WithContext(ctx)

		// Add to response header so frontend can track it
		c.Header("X-Correlation-ID", correlationID)

		c.Next()
	}
}
