package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Revvfi/revvfi-backend/internal/api/dto/request"
	"github.com/Revvfi/revvfi-backend/internal/models"
	appErr "github.com/Revvfi/revvfi-backend/internal/pkg/errors"
)

/*
@file admin_test.go

@desc
Unit tests for admin HTTP handlers.
Tests all CRUD operations, error handling, and response mapping.

@test_structure
- Table-driven test patterns for multiple scenarios
- Mock services for isolation
- Test error cases and edge conditions
- Verify response DTOs and status codes
- Test input validation
*/

/*
@test mockMarketService

@desc
Mock implementation of MarketService interface for testing.
*/
type mockMarketService struct {
	createMarketFn  func(ctx context.Context, address string, borrower, borrowAsset, collateralAsset, collateralOracle string, minRatio, liquidationThreshold int32) (*models.Market, error)
	getMarketFn     func(ctx context.Context, address string) (*models.Market, error)
	listMarketsFn   func(ctx context.Context, limit, offset int32) ([]models.Market, error)
	calculateMetricsFn func(ctx context.Context, market *models.Market) (map[string]interface{}, error)
}

/*
@function CreateMarket (mock)

@desc
Mock implementation of CreateMarket service method.
*/
func (m *mockMarketService) CreateMarket(ctx context.Context, address string, borrower, borrowAsset, collateralAsset, collateralOracle string, minRatio, liquidationThreshold int32) (*models.Market, error) {
	if m.createMarketFn != nil {
		return m.createMarketFn(ctx, address, borrower, borrowAsset, collateralAsset, collateralOracle, minRatio, liquidationThreshold)
	}
	return nil, fmt.Errorf("not implemented")
}

/*
@function GetMarket (mock)

@desc
Mock implementation of GetMarket service method.
*/
func (m *mockMarketService) GetMarket(ctx context.Context, address string) (*models.Market, error) {
	if m.getMarketFn != nil {
		return m.getMarketFn(ctx, address)
	}
	return nil, fmt.Errorf("not implemented")
}

/*
@function ListMarkets (mock)

@desc
Mock implementation of ListMarkets service method.
*/
func (m *mockMarketService) ListMarkets(ctx context.Context, limit, offset int32) ([]models.Market, error) {
	if m.listMarketsFn != nil {
		return m.listMarketsFn(ctx, limit, offset)
	}
	return nil, fmt.Errorf("not implemented")
}

/*
@function CalculateMetrics (mock)

@desc
Mock implementation of CalculateMetrics service method.
*/
func (m *mockMarketService) CalculateMetrics(ctx context.Context, market *models.Market) (map[string]interface{}, error) {
	if m.calculateMetricsFn != nil {
		return m.calculateMetricsFn(ctx, market)
	}
	return nil, fmt.Errorf("not implemented")
}

/*
@test mockMarketValidator

@desc
Mock implementation of MarketValidator interface for testing.
*/
type mockMarketValidator struct {
	validateCreationFn func(borrower, borrowAsset, collateralAsset string, minRatio, liquidationThreshold int32) error
}

/*
@function ValidateMarketCreation (mock)

@desc
Mock implementation of validation method.
*/
func (m *mockMarketValidator) ValidateMarketCreation(borrower, borrowAsset, collateralAsset string, minRatio, liquidationThreshold int32) error {
	if m.validateCreationFn != nil {
		return m.validateCreationFn(borrower, borrowAsset, collateralAsset, minRatio, liquidationThreshold)
	}
	return nil
}

