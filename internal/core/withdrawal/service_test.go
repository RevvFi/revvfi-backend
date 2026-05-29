package withdrawal

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/Revvfi/revvfi-backend/internal/models"
)

/*
@file service_test.go

@desc
Unit tests for withdrawal service.
Tests request lifecycle, epoch processing, and fulfillment calculations.

@test_coverage
- RequestWithdrawal: valid and invalid requests
- CancelWithdrawalRequest: cancellation authorization and state
- GetWithdrawalRequests: pagination and filtering
- ProcessEpoch: fulfillment calculations
- GetCurrentEpoch: epoch retrieval
*/

/*
@struct MockWithdrawalRepository

@desc
Mock implementation of WithdrawalRepository for testing.
*/
type MockWithdrawalRepository struct {
	requests map[int64]*models.WithdrawalRequest
	epochs   map[int64]*models.WithdrawalEpoch
	current  *models.WithdrawalEpoch
}

func NewMockWithdrawalRepository() *MockWithdrawalRepository {
	return &MockWithdrawalRepository{
		requests: make(map[int64]*models.WithdrawalRequest),
		epochs:   make(map[int64]*models.WithdrawalEpoch),
		current: &models.WithdrawalEpoch{
			ID:            1,
			EpochNumber:   0,
			StartTime:     time.Now(),
			EndTime:       time.Now().Add(7 * 24 * time.Hour),
			TotalRequested: big.NewInt(0),
			TotalFulfilled: big.NewInt(0),
			Status:        "pending",
			CreatedAt:     time.Now(),
		},
	}
}

func (m *MockWithdrawalRepository) CreateRequest(ctx context.Context, request *models.WithdrawalRequest) error {
	request.ID = int64(len(m.requests) + 1)
	request.RequestID = request.ID
	m.requests[request.ID] = request
	return nil
}

func (m *MockWithdrawalRepository) GetRequestByID(ctx context.Context, requestID int64) (*models.WithdrawalRequest, error) {
	return m.requests[requestID], nil
}

func (m *MockWithdrawalRepository) GetRequestsByLender(ctx context.Context, lender string, limit, offset int32) ([]models.WithdrawalRequest, error) {
	var result []models.WithdrawalRequest
	for _, req := range m.requests {
		if req.Lender == lender {
			result = append(result, *req)
		}
	}
	return result, nil
}

func (m *MockWithdrawalRepository) UpdateRequest(ctx context.Context, request *models.WithdrawalRequest) error {
	m.requests[request.ID] = request
	return nil
}

func (m *MockWithdrawalRepository) GetActiveRequests(ctx context.Context, epochNumber int64) ([]models.WithdrawalRequest, error) {
	var result []models.WithdrawalRequest
	for _, req := range m.requests {
		if req.EpochNumber == int32(epochNumber) && (req.Status == "pending" || req.Status == "fulfilled") {
			result = append(result, *req)
		}
	}
	return result, nil
}

func (m *MockWithdrawalRepository) CreateEpoch(ctx context.Context, epoch *models.WithdrawalEpoch) error {
	m.epochs[int64(epoch.EpochNumber)] = epoch
	return nil
}

func (m *MockWithdrawalRepository) GetEpochByNumber(ctx context.Context, epochNumber int64) (*models.WithdrawalEpoch, error) {
	return m.epochs[epochNumber], nil
}

func (m *MockWithdrawalRepository) UpdateEpoch(ctx context.Context, epoch *models.WithdrawalEpoch) error {
	m.epochs[int64(epoch.EpochNumber)] = epoch
	return nil
}

func (m *MockWithdrawalRepository) GetCurrentEpoch(ctx context.Context) (*models.WithdrawalEpoch, error) {
	return m.current, nil
}

/*
@struct MockPositionRepository

@desc
Mock implementation of PositionRepository for testing.
*/
type MockPositionRepository struct {
	positions map[int64]*models.Position
}

func NewMockPositionRepository() *MockPositionRepository {
	return &MockPositionRepository{
		positions: make(map[int64]*models.Position),
	}
}

func (m *MockPositionRepository) GetByTokenID(ctx context.Context, tokenID int64) (*models.Position, error) {
	return m.positions[tokenID], nil
}

func (m *MockPositionRepository) UpdatePosition(ctx context.Context, position *models.Position) error {
	m.positions[position.TokenID] = position
	return nil
}

// Test RequestWithdrawal - Valid Request
func TestRequestWithdrawalValid(t *testing.T) {
	ctx := context.Background()
	mockWithdrawal := NewMockWithdrawalRepository()
	mockPosition := NewMockPositionRepository()

	// Create test position
	position := &models.Position{
		TokenID:          1,
		Lender:           "0x123",
		MarketAddress:    "0xabc",
		Principal:        big.NewInt(1000000),
		CurrentPrincipal: big.NewInt(1000000),
		ClaimableAmount:  big.NewInt(500000),
		APR:              500,
		Status:           "active",
	}
	mockPosition.positions[1] = position

	service := NewWithdrawalService(mockWithdrawal, mockPosition)

	// Request withdrawal
	request, err := service.RequestWithdrawal(ctx, "0x123", 1, big.NewInt(300000))
	if err != nil {
		t.Fatalf("RequestWithdrawal failed: %v", err)
	}

	if request.Status != "pending" {
		t.Errorf("Expected status 'pending', got '%s'", request.Status)
	}

	if request.Lender != "0x123" {
		t.Errorf("Expected lender '0x123', got '%s'", request.Lender)
	}
}

