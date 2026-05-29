package request

/*
@file position.go

@desc
Request DTOs for position endpoints.
Handles position claims and withdrawal requests.
*/

/*
@struct ClaimPositionRequest

@desc
Request to claim settled position (principal + interest).

@fields
- TokenID: position NFT token ID
- Amount: optional specific amount to claim
*/
type ClaimPositionRequest struct {
	TokenID int64  `json:"token_id" binding:"required,min=1"`
	Amount  string `json:"amount" binding:"omitempty,numeric"`
}

/*
@struct PositionListQuery

@desc
Query parameters for position listing.

@fields
- Page: pagination
- PageSize: results per page
- Status: active|settled|liquidated
- SortBy: value|date
- Order: asc|desc
*/
type PositionListQuery struct {
	Page     int32  `form:"page" binding:"omitempty,min=1"`
	PageSize int32  `form:"page_size" binding:"omitempty,min=1,max=100"`
	Status   string `form:"status" binding:"omitempty,oneof=active settled liquidated"`
	SortBy   string `form:"sort_by" binding:"omitempty,oneof=value date"`
	Order    string `form:"order" binding:"omitempty,oneof=asc desc"`
}
