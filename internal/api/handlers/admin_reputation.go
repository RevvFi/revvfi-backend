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
@file admin_reputation.go

@desc
HTTP handlers for borrower reputation management admin endpoints.

@handler_set
- GetBorrowerReputation: GET /api/v1/admin/reputation/:address
- PrepareSetReputation: POST /api/v1/admin/reputation/:address/prepare
- GetDefaultedBorrowers: GET /api/v1/admin/reputation/defaulted

@dependencies
- AdminReputationServiceInterface
*/

/*
@interface AdminReputationServiceInterface

@desc
Service interface for borrower reputation management.
*/
type AdminReputationServiceInterface interface {
	GetBorrowerReputation(ctx interface{}, address string) (*response.BorrowerReputation, error)
	PrepareSetReputation(ctx interface{}, address string, newScore int32, adminAddress, reason string) (*response.TransactionPreparation, error)
	GetDefaultedBorrowers(ctx interface{}, page, limit int32) (*response.DefaultedBorrowersResponse, error)
}

/*
@struct AdminReputationHandler

@desc
HTTP handler for borrower reputation management endpoints.

@fields
- service: AdminReputationServiceInterface implementation
*/
type AdminReputationHandler struct {
	service AdminReputationServiceInterface
}

/*
@function NewAdminReputationHandler

@desc
Creates a new AdminReputationHandler.

@params
- svc: AdminReputationServiceInterface implementation

@returns
- *AdminReputationHandler
*/
func NewAdminReputationHandler(svc AdminReputationServiceInterface) *AdminReputationHandler {
	return &AdminReputationHandler{service: svc}
}

/*
@handler GetBorrowerReputation

@desc
Returns the full reputation profile for a borrower address.

@route GET /api/v1/admin/reputation/:address

@request
Path: address (borrower wallet address)

@response
HTTP 200 — { "success": true, "data": { BorrowerReputation } }

@error_cases
- 400: Missing or blank address
- 404: Borrower not found
- 500: Service error
*/
func (h *AdminReputationHandler) GetBorrowerReputation(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "INVALID_ADDRESS", "message": "borrower address is required"},
		})
		return
	}

	result, err := h.service.GetBorrowerReputation(c, address)
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
@handler PrepareSetReputation

@desc
Prepares a transaction to override a borrower's reputation score.

@route POST /api/v1/admin/reputation/:address/prepare

@request
Path: address
Body: SetReputationRequest { "new_score": 750, "reason": "..." }

@response
HTTP 200 — { "success": true, "data": { TransactionPreparation } }

@error_cases
- 400: Invalid address, score out of range, or missing reason
- 404: Borrower not found
- 500: Service error
*/
func (h *AdminReputationHandler) PrepareSetReputation(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "INVALID_ADDRESS", "message": "borrower address is required"},
		})
		return
	}

	var req request.SetReputationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "INVALID_REQUEST", "message": fmt.Sprintf("invalid request: %v", err)},
		})
		return
	}

	adminAddr, _ := c.Get("wallet_address")
	adminAddress, _ := adminAddr.(string)

	result, err := h.service.PrepareSetReputation(c, address, int32(req.NewScore), adminAddress, req.Reason)
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
@handler GetDefaultedBorrowers

@desc
Returns a paginated list of borrowers with at least one defaulted loan.

@route GET /api/v1/admin/reputation/defaulted

@request
Query: page (default 1), limit (default 20, max 100)

@response
HTTP 200 — { "success": true, "data": { DefaultedBorrowersResponse } }

@error_cases
- 400: Invalid pagination parameters
- 500: Service error
*/
func (h *AdminReputationHandler) GetDefaultedBorrowers(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	page, err := strconv.ParseInt(pageStr, 10, 32)
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	result, svcErr := h.service.GetDefaultedBorrowers(c, int32(page), int32(limit))
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
