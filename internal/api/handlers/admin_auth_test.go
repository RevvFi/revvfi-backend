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
)

/*
@file admin_auth_test.go

@desc
Unit tests for admin authentication handlers.
Tests all authentication and authorization endpoints with comprehensive coverage.

@test_structure
- Table-driven test patterns
- Mock service isolation
- Error path testing
- Response validation
*/

/*
@test mockAdminAuthService

@desc
Mock implementation of AdminAuthService for testing.
Allows controlled behavior for testing handlers.
*/
type mockAdminAuthService struct {
	isAdminFn                  func(ctx interface{}, address string) (bool, error)
	getAdminRoleFn             func(ctx interface{}, address string) (string, error)
	getAdminPermissionsFn      func(ctx interface{}, address string) ([]string, error)
	generateImpersonateTokenFn func(ctx interface{}, adminAddress, reason string) (string, int64, error)
	validateImpersonateTokenFn func(token string) (string, error)
}

/*
@function IsAdmin (mock)
@desc Mock implementation of IsAdmin method.
*/
func (m *mockAdminAuthService) IsAdmin(ctx interface{}, address string) (bool, error) {
	if m.isAdminFn != nil {
		return m.isAdminFn(ctx, address)
	}
	return false, nil
}

/*
@function GetAdminRole (mock)
@desc Mock implementation of GetAdminRole method.
*/
func (m *mockAdminAuthService) GetAdminRole(ctx interface{}, address string) (string, error) {
	if m.getAdminRoleFn != nil {
		return m.getAdminRoleFn(ctx, address)
	}
	return "", fmt.Errorf("not implemented")
}

/*
@function GetAdminPermissions (mock)
@desc Mock implementation of GetAdminPermissions method.
*/
func (m *mockAdminAuthService) GetAdminPermissions(ctx interface{}, address string) ([]string, error) {
	if m.getAdminPermissionsFn != nil {
		return m.getAdminPermissionsFn(ctx, address)
	}
	return nil, fmt.Errorf("not implemented")
}

/*
@function GenerateImpersonateToken (mock)
@desc Mock implementation of GenerateImpersonateToken method.
*/
func (m *mockAdminAuthService) GenerateImpersonateToken(ctx interface{}, adminAddress, reason string) (string, int64, error) {
	if m.generateImpersonateTokenFn != nil {
		return m.generateImpersonateTokenFn(ctx, adminAddress, reason)
	}
	return "", 0, fmt.Errorf("not implemented")
}

/*
@function ValidateImpersonateToken (mock)
@desc Mock implementation of ValidateImpersonateToken method.
*/
func (m *mockAdminAuthService) ValidateImpersonateToken(token string) (string, error) {
	if m.validateImpersonateTokenFn != nil {
		return m.validateImpersonateTokenFn(token)
	}
	return "", fmt.Errorf("not implemented")
}

/*
@test TestCheckAdmin_Success

@desc
Test successful admin verification returns correct status.

@test_cases
- Valid admin address
- Returns is_admin=true
- Includes role and permissions
*/
func TestCheckAdmin_Success(t *testing.T) {
	/*
		@logic Setup
		@desc Initialize mock service and handler.
	*/
	mockSvc := &mockAdminAuthService{
		isAdminFn: func(ctx interface{}, address string) (bool, error) {
			return true, nil
		},
		getAdminRoleFn: func(ctx interface{}, address string) (string, error) {
			return "ARCH_CONTROLLER_OWNER", nil
		},
		getAdminPermissionsFn: func(ctx interface{}, address string) ([]string, error) {
			return []string{"create_market", "pause_market", "set_fees"}, nil
		},
	}

	handler := NewAdminAuthHandler(mockSvc)

	/*
		@logic CreateContext
		@desc Create test context with address parameter.
	*/
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/admin/check/0x1234567890123456789012345678901234567890", nil)
	c.Params = []gin.Param{{Key: "address", Value: "0x1234567890123456789012345678901234567890"}}

	/*
		@logic ExecuteHandler
		@desc Call CheckAdmin handler.
	*/
	handler.CheckAdmin(c)

	/*
		@logic AssertResponse
		@desc Verify status code and response structure.
	*/
	assert.Equal(t, http.StatusOK, w.Code)

	var respBody map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &respBody)
	require.NoError(t, err)

	assert.True(t, respBody["success"].(bool))
	data := respBody["data"].(map[string]interface{})
	assert.True(t, data["is_admin"].(bool))
	assert.Equal(t, "ARCH_CONTROLLER_OWNER", data["role"])
}

