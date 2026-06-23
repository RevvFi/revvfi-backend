package response

/*
@file admin_stats.go

@desc
Response DTOs for admin dashboard statistics endpoints.
Provides protocol-wide analytics for monitoring and governance.

@responsibilities
- Define overview statistics response
- Define per-domain statistics responses (borrowers, markets, revenue, liquidations, positions)
*/

/*
@struct OverviewStats

@desc
Protocol-wide summary statistics for admin dashboard.

@fields
- TotalMarkets: total number of markets created
- ActiveMarkets: markets currently operational
- TotalBorrowers: total registered borrowers
- ActiveBorrowers: borrowers with active loans
- ActivePositions: lender positions currently active
- ActiveOffers: lender offers currently open
- ActiveAuctions: liquidation auctions in progress
- TotalDebt: total outstanding debt across all markets (wei string)
- TotalLiquidity: total available liquidity across all markets (wei string)
- TotalPrincipal: total principal deployed across all markets (wei string)
- ProtocolHealth: health indicator (healthy/degraded/critical)
*/
type OverviewStats struct {
	TotalMarkets    int64   `json:"total_markets"`
	ActiveMarkets   int64   `json:"active_markets"`
	TotalBorrowers  int64   `json:"total_borrowers"`
	ActiveBorrowers int64   `json:"active_borrowers"`
	ActivePositions int64   `json:"active_positions"`
	ActiveOffers    int64   `json:"active_offers"`
	ActiveAuctions  int64   `json:"active_auctions"`
	TotalDebt       string  `json:"total_debt_wei"`
	TotalLiquidity  string  `json:"total_liquidity_wei"`
	TotalPrincipal  string  `json:"total_principal_wei"`
	ProtocolHealth  string  `json:"protocol_health"`
}

/*
@struct BorrowerGrowthStats

@desc
Borrower growth metrics and reputation distribution for admin analytics.

@fields
- TotalBorrowers: total registered borrower count
- ActiveBorrowers: count with active loans
- DefaultedCount: count with at least one default
- AverageReputation: mean reputation score across all borrowers
- AverageSuccessRate: mean loan success rate
- TotalVolume: cumulative borrowed amount across all borrowers (wei string)
*/
type BorrowerGrowthStats struct {
	TotalBorrowers    int64   `json:"total_borrowers"`
	ActiveBorrowers   int64   `json:"active_borrowers"`
	DefaultedCount    int64   `json:"defaulted_count"`
	AverageReputation float64 `json:"average_reputation"`
	AverageSuccessRate float64 `json:"average_success_rate"`
	TotalVolume       string  `json:"total_volume_wei"`
}

/*
@struct MarketCreationStats

@desc
Market creation metrics and distribution analytics.

@fields
- TotalMarkets: total markets ever created
- ActiveMarkets: currently operational markets
- ClosedMarkets: permanently closed markets
- LiquidatingMarkets: markets currently under liquidation
- AverageUtilization: mean utilization rate across active markets
- AverageAPR: mean weighted APR across active markets (basis points)
*/
type MarketCreationStats struct {
	TotalMarkets        int64   `json:"total_markets"`
	ActiveMarkets       int64   `json:"active_markets"`
	ClosedMarkets       int64   `json:"closed_markets"`
	LiquidatingMarkets  int64   `json:"liquidating_markets"`
	AverageUtilization  float64 `json:"average_utilization"`
	AverageAPR          float64 `json:"average_apr_bps"`
}

/*
@struct RevenueStats

@desc
Fee revenue analytics from protocol activity.

@fields
- TotalInterestCollected: sum of all interest paid across repayments (wei string)
- TotalRepaid: total amount repaid by borrowers (wei string)
- TotalRepayments: number of repayment transactions
*/
type RevenueStats struct {
	TotalInterestCollected string `json:"total_interest_collected_wei"`
	TotalRepaid            string `json:"total_repaid_wei"`
	TotalRepayments        int64  `json:"total_repayments"`
}

/*
@struct LiquidationStats

@desc
Liquidation statistics and recovery rates.

@fields
- TotalAuctions: total number of liquidation auctions
- ActiveAuctions: currently running auctions
- SettledAuctions: completed auctions
- AverageRecoveryRate: mean recovery rate across settled auctions (0.0-1.0)
- TotalCollateralLiquidated: cumulative collateral liquidated (wei string)
- TotalDebtLiquidated: cumulative debt resolved through liquidation (wei string)
*/
type LiquidationStats struct {
	TotalAuctions              int64   `json:"total_auctions"`
	ActiveAuctions             int64   `json:"active_auctions"`
	SettledAuctions            int64   `json:"settled_auctions"`
	AverageRecoveryRate        float64 `json:"average_recovery_rate"`
	TotalCollateralLiquidated  string  `json:"total_collateral_liquidated_wei"`
	TotalDebtLiquidated        string  `json:"total_debt_liquidated_wei"`
}

/*
@struct PositionDistributionStats

@desc
Position distribution across seniority tiers and status.

@fields
- TotalPositions: all positions ever created
- ActivePositions: positions currently active
- SettledPositions: positions that have been settled
- SeniorPositions: count of senior-tier positions (seniority=0)
- JuniorPositions: count of junior-tier positions (seniority=1)
- TotalPrincipalLocked: total principal across all active positions (wei string)
- TotalAccruedInterest: total accrued interest across all positions (wei string)
*/
type PositionDistributionStats struct {
	TotalPositions      int64  `json:"total_positions"`
	ActivePositions     int64  `json:"active_positions"`
	SettledPositions    int64  `json:"settled_positions"`
	SeniorPositions     int64  `json:"senior_positions"`
	JuniorPositions     int64  `json:"junior_positions"`
	TotalPrincipalLocked string `json:"total_principal_locked_wei"`
	TotalAccruedInterest string `json:"total_accrued_interest_wei"`
}
