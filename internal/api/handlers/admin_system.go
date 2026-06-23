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
@file admin_system.go

@desc
HTTP handlers for system configuration admin endpoints.

@handler_set
- GetSystemConfig: GET /api/v1/admin/system/config
- PrepareUpdateSystemConfig: POST /api/v1/admin/system/config/prepare

@dependencies
- AdminSystemServiceInterface
*/

/*
@interface AdminSystemServiceInterface

@desc
Service interface for system configuration management.
*/
type AdminSystemServiceInterface interface {
	GetSystemConfig(ctx interface{}) (*response.SystemConfigListResponse, error)
	PrepareUpdateSystemConfig(ctx interface{}, key, value, adminAddress, reason string) (*response.TransactionPreparation, error)
}

/*
@struct AdminSystemHandler

@desc
HTTP handler for system configuration endpoints.

@fields
- service: AdminSystemServiceInterface implementation
*/
type AdminSystemHandler struct {
	service AdminSystemServiceInterface
}

/*
@function NewAdminSystemHandler

@desc
Creates a new AdminSystemHandler.

@params
- svc: AdminSystemServiceInterface implementation

@returns
- *AdminSystemHandler
*/
func NewAdminSystemHandler(svc AdminSystemServiceInterface) *AdminSystemHandler {
	return &AdminSystemHandler{service: svc}
}

/*
@handler GetSystemConfig

@desc
Returns all system configuration key-value entries.
Sensitive values are masked in the response.

@route GET /api/v1/admin/system/config

@response
HTTP 200 — { "success": true, "data": { SystemConfigListResponse } }

@error_cases
- 500: Service error
*/
func (h *AdminSystemHandler) GetSystemConfig(c *gin.Context) {
	result, err := h.service.GetSystemConfig(c)
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
@handler PrepareUpdateSystemConfig

@desc
Updates a system configuration value and returns encoded governance calldata.

@route POST /api/v1/admin/system/config/prepare

@request
Body: UpdateSystemConfigRequest { "key": "...", "value": "...", "reason": "..." }

@response
HTTP 200 — { "success": true, "data": { TransactionPreparation } }

@error_cases
- 400: Missing key, value, or invalid request
- 500: Service error
*/
func (h *AdminSystemHandler) PrepareUpdateSystemConfig(c *gin.Context) {
	var req request.UpdateSystemConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "INVALID_REQUEST", "message": fmt.Sprintf("invalid request: %v", err)},
		})
		return
	}

	adminAddr, _ := c.Get("wallet_address")
	adminAddress, _ := adminAddr.(string)

	result, err := h.service.PrepareUpdateSystemConfig(c, req.Key, req.Value, adminAddress, req.Reason)
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
