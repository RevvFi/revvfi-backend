package request

/*
@file borrower_request.go

@desc
Request DTOs for the off-chain borrower access request queue.
*/

/*
@struct RejectBorrowerRequest

@desc
Admin decision body for rejecting a pending borrower request.
*/
type RejectBorrowerRequest struct {
	Note string `json:"note" binding:"omitempty,max=500"`
}
