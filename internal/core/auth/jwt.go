package auth

import (
"fmt"
"time"

"github.com/golang-jwt/jwt/v5"
)

/*
@struct JWTManager

@desc
Manages JWT token generation and verification.
Handles token lifecycle with configurable TTL.

@fields
- secretKey: JWT signing secret
- tokenTTL: token time-to-live duration
- issuer: token issuer identifier
*/
type JWTManager struct {
	secretKey []byte
	tokenTTL  time.Duration
	issuer    string
}

/*
@struct Claims

@desc
JWT token claims payload.

@fields
- Wallet: authenticated wallet address
- RegisteredClaims: standard JWT claims
*/
type Claims struct {
	Wallet string `json:"wallet"`
	jwt.RegisteredClaims
}

/*
@function NewJWTManager

@desc
Creates new JWT manager instance.

@params
- secretKey: JWT signing secret (minimum 32 bytes)
- tokenTTL: token time-to-live
- issuer: token issuer

@returns
- *JWTManager
- error if secret key too short
*/
func NewJWTManager(secretKey string, tokenTTL time.Duration, issuer string) (*JWTManager, error) {
	if len(secretKey) < 32 {
		return nil, fmt.Errorf("secret key must be at least 32 bytes")
	}

	return &JWTManager{
		secretKey: []byte(secretKey),
		tokenTTL:  tokenTTL,
		issuer:    issuer,
	}, nil
}

/*
@method GenerateToken

@desc
Generates new JWT token for wallet address.

@params
- wallet: wallet address for token

@returns
- token: signed JWT token string
- expiresAt: token expiration unix timestamp
- error

@notes
- Token includes wallet address in claims
- Expiration is current time + tokenTTL
- Signed with HS256 algorithm
*/
func (m *JWTManager) GenerateToken(wallet string) (string, int64, error) {
	expiresAt := time.Now().Add(m.tokenTTL)

	claims := &Claims{
		Wallet: wallet,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    m.issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.secretKey)
	if err != nil {
		return "", 0, fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, expiresAt.Unix(), nil
}

/*
@method VerifyToken

@desc
Verifies JWT token and extracts claims.

@params
- token: JWT token string to verify

@returns
- wallet: wallet address from token
- error if verification fails

@notes
- Validates signature
- Checks expiration
- Extracts wallet claim
- Returns error if token invalid or expired
*/
func (m *JWTManager) VerifyToken(token string) (string, error) {
	claims := &Claims{}

	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secretKey, nil
	})

	if err != nil {
		return "", fmt.Errorf("token parsing failed: %w", err)
	}

	if !parsedToken.Valid {
		return "", fmt.Errorf("token is invalid")
	}

	if claims.Wallet == "" {
		return "", fmt.Errorf("wallet claim missing")
	}

	return claims.Wallet, nil
}
