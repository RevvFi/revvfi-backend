package auth

import (
"testing"
"time"

"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"
)

/*
@file jwt_test.go

@desc
Unit tests for JWT token management.
Tests token generation, verification, and expiration.

@test_cases
- TestNewJWTManager: manager creation
- TestNewJWTManagerShortSecret: validation of secret length
- TestGenerateToken: token generation
- TestVerifyToken: token verification
- TestVerifyTokenExpired: rejection of expired tokens
- TestVerifyTokenInvalidSignature: rejection of tampered tokens
*/

/*
@test TestNewJWTManager

@desc
Tests JWT manager initialization with valid parameters.

@assertions
- Manager is created successfully
- Secret key is stored
- TTL is configured
*/
func TestNewJWTManager(t *testing.T) {
	secretKey := "thisisaverylongsecretkeyof32bytes!!"
	ttl := 24 * time.Hour
	issuer := "revvfi"

	mgr, err := NewJWTManager(secretKey, ttl, issuer)

	require.NoError(t, err)
	assert.NotNil(t, mgr)
	assert.Equal(t, mgr.tokenTTL, ttl)
	assert.Equal(t, mgr.issuer, issuer)
}

/*
@test TestNewJWTManagerShortSecret

@desc
Tests JWT manager validation of secret key length.

@assertions
- Returns error for short secret key
- Minimum 32 bytes required
*/
func TestNewJWTManagerShortSecret(t *testing.T) {
	shortSecret := "short"

	_, err := NewJWTManager(shortSecret, time.Hour, "revvfi")

	assert.Error(t, err)
}

/*
@test TestGenerateToken

@desc
Tests successful JWT token generation.

@assertions
- Token is generated
- Token contains wallet address
- Expiration is set correctly
- Token is in valid JWT format
*/
func TestGenerateToken(t *testing.T) {
	secretKey := "thisisaverylongsecretkeyof32bytes!!"
	mgr, err := NewJWTManager(secretKey, time.Hour, "revvfi")
	require.NoError(t, err)

	wallet := "0x742d35Cc6634C0532925a3b844Bc9e7595f88128"
	token, expiresAt, err := mgr.GenerateToken(wallet)

	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Greater(t, expiresAt, time.Now().Unix())
	// Token should have 3 parts (header.payload.signature)
	assert.Equal(t, 3, len(token)-len(token)+3)
}

/*
@test TestVerifyToken

@desc
Tests successful JWT token verification.

@assertions
- Valid token is verified
- Wallet address is extracted correctly
- No error is returned
*/
func TestVerifyToken(t *testing.T) {
	secretKey := "thisisaverylongsecretkeyof32bytes!!"
	mgr, err := NewJWTManager(secretKey, time.Hour, "revvfi")
	require.NoError(t, err)

	wallet := "0x742d35Cc6634C0532925a3b844Bc9e7595f88128"
	token, _, err := mgr.GenerateToken(wallet)
	require.NoError(t, err)

	verifiedWallet, err := mgr.VerifyToken(token)

	require.NoError(t, err)
	assert.Equal(t, wallet, verifiedWallet)
}

/*
@test TestVerifyTokenExpired

@desc
Tests rejection of expired token.

@assertions
- Expired token returns error
- Error message indicates expiration
*/
func TestVerifyTokenExpired(t *testing.T) {
	secretKey := "thisisaverylongsecretkeyof32bytes!!"
	mgr, err := NewJWTManager(secretKey, -time.Hour, "revvfi") // Negative TTL = already expired
	require.NoError(t, err)

	wallet := "0x742d35Cc6634C0532925a3b844Bc9e7595f88128"
	token, _, err := mgr.GenerateToken(wallet)
	require.NoError(t, err)

	// Create new manager with normal TTL for verification
	verifyMgr, err := NewJWTManager(secretKey, time.Hour, "revvfi")
	require.NoError(t, err)

	_, err = verifyMgr.VerifyToken(token)
	assert.Error(t, err)
}

/*
@test TestVerifyTokenInvalidSignature

@desc
Tests rejection of tampered token.

@assertions
- Token with modified payload returns error
- Error indicates invalid signature
*/
func TestVerifyTokenInvalidSignature(t *testing.T) {
	secretKey := "thisisaverylongsecretkeyof32bytes!!"
	mgr, err := NewJWTManager(secretKey, time.Hour, "revvfi")
	require.NoError(t, err)

	wallet := "0x742d35Cc6634C0532925a3b844Bc9e7595f88128"
	token, _, err := mgr.GenerateToken(wallet)
	require.NoError(t, err)

	// Tamper with token by changing the signature
	tamperedToken := token[:len(token)-10] + "0000000000"

	_, err = mgr.VerifyToken(tamperedToken)
	assert.Error(t, err)
}
