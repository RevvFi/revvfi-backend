package request

/*
@file admin_market_risk.go

@desc
Request DTOs for market risk parameter management admin endpoints.
Covers minimum collateral ratio, liquidation threshold, and oracle address changes.
All mutations return unsigned transaction calldata rather than executing on-chain.

@responsibilities
- Validate market risk parameter updates
- Provide binding tags for Gin request parsing
*/

/*
@struct SetMinCRRequest

@desc
Request body for preparing a minimum collateral ratio update transaction.

@fields
- MinCollateralRatioBps: new minimum collateral ratio in basis points (must be >= 10000)
- Reason: optional reason for the change (max 500 chars, used for audit trail)
*/
type SetMinCRRequest struct {
	MinCollateralRatioBps int    `json:"min_collateral_ratio_bps" binding:"required,min=10000"`
	Reason                string `json:"reason" binding:"max=500"`
}

/*
@struct SetLiqThresholdRequest

@desc
Request body for preparing a liquidation threshold update transaction.

@fields
- LiquidationThresholdBps: new liquidation threshold in basis points
- Reason: optional reason for the change
*/
type SetLiqThresholdRequest struct {
	LiquidationThresholdBps int    `json:"liquidation_threshold_bps" binding:"required,min=5000"`
	Reason                  string `json:"reason" binding:"max=500"`
}

/*
@struct SetOracleRequest

@desc
Request body for preparing an oracle address update transaction.

@fields
- OracleAddress: new Chainlink oracle contract address
- Reason: optional reason for the change
*/
type SetOracleRequest struct {
	OracleAddress string `json:"oracle_address" binding:"required"`
	Reason        string `json:"reason" binding:"max=500"`
}