// Test RequestWithdrawal - Unauthorized
func TestRequestWithdrawalUnauthorized(t *testing.T) {
	ctx := context.Background()
	mockWithdrawal := NewMockWithdrawalRepository()
	mockPosition := NewMockPositionRepository()

	position := &models.Position{
		TokenID:          1,
		Lender:           "0x123",
		MarketAddress:    "0xabc",
		Principal:        big.NewInt(1000000),
		CurrentPrincipal: big.NewInt(1000000),
		ClaimableAmount:  big.NewInt(500000),
	}
	mockPosition.positions[1] = position

	service := NewWithdrawalService(mockWithdrawal, mockPosition)

	// Try to withdraw from different lender
	_, err := service.RequestWithdrawal(ctx, "0x456", 1, big.NewInt(300000))
	if err == nil {
		t.Fatal("Expected error for unauthorized withdrawal")
	}
}

// Test RequestWithdrawal - Insufficient Balance
func TestRequestWithdrawalInsufficientBalance(t *testing.T) {
	ctx := context.Background()
	mockWithdrawal := NewMockWithdrawalRepository()
	mockPosition := NewMockPositionRepository()

	position := &models.Position{
		TokenID:          1,
		Lender:           "0x123",
		MarketAddress:    "0xabc",
		Principal:        big.NewInt(1000000),
		CurrentPrincipal: big.NewInt(1000000),
		ClaimableAmount:  big.NewInt(500000),
	}
	mockPosition.positions[1] = position

	service := NewWithdrawalService(mockWithdrawal, mockPosition)

	// Try to withdraw more than claimable
	_, err := service.RequestWithdrawal(ctx, "0x123", 1, big.NewInt(600000))
	if err == nil {
		t.Fatal("Expected error for insufficient claimable balance")
	}
}

// Test CancelWithdrawalRequest - Valid Cancellation
func TestCancelWithdrawalRequestValid(t *testing.T) {
	ctx := context.Background()
	mockWithdrawal := NewMockWithdrawalRepository()
	mockPosition := NewMockPositionRepository()

	// Create request
	request := &models.WithdrawalRequest{
		ID:              1,
		RequestID:       1,
		Lender:          "0x123",
		PositionID:      1,
		RequestedAmount: big.NewInt(300000),
		Status:          "pending",
		CreatedAt:       time.Now(),
	}
	mockWithdrawal.requests[1] = request

	service := NewWithdrawalService(mockWithdrawal, mockPosition)

	// Cancel request
	err := service.CancelWithdrawalRequest(ctx, 1, "0x123")
	if err != nil {
		t.Fatalf("CancelWithdrawalRequest failed: %v", err)
	}

	// Verify status changed
	updatedRequest := mockWithdrawal.requests[1]
	if updatedRequest.Status != "cancelled" {
		t.Errorf("Expected status 'cancelled', got '%s'", updatedRequest.Status)
	}
}

// Test CancelWithdrawalRequest - Unauthorized
func TestCancelWithdrawalRequestUnauthorized(t *testing.T) {
	ctx := context.Background()
	mockWithdrawal := NewMockWithdrawalRepository()
	mockPosition := NewMockPositionRepository()

	request := &models.WithdrawalRequest{
		ID:        1,
		RequestID: 1,
		Lender:    "0x123",
		Status:    "pending",
	}
	mockWithdrawal.requests[1] = request

	service := NewWithdrawalService(mockWithdrawal, mockPosition)

	// Try to cancel with different lender
	err := service.CancelWithdrawalRequest(ctx, 1, "0x456")
	if err == nil {
		t.Fatal("Expected error for unauthorized cancellation")
	}
}

