// internal/pkg/errors/auth_errors.go
package errors

import "errors"

/*
@file auth_errors.go

@desc
Authentication-specific error definitions for SIWE and JWT operations.
These errors are used across auth service, handlers, and repositories.

@error_categories
- Nonce errors: Storage, validation, consumption
- Message errors: SIWE parsing, size limits
- Address errors: Validation, mismatches
- Chain validation: Chain ID validation
- URI/Domain validation: Format and host validation
- Signature errors: Verification failures
- Rate limiting: Request throttling
- Time validation: Expiration, future dates
- Session errors: Storage, revocation
- JWT errors: Token generation, verification
*/

// =====================================================
// NONCE ERRORS
// =====================================================

/*
@var ErrNonceNotFound
@desc Nonce does not exist in database
@trigger When a nonce from SIWE message is not found in auth_nonces table
*/
var ErrNonceNotFound = errors.New("nonce not found")

/*
@var ErrNonceAlreadyUsed
@desc Nonce has already been consumed (replay attack detected)
@trigger When a nonce that has already been used is presented again
*/
var ErrNonceAlreadyUsed = errors.New("nonce already used")

/*
@var ErrNonceExpired
@desc Nonce has passed its expiration time (TTL exceeded)
@trigger When a nonce older than 5 minutes is used
*/
var ErrNonceExpired = errors.New("nonce expired")

/*
@var ErrInvalidNonce
@desc Nonce format is invalid (wrong length)
@trigger When nonce length is less than 8 or greater than 64 characters
*/
var ErrInvalidNonce = errors.New("invalid nonce format")

/*
@var ErrInvalidNonceFormat
@desc Nonce contains invalid characters (not alphanumeric)
@trigger When nonce contains characters outside A-Z, a-z, 0-9
@spec EIP-4361 requires nonce to be ALPHA / DIGIT only
*/
var ErrInvalidNonceFormat = errors.New("nonce must contain only alphanumeric characters (A-Z, a-z, 0-9)")

// =====================================================
// MESSAGE ERRORS
// =====================================================

/*
@var ErrMessageTooLarge
@desc SIWE message exceeds maximum allowed size
@trigger When message length > MaxSIWEMessageSize (8192 bytes)
@security Prevents DoS attacks via large message payloads
*/
var ErrMessageTooLarge = errors.New("SIWE message too large (max 8KB)")

/*
@var ErrInvalidSIWEMessage
@desc SIWE message format is invalid or incomplete
@trigger When message doesn't conform to EIP-4361 specification
*/
var ErrInvalidSIWEMessage = errors.New("invalid SIWE message format")

// =====================================================
// ADDRESS ERRORS
// =====================================================

/*
@var ErrInvalidAddress
@desc Ethereum address format is invalid
@trigger When address fails common.IsHexAddress() validation
*/
var ErrInvalidAddress = errors.New("invalid Ethereum address")

/*
@var ErrAddressMismatch
@desc Wallet address in message doesn't match authenticated wallet
@trigger When SIWE message address != wallet from auth context
*/
var ErrAddressMismatch = errors.New("address mismatch between message and authenticated wallet")

// =====================================================
// CHAIN VALIDATION ERRORS
// =====================================================

/*
@var ErrInvalidChainID
@desc Chain ID in SIWE message is not allowed
@trigger When chain ID in message is not in AllowedChains config
*/
var ErrInvalidChainID = errors.New("chain ID not allowed")

// =====================================================
// URI/DOMAIN VALIDATION ERRORS
// =====================================================

/*
@var ErrInvalidDomain
@desc Domain in SIWE message doesn't match expected domain
@trigger When domain in message != config.Domain
*/
var ErrInvalidDomain = errors.New("invalid domain")

/*
@var ErrInvalidURI
@desc URI format is invalid (cannot be parsed)
@trigger When url.Parse() fails on the URI field
*/
var ErrInvalidURI = errors.New("invalid URI format")

/*
@var ErrHostNotAllowed
@desc URI host is not in allowed hosts whitelist
@trigger When parsed hostname not in config.AllowedHosts
*/
var ErrHostNotAllowed = errors.New("URI host not allowed")

/*
@var ErrInvalidScheme
@desc URI scheme is not allowed (only https/http)
@trigger When scheme is not "https" or "http"
*/
var ErrInvalidScheme = errors.New("invalid URI scheme (only https/http allowed)")

// =====================================================
// SIGNATURE ERRORS
// =====================================================

/*
@var ErrInvalidSignature
@desc Cryptographic signature verification failed
@trigger When recovered address doesn't match wallet, or crypto.VerifySignature fails
*/
var ErrInvalidSignature = errors.New("invalid signature")

/*
@var ErrInvalidVValue
@desc Signature recovery ID (v) is not in allowed values
@trigger When v is not 0, 1, 27, or 28
*/
var ErrInvalidVValue = errors.New("invalid signature v value")

// =====================================================
// RATE LIMITING ERRORS
// =====================================================

