package request

/*
@file withdrawal.go

@desc
Request DTOs for withdrawal endpoints.
*/

/*
@struct WithdrawalRequest

@desc
Request body for queueing a lender withdrawal.

@fields
- PositionID: lender position NFT token ID
- Amount: withdrawal amount in wei
*/
type WithdrawalRequest struct {
	PositionID int64  `json:"position_id" binding:"required,min=1"`
	Amount     string `json:"amount" binding:"required,numeric"`
}

/*
@struct WithdrawalCancelRequest

@desc
Request body for cancelling a pending withdrawal.

@fields
- RequestID: withdrawal request identifier
*/
type WithdrawalCancelRequest struct {
	RequestID int64 `json:"request_id" binding:"required,min=1"`
}
