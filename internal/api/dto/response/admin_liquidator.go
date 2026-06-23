package response

/*
@file admin_liquidator.go

@desc
Response DTOs for liquidator configuration and active auction admin endpoints.

@responsibilities
- Structure liquidator parameter query responses
- Structure active auction listing responses
*/

/*
@struct LiquidatorConfig

@desc
Current liquidator contract configuration parameters.

@fields
- AuctionDurationSeconds: Dutch auction duration in seconds (default 259200 = 3 days)
- PriceDropRateBps: hourly price drop rate in basis points (default 500 = 5%)
- MinBidIncrementBps: minimum required bid increment in basis points
- LiquidationIncentiveBps: bonus awarded to successful liquidators in basis points
- MaxSlippageBps: maximum allowable price slippage in basis points
- UpdatedAt: Unix timestamp of last configuration update
- UpdatedBy: admin address that performed the last update
*/
type LiquidatorConfig struct {
	AuctionDurationSeconds  int    `json:"auction_duration_seconds"`
	PriceDropRateBps        int    `json:"price_drop_rate_bps"`
	MinBidIncrementBps      int    `json:"min_bid_increment_bps"`
	LiquidationIncentiveBps int    `json:"liquidation_incentive_bps"`
	MaxSlippageBps          int    `json:"max_slippage_bps"`
	UpdatedAt               int64  `json:"updated_at"`
	UpdatedBy               string `json:"updated_by"`
}

/*
@struct ActiveAuction

@desc
Summary of a currently active liquidation auction.

@fields
- AuctionID: unique blockchain auction identifier
- MarketAddress: market contract being liquidated
- BorrowerAddress: address of the borrower being liquidated
- CollateralAmount: collateral amount being auctioned (wei string)
- DebtAmount: total debt triggering the liquidation (wei string)
- CurrentPrice: current Dutch auction price (wei string)
- StartTime: auction start Unix timestamp
- EndTime: auction end Unix timestamp
- Status: current auction status
*/
type ActiveAuction struct {
	AuctionID        int64  `json:"auction_id"`
	MarketAddress    string `json:"market_address"`
	BorrowerAddress  string `json:"borrower_address"`
	CollateralAmount string `json:"collateral_amount"`
	DebtAmount       string `json:"debt_amount"`
	CurrentPrice     string `json:"current_price"`
	StartTime        int64  `json:"start_time"`
	EndTime          int64  `json:"end_time"`
	Status           string `json:"status"`
}

/*
@struct ActiveAuctionsResponse

@desc
Response wrapper for the active auctions listing endpoint.

@fields
- Auctions: list of currently active auctions
- Total: total count of active auctions
*/
type ActiveAuctionsResponse struct {
	Auctions []ActiveAuction `json:"auctions"`
	Total    int             `json:"total"`
}
