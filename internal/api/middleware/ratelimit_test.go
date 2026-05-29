package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

/*
@function TestRateLimitRejectsExcessRequests

@desc
Tests that rate limiting rejects requests above the configured allowance.
*/
func TestRateLimitRejectsExcessRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RateLimit(1, time.Minute))
	router.GET("/ping", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	first := httptest.NewRecorder()
	router.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/ping", nil))
	if first.Code != http.StatusOK {
		t.Fatalf("expected first request 200, got %d", first.Code)
	}

	second := httptest.NewRecorder()
	router.ServeHTTP(second, httptest.NewRequest(http.MethodGet, "/ping", nil))
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second request 429, got %d", second.Code)
	}
}
