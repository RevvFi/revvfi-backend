package response

/*
@file liquidation.go

@desc
Response DTOs for liquidation and auction endpoints.
Returns liquidation opportunities and auction state.
*/

/*
@struct LiquidatableResponse

@desc
Response containing liquidatable markets.

@fields
- Markets: markets eligible for liquidation
- Count: count of liquidatable markets
- TotalCollateral: total collateral value
- TotalDebt: total debt value
*/
type LiquidatableResponse struct {
	Markets        []LiquidatableMarket `json:"markets"`
	Count          int32                `json:"count"`
	TotalCollateral string               `json:"total_collateral"`
	TotalDebt      string               `json:"total_debt"`
}

/*
@struct LiquidatableMarket

@desc
Details of a single liquidatable market.

@fields
- MarketAddress: market address
- Borrower: borrower address
- CollateralAmount: collateral amount
- DebtAmount: debt amount
- CollateralRatio: current ratio
- HealthFactor: market health
*/
type LiquidatableMarket struct {
	MarketAddress   string  `json:"market_address"`
	Borrower        string  `json:"borrower"`
	CollateralAmount string  `json:"collateral_amount"`
	DebtAmount      string  `json:"debt_amount"`
	CollateralRatio float64 `json:"collateral_ratio"`
	HealthFactor    float64 `json:"health_factor"`
}

/*
@struct AuctionResponse

@desc
Response containing auction details.

@fields
- AuctionID: auction identifier
- MarketAddress: market being auctioned
- Borrower: borrower address
- CollateralAmount: collateral being sold
- DebtAmount: debt to recover
- CurrentPrice: current Dutch auction price
- HighestBid: current highest bid
- HighestBidder: bidder address
- Status: active|settled|cancelled|failed
- TimeRemaining: seconds until auction ends
- StartTime: auction start
- EndTime: auction end
*/
type AuctionResponse struct {
	AuctionID        int64  `json:"auction_id"`
	MarketAddress    string `json:"market_address"`
	Borrower         string `json:"borrower"`
	CollateralAmount string `json:"collateral_amount"`
	DebtAmount       string `json:"debt_amount"`
	CurrentPrice     string `json:"current_price"`
	HighestBid       string `json:"highest_bid"`
	HighestBidder    string `json:"highest_bidder,omitempty"`
	Status           string `json:"status"`
	TimeRemaining    int64  `json:"time_remaining"`
	StartTime        int64  `json:"start_time"`
	EndTime          int64  `json:"end_time"`
}

/*
@struct AuctionPriceResponse

@desc
Response containing real-time Dutch auction price calculation.

@fields
- AuctionID: auction identifier
- CurrentPrice: current bid price
- MinPrice: reserve price (minimum)
- PriceDecay: price decay rate
- TimeRemaining: seconds until auction ends
- TimeElapsed: seconds since auction started
*/
type AuctionPriceResponse struct {
	AuctionID       int64  `json:"auction_id"`
	CurrentPrice    string `json:"current_price"`
	MinPrice        string `json:"min_price"`
	PriceDecay      float64 `json:"price_decay"`
	TimeRemaining   int64  `json:"time_remaining"`
	TimeElapsed     int64  `json:"time_elapsed"`
}

/*
@struct LiquidationHistoryResponse

@desc
Response containing liquidation history.

@fields
- Events: liquidation events
- Count: count of events
- TotalRecovered: total recovered value
- AverageRecoveryRate: recovery rate percentage
*/
type LiquidationHistoryResponse struct {
	Events              []LiquidationEvent `json:"events"`
	Count               int32              `json:"count"`
	TotalRecovered      string             `json:"total_recovered"`
	AverageRecoveryRate float64            `json:"average_recovery_rate"`
}

/*
@struct LiquidationEvent

@desc
Details of a liquidation event.

@fields
- MarketAddress: market address
- Borrower: borrower address
- CollateralAmount: collateral liquidated
- DebtAtLiquidation: debt at time of liquidation
- RecoveryRate: percentage recovered
- InitiatedAt: liquidation start
- CompletedAt: liquidation completion
*/
type LiquidationEvent struct {
	MarketAddress       string  `json:"market_address"`
	Borrower            string  `json:"borrower"`
	CollateralAmount    string  `json:"collateral_amount"`
	DebtAtLiquidation   string  `json:"debt_at_liquidation"`
	RecoveryRate        float64 `json:"recovery_rate"`
	InitiatedAt         int64   `json:"initiated_at"`
	CompletedAt         *int64  `json:"completed_at,omitempty"`
}
