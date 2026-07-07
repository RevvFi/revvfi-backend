package handlers

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	"github.com/Revvfi/revvfi-backend/internal/api/dto/request"
	"github.com/Revvfi/revvfi-backend/internal/api/dto/response"
	"github.com/Revvfi/revvfi-backend/internal/models"
	appErr "github.com/Revvfi/revvfi-backend/internal/pkg/errors"
	"github.com/gin-gonic/gin"
)

/*
@file admin.go

@desc
HTTP handlers for admin operations on RevvFi protocol.
Manages market creation, configuration, and administrative functions.

@responsibilities
- Market creation and configuration
- Market status updates
- Administrative queries
- Protocol parameter management
- Error handling and validation

@dependencies
- MarketService: business logic for market operations
- Validator: input validation
- Logger: structured logging
*/

/*
@interface MarketServiceInterface

@desc
Interface for market service operations (used for dependency injection and testing).
*/
type MarketServiceInterface interface {
	CreateMarket(ctx context.Context, marketAddr string, borrower, borrowAsset, collateralAsset, collateralOracle string, minRatio, liquidationThreshold int32) (*models.Market, error)
	GetMarket(ctx context.Context, address string) (*models.Market, error)
	ListMarkets(ctx context.Context, limit, offset int32, borrower string) ([]models.Market, error)
	CalculateMetrics(ctx context.Context, market *models.Market) (map[string]any, error)
}

/*
@interface MarketValidatorInterface

@desc
Interface for market validation operations (used for dependency injection and testing).
*/
type MarketValidatorInterface interface {
	ValidateMarketCreation(borrower, borrowAsset, collateralAsset string, minRatio, liquidationThreshold int32) error
}

/*
@struct AdminHandler

@desc
Handles HTTP requests for administrative operations.
Implements admin route handlers following clean architecture principles.

@responsibilities
- Parse and validate HTTP requests
- Call market service layer
- Return standardized HTTP responses
- Handle errors with proper status codes

@dependencies
- marketService: core business logic
- marketValidator: input validation
*/
type AdminHandler struct {
	marketService   MarketServiceInterface
	marketValidator MarketValidatorInterface
}

/*
@function NewAdminHandler

@desc
Creates new admin handler instance with dependency injection.
Follows constructor pattern for clean dependency management.

@params
- marketService: market service dependency
- marketValidator: validator dependency

@returns
- *AdminHandler: initialized handler
*/
func NewAdminHandler(
	marketService MarketServiceInterface,
	marketValidator MarketValidatorInterface,
) *AdminHandler {
	return &AdminHandler{
		marketService:   marketService,
		marketValidator: marketValidator,
	}
}

