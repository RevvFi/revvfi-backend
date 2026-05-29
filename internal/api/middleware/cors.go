package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

/*
@function CORS

@desc
Creates CORS middleware with configured origins, methods, and headers.

@responsibilities
- Match request origin against allowed origins
- Set browser CORS headers
- Short-circuit preflight requests

@params
- origins: allowed origins
- methods: allowed HTTP methods
- headers: allowed HTTP headers

@returns
- gin.HandlerFunc
*/
func CORS(origins []string, methods []string, headers []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(origins))
	allowAll := false
	for _, origin := range origins {
		if origin == "*" {
			allowAll = true
		}
		allowed[origin] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if allowAll {
			c.Header("Access-Control-Allow-Origin", "*")
		} else if _, ok := allowed[origin]; ok {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}

		c.Header("Access-Control-Allow-Methods", strings.Join(methods, ", "))
		c.Header("Access-Control-Allow-Headers", strings.Join(headers, ", "))
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
