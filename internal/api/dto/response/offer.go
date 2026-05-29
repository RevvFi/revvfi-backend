package response

/*
@file offer.go

@desc
Response DTOs for offer endpoints.
Returns offer data, quotes, and matching results.
*/

/*
@struct OfferResponse

@desc
Response containing offer details.

@fields
- OfferID: blockchain offer ID
- Lender: offer creator
- MarketAddress: target market
- Amount: total liquidity offered
- RemainingAmount: unfilled amount
- FilledAmount: filled amount
- APR: annual percentage rate
- Seniority: 0=Senior, 1=Junior
- Status: active|partially_filled|filled|cancelled|expired
- FillPercentage: percent of offer filled
- Expiry: expiration timestamp
- CreatedAt: creation timestamp
*/
type OfferResponse struct {
	OfferID          int64   `json:"offer_id"`
	Lender           string  `json:"lender"`
	MarketAddress    string  `json:"market_address"`
	Amount           string  `json:"amount"`
	RemainingAmount  string  `json:"remaining_amount"`
	FilledAmount     string  `json:"filled_amount"`
	APR              int32   `json:"apr"`
	Seniority        int16   `json:"seniority"`
	Status           string  `json:"status"`
	FillPercentage   float64 `json:"fill_percentage"`
	Expiry           int64   `json:"expiry"`
	CreatedAt        int64   `json:"created_at"`
}

/*
@struct OfferQuoteResponse

@desc
Response from offer matching quote calculation.

@fields
- OptimalAPR: best available APR
- MatchedOffers: offers that would be matched
- TotalLiquidity: available liquidity at this APR
- EstimatedRate: effective borrowing rate
- GasEstimate: estimated gas cost
*/
type OfferQuoteResponse struct {
	OptimalAPR        int32            `json:"optimal_apr"`
	MatchedOffers     []OfferResponse  `json:"matched_offers"`
	TotalLiquidity    string           `json:"total_liquidity"`
	EstimatedRate     float64          `json:"estimated_rate"`
	GasEstimate       string           `json:"gas_estimate"`
}

/*
@struct OfferSimulateResponse

@desc
Response from offer submission simulation.

@fields
- IsValid: offer would be valid
- Errors: validation errors (if any)
- Warnings: non-blocking warnings
- EstimatedFillRate: likely fill rate
- EstimatedAPRPosition: where APR ranks
*/
type OfferSimulateResponse struct {
	IsValid              bool     `json:"is_valid"`
	Errors               []string `json:"errors,omitempty"`
	Warnings             []string `json:"warnings,omitempty"`
	EstimatedFillRate    float64  `json:"estimated_fill_rate"`
	EstimatedAPRPosition int32    `json:"estimated_apr_position"`
}

/*
@struct BestOffersResponse

@desc
Response containing best available offers for borrowing.

@fields
- Offers: ranked offers by APR
- TotalLiquidity: total available
- WeightedAPR: average APR
- BestAPR: best (lowest) APR available
*/
type BestOffersResponse struct {
	Offers          []OfferResponse `json:"offers"`
	TotalLiquidity  string          `json:"total_liquidity"`
	WeightedAPR     int32           `json:"weighted_apr"`
	BestAPR         int32           `json:"best_apr"`
}
