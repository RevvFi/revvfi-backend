package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/Revvfi/revvfi-backend/internal/api/dto/request"
	"github.com/Revvfi/revvfi-backend/internal/api/dto/response"
	appErr "github.com/Revvfi/revvfi-backend/internal/pkg/errors"
)

/*
@file admin_borrowers.go

@desc
HTTP handlers for borrower management admin endpoints.
Manages borrower addition, removal, verification, and listing.
Provides administrative oversight of the borrower registry.

@handler_set
- ListBorrowers: GET /api/v1/admin/borrowers - List all borrowers (paginated)
- GetBorrower: GET /api/v1/admin/borrowers/:address - Get single borrower details
- GetPendingBorrowers: GET /api/v1/admin/borrowers/pending - List pending borrowers
- PrepareBorrowerAddition: POST /api/v1/admin/borrowers/:address/prepare - Prepare addBorrower tx
- PrepareBorrowerRemoval: DELETE /api/v1/admin/borrowers/:address/prepare - Prepare removeBorrower tx

@dependencies
- AdminBorrowerServiceInterface: Service for borrower operations
*/

/*
@interface AdminBorrowerServiceInterface

@desc
Service interface for borrower management operations.
Implements clean architecture separation of concerns.
All methods require context for potential async operations.

@methods
- ListBorrowers: Retrieve paginated list of all borrowers
- GetBorrower: Fetch single borrower with detailed metrics
- GetPendingBorrowers: List borrowers awaiting verification
- PrepareBorrowerAddition: Prepare transaction to add borrower
- PrepareBorrowerRemoval: Prepare transaction to remove borrower
*/
type AdminBorrowerServiceInterface interface {
	/*
		@method ListBorrowers
		@desc Retrieve paginated list of all borrowers with filtering.
		@params
		- ctx: Request context
		- page: Page number (1-indexed)
		- limit: Items per page (1-100)
		- status: Filter by verification status (optional)
		- search: Search borrower address/name (optional)
		@returns
		- response.BorrowerListResponse: Paginated borrower list
		- error: Any service error
	*/
	ListBorrowers(ctx interface{}, page, limit int32, status, search string) (*response.BorrowerListResponse, error)

	/*
		@method GetBorrower
		@desc Fetch detailed information about single borrower.
		@params
		- ctx: Request context
		- borrowerAddress: Ethereum address of borrower
		@returns
		- response.BorrowerInfo: Borrower details
		- error: Any service error (e.g., not found)
	*/
	GetBorrower(ctx interface{}, borrowerAddress string) (*response.BorrowerInfo, error)

	/*
		@method GetPendingBorrowers
		@desc Retrieve list of borrowers pending verification.
		@params
		- ctx: Request context
		@returns
		- response.PendingBorrowerResponse: List of pending borrowers
		- error: Any service error
	*/
	GetPendingBorrowers(ctx interface{}) (*response.PendingBorrowerResponse, error)

	/*
		@method PrepareBorrowerAddition
		@desc Prepare transaction to add borrower to marketplace.
		@params
		- ctx: Request context
		- borrowerAddress: Address to add
		- isVerified: Whether to mark as verified
		@returns
		- string: Encoded transaction data
		- error: Any service error
	*/
	PrepareBorrowerAddition(ctx interface{}, borrowerAddress string, isVerified bool) (string, error)

	/*
		@method PrepareBorrowerRemoval
		@desc Prepare transaction to remove borrower from marketplace.
		@params
		- ctx: Request context
		- borrowerAddress: Address to remove
		- reason: Reason for removal (audit trail)
		- closePositions: Whether to force-close open positions
		@returns
		- string: Encoded transaction data
		- error: Any service error
	*/
	PrepareBorrowerRemoval(ctx interface{}, borrowerAddress, reason string, closePositions bool) (string, error)
}

/*
@struct AdminBorrowerHandler

@desc
HTTP handler for borrower management endpoints.
Delegates business logic to service layer.
Responsible for HTTP parsing, validation, and response formatting.

@fields
- borrowerService: AdminBorrowerServiceInterface for dependency injection
*/
type AdminBorrowerHandler struct {
	borrowerService AdminBorrowerServiceInterface
}

/*
@constructor NewAdminBorrowerHandler

@desc
Initialize new admin borrower handler with dependencies.
Implements dependency injection pattern for testability.

@params
- svc: AdminBorrowerServiceInterface implementation

@returns
- *AdminBorrowerHandler: Configured handler
*/
func NewAdminBorrowerHandler(svc AdminBorrowerServiceInterface) *AdminBorrowerHandler {
	return &AdminBorrowerHandler{
		borrowerService: svc,
	}
}

