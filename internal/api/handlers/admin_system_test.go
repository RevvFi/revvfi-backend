package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Revvfi/revvfi-backend/internal/api/dto/response"
)

/*
@file admin_system_test.go

@desc
Unit tests for system configuration management handlers.
Covers both endpoints: GetSystemConfig and PrepareUpdateSystemConfig.
*/

/*
@test mockAdminSystemService

@desc
Mock implementation of AdminSystemServiceInterface for handler testing.
*/
type mockAdminSystemService struct {
	getSystemConfigFn           func(ctx interface{}) (*response.SystemConfigListResponse, error)
	prepareUpdateSystemConfigFn func(ctx interface{}, key, value, adminAddress, reason string) (*response.TransactionPreparation, error)
}

/*
@function GetSystemConfig (mock)
@desc Mock returning controlled system configuration list.
*/
func (m *mockAdminSystemService) GetSystemConfig(ctx interface{}) (*response.SystemConfigListResponse, error) {
	if m.getSystemConfigFn != nil {
		return m.getSystemConfigFn(ctx)
	}
	return nil, fmt.Errorf("not implemented")
}

/*
@function PrepareUpdateSystemConfig (mock)
@desc Mock returning controlled system config update transaction.
*/
func (m *mockAdminSystemService) PrepareUpdateSystemConfig(ctx interface{}, key, value, adminAddress, reason string) (*response.TransactionPreparation, error) {
	if m.prepareUpdateSystemConfigFn != nil {
		return m.prepareUpdateSystemConfigFn(ctx, key, value, adminAddress, reason)
	}
	return nil, fmt.Errorf("not implemented")
}

func TestAdminSystemHandler_GetSystemConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		svc := &mockAdminSystemService{
			getSystemConfigFn: func(_ interface{}) (*response.SystemConfigListResponse, error) {
				return &response.SystemConfigListResponse{
					Configs: []response.SystemConfigEntry{
						{Key: "max_markets", Value: "50", UpdatedBy: "0xAdmin", UpdatedAt: 1700000000},
					},
					Total: 1,
				}, nil
			},
		}
		h := NewAdminSystemHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/admin/system/config", nil)

		h.GetSystemConfig(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var body map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.True(t, body["success"].(bool))
	})

	t.Run("service error", func(t *testing.T) {
		svc := &mockAdminSystemService{
			getSystemConfigFn: func(_ interface{}) (*response.SystemConfigListResponse, error) {
				return nil, fmt.Errorf("db error")
			},
		}
		h := NewAdminSystemHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/admin/system/config", nil)

		h.GetSystemConfig(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestAdminSystemHandler_PrepareUpdateSystemConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		svc := &mockAdminSystemService{
			prepareUpdateSystemConfigFn: func(_ interface{}, key, value, _, _ string) (*response.TransactionPreparation, error) {
				return &response.TransactionPreparation{
					Target:      "0xArch",
					TxData:      "0xcfg",
					Value:       "0",
					Description: fmt.Sprintf("Set %s = %s", key, value),
				}, nil
			},
		}
		h := NewAdminSystemHandler(svc)

		body, _ := json.Marshal(map[string]interface{}{"key": "max_markets", "value": "100", "reason": "scaling up"})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/admin/system/config/prepare", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		h.PrepareUpdateSystemConfig(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.True(t, resp["success"].(bool))
	})

	t.Run("invalid body", func(t *testing.T) {
		h := NewAdminSystemHandler(&mockAdminSystemService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{bad json}`)))
		c.Request.Header.Set("Content-Type", "application/json")
		h.PrepareUpdateSystemConfig(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing required key field", func(t *testing.T) {
		h := NewAdminSystemHandler(&mockAdminSystemService{})
		// value is present but key is missing — should fail binding validation
		body, _ := json.Marshal(map[string]interface{}{"value": "100", "reason": "test"})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		h.PrepareUpdateSystemConfig(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		svc := &mockAdminSystemService{
			prepareUpdateSystemConfigFn: func(_ interface{}, _, _, _, _ string) (*response.TransactionPreparation, error) {
				return nil, fmt.Errorf("contract error")
			},
		}
		h := NewAdminSystemHandler(svc)

		body, _ := json.Marshal(map[string]interface{}{"key": "max_markets", "value": "100", "reason": "test"})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		h.PrepareUpdateSystemConfig(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
