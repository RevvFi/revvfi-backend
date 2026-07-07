package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

/*
@struct rateBucket

@desc
Tracks request count for one client inside a fixed rate-limit window.

@responsibilities
- Store request count
- Store window reset time
*/
type rateBucket struct {
	count   int
	resetAt time.Time
}

/*
@function RateLimit

@desc
Creates simple in-memory IP-based rate limiting middleware.

@responsibilities
- Track request counts by client IP
- Reset counters after configured window
- Reject clients exceeding the request allowance

@params
- maxRequests: maximum requests per window
- window: rate-limit window duration

@returns
- gin.HandlerFunc
*/
func RateLimit(maxRequests int, window time.Duration) gin.HandlerFunc {
	var mu sync.Mutex
	buckets := make(map[string]*rateBucket)

	// Buckets are only ever added, never removed, so a long-lived process
	// accumulates one entry per distinct client IP forever. Periodically
	// drop entries whose window has already expired - they'd be replaced on
	// next access anyway (see the now.After(bucket.resetAt) check below),
	// so this only reclaims memory and never changes a limiting decision.
	if maxRequests > 0 && window > 0 {
		go func() {
			ticker := time.NewTicker(window)
			defer ticker.Stop()
			for range ticker.C {
				cutoff := time.Now()
				mu.Lock()
				for ip, bucket := range buckets {
					if cutoff.After(bucket.resetAt) {
						delete(buckets, ip)
					}
				}
				mu.Unlock()
			}
		}()
	}

	return func(c *gin.Context) {
		if maxRequests <= 0 {
			c.Next()
			return
		}

		now := time.Now()
		ip := c.ClientIP()

		mu.Lock()
		bucket, ok := buckets[ip]
		if !ok || now.After(bucket.resetAt) {
			bucket = &rateBucket{count: 0, resetAt: now.Add(window)}
			buckets[ip] = bucket
		}
		bucket.count++
		allowed := bucket.count <= maxRequests
		resetAt := bucket.resetAt
		mu.Unlock()

		c.Header("X-RateLimit-Limit", itoa(maxRequests))
		c.Header("X-RateLimit-Reset", itoa(int(resetAt.Unix())))

		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": gin.H{"code": "RATE_LIMITED", "message": "rate limit exceeded"}})
			return
		}
		c.Next()
	}
}

/*
@function itoa

@desc
Formats an integer for HTTP header values.

@responsibilities
- Convert integer values to strings without leaking strconv imports to call sites

@params
- value: integer value

@returns
- string
*/
func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	digits := make([]byte, 0, 20)
	negative := value < 0
	if negative {
		value = -value
	}
	for value > 0 {
		digits = append(digits, byte('0'+value%10))
		value /= 10
	}
	if negative {
		digits = append(digits, '-')
	}
	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}
	return string(digits)
}
