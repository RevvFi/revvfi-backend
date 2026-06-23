package request

/*
@file offer.go

@desc
Request DTOs for offer endpoints.
Handles offer creation, quotes, and simulations.
*/

/*
@struct CreateOfferRequest

@desc
Request to create a new lender offer.
Builds transaction payload for blockchain submission.

@fields
- MarketAddress: target market address
- Amount: liquidity amount offered
- APR: annual percentage rate in bps
- Seniority: 0=Senior, 1=Junior tier
- ExpiryDays: days until offer expiration
*/
type CreateOfferRequest struct {
	MarketAddress string `json:"market_address" binding:"required,eth_addr"`
	Amount        string `json:"amount" binding:"required,numeric"`
	APR           int32  `json:"apr" binding:"required,min=1,max=5000"`
	Seniority     int16  `json:"seniority" binding:"omitempty,oneof=0 1"`
	ExpiryDays    int32  `json:"expiry_days" binding:"required,min=1,max=365"`
}

/*
@struct OfferQuoteRequest

@desc
Request to calculate optimal offer matching and borrow routing.
Pure computation - no blockchain mutation.

@fields
- MarketAddress: target market
- BorrowAmount: amount borrower wants to borrow
- CollateralAmount: collateral to provide
- PreferredAPR: maximum APR lender willing to pay
*/
type OfferQuoteRequest struct {
	MarketAddress   string `json:"market_address" binding:"required,eth_addr"`
	BorrowAmount    string `json:"borrow_amount" binding:"required,numeric"`
	CollateralAmount string `json:"collateral_amount" binding:"required,numeric"`
	PreferredAPR    int32  `json:"preferred_apr" binding:"omitempty,min=1,max=5000"`
}

/*
@struct OfferSimulateRequest

@desc
Request to simulate offer submission before blockchain execution.
Validates protocol constraints without mutation.

@fields
- MarketAddress: target market
- Amount: offer amount
- APR: offer APR
- Seniority: seniority tier
*/
type OfferSimulateRequest struct {
	MarketAddress string `json:"market_address" binding:"required,eth_addr"`
	Amount        string `json:"amount" binding:"required,numeric"`
	APR           int32  `json:"apr" binding:"required,min=1,max=5000"`
	Seniority     int16  `json:"seniority" binding:"omitempty,oneof=0 1"`
}

/*
@struct OfferListQuery

@desc
Query parameters for offer listing.

@fields
- Page: pagination
- PageSize: results per page
- MarketAddress: filter by market
- SortBy: sort column
- Order: asc|desc
*/
type OfferListQuery struct {
	Page          int32  `form:"page" binding:"omitempty,min=1"`
	PageSize      int32  `form:"page_size" binding:"omitempty,min=1,max=100"`
	MarketAddress string `form:"market_address" binding:"omitempty,eth_addr"`
	SortBy        string `form:"sort_by" binding:"omitempty,oneof=apr liquidity expiry"`
	Order         string `form:"order" binding:"omitempty,oneof=asc desc"`
}
