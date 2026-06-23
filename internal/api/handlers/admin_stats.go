package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Revvfi/revvfi-backend/internal/api/dto/response"
	appErr "github.com/Revvfi/revvfi-backend/internal/pkg/errors"
)

/*
@file admin_stats.go

@desc
HTTP handlers for admin dashboard statistics endpoints.

@handler_set
- GetOverview: GET /api/v1/admin/stats/overview
- GetBorrowerStats: GET /api/v1/admin/stats/borrowers
- GetMarketStats: GET /api/v1/admin/stats/markets
- GetRevenueStats: GET /api/v1/admin/stats/revenue
- GetLiquidationStats: GET /api/v1/admin/stats/liquidations
- GetPositionStats: GET /api/v1/admin/stats/positions

@dependencies
- AdminStatsServiceInterface
*/

/*
@interface AdminStatsServiceInterface

@desc
Service interface for admin dashboard statistics.
*/
type AdminStatsServiceInterface interface {
	GetOverview(ctx interface{}) (*response.OverviewStats, error)
	GetBorrowerStats(ctx interface{}) (*response.BorrowerGrowthStats, error)
	GetMarketStats(ctx interface{}) (*response.MarketCreationStats, error)
	GetRevenueStats(ctx interface{}) (*response.RevenueStats, error)
	GetLiquidationStats(ctx interface{}) (*response.LiquidationStats, error)
	GetPositionStats(ctx interface{}) (*response.PositionDistributionStats, error)
}

/*
@struct AdminStatsHandler

@desc
HTTP handler for admin dashboard statistics endpoints.

@fields
- service: AdminStatsServiceInterface implementation
*/
type AdminStatsHandler struct {
	service AdminStatsServiceInterface
}

/*
@function NewAdminStatsHandler

@desc
Creates a new AdminStatsHandler.

@params
- svc: AdminStatsServiceInterface implementation

@returns
- *AdminStatsHandler
*/
func NewAdminStatsHandler(svc AdminStatsServiceInterface) *AdminStatsHandler {
	return &AdminStatsHandler{service: svc}
}

/*
@handler GetOverview

@desc
Returns protocol-wide overview statistics for the admin dashboard.

@route GET /api/v1/admin/stats/overview

@response
HTTP 200 — { "success": true, "data": { OverviewStats } }

@error_cases
- 500: Service error
*/
func (h *AdminStatsHandler) GetOverview(c *gin.Context) {
	result, err := h.service.GetOverview(c)
	if err != nil {
		statusCode := appErr.StatusCode(err)
		c.JSON(statusCode, gin.H{
			"success": false,
			"error":   gin.H{"code": appErr.ErrorCode(err), "message": err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

/*
@handler GetBorrowerStats

@desc
Returns borrower growth and reputation distribution metrics.

@route GET /api/v1/admin/stats/borrowers

@response
HTTP 200 — { "success": true, "data": { BorrowerGrowthStats } }

@error_cases
- 500: Service error
*/
func (h *AdminStatsHandler) GetBorrowerStats(c *gin.Context) {
	result, err := h.service.GetBorrowerStats(c)
	if err != nil {
		statusCode := appErr.StatusCode(err)
		c.JSON(statusCode, gin.H{
			"success": false,
			"error":   gin.H{"code": appErr.ErrorCode(err), "message": err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

/*
@handler GetMarketStats

@desc
Returns market creation and utilization metrics.

@route GET /api/v1/admin/stats/markets

@response
HTTP 200 — { "success": true, "data": { MarketCreationStats } }

@error_cases
- 500: Service error
*/
func (h *AdminStatsHandler) GetMarketStats(c *gin.Context) {
	result, err := h.service.GetMarketStats(c)
	if err != nil {
		statusCode := appErr.StatusCode(err)
		c.JSON(statusCode, gin.H{
			"success": false,
			"error":   gin.H{"code": appErr.ErrorCode(err), "message": err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

/*
@handler GetRevenueStats

@desc
Returns protocol revenue analytics from repayment data.

@route GET /api/v1/admin/stats/revenue

@response
HTTP 200 — { "success": true, "data": { RevenueStats } }

@error_cases
- 500: Service error
*/
func (h *AdminStatsHandler) GetRevenueStats(c *gin.Context) {
	result, err := h.service.GetRevenueStats(c)
	if err != nil {
		statusCode := appErr.StatusCode(err)
		c.JSON(statusCode, gin.H{
			"success": false,
			"error":   gin.H{"code": appErr.ErrorCode(err), "message": err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

/*
@handler GetLiquidationStats

@desc
Returns liquidation auction statistics and recovery metrics.

@route GET /api/v1/admin/stats/liquidations

@response
HTTP 200 — { "success": true, "data": { LiquidationStats } }

@error_cases
- 500: Service error
*/
func (h *AdminStatsHandler) GetLiquidationStats(c *gin.Context) {
	result, err := h.service.GetLiquidationStats(c)
	if err != nil {
		statusCode := appErr.StatusCode(err)
		c.JSON(statusCode, gin.H{
			"success": false,
			"error":   gin.H{"code": appErr.ErrorCode(err), "message": err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

/*
@handler GetPositionStats

@desc
Returns position distribution across seniority tiers and status.

@route GET /api/v1/admin/stats/positions

@response
HTTP 200 — { "success": true, "data": { PositionDistributionStats } }

@error_cases
- 500: Service error
*/
func (h *AdminStatsHandler) GetPositionStats(c *gin.Context) {
	result, err := h.service.GetPositionStats(c)
	if err != nil {
		statusCode := appErr.StatusCode(err)
		c.JSON(statusCode, gin.H{
			"success": false,
			"error":   gin.H{"code": appErr.ErrorCode(err), "message": err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}