/*
@test TestCreateMarket_Success

@desc
Test successful market creation with valid parameters.

@test_cases
- Valid request parameters
- Borrower extracted from context
- Service returns valid market
- Response contains all required fields
*/
func TestCreateMarket_Success(t *testing.T) {
	/*
	@logic Setup
	@desc Initialize mock service and handler.
	*/
	mockSvc := &mockMarketService{
		createMarketFn: func(ctx context.Context, address string, borrower, borrowAsset, collateralAsset, collateralOracle string, minRatio, liquidationThreshold int32) (*models.Market, error) {
			return &models.Market{
				ID:                      1,
				Address:                 address,
				Borrower:                borrower,
				BorrowAsset:             borrowAsset,
				CollateralAsset:         collateralAsset,
				CollateralOracle:        collateralOracle,
				MinCollateralRatio:      minRatio,
				LiquidationThreshold:    liquidationThreshold,
				TotalPrincipal:          big.NewInt(0),
				TotalAccruedInterest:    big.NewInt(0),
				TotalDebt:               big.NewInt(0),
				TotalLiquidity:          big.NewInt(0),
				WeightedAvgAPR:          500,
				UtilizationRate:         0.0,
				IsActive:                true,
				IsLiquidating:           false,
				IsClosed:                false,
				CreatedAt:               time.Now(),
				LastInterestAccrual:     time.Now(),
			}, nil
		},
	}

	mockValidator := &mockMarketValidator{
		validateCreationFn: func(borrower, borrowAsset, collateralAsset string, minRatio, liquidationThreshold int32) error {
			return nil
		},
	}

	handler := &AdminHandler{
		marketService:   mockSvc,
		marketValidator: mockValidator,
	}

	/*
	@logic BuildRequest
	@desc Create valid CreateMarketRequest payload.
	*/
	req := request.CreateMarketRequest{
		BorrowAsset:             "0x6b175474e89094c44da98b954eedeac495271d0f",
		BorrowAssetDecimals:     6,
		CollateralAsset:         "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
		CollateralAssetDecimals: 18,
		CollateralOracle:        "0x5f4ec3df9cbd43714fe2740f5e3616155c5b8419",
		MinCollateralRatio:      15000,
		LiquidationThreshold:    12000,
	}

	reqBody, err := json.Marshal(req)
	require.NoError(t, err)

	/*
	@logic CreateContext
	@desc Create Gin test context with wallet_address set.
	*/
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/admin/markets", bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("wallet_address", "0x1234567890123456789012345678901234567890")

	/*
	@logic ExecuteHandler
	@desc Call handler with mocked context.
	*/
	handler.CreateMarket(c)

	/*
	@logic AssertResponse
	@desc Verify status code and response structure.
	*/
	assert.Equal(t, http.StatusCreated, w.Code)

	var respBody map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &respBody)
	require.NoError(t, err)

	assert.True(t, respBody["success"].(bool))
	assert.NotNil(t, respBody["data"])

	/*
	@logic AssertData
	@desc Verify market data in response.
	*/
	data := respBody["data"].(map[string]interface{})
	assert.Equal(t, "0x1234567890123456789012345678901234567890", data["borrower"])
	assert.Equal(t, true, data["is_active"])
	assert.NotEmpty(t, data["address"])
}

/*
@test TestCreateMarket_MissingWallet

@desc
Test market creation fails when wallet not in context.

@error_case
- Missing wallet_address in context → 401 Unauthorized

@notes
The wallet check happens BEFORE JSON parsing, so we need valid JSON.
*/
func TestCreateMarket_MissingWallet(t *testing.T) {
	/*
	@logic Setup
	@desc Create handler without wallet in context.
	*/
	handler := &AdminHandler{
		marketService:   &mockMarketService{},
		marketValidator: &mockMarketValidator{},
	}

	req := request.CreateMarketRequest{
		BorrowAsset:             "0x6b175474e89094c44da98b954eedeac495271d0f",
		BorrowAssetDecimals:     6,
		CollateralAsset:         "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
		CollateralAssetDecimals: 18,
		CollateralOracle:        "0x5f4ec3df9cbd43714fe2740f5e3616155c5b8419",
		MinCollateralRatio:      15000,
		LiquidationThreshold:    12000,
	}

	reqBody, _ := json.Marshal(req)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/admin/markets", bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	// No wallet_address in context

	/*
	@logic ExecuteHandler
	@desc Call handler without wallet.
	*/
	handler.CreateMarket(c)

	/*
	@logic AssertError
	@desc Verify 401 response.
	*/
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var respBody map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &respBody)
	assert.False(t, respBody["success"].(bool))
	assert.Equal(t, "UNAUTHORIZED", respBody["error"].(map[string]interface{})["code"])
}

/*
@test TestCreateMarket_InvalidRequest

@desc
Test market creation with malformed JSON.

@error_case
- Invalid JSON → 400 Bad Request
*/
func TestCreateMarket_InvalidRequest(t *testing.T) {
	/*
	@logic Setup
	*/
	handler := &AdminHandler{
		marketService:   &mockMarketService{},
		marketValidator: &mockMarketValidator{},
	}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/admin/markets", bytes.NewBuffer([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("wallet_address", "0x1234567890123456789012345678901234567890")

	/*
	@logic ExecuteHandler
	@desc Call with malformed payload.
	*/
	handler.CreateMarket(c)

	/*
	@logic AssertError
	@desc Verify 400 response.
	*/
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var respBody map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &respBody)
	assert.False(t, respBody["success"].(bool))
	assert.Equal(t, "INVALID_REQUEST", respBody["error"].(map[string]interface{})["code"])
}

