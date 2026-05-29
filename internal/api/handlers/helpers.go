package handlers

import (
	"errors"
	"math/big"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	appErr "github.com/Revvfi/revvfi-backend/internal/pkg/errors"
)

/*
@function writeError

@desc
Writes a standardized JSON error response from a domain or infrastructure error.

@responsibilities
- Map known domain errors to HTTP status codes
- Return a stable error envelope
- Hide internal details from clients

@params
- c: Gin request context
- err: error to render
*/
func writeError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	status := appErr.StatusCode(rootError(err))
	c.JSON(status, gin.H{
		"error": gin.H{
			"code":    appErr.ErrorCode(rootError(err)),
			"message": err.Error(),
		},
	})
}

/*
@function rootError

@desc
Finds the deepest wrapped error for domain error mapping.

@responsibilities
- Walk wrapped error chains
- Preserve sentinel errors for status mapping

@params
- err: wrapped error

@returns
- error
*/
func rootError(err error) error {
	if err == nil {
		return nil
	}

	known := []error{
		appErr.ErrUnauthorized,
		appErr.ErrInvalidToken,
		appErr.ErrInvalidNonce,
		appErr.ErrMarketNotFound,
		appErr.ErrPositionNotFound,
		appErr.ErrOfferNotFound,
		appErr.ErrBorrowerNotFound,
		appErr.ErrAuctionNotFound,
		appErr.ErrWithdrawalNotFound,
		appErr.ErrInvalidInput,
		appErr.ErrInvalidAddress,
		appErr.ErrInvalidAmount,
		appErr.ErrInvalidAPR,
		appErr.ErrInsufficientLiquidity,
	}
	for _, target := range known {
		if errors.Is(err, target) {
			return target
		}
	}
	return err
}

/*
@function parseBigInt

@desc
Parses a base-10 integer string into a big.Int pointer.

@responsibilities
- Validate decimal integer input
- Return domain validation errors for malformed values

@params
- value: decimal string

@returns
- *big.Int
- error
*/
func parseBigInt(value string) (*big.Int, error) {
	parsed, ok := new(big.Int).SetString(value, 10)
	if !ok || parsed.Sign() < 0 {
		return nil, appErr.ErrInvalidAmount
	}
	return parsed, nil
}

/*
@function pagination

@desc
Calculates limit and offset values from page and page-size inputs.

@responsibilities
- Apply default pagination settings
- Convert 1-indexed page numbers to offsets

@params
- page: requested page number
- pageSize: requested page size

@returns
- int32
- int32
*/
func pagination(page int32, pageSize int32) (int32, int32) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return pageSize, (page - 1) * pageSize
}

/*
@function pathInt64

@desc
Parses an int64 path parameter from Gin context.

@responsibilities
- Read path parameter by name
- Return validation error for malformed values

@params
- c: Gin request context
- key: path parameter name

@returns
- int64
- error
*/
func pathInt64(c *gin.Context, key string) (int64, error) {
	value := c.Param(key)
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		return 0, appErr.ErrInvalidInput
	}
	return parsed, nil
}

/*
@function ok

@desc
Writes a successful JSON response.

@responsibilities
- Keep response writing consistent across handlers
- Use caller-provided HTTP status codes

@params
- c: Gin request context
- status: HTTP status code
- payload: response payload
*/
func ok(c *gin.Context, status int, payload interface{}) {
	if status == 0 {
		status = http.StatusOK
	}
	c.JSON(status, payload)
}
