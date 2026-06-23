package request

/*
@file admin_reputation.go

@desc
Request DTOs for borrower reputation management admin endpoints.

@responsibilities
- Validate reputation score adjustment requests
*/

/*
@struct SetReputationRequest

@desc
Request body for preparing a borrower reputation adjustment transaction.

@fields
- NewScore: new reputation score to set (0-1000)
- Reason: required justification for manual override
*/
type SetReputationRequest struct {
	NewScore int    `json:"new_score" binding:"required,min=0,max=1000"`
	Reason   string `json:"reason" binding:"required,max=500"`
}
