package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/Revvfi/revvfi-backend/internal/api/dto/request"
	"github.com/Revvfi/revvfi-backend/internal/api/dto/response"
	appErr "github.com/Revvfi/revvfi-backend/internal/pkg/errors"
)

/*
@file admin_liquidator.go

@desc
HTTP handlers for liquidator configuration and active auction management.

@handler_set
- GetLiquidatorConfig: GET /api/v1/admin/liquidator/config
- PrepareSetLiquidatorParams: POST /api/v1/admin/liquidator/config/prepare
- GetActiveAuctions: GET /api/v1/admin/liquidator/auctions
- PrepareStopAuction: POST /api/v1/admin/liquidator/auctions/:auctionID/stop/prepare

@dependencies
- AdminLiquidatorServiceInterface
*/

/*
@interface AdminLiquidatorServiceInterface

@desc
Service interface for liquidator management.
*/
type AdminLiquidatorServiceInterface interface {
	GetLiquidatorConfig(ctx interface{}) (*response.LiquidatorConfig, error)
	PrepareSetLiquidatorParams(ctx interface{}, auctionDurationSecs, priceDropRateBps, minBidIncrementBps, liquidationIncentiveBps, maxSlippageBps int, updatedBy, reason string) (*response.TransactionPreparation, error)
	GetActiveAuctions(ctx interface{}) (*response.ActiveAuctionsResponse, error)
	PrepareStopAuction(ctx interface{}, auctionID int64, adminAddress, reason string) (*response.TransactionPreparation, error)
}

/*
@struct AdminLiquidatorHandler

@desc
HTTP handler for liquidator configuration endpoints.

@fields
- service: AdminLiquidatorServiceInterface implementation
*/
type AdminLiquidatorHandler struct {
	service AdminLiquidatorServiceInterface
}

/*
@function NewAdminLiquidatorHandler

@desc
Creates a new AdminLiquidatorHandler.

@params
- svc: AdminLiquidatorServiceInterface implementation

@returns
- *AdminLiquidatorHandler
*/
func NewAdminLiquidatorHandler(svc AdminLiquidatorServiceInterface) *AdminLiquidatorHandler {
	return &AdminLiquidatorHandler{service: svc}
}

/*
@handler GetLiquidatorConfig

@desc
Returns the current liquidator Dutch auction parameters.

@route GET /api/v1/admin/liquidator/config

@response
HTTP 200 — { "success": true, "data": { LiquidatorConfig } }

@error_cases
- 500: Service error
*/
func (h *AdminLiquidatorHandler) GetLiquidatorConfig(c *gin.Context) {
	result, err := h.service.GetLiquidatorConfig(c)
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
@handler PrepareSetLiquidatorParams

@desc
Prepares a transaction to update the liquidator contract parameters.

@route POST /api/v1/admin/liquidator/config/prepare

@request
Body: SetLiquidatorParamsRequest

@response
HTTP 200 — { "success": true, "data": { TransactionPreparation } }

@error_cases
- 400: Invalid request body
- 500: Service error
*/
func (h *AdminLiquidatorHandler) PrepareSetLiquidatorParams(c *gin.Context) {
	var req request.SetLiquidatorParamsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "INVALID_REQUEST", "message": fmt.Sprintf("invalid request: %v", err)},
		})
		return
	}

	adminAddr, _ := c.Get("wallet_address")
	adminAddress, _ := adminAddr.(string)

	result, err := h.service.PrepareSetLiquidatorParams(
		c,
		req.AuctionDurationSeconds, req.PriceDropRateBps,
		req.MinBidIncrementBps, req.LiquidationIncentiveBps, req.MaxSlippageBps,
		adminAddress, req.Reason,
	)
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
@handler GetActiveAuctions

@desc
Returns all currently active liquidation auctions.

@route GET /api/v1/admin/liquidator/auctions

@response
HTTP 200 — { "success": true, "data": { ActiveAuctionsResponse } }

@error_cases
- 500: Service error
*/
func (h *AdminLiquidatorHandler) GetActiveAuctions(c *gin.Context) {
	result, err := h.service.GetActiveAuctions(c)
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
@handler PrepareStopAuction

@desc
Prepares a transaction to cancel an active auction.

@route POST /api/v1/admin/liquidator/auctions/:auctionID/stop/prepare

@request
Path: auctionID (numeric auction identifier)
Body: StopAuctionRequest

@response
HTTP 200 — { "success": true, "data": { TransactionPreparation } }

@error_cases
- 400: Invalid auctionID or request body
- 500: Service error
*/
func (h *AdminLiquidatorHandler) PrepareStopAuction(c *gin.Context) {
	auctionIDStr := c.Param("auctionID")
	auctionID, err := strconv.ParseInt(auctionIDStr, 10, 64)
	if err != nil || auctionID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "INVALID_AUCTION_ID", "message": "invalid auction ID"},
		})
		return
	}

	var req request.StopAuctionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "INVALID_REQUEST", "message": fmt.Sprintf("invalid request: %v", err)},
		})
		return
	}

	adminAddr, _ := c.Get("wallet_address")
	adminAddress, _ := adminAddr.(string)

	result, svcErr := h.service.PrepareStopAuction(c, auctionID, adminAddress, req.Reason)
	if svcErr != nil {
		statusCode := appErr.StatusCode(svcErr)
		c.JSON(statusCode, gin.H{
			"success": false,
			"error":   gin.H{"code": appErr.ErrorCode(svcErr), "message": svcErr.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}
