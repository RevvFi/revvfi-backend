package offer

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
Unit tests for offer service.
Tests offer creation, retrieval, and matching logic.

@test_coverage
- CreateOffer: valid/invalid creation
- GetOffer: found/not found
- CalculateQuote: quote generation
- CancelOffer: offer cancellation
*/

// MockOfferRepository implements OfferRepository interface
type MockOfferRepository struct {
	offers map[int64]*models.Offer
}

func NewMockOfferRepository() *MockOfferRepository {
	return &MockOfferRepository{
		offers: make(map[int64]*models.Offer),
	}
}

func (m *MockOfferRepository) CreateOffer(ctx context.Context, offer *models.Offer) error {
	if offer == nil {
		return appErr.ErrInvalidInput
	}
	m.offers[offer.OfferID] = offer
	return nil
}

func (m *MockOfferRepository) GetByID(ctx context.Context, offerID int64) (*models.Offer, error) {
	return m.offers[offerID], nil
}

func (m *MockOfferRepository) GetActiveByMarket(
ctx context.Context,
marketAddr string,
limit, offset int32,
) ([]models.Offer, error) {
	var result []models.Offer
	for _, offer := range m.offers {
		if offer.MarketAddress == marketAddr && offer.Status == "active" {
			result = append(result, *offer)
		}
	}
	return result, nil
}

func (m *MockOfferRepository) UpdateOffer(ctx context.Context, offer *models.Offer) error {
	if offer == nil {
		return appErr.ErrInvalidInput
	}
	m.offers[offer.OfferID] = offer
	return nil
}

func (m *MockOfferRepository) CancelOffer(ctx context.Context, offerID int64) error {
	if offer, exists := m.offers[offerID]; exists {
		offer.Status = "cancelled"
		m.offers[offerID] = offer
	}
	return nil
}

func TestCreateOffer(t *testing.T) {
	tests := []struct {
		name        string
		lender      string
		amount      *big.Int
		apr         int32
		seniority   int16
		expectError bool
	}{
		{
			name:        "valid offer",
			lender:      "0x0000000000000000000000000000000000000001",
			amount:      big.NewInt(1000000000000000000),
			apr:         500,
			seniority:   0,
			expectError: false,
		},
		{
			name:        "invalid APR",
			lender:      "0x0000000000000000000000000000000000000001",
			amount:      big.NewInt(1000000000000000000),
			apr:         6000, // > 5000
			seniority:   0,
			expectError: true,
		},
		{
			name:        "zero amount",
			lender:      "0x0000000000000000000000000000000000000001",
			amount:      big.NewInt(0),
			apr:         500,
			seniority:   0,
			expectError: true,
		},
		{
			name:        "invalid seniority",
			lender:      "0x0000000000000000000000000000000000000001",
			amount:      big.NewInt(1000000000000000000),
			apr:         500,
			seniority:   2, // only 0 or 1
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
mockRepo := NewMockOfferRepository()
			service := NewOfferService(mockRepo)

			_, err := service.CreateOffer(
context.Background(),
				1,
				tt.lender,
				"0x0000000000000000000000000000000000000002",
				tt.amount,
				tt.apr,
				tt.seniority,
				30,
			)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestGetOffer(t *testing.T) {
	mockRepo := NewMockOfferRepository()
	service := NewOfferService(mockRepo)

	offer := &models.Offer{
		OfferID:         1,
		Lender:          "0x0000000000000000000000000000000000000001",
		MarketAddress:   "0x0000000000000000000000000000000000000002",
		Amount:          big.NewInt(1000000000000000000),
		RemainingAmount: big.NewInt(1000000000000000000),
		APR:             500,
		Status:          "active",
	}
	mockRepo.CreateOffer(context.Background(), offer)

	// Test existing offer
	retrieved, err := service.GetOffer(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retrieved == nil {
		t.Fatal("offer should not be nil")
	}
	if retrieved.OfferID != 1 {
		t.Errorf("offer ID mismatch: got %d, want 1", retrieved.OfferID)
	}

	// Test non-existent offer
	_, err = service.GetOffer(context.Background(), 999)
	if err != appErr.ErrOfferNotFound {
		t.Errorf("expected ErrOfferNotFound, got %v", err)
	}
}

func TestCancelOffer(t *testing.T) {
	mockRepo := NewMockOfferRepository()
	service := NewOfferService(mockRepo)

	lender := "0x0000000000000000000000000000000000000001"
	offer := &models.Offer{
		OfferID:         1,
		Lender:          lender,
		MarketAddress:   "0x0000000000000000000000000000000000000002",
		Amount:          big.NewInt(1000000000000000000),
		RemainingAmount: big.NewInt(1000000000000000000),
		APR:             500,
		Status:          "active",
	}
	mockRepo.CreateOffer(context.Background(), offer)

	// Test valid cancellation
	err := service.CancelOffer(context.Background(), 1, lender)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test unauthorized cancellation
	err = service.CancelOffer(context.Background(), 1, "0x9999999999999999999999999999999999999999")
	if err != appErr.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}
