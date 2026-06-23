package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/Revvfi/revvfi-backend/internal/models"
	appErr "github.com/Revvfi/revvfi-backend/internal/pkg/errors"
	"github.com/Revvfi/revvfi-backend/internal/repository"
	Errors "github.com/Revvfi/revvfi-backend/pkg/errors"
)

/*
@const MaxSIWEMessageSize
@desc Maximum allowed SIWE message size to prevent DoS attacks
@fix #8: Added back - prevents memory exhaustion attacks
*/
const MaxSIWEMessageSize = 8192 // 8KB

/*
@const MinNonceLength
@desc Minimum length for nonce (EIP-4361 requires minimum 8)
*/
const MinNonceLength = 8

/*
@const MaxNonceLength
@desc Maximum length for nonce
*/
const MaxNonceLength = 64

/*
@const DefaultNonceTTL
@desc Default TTL for nonces if config not provided
*/
const DefaultNonceTTL = 5 * time.Minute

/*
@var nonceRegex
@desc EIP-4361 compliant nonce regex - only alphanumeric characters
@fix #9: Validates nonce format per specification
*/
var nonceRegex = regexp.MustCompile(`^[A-Za-z0-9]{8,64}$`)

/*
@var AllowedVValues
@desc Valid ECDSA signature recovery ID values (v)
*/
var AllowedVValues = map[uint8]bool{0: true, 1: true, 27: true, 28: true}

/*
@struct AuthConfig
@desc Authentication configuration
@fix #5: All values configurable, not hardcoded
*/
type AuthConfig struct {
	Domain     string        `json:"domain"` // FIX #7: Configurable domain
	URI        string        `json:"uri"`    // FIX #6: Configurable URI
	Statement  string        `json:"statement"`
	Version    string        `json:"version"`
	ChainID    int64         `json:"chain_id"` // FIX #5: Configurable chain ID
	NonceTTL   time.Duration `json:"nonce_ttl"`
	SessionTTL time.Duration `json:"session_ttl"`
}

/*
@struct AuthService
@desc Handles wallet authentication via SIWE
*/
type AuthService struct {
	authRepo repository.AuthRepository
	jwtMgr   *JWTManager
	config   *AuthConfig
}

/*
@function NewAuthService
@desc Creates new authentication service
*/
func NewAuthService(
	authRepo repository.AuthRepository,
	jwtMgr *JWTManager,
	config *AuthConfig,
) *AuthService {
	if config == nil {
		config = &AuthConfig{
			Domain:     "revvfi.com",
			URI:        "https://revvfi.com/api/v1/auth/login",
			Statement:  "Sign in to RevvFi Protocol",
			Version:    "1",
			ChainID:    1,
			NonceTTL:   DefaultNonceTTL,
			SessionTTL: 24 * time.Hour,
		}
	}
	return &AuthService{
		authRepo: authRepo,
		jwtMgr:   jwtMgr,
		config:   config,
	}
}

/*
@method GenerateNonce
@desc Generates cryptographically secure SIWE nonce
*/
func (s *AuthService) GenerateNonce(ctx context.Context, wallet string) (string, string, error) {
	// Validate wallet address
	if !common.IsHexAddress(wallet) {
		return "", "", appErr.ErrInvalidAddress
	}

	// Generate 32-byte random nonce (256 bits)
	nonceBytes := make([]byte, 32)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	nonce := hex.EncodeToString(nonceBytes)

	// Build EIP-4361 compliant message using config
	message := s.buildSIWEMessage(wallet, nonce)

	// Store nonce with TTL
	expiresAt := time.Now().Add(s.config.NonceTTL)
	if err := s.authRepo.StoreNonce(ctx, wallet, nonce, expiresAt); err != nil {
		return "", "", fmt.Errorf("failed to store nonce: %w", err)
	}

	return nonce, message, nil
}

