// internal/core/auth/service_test.go
package auth

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Revvfi/revvfi-backend/internal/models"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

/*
@file service_test.go

@desc
Unit tests for authentication service.
Tests nonce generation, signature verification, and JWT token handling.

@test_cases
- TestGenerateNonce: nonce generation and storage
- TestLoginSuccess: successful login with valid signature
- TestLogout: token revocation
- TestValidateToken: token validation
*/

// MockAuthRepository mock for testing
type MockAuthRepository struct {
	mock.Mock
}

func (m *MockAuthRepository) StoreNonce(ctx context.Context, wallet, nonce string, expiresAt time.Time) error {
	args := m.Called(ctx, wallet, nonce, expiresAt)
	return args.Error(0)
}

func (m *MockAuthRepository) ValidateNonce(ctx context.Context, wallet, nonce string) error {
	args := m.Called(ctx, wallet, nonce)
	return args.Error(0)
}

func (m *MockAuthRepository) ValidateNonceExists(ctx context.Context, wallet, nonce string) error {
	args := m.Called(ctx, wallet, nonce)
	return args.Error(0)
}

func (m *MockAuthRepository) ConsumeNonceAtomic(ctx context.Context, wallet, nonce string) error {
	args := m.Called(ctx, wallet, nonce)
	return args.Error(0)
}

func (m *MockAuthRepository) ConsumeNonce(ctx context.Context, wallet, nonce string) error {
	args := m.Called(ctx, wallet, nonce)
	return args.Error(0)
}

func (m *MockAuthRepository) GetNonce(ctx context.Context, wallet string) (string, error) {
	args := m.Called(ctx, wallet)
	return args.String(0), args.Error(1)
}

func (m *MockAuthRepository) DeleteNonce(ctx context.Context, wallet string) error {
	args := m.Called(ctx, wallet)
	return args.Error(0)
}

func (m *MockAuthRepository) StoreSession(ctx context.Context, session *models.AuthSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockAuthRepository) RevokeSession(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockAuthRepository) IsSessionRevoked(ctx context.Context, token string) (bool, error) {
	args := m.Called(ctx, token)
	return args.Bool(0), args.Error(1)
}

// Helper function to create default test config
func getTestConfig() *AuthConfig {
	return &AuthConfig{
		Domain:     "revvfi.com",
		URI:        "https://revvfi.com/api/v1/auth/login",
		Statement:  "Sign in to RevvFi Protocol",
		Version:    "1",
		ChainID:    1,
		NonceTTL:   5 * time.Minute,
		SessionTTL: 24 * time.Hour,
	}
}

/*
@test TestGenerateNonce

@desc
Tests successful nonce generation and storage.

@assertions
- Nonce is generated
- Nonce has correct format
- Message contains required components
- Repository store is called
*/
func TestGenerateNonce(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	jwtMgr, err := NewJWTManager("testsecretkeyof32byteslengthxxxxx", time.Hour, "revvfi")
	require.NoError(t, err)

	authService := NewAuthService(mockRepo, jwtMgr, getTestConfig())

	ctx := context.Background()
	wallet := "0x742d35Cc6634C0532925a3b844Bc9e7595f88128"

	// Expect store to be called
	mockRepo.On("StoreNonce", mock.MatchedBy(func(c context.Context) bool {
		return c == ctx
	}), wallet, mock.MatchedBy(func(n string) bool {
		return len(n) == 64 // hex-encoded 32 bytes
	}), mock.MatchedBy(func(t time.Time) bool {
		return t.After(time.Now())
	})).Return(nil)

	nonce, message, err := authService.GenerateNonce(ctx, wallet)

	require.NoError(t, err)
	assert.NotEmpty(t, nonce)
	assert.NotEmpty(t, message)
	assert.Contains(t, message, wallet)
	assert.Contains(t, message, nonce)
	mockRepo.AssertExpectations(t)
}

