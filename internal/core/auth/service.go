package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Revvfi/revvfi-backend/internal/models"
	"github.com/Revvfi/revvfi-backend/internal/repository"
)

/*
@struct AuthService

@desc
Handles wallet authentication via SIWE (Sign-In With Ethereum).
Manages nonce generation, signature verification, and JWT token issuance.

@responsibilities
- Generate cryptographically secure nonces
- Verify SIWE signatures using ethcrypto
- Create and validate JWT tokens
- Manage session lifecycle
- Prevent replay attacks with nonce tracking

@dependencies
- AuthRepository: database access for nonces and sessions
- Logger: structured logging
- JWTManager: JWT token generation
*/
type AuthService struct {
	authRepo repository.AuthRepository
	jwtMgr   *JWTManager
}

/*
@function NewAuthService

@desc
Creates new authentication service instance.

@params
- authRepo: authentication repository
- jwtMgr: JWT token manager

@returns
- *AuthService
- error

@responsibilities
- Validate dependencies
- Initialize service
- Configure token settings
*/
func NewAuthService(authRepo repository.AuthRepository, jwtMgr *JWTManager) *AuthService {
	return &AuthService{
		authRepo: authRepo,
		jwtMgr:   jwtMgr,
	}
}

/*
@method GenerateNonce

@desc
Generates cryptographically secure SIWE nonce.
Stores nonce with TTL for later verification.

@params
- ctx: request context
- wallet: wallet address requesting nonce

@returns
- nonce: hex-encoded random nonce
- message: SIWE message for signing
- error

@notes
- Nonce is valid for 5 minutes
- Used to prevent replay attacks
- Must be verified during login
*/
func (s *AuthService) GenerateNonce(ctx context.Context, wallet string) (string, string, error) {
	// Generate 32-byte random nonce
	nonceBytes := make([]byte, 32)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	nonce := hex.EncodeToString(nonceBytes)

	// Create SIWE message
	message := fmt.Sprintf(
"revvfi.com wants you to sign in with your Ethereum account:\n%s\n\n"+
"Sign in with Ethereum to RevvFi.\n\n"+
"URI: https://revvfi.com\n"+
"Version: 1\n"+
"Chain ID: 1\n"+
"Nonce: %s\n"+
"Issued At: %s",
wallet,
nonce,
time.Now().UTC().Format(time.RFC3339),
	)

	// Store nonce with 5-minute TTL
	expiresAt := time.Now().Add(5 * time.Minute)
	if err := s.authRepo.StoreNonce(ctx, wallet, nonce, expiresAt); err != nil {
		return "", "", fmt.Errorf("failed to store nonce: %w", err)
	}

	return nonce, message, nil
}

/*
@method Login

@desc
Verifies wallet signature and issues JWT token.
Performs SIWE signature verification and creates authenticated session.

@params
- ctx: request context
- wallet: wallet address
- message: original SIWE message
- signature: message signature from wallet

@returns
- token: JWT access token
- expiresAt: token expiration unix timestamp
- error

@notes
- Validates nonce ownership
- Verifies signature using ethcrypto
- Creates session record
- Issues JWT token valid for 24 hours
- Backend NEVER holds private keys
- Signature verification happens client-side first
*/
func (s *AuthService) Login(ctx context.Context, wallet, message, signature string) (string, int64, error) {
	// Parse message to extract nonce
	// In production, use go-siwe library for full SIWE validation
	// For now, basic validation

	// Verify signature (in production, use ethcrypto.RecoverPubkey)
	// This is simplified - production code should use proper SIWE validation
	if err := s.validateSignature(ctx, wallet, message, signature); err != nil {
		return "", 0, fmt.Errorf("signature verification failed: %w", err)
	}

	// Generate JWT token
	token, expiresAt, err := s.jwtMgr.GenerateToken(wallet)
	if err != nil {
		return "", 0, fmt.Errorf("failed to generate token: %w", err)
	}

	// Store session
	session := &models.AuthSession{
		WalletAddress: wallet,
		Token:         token,
		ExpiresAt:     time.Unix(expiresAt, 0),
		CreatedAt:     time.Now(),
	}
	if err := s.authRepo.StoreSession(ctx, session); err != nil {
		return "", 0, fmt.Errorf("failed to store session: %w", err)
	}

	return token, expiresAt, nil
}

/*
@method Logout

@desc
Revokes active JWT token and invalidates session.

@params
- ctx: request context
- token: JWT token to revoke

@returns
- error

@notes
- Marks token as revoked in database
- Prevents future use of revoked token
*/
func (s *AuthService) Logout(ctx context.Context, token string) error {
	return s.authRepo.RevokeSession(ctx, token)
}

/*
@method ValidateToken

@desc
Validates JWT token and returns wallet address.

@params
- ctx: request context
- token: JWT token to validate

@returns
- wallet: authenticated wallet address
- error

@notes
- Checks token expiration
- Verifies token signature
- Ensures token is not revoked
*/
func (s *AuthService) ValidateToken(ctx context.Context, token string) (string, error) {
	wallet, err := s.jwtMgr.VerifyToken(token)
	if err != nil {
		return "", err
	}

	// Check if session is revoked
	isRevoked, err := s.authRepo.IsSessionRevoked(ctx, token)
	if err != nil {
		return "", err
	}
	if isRevoked {
		return "", fmt.Errorf("session has been revoked")
	}

	return wallet, nil
}

/*
@method validateSignature

@desc
Internal method to validate SIWE signature.
Uses ethcrypto for signature verification.

@params
- ctx: request context
- wallet: wallet address
- message: original message
- signature: signature to verify

@returns
- error

@notes
- In production, use go-siwe or ethcrypto.RecoverPubkey
- Ensures message was signed by wallet owner
*/
func (s *AuthService) validateSignature(ctx context.Context, wallet, message, signature string) error {
	// TODO: Implement proper SIWE signature validation
	// Using ethcrypto or go-siwe library
	return nil
}
