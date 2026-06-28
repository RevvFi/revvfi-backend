package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Revvfi/revvfi-backend/internal/api/dto/request"
	"github.com/Revvfi/revvfi-backend/internal/api/dto/response"
)

/*
@file admin_borrowers_test.go

@desc
Unit tests for borrower management handlers.
Tests all borrower administration endpoints with comprehensive coverage.

@test_structure
- Table-driven test patterns
- Mock service isolation
- Error path testing
- Response validation
*/

/*
@test mockAdminBorrowerService

@desc
Mock implementation of AdminBorrowerService for testing.
Allows controlled behavior for testing handlers.
*/
type mockAdminBorrowerService struct {
	listBorrowersFn           func(ctx interface{}, page, limit int32, status, search string) (*response.BorrowerListResponse, error)
	getBorrowerFn             func(ctx interface{}, borrowerAddress string) (*response.BorrowerInfo, error)
	getPendingBorrowersFn     func(ctx interface{}) (*response.PendingBorrowerResponse, error)
	prepareBorrowerAdditionFn func(ctx interface{}, borrowerAddress string, isVerified bool) (string, error)
	prepareBorrowerRemovalFn  func(ctx interface{}, borrowerAddress, reason string, closePositions bool) (string, error)
}

/*
@function ListBorrowers (mock)
@desc Mock implementation of ListBorrowers method.
*/
func (m *mockAdminBorrowerService) ListBorrowers(ctx interface{}, page, limit int32, status, search string) (*response.BorrowerListResponse, error) {
	if m.listBorrowersFn != nil {
		return m.listBorrowersFn(ctx, page, limit, status, search)
	}
	return nil, fmt.Errorf("not implemented")
}

/*
@function GetBorrower (mock)
@desc Mock implementation of GetBorrower method.
*/
func (m *mockAdminBorrowerService) GetBorrower(ctx interface{}, borrowerAddress string) (*response.BorrowerInfo, error) {
	if m.getBorrowerFn != nil {
		return m.getBorrowerFn(ctx, borrowerAddress)
	}
	return nil, fmt.Errorf("not implemented")
}

/*
@function GetPendingBorrowers (mock)
@desc Mock implementation of GetPendingBorrowers method.
*/
func (m *mockAdminBorrowerService) GetPendingBorrowers(ctx interface{}) (*response.PendingBorrowerResponse, error) {
	if m.getPendingBorrowersFn != nil {
		return m.getPendingBorrowersFn(ctx)
	}
	return nil, fmt.Errorf("not implemented")
}

/*
@function PrepareBorrowerAddition (mock)
@desc Mock implementation of PrepareBorrowerAddition method.
*/
func (m *mockAdminBorrowerService) PrepareBorrowerAddition(ctx interface{}, borrowerAddress string, isVerified bool) (string, error) {
	if m.prepareBorrowerAdditionFn != nil {
		return m.prepareBorrowerAdditionFn(ctx, borrowerAddress, isVerified)
	}
	return "", fmt.Errorf("not implemented")
}

/*
@function PrepareBorrowerRemoval (mock)
@desc Mock implementation of PrepareBorrowerRemoval method.
*/
func (m *mockAdminBorrowerService) PrepareBorrowerRemoval(ctx interface{}, borrowerAddress, reason string, closePositions bool) (string, error) {
	if m.prepareBorrowerRemovalFn != nil {
		return m.prepareBorrowerRemovalFn(ctx, borrowerAddress, reason, closePositions)
	}
	return "", fmt.Errorf("not implemented")
}

