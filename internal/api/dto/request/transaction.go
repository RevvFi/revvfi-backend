package request

/*
@file transaction.go

@desc
Request DTOs for transaction build and quote endpoints.
*/

/*
@struct BorrowTransactionRequest

@desc
Request body for building an unsigned borrow transaction.

@fields
- MarketAddress: target market contract
- Amount: borrow amount in wei
- MaxAPR: maximum acceptable APR in basis points
- UseSeniorOnly: whether to only use senior tranche offers (optional, defaults to false)
*/
type BorrowTransactionRequest struct {
	MarketAddress string `json:"market_address" binding:"required,eth_addr"`
	Amount        string `json:"amount" binding:"required,numeric"`
	MaxAPR        int32  `json:"max_apr" binding:"required,min=1,max=5000"`
	UseSeniorOnly bool   `json:"use_senior_only"` // defaults to false if not provided
}

/*
@struct RepayTransactionRequest

@desc
Request body for building an unsigned repayment transaction.

@fields
- MarketAddress: target market contract
- Amount: repayment amount in wei
*/
type RepayTransactionRequest struct {
	MarketAddress string `json:"market_address" binding:"required,eth_addr"`
	Amount        string `json:"amount" binding:"required,numeric"`
}

/*
@struct LiquidateTransactionRequest

@desc
Request body for building an unsigned liquidation bid transaction.

@fields
- MarketAddress: target market contract
- AuctionID: target auction identifier
- BidAmount: bid amount in wei
*/
type LiquidateTransactionRequest struct {
	MarketAddress string `json:"market_address" binding:"required,eth_addr"`
	AuctionID     int64  `json:"auction_id" binding:"required,min=1"`
	BidAmount     string `json:"bid_amount" binding:"required,numeric"`
}

/*
@struct TransactionQuoteRequest

@desc
Request body for estimating transaction gas cost.

@fields
- Type: transaction type
- Params: opaque transaction parameters
*/
type TransactionQuoteRequest struct {
	Type   string                 `json:"type" binding:"required,oneof=borrow repay liquidate withdraw"`
	Params map[string]interface{} `json:"params" binding:"omitempty"`
}
