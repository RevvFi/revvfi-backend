package request

/*
@file admin_emergency.go

@desc
Request DTOs for emergency control admin endpoints.

@responsibilities
- Validate emergency pause/unpause/drain-fees requests
*/

/*
@struct EmergencyActionRequest

@desc
Request body for emergency pause/unpause operations.
Requires a reason for the audit trail.

@fields
- Reason: required justification for the emergency action (max 1000 chars)
*/
type EmergencyActionRequest struct {
	Reason string `json:"reason" binding:"required,max=1000"`
}

/*
@struct DrainFeesRequest

@desc
Request body for preparing an emergency fee drain transaction.

@fields
- Recipient: address to receive the drained fees
- Reason: required justification for draining fees
*/
type DrainFeesRequest struct {
	Recipient string `json:"recipient" binding:"required"`
	Reason    string `json:"reason" binding:"required,max=1000"`
}
