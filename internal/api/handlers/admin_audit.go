package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/Revvfi/revvfi-backend/internal/api/dto/response"
	appErr "github.com/Revvfi/revvfi-backend/internal/pkg/errors"
)

/*
@file admin_audit.go

@desc
HTTP handlers for admin audit log and compliance monitoring endpoints.

@handler_set
- GetAuditLogs: GET /api/v1/admin/audit/logs
- GetAuditStats: GET /api/v1/admin/audit/stats
- ExportAuditLogs: GET /api/v1/admin/audit/export
- GetAdminActivity: GET /api/v1/admin/audit/activity/:adminAddress
- GetActionHistory: GET /api/v1/admin/audit/actions/:action

@dependencies
- AdminAuditServiceInterface
*/

/*
@interface AdminAuditServiceInterface

@desc
Service interface for audit log querying and statistics.
*/
type AdminAuditServiceInterface interface {
	GetAuditLogs(ctx interface{}, page, limit int32, adminAddr, action, targetType string) (*response.AuditLogListResponse, error)
	GetAuditStats(ctx interface{}) (*response.AuditStats, error)
	ExportAuditLogs(ctx interface{}, fromUnix, toUnix int64, format string) (*response.AuditExportResponse, error)
	GetAdminActivity(ctx interface{}, adminAddr string) (*response.AuditLogListResponse, error)
	GetActionHistory(ctx interface{}, action string) (*response.AuditLogListResponse, error)
}

/*
@struct AdminAuditHandler

@desc
HTTP handler for admin audit monitoring endpoints.

@fields
- service: AdminAuditServiceInterface implementation
*/
type AdminAuditHandler struct {
	service AdminAuditServiceInterface
}

/*
@function NewAdminAuditHandler

@desc
Creates a new AdminAuditHandler.

@params
- svc: AdminAuditServiceInterface implementation

@returns
- *AdminAuditHandler
*/
func NewAdminAuditHandler(svc AdminAuditServiceInterface) *AdminAuditHandler {
	return &AdminAuditHandler{service: svc}
}

/*
@handler GetAuditLogs

@desc
Returns paginated, filtered admin audit log entries.

@route GET /api/v1/admin/audit/logs

@request
Query: page, limit, admin_address, action, target_type

@response
HTTP 200 — { "success": true, "data": { AuditLogListResponse } }

@error_cases
- 500: Service error
*/
func (h *AdminAuditHandler) GetAuditLogs(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")
	adminAddr := c.Query("admin_address")
	action := c.Query("action")
	targetType := c.Query("target_type")

	page, err := strconv.ParseInt(pageStr, 10, 32)
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	result, svcErr := h.service.GetAuditLogs(c, int32(page), int32(limit), adminAddr, action, targetType)
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

/*
@handler GetAuditStats

@desc
Returns aggregate statistics over the entire audit log.

@route GET /api/v1/admin/audit/stats

@response
HTTP 200 — { "success": true, "data": { AuditStats } }

@error_cases
- 500: Service error
*/
func (h *AdminAuditHandler) GetAuditStats(c *gin.Context) {
	result, err := h.service.GetAuditStats(c)
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
@handler ExportAuditLogs

@desc
Exports audit logs for a time range in the requested format.

@route GET /api/v1/admin/audit/export

@request
Query: from (Unix timestamp), to (Unix timestamp), format (json|csv)

@response
HTTP 200 — { "success": true, "data": { AuditExportResponse } }

@error_cases
- 500: Service error
*/
func (h *AdminAuditHandler) ExportAuditLogs(c *gin.Context) {
	fromStr := c.DefaultQuery("from", "0")
	toStr := c.DefaultQuery("to", "0")
	format := c.DefaultQuery("format", "json")

	from, _ := strconv.ParseInt(fromStr, 10, 64)
	to, _ := strconv.ParseInt(toStr, 10, 64)

	result, err := h.service.ExportAuditLogs(c, from, to, format)
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
@handler GetAdminActivity

@desc
Returns recent audit log entries for a specific admin address.

@route GET /api/v1/admin/audit/activity/:adminAddress

@request
Path: adminAddress (wallet address)

@response
HTTP 200 — { "success": true, "data": { AuditLogListResponse } }

@error_cases
- 400: Missing address
- 500: Service error
*/
func (h *AdminAuditHandler) GetAdminActivity(c *gin.Context) {
	adminAddress := c.Param("adminAddress")
	if adminAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "INVALID_ADDRESS", "message": "admin address is required"},
		})
		return
	}

	result, err := h.service.GetAdminActivity(c, adminAddress)
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
@handler GetActionHistory

@desc
Returns all audit log entries for a specific action type.

@route GET /api/v1/admin/audit/actions/:action

@request
Path: action (action name, e.g., update_min_cr)

@response
HTTP 200 — { "success": true, "data": { AuditLogListResponse } }

@error_cases
- 400: Missing action
- 500: Service error
*/
func (h *AdminAuditHandler) GetActionHistory(c *gin.Context) {
	action := c.Param("action")
	if action == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "INVALID_ACTION", "message": "action is required"},
		})
		return
	}

	result, err := h.service.GetActionHistory(c, action)
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
