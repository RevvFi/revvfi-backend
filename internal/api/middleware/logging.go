package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

/*
@function Logging

@desc
Creates structured request logging middleware using the standard logger.

@responsibilities
- Record request method, path, status, latency, and client IP
- Keep logging dependency-free for the current backend phase

@returns
- gin.HandlerFunc
*/
func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Printf("method=%s path=%s status=%d latency=%s ip=%s", c.Request.Method, c.Request.URL.Path, c.Writer.Status(), time.Since(start), c.ClientIP())
	}
}
