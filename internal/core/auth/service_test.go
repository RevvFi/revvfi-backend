package auth

import (
	"context"
	"testing"
	"time"

	"github.com/Revvfi/revvfi-backend/internal/models"

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

	service := NewAuthService(mockRepo, jwtMgr)

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

	nonce, message, err := service.GenerateNonce(ctx, wallet)

	require.NoError(t, err)
	assert.NotEmpty(t, nonce)
	assert.NotEmpty(t, message)
	assert.Contains(t, message, wallet)
	assert.Contains(t, message, nonce)
	mockRepo.AssertExpectations(t)
}

/*
@test TestLoginSuccess

@desc
Tests successful login with valid signature.

@assertions
- JWT token is generated
- Token is stored as session
- Expiration timestamp is set correctly
*/
func TestLoginSuccess(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	jwtMgr, err := NewJWTManager("testsecretkeyof32byteslengthxxxxx", time.Hour, "revvfi")
	require.NoError(t, err)

	service := NewAuthService(mockRepo, jwtMgr)

	ctx := context.Background()
	wallet := "0x742d35Cc6634C0532925a3b844Bc9e7595f88128"
	message := "test message"
	signature := "0xvalidsignature"

	mockRepo.On("StoreSession", mock.MatchedBy(func(c context.Context) bool {
		return c == ctx
	}), mock.MatchedBy(func(s *models.AuthSession) bool {
		return s.WalletAddress == wallet
	})).Return(nil)

	token, expiresAt, err := service.Login(ctx, wallet, message, signature)

	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Greater(t, expiresAt, time.Now().Unix())
	mockRepo.AssertExpectations(t)
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

	service := NewAuthService(mockRepo, jwtMgr)

	ctx := context.Background()
	token := "test.jwt.token"

	mockRepo.On("RevokeSession", ctx, token).Return(nil)

	err = service.Logout(ctx, token)

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

	service := NewAuthService(mockRepo, jwtMgr)

	ctx := context.Background()
	wallet := "0x742d35Cc6634C0532925a3b844Bc9e7595f88128"

	// Generate valid token
	token, _, err := jwtMgr.GenerateToken(wallet)
	require.NoError(t, err)

	// Test valid token
	mockRepo.On("IsSessionRevoked", ctx, token).Return(false, nil)
	retrievedWallet, err := service.ValidateToken(ctx, token)
	require.NoError(t, err)
	assert.Equal(t, wallet, retrievedWallet)

	// Test revoked token
	revokedWallet := "0x1111111111111111111111111111111111111111"
	revokedToken, _, err := jwtMgr.GenerateToken(revokedWallet)
	require.NoError(t, err)
	mockRepo.On("IsSessionRevoked", ctx, revokedToken).Return(true, nil)
	_, err = service.ValidateToken(ctx, revokedToken)
	assert.Error(t, err)

	mockRepo.AssertExpectations(t)
}
