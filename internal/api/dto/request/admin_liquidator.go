package request

/*
@file admin_liquidator.go

@desc
Request DTOs for liquidator configuration admin endpoints.

@responsibilities
- Validate liquidator parameter updates
- Validate auction stop requests
*/

/*
@struct SetLiquidatorParamsRequest

@desc
Request body for preparing a liquidator parameter update transaction.
All parameters are in basis points unless otherwise noted.

@fields
- AuctionDurationSeconds: Dutch auction duration in seconds (default 259200 = 3 days)
- PriceDropRateBps: hourly price drop rate in basis points (default 500 = 5%)
- MinBidIncrementBps: minimum bid increment in basis points (default 100 = 1%)
- LiquidationIncentiveBps: bonus for successful liquidators in bps
- MaxSlippageBps: maximum allowable price slippage in bps
- Reason: justification for parameter change
*/
type SetLiquidatorParamsRequest struct {
	AuctionDurationSeconds  int    `json:"auction_duration_seconds" binding:"required,min=3600"`
	PriceDropRateBps        int    `json:"price_drop_rate_bps" binding:"required,min=0,max=10000"`
	MinBidIncrementBps      int    `json:"min_bid_increment_bps" binding:"required,min=0,max=5000"`
	LiquidationIncentiveBps int    `json:"liquidation_incentive_bps" binding:"required,min=0,max=5000"`
	MaxSlippageBps          int    `json:"max_slippage_bps" binding:"required,min=0,max=10000"`
	Reason                  string `json:"reason" binding:"max=500"`
}

/*
@struct StopAuctionRequest

@desc
Request body for preparing a transaction to stop an active auction.

@fields
- Reason: required justification for stopping the auction (audit trail)
*/
type StopAuctionRequest struct {
	Reason string `json:"reason" binding:"required,max=500"`
}