/*
@handler ListBorrowers

@desc
Retrieve paginated list of borrowers with optional filtering.
Allows searching by status (verified/pending/blocked) and address.

@route GET /api/v1/admin/borrowers

@request
Query Parameters:
- page: Page number (default: 1, min: 1)
- limit: Items per page (default: 20, min: 1, max: 100)
- status: Filter by status - "verified", "pending", "blocked" (optional)
- search: Search by address or name (optional, max 256 chars)

@response
HTTP 200 OK
Body:

	{
	  "success": true,
	  "data": {
	    "borrowers": [
	      {
	        "address": "0x...",
	        "is_verified": true,
	        "reputation_score": 850,
	        "total_borrowed": "5000000000000000000",
	        "total_repaid": "4500000000000000000",
	        "active_offers": 2,
	        "default_count": 0,
	        "added_at": 1685356800,
	        "verified_at": 1685360000
	      }
	    ],
	    "pagination": {
	      "page": 1,
	      "limit": 20,
	      "total": 150
	    },
	    "total_reputation_average": 820.5
	  }
	}

@error_cases
- 400: Invalid page/limit values
- 401: Unauthorized (not authenticated)
- 403: Forbidden (insufficient permissions)
- 500: Service error
*/
func (h *AdminBorrowerHandler) ListBorrowers(c *gin.Context) {
	/*
		@logic ExtractParameters
		@desc Extract and validate pagination parameters from query string.
	*/
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")
	status := c.Query("status")
	search := c.Query("search")

	/*
		@logic ParseParameters
		@desc Convert string parameters to integers with validation.
	*/
	page, err := strconv.ParseInt(pageStr, 10, 32)
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_PAGINATION",
				"message": "page must be a positive integer",
			},
		})
		return
	}

	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_PAGINATION",
				"message": "limit must be between 1 and 100",
			},
		})
		return
	}

	/*
		@logic ValidateStatus
		@desc Validate status filter if provided.
	*/
	if status != "" && status != "verified" && status != "pending" && status != "blocked" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_STATUS_FILTER",
				"message": "status must be one of: verified, pending, blocked",
			},
		})
		return
	}

	/*
		@logic CallService
		@desc Fetch borrower list from service layer.
	*/
	result, err := h.borrowerService.ListBorrowers(c, int32(page), int32(limit), status, search)
	if err != nil {
		statusCode := appErr.StatusCode(err)
		c.JSON(statusCode, gin.H{
			"success": false,
			"error": gin.H{
				"code":    appErr.ErrorCode(err),
				"message": err.Error(),
			},
		})
		return
	}

	/*
		@logic ReturnSuccess
		@desc Return paginated borrower list with metadata.
	*/
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

/*
@handler GetBorrower

@desc
Retrieve detailed information about a specific borrower.
Includes reputation metrics, borrowing history, and platform activity.

@route GET /api/v1/admin/borrowers/:address

@request
Path Parameters:
- address: Ethereum address of borrower (required, valid eth address)

@response
HTTP 200 OK
Body:

	{
	  "success": true,
	  "data": {
	    "address": "0x...",
	    "is_verified": true,
	    "reputation_score": 850,
	    "total_borrowed": "5000000000000000000",
	    "total_repaid": "4500000000000000000",
	    "active_offers": 2,
	    "default_count": 0,
	    "added_at": 1685356800,
	    "verified_at": 1685360000
	  }
	}

@error_cases
- 400: Invalid/missing address parameter
- 404: Borrower not found
- 500: Service error
*/
func (h *AdminBorrowerHandler) GetBorrower(c *gin.Context) {
	/*
		@logic ExtractParameter
		@desc Extract borrower address from URL path.
	*/
	address := c.Param("address")

	/*
		@logic ValidateAddress
		@desc Validate address format (basic check, further validation in service).
	*/
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_ADDRESS",
				"message": "borrower address is required",
			},
		})
		return
	}

	/*
		@logic CallService
		@desc Fetch borrower from service layer.
	*/
	borrower, err := h.borrowerService.GetBorrower(c, address)
	if err != nil {
		statusCode := appErr.StatusCode(err)
		c.JSON(statusCode, gin.H{
			"success": false,
			"error": gin.H{
				"code":    appErr.ErrorCode(err),
				"message": err.Error(),
			},
		})
		return
	}

	/*
		@logic ReturnSuccess
		@desc Return borrower details.
	*/
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    borrower,
	})
}