/*
@test TestListBorrowers_Success

@desc
Test successful borrower listing with pagination.

@test_cases
- Valid page and limit parameters
- Returns paginated list
- Includes pagination metadata
*/
func TestListBorrowers_Success(t *testing.T) {
	mockSvc := &mockAdminBorrowerService{
		listBorrowersFn: func(ctx interface{}, page, limit int32, status, search string) (*response.BorrowerListResponse, error) {
			return &response.BorrowerListResponse{
				Borrowers: []response.BorrowerInfo{
					{
						Address:         "0x1234567890123456789012345678901234567890",
						IsVerified:      true,
						ReputationScore: 850,
						TotalBorrowed:   "5000000000000000000",
						TotalRepaid:     "4500000000000000000",
						ActiveOffers:    2,
						DefaultCount:    0,
						AddedAt:         1685356800,
						VerifiedAt:      1685360000,
					},
				},
				Pagination: &response.PaginationInfo{
					Page:  1,
					Limit: 20,
					Total: 150,
				},
				TotalReputationAverage: 820.5,
			}, nil
		},
	}

	handler := NewAdminBorrowerHandler(mockSvc)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/admin/borrowers?page=1&limit=20", nil)

	handler.ListBorrowers(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var respBody map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &respBody)
	require.NoError(t, err)

	assert.True(t, respBody["success"].(bool))
	data := respBody["data"].(map[string]interface{})
	assert.NotNil(t, data["borrowers"])
	assert.NotNil(t, data["pagination"])
}

/*
@test TestListBorrowers_InvalidPagination

@desc
Test borrower listing with invalid pagination parameters.

@error_case
- Invalid page/limit values → 400 Bad Request
*/
func TestListBorrowers_InvalidPagination(t *testing.T) {
	mockSvc := &mockAdminBorrowerService{}
	handler := NewAdminBorrowerHandler(mockSvc)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/admin/borrowers?page=0&limit=20", nil)

	handler.ListBorrowers(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var respBody map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &respBody)
	assert.False(t, respBody["success"].(bool))
}

/*
@test TestListBorrowers_InvalidStatusFilter

@desc
Test borrower listing with invalid status filter.

@error_case
- Invalid status value → 400 Bad Request
*/
func TestListBorrowers_InvalidStatusFilter(t *testing.T) {
	mockSvc := &mockAdminBorrowerService{}
	handler := NewAdminBorrowerHandler(mockSvc)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/admin/borrowers?page=1&limit=20&status=invalid", nil)

	handler.ListBorrowers(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var respBody map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &respBody)
	assert.False(t, respBody["success"].(bool))
	assert.Equal(t, "INVALID_STATUS_FILTER", respBody["error"].(map[string]interface{})["code"])
}

/*
@test TestGetBorrower_Success

@desc
Test successful retrieval of borrower details.

@test_cases
- Valid borrower address
- Returns borrower information with metrics
*/
func TestGetBorrower_Success(t *testing.T) {
	mockSvc := &mockAdminBorrowerService{
		getBorrowerFn: func(ctx interface{}, borrowerAddress string) (*response.BorrowerInfo, error) {
			return &response.BorrowerInfo{
				Address:         borrowerAddress,
				IsVerified:      true,
				ReputationScore: 850,
				TotalBorrowed:   "5000000000000000000",
				TotalRepaid:     "4500000000000000000",
				ActiveOffers:    2,
				DefaultCount:    0,
				AddedAt:         1685356800,
				VerifiedAt:      1685360000,
			}, nil
		},
	}

	handler := NewAdminBorrowerHandler(mockSvc)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/admin/borrowers/0x1234567890123456789012345678901234567890", nil)
	c.Params = []gin.Param{{Key: "address", Value: "0x1234567890123456789012345678901234567890"}}

	handler.GetBorrower(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var respBody map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &respBody)
	assert.True(t, respBody["success"].(bool))
}

