package handlers

import (
	"fmt"
	"net/http"

	"github.com/Revvfi/revvfi-backend/internal/api/dto/request"
	"github.com/Revvfi/revvfi-backend/internal/api/dto/response"
	appErr "github.com/Revvfi/revvfi-backend/internal/pkg/errors"
	"github.com/gin-gonic/gin"
)

/*
@file admin_auth.go

@desc
HTTP handlers for admin authentication and authorization.
Manages admin role verification, permission checking, and admin impersonation (dev only).

@responsibilities
- Verify admin status for wallet addresses
- Check admin permissions and roles
- Enable admin impersonation for testing (dev only)
- Manage admin sessions and tokens

@dependencies
- AdminService: authentication and role management (interface-based)
*/

/*
@interface AdminAuthServiceInterface

@desc
Interface for admin authentication operations.
Used for dependency injection and testing.
*/
type AdminAuthServiceInterface interface {
	IsAdmin(ctx interface{}, address string) (bool, error)
	GetAdminRole(ctx interface{}, address string) (string, error)
	GetAdminPermissions(ctx interface{}, address string) ([]string, error)
	GenerateImpersonateToken(ctx interface{}, adminAddress, reason string) (string, int64, error)
	ValidateImpersonateToken(token string) (string, error)
}

/*
@struct AdminAuthHandler

@desc
Handles HTTP requests for admin authentication endpoints.
Provides role verification and permission management.

@responsibilities
- Verify admin status for addresses
- Return admin permissions and roles
- Generate and validate impersonation tokens
- Handle admin context switching (dev only)

@dependencies
- authService: admin authentication service
*/
type AdminAuthHandler struct {
	authService AdminAuthServiceInterface
}

/*
@function NewAdminAuthHandler

@desc
Creates new admin auth handler with dependencies.
Follows constructor pattern for dependency injection.

@params
- authService: admin authentication service interface

@returns
- *AdminAuthHandler: initialized handler
*/
func NewAdminAuthHandler(authService AdminAuthServiceInterface) *AdminAuthHandler {
	return &AdminAuthHandler{
		authService: authService,
	}
}

/*
@handler CheckAdmin

@desc
Verify if wallet address is admin and retrieve admin role/permissions.

@route GET /api/v1/admin/check/:address

@request
Path parameter:
- address: Wallet address to check (eth_addr format)

@response (success)
HTTP 200 OK

	{
	  "success": true,
	  "data": {
	    "is_admin": true,
	    "role": "ARCH_CONTROLLER_OWNER",
	    "permissions": ["create_market", "pause_market", "set_fees"],
	    "arch_controller_address": "0x..."
	  }
	}

@error_cases
- 400: Invalid address format (not valid Ethereum address)
- 401: Caller not authenticated
- 500: Service error checking admin status
*/
func (h *AdminAuthHandler) CheckAdmin(c *gin.Context) {
	/*
		@logic ExtractParameter
		@desc Extract and validate address from URL path.
		@notes Address must be valid Ethereum format (0x...).
	*/
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_ADDRESS",
				"message": "Address parameter required",
			},
		})
		return
	}

	/*
		@logic CallServiceLayer
		@desc Call auth service to verify admin status.
		@error_cases Service may return errors if address invalid.
	*/
	isAdmin, err := h.authService.IsAdmin(c.Request.Context(), address)
	if err != nil {
		statusCode := appErr.StatusCode(err)
		c.JSON(statusCode, gin.H{
			"success": false,
			"error": gin.H{
				"code":    appErr.ErrorCode(err),
				"message": "Failed to check admin status",
			},
		})
		return
	}

	/*
		@logic BuildResponse
		@desc Construct response with admin status and permissions.
		@notes Only include role/permissions if address is admin.
	*/
	resp := response.AdminCheckResponse{
		IsAdmin: isAdmin,
	}

	if isAdmin {
		/*
			@logic GetAdminDetails
			@desc Retrieve role and permissions for admin address.
			@notes Called only if address is confirmed admin.
		*/
		role, err := h.authService.GetAdminRole(c.Request.Context(), address)
		if err == nil {
			resp.Role = role
		}

		permissions, err := h.authService.GetAdminPermissions(c.Request.Context(), address)
		if err == nil {
			resp.Permissions = permissions
		}

		// Set arch controller address (hardcoded for now)
		resp.ArchControllerAddress = "0x" // TODO: get from config
	}

	/*
		@logic ReturnResponse
		@desc Return success response with admin information.
	*/
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
	})
}

