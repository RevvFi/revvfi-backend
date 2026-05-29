package response

/*
@file borrower.go

@desc
Response DTOs for borrower endpoints.
Returns borrower reputation and market data.
*/

/*
@struct BorrowerReputationResponse

@desc
Response containing borrower reputation and trust metrics.

@fields
- Address: borrower wallet
- ReputationScore: 0-1000 score
- RiskLabel: AAA|AA|A|B|C|D
- SuccessRate: successful loan percentage
- DefaultRate: defaulted loan percentage
- TotalBorrowed: cumulative borrowed amount
- TotalRepaid: cumulative repaid amount
- OutstandingDebt: current debt
- ActiveLoans: number of active loans
- FailedLoans: number of defaults
*/
type BorrowerReputationResponse struct {
	Address            string  `json:"address"`
	ReputationScore    int32   `json:"reputation_score"`
	RiskLabel          string  `json:"risk_label"`
	SuccessRate        float64 `json:"success_rate"`
	DefaultRate        float64 `json:"default_rate"`
	TotalBorrowed      string  `json:"total_borrowed"`
	TotalRepaid        string  `json:"total_repaid"`
	OutstandingDebt    string  `json:"outstanding_debt"`
	ActiveLoans        int32   `json:"active_loans"`
	FailedLoans        int32   `json:"failed_loans"`
	RegisteredAt       int64   `json:"registered_at"`
	LastActivity       *int64  `json:"last_activity,omitempty"`
}

/*
@struct BorrowerMarketsResponse

@desc
Response with borrower-owned markets.

@fields
- Address: borrower address
- Markets: list of markets created
- TotalMarkets: count of markets
- TotalTVL: aggregate TVL
- AverageUtilization: average utilization
*/
type BorrowerMarketsResponse struct {
	Address                string            `json:"address"`
	Markets                []MarketResponse  `json:"markets"`
	TotalMarkets           int32             `json:"total_markets"`
	TotalTVL               string            `json:"total_tvl"`
	AverageUtilization     float64           `json:"average_utilization"`
}

/*
@struct LiquidationRiskResponse

@desc
Response containing borrower liquidation risk analysis.

@fields
- Address: borrower address
- HealthFactor: market health factor
- CollateralRatio: current collateral ratio
- LiquidationThreshold: threshold level
- RiskLevel: low|medium|high|critical
- TimeToLiquidation: estimated time (hours)
- RecommendedAction: suggested action
*/
type LiquidationRiskResponse struct {
	Address               string  `json:"address"`
	HealthFactor          float64 `json:"health_factor"`
	CollateralRatio       float64 `json:"collateral_ratio"`
	LiquidationThreshold  float64 `json:"liquidation_threshold"`
	RiskLevel             string  `json:"risk_level"`
	TimeToLiquidation     *int64  `json:"time_to_liquidation,omitempty"`
	RecommendedAction     string  `json:"recommended_action,omitempty"`
}
