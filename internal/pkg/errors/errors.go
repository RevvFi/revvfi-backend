package errors

import (
	"fmt"

	/*
	   @file errors.go

	   @desc
	   Custom domain-specific error types for RevvFi protocol.
	   Provides structured error handling across the entire application.

	   @responsibilities
	   - Define standard error constants
	   - Create typed errors for API responses
	   - Map errors to HTTP status codes
	*/

	stdErrors "errors"
)

var (
	// Authentication errors
	ErrUnauthorized = stdErrors.New("unauthorized")
	ErrInvalidToken = stdErrors.New("invalid or expired token")
	ErrInvalidNonce = stdErrors.New("invalid or expired nonce")

	// Resource not found errors
	ErrMarketNotFound         = stdErrors.New("market not found")
	ErrPositionNotFound       = stdErrors.New("position not found")
	ErrOfferNotFound          = stdErrors.New("offer not found")
	ErrBorrowerNotFound       = stdErrors.New("borrower not found")
	ErrAuctionNotFound        = stdErrors.New("auction not found")
	ErrWithdrawalNotFound     = stdErrors.New("withdrawal request not found")
	ErrWithdrawalEpochNotFound = stdErrors.New("withdrawal epoch not found")

	// Validation errors
	ErrInvalidInput         = stdErrors.New("invalid input")
	ErrInvalidAddress       = stdErrors.New("invalid ethereum address")
	ErrInvalidAmount        = stdErrors.New("invalid amount")
	ErrInvalidAPR           = stdErrors.New("invalid APR")
	ErrInvalidSignature     = stdErrors.New("invalid signature")
	ErrInvalidMarketConfig  = stdErrors.New("invalid market configuration")

	// Business logic errors
	ErrInsufficientLiquidity      = stdErrors.New("insufficient liquidity available")
	ErrInsufficientCollateral     = stdErrors.New("insufficient collateral")
	ErrMarketAlreadyExists        = stdErrors.New("market already exists")
	ErrMarketClosed               = stdErrors.New("market is closed")
	ErrMarketLiquidating          = stdErrors.New("market is liquidating")
	ErrOfferExpired               = stdErrors.New("offer has expired")
	ErrOfferNotActive             = stdErrors.New("offer is not active")
	ErrPositionAlreadySettled     = stdErrors.New("position already settled")
	ErrCannotWithdrawFromActive   = stdErrors.New("cannot withdraw from active position")

	// Liquidation errors
	ErrNotLiquidatable        = stdErrors.New("position is not liquidatable")
	ErrAuctionAlreadyActive   = stdErrors.New("auction already active for this market")
	ErrAuctionEnded           = stdErrors.New("auction has ended")
	ErrAuctionNotActive       = stdErrors.New("auction is not active")
	ErrAuctionExpired         = stdErrors.New("auction has expired")
	ErrLiquidationNotRequired = stdErrors.New("liquidation not required")
	ErrInsufficientBidAmount  = stdErrors.New("bid amount below current price")
	ErrAuctionNotEnded        = stdErrors.New("auction has not ended")
	ErrAuctionAlreadySettled  = stdErrors.New("auction already settled")

	// Blockchain errors
	ErrBlockchainError      = stdErrors.New("blockchain interaction failed")
	ErrOracleStale          = stdErrors.New("price oracle data is stale")
	ErrOracleNotAvailable   = stdErrors.New("price oracle not available")
	ErrContractNotFound     = stdErrors.New("contract not found on chain")
	ErrTransactionFailed    = stdErrors.New("transaction execution failed")

	// Database errors
	ErrDatabaseError       = stdErrors.New("database operation failed")
	ErrDuplicateEntry      = stdErrors.New("duplicate entry")
	ErrDatabaseTxFailed    = stdErrors.New("database transaction failed")

	// System errors
	ErrInternal       = stdErrors.New("internal server error")
	ErrNotImplemented = stdErrors.New("not implemented")

	// Withdrawal errors
	ErrInvalidWithdrawalStatus = stdErrors.New("invalid withdrawal status")
	ErrInsufficientBalance     = stdErrors.New("insufficient balance")
	ErrEpochNotFound           = stdErrors.New("epoch not found")
)