/*
@handler CreateMarket

@desc
HTTP POST handler for creating new isolated lending market.
Validates request, enforces protocol constraints, and persists market.

@route
POST /api/admin/markets

@request
CreateMarketRequest DTO containing:
- borrower: market creator address (42-char Ethereum address)
- borrowAsset: ERC20 token to borrow
- collateralAsset: ERC20 collateral token
- collateralOracle: Chainlink oracle address
- minCollateralRatio: minimum ratio in basis points (must be >= 10000)
- liquidationThreshold: liquidation trigger in bps (must be < minCollateralRatio)

@response
Success (201 Created):

	{
	  "success": true,
	  "data": {
	    "market_address": "0x...",
	    "borrower": "0x...",
	    "borrow_asset": "0x...",
	    "collateral_asset": "0x...",
	    "min_collateral_ratio": 11000,
	    "liquidation_threshold": 9500,
	    "created_at": "2026-05-30T12:00:00Z"
	  }
	}

Error (400/422):

	{
	  "success": false,
	  "error": {
	    "code": "INVALID_MARKET_CONFIG",
	    "message": "collateral ratio must be >= 100%"
	  }
	}

@error_cases
- ErrInvalidInput: malformed request
- ErrInvalidAddress: invalid Ethereum address
- ErrInvalidMarketConfig: protocol constraint violation
- ErrMarketAlreadyExists: market already exists
*/
func (h *AdminHandler) CreateMarket(c *gin.Context) {
	ctx := c.Request.Context()
	ctx = c.Request.Context()

	/*
		@logic ParseRequest

		@desc
		Parse JSON request body into CreateMarketRequest DTO.
		Gin automatically validates struct tags (eth_addr, numeric bounds).

		@notes
		- Validates format but not business logic
		- Business validation happens in service layer
	*/
	var req request.CreateMarketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": fmt.Sprintf("invalid request format: %v", err),
			},
		})
		return
	}

	/*
		@logic ExtractBorrower

		@desc
		Extract borrower wallet address from authentication context.
		Borrower MUST come from SIWE auth middleware (wallet verification).
		NOT from request body for security reasons.

		@notes
		- Middleware should set "wallet_address" in context
		- We trust middleware to validate wallet signature
	*/
	borrowerInterface, exists := c.Get("wallet_address")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "wallet address not found in context",
			},
		})
		return
	}
	borrower := borrowerInterface.(string)

	/*
		@logic ValidateRequest

		@desc
		Call service layer validation to enforce protocol constraints.
		This is the business logic validation layer that ensures protocol safety.
	*/
	if err := h.marketValidator.ValidateMarketCreation(
		borrower,
		req.BorrowAsset,
		req.CollateralAsset,
		int32(req.MinCollateralRatio),
		int32(req.LiquidationThreshold),
	); err != nil {
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
		@logic GenerateMarketAddress

		@desc
		Generate deterministic market address for new market.
		In production: derive from factory contract or use Ethereum address generation.
		In this stub: simple hash-based generation for testing.

		@notes
		- Market address should be deterministic (same inputs = same address)
		- Should avoid collisions (use factory contract in production)
	*/
	marketAddress := fmt.Sprintf("0xmarket%s%s", borrower[:8], req.BorrowAsset[:8])

	/*
		@logic CallServiceLayer

		@desc
		Call market service to create market with validated parameters.
		Service layer handles all business logic and persistence.

		@params
		- ctx: request context
		- marketAddress: deterministic market address
		- borrower: from auth context (trusted)
		- borrowAsset: from request
		- collateralAsset: from request
		- collateralOracle: from request (Chainlink oracle)
		- minCollateralRatio: from request
		- liquidationThreshold: from request
	*/
	market, err := h.marketService.CreateMarket(
		ctx,
		marketAddress,
		borrower,
		req.BorrowAsset,
		req.CollateralAsset,
		req.CollateralOracle,
		int32(req.MinCollateralRatio),
		int32(req.LiquidationThreshold),
	)
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
		@logic MapResponseDTO

		@desc
		Convert domain model to response DTO.
		This ensures API contract remains stable even if domain models change.
	*/
	resp := response.MarketResponse{
		Address:         market.Address,
		Borrower:        market.Borrower,
		BorrowAsset:     response.TokenInfo{Address: market.BorrowAsset},
		CollateralAsset: response.TokenInfo{Address: market.CollateralAsset},
		TotalPrincipal:  market.TotalPrincipal.String(),
		TotalLiquidity:  market.TotalLiquidity.String(),
		TotalDebt:       market.TotalDebt.String(),
		UtilizationRate: market.UtilizationRate,
		WeightedAPR:     market.WeightedAvgAPR,
		IsActive:        market.IsActive,
		IsLiquidating:   market.IsLiquidating,
		IsClosed:        market.IsClosed,
		CreatedAt:       market.CreatedAt.Unix(),
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    resp,
	})
}

/*
@handler GetMarket

@desc
HTTP GET handler for retrieving market details by address.
Returns market configuration and current state metrics.

@route
GET /api/admin/markets/:address

@params
- address: market contract address (path parameter, 42-char hex string)

@response
Success (200 OK):

	{
	  "success": true,
	  "data": {
	    "market_address": "0x...",
	    "borrower": "0x...",
	    "borrow_asset": "0x...",
	    "collateral_asset": "0x...",
	    "min_collateral_ratio": 11000,
	    "liquidation_threshold": 9500,
	    "total_principal": "1000000000000000000",
	    "total_interest": "50000000000000000",
	    "is_active": true,
	    "created_at": 1685356800
	  }
	}

Error (404):

	{
	  "success": false,
	  "error": {
	    "code": "MARKET_NOT_FOUND",
	    "message": "market not found"
	  }
	}

@error_cases
- ErrMarketNotFound: market address not found in database
- ErrInvalidAddress: invalid market address format
*/
func (h *AdminHandler) GetMarket(c *gin.Context) {
	ctx := c.Request.Context()
	ctx = c.Request.Context()

	/*
		@logic ExtractPathParam

		@desc
		Extract market address from URL path parameter.
		Path parameters are automatically URL-decoded by Gin.
	*/
	marketAddress := c.Param("address")

	/*
		@logic ValidateAddressFormat

		@desc
		Quick format validation of market address.
		Full validation happens in service layer.
	*/
	if marketAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "market address is required",
			},
		})
		return
	}

	/*
		@logic CallServiceLayer

		@desc
		Call market service to fetch market details.
		Service applies business logic validation.
	*/
	mkt, err := h.marketService.GetMarket(ctx, marketAddress)
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
		@logic MapResponseDTO

		@desc
		Map domain model to response DTO, including computed metrics.
	*/
	resp := response.MarketResponse{
		Address:         mkt.Address,
		Borrower:        mkt.Borrower,
		BorrowAsset:     response.TokenInfo{Address: mkt.BorrowAsset},
		CollateralAsset: response.TokenInfo{Address: mkt.CollateralAsset},
		TotalPrincipal:  mkt.TotalPrincipal.String(),
		TotalLiquidity:  mkt.TotalLiquidity.String(),
		TotalDebt:       mkt.TotalDebt.String(),
		UtilizationRate: mkt.UtilizationRate,
		WeightedAPR:     mkt.WeightedAvgAPR,
		IsActive:        mkt.IsActive,
		IsLiquidating:   mkt.IsLiquidating,
		IsClosed:        mkt.IsClosed,
		CreatedAt:       mkt.CreatedAt.Unix(),
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
	})
}

