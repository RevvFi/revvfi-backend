package response

/*
@file admin_emergency.go

@desc
Response DTOs for emergency control admin endpoints.
Emergency actions produce on-chain transactions prepared as calldata.

@responsibilities
- Define emergency action response shapes
- Reuse TransactionPreparation for calldata returns
*/

/*
@struct EmergencyStatus

@desc
Current protocol emergency state.

@fields
- IsPaused: whether the protocol is globally paused
- PausedAt: Unix timestamp when the protocol was paused (0 if not paused)
- PausedBy: admin address that triggered the pause
- ActiveAuctions: count of auctions still running under the emergency
*/
type EmergencyStatus struct {
	IsPaused       bool   `json:"is_paused"`
	PausedAt       int64  `json:"paused_at,omitempty"`
	PausedBy       string `json:"paused_by,omitempty"`
	ActiveAuctions int64  `json:"active_auctions"`
}
