package request

/*
@file admin_borrowers.go

@desc
Request DTOs for borrower management admin endpoints.
Handles adding/removing borrowers and listing borrower information.

@structure
- AddBorrowerRequest: Add borrower to marketplace
- RemoveBorrowerRequest: Remove borrower from marketplace
- BorrowerFilterRequest: Filter parameters for borrower listing
*/

/*
@struct AddBorrowerRequest

@desc
Request body for adding a new borrower to the marketplace.
Validates borrower address and collateral requirements.

@fields
- borrower_address: Ethereum address of borrower (required, must be valid eth address)
- is_verified: Whether borrower is pre-verified (optional)
*/
type AddBorrowerRequest struct {
	BorrowerAddress string `json:"borrower_address" binding:"required,eth_addr"`
	IsVerified      bool   `json:"is_verified"`
}

/*
@struct RemoveBorrowerRequest

@desc
Request body for removing borrower from marketplace.
Requires reason for audit trail.

@fields
- reason: Reason for removal (required, max 256 chars)
- close_existing_positions: Whether to force-close all borrower positions (optional)
*/
type RemoveBorrowerRequest struct {
	Reason                 string `json:"reason" binding:"required,max=256"`
	CloseExistingPositions bool   `json:"close_existing_positions"`
}

/*
@struct BorrowerFilterRequest

@desc
Query parameters for filtering borrower list.
Supports pagination and filtering by status.

@fields
- page: Page number for pagination (default 1)
- limit: Items per page (default 20, max 100)
- status: Filter by borrower status (verified, pending, blocked)
- search: Search by borrower address or name (optional)
*/
type BorrowerFilterRequest struct {
	Page   int32  `form:"page" binding:"gte=1"`
	Limit  int32  `form:"limit" binding:"gte=1,lte=100"`
	Status string `form:"status"`
	Search string `form:"search"`
}
