package response

/*
@file position.go

@desc
Response DTOs for position endpoints.
Returns position data, valuations, and portfolios.
*/

/*
@struct PositionResponse

@desc
Response containing detailed lender position information.

@fields
- TokenID: NFT token ID
- Lender: position owner
- MarketAddress: target market
- Principal: original deposit
- CurrentValue: principal + accrued interest
- AccruedInterest: unpaid interest
- APR: annual percentage rate
- Seniority: 0=Senior, 1=Junior
- Status: active|settled|liquidated
- MintedAt: creation timestamp
- SettledAt: settlement timestamp
*/
type PositionResponse struct {
	TokenID          int64  `json:"token_id"`
	Lender           string `json:"lender"`
	MarketAddress    string `json:"market_address"`
	Principal        string `json:"principal"`
	CurrentValue     string `json:"current_value"`
	AccruedInterest  string `json:"accrued_interest"`
	ClaimableAmount  string `json:"claimable_amount"`
	APR              int32  `json:"apr"`
	Seniority        int16  `json:"seniority"`
	Status           string `json:"status"`
	MintedAt         int64  `json:"minted_at"`
	SettledAt        *int64 `json:"settled_at,omitempty"`
}

/*
@struct PortfolioSummaryResponse

@desc
Aggregated lender portfolio summary.

@fields
- TotalSupplied: total capital supplied
- TotalValue: current portfolio value
- EarnedInterest: total interest earned
- UnrealizedEarnings: unrealized interest
- AvgAPR: weighted average APR
- PositionCount: number of positions
- ActivePositions: number of active positions
- SettledPositions: number of settled positions
*/
type PortfolioSummaryResponse struct {
	TotalSupplied       string  `json:"total_supplied"`
	TotalValue          string  `json:"total_value"`
	EarnedInterest      string  `json:"earned_interest"`
	UnrealizedEarnings  string  `json:"unrealized_earnings"`
	AvgAPR              float64 `json:"avg_apr"`
	PositionCount       int32   `json:"position_count"`
	ActivePositions     int32   `json:"active_positions"`
	SettledPositions    int32   `json:"settled_positions"`
	ClaimableAmount     string  `json:"claimable_amount"`
}

/*
@struct ClaimableResponse

@desc
Claimable positions and reward aggregates.

@fields
- ClaimableAmount: total withdrawable
- ClaimablePositions: positions ready to claim
- PendingClaims: claims waiting for processing
- UnclaimedInterest: unclaimed interest earnings
*/
type ClaimableResponse struct {
	ClaimableAmount     string                `json:"claimable_amount"`
	ClaimablePositions  []PositionResponse    `json:"claimable_positions"`
	PendingClaims       string                `json:"pending_claims"`
	UnclaimedInterest   string                `json:"unclaimed_interest"`
}
