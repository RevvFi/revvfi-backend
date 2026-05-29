package response

/*
@file auth.go

@desc
Response DTOs for authentication endpoints.
Returns nonces, tokens, and user information.
*/

/*
@struct NonceResponse

@desc
Response containing SIWE nonce for wallet signing.

@fields
- Nonce: random nonce string
- ExpiresAt: nonce expiration timestamp
- Message: SIWE message for signing
*/
type NonceResponse struct {
	Nonce     string `json:"nonce"`
	ExpiresAt int64  `json:"expires_at"`
	Message   string `json:"message"`
}

/*
@struct LoginResponse

@desc
Response containing JWT token after signature verification.

@fields
- Token: JWT access token
- ExpiresAt: token expiration unix timestamp
- RefreshToken: refresh token (if applicable)
- WalletAddress: authenticated wallet
*/
type LoginResponse struct {
	Token        string `json:"token"`
	ExpiresAt    int64  `json:"expires_at"`
	RefreshToken string `json:"refresh_token,omitempty"`
	WalletAddress string `json:"wallet_address"`
}

/*
@struct CurrentUserResponse

@desc
Response with authenticated user information.

@fields
- WalletAddress: user wallet
- Role: lender|borrower|both
- RegisteredAt: registration timestamp
- LastActivity: last activity timestamp
- IsVerified: KYC verification status
*/
type CurrentUserResponse struct {
	WalletAddress string `json:"wallet_address"`
	Role          string `json:"role"`
	RegisteredAt   int64  `json:"registered_at"`
	LastActivity   *int64 `json:"last_activity,omitempty"`
	IsVerified    bool   `json:"is_verified"`
}