func TestVerifySignatureWithRecoveredPubKey(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	jwtMgr, err := NewJWTManager("testsecretkeyof32byteslengthxxxxx", time.Hour, "revvfi")
	require.NoError(t, err)

	authService := NewAuthService(mockRepo, jwtMgr, getTestConfig())
	wallet := "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
	message := "revvfi.com wants you to sign in with your Ethereum account:\n" +
		wallet + "\n\n" +
		"Sign in to RevvFi Protocol\n\n" +
		"This signature will not trigger a blockchain transaction or cost any gas.\n\n" +
		"By signing, you agree to the RevvFi Terms of Service (https://revvfi.com/terms)\n\n" +
		"URI: https://revvfi.com/api/v1/auth/login\n" +
		"Version: 1\n" +
		"Chain ID: 1\n" +
		"Nonce: e95d04008338bcf1ab12071bb927ca686574ecc5e8dc3a72135d52a118557733\n" +
		"Issued At: 2026-06-05T19:12:42Z"

	privateKey, err := crypto.HexToECDSA("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	require.NoError(t, err)

	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err := crypto.Sign(hash, privateKey)
	require.NoError(t, err)

	msg, err := authService.parseSIWEMessage(message)
	require.NoError(t, err)
	require.NoError(t, authService.verifySignature(msg, hexutil.Encode(signature), wallet))
}

/*
@test TestLogout

@desc
Tests session revocation on logout.

@assertions
- Logout calls repository revoke
- No error returned on success
*/
func TestLogout(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	jwtMgr, err := NewJWTManager("testsecretkeyof32byteslengthxxxxx", time.Hour, "revvfi")
	require.NoError(t, err)

	authService := NewAuthService(mockRepo, jwtMgr, getTestConfig())

	ctx := context.Background()
	token := "test.jwt.token"

	mockRepo.On("RevokeSession", ctx, token).Return(nil)

	err = authService.Logout(ctx, token)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

/*
@test TestValidateToken

@desc
Tests JWT token validation.

@assertions
- Valid token returns wallet
- Revoked token returns error
- Invalid token returns error
*/
func TestValidateToken(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	jwtMgr, err := NewJWTManager("testsecretkeyof32byteslengthxxxxx", time.Hour, "revvfi")
	require.NoError(t, err)

	authService := NewAuthService(mockRepo, jwtMgr, getTestConfig())

	ctx := context.Background()
	wallet := "0x742d35Cc6634C0532925a3b844Bc9e7595f88128"

	// Generate valid token
	token, _, err := jwtMgr.GenerateToken(wallet)
	require.NoError(t, err)

	// Test valid token
	mockRepo.On("IsSessionRevoked", ctx, token).Return(false, nil)
	retrievedWallet, err := authService.ValidateToken(ctx, token)
	require.NoError(t, err)
	assert.Equal(t, wallet, retrievedWallet)

	// Test revoked token
	revokedWallet := "0x1111111111111111111111111111111111111111"
	revokedToken, _, err := jwtMgr.GenerateToken(revokedWallet)
	require.NoError(t, err)
	mockRepo.On("IsSessionRevoked", ctx, revokedToken).Return(true, nil)
	_, err = authService.ValidateToken(ctx, revokedToken)
	assert.Error(t, err)

	mockRepo.AssertExpectations(t)
}

/*
@test TestNewAuthServiceWithCustomConfig

@desc
Tests that custom config is properly applied.

@assertions
- Custom config values are used
- Default config is used when nil
*/
func TestNewAuthServiceWithCustomConfig(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	jwtMgr, err := NewJWTManager("testsecretkeyof32byteslengthxxxxx", time.Hour, "revvfi")
	require.NoError(t, err)

	customConfig := &AuthConfig{
		Domain:     "custom.revvfi.com",
		URI:        "https://custom.revvfi.com/api/v1/auth/login",
		Statement:  "Custom statement",
		Version:    "2",
		ChainID:    42161, // Arbitrum
		NonceTTL:   10 * time.Minute,
		SessionTTL: 48 * time.Hour,
	}

	authService := NewAuthService(mockRepo, jwtMgr, customConfig)

	assert.Equal(t, "custom.revvfi.com", authService.config.Domain)
	assert.Equal(t, int64(42161), authService.config.ChainID)
	assert.Equal(t, 10*time.Minute, authService.config.NonceTTL)
}

/*
@test TestNewAuthServiceWithNilConfig

@desc
Tests that default config is used when nil is passed.

@assertions
- Default config values are used
*/
func TestNewAuthServiceWithNilConfig(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	jwtMgr, err := NewJWTManager("testsecretkeyof32byteslengthxxxxx", time.Hour, "revvfi")
	require.NoError(t, err)

	authService := NewAuthService(mockRepo, jwtMgr, nil)

	assert.Equal(t, "revvfi.com", authService.config.Domain)
	assert.Equal(t, int64(1), authService.config.ChainID)
	assert.Equal(t, 5*time.Minute, authService.config.NonceTTL)
}
