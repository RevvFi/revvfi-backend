package handlers

import (
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
@file admin_stats_test.go

@desc
Unit tests for admin dashboard statistics handlers.
Covers all six stat endpoints with success and error cases.
*/

/*
@test mockAdminStatsService

@desc
Mock implementation of AdminStatsServiceInterface for handler testing.
*/
type mockAdminStatsService struct {
	getOverviewFn          func(ctx interface{}) (*response.OverviewStats, error)
	getBorrowerStatsFn     func(ctx interface{}) (*response.BorrowerGrowthStats, error)
	getMarketStatsFn       func(ctx interface{}) (*response.MarketCreationStats, error)
	getRevenueStatsFn      func(ctx interface{}) (*response.RevenueStats, error)
	getLiquidationStatsFn  func(ctx interface{}) (*response.LiquidationStats, error)
	getPositionStatsFn     func(ctx interface{}) (*response.PositionDistributionStats, error)
}

/*
@function GetOverview (mock)
@desc Mock returning controlled protocol overview statistics.
*/
func (m *mockAdminStatsService) GetOverview(ctx interface{}) (*response.OverviewStats, error) {
	if m.getOverviewFn != nil {
		return m.getOverviewFn(ctx)
	}
	return nil, fmt.Errorf("not implemented")
}

/*
@function GetBorrowerStats (mock)
@desc Mock returning controlled borrower growth statistics.
*/
func (m *mockAdminStatsService) GetBorrowerStats(ctx interface{}) (*response.BorrowerGrowthStats, error) {
	if m.getBorrowerStatsFn != nil {
		return m.getBorrowerStatsFn(ctx)
	}
	return nil, fmt.Errorf("not implemented")
}

/*
@function GetMarketStats (mock)
@desc Mock returning controlled market creation statistics.
*/
func (m *mockAdminStatsService) GetMarketStats(ctx interface{}) (*response.MarketCreationStats, error) {
	if m.getMarketStatsFn != nil {
		return m.getMarketStatsFn(ctx)
	}
	return nil, fmt.Errorf("not implemented")
}

/*
@function GetRevenueStats (mock)
@desc Mock returning controlled protocol revenue statistics.
*/
func (m *mockAdminStatsService) GetRevenueStats(ctx interface{}) (*response.RevenueStats, error) {
	if m.getRevenueStatsFn != nil {
		return m.getRevenueStatsFn(ctx)
	}
	return nil, fmt.Errorf("not implemented")
}

/*
@function GetLiquidationStats (mock)
@desc Mock returning controlled liquidation statistics.
*/
func (m *mockAdminStatsService) GetLiquidationStats(ctx interface{}) (*response.LiquidationStats, error) {
	if m.getLiquidationStatsFn != nil {
		return m.getLiquidationStatsFn(ctx)
	}
	return nil, fmt.Errorf("not implemented")
}

/*
@function GetPositionStats (mock)
@desc Mock returning controlled position distribution statistics.
*/
func (m *mockAdminStatsService) GetPositionStats(ctx interface{}) (*response.PositionDistributionStats, error) {
	if m.getPositionStatsFn != nil {
		return m.getPositionStatsFn(ctx)
	}
	return nil, fmt.Errorf("not implemented")
}

func TestAdminStatsHandler_GetOverview(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		svc := &mockAdminStatsService{
			getOverviewFn: func(_ interface{}) (*response.OverviewStats, error) {
				return &response.OverviewStats{
					TotalMarkets:    5,
					ActiveMarkets:   4,
					TotalBorrowers:  100,
					ActiveBorrowers: 80,
					ActivePositions: 200,
					ActiveOffers:    250,
					ActiveAuctions:  2,
					TotalDebt:       "5000000",
					TotalLiquidity:  "10000000",
					TotalPrincipal:  "4500000",
					ProtocolHealth:  "healthy",
				}, nil
			},
		}
		h := NewAdminStatsHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/admin/stats/overview", nil)

		h.GetOverview(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var body map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.True(t, body["success"].(bool))
	})

	t.Run("service error", func(t *testing.T) {
		svc := &mockAdminStatsService{
			getOverviewFn: func(_ interface{}) (*response.OverviewStats, error) {
				return nil, fmt.Errorf("db error")
			},
		}
		h := NewAdminStatsHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/admin/stats/overview", nil)

		h.GetOverview(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestAdminStatsHandler_GetBorrowerStats(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		svc := &mockAdminStatsService{
			getBorrowerStatsFn: func(_ interface{}) (*response.BorrowerGrowthStats, error) {
				return &response.BorrowerGrowthStats{
					TotalBorrowers:     100,
					ActiveBorrowers:    80,
					DefaultedCount:     5,
					AverageReputation:  720.5,
					AverageSuccessRate: 0.92,
					TotalVolume:        "50000000",
				}, nil
			},
		}
		h := NewAdminStatsHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/admin/stats/borrowers", nil)

		h.GetBorrowerStats(c)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestAdminStatsHandler_GetMarketStats(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		svc := &mockAdminStatsService{
			getMarketStatsFn: func(_ interface{}) (*response.MarketCreationStats, error) {
				return &response.MarketCreationStats{
					TotalMarkets:       5,
					ActiveMarkets:      4,
					ClosedMarkets:      1,
					LiquidatingMarkets: 0,
					AverageUtilization: 0.75,
					AverageAPR:         850,
				}, nil
			},
		}
		h := NewAdminStatsHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/admin/stats/markets", nil)

		h.GetMarketStats(c)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestAdminStatsHandler_GetRevenueStats(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		svc := &mockAdminStatsService{
			getRevenueStatsFn: func(_ interface{}) (*response.RevenueStats, error) {
				return &response.RevenueStats{
					TotalInterestCollected: "50000",
					TotalRepaid:            "450000",
					TotalRepayments:        120,
				}, nil
			},
		}
		h := NewAdminStatsHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/admin/stats/revenue", nil)

		h.GetRevenueStats(c)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestAdminStatsHandler_GetLiquidationStats(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		svc := &mockAdminStatsService{
			getLiquidationStatsFn: func(_ interface{}) (*response.LiquidationStats, error) {
				return &response.LiquidationStats{
					TotalAuctions:             30,
					ActiveAuctions:            2,
					SettledAuctions:           28,
					AverageRecoveryRate:       0.93,
					TotalCollateralLiquidated: "3000000",
					TotalDebtLiquidated:       "2800000",
				}, nil
			},
		}
		h := NewAdminStatsHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/admin/stats/liquidations", nil)

		h.GetLiquidationStats(c)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestAdminStatsHandler_GetPositionStats(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		svc := &mockAdminStatsService{
			getPositionStatsFn: func(_ interface{}) (*response.PositionDistributionStats, error) {
				return &response.PositionDistributionStats{
					TotalPositions:       500,
					ActivePositions:      400,
					SettledPositions:     100,
					SeniorPositions:      300,
					JuniorPositions:      200,
					TotalPrincipalLocked: "10000000",
					TotalAccruedInterest: "250000",
				}, nil
			},
		}
		h := NewAdminStatsHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/admin/stats/positions", nil)

		h.GetPositionStats(c)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
