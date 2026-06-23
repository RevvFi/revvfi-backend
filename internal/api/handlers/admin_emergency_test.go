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
@file admin_emergency_test.go

@desc
Unit tests for emergency protocol control handlers.
Covers all three endpoints: pause, unpause, and drain-fees.
*/

/*
@test mockAdminEmergencyService

@desc
Mock implementation of AdminEmergencyServiceInterface for handler testing.
*/
type mockAdminEmergencyService struct {
	prepareEmergencyPauseFn   func(ctx interface{}, adminAddress, reason string) (*response.TransactionPreparation, error)
	prepareEmergencyUnpauseFn func(ctx interface{}, adminAddress, reason string) (*response.TransactionPreparation, error)
	prepareDrainFeesFn        func(ctx interface{}, recipient, adminAddress, reason string) (*response.TransactionPreparation, error)
}

/*
@function PrepareEmergencyPause (mock)
@desc Mock returning controlled emergency pause transaction data.
*/
func (m *mockAdminEmergencyService) PrepareEmergencyPause(ctx interface{}, adminAddress, reason string) (*response.TransactionPreparation, error) {
	if m.prepareEmergencyPauseFn != nil {
		return m.prepareEmergencyPauseFn(ctx, adminAddress, reason)
	}
	return nil, fmt.Errorf("not implemented")
}

/*
@function PrepareEmergencyUnpause (mock)
@desc Mock returning controlled emergency unpause transaction data.
*/
func (m *mockAdminEmergencyService) PrepareEmergencyUnpause(ctx interface{}, adminAddress, reason string) (*response.TransactionPreparation, error) {
	if m.prepareEmergencyUnpauseFn != nil {
		return m.prepareEmergencyUnpauseFn(ctx, adminAddress, reason)
	}
	return nil, fmt.Errorf("not implemented")
}

/*
@function PrepareDrainFees (mock)
@desc Mock returning controlled drain-fees transaction data.
*/
func (m *mockAdminEmergencyService) PrepareDrainFees(ctx interface{}, recipient, adminAddress, reason string) (*response.TransactionPreparation, error) {
	if m.prepareDrainFeesFn != nil {
		return m.prepareDrainFeesFn(ctx, recipient, adminAddress, reason)
	}
	return nil, fmt.Errorf("not implemented")
}

func TestAdminEmergencyHandler_PrepareEmergencyPause(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		svc := &mockAdminEmergencyService{
			prepareEmergencyPauseFn: func(_ interface{}, _, _ string) (*response.TransactionPreparation, error) {
				return &response.TransactionPreparation{Target: "0xArch", TxData: "0xpause", Value: "0", Description: "Emergency pause"}, nil
			},
		}
		h := NewAdminEmergencyHandler(svc)

		body, _ := json.Marshal(map[string]interface{}{"reason": "critical exploit detected"})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/admin/emergency/pause/prepare", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		h.PrepareEmergencyPause(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.True(t, resp["success"].(bool))
	})

	t.Run("invalid body", func(t *testing.T) {
		h := NewAdminEmergencyHandler(&mockAdminEmergencyService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{bad json}`)))
		c.Request.Header.Set("Content-Type", "application/json")
		h.PrepareEmergencyPause(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		svc := &mockAdminEmergencyService{
			prepareEmergencyPauseFn: func(_ interface{}, _, _ string) (*response.TransactionPreparation, error) {
				return nil, fmt.Errorf("rpc error")
			},
		}
		h := NewAdminEmergencyHandler(svc)
		body, _ := json.Marshal(map[string]interface{}{"reason": "test"})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		h.PrepareEmergencyPause(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestAdminEmergencyHandler_PrepareEmergencyUnpause(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		svc := &mockAdminEmergencyService{
			prepareEmergencyUnpauseFn: func(_ interface{}, _, _ string) (*response.TransactionPreparation, error) {
				return &response.TransactionPreparation{Target: "0xArch", TxData: "0xunpause", Value: "0"}, nil
			},
		}
		h := NewAdminEmergencyHandler(svc)

		body, _ := json.Marshal(map[string]interface{}{"reason": "exploit patched"})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/admin/emergency/unpause/prepare", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		h.PrepareEmergencyUnpause(c)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := NewAdminEmergencyHandler(&mockAdminEmergencyService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{bad}`)))
		c.Request.Header.Set("Content-Type", "application/json")
		h.PrepareEmergencyUnpause(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAdminEmergencyHandler_PrepareDrainFees(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		svc := &mockAdminEmergencyService{
			prepareDrainFeesFn: func(_ interface{}, recipient, _, _ string) (*response.TransactionPreparation, error) {
				return &response.TransactionPreparation{Target: "0xArch", TxData: "0xdrain", Value: "0"}, nil
			},
		}
		h := NewAdminEmergencyHandler(svc)

		body, _ := json.Marshal(map[string]interface{}{"recipient": "0xTreasury", "reason": "quarterly drain"})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/admin/emergency/drain-fees/prepare", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		h.PrepareDrainFees(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.True(t, resp["success"].(bool))
	})

	t.Run("invalid body", func(t *testing.T) {
		h := NewAdminEmergencyHandler(&mockAdminEmergencyService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{bad}`)))
		c.Request.Header.Set("Content-Type", "application/json")
		h.PrepareDrainFees(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
