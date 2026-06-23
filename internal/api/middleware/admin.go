package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Revvfi/revvfi-backend/internal/core/admin"
)

/*
@function AdminAuth

@desc
Creates admin authorization middleware for admin-only API routes.
Must be used AFTER Auth middleware to access wallet from context.

@responsibilities
- Read wallet from Gin context (set by Auth middleware)
- Verify wallet is an admin through AdminAuthService
- Reject unauthorized requests

@params
- adminAuthService: admin authentication service for role verification

@returns
- gin.HandlerFunc
*/
func AdminAuth(adminAuthService *admin.AdminAuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get wallet from context (set by Auth middleware)
		walletVal, exists := c.Get("wallet")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "authentication required",
				},
			})
			return
		}

		wallet, ok := walletVal.(string)
		if !ok || wallet == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "invalid wallet address",
				},
			})
			return
		}

		// Check if wallet is an admin
		isAdmin, err := adminAuthService.IsAdmin(c, wallet)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "failed to verify admin status",
				},
			})
			return
		}

		if !isAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "FORBIDDEN",
					"message": "admin access required",
				},
			})
			return
		}

		// Store admin status in context
		c.Set("is_admin", true)
		c.Next()
	}
}