/*
@struct AppError

@desc
Structured error wrapper for API responses.
Includes error code, message, and optional details.

@fields
- Code: machine-readable error code
- Message: human-readable message
- Details: optional error details
- Err: underlying error
*/
type AppError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details any `json:"details,omitempty"`
	Err     error       `json:"-"`
}

/*
@method Error

@desc
Implements error interface for AppError.

@returns
- string: formatted error message
*/
func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

/*
@method Unwrap

@desc
Implements errors.Wrapper interface.

@returns
- error: wrapped error
*/
func (e *AppError) Unwrap() error {
	return e.Err
}

/*
@function NewAppError

@desc
Creates new AppError.

@params
- code: error code
- message: error message
- err: underlying error

@returns
- *AppError
*/
func NewAppError(code string, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

/*
@function NewAppErrorWithDetails

@desc
Creates new AppError with details.

@params
- code: error code
- message: error message
- details: optional details
- err: underlying error

@returns
- *AppError
*/
func NewAppErrorWithDetails(code string, message string, details any, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Details: details,
		Err:     err,
	}
}

/*
@function StatusCode

@desc
Maps error to HTTP status code.

@params
- err: error to check

@returns
- int: HTTP status code
*/
func StatusCode(err error) int {
	if err == nil {
		return 200
	}

	switch err {
	case ErrUnauthorized, ErrInvalidToken, ErrInvalidNonce:
		return 401
	case ErrMarketNotFound, ErrPositionNotFound, ErrOfferNotFound, ErrBorrowerNotFound,
		ErrAuctionNotFound, ErrWithdrawalNotFound:
		return 404
	case ErrInvalidInput, ErrInvalidAddress, ErrInvalidAmount, ErrInvalidAPR,
		ErrInvalidSignature, ErrInvalidMarketConfig:
		return 400
	case ErrInsufficientLiquidity, ErrInsufficientCollateral, ErrMarketAlreadyExists,
		ErrMarketClosed, ErrMarketLiquidating, ErrOfferExpired, ErrOfferNotActive:
		return 422
	case ErrInternal, ErrDatabaseError:
		return 500
	default:
		return 500
	}
}

/*
@function ErrorCode

@desc
Maps error to machine-readable code.

@params
- err: error to check

@returns
- string: error code
*/
func ErrorCode(err error) string {
	if err == nil {
		return "OK"
	}

	switch err {
	case ErrUnauthorized:
		return "UNAUTHORIZED"
	case ErrInvalidToken:
		return "INVALID_TOKEN"
	case ErrInvalidNonce:
		return "INVALID_NONCE"
	case ErrMarketNotFound:
		return "MARKET_NOT_FOUND"
	case ErrPositionNotFound:
		return "POSITION_NOT_FOUND"
	case ErrOfferNotFound:
		return "OFFER_NOT_FOUND"
	case ErrBorrowerNotFound:
		return "BORROWER_NOT_FOUND"
	case ErrAuctionNotFound:
		return "AUCTION_NOT_FOUND"
	case ErrWithdrawalNotFound:
		return "WITHDRAWAL_NOT_FOUND"
	case ErrWithdrawalEpochNotFound:
		return "EPOCH_NOT_FOUND"
	case ErrInvalidInput, ErrInvalidAddress, ErrInvalidAmount, ErrInvalidAPR:
		return "INVALID_REQUEST"
	case ErrInsufficientLiquidity:
		return "INSUFFICIENT_LIQUIDITY"
	case ErrInsufficientCollateral:
		return "INSUFFICIENT_COLLATERAL"
	case ErrMarketClosed:
		return "MARKET_CLOSED"
	case ErrMarketLiquidating:
		return "MARKET_LIQUIDATING"
	case ErrOfferExpired:
		return "OFFER_EXPIRED"
	case ErrAuctionNotActive:
		return "AUCTION_NOT_ACTIVE"
	case ErrAuctionAlreadySettled:
		return "AUCTION_ALREADY_SETTLED"
	case ErrNotLiquidatable:
		return "NOT_LIQUIDATABLE"
	default:
		return "INTERNAL_ERROR"
	}
}
