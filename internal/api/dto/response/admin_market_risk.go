package response

/*
@file admin_market_risk.go

@desc
Response DTOs for market risk parameter admin endpoints.

@responsibilities
- Structure market risk parameter query responses
- Provide typed fields for risk configuration display
*/

/*
@struct MarketRiskParams

@desc
Current risk parameters configured for a specific market.

@fields
- MarketAddress: contract address of the market
- MinCollateralRatioBps: minimum collateral ratio in basis points (e.g. 11000 = 110%)
- LiquidationThresholdBps: liquidation trigger threshold in basis points
- CollateralOracle: Chainlink oracle contract address providing price feed
- IsActive: whether the market is currently active
- IsLiquidating: whether the market is currently in liquidation
*/
type MarketRiskParams struct {
	MarketAddress           string `json:"market_address"`
	MinCollateralRatioBps   int32  `json:"min_collateral_ratio_bps"`
	LiquidationThresholdBps int32  `json:"liquidation_threshold_bps"`
	CollateralOracle        string `json:"collateral_oracle"`
	IsActive                bool   `json:"is_active"`
	IsLiquidating           bool   `json:"is_liquidating"`
}