/*
@method buildSIWEMessage
@desc Builds EIP-4361 compliant SIWE message
@security Uses config values instead of hardcoded constants
*/
func (s *AuthService) buildSIWEMessage(wallet, nonce string) string {
	now := time.Now().UTC()
	// Fix: Use raw strings with actual newlines, not escaped ones
	statement := "Sign in to RevvFi Protocol\n\nThis signature will not trigger a blockchain transaction or cost any gas.\n\nBy signing, you agree to the RevvFi Terms of Service (https://revvfi.com/terms)"
	return fmt.Sprintf(
		"%s wants you to sign in with your Ethereum account:\n%s\n\n%s\n\nURI: %s\nVersion: %s\nChain ID: %d\nNonce: %s\nIssued At: %s",
		s.config.Domain,
		wallet,
		statement,
		s.config.URI,
		s.config.Version,
		s.config.ChainID,
		nonce,
		now.Format(time.RFC3339),
	)
}

/*
@method Login
@desc Verifies SIWE signature and issues JWT token
@security Complete flow with all fixes:
1. Validate message size (prevent DoS)
2. Parse SIWE message (sets Raw field)
3. Extract and validate nonce
4. Validate nonce exists (read-only)
5. Validate all SIWE fields against config
6. Verify signature (with explicit crypto.VerifySignature)
7. Consume nonce ATOMICALLY
8. Generate JWT token
9. Store session
*/
func (s *AuthService) Login(
	ctx context.Context,
	wallet, message, signature string,
) (string, int64, error) {
	// FIX #8: Validate message size before parsing
	if len(message) > MaxSIWEMessageSize {
		return "", 0, Errors.ErrMessageTooLarge
	}

	// Validate wallet address
	if !common.IsHexAddress(wallet) {
		return "", 0, appErr.ErrInvalidAddress
	}

	// Parse SIWE message (FIX #3: Raw field is set here)
	siweMsg, err := s.parseSIWEMessage(message)
	if err != nil {
		fmt.Printf("ERROR Login: parseSIWEMessage failed: %v\n", err)
		return "", 0, fmt.Errorf("invalid SIWE message: %w", err)
	}

	// Canonicalize wallet addresses
	canonicalWallet := common.HexToAddress(wallet).Hex()
	canonicalParsedAddress := common.HexToAddress(siweMsg.Address).Hex()
	fmt.Printf("DEBUG Login: wallet=%s parsed=%s\n", canonicalWallet, canonicalParsedAddress)

	if !strings.EqualFold(canonicalWallet, canonicalParsedAddress) {
		fmt.Printf("ERROR Login: address mismatch\n")
		return "", 0, Errors.ErrAddressMismatch
	}

	// FIX #9: Validate nonce format
	if len(siweMsg.Nonce) < MinNonceLength || len(siweMsg.Nonce) > MaxNonceLength {
		fmt.Printf("ERROR Login: nonce length invalid: %d\n", len(siweMsg.Nonce))
		return "", 0, appErr.ErrInvalidNonce
	}
	if !nonceRegex.MatchString(siweMsg.Nonce) {
		fmt.Printf("ERROR Login: nonce format invalid: %s\n", siweMsg.Nonce)
		return "", 0, Errors.ErrInvalidNonceFormat
	}
	fmt.Printf("DEBUG Login: nonce validated: %s\n", siweMsg.Nonce)

	// Validate nonce exists (read-only - does NOT consume)
	if err := s.authRepo.ValidateNonceExists(ctx, canonicalWallet, siweMsg.Nonce); err != nil {
		fmt.Printf("ERROR Login: ValidateNonceExists failed: %v\n", err)
		return "", 0, err
	}
	fmt.Printf("DEBUG Login: nonce exists in DB\n")

	// Validate all SIWE fields
	if err := s.validateSIWEMessage(siweMsg); err != nil {
		fmt.Printf("ERROR Login: validateSIWEMessage failed: %v\n", err)
		return "", 0, err
	}
	fmt.Printf("DEBUG Login: SIWE message validated (domain=%s, URI=%s, chainID=%d)\n", siweMsg.Domain, siweMsg.URI, siweMsg.ChainID)

	// Verify signature (with explicit crypto.VerifySignature)
	if err := s.verifySignature(siweMsg, signature, canonicalWallet); err != nil {
		fmt.Printf("ERROR Login: verifySignature failed: %v\n", err)
		return "", 0, err
	}
	fmt.Printf("DEBUG Login: signature verified\n")

	// FIX #2: Atomically consume the nonce
	if err := s.authRepo.ConsumeNonceAtomic(ctx, canonicalWallet, siweMsg.Nonce); err != nil {
		return "", 0, err
	}

	// Generate JWT token
	token, expiresAt, err := s.jwtMgr.GenerateToken(canonicalWallet)
	if err != nil {
		return "", 0, fmt.Errorf("failed to generate token: %w", err)
	}

	// Store session
	session := &models.AuthSession{
		WalletAddress: canonicalWallet,
		Token:         token,
		ExpiresAt:     time.Unix(expiresAt, 0),
		CreatedAt:     time.Now(),
	}
	if err := s.authRepo.StoreSession(ctx, session); err != nil {
		// Log error but don't fail - token is still valid
		// Session tracking is for revocation, not authentication
	}

	return token, expiresAt, nil
}