/*
@var ErrRateLimitExceeded
@desc Rate limit exceeded for auth endpoint
@trigger When too many requests from same wallet/IP in time window
@limits /auth/nonce: 10 per minute, /auth/login: 20 per minute
*/
var ErrRateLimitExceeded = errors.New("rate limit exceeded. Please wait before retrying")

// =====================================================
// TIME VALIDATION ERRORS
// =====================================================

/*
@var ErrMessageExpired
@desc SIWE message has passed its expiration time
@trigger When current time > ExpirationTime field in message
*/
var ErrMessageExpired = errors.New("SIWE message has expired")

/*
@var ErrMessageNotYetValid
@desc SIWE message is not yet valid (Not Before in future)
@trigger When current time < NotBefore field in message
*/
var ErrMessageNotYetValid = errors.New("SIWE message not yet valid")

/*
@var ErrFutureIssuedAt
@desc Issued At timestamp is too far in the future (>5 minutes ahead)
@trigger When IssuedAt > now + 5 minutes (clock drift allowance)
*/
var ErrFutureIssuedAt = errors.New("issued at timestamp is in the future")

// =====================================================
// SESSION ERRORS
// =====================================================

/*
@var ErrSessionNotFound
@desc Session not found for given token
@trigger When token doesn't exist in auth_sessions table
*/
var ErrSessionNotFound = errors.New("session not found")

/*
@var ErrSessionExpired
@desc Session has passed its expiration time
@trigger When current time > session.expires_at
*/
var ErrSessionExpired = errors.New("session has expired")

/*
@var ErrSessionRevoked
@desc Session has been explicitly revoked
@trigger When session.revoked_at is not null
*/
var ErrSessionRevoked = errors.New("session has been revoked")

// =====================================================
// JWT ERRORS
// =====================================================

/*
@var ErrInvalidToken
@desc JWT token format is invalid or malformed
@trigger When token cannot be parsed as JWT
*/
var ErrInvalidToken = errors.New("invalid token format")

/*
@var ErrTokenExpired
@desc JWT token has passed its expiration
@trigger When token.Exp > current time
*/
var ErrTokenExpired = errors.New("token has expired")

/*
@var ErrInvalidTokenSignature
@desc JWT token signature verification failed
@trigger When token signature doesn't match signing secret
*/
var ErrInvalidTokenSignature = errors.New("invalid token signature")

/*
@var ErrMissingTokenClaim
@desc Required claim (e.g., wallet) is missing from JWT
@trigger When token payload missing required fields
*/
var ErrMissingTokenClaim = errors.New("missing required token claim")

// =====================================================
// AUTHORIZATION ERRORS
// =====================================================

/*
@var ErrUnauthorized
@desc Generic unauthorized error for auth failures
@trigger When authentication fails for any reason
*/
var ErrUnauthorized = errors.New("unauthorized")

/*
@var ErrInsufficientPermissions
@desc Authenticated user lacks required permissions
@trigger When trying to access admin-only endpoint with non-admin token
*/
var ErrInsufficientPermissions = errors.New("insufficient permissions")

/*
@var ErrInvalidAuthHeader
@desc Authorization header format is invalid
@trigger When header doesn't contain "Bearer " prefix or token
*/
var ErrInvalidAuthHeader = errors.New("invalid authorization header format")

// =====================================================
// EIP-1271 ERRORS (Contract Wallets)
// =====================================================

/*
@var ErrContractWalletNotSupported
@desc Contract wallet (EIP-1271) verification not implemented
@trigger When attempting to authenticate with a contract wallet
@phase This will be implemented in Phase 2
*/
var ErrContractWalletNotSupported = errors.New("contract wallet not supported (EIP-1271 not implemented)")

/*
@var ErrEIP1271VerificationFailed
@desc EIP-1271 signature verification failed on contract
@trigger When contract.isValidSignature returns invalid magic value
*/
var ErrEIP1271VerificationFailed = errors.New("EIP-1271 signature verification failed")

// =====================================================
// DATABASE/INTERNAL ERRORS
// =====================================================

/*
@var ErrDatabaseError
@desc Generic database operation error
@trigger When database query fails (connection, constraint, etc.)
*/
var ErrDatabaseError = errors.New("database error")

/*
@var ErrDuplicateEntry
@desc Unique constraint violation in database
@trigger When attempting to insert duplicate (wallet, nonce)
*/
var ErrDuplicateEntry = errors.New("duplicate entry")

/*
@var ErrInternalServer
@desc Generic internal server error
@trigger For unexpected errors that shouldn't be exposed to clients
*/
var ErrInternalServer = errors.New("internal server error")

// =====================================================
// ERROR MESSAGE HELPERS
// =====================================================

