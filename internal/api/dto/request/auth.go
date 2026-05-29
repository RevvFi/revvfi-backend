package request

/*
@file auth.go

@desc
Request DTOs for authentication endpoints.
Handles SIWE nonce generation and JWT login flow.
*/

/*
@struct NonceRequest

@desc
Request body for generating SIWE authentication nonce.
No authentication required - nonce generation is public.

@fields
- WalletAddress: wallet address requesting nonce
*/
type NonceRequest struct {
	WalletAddress string `json:"wallet_address" binding:"required,eth_addr"`
}

/*
@struct LoginRequest

@desc
Request body for wallet signature verification and JWT issuance.
Contains SIWE message and signature for verification.

@fields
- WalletAddress: wallet address signing
- Message: SIWE message
- Signature: message signature
- Nonce: nonce from previous request
*/
type LoginRequest struct {
	WalletAddress string `json:"wallet_address" binding:"required,eth_addr"`
	Message       string `json:"message" binding:"required"`
	Signature     string `json:"signature" binding:"required"`
	Nonce         string `json:"nonce" binding:"required"`
}

/*
@struct LogoutRequest

@desc
Request body for invalidating active session.

@fields
- Token: JWT token to revoke (optional, uses auth header if not provided)
*/
type LogoutRequest struct {
	Token string `json:"token" binding:"omitempty"`
}