/*
@handler GetPendingBorrowers

@desc
Retrieve list of all borrowers awaiting verification.
Helps admins track pending approvals and timeline.

@route GET /api/v1/admin/borrowers/pending

@request
No query parameters required.

@response
HTTP 200 OK
Body:

	{
	  "success": true,
	  "data": {
	    "pending_count": 5,
	    "borrowers": [
	      {
	        "address": "0x...",
	        "is_verified": false,
	        "reputation_score": 0,
	        "total_borrowed": "0",
	        "total_repaid": "0",
	        "active_offers": 0,
	        "default_count": 0,
	        "added_at": 1685356800
	      }
	    ],
	    "oldest_pending_at": 1685356800
	  }
	}

@error_cases
- 401: Unauthorized (not authenticated)
- 403: Forbidden (insufficient permissions)
- 500: Service error
*/
func (h *AdminBorrowerHandler) GetPendingBorrowers(c *gin.Context) {
	/*
		@logic CallService
		@desc Fetch pending borrowers from service layer.
	*/
	result, err := h.borrowerService.GetPendingBorrowers(c)
	if err != nil {
		statusCode := appErr.StatusCode(err)
		c.JSON(statusCode, gin.H{
			"success": false,
			"error": gin.H{
				"code":    appErr.ErrorCode(err),
				"message": err.Error(),
			},
		})
		return
	}

	/*
		@logic ReturnSuccess
		@desc Return list of pending borrowers.
	*/
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

/*
@handler PrepareBorrowerAddition

@desc
Prepare transaction to add borrower to the marketplace.
Can mark borrower as pre-verified at addition time.
Returns encoded transaction for signing/execution.

@route POST /api/v1/admin/borrowers/:address/prepare

@request
Path Parameters:
- address: Ethereum address to add as borrower

Body (JSON):

	{
	  "is_verified": true
	}

@response
HTTP 200 OK
Body:

	{
	  "success": true,
	  "data": {
	    "tx_data": "0x...",
	    "target": "0x...",
	    "value": "0",
	    "description": "Add borrower: 0x... (verified)",
	    "encoded_proposal": "0x..."
	  }
	}

@error_cases
- 400: Invalid address or request body
- 409: Borrower already exists
- 500: Service error
*/
func (h *AdminBorrowerHandler) PrepareBorrowerAddition(c *gin.Context) {
	/*
		@logic ExtractParameter
		@desc Extract borrower address from URL path.
	*/
	address := c.Param("address")

	/*
		@logic ValidateAddress
		@desc Validate address format.
	*/
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_ADDRESS",
				"message": "borrower address is required",
			},
		})
		return
	}

	/*
		@logic ParseRequest
		@desc Parse request body for verification flag.
	*/
	var req struct {
		IsVerified bool `json:"is_verified"`
	}

	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": fmt.Sprintf("invalid request body: %v", err),
			},
		})
		return
	}

	/*
		@logic CallService
		@desc Prepare transaction from service layer.
	*/
	txData, err := h.borrowerService.PrepareBorrowerAddition(c, address, req.IsVerified)
	if err != nil {
		statusCode := appErr.StatusCode(err)
		c.JSON(statusCode, gin.H{
			"success": false,
			"error": gin.H{
				"code":    appErr.ErrorCode(err),
				"message": err.Error(),
			},
		})
		return
	}

	/*
		@logic ReturnSuccess
		@desc Return encoded transaction data.
	*/
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"tx_data": txData,
		},
	})
}

/*
@handler PrepareBorrowerRemoval

@desc
Prepare transaction to remove borrower from marketplace.
Can optionally force-close all borrower positions.
Returns encoded transaction for signing/execution.

@route DELETE /api/v1/admin/borrowers/:address/prepare

@request
Path Parameters:
- address: Ethereum address to remove

Body (JSON):

	{
	  "reason": "Violation of terms",
	  "close_existing_positions": true
	}

@response
HTTP 200 OK
Body:

	{
	  "success": true,
	  "data": {
	    "tx_data": "0x...",
	    "target": "0x...",
	    "value": "0",
	    "description": "Remove borrower: 0x... (Reason: Violation of terms)",
	    "encoded_proposal": "0x..."
	  }
	}

@error_cases
- 400: Invalid address or request body, missing reason
- 404: Borrower not found
- 409: Cannot remove borrower (e.g., has active protected positions)
- 500: Service error
*/
func (h *AdminBorrowerHandler) PrepareBorrowerRemoval(c *gin.Context) {
	/*
		@logic ExtractParameter
		@desc Extract borrower address from URL path.
	*/
	address := c.Param("address")

	/*
		@logic ValidateAddress
		@desc Validate address format.
	*/
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_ADDRESS",
				"message": "borrower address is required",
			},
		})
		return
	}

	/*
		@logic ParseRequest
		@desc Parse request body for removal reason and position handling.
	*/
	var req request.RemoveBorrowerRequest

	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": fmt.Sprintf("invalid request body: %v", err),
			},
		})
		return
	}

	/*
		@logic ValidateRequest
		@desc Validate required fields in request.
	*/
	if req.Reason == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "MISSING_FIELD",
				"message": "reason is required",
			},
		})
		return
	}

	/*
		@logic CallService
		@desc Prepare transaction from service layer.
	*/
	txData, err := h.borrowerService.PrepareBorrowerRemoval(c, address, req.Reason, req.CloseExistingPositions)
	if err != nil {
		statusCode := appErr.StatusCode(err)
		c.JSON(statusCode, gin.H{
			"success": false,
			"error": gin.H{
				"code":    appErr.ErrorCode(err),
				"message": err.Error(),
			},
		})
		return
	}

	/*
		@logic ReturnSuccess
		@desc Return encoded transaction data.
	*/
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"tx_data": txData,
		},
	})
}