/*
@test TestGetBorrower_MissingAddress

@desc
Test borrower retrieval with missing address parameter.

@error_case
- No address in path → 400 Bad Request
*/
func TestGetBorrower_MissingAddress(t *testing.T) {
	mockSvc := &mockAdminBorrowerService{}
	handler := NewAdminBorrowerHandler(mockSvc)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/admin/borrowers/", nil)
	c.Params = []gin.Param{}

	handler.GetBorrower(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

/*
@test TestGetPendingBorrowers_Success

@desc
Test successful retrieval of pending borrowers.

@test_cases
- Returns list of borrowers awaiting verification
- Includes pending count and oldest pending timestamp
*/
func TestGetPendingBorrowers_Success(t *testing.T) {
	mockSvc := &mockAdminBorrowerService{
		getPendingBorrowersFn: func(ctx interface{}) (*response.PendingBorrowerResponse, error) {
			return &response.PendingBorrowerResponse{
				PendingCount:    5,
				OldestPendingAt: 1685356800,
				Borrowers: []response.BorrowerInfo{
					{
						Address:    "0x1234567890123456789012345678901234567890",
						IsVerified: false,
						AddedAt:    1685356800,
					},
				},
			}, nil
		},
	}

	handler := NewAdminBorrowerHandler(mockSvc)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/admin/borrowers/pending", nil)

	handler.GetPendingBorrowers(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var respBody map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &respBody)
	assert.True(t, respBody["success"].(bool))
	data := respBody["data"].(map[string]interface{})
	assert.Equal(t, float64(5), data["pending_count"])
}

/*
@test TestPrepareBorrowerAddition_Success

@desc
Test successful borrower addition preparation.

@test_cases
- Valid borrower address and request
- Returns encoded transaction data
*/
func TestPrepareBorrowerAddition_Success(t *testing.T) {
	mockSvc := &mockAdminBorrowerService{
		prepareBorrowerAdditionFn: func(ctx interface{}, borrowerAddress string, isVerified bool) (string, error) {
			return "0xencoded_transaction_data", nil
		},
	}

	handler := NewAdminBorrowerHandler(mockSvc)

	req := struct {
		IsVerified bool `json:"is_verified"`
	}{IsVerified: true}

	reqBody, _ := json.Marshal(req)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/v1/admin/borrowers/0x1234567890123456789012345678901234567890/prepare", bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = []gin.Param{{Key: "address", Value: "0x1234567890123456789012345678901234567890"}}

	handler.PrepareBorrowerAddition(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var respBody map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &respBody)
	assert.True(t, respBody["success"].(bool))
	assert.NotNil(t, respBody["data"])
}

/*
@test TestPrepareBorrowerAddition_InvalidRequest

@desc
Test borrower addition with invalid request body.

@error_case
- Malformed JSON → 400 Bad Request
*/
func TestPrepareBorrowerAddition_InvalidRequest(t *testing.T) {
	mockSvc := &mockAdminBorrowerService{}
	handler := NewAdminBorrowerHandler(mockSvc)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/v1/admin/borrowers/0x1234567890123456789012345678901234567890/prepare", bytes.NewBuffer([]byte("invalid")))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = []gin.Param{{Key: "address", Value: "0x1234567890123456789012345678901234567890"}}

	handler.PrepareBorrowerAddition(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

/*
@test TestPrepareBorrowerRemoval_Success

@desc
Test successful borrower removal preparation.

@test_cases
- Valid borrower address and removal reason
- Returns encoded transaction data
*/
func TestPrepareBorrowerRemoval_Success(t *testing.T) {
	mockSvc := &mockAdminBorrowerService{
		prepareBorrowerRemovalFn: func(ctx interface{}, borrowerAddress, reason string, closePositions bool) (string, error) {
			return "0xencoded_removal_tx", nil
		},
	}

	handler := NewAdminBorrowerHandler(mockSvc)

	req := request.RemoveBorrowerRequest{
		Reason:                 "Violation of terms",
		CloseExistingPositions: true,
	}

	reqBody, _ := json.Marshal(req)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("DELETE", "/api/v1/admin/borrowers/0x1234567890123456789012345678901234567890/prepare", bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = []gin.Param{{Key: "address", Value: "0x1234567890123456789012345678901234567890"}}

	handler.PrepareBorrowerRemoval(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var respBody map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &respBody)
	assert.True(t, respBody["success"].(bool))
}

/*
@test TestPrepareBorrowerRemoval_MissingReason

@desc
Test borrower removal with missing reason.

@error_case
- Missing required reason field → 400 Bad Request
*/
func TestPrepareBorrowerRemoval_MissingReason(t *testing.T) {
	mockSvc := &mockAdminBorrowerService{}
	handler := NewAdminBorrowerHandler(mockSvc)

	req := request.RemoveBorrowerRequest{
		Reason:                 "", // Empty reason
		CloseExistingPositions: false,
	}

	reqBody, _ := json.Marshal(req)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("DELETE", "/api/v1/admin/borrowers/0x1234567890123456789012345678901234567890/prepare", bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = []gin.Param{{Key: "address", Value: "0x1234567890123456789012345678901234567890"}}

	handler.PrepareBorrowerRemoval(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var respBody map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &respBody)
	assert.False(t, respBody["success"].(bool))
	assert.Equal(t, "MISSING_FIELD", respBody["error"].(map[string]interface{})["code"])
}