/*
@handler ListMarkets

@desc
HTTP GET handler for listing all active markets with pagination.
Returns paginated list of markets with optional filtering.

@route
GET /api/admin/markets?page=1&limit=20&active=true

@query_params
- page: page number (default: 1, minimum: 1)
- limit: results per page (default: 20, max: 100)
- active: filter by active status (optional, true/false)

@response
Success (200 OK):

	{
	  "success": true,
	  "data": {
	    "markets": [
	      {
	        "market_address": "0x...",
	        "borrower": "0x...",
	        "borrow_asset": "0x...",
	        "collateral_asset": "0x...",
	        "is_active": true,
	        "created_at": 1685356800
	      }
	    ],
	    "pagination": {
	      "page": 1,
	      "limit": 20,
	      "total": 45
	    }
	  }
	}

@error_cases
- Invalid pagination parameters
- Database connection errors
*/
func (h *AdminHandler) ListMarkets(c *gin.Context) {
	ctx := c.Request.Context()
	ctx = c.Request.Context()

	/*
		@logic ExtractQueryParams

		@desc
		Extract and validate pagination query parameters from request.
		Query parameters are strings and must be converted to integers.
	*/
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	/*
		@logic ParseAndValidatePagination

		@desc
		Parse pagination parameters with validation and bounds checking.
		Prevents invalid queries from reaching database layer.
	*/
	page, err := strconv.ParseInt(pageStr, 10, 32)
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
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
				"code":    "INVALID_REQUEST",
				"message": "limit must be between 1 and 100",
			},
		})
		return
	}

	/*
		@logic CalculateOffset

		@desc
		Calculate database offset from page number.
		Offset = (page - 1) * limit
	*/
	offset := (page - 1) * limit

	/*
		@logic CallServiceLayer

		@desc
		Call market service with pagination parameters.
		Service layer handles database query and result mapping.
	*/
	markets, err := h.marketService.ListMarkets(
		ctx,
		int32(limit),
		int32(offset),
		"",
	)
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
		@logic MapResponses

		@desc
		Map domain models to response DTOs for API contract.
	*/
	var resp []response.MarketResponse
	for _, mkt := range markets {
		resp = append(resp, response.MarketResponse{
			Address:         mkt.Address,
			Borrower:        mkt.Borrower,
			BorrowAsset:     response.TokenInfo{Address: mkt.BorrowAsset},
			CollateralAsset: response.TokenInfo{Address: mkt.CollateralAsset},
			TotalPrincipal:  mkt.TotalPrincipal.String(),
			TotalLiquidity:  mkt.TotalLiquidity.String(),
			TotalDebt:       mkt.TotalDebt.String(),
			UtilizationRate: mkt.UtilizationRate,
			WeightedAPR:     mkt.WeightedAvgAPR,
			IsActive:        mkt.IsActive,
			IsLiquidating:   mkt.IsLiquidating,
			IsClosed:        mkt.IsClosed,
			CreatedAt:       mkt.CreatedAt.Unix(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"markets": resp,
			"pagination": gin.H{
				"page":  page,
				"limit": limit,
			},
		},
	})
}