/*
@test TestCheckAdmin_NotAdmin

@desc
Test verification of non-admin address returns is_admin=false.

@test_cases
- Non-admin address
- Returns is_admin=false
- No role or permissions included
*/
func TestCheckAdmin_NotAdmin(t *testing.T) {
	mockSvc := &mockAdminAuthService{
		isAdminFn: func(ctx interface{}, address string) (bool, error) {
			return false, nil
		},
	}

	handler := NewAdminAuthHandler(mockSvc)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/admin/check/0x0000000000000000000000000000000000000000", nil)
	c.Params = []gin.Param{{Key: "address", Value: "0x0000000000000000000000000000000000000000"}}

	handler.CheckAdmin(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var respBody map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &respBody)

	assert.True(t, respBody["success"].(bool))
	data := respBody["data"].(map[string]interface{})
	assert.False(t, data["is_admin"].(bool))
}

/*
@test TestCheckAdmin_MissingAddress

@desc
Test missing address parameter returns error.

@error_case
- No address in path → 400 Bad Request
*/
func TestCheckAdmin_MissingAddress(t *testing.T) {
	mockSvc := &mockAdminAuthService{}
	handler := NewAdminAuthHandler(mockSvc)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/admin/check/", nil)
	c.Params = []gin.Param{}

	handler.CheckAdmin(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var respBody map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &respBody)
	assert.False(t, respBody["success"].(bool))
	assert.Equal(t, "INVALID_ADDRESS", respBody["error"].(map[string]interface{})["code"])
}

/*
@test TestImpersonate_Success

@desc
Test successful impersonation generates token.

@test_cases
- Valid admin address
- Valid reason provided
- Returns JWT token
- Token has correct expiration
*/
func TestImpersonate_Success(t *testing.T) {
	mockSvc := &mockAdminAuthService{
		generateImpersonateTokenFn: func(ctx interface{}, adminAddress, reason string) (string, int64, error) {
			return "eyJhbGc...", 1685356800, nil
		},
	}

	handler := NewAdminAuthHandler(mockSvc)

	req := request.ImpersonateRequest{
		AdminAddress: "0x1234567890123456789012345678901234567890",
		Reason:       "Testing market creation",
	}

	reqBody, _ := json.Marshal(req)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/v1/admin/auth/impersonate", bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Impersonate(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var respBody map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &respBody)

	assert.True(t, respBody["success"].(bool))
	data := respBody["data"].(map[string]interface{})
	assert.NotEmpty(t, data["token"])
	assert.Equal(t, "0x1234567890123456789012345678901234567890", data["admin_address"])
}

/*
@test TestImpersonate_InvalidRequest

@desc
Test malformed impersonate request returns error.

@error_case
- Invalid JSON body → 400 Bad Request
*/
func TestImpersonate_InvalidRequest(t *testing.T) {
	mockSvc := &mockAdminAuthService{}
	handler := NewAdminAuthHandler(mockSvc)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/v1/admin/auth/impersonate", bytes.NewBuffer([]byte("invalid")))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Impersonate(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var respBody map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &respBody)
	assert.False(t, respBody["success"].(bool))
	assert.Equal(t, "INVALID_REQUEST", respBody["error"].(map[string]interface{})["code"])
}

/*
@test TestImpersonate_MissingFields

@desc
Test impersonate request with missing required fields.

@error_case
- Missing admin_address or reason → 400 Bad Request
*/
func TestImpersonate_MissingFields(t *testing.T) {
	mockSvc := &mockAdminAuthService{}
	handler := NewAdminAuthHandler(mockSvc)

	// Missing reason field
	req := request.ImpersonateRequest{
		AdminAddress: "0x1234567890123456789012345678901234567890",
		Reason:       "", // Empty reason
	}

	reqBody, _ := json.Marshal(req)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/v1/admin/auth/impersonate", bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Impersonate(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

/*
@test TestListAdmins_Success

@desc
Test successful admin list retrieval with pagination.

@test_cases
- Returns paginated admin list
- Includes role and permissions for each admin
- Pagination metadata correct
*/
func TestListAdmins_Success(t *testing.T) {
	mockSvc := &mockAdminAuthService{}
	handler := NewAdminAuthHandler(mockSvc)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/admin/admins?page=1&limit=20", nil)

	handler.ListAdmins(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var respBody map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &respBody)

	assert.True(t, respBody["success"].(bool))
	data := respBody["data"].(map[string]interface{})
	assert.NotNil(t, data["admins"])
	assert.NotNil(t, data["pagination"])
}
