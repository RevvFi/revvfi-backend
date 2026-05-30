package response

/*
@file admin_auth.go

@desc
Response DTOs for admin authentication endpoints.
Returns admin verification results, role information, and token details.
*/

/*
@struct AdminCheckResponse

@desc
Response containing admin status and permission details.
Returned by admin verification endpoint.

@fields
- IsAdmin: whether address has admin privileges
- Role: admin role level (OWNER, MANAGER, VIEWER)
- Permissions: list of specific permissions granted
- ArchControllerAddress: architecture controller contract address
*/
type AdminCheckResponse struct {
	IsAdmin               bool     `json:"is_admin"`
	Role                  string   `json:"role,omitempty"`
	Permissions           []string `json:"permissions,omitempty"`
	ArchControllerAddress string   `json:"arch_controller_address,omitempty"`
}

/*
@struct ImpersonateResponse

@desc
Response containing impersonation token and validity details.
Token should be used in Authorization header for subsequent requests.

@fields
- AdminAddress: address being impersonated
- Token: JWT token for authentication
- Reason: reason provided for impersonation
- ExpiresAt: unix timestamp when token expires
- ValidForSeconds: duration token is valid (convenience field)
*/
type ImpersonateResponse struct {
	AdminAddress    string `json:"admin_address"`
	Token           string `json:"token"`
	Reason          string `json:"reason"`
	ExpiresAt       int64  `json:"expires_at"`
	ValidForSeconds int64  `json:"valid_for_seconds"`
}

/*
@struct AdminInfo

@desc
Information about a single admin address.
Used in admin list responses.

@fields
- Address: wallet address with admin privileges
- Role: admin role level
- Permissions: list of granted permissions
- AddedAt: unix timestamp when admin was added
*/
type AdminInfo struct {
	Address     string   `json:"address"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	AddedAt     int64    `json:"added_at"`
}

/*
@struct AdminListResponse

@desc
Paginated list of admin addresses and their roles.

@fields
- Admins: list of admin information
- Pagination: pagination metadata
*/
type AdminListResponse struct {
	Admins     []AdminInfo     `json:"admins"`
	Pagination PaginationInfo  `json:"pagination"`
}

/*
@struct PaginationInfo

@desc
Metadata for paginated responses.

@fields
- Page: current page number (1-indexed)
- Limit: results per page
- Total: total number of items
*/
type PaginationInfo struct {
	Page  int32 `json:"page"`
	Limit int32 `json:"limit"`
	Total int64 `json:"total"`
}