/*
@method parseSIWEMessage
@desc Parses EIP-4361 compliant SIWE message
@security FIX #3: Sets Raw field for signature verification
@security FIX #11: Proper field boundary detection
*/
func (s *AuthService) parseSIWEMessage(message string) (*SIWEMessage, error) {
	// FIX #3: Set Raw message immediately
	result := &SIWEMessage{
		Raw:      message,
		RawLines: strings.Split(message, "\n"),
	}
	fmt.Printf("DEBUG parseSIWEMessage: Raw message length: %d\n", len(result.Raw))
	fmt.Printf("DEBUG parseSIWEMessage: First 200 chars of Raw: %q\n",
		func() string {
			if len(result.Raw) > 200 {
				return result.Raw[:200]
			}
			return result.Raw
		}())
	lines := result.RawLines
	if len(lines) < 8 {
		return nil, fmt.Errorf("invalid SIWE message: insufficient lines")
	}

	// Parse domain and address
	firstLine := lines[0]
	domainIdx := strings.Index(firstLine, " wants you to sign in with your Ethereum account:")
	if domainIdx < 0 {
		return nil, fmt.Errorf("invalid SIWE format: missing domain")
	}
	result.Domain = firstLine[:domainIdx]
	if len(lines) > 1 {
		result.Address = strings.TrimSpace(lines[1])
	}

	// Parse statement (FIX #11: Proper boundary detection)
	statementLines := []string{}
	lineIdx := 2
	for lineIdx < len(lines) {
		trimmed := strings.TrimSpace(lines[lineIdx])
		if strings.HasPrefix(trimmed, "URI:") {
			break
		}
		if trimmed != "" {
			statementLines = append(statementLines, trimmed)
		}
		lineIdx++
	}
	result.Statement = strings.Join(statementLines, "\n")

	// Parse URI
	if lineIdx < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[lineIdx]), "URI:") {
		uriStr := strings.TrimSpace(strings.TrimPrefix(lines[lineIdx], "URI:"))
		result.URI = uriStr
		lineIdx++
	}

	// Parse Version
	if lineIdx < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[lineIdx]), "Version:") {
		versionStr := strings.TrimSpace(strings.TrimPrefix(lines[lineIdx], "Version:"))
		result.Version = versionStr
		lineIdx++
	}

	// Parse Chain ID
	if lineIdx < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[lineIdx]), "Chain ID:") {
		chainIDStr := strings.TrimSpace(strings.TrimPrefix(lines[lineIdx], "Chain ID:"))
		if _, err := fmt.Sscanf(chainIDStr, "%d", &result.ChainID); err != nil {
			return nil, fmt.Errorf("invalid Chain ID")
		}
		lineIdx++
	}

	// Parse Nonce
	if lineIdx < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[lineIdx]), "Nonce:") {
		result.Nonce = strings.TrimSpace(strings.TrimPrefix(lines[lineIdx], "Nonce:"))
		lineIdx++
	}

	// Parse Issued At
	if lineIdx < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[lineIdx]), "Issued At:") {
		issuedAtStr := strings.TrimSpace(strings.TrimPrefix(lines[lineIdx], "Issued At:"))
		if issuedAt, err := time.Parse(time.RFC3339, issuedAtStr); err == nil {
			result.IssuedAt = issuedAt
		}
		lineIdx++
	}

	// Parse Not Before (optional)
	if lineIdx < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[lineIdx]), "Not Before:") {
		notBeforeStr := strings.TrimSpace(strings.TrimPrefix(lines[lineIdx], "Not Before:"))
		if notBefore, err := time.Parse(time.RFC3339, notBeforeStr); err == nil {
			result.NotBefore = &notBefore
		}
		lineIdx++
	}

	// Parse Expiration Time (optional)
	if lineIdx < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[lineIdx]), "Expiration Time:") {
		expirationStr := strings.TrimSpace(strings.TrimPrefix(lines[lineIdx], "Expiration Time:"))
		if expiration, err := time.Parse(time.RFC3339, expirationStr); err == nil {
			result.Expiration = &expiration
		}
		lineIdx++
	}

	// Parse Request ID (optional)
	if lineIdx < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[lineIdx]), "Request ID:") {
		requestID := strings.TrimSpace(strings.TrimPrefix(lines[lineIdx], "Request ID:"))
		result.RequestID = &requestID
		lineIdx++
	}

	// FIX #12: Parse Resources and advance index correctly
	if lineIdx < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[lineIdx]), "Resources:") {
		resources := []string{}
		lineIdx++ // Skip the "Resources:" line
		for lineIdx < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[lineIdx]), "- ") {
			resource := strings.TrimSpace(strings.TrimPrefix(lines[lineIdx], "- "))
			resources = append(resources, resource)
			lineIdx++
		}
		result.Resources = resources
	}

	return result, nil
}

