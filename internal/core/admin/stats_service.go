package admin

import (
	"fmt"

	"github.com/Revvfi/revvfi-backend/internal/api/dto/response"
	"github.com/Revvfi/revvfi-backend/internal/repository/postgres"
)

/*
@file stats_service.go

@desc
Service for admin dashboard statistics and analytics.

@responsibilities
- Return protocol-wide overview statistics
- Return borrower growth metrics
- Return market creation and utilization metrics
- Return protocol revenue analytics
- Return liquidation statistics
- Return position distribution metrics
*/

/*
@struct AdminStatsService

@desc
Implements admin dashboard statistics aggregation.

@fields
- repo: admin repository for cross-table aggregate queries
*/
type AdminStatsService struct {
	repo *postgres.AdminRepository
}

/*
@function NewAdminStatsService

@desc
Creates a new AdminStatsService.

@params
- repo: admin PostgreSQL repository

@returns
- *AdminStatsService
*/
func NewAdminStatsService(repo *postgres.AdminRepository) *AdminStatsService {
	return &AdminStatsService{repo: repo}
}

/*
@method GetOverview

@desc
Returns high-level protocol statistics for the admin dashboard landing page.

@params
- ctx: handler context

@returns
- *response.OverviewStats: aggregate protocol metrics
- error
*/
func (s *AdminStatsService) GetOverview(ctx interface{}) (*response.OverviewStats, error) {
	c := extractCtx(ctx)
	row, err := s.repo.GetOverviewStats(c)
	if err != nil {
		return nil, fmt.Errorf("get overview stats: %w", err)
	}

	health := "healthy"
	if row.ActiveAuctions > 10 {
		health = "degraded"
	}
	return &response.OverviewStats{
		TotalMarkets:    row.TotalMarkets,
		ActiveMarkets:   row.ActiveMarkets,
		TotalBorrowers:  row.TotalBorrowers,
		ActiveBorrowers: row.ActiveBorrowers,
		ActivePositions: row.ActivePositions,
		ActiveOffers:    row.ActiveOffers,
		ActiveAuctions:  row.ActiveAuctions,
		TotalDebt:       row.TotalDebt,
		TotalLiquidity:  row.TotalLiquidity,
		TotalPrincipal:  row.TotalPrincipal,
		ProtocolHealth:  health,
	}, nil
}

/*
@method GetBorrowerStats

@desc
Returns borrower growth and reputation distribution statistics.

@params
- ctx: handler context

@returns
- *response.BorrowerGrowthStats: borrower metrics
- error
*/
func (s *AdminStatsService) GetBorrowerStats(ctx interface{}) (*response.BorrowerGrowthStats, error) {
	c := extractCtx(ctx)
	row, err := s.repo.GetBorrowerStats(c)
	if err != nil {
		return nil, fmt.Errorf("get borrower stats: %w", err)
	}
	return &response.BorrowerGrowthStats{
		TotalBorrowers:     row.TotalBorrowers,
		ActiveBorrowers:    row.ActiveBorrowers,
		DefaultedCount:     row.DefaultedCount,
		AverageReputation:  row.AverageReputation,
		AverageSuccessRate: row.AverageSuccessRate,
		TotalVolume:        row.TotalVolume,
	}, nil
}

/*
@method GetMarketStats

@desc
Returns market creation and utilization statistics.

@params
- ctx: handler context

@returns
- *response.MarketCreationStats: market metrics
- error
*/
func (s *AdminStatsService) GetMarketStats(ctx interface{}) (*response.MarketCreationStats, error) {
	c := extractCtx(ctx)
	row, err := s.repo.GetMarketStats(c)
	if err != nil {
		return nil, fmt.Errorf("get market stats: %w", err)
	}
	return &response.MarketCreationStats{
		TotalMarkets:       row.TotalMarkets,
		ActiveMarkets:      row.ActiveMarkets,
		ClosedMarkets:      row.ClosedMarkets,
		LiquidatingMarkets: row.LiquidatingMarkets,
		AverageUtilization: row.AverageUtilization,
		AverageAPR:         row.AverageAPR,
	}, nil
}

/*
@method GetRevenueStats

@desc
Returns protocol revenue analytics aggregated from repayment records.

@params
- ctx: handler context

@returns
- *response.RevenueStats: revenue metrics
- error
*/
func (s *AdminStatsService) GetRevenueStats(ctx interface{}) (*response.RevenueStats, error) {
	c := extractCtx(ctx)
	row, err := s.repo.GetRevenueStats(c)
	if err != nil {
		return nil, fmt.Errorf("get revenue stats: %w", err)
	}
	return &response.RevenueStats{
		TotalInterestCollected: row.TotalInterestCollected,
		TotalRepaid:            row.TotalRepaid,
		TotalRepayments:        row.TotalRepayments,
	}, nil
}

/*
@method GetLiquidationStats

@desc
Returns liquidation auction statistics and recovery metrics.

@params
- ctx: handler context

@returns
- *response.LiquidationStats: liquidation metrics
- error
*/
func (s *AdminStatsService) GetLiquidationStats(ctx interface{}) (*response.LiquidationStats, error) {
	c := extractCtx(ctx)
	row, err := s.repo.GetLiquidationStats(c)
	if err != nil {
		return nil, fmt.Errorf("get liquidation stats: %w", err)
	}
	return &response.LiquidationStats{
		TotalAuctions:             row.TotalAuctions,
		ActiveAuctions:            row.ActiveAuctions,
		SettledAuctions:           row.SettledAuctions,
		AverageRecoveryRate:       row.AverageRecoveryRate,
		TotalCollateralLiquidated: row.TotalCollateralLiquidated,
		TotalDebtLiquidated:       row.TotalDebtLiquidated,
	}, nil
}

/*
@method GetPositionStats

@desc
Returns position distribution metrics across seniority tiers.

@params
- ctx: handler context

@returns
- *response.PositionDistributionStats: position metrics
- error
*/
func (s *AdminStatsService) GetPositionStats(ctx interface{}) (*response.PositionDistributionStats, error) {
	c := extractCtx(ctx)
	row, err := s.repo.GetPositionStats(c)
	if err != nil {
		return nil, fmt.Errorf("get position stats: %w", err)
	}
	return &response.PositionDistributionStats{
		TotalPositions:       row.TotalPositions,
		ActivePositions:      row.ActivePositions,
		SettledPositions:     row.SettledPositions,
		SeniorPositions:      row.SeniorPositions,
		JuniorPositions:      row.JuniorPositions,
		TotalPrincipalLocked: row.TotalPrincipalLocked,
		TotalAccruedInterest: row.TotalAccruedInterest,
	}, nil
}
