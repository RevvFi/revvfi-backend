package request

/*
@file admin_auth.go

@desc
Request DTOs for admin authentication endpoints.
Handles admin verification and impersonation requests.
*/

/*
@struct ImpersonateRequest

@desc
Request body for admin impersonation endpoint.
Used to create temporary token for admin context switching (dev only).

@fields
- AdminAddress: wallet address to impersonate
- Reason: human-readable reason for impersonation (audit trail)
*/
type ImpersonateRequest struct {
	AdminAddress string `json:"admin_address" binding:"required,eth_addr"`
	Reason       string `json:"reason" binding:"required,max=256"`
}
