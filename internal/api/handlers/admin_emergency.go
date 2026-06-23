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
@file admin_emergency.go

@desc
HTTP handlers for emergency protocol control endpoints.

@handler_set
- PrepareEmergencyPause: POST /api/v1/admin/emergency/pause/prepare
- PrepareEmergencyUnpause: POST /api/v1/admin/emergency/unpause/prepare
- PrepareDrainFees: POST /api/v1/admin/emergency/drain-fees/prepare

@dependencies
- AdminEmergencyServiceInterface
*/

/*
@interface AdminEmergencyServiceInterface

@desc
Service interface for emergency protocol controls.
*/
type AdminEmergencyServiceInterface interface {
	PrepareEmergencyPause(ctx interface{}, adminAddress, reason string) (*response.TransactionPreparation, error)
	PrepareEmergencyUnpause(ctx interface{}, adminAddress, reason string) (*response.TransactionPreparation, error)
	PrepareDrainFees(ctx interface{}, recipient, adminAddress, reason string) (*response.TransactionPreparation, error)
}

/*
@struct AdminEmergencyHandler

@desc
HTTP handler for emergency control endpoints.

@fields
- service: AdminEmergencyServiceInterface implementation
*/
type AdminEmergencyHandler struct {
	service AdminEmergencyServiceInterface
}

/*
@function NewAdminEmergencyHandler

@desc
Creates a new AdminEmergencyHandler.

@params
- svc: AdminEmergencyServiceInterface implementation

@returns
- *AdminEmergencyHandler
*/
func NewAdminEmergencyHandler(svc AdminEmergencyServiceInterface) *AdminEmergencyHandler {
	return &AdminEmergencyHandler{service: svc}
}

/*
@handler PrepareEmergencyPause

@desc
Prepares a transaction to emergency-pause all protocol markets.

@route POST /api/v1/admin/emergency/pause/prepare

@request
Body: EmergencyActionRequest { "reason": "..." }

@response
HTTP 200 — { "success": true, "data": { TransactionPreparation } }

@error_cases
- 400: Missing or invalid reason
- 500: Service error
*/
func (h *AdminEmergencyHandler) PrepareEmergencyPause(c *gin.Context) {
	var req request.EmergencyActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "INVALID_REQUEST", "message": fmt.Sprintf("invalid request: %v", err)},
		})
		return
	}

	adminAddr, _ := c.Get("wallet_address")
	adminAddress, _ := adminAddr.(string)

	result, err := h.service.PrepareEmergencyPause(c, adminAddress, req.Reason)
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
@handler PrepareEmergencyUnpause

@desc
Prepares a transaction to unpause all protocol markets after an emergency.

@route POST /api/v1/admin/emergency/unpause/prepare

@request
Body: EmergencyActionRequest { "reason": "..." }

@response
HTTP 200 — { "success": true, "data": { TransactionPreparation } }

@error_cases
- 400: Missing or invalid reason
- 500: Service error
*/
func (h *AdminEmergencyHandler) PrepareEmergencyUnpause(c *gin.Context) {
	var req request.EmergencyActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "INVALID_REQUEST", "message": fmt.Sprintf("invalid request: %v", err)},
		})
		return
	}

	adminAddr, _ := c.Get("wallet_address")
	adminAddress, _ := adminAddr.(string)

	result, err := h.service.PrepareEmergencyUnpause(c, adminAddress, req.Reason)
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
@handler PrepareDrainFees

@desc
Prepares a transaction to drain accumulated protocol fees to a recipient.

@route POST /api/v1/admin/emergency/drain-fees/prepare

@request
Body: DrainFeesRequest { "recipient": "0x...", "reason": "..." }

@response
HTTP 200 — { "success": true, "data": { TransactionPreparation } }

@error_cases
- 400: Missing recipient or reason
- 500: Service error
*/
func (h *AdminEmergencyHandler) PrepareDrainFees(c *gin.Context) {
	var req request.DrainFeesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "INVALID_REQUEST", "message": fmt.Sprintf("invalid request: %v", err)},
		})
		return
	}

	adminAddr, _ := c.Get("wallet_address")
	adminAddress, _ := adminAddr.(string)

	result, err := h.service.PrepareDrainFees(c, req.Recipient, adminAddress, req.Reason)
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