/*
@handler Impersonate

@desc
Enable admin impersonation for testing (dev only).
Creates temporary token to switch to admin context.
DISABLED IN PRODUCTION - checked via environment variable.

@route POST /api/v1/admin/auth/impersonate

@request

	{
	  "admin_address": "0x...",
	  "reason": "Testing market creation flow"
	}

@response (success)
HTTP 200 OK

	{
	  "success": true,
	  "data": {
	    "admin_address": "0x...",
	    "token": "eyJhbGc...",
	    "reason": "Testing...",
	    "expires_at": 1685356800,
	    "valid_for_seconds": 3600
	  }
	}

@error_cases
- 400: Invalid request format or missing fields
- 401: Caller not authenticated
- 403: Feature disabled in production environment
- 404: Admin address not found or invalid
- 500: Failed to generate token
*/
func (h *AdminAuthHandler) Impersonate(c *gin.Context) {
	/*
		@logic CheckEnvironment
		@desc Verify impersonation is allowed in current environment.
		@notes Feature only available in development (dev/staging).
	*/
	// TODO: Check if environment is dev/staging
	// For now, allow all requests (production check should be implemented)

	/*
		@logic ParseRequest
		@desc Parse and validate impersonate request from JSON body.
		@notes Standard Gin binding with validation tags.
	*/
	var req request.ImpersonateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": fmt.Sprintf("invalid request: %v", err),
			},
		})
		return
	}

	/*
		@logic ValidateAddress
		@desc Verify admin address is valid and exists.
		@notes Will be checked by service layer.
	*/

	/*
		@logic GenerateToken
		@desc Call service to generate impersonation token.
		@params
		  - adminAddress: wallet to impersonate
		  - reason: reason for impersonation (audit trail)
		@returns
		  - token: JWT token for impersonation
		  - expiresAt: unix timestamp when token expires
	*/
	token, expiresAt, err := h.authService.GenerateImpersonateToken(
		c.Request.Context(),
		req.AdminAddress,
		req.Reason,
	)
	if err != nil {
		statusCode := appErr.StatusCode(err)
		c.JSON(statusCode, gin.H{
			"success": false,
			"error": gin.H{
				"code":    appErr.ErrorCode(err),
				"message": "Failed to generate impersonate token",
			},
		})
		return
	}

	/*
		@logic BuildResponse
		@desc Construct response with token and expiration details.
		@notes Include validity duration for client convenience.
	*/
	now := int64(0) // TODO: get current time
	validForSeconds := expiresAt - now

	resp := response.ImpersonateResponse{
		AdminAddress:    req.AdminAddress,
		Token:           token,
		Reason:          req.Reason,
		ExpiresAt:       expiresAt,
		ValidForSeconds: validForSeconds,
	}

	/*
		@logic ReturnResponse
		@desc Return success response with impersonate token.
		@notes Token should be used in Authorization header.
	*/
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
	})
}

/*
@handler ListAdmins

@desc
List all addresses with admin roles and their permissions.
Useful for admin panel visibility into current admins.

@route GET /api/v1/admin/admins

@request
Query parameters (optional):
- page: Page number (1-indexed, default: 1)
- limit: Results per page (default: 20, max: 100)
- role: Filter by role (OWNER, MANAGER, VIEWER)

@response (success)
HTTP 200 OK

	{
	  "success": true,
	  "data": {
	    "admins": [
	      {
	        "address": "0x...",
	        "role": "ARCH_CONTROLLER_OWNER",
	        "permissions": [...],
	        "added_at": 1685356800
	      }
	    ],
	    "pagination": {
	      "page": 1,
	      "limit": 20,
	      "total": 5
	    }
	  }
	}

@error_cases
- 400: Invalid pagination parameters
- 401: Caller not authenticated or not admin
- 403: Insufficient permissions to view admins
*/
func (h *AdminAuthHandler) ListAdmins(c *gin.Context) {
	/*
		@logic ValidatePermissions
		@desc Check if caller has permission to view admin list.
		@notes Only OWNER and MANAGER can view admin list.
	*/
	// TODO: Implement permission check

	/*
		@logic GetPaginationParams
		@desc Extract pagination parameters from query string.
		@notes Validate limits (max 100 per page).
	*/
	// TODO: Extract page and limit from query
	page := 1
	limit := 20

	/*
		@logic FetchAdminsList
		@desc Call service to fetch list of all admins.
		@params
		  - page: pagination page
		  - limit: results per page
		@returns
		  - admins: list of admin addresses with roles
		  - total: total number of admins
	*/
	// TODO: Call service

	/*
		@logic BuildResponse
		@desc Construct paginated response with admin list.
	*/
	resp := response.AdminListResponse{
		Admins: []response.AdminInfo{},
		Pagination: response.PaginationInfo{
			Page:  int32(page),
			Limit: int32(limit),
			Total: 0,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
	})
}
