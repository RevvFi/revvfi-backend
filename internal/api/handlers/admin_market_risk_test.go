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
@file admin_market_risk_test.go

@desc
Unit tests for market risk parameter management handlers.
Tests all four endpoints with success paths and error cases.
*/

type mockAdminMarketRiskService struct {
	getMarketRiskParamsFn    func(ctx interface{}, address string) (*response.MarketRiskParams, error)
	prepareSetMinCRFn        func(ctx interface{}, address string, minCRBps int32, reason string) (*response.TransactionPreparation, error)
	prepareSetLiqThresholdFn func(ctx interface{}, address string, thresholdBps int32, reason string) (*response.TransactionPreparation, error)
	prepareSetOracleFn       func(ctx interface{}, address, oracleAddress, reason string) (*response.TransactionPreparation, error)
}

func (m *mockAdminMarketRiskService) GetMarketRiskParams(ctx interface{}, address string) (*response.MarketRiskParams, error) {
	if m.getMarketRiskParamsFn != nil {
		return m.getMarketRiskParamsFn(ctx, address)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockAdminMarketRiskService) PrepareSetMinCR(ctx interface{}, address string, minCRBps int32, reason string) (*response.TransactionPreparation, error) {
	if m.prepareSetMinCRFn != nil {
		return m.prepareSetMinCRFn(ctx, address, minCRBps, reason)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockAdminMarketRiskService) PrepareSetLiqThreshold(ctx interface{}, address string, thresholdBps int32, reason string) (*response.TransactionPreparation, error) {
	if m.prepareSetLiqThresholdFn != nil {
		return m.prepareSetLiqThresholdFn(ctx, address, thresholdBps, reason)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockAdminMarketRiskService) PrepareSetOracle(ctx interface{}, address, oracleAddress, reason string) (*response.TransactionPreparation, error) {
	if m.prepareSetOracleFn != nil {
		return m.prepareSetOracleFn(ctx, address, oracleAddress, reason)
	}
	return nil, fmt.Errorf("not implemented")
}

func TestAdminMarketRiskHandler_GetMarketRiskParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		svc := &mockAdminMarketRiskService{
			getMarketRiskParamsFn: func(_ interface{}, addr string) (*response.MarketRiskParams, error) {
				return &response.MarketRiskParams{
					MarketAddress:           addr,
					MinCollateralRatioBps:   11000,
					LiquidationThresholdBps: 9500,
					CollateralOracle:        "0xOracle",
					IsActive:                true,
					IsLiquidating:           false,
				}, nil
			},
		}
		h := NewAdminMarketRiskHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/admin/markets/0xMarket/risk", nil)
		c.Params = gin.Params{{Key: "address", Value: "0xMarket"}}

		h.GetMarketRiskParams(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var body map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.True(t, body["success"].(bool))
	})

	t.Run("missing address", func(t *testing.T) {
		h := NewAdminMarketRiskHandler(&mockAdminMarketRiskService{})

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/admin/markets//risk", nil)
		c.Params = gin.Params{{Key: "address", Value: ""}}

		h.GetMarketRiskParams(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		svc := &mockAdminMarketRiskService{
			getMarketRiskParamsFn: func(_ interface{}, _ string) (*response.MarketRiskParams, error) {
				return nil, fmt.Errorf("db error")
			},
		}
		h := NewAdminMarketRiskHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/admin/markets/0xMarket/risk", nil)
		c.Params = gin.Params{{Key: "address", Value: "0xMarket"}}

		h.GetMarketRiskParams(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestAdminMarketRiskHandler_PrepareSetMinCR(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		svc := &mockAdminMarketRiskService{
			prepareSetMinCRFn: func(_ interface{}, addr string, bps int32, _ string) (*response.TransactionPreparation, error) {
				return &response.TransactionPreparation{
					Target:      addr,
					TxData:      "0xabc",
					Value:       "0",
					Description: fmt.Sprintf("Set min CR to %d bps", bps),
				}, nil
			},
		}
		h := NewAdminMarketRiskHandler(svc)

		body, _ := json.Marshal(map[string]interface{}{"min_collateral_ratio_bps": 11000, "reason": "test"})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/admin/markets/0xMarket/risk/min-cr/prepare", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "address", Value: "0xMarket"}}

		h.PrepareSetMinCR(c)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := NewAdminMarketRiskHandler(&mockAdminMarketRiskService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/admin/markets/0xMarket/risk/min-cr/prepare", bytes.NewReader([]byte(`{invalid}`)))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "address", Value: "0xMarket"}}
		h.PrepareSetMinCR(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAdminMarketRiskHandler_PrepareSetLiqThreshold(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		svc := &mockAdminMarketRiskService{
			prepareSetLiqThresholdFn: func(_ interface{}, addr string, bps int32, _ string) (*response.TransactionPreparation, error) {
				return &response.TransactionPreparation{Target: addr, TxData: "0xdef", Value: "0"}, nil
			},
		}
		h := NewAdminMarketRiskHandler(svc)
		body, _ := json.Marshal(map[string]interface{}{"liquidation_threshold_bps": 9500, "reason": "test"})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "address", Value: "0xMarket"}}
		h.PrepareSetLiqThreshold(c)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestAdminMarketRiskHandler_PrepareSetOracle(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		svc := &mockAdminMarketRiskService{
			prepareSetOracleFn: func(_ interface{}, addr, oracle, _ string) (*response.TransactionPreparation, error) {
				return &response.TransactionPreparation{Target: addr, TxData: "0xoracle", Value: "0"}, nil
			},
		}
		h := NewAdminMarketRiskHandler(svc)
		body, _ := json.Marshal(map[string]interface{}{"oracle_address": "0xOracle", "reason": "test"})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "address", Value: "0xMarket"}}
		h.PrepareSetOracle(c)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
