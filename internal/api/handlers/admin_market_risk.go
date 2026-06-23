package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Revvfi/revvfi-backend/internal/api/dto/request"
	"github.com/Revvfi/revvfi-backend/internal/api/dto/response"
	appErr "github.com/Revvfi/revvfi-backend/internal/pkg/errors"
)

/*
@file admin_market_risk.go

@desc
HTTP handlers for market risk parameter admin endpoints.
Allows admins to read and update risk parameters for individual markets.

@handler_set
- GetMarketRiskParams: GET /api/v1/admin/markets/:address/risk
- PrepareSetMinCR: POST /api/v1/admin/markets/:address/risk/min-cr/prepare
- PrepareSetLiqThreshold: POST /api/v1/admin/markets/:address/risk/liquidation-threshold/prepare
- PrepareSetOracle: POST /api/v1/admin/markets/:address/risk/oracle/prepare

@dependencies
- AdminMarketRiskServiceInterface: service for market risk operations
*/

/*
@interface AdminMarketRiskServiceInterface

@desc
Service interface for market risk parameter management.

@methods
- GetMarketRiskParams: read current risk parameters for a market
- PrepareSetMinCR: prepare tx to update min collateral ratio
- PrepareSetLiqThreshold: prepare tx to update liquidation threshold
- PrepareSetOracle: prepare tx to update collateral oracle
*/
type AdminMarketRiskServiceInterface interface {
	GetMarketRiskParams(ctx interface{}, address string) (*response.MarketRiskParams, error)
	PrepareSetMinCR(ctx interface{}, address string, minCRBps int32, reason string) (*response.TransactionPreparation, error)
	PrepareSetLiqThreshold(ctx interface{}, address string, thresholdBps int32, reason string) (*response.TransactionPreparation, error)
	PrepareSetOracle(ctx interface{}, address, oracleAddress, reason string) (*response.TransactionPreparation, error)
}

/*
@struct AdminMarketRiskHandler

@desc
HTTP handler for market risk parameter endpoints.

@fields
- service: AdminMarketRiskServiceInterface implementation
*/
type AdminMarketRiskHandler struct {
	service AdminMarketRiskServiceInterface
}

/*
@function NewAdminMarketRiskHandler

@desc
Creates a new AdminMarketRiskHandler.

@params
- svc: AdminMarketRiskServiceInterface implementation

@returns
- *AdminMarketRiskHandler
*/
func NewAdminMarketRiskHandler(svc AdminMarketRiskServiceInterface) *AdminMarketRiskHandler {
	return &AdminMarketRiskHandler{service: svc}
}

/*
@handler GetMarketRiskParams

@desc
Returns the current risk parameters for a specific market.

@route GET /api/v1/admin/markets/:address/risk

@request
Path: address (market contract address)

@response
HTTP 200 — { "success": true, "data": { MarketRiskParams } }

@error_cases
- 400: Missing or blank address
- 404: Market not found
- 500: Service error
*/
func (h *AdminMarketRiskHandler) GetMarketRiskParams(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "INVALID_ADDRESS", "message": "market address is required"},
		})
		return
	}

	result, err := h.service.GetMarketRiskParams(c, address)
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
@handler PrepareSetMinCR

@desc
Prepares a transaction to update the minimum collateral ratio for a market.

@route POST /api/v1/admin/markets/:address/risk/min-cr/prepare

@request
Path: address
Body: { "min_collateral_ratio_bps": 11000, "reason": "..." }

@response
HTTP 200 — { "success": true, "data": { TransactionPreparation } }

@error_cases
- 400: Invalid address or request body
- 404: Market not found
- 500: Service error
*/
func (h *AdminMarketRiskHandler) PrepareSetMinCR(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "INVALID_ADDRESS", "message": "market address is required"},
		})
		return
	}

	var req request.SetMinCRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "INVALID_REQUEST", "message": fmt.Sprintf("invalid request: %v", err)},
		})
		return
	}

	result, err := h.service.PrepareSetMinCR(c, address, int32(req.MinCollateralRatioBps), req.Reason)
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
@handler PrepareSetLiqThreshold

@desc
Prepares a transaction to update the liquidation threshold for a market.

@route POST /api/v1/admin/markets/:address/risk/liquidation-threshold/prepare

@request
Path: address
Body: { "liquidation_threshold_bps": 9500, "reason": "..." }

@response
HTTP 200 — { "success": true, "data": { TransactionPreparation } }

@error_cases
- 400: Invalid address or request body
- 404: Market not found
- 500: Service error
*/
func (h *AdminMarketRiskHandler) PrepareSetLiqThreshold(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "INVALID_ADDRESS", "message": "market address is required"},
		})
		return
	}

	var req request.SetLiqThresholdRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "INVALID_REQUEST", "message": fmt.Sprintf("invalid request: %v", err)},
		})
		return
	}

	result, err := h.service.PrepareSetLiqThreshold(c, address, int32(req.LiquidationThresholdBps), req.Reason)
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
@handler PrepareSetOracle

@desc
Prepares a transaction to update the collateral oracle address for a market.

@route POST /api/v1/admin/markets/:address/risk/oracle/prepare

@request
Path: address
Body: { "oracle_address": "0x...", "reason": "..." }

@response
HTTP 200 — { "success": true, "data": { TransactionPreparation } }

@error_cases
- 400: Invalid address or request body
- 404: Market not found
- 500: Service error
*/
func (h *AdminMarketRiskHandler) PrepareSetOracle(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "INVALID_ADDRESS", "message": "market address is required"},
		})
		return
	}

	var req request.SetOracleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "INVALID_REQUEST", "message": fmt.Sprintf("invalid request: %v", err)},
		})
		return
	}

	result, err := h.service.PrepareSetOracle(c, address, req.OracleAddress, req.Reason)
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
