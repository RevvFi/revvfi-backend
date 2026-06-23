package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"

	"github.com/Revvfi/revvfi-backend/internal/api/dto/response"
)

/*
@interface HealthPinger

@desc
Minimal database health dependency.

@responsibilities
- Verify database connectivity through PingContext
*/
type HealthPinger interface {
	PingContext(ctx context.Context) error
}

/*
@struct HealthHandler

@desc
HTTP handler for service health endpoints.

@responsibilities
- Report liveness and readiness
- Check database connectivity when configured
- Check RPC connectivity when configured
- Include uptime metadata
*/
type HealthHandler struct {
	startedAt time.Time
	db        HealthPinger
	rpcURL    string
}

/*
@function NewHealthHandler

@desc
Creates a health HTTP handler.

@params
- db: optional database pinger
- rpcURL: optional RPC endpoint URL

@returns
- *HealthHandler
*/
func NewHealthHandler(db HealthPinger, rpcURL string) *HealthHandler {
	return &HealthHandler{
		startedAt: time.Now(),
		db:        db,
		rpcURL:    rpcURL,
	}
}

/*
@method Health

@desc
Handles liveness health checks.

@params
- c: Gin context
*/
func (h *HealthHandler) Health(c *gin.Context) {
	status := "healthy"
	dbStatus := "not_configured"
	rpcStatus := "not_configured"

	// Check database
	if h.db != nil {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := h.db.PingContext(ctx); err != nil {
			status = "degraded"
			dbStatus = "unhealthy"
		} else {
			dbStatus = "healthy"
		}
	}

	// Check RPC
	if h.rpcURL != "" {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		client, err := ethclient.DialContext(ctx, h.rpcURL)
		if err != nil {
			status = "degraded"
			rpcStatus = "unhealthy"
		} else {
			defer client.Close()
			// Try to get block number to verify connection
			if _, err := client.BlockNumber(ctx); err != nil {
				status = "degraded"
				rpcStatus = "unhealthy"
			} else {
				rpcStatus = "healthy"
			}
		}
	}

	ok(c, http.StatusOK, response.HealthResponse{
		Status:    status,
		Database:  dbStatus,
		RPC:       rpcStatus,
		Uptime:    int64(time.Since(h.startedAt).Seconds()),
		Timestamp: time.Now().Unix(),
	})
}

/*
@method Ready

@desc
Handles readiness checks for deployment probes.

@params
- c: Gin context
*/
func (h *HealthHandler) Ready(c *gin.Context) {
	if h.db == nil {
		ok(c, http.StatusOK, response.ReadinessResponse{Ready: true})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	if err := h.db.PingContext(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, response.ReadinessResponse{Ready: false, Reason: err.Error()})
		return
	}
	ok(c, http.StatusOK, response.ReadinessResponse{Ready: true})
}
