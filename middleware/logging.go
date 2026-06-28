package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/Revvfi/revvfi-backend/internal/logger"
)

// RequestLogger middleware logs all HTTP requests with structured data
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Record start time
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Get correlation ID from context
		correlationID, _ := c.Get("correlation_id")

		// Get wallet address if authenticated (from auth middleware)
		walletAddress, _ := c.Get("wallet_address")

		// Log request start
		logger.Debug("HTTP request started",
			logger.WithRoute(path),
			logger.WithMethod(method),
			"correlation_id", correlationID,
			logger.WithWalletAddress(toString(walletAddress)),
		)

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)
		statusCode := c.Writer.Status()

		// Determine log level based on status code
		logFunc := logger.Info
		if statusCode >= 500 {
			logFunc = logger.Error
		} else if statusCode >= 400 {
			logFunc = logger.Warn
		}

		// Log request completion
		logFunc("HTTP request completed",
			logger.WithRoute(path),
			logger.WithMethod(method),
			logger.WithStatusCode(statusCode),
			logger.WithLatency(latency),
			"correlation_id", correlationID,
			logger.WithWalletAddress(toString(walletAddress)),
			"client_ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)
	}
}

// toString safely converts interface{} to string
func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if str, ok := v.(string); ok {
		return str
	}
	return ""
}
