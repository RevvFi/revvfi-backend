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
@file admin_liquidator_test.go

@desc
Unit tests for liquidator configuration and active auction handlers.
Tests all four endpoints with success paths and error cases.
*/

/*
@test mockAdminLiquidatorService

@desc
Mock implementation of AdminLiquidatorServiceInterface for handler testing.
*/
type mockAdminLiquidatorService struct {
	getLiquidatorConfigFn        func(ctx interface{}) (*response.LiquidatorConfig, error)
	prepareSetLiquidatorParamsFn func(ctx interface{}, auctionDurationSecs, priceDropRateBps, minBidIncrementBps, liquidationIncentiveBps, maxSlippageBps int, updatedBy, reason string) (*response.TransactionPreparation, error)
	getActiveAuctionsFn          func(ctx interface{}) (*response.ActiveAuctionsResponse, error)
	prepareStopAuctionFn         func(ctx interface{}, auctionID int64, adminAddress, reason string) (*response.TransactionPreparation, error)
}

/*
@function GetLiquidatorConfig (mock)
@desc Mock implementation returning controlled liquidator config.
*/
func (m *mockAdminLiquidatorService) GetLiquidatorConfig(ctx interface{}) (*response.LiquidatorConfig, error) {
	if m.getLiquidatorConfigFn != nil {
		return m.getLiquidatorConfigFn(ctx)
	}
	return nil, fmt.Errorf("not implemented")
}

/*
@function PrepareSetLiquidatorParams (mock)
@desc Mock implementation returning controlled transaction prep.
*/
func (m *mockAdminLiquidatorService) PrepareSetLiquidatorParams(ctx interface{}, auctionDurationSecs, priceDropRateBps, minBidIncrementBps, liquidationIncentiveBps, maxSlippageBps int, updatedBy, reason string) (*response.TransactionPreparation, error) {
	if m.prepareSetLiquidatorParamsFn != nil {
		return m.prepareSetLiquidatorParamsFn(ctx, auctionDurationSecs, priceDropRateBps, minBidIncrementBps, liquidationIncentiveBps, maxSlippageBps, updatedBy, reason)
	}
	return nil, fmt.Errorf("not implemented")
}

/*
@function GetActiveAuctions (mock)
@desc Mock implementation returning controlled active auctions list.
*/
func (m *mockAdminLiquidatorService) GetActiveAuctions(ctx interface{}) (*response.ActiveAuctionsResponse, error) {
	if m.getActiveAuctionsFn != nil {
		return m.getActiveAuctionsFn(ctx)
	}
	return nil, fmt.Errorf("not implemented")
}

/*
@function PrepareStopAuction (mock)
@desc Mock implementation returning controlled stop-auction tx prep.
*/
func (m *mockAdminLiquidatorService) PrepareStopAuction(ctx interface{}, auctionID int64, adminAddress, reason string) (*response.TransactionPreparation, error) {
	if m.prepareStopAuctionFn != nil {
		return m.prepareStopAuctionFn(ctx, auctionID, adminAddress, reason)
	}
	return nil, fmt.Errorf("not implemented")
}

func TestAdminLiquidatorHandler_GetLiquidatorConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		svc := &mockAdminLiquidatorService{
			getLiquidatorConfigFn: func(_ interface{}) (*response.LiquidatorConfig, error) {
				return &response.LiquidatorConfig{
					AuctionDurationSeconds:  86400,
					PriceDropRateBps:        500,
					MinBidIncrementBps:      100,
					LiquidationIncentiveBps: 300,
					MaxSlippageBps:          200,
				}, nil
			},
		}
		h := NewAdminLiquidatorHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/admin/liquidator/config", nil)

		h.GetLiquidatorConfig(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var body map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.True(t, body["success"].(bool))
	})

	t.Run("service error", func(t *testing.T) {
		svc := &mockAdminLiquidatorService{
			getLiquidatorConfigFn: func(_ interface{}) (*response.LiquidatorConfig, error) {
				return nil, fmt.Errorf("db error")
			},
		}
		h := NewAdminLiquidatorHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/admin/liquidator/config", nil)

		h.GetLiquidatorConfig(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestAdminLiquidatorHandler_PrepareSetLiquidatorParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		svc := &mockAdminLiquidatorService{
			prepareSetLiquidatorParamsFn: func(_ interface{}, _, _, _, _, _ int, _, _ string) (*response.TransactionPreparation, error) {
				return &response.TransactionPreparation{Target: "0xLiquidator", TxData: "0xabc", Value: "0"}, nil
			},
		}
		h := NewAdminLiquidatorHandler(svc)

		body, _ := json.Marshal(map[string]interface{}{
			"auction_duration_seconds":   86400,
			"price_drop_rate_bps":        500,
			"min_bid_increment_bps":      100,
			"liquidation_incentive_bps":  300,
			"max_slippage_bps":           200,
			"reason":                     "tuning",
		})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/admin/liquidator/config/prepare", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		h.PrepareSetLiquidatorParams(c)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		h := NewAdminLiquidatorHandler(&mockAdminLiquidatorService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{invalid}`)))
		c.Request.Header.Set("Content-Type", "application/json")
		h.PrepareSetLiquidatorParams(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAdminLiquidatorHandler_GetActiveAuctions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		svc := &mockAdminLiquidatorService{
			getActiveAuctionsFn: func(_ interface{}) (*response.ActiveAuctionsResponse, error) {
				return &response.ActiveAuctionsResponse{Total: 0, Auctions: []response.ActiveAuction{}}, nil
			},
		}
		h := NewAdminLiquidatorHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/admin/liquidator/auctions", nil)

		h.GetActiveAuctions(c)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestAdminLiquidatorHandler_PrepareStopAuction(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		svc := &mockAdminLiquidatorService{
			prepareStopAuctionFn: func(_ interface{}, id int64, _, _ string) (*response.TransactionPreparation, error) {
				return &response.TransactionPreparation{Target: "0xLiq", TxData: "0xstop", Value: "0"}, nil
			},
		}
		h := NewAdminLiquidatorHandler(svc)

		body, _ := json.Marshal(map[string]interface{}{"reason": "manual stop"})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "auctionID", Value: "42"}}

		h.PrepareStopAuction(c)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid auction id", func(t *testing.T) {
		h := NewAdminLiquidatorHandler(&mockAdminLiquidatorService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/", nil)
		c.Params = gin.Params{{Key: "auctionID", Value: "notanumber"}}
		h.PrepareStopAuction(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