// Test ProcessEpoch - Fulfillment Calculation
func TestProcessEpochFulfillment(t *testing.T) {
	ctx := context.Background()
	mockWithdrawal := NewMockWithdrawalRepository()
	mockPosition := NewMockPositionRepository()

	// Setup epoch
	epoch := &models.WithdrawalEpoch{
		ID:            1,
		EpochNumber:   0,
		StartTime:     time.Now(),
		EndTime:       time.Now().Add(7 * 24 * time.Hour),
		TotalRequested: big.NewInt(1000000),
		TotalFulfilled: big.NewInt(0),
		Status:        "pending",
		CreatedAt:     time.Now(),
	}
	mockWithdrawal.epochs[0] = epoch

	// Create requests
	request1 := &models.WithdrawalRequest{
		ID:              1,
		RequestID:       1,
		Lender:          "0x123",
		PositionID:      1,
		RequestedAmount: big.NewInt(600000),
		Status:          "pending",
		EpochNumber:     0,
	}
	request2 := &models.WithdrawalRequest{
		ID:              2,
		RequestID:       2,
		Lender:          "0x456",
		PositionID:      2,
		RequestedAmount: big.NewInt(400000),
		Status:          "pending",
		EpochNumber:     0,
	}
	mockWithdrawal.requests[1] = request1
	mockWithdrawal.requests[2] = request2

	service := NewWithdrawalService(mockWithdrawal, mockPosition)

	// Process epoch
	err := service.ProcessEpoch(ctx, 0)
	if err != nil {
		t.Fatalf("ProcessEpoch failed: %v", err)
	}

	// Verify epoch status changed
	updatedEpoch := mockWithdrawal.epochs[0]
	if updatedEpoch.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", updatedEpoch.Status)
	}

	// Verify requests fulfilled
	updatedRequest1 := mockWithdrawal.requests[1]
	if updatedRequest1.Status != "fulfilled" {
		t.Errorf("Expected request status 'fulfilled', got '%s'", updatedRequest1.Status)
	}
}

// Test GetCurrentEpoch
func TestGetCurrentEpoch(t *testing.T) {
	ctx := context.Background()
	mockWithdrawal := NewMockWithdrawalRepository()
	mockPosition := NewMockPositionRepository()

	service := NewWithdrawalService(mockWithdrawal, mockPosition)

	epoch, err := service.GetCurrentEpoch(ctx)
	if err != nil {
		t.Fatalf("GetCurrentEpoch failed: %v", err)
	}

	if epoch.EpochNumber != 0 {
		t.Errorf("Expected epoch number 0, got %d", epoch.EpochNumber)
	}
}

// Test GetWithdrawalRequests - Pagination
func TestGetWithdrawalRequestsPagination(t *testing.T) {
	ctx := context.Background()
	mockWithdrawal := NewMockWithdrawalRepository()
	mockPosition := NewMockPositionRepository()

	// Create multiple requests
	for i := 0; i < 5; i++ {
		request := &models.WithdrawalRequest{
			ID:        int64(i + 1),
			RequestID: int64(i + 1),
			Lender:    "0x123",
			Status:    "pending",
		}
		mockWithdrawal.requests[int64(i+1)] = request
	}

	service := NewWithdrawalService(mockWithdrawal, mockPosition)

	// Get requests
	requests, err := service.GetWithdrawalRequests(ctx, "0x123", 10, 0)
	if err != nil {
		t.Fatalf("GetWithdrawalRequests failed: %v", err)
	}

	if len(requests) != 5 {
		t.Errorf("Expected 5 requests, got %d", len(requests))
	}
}

// Test GetEpochStatus
func TestGetEpochStatus(t *testing.T) {
	ctx := context.Background()
	mockWithdrawal := NewMockWithdrawalRepository()
	mockPosition := NewMockPositionRepository()

	// Setup epoch
	epoch := &models.WithdrawalEpoch{
		ID:            1,
		EpochNumber:   0,
		StartTime:     time.Now(),
		EndTime:       time.Now().Add(7 * 24 * time.Hour),
		TotalRequested: big.NewInt(1000000),
		TotalFulfilled: big.NewInt(900000),
		Status:        "completed",
		CreatedAt:     time.Now(),
	}
	mockWithdrawal.epochs[0] = epoch

	service := NewWithdrawalService(mockWithdrawal, mockPosition)

	status, err := service.GetEpochStatus(ctx, 0)
	if err != nil {
		t.Fatalf("GetEpochStatus failed: %v", err)
	}

	if status["status"] != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", status["status"])
	}

	if epochNum, ok := status["epoch_number"].(int32); !ok || epochNum != 0 {
		t.Errorf("Expected epoch number 0 in status")
	}
}

// Test InvalidAmount
func TestRequestWithdrawalInvalidAmount(t *testing.T) {
	ctx := context.Background()
	mockWithdrawal := NewMockWithdrawalRepository()
	mockPosition := NewMockPositionRepository()

	position := &models.Position{
		TokenID:          1,
		Lender:           "0x123",
		ClaimableAmount:  big.NewInt(500000),
	}
	mockPosition.positions[1] = position

	service := NewWithdrawalService(mockWithdrawal, mockPosition)

	// Try to withdraw zero amount
	_, err := service.RequestWithdrawal(ctx, "0x123", 1, big.NewInt(0))
	if err == nil {
		t.Fatal("Expected error for zero amount")
	}
}

// Test PositionNotFound
func TestRequestWithdrawalPositionNotFound(t *testing.T) {
	ctx := context.Background()
	mockWithdrawal := NewMockWithdrawalRepository()
	mockPosition := NewMockPositionRepository()

	service := NewWithdrawalService(mockWithdrawal, mockPosition)

	// Try to withdraw from non-existent position
	_, err := service.RequestWithdrawal(ctx, "0x123", 999, big.NewInt(300000))
	if err == nil {
		t.Fatal("Expected error for non-existent position")
	}
}
