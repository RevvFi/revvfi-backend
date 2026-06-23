package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"encoding/json"

	"github.com/Revvfi/revvfi-backend/internal/api/dto/response"
)

/*
@file admin_audit_test.go

@desc
Unit tests for audit log and compliance monitoring handlers.
Tests all five endpoints covering success paths, pagination, and service errors.
*/

/*
@test mockAdminAuditService

@desc
Mock implementation of AdminAuditServiceInterface for handler testing.
*/
type mockAdminAuditService struct {
	getAuditLogsFn      func(ctx interface{}, page, limit int32, adminAddr, action, targetType string) (*response.AuditLogListResponse, error)
	getAuditStatsFn     func(ctx interface{}) (*response.AuditStats, error)
	exportAuditLogsFn   func(ctx interface{}, fromUnix, toUnix int64, format string) (*response.AuditExportResponse, error)
	getAdminActivityFn  func(ctx interface{}, adminAddr string) (*response.AuditLogListResponse, error)
	getActionHistoryFn  func(ctx interface{}, action string) (*response.AuditLogListResponse, error)
}

/*
@function GetAuditLogs (mock)
@desc Mock returning a controlled paginated audit log list.
*/
func (m *mockAdminAuditService) GetAuditLogs(ctx interface{}, page, limit int32, adminAddr, action, targetType string) (*response.AuditLogListResponse, error) {
	if m.getAuditLogsFn != nil {
		return m.getAuditLogsFn(ctx, page, limit, adminAddr, action, targetType)
	}
	return nil, fmt.Errorf("not implemented")
}

/*
@function GetAuditStats (mock)
@desc Mock returning controlled audit statistics.
*/
func (m *mockAdminAuditService) GetAuditStats(ctx interface{}) (*response.AuditStats, error) {
	if m.getAuditStatsFn != nil {
		return m.getAuditStatsFn(ctx)
	}
	return nil, fmt.Errorf("not implemented")
}

/*
@function ExportAuditLogs (mock)
@desc Mock returning controlled audit export data.
*/
func (m *mockAdminAuditService) ExportAuditLogs(ctx interface{}, fromUnix, toUnix int64, format string) (*response.AuditExportResponse, error) {
	if m.exportAuditLogsFn != nil {
		return m.exportAuditLogsFn(ctx, fromUnix, toUnix, format)
	}
	return nil, fmt.Errorf("not implemented")
}

/*
@function GetAdminActivity (mock)
@desc Mock returning controlled activity log for an admin address.
*/
func (m *mockAdminAuditService) GetAdminActivity(ctx interface{}, adminAddr string) (*response.AuditLogListResponse, error) {
	if m.getAdminActivityFn != nil {
		return m.getAdminActivityFn(ctx, adminAddr)
	}
	return nil, fmt.Errorf("not implemented")
}

/*
@function GetActionHistory (mock)
@desc Mock returning controlled action history for a named action.
*/
func (m *mockAdminAuditService) GetActionHistory(ctx interface{}, action string) (*response.AuditLogListResponse, error) {
	if m.getActionHistoryFn != nil {
		return m.getActionHistoryFn(ctx, action)
	}
	return nil, fmt.Errorf("not implemented")
}

func TestAdminAuditHandler_GetAuditLogs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		svc := &mockAdminAuditService{
			getAuditLogsFn: func(_ interface{}, _, _ int32, _, _, _ string) (*response.AuditLogListResponse, error) {
				return &response.AuditLogListResponse{
					Logs:       []response.AuditLogEntry{},
					Pagination: &response.PaginationInfo{Page: 1, Limit: 20},
					Total:      0,
				}, nil
			},
		}
		h := NewAdminAuditHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/admin/audit/logs?page=1&limit=20", nil)

		h.GetAuditLogs(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var body map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.True(t, body["success"].(bool))
	})

	t.Run("service error", func(t *testing.T) {
		svc := &mockAdminAuditService{
			getAuditLogsFn: func(_ interface{}, _, _ int32, _, _, _ string) (*response.AuditLogListResponse, error) {
				return nil, fmt.Errorf("db error")
			},
		}
		h := NewAdminAuditHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/admin/audit/logs", nil)

		h.GetAuditLogs(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestAdminAuditHandler_GetAuditStats(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		svc := &mockAdminAuditService{
			getAuditStatsFn: func(_ interface{}) (*response.AuditStats, error) {
				return &response.AuditStats{
					TotalActions:    100,
					RecentActivity:  10,
					SuccessRate:     0.95,
					ActionBreakdown: map[string]int64{"SET_CR": 40, "SET_ORACLE": 60},
					AdminBreakdown:  map[string]int64{"0xAdmin": 100},
				}, nil
			},
		}
		h := NewAdminAuditHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/admin/audit/stats", nil)

		h.GetAuditStats(c)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestAdminAuditHandler_ExportAuditLogs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success csv", func(t *testing.T) {
		svc := &mockAdminAuditService{
			exportAuditLogsFn: func(_ interface{}, _, _ int64, format string) (*response.AuditExportResponse, error) {
				return &response.AuditExportResponse{Format: format, TotalRecords: 5, Data: []response.AuditLogEntry{}}, nil
			},
		}
		h := NewAdminAuditHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/admin/audit/export?from=0&to=9999999999&format=csv", nil)

		h.ExportAuditLogs(c)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestAdminAuditHandler_GetAdminActivity(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		svc := &mockAdminAuditService{
			getAdminActivityFn: func(_ interface{}, addr string) (*response.AuditLogListResponse, error) {
				return &response.AuditLogListResponse{Logs: []response.AuditLogEntry{}, Pagination: nil}, nil
			},
		}
		h := NewAdminAuditHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/admin/audit/activity/0xAdmin", nil)
		c.Params = gin.Params{{Key: "adminAddress", Value: "0xAdmin"}}

		h.GetAdminActivity(c)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("missing admin address", func(t *testing.T) {
		h := NewAdminAuditHandler(&mockAdminAuditService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)
		c.Params = gin.Params{{Key: "adminAddress", Value: ""}}
		h.GetAdminActivity(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAdminAuditHandler_GetActionHistory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		svc := &mockAdminAuditService{
			getActionHistoryFn: func(_ interface{}, action string) (*response.AuditLogListResponse, error) {
				return &response.AuditLogListResponse{Logs: []response.AuditLogEntry{}, Pagination: nil}, nil
			},
		}
		h := NewAdminAuditHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/admin/audit/actions/SET_CR", nil)
		c.Params = gin.Params{{Key: "action", Value: "SET_CR"}}

		h.GetActionHistory(c)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("missing action", func(t *testing.T) {
		h := NewAdminAuditHandler(&mockAdminAuditService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)
		c.Params = gin.Params{{Key: "action", Value: ""}}
		h.GetActionHistory(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
