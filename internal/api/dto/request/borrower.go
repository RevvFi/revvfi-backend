package request

/*
@file borrower.go

@desc
Request DTOs for borrower endpoints.
Handles borrower registration and market operations.
*/

/*
@struct BorrowerRegistrationRequest

@desc
Request to build borrower registration transaction.

@fields
- WalletAddress: borrower wallet address
- Metadata: optional profile metadata
*/
type BorrowerRegistrationRequest struct {
	WalletAddress string            `json:"wallet_address" binding:"required,eth_addr"`
	Metadata      map[string]string `json:"metadata" binding:"omitempty"`
}

/*
@struct BorrowerQuery

@desc
Query parameters for borrower operations.

@fields
- SortBy: reputation|activity
- Order: asc|desc
*/
type BorrowerQuery struct {
	SortBy string `form:"sort_by" binding:"omitempty,oneof=reputation activity"`
	Order  string `form:"order" binding:"omitempty,oneof=asc desc"`
}