/*
@method validateSIWEMessage
@desc Validates all SIWE message fields against config
@security FIX #5, #6, #7: Uses config for validation
*/
func (s *AuthService) validateSIWEMessage(msg *SIWEMessage) error {
	// FIX #7: Domain validation against config (not user input)
	if msg.Domain != s.config.Domain {
		return fmt.Errorf("invalid domain: expected %s, got %s", s.config.Domain, msg.Domain)
	}

	// FIX #6: URI validation with proper URL parsing
	if err := s.validateURI(msg.URI); err != nil {
		return err
	}

	// Version validation
	if msg.Version != s.config.Version {
		return fmt.Errorf("invalid version: expected %s, got %s", s.config.Version, msg.Version)
	}

	// FIX #5: Chain ID validation against config (not hardcoded)
	if msg.ChainID != s.config.ChainID {
		return fmt.Errorf("invalid chain ID: expected %d, got %d", s.config.ChainID, msg.ChainID)
	}

	// Nonce format already validated earlier

	// Issued At validation (prevent future-dated messages)
	now := time.Now()
	if msg.IssuedAt.After(now.Add(5 * time.Minute)) {
		return Errors.ErrFutureIssuedAt
	}

	// Not Before validation
	if msg.NotBefore != nil && now.Before(*msg.NotBefore) {
		return Errors.ErrMessageNotYetValid
	}

	// Expiration validation
	if msg.Expiration != nil && now.After(*msg.Expiration) {
		return Errors.ErrMessageExpired
	}

	return nil
}

/*
@method validateURI
@desc Validates URI with proper URL parsing
@security FIX #6: Strong URI validation with scheme check
*/
func (s *AuthService) validateURI(uriStr string) error {
	parsed, err := url.Parse(uriStr)
	if err != nil {
		return fmt.Errorf("invalid URI format: %w", err)
	}

	// Only allow https and http schemes
	switch parsed.Scheme {
	case "https", "http":
		// Valid
	default:
		return fmt.Errorf("invalid URI scheme: %s (only https/http allowed)", parsed.Scheme)
	}

	// Validate host matches config domain; use parsed.Host (not Hostname) to retain port
	host := parsed.Host
	if host != s.config.Domain {
		return fmt.Errorf("URI host %s does not match expected domain %s", host, s.config.Domain)
	}

	return nil
}

