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
@file admin_reputation_test.go

@desc
Unit tests for borrower reputation management handlers.
Covers all three endpoints with success and error paths.
*/

/*
@test mockAdminReputationService

@desc
Mock implementation of AdminReputationServiceInterface for handler testing.
*/
type mockAdminReputationService struct {
	getBorrowerReputationFn  func(ctx interface{}, address string) (*response.BorrowerReputation, error)
	prepareSetReputationFn   func(ctx interface{}, address string, newScore int32, adminAddress, reason string) (*response.TransactionPreparation, error)
	getDefaultedBorrowersFn  func(ctx interface{}, page, limit int32) (*response.DefaultedBorrowersResponse, error)
}

/*
@function GetBorrowerReputation (mock)
@desc Mock implementation returning controlled borrower reputation data.
*/
func (m *mockAdminReputationService) GetBorrowerReputation(ctx interface{}, address string) (*response.BorrowerReputation, error) {
	if m.getBorrowerReputationFn != nil {
		return m.getBorrowerReputationFn(ctx, address)
	}
	return nil, fmt.Errorf("not implemented")
}

/*
@function PrepareSetReputation (mock)
@desc Mock implementation returning controlled transaction preparation.
*/
func (m *mockAdminReputationService) PrepareSetReputation(ctx interface{}, address string, newScore int32, adminAddress, reason string) (*response.TransactionPreparation, error) {
	if m.prepareSetReputationFn != nil {
		return m.prepareSetReputationFn(ctx, address, newScore, adminAddress, reason)
	}
	return nil, fmt.Errorf("not implemented")
}

/*
@function GetDefaultedBorrowers (mock)
@desc Mock implementation returning controlled defaulted borrowers list.
*/
func (m *mockAdminReputationService) GetDefaultedBorrowers(ctx interface{}, page, limit int32) (*response.DefaultedBorrowersResponse, error) {
	if m.getDefaultedBorrowersFn != nil {
		return m.getDefaultedBorrowersFn(ctx, page, limit)
	}
	return nil, fmt.Errorf("not implemented")
}

func TestAdminReputationHandler_GetBorrowerReputation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		svc := &mockAdminReputationService{
			getBorrowerReputationFn: func(_ interface{}, addr string) (*response.BorrowerReputation, error) {
				return &response.BorrowerReputation{
					Address:         addr,
					ReputationScore: 750,
					RiskLabel:       "A",
				}, nil
			},
		}
		h := NewAdminReputationHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/admin/reputation/0xBorrower", nil)
		c.Params = gin.Params{{Key: "address", Value: "0xBorrower"}}

		h.GetBorrowerReputation(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var body map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.True(t, body["success"].(bool))
	})

	t.Run("missing address", func(t *testing.T) {
		h := NewAdminReputationHandler(&mockAdminReputationService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/admin/reputation/", nil)
		c.Params = gin.Params{{Key: "address", Value: ""}}
		h.GetBorrowerReputation(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		svc := &mockAdminReputationService{
			getBorrowerReputationFn: func(_ interface{}, _ string) (*response.BorrowerReputation, error) {
				return nil, fmt.Errorf("not found")
			},
		}
		h := NewAdminReputationHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/admin/reputation/0xBorrower", nil)
		c.Params = gin.Params{{Key: "address", Value: "0xBorrower"}}

		h.GetBorrowerReputation(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestAdminReputationHandler_PrepareSetReputation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		svc := &mockAdminReputationService{
			prepareSetReputationFn: func(_ interface{}, addr string, score int32, _, _ string) (*response.TransactionPreparation, error) {
				return &response.TransactionPreparation{Target: addr, TxData: "0xrep", Value: "0"}, nil
			},
		}
		h := NewAdminReputationHandler(svc)

		body, _ := json.Marshal(map[string]interface{}{"new_score": 800, "reason": "good standing"})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "address", Value: "0xBorrower"}}

		h.PrepareSetReputation(c)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("missing address", func(t *testing.T) {
		h := NewAdminReputationHandler(&mockAdminReputationService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{"new_score":800,"reason":"ok"}`)))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "address", Value: ""}}
		h.PrepareSetReputation(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := NewAdminReputationHandler(&mockAdminReputationService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{bad}`)))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "address", Value: "0xBorrower"}}
		h.PrepareSetReputation(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAdminReputationHandler_GetDefaultedBorrowers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		svc := &mockAdminReputationService{
			getDefaultedBorrowersFn: func(_ interface{}, page, limit int32) (*response.DefaultedBorrowersResponse, error) {
				return &response.DefaultedBorrowersResponse{TotalDefaulted: 0, Borrowers: []response.BorrowerInfo{}}, nil
			},
		}
		h := NewAdminReputationHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/admin/reputation/defaulted?page=1&limit=20", nil)

		h.GetDefaultedBorrowers(c)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
