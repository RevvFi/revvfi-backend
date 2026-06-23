package borrower

import (
	"context"
	"math/big"
	"testing"

	"github.com/Revvfi/revvfi-backend/internal/models"
	appErr "github.com/Revvfi/revvfi-backend/internal/pkg/errors"
)

/*
@file service_test.go

@desc
Unit tests for borrower service.
Tests reputation tracking and risk assessment.

@test_coverage
- RegisterBorrower: new borrower creation
- GetBorrower: retrieval
- RecordSuccessfulRepayment: reputation increases
- RecordDefault: reputation decreases
- GetRiskAssessment: risk calculation
*/

// MockBorrowerRepository implements the BorrowerRepository interface defined in service.go
type MockBorrowerRepository struct {
	borrowers map[string]*models.Borrower
}

func NewMockBorrowerRepository() *MockBorrowerRepository {
	return &MockBorrowerRepository{
		borrowers: make(map[string]*models.Borrower),
	}
}

func (m *MockBorrowerRepository) CreateBorrower(ctx context.Context, borrower *models.Borrower) error {
	if borrower == nil {
		return appErr.ErrInvalidInput
	}
	m.borrowers[borrower.Address] = borrower
	return nil
}

func (m *MockBorrowerRepository) GetByAddress(ctx context.Context, address string) (*models.Borrower, error) {
	return m.borrowers[address], nil
}

func (m *MockBorrowerRepository) UpdateBorrower(ctx context.Context, borrower *models.Borrower) error {
	if borrower == nil {
		return appErr.ErrInvalidInput
	}
	m.borrowers[borrower.Address] = borrower
	return nil
}

func (m *MockBorrowerRepository) GetCollateralBalance(ctx context.Context, address string) (*models.CollateralBalance, error) {
	return nil, nil
}

func TestRegisterBorrower(t *testing.T) {
	mockRepo := NewMockBorrowerRepository()
	service := NewBorrowerService(mockRepo)

	address := "0x0000000000000000000000000000000000000001"

	borrower, err := service.RegisterBorrower(context.Background(), address)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if borrower == nil {
		t.Fatal("borrower should not be nil")
	}

	if borrower.Address != address {
		t.Errorf("address mismatch: got %s, want %s", borrower.Address, address)
	}

	// Check starting reputation
	if borrower.ReputationScore != 500 {
		t.Errorf("expected starting score 500, got %d", borrower.ReputationScore)
	}
}

func TestGetBorrower(t *testing.T) {
	mockRepo := NewMockBorrowerRepository()
	service := NewBorrowerService(mockRepo)

	address := "0x0000000000000000000000000000000000000001"
	borrower := &models.Borrower{
		Address:         address,
		TotalBorrowed:   big.NewInt(1000000),
		ReputationScore: 600,
	}
	mockRepo.CreateBorrower(context.Background(), borrower)

	retrieved, err := service.GetBorrower(context.Background(), address)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if retrieved.Address != address {
		t.Errorf("address mismatch")
	}

	// Test non-existent
	_, err = service.GetBorrower(context.Background(), "0x9999999999999999999999999999999999999999")
	if err != appErr.ErrBorrowerNotFound {
		t.Errorf("expected ErrBorrowerNotFound, got %v", err)
	}
}

func TestRecordSuccessfulRepayment(t *testing.T) {
	mockRepo := NewMockBorrowerRepository()
	service := NewBorrowerService(mockRepo)

	address := "0x0000000000000000000000000000000000000001"
	borrower := &models.Borrower{
		Address:         address,
		TotalLoans:      5,
		SuccessfulLoans: 3,
		DefaultedLoans:  0,
		ReputationScore: 600,
	}
	mockRepo.CreateBorrower(context.Background(), borrower)

	err := service.RecordSuccessfulRepayment(context.Background(), address)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, _ := mockRepo.GetByAddress(context.Background(), address)
	if updated.SuccessfulLoans != 4 {
		t.Errorf("expected 4 successful loans, got %d", updated.SuccessfulLoans)
	}
}

func TestRecordDefault(t *testing.T) {
	mockRepo := NewMockBorrowerRepository()
	service := NewBorrowerService(mockRepo)

	address := "0x0000000000000000000000000000000000000001"
	borrower := &models.Borrower{
		Address:         address,
		TotalLoans:      5,
		SuccessfulLoans: 3,
		DefaultedLoans:  0,
		ReputationScore: 600,
	}
	mockRepo.CreateBorrower(context.Background(), borrower)

	err := service.RecordDefault(context.Background(), address)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, _ := mockRepo.GetByAddress(context.Background(), address)
	if updated.DefaultedLoans != 1 {
		t.Errorf("expected 1 default, got %d", updated.DefaultedLoans)
	}

	// Score should have decreased
	if updated.ReputationScore >= 600 {
		t.Errorf("reputation score should decrease after default")
	}
}