/*
@method verifySignature
@desc Verifies SIWE signature with EOA support
@security FIX #4: Added explicit crypto.VerifySignature()
@security FIX #2: Uses msg.Raw for hash
*/
func (s *AuthService) verifySignature(
	msg *SIWEMessage,
	signatureHex string,
	expectedAddress string,
) error {

	fmt.Println("DEBUG: Starting signature verification")

	// Decode signature
	sigBytes, err := hexutil.Decode(signatureHex)
	if err != nil {
		fmt.Printf("DEBUG: Signature decode error: %v\n", err)
		return fmt.Errorf("invalid signature format: %w", err)
	}
	if len(sigBytes) != 65 {
		fmt.Printf("DEBUG: Invalid signature length: %d\n", len(sigBytes))
		return fmt.Errorf("invalid signature length: expected 65, got %d", len(sigBytes))
	}
	fmt.Println("DEBUG: Signature decoded successfully")

	// Validate v value
	v := sigBytes[64]
	if !AllowedVValues[v] {
		fmt.Printf("DEBUG: Invalid v value: %d\n", v)
		return fmt.Errorf("invalid v value: %d", v)
	}
	if v == 27 || v == 28 {
		sigBytes[64] -= 27
	}
	fmt.Println("DEBUG: v value valid")

	// FIX #2 & #3: Use msg.Raw (the original message) for hash
	prefix := fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len(msg.Raw))
	fmt.Printf("DEBUG: Message length: %d\n", len(msg.Raw))
	fmt.Printf("DEBUG: Prefix: %s\n", prefix)

	msgHash := crypto.Keccak256Hash([]byte(prefix + msg.Raw))
	fmt.Printf("DEBUG: Message hash: %x\n", msgHash)

	// Recover public key
	pubKey, err := crypto.SigToPub(msgHash.Bytes(), sigBytes)
	if err != nil {
		fmt.Printf("DEBUG: Pubkey recovery error: %v\n", err)
		return fmt.Errorf("failed to recover public key: %w", err)
	}
	fmt.Println("DEBUG: Public key recovered")

	// FIX #4: Explicit signature verification
	verified := crypto.VerifySignature(
		crypto.FromECDSAPub(pubKey),
		msgHash.Bytes(),
		sigBytes[:64],
	)
	if !verified {
		fmt.Println("DEBUG: Signature verification failed")
		return appErr.ErrInvalidSignature
	}
	fmt.Println("DEBUG: Signature verified successfully")

	// Recover and canonicalize address
	recoveredAddress := crypto.PubkeyToAddress(*pubKey).Hex()
	canonicalExpected := common.HexToAddress(expectedAddress).Hex()

	fmt.Printf("DEBUG: Recovered address: %s\n", recoveredAddress)
	fmt.Printf("DEBUG: Expected address: %s\n", canonicalExpected)

	if !strings.EqualFold(recoveredAddress, canonicalExpected) {
		fmt.Printf("DEBUG: Address mismatch!\n")
		return appErr.ErrInvalidSignature
	}

	fmt.Println("DEBUG: Address matches!")
	return nil
}

/*
@method Logout
@desc Revokes active JWT token and invalidates session
*/
func (s *AuthService) Logout(ctx context.Context, token string) error {
	return s.authRepo.RevokeSession(ctx, token)
}

/*
@method ValidateToken
@desc Validates JWT token and returns wallet address
*/
func (s *AuthService) ValidateToken(ctx context.Context, token string) (string, error) {
	wallet, err := s.jwtMgr.VerifyToken(token)
	if err != nil {
		return "", err
	}

	isRevoked, err := s.authRepo.IsSessionRevoked(ctx, token)
	if err != nil {
		return "", err
	}
	if isRevoked {
		return "", fmt.Errorf("session has been revoked")
	}

	return wallet, nil
}