/*
@test TestCreateMarket_ValidationError

@desc
Test market creation with business logic validation failure.

@error_case
- Min collateral ratio < 100% → 422 Unprocessable Entity
*/
func TestCreateMarket_ValidationError(t *testing.T) {
	/*
	@logic Setup
	@desc Mock validator to return error.
	*/
	mockValidator := &mockMarketValidator{
		validateCreationFn: func(borrower, borrowAsset, collateralAsset string, minRatio, liquidationThreshold int32) error {
			return appErr.ErrInvalidInput // Use an error in the ErrorCode switch
		},
	}

	handler := &AdminHandler{
		marketService:   &mockMarketService{},
		marketValidator: mockValidator,
	}

	req := request.CreateMarketRequest{
		BorrowAsset:             "0x6b175474e89094c44da98b954eedeac495271d0f",
		BorrowAssetDecimals:     6,
		CollateralAsset:         "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
		CollateralAssetDecimals: 18,
		CollateralOracle:        "0x5f4ec3df9cbd43714fe2740f5e3616155c5b8419",
		MinCollateralRatio:      11000, // Valid (110%) - but validation logic will fail
		LiquidationThreshold:    9500,  // Valid (95%)
	}

	reqBody, _ := json.Marshal(req)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/admin/markets", bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("wallet_address", "0x1234567890123456789012345678901234567890")

	/*
	@logic ExecuteHandler
	@desc Call with invalid parameters.
	*/
	handler.CreateMarket(c)

	/*
	@logic AssertError
	@desc Verify 400 response with validation error (ErrInvalidInput maps to 400).
	*/
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var respBody map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &respBody)
	assert.False(t, respBody["success"].(bool))
	errorCode := respBody["error"].(map[string]interface{})["code"]
	assert.NotEmpty(t, errorCode)
}

/*
@test TestGetMarket_Success

@desc
Test successful market retrieval by address.

@test_cases
- Market exists
- Return complete market data
- All fields properly mapped
*/
func TestGetMarket_Success(t *testing.T) {
	/*
	@logic Setup
	@desc Mock service to return market.
	*/
	mockSvc := &mockMarketService{
		getMarketFn: func(ctx context.Context, address string) (*models.Market, error) {
			return &models.Market{
				ID:                      1,
				Address:                 address,
				Borrower:                "0x1234567890123456789012345678901234567890",
				BorrowAsset:             "0x6b175474e89094c44da98b954eedeac495271d0f",
				CollateralAsset:         "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
				TotalPrincipal:          big.NewInt(1000000),
				TotalLiquidity:          big.NewInt(2000000),
				TotalDebt:               big.NewInt(500000),
				UtilizationRate:         0.25,
				WeightedAvgAPR:          500,
				IsActive:                true,
				CreatedAt:               time.Now(),
			}, nil
		},
	}

	handler := &AdminHandler{
		marketService: mockSvc,
	}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/admin/markets/0xmarket", nil)
	c.Params = []gin.Param{{Key: "address", Value: "0xmarket"}}

	/*
	@logic ExecuteHandler
	*/
	handler.GetMarket(c)

	/*
	@logic AssertSuccess
	@desc Verify 200 response with market data.
	*/
	assert.Equal(t, http.StatusOK, w.Code)

	var respBody map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &respBody)
	assert.True(t, respBody["success"].(bool))
	
	data := respBody["data"].(map[string]interface{})
	assert.Equal(t, "0xmarket", data["address"])
	assert.Equal(t, true, data["is_active"])
}

/*
@test TestGetMarket_NotFound

@desc
Test market retrieval when market doesn't exist.

@error_case
- Market not found → 404 Not Found
*/
func TestGetMarket_NotFound(t *testing.T) {
	/*
	@logic Setup
	@desc Mock service to return not found error.
	*/
	mockSvc := &mockMarketService{
		getMarketFn: func(ctx context.Context, address string) (*models.Market, error) {
			return nil, appErr.ErrMarketNotFound
		},
	}

	handler := &AdminHandler{
		marketService: mockSvc,
	}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/admin/markets/0xnonexistent", nil)
	c.Params = []gin.Param{{Key: "address", Value: "0xnonexistent"}}

	/*
	@logic ExecuteHandler
	*/
	handler.GetMarket(c)

	/*
	@logic AssertError
	@desc Verify 404 response with market not found error code.
	*/
	assert.Equal(t, http.StatusNotFound, w.Code)

	var respBody map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &respBody)
	assert.False(t, respBody["success"].(bool))
	assert.Equal(t, "MARKET_NOT_FOUND", respBody["error"].(map[string]interface{})["code"])
}

