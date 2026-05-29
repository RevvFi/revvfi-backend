package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

/*
@function Recovery

@desc
Creates panic recovery middleware returning a stable JSON error payload.

@responsibilities
- Recover panics from downstream handlers
- Return internal error JSON
- Prevent server process crashes from request panics

@returns
- gin.HandlerFunc
*/
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "internal server error"}})
	})
}
