package admin

import (
	"fmt"
	"strings"

	"github.com/Revvfi/revvfi-backend/internal/config"
	"github.com/Revvfi/revvfi-backend/internal/core/auth"
	appErr "github.com/Revvfi/revvfi-backend/internal/pkg/errors"
)

/*
@file auth_service.go

@desc
Service for admin authentication and authorization.
Verifies admin status against the ArchController address from config.

@responsibilities
- Determine whether an Ethereum address is an admin
- Return admin role and permissions based on address match
- Generate short-lived impersonation tokens via the JWT manager
- Validate impersonation tokens
*/

/*
@struct AdminAuthService

@desc
Implements AdminAuthServiceInterface for admin role verification.

@fields
- cfg: blockchain configuration (ArchControllerAddress is the owner)
- jwtMgr: JWT manager for generating impersonation tokens
*/
type AdminAuthService struct {
	cfg    config.BlockchainConfig
	jwtMgr *auth.JWTManager
}

/*
@function NewAdminAuthService

@desc
Creates a new AdminAuthService.

@params
- cfg: blockchain configuration
- jwtMgr: JWT manager for impersonation token generation

@returns
- *AdminAuthService
*/
func NewAdminAuthService(cfg config.BlockchainConfig, jwtMgr *auth.JWTManager) *AdminAuthService {
	return &AdminAuthService{cfg: cfg, jwtMgr: jwtMgr}
}

/*
@method IsAdmin

@desc
Returns true if the address matches the ArchController owner address.
Admin status is determined by the configured ArchControllerAddress.

@params
- ctx: handler context (unused, included for interface compatibility)
- address: wallet address to check

@returns
- bool: true if address is admin
- error
*/
func (s *AdminAuthService) IsAdmin(_ interface{}, address string) (bool, error) {
	if address == "" {
		return false, appErr.ErrInvalidAddress
	}

	// Normalize the input address
	normalizedAddress := strings.ToLower(strings.TrimSpace(address))

	// Check against all admin wallets
	for _, adminWallet := range s.cfg.AdminWallets {
		if strings.EqualFold(normalizedAddress, strings.TrimSpace(adminWallet)) {
			return true, nil
		}
	}

	return false, nil
}

/*
@method GetAdminRole

@desc
Returns the role name for an admin address.
Single-owner model: only the ArchControllerAddress has OWNER role.

@params
- ctx: handler context
- address: wallet address

@returns
- string: role name
- error: ErrUnauthorized if not admin
*/
func (s *AdminAuthService) GetAdminRole(_ interface{}, address string) (string, error) {
	isAdmin, err := s.IsAdmin(nil, address)
	if err != nil {
		return "", err
	}
	if !isAdmin {
		return "", appErr.ErrUnauthorized
	}
	return "ARCH_CONTROLLER_OWNER", nil
}

/*
@method GetAdminPermissions

@desc
Returns the list of permissions for an admin address.

@params
- ctx: handler context
- address: wallet address

@returns
- []string: granted permissions
- error
*/
func (s *AdminAuthService) GetAdminPermissions(_ interface{}, address string) ([]string, error) {
	isAdmin, err := s.IsAdmin(nil, address)
	if err != nil {
		return nil, err
	}
	if !isAdmin {
		return nil, appErr.ErrUnauthorized
	}
	return []string{
		"create_market", "pause_market", "set_fees", "manage_borrowers",
		"manage_liquidator", "manage_reputation", "emergency_controls",
		"update_system_config", "view_audit_logs",
	}, nil
}

/*
@method GenerateImpersonateToken

@desc
Generates a short-lived JWT impersonation token for the given admin address.

@params
- ctx: handler context
- adminAddress: wallet address to impersonate
- reason: justification for impersonation (included in token claims)

@returns
- string: JWT token
- int64: expiration Unix timestamp
- error
*/
func (s *AdminAuthService) GenerateImpersonateToken(_ interface{}, adminAddress, reason string) (string, int64, error) {
	if adminAddress == "" {
		return "", 0, appErr.ErrInvalidAddress
	}
	_ = reason

	token, expiresAt, err := s.jwtMgr.GenerateToken(adminAddress)
	if err != nil {
		return "", 0, fmt.Errorf("generate impersonate token: %w", err)
	}
	return token, expiresAt, nil
}

/*
@method ValidateImpersonateToken

@desc
Validates a previously issued impersonation token and returns the admin address.

@params
- token: JWT impersonation token

@returns
- string: admin wallet address embedded in the token
- error: ErrInvalidToken if the token is invalid or expired
*/
func (s *AdminAuthService) ValidateImpersonateToken(token string) (string, error) {
	wallet, err := s.jwtMgr.VerifyToken(token)
	if err != nil {
		return "", appErr.ErrInvalidToken
	}
	return wallet, nil
}