/*
@handler GetMarketMetrics

@desc
HTTP GET handler for retrieving market metrics and analytics.
Returns utilization, APR, and other protocol metrics.

@route
GET /api/admin/markets/:address/metrics

@params
- address: market contract address

@response
Success (200 OK):

	{
	  "success": true,
	  "data": {
	    "market_address": "0x...",
	    "utilization_rate": 0.65,
	    "average_apr": 500,
	    "total_value_locked": "5000000000000000000",
	    "total_interest": "150000000000000000",
	    "active_positions": 42,
	    "active_offers": 12,
	    "status": "healthy"
	  }
	}

@error_cases
- ErrMarketNotFound: market not found
- Calculation errors in metrics
*/
func (h *AdminHandler) GetMarketMetrics(c *gin.Context) {
	ctx := c.Request.Context()
	ctx = c.Request.Context()

	/*
		@logic ExtractPathParam

		@desc
		Extract market address from URL path.
	*/
	marketAddress := c.Param("address")

	if marketAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "market address is required",
			},
		})
		return
	}

	/*
		@logic GetMarketData

		@desc
		Fetch market from service first.
		CalculateMetrics requires *models.Market, not just address string.
	*/
	market, err := h.marketService.GetMarket(ctx, marketAddress)
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
		@logic CallServiceLayer

		@desc
		Call market service to calculate metrics.
		Service performs all calculations and validation.
	*/
	metrics, err := h.marketService.CalculateMetrics(ctx, market)
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
		@logic MapResponseDTO

		@desc
		Convert metrics to response format with proper data types.
		Extract relevant metrics from the metrics map returned by service.
	*/
	activePositions := int32(0)
	if ap, ok := metrics["active_positions"]; ok {
		switch v := ap.(type) {
		case int:
			activePositions = int32(v)
		case int32:
			activePositions = v
		case int64:
			activePositions = int32(v)
		}
	}
	tvl := ""
	if t, ok := metrics["tvl"]; ok {
		tvl, _ = t.(string)
	}
	resp := response.MarketMetricsResponse{
		Address:            marketAddress,
		TVLHistory:         []response.DataPoint{},
		APRHistory:         []response.DataPoint{},
		UtilizationHistory: []response.DataPoint{},
		ActivePositions:    activePositions,
		TotalVolumeTraded:  tvl,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
	})
}

/*
@handler UpdateMarketStatus

@desc
HTTP PATCH handler for updating market operational status.
Allows enabling/disabling market and updating parameters.

@route
PATCH /api/admin/markets/:address/status

@params
- address: market contract address

@request

	{
	  "is_active": true,
	  "is_liquidating": false
	}

@response
Success (200 OK):

	{
	  "success": true,
	  "data": {
	    "market_address": "0x...",
	    "is_active": true,
	    "is_liquidating": false,
	    "updated_at": 1685356800
	  }
	}

@error_cases
- ErrMarketNotFound: market not found
- ErrUnauthorized: insufficient permissions
*/
func (h *AdminHandler) UpdateMarketStatus(c *gin.Context) {
	ctx := c.Request.Context()
	ctx = c.Request.Context()

	/*
		@logic ExtractPathParam

		@desc
		Extract market address from URL path.
	*/
	marketAddress := c.Param("address")

	if marketAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "market address is required",
			},
		})
		return
	}

	/*
		@logic ParseRequestBody

		@desc
		Parse status update request from JSON body.
	*/
	var req struct {
		IsActive      *bool `json:"is_active"`
		IsLiquidating *bool `json:"is_liquidating"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": fmt.Sprintf("invalid request format: %v", err),
			},
		})
		return
	}

	/*
		@logic FetchCurrentMarket

		@desc
		Get current market state to apply updates.
		Only update provided fields (patch semantics).
	*/
	mkt, err := h.marketService.GetMarket(ctx, marketAddress)
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
		@logic ApplyUpdates

		@desc
		Apply partial updates to market model.
		Only modify fields that were provided in request (nil checks).
	*/
	if req.IsActive != nil {
		mkt.IsActive = *req.IsActive
	}
	if req.IsLiquidating != nil {
		mkt.IsLiquidating = *req.IsLiquidating
	}

	/*
		@logic PersistChanges

		@desc
		Persist updated market to database.
		This is a simplified example - real implementation would use repository.
	*/
	// In a real implementation, we would call repository to update
	// For now, we'll return success with updated values

	resp := gin.H{
		"market_address": mkt.Address,
		"is_active":      mkt.IsActive,
		"is_liquidating": mkt.IsLiquidating,
		"updated_at":     mkt.LastInterestAccrual.Unix(),
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
	})
}

/*
@handler HealthCheck

@desc
HTTP GET handler for protocol and market health status.
Used for monitoring and operational status checks.

@route
GET /api/admin/health

@response
Success (200 OK):

	{
	  "success": true,
	  "data": {
	    "status": "healthy",
	    "total_markets": 15,
	    "total_tvl": "10000000000000000000",
	    "active_positions": 250,
	    "timestamp": 1685356800
	  }
	}

@error_cases
- Database connection issues
- Calculation failures
*/
func (h *AdminHandler) HealthCheck(c *gin.Context) {
	/*
		@logic QuerySystemMetrics

		@desc
		Gather high-level system metrics for health status.
		This is a simplified example - real implementation would aggregate data.
	*/
	// In a real implementation, we would call repository methods to fetch:
	// - Total market count
	// - Total TVL across all markets
	// - Active positions count
	// - System status indicators

	resp := gin.H{
		"status":           "healthy",
		"total_markets":    0,
		"total_tvl":        big.NewInt(0).String(),
		"active_positions": 0,
		"timestamp":        c.GetInt64("request_time"),
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
	})
}