/*
@test TestListMarkets_Success

@desc
Test successful market list retrieval with pagination.

@test_cases
- Multiple markets returned
- Pagination parameters working
- All markets properly mapped
*/
func TestListMarkets_Success(t *testing.T) {
	/*
	@logic Setup
	@desc Mock service to return list of markets.
	*/
	mockSvc := &mockMarketService{
		listMarketsFn: func(ctx context.Context, limit, offset int32) ([]models.Market, error) {
			return []models.Market{
				{
					Address:    "0xmarket1",
					Borrower:   "0x1111111111111111111111111111111111111111",
					IsActive:   true,
					CreatedAt:  time.Now(),
					TotalPrincipal: big.NewInt(1000000),
					TotalLiquidity: big.NewInt(2000000),
					TotalDebt:      big.NewInt(500000),
				},
				{
					Address:    "0xmarket2",
					Borrower:   "0x2222222222222222222222222222222222222222",
					IsActive:   true,
					CreatedAt:  time.Now(),
					TotalPrincipal: big.NewInt(2000000),
					TotalLiquidity: big.NewInt(4000000),
					TotalDebt:      big.NewInt(1000000),
				},
			}, nil
		},
	}

	handler := &AdminHandler{
		marketService: mockSvc,
	}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/admin/markets?page=1&limit=20", nil)

	/*
	@logic ExecuteHandler
	*/
	handler.ListMarkets(c)

	/*
	@logic AssertSuccess
	@desc Verify 200 response with market list.
	*/
	assert.Equal(t, http.StatusOK, w.Code)

	var respBody map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &respBody)
	assert.True(t, respBody["success"].(bool))

	data := respBody["data"].(map[string]interface{})
	markets := data["markets"].([]interface{})
	assert.Equal(t, 2, len(markets))
}

/*
@test TestGetMarketMetrics_Success

@desc
Test successful market metrics retrieval.

@test_cases
- Market exists
- Metrics calculated successfully
- Response contains all metric fields
*/
func TestGetMarketMetrics_Success(t *testing.T) {
	/*
	@logic Setup
	@desc Mock service to return metrics.
	*/
	mockSvc := &mockMarketService{
		getMarketFn: func(ctx context.Context, address string) (*models.Market, error) {
			return &models.Market{
				Address: address,
				IsActive: true,
			}, nil
		},
		calculateMetricsFn: func(ctx context.Context, market *models.Market) (map[string]interface{}, error) {
			return map[string]interface{}{
				"active_positions": int32(42),
				"tvl":              "1000000000000000000",
				"utilization_rate": 0.65,
				"average_apr":      500,
			}, nil
		},
	}

	handler := &AdminHandler{
		marketService: mockSvc,
	}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/admin/markets/0xmarket/metrics", nil)
	c.Params = []gin.Param{{Key: "address", Value: "0xmarket"}}

	/*
	@logic ExecuteHandler
	*/
	handler.GetMarketMetrics(c)

	/*
	@logic AssertSuccess
	@desc Verify 200 response with metrics.
	*/
	assert.Equal(t, http.StatusOK, w.Code)

	var respBody map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &respBody)
	assert.True(t, respBody["success"].(bool))

	data := respBody["data"].(map[string]interface{})
	assert.Equal(t, "0xmarket", data["address"])
	assert.NotNil(t, data["active_positions"])
}

/*
@test TestHealthCheck_Success

@desc
Test health check endpoint returns system status.

@test_cases
- Always returns 200 OK
- Status is "healthy"
- Includes timestamp
*/
func TestHealthCheck_Success(t *testing.T) {
	/*
	@logic Setup
	*/
	handler := &AdminHandler{
		marketService: &mockMarketService{},
	}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/admin/health", nil)

	/*
	@logic ExecuteHandler
	*/
	handler.HealthCheck(c)

	/*
	@logic AssertSuccess
	@desc Verify 200 response.
	*/
	assert.Equal(t, http.StatusOK, w.Code)

	var respBody map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &respBody)
	assert.True(t, respBody["success"].(bool))

	data := respBody["data"].(map[string]interface{})
	assert.Equal(t, "healthy", data["status"])
}
