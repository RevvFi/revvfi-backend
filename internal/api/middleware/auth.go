package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Revvfi/revvfi-backend/internal/core/auth"
)

/*
@function Auth

@desc
Creates JWT authentication middleware for protected API routes.

@responsibilities
- Read bearer token from Authorization header
- Validate token through auth service
- Store wallet and token on Gin context
- Reject unauthenticated requests

@params
- authService: authentication service used for token validation

@returns
- gin.HandlerFunc
*/
func Auth(authService *auth.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := strings.TrimSpace(c.GetHeader("Authorization"))
		if header == "" || !strings.HasPrefix(strings.ToLower(header), "bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "missing bearer token"}})
			return
		}

		token := strings.TrimSpace(header[len("Bearer "):])
		wallet, err := authService.ValidateToken(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "INVALID_TOKEN", "message": err.Error()}})
			return
		}

		c.Set("wallet", wallet)
		c.Set("token", token)
		c.Next()
	}
}