/*
@function ErrorCode
@desc Returns a string code for each error type
@param err error - The error to get code for
@returns string - Error code for API responses
*/
func ErrorCode(err error) string {
    switch err {
    // Nonce errors
    case ErrNonceNotFound:
        return "NONCE_NOT_FOUND"
    case ErrNonceAlreadyUsed:
        return "NONCE_ALREADY_USED"
    case ErrNonceExpired:
        return "NONCE_EXPIRED"
    case ErrInvalidNonce:
        return "INVALID_NONCE"
    case ErrInvalidNonceFormat:
        return "INVALID_NONCE_FORMAT"
    
    // Message errors
    case ErrMessageTooLarge:
        return "MESSAGE_TOO_LARGE"
    case ErrInvalidSIWEMessage:
        return "INVALID_SIWE_MESSAGE"
    
    // Address errors
    case ErrInvalidAddress:
        return "INVALID_ADDRESS"
    case ErrAddressMismatch:
        return "ADDRESS_MISMATCH"
    
    // Chain errors
    case ErrInvalidChainID:
        return "INVALID_CHAIN_ID"
    
    // URI/Domain errors
    case ErrInvalidDomain:
        return "INVALID_DOMAIN"
    case ErrInvalidURI:
        return "INVALID_URI"
    case ErrHostNotAllowed:
        return "HOST_NOT_ALLOWED"
    case ErrInvalidScheme:
        return "INVALID_SCHEME"
    
    // Signature errors
    case ErrInvalidSignature:
        return "INVALID_SIGNATURE"
    case ErrInvalidVValue:
        return "INVALID_V_VALUE"
    
    // Rate limiting
    case ErrRateLimitExceeded:
        return "RATE_LIMIT_EXCEEDED"
    
    // Time validation
    case ErrMessageExpired:
        return "MESSAGE_EXPIRED"
    case ErrMessageNotYetValid:
        return "MESSAGE_NOT_YET_VALID"
    case ErrFutureIssuedAt:
        return "FUTURE_ISSUED_AT"
    
    // Session errors
    case ErrSessionNotFound:
        return "SESSION_NOT_FOUND"
    case ErrSessionExpired:
        return "SESSION_EXPIRED"
    case ErrSessionRevoked:
        return "SESSION_REVOKED"
    
    // JWT errors
    case ErrInvalidToken:
        return "INVALID_TOKEN"
    case ErrTokenExpired:
        return "TOKEN_EXPIRED"
    case ErrInvalidTokenSignature:
        return "INVALID_TOKEN_SIGNATURE"
    case ErrMissingTokenClaim:
        return "MISSING_TOKEN_CLAIM"
    
    // Authorization errors
    case ErrUnauthorized:
        return "UNAUTHORIZED"
    case ErrInsufficientPermissions:
        return "INSUFFICIENT_PERMISSIONS"
    case ErrInvalidAuthHeader:
        return "INVALID_AUTH_HEADER"
    
    // EIP-1271 errors
    case ErrContractWalletNotSupported:
        return "CONTRACT_WALLET_NOT_SUPPORTED"
    case ErrEIP1271VerificationFailed:
        return "EIP1271_VERIFICATION_FAILED"
    
    // Internal errors
    case ErrDatabaseError:
        return "DATABASE_ERROR"
    case ErrDuplicateEntry:
        return "DUPLICATE_ENTRY"
    
    default:
        return "INTERNAL_ERROR"
    }
}

/*
@function StatusCode
@desc Returns HTTP status code for each error type
@param err error - The error to get status code for
@returns int - HTTP status code
*/
func StatusCode(err error) int {
    switch err {
    // Bad Request (400)
    case ErrInvalidAddress, ErrInvalidNonce, ErrInvalidNonceFormat,
         ErrInvalidSIWEMessage, ErrMessageTooLarge, ErrInvalidURI,
         ErrInvalidDomain, ErrInvalidScheme, ErrInvalidVValue,
         ErrInvalidToken, ErrMissingTokenClaim, ErrInvalidAuthHeader:
        return 400
    
    // Unauthorized (401)
    case ErrUnauthorized, ErrInvalidSignature, ErrNonceNotFound,
         ErrNonceAlreadyUsed, ErrNonceExpired, ErrAddressMismatch,
         ErrInvalidChainID, ErrHostNotAllowed, ErrMessageExpired,
         ErrFutureIssuedAt, ErrInvalidTokenSignature, ErrTokenExpired,
         ErrSessionNotFound, ErrSessionExpired, ErrSessionRevoked:
        return 401
    
    // Forbidden (403)
    case ErrInsufficientPermissions:
        return 403
    
    // Too Many Requests (429)
    case ErrRateLimitExceeded:
        return 429
    
    // Not Found (404)
    case ErrNonceNotFound:
        return 404
    
    // Not Implemented (501)
    case ErrContractWalletNotSupported:
        return 501
    
    // Conflict (409)
    case ErrDuplicateEntry:
        return 409
    
    // Internal Server Error (500)
    case ErrDatabaseError, ErrInternalServer, ErrEIP1271VerificationFailed:
        return 500
    
    default:
        return 500
    }
}