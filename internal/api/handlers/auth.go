package handlers

import (
	"net/http"

	"github.com/Revvfi/revvfi-backend/internal/api/dto/request"
	"github.com/Revvfi/revvfi-backend/internal/api/dto/response"
	"github.com/Revvfi/revvfi-backend/internal/core/auth"
	"github.com/gin-gonic/gin"
)

/*
@struct AuthHandler

@desc
HTTP handlers for authentication endpoints.
Handles SIWE nonce generation and JWT login/logout.

@responsibilities
- Validate request payloads
- Call auth service layer
- Return standardized responses
- Handle error cases with proper HTTP status

@dependencies
- AuthService: authentication business logic
- Logger: structured logging
*/
type AuthHandler struct {
	authService *auth.AuthService
}

/*
@function NewAuthHandler

@desc
Creates new auth handler.

@params
- authService: authentication service

@returns
- *AuthHandler
*/
func NewAuthHandler(authService *auth.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

/*
@method GetNonce

@desc
HTTP handler for POST /api/auth/nonce.
Generates SIWE nonce for wallet signing.

@params
- c: Gin context

@request
- wallet_address: ethereum wallet address

@response
- nonce: random nonce string
- message: SIWE message for signing
- expires_at: nonce expiration timestamp

@errors
- 400: invalid wallet address
- 500: nonce generation failure
*/
func (h *AuthHandler) GetNonce(c *gin.Context) {
	var req request.NonceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
    
	nonce, message, err := h.authService.GenerateNonce(c.Request.Context(), req.WalletAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.NonceResponse{
		Nonce:   nonce,
		Message: message,
	})
}

/*
@method Login

@desc
HTTP handler for POST /api/auth/login.
Verifies wallet signature and issues JWT token.

@params
- c: Gin context

@request
- wallet_address: wallet address
- message: original SIWE message
- signature: wallet signature

@response
- token: JWT access token
- expires_at: token expiration
- wallet_address: authenticated wallet

@errors
- 400: invalid request
- 401: signature verification failed
- 500: internal error
*/
func (h *AuthHandler) Login(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, expiresAt, err := h.authService.Login(c.Request.Context(), req.WalletAddress, req.Message, req.Signature)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication failed"})
		return
	}

	c.JSON(http.StatusOK, response.LoginResponse{
		Token:         token,
		ExpiresAt:     expiresAt,
		WalletAddress: req.WalletAddress,
	})
}

/*
@method Logout

@desc
HTTP handler for POST /api/auth/logout.
Revokes active JWT token.

@params
- c: Gin context (expects Authorization header)

@errors
- 401: no token provided
- 500: revocation failure
*/
func (h *AuthHandler) Logout(c *gin.Context) {
	token := c.GetString("token") // Set by auth middleware
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	if err := h.authService.Logout(c.Request.Context(), token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

/*
@method GetMe

@desc
HTTP handler for GET /api/auth/me.
Returns authenticated user information.

@params
- c: Gin context (requires valid JWT)

@response
- wallet_address: user wallet
- role: lender|borrower|both
- registered_at: registration timestamp
- is_verified: KYC status

@errors
- 401: not authenticated
- 500: internal error
*/
func (h *AuthHandler) GetMe(c *gin.Context) {
	wallet := c.GetString("wallet") // Set by auth middleware
	if wallet == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	// TODO: Fetch user profile from database
	c.JSON(http.StatusOK, response.CurrentUserResponse{
		WalletAddress: wallet,
		Role:          "lender",
		RegisteredAt:  0, // TODO: Set from DB
		IsVerified:    false,
	})
}
