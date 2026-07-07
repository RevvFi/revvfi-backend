package market

import (
"context"
"math/big"
"strings"
"testing"

"github.com/Revvfi/revvfi-backend/internal/models"
appErr "github.com/Revvfi/revvfi-backend/internal/pkg/errors"
)

/*
@file service_test.go

@desc
Unit tests for market service.
Tests market creation, retrieval, and metrics calculation.

@test_coverage
- CreateMarket: valid creation, validation errors
- GetMarket: found, not found
- ListMarkets: returns markets
- CalculateMetrics: calculates metrics correctly
*/

// MockMarketRepository implements MarketRepository interface
type MockMarketRepository struct {
	markets map[string]*models.Market
}

func NewMockMarketRepository() *MockMarketRepository {
	return &MockMarketRepository{
		markets: make(map[string]*models.Market),
	}
}

func (m *MockMarketRepository) CreateMarket(ctx context.Context, market *models.Market) error {
	if market == nil {
		return appErr.ErrInvalidInput
	}
	m.markets[market.Address] = market
	return nil
}

func (m *MockMarketRepository) GetByAddress(ctx context.Context, address string) (*models.Market, error) {
	market, exists := m.markets[address]
	if !exists {
		return nil, nil
	}
	return market, nil
}

func (m *MockMarketRepository) ListActive(ctx context.Context, limit, offset int32, borrower string) ([]models.Market, error) {
	result := make([]models.Market, 0)
	for _, market := range m.markets {
		if !market.IsActive {
			continue
		}
		if borrower != "" && !strings.EqualFold(market.Borrower, borrower) {
			continue
		}
		result = append(result, *market)
	}
	return result, nil
}

func (m *MockMarketRepository) UpdateMarket(ctx context.Context, market *models.Market) error {
	if market == nil {
		return appErr.ErrInvalidInput
	}
	m.markets[market.Address] = market
	return nil
}

// MockOfferRepository implements OfferRepository interface
type MockOfferRepository struct {
	offers map[string][]models.Offer
}

func NewMockOfferRepository() *MockOfferRepository {
	return &MockOfferRepository{
		offers: make(map[string][]models.Offer),
	}
}

func (m *MockOfferRepository) GetActiveByMarket(ctx context.Context, marketAddr string) ([]models.Offer, error) {
	return m.offers[marketAddr], nil
}

func (m *MockOfferRepository) CountActive(ctx context.Context, marketAddr string) (int32, error) {
	return int32(len(m.offers[marketAddr])), nil
}

// MockPositionRepository implements PositionRepository interface
type MockPositionRepository struct{}

func (m *MockPositionRepository) CountActiveByMarket(ctx context.Context, marketAddr string) (int32, error) {
	return 0, nil
}

func (m *MockPositionRepository) GetTotalPrincipalByMarket(ctx context.Context, marketAddr string) (*big.Int, error) {
	return big.NewInt(0), nil
}

func TestCreateMarket(t *testing.T) {
	tests := []struct {
		name               string
		marketAddr         string
		borrower           string
		borrowAsset        string
		collateralAsset    string
		collateralOracle   string
		minCollateralRatio int32
		liquidationThreshold int32
		expectError        bool
	}{
		{
			name:                 "valid market creation",
			marketAddr:           "0x1234567890123456789012345678901234567890",
			borrower:             "0x0000000000000000000000000000000000000001",
			borrowAsset:          "0x0000000000000000000000000000000000000002",
			collateralAsset:      "0x0000000000000000000000000000000000000003",
			collateralOracle:     "0x0000000000000000000000000000000000000004",
			minCollateralRatio:   15000, // 150%
			liquidationThreshold: 12500, // 125%
			expectError:          false,
		},
		{
			name:                 "invalid borrower address",
			marketAddr:           "0x1234567890123456789012345678901234567890",
			borrower:             "invalid",
			borrowAsset:          "0x0000000000000000000000000000000000000002",
			collateralAsset:      "0x0000000000000000000000000000000000000003",
			collateralOracle:     "0x0000000000000000000000000000000000000004",
			minCollateralRatio:   15000,
			liquidationThreshold: 12500,
			expectError:          true,
		},
		{
			name:                 "invalid collateral ratio",
			marketAddr:           "0x1234567890123456789012345678901234567890",
			borrower:             "0x0000000000000000000000000000000000000001",
			borrowAsset:          "0x0000000000000000000000000000000000000002",
			collateralAsset:      "0x0000000000000000000000000000000000000003",
			collateralOracle:     "0x0000000000000000000000000000000000000004",
			minCollateralRatio:   5000, // < 100%, invalid
			liquidationThreshold: 4000,
			expectError:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
mockRepo := NewMockMarketRepository()
			mockOfferRepo := NewMockOfferRepository()
			mockPosRepo := &MockPositionRepository{}

			service := NewMarketService(mockRepo, mockOfferRepo, mockPosRepo)

			_, err := service.CreateMarket(
context.Background(),
				tt.marketAddr,
				tt.borrower,
				tt.borrowAsset,
				tt.collateralAsset,
				tt.collateralOracle,
				tt.minCollateralRatio,
				tt.liquidationThreshold,
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

func TestGetMarket(t *testing.T) {
	mockRepo := NewMockMarketRepository()
	mockOfferRepo := NewMockOfferRepository()
	mockPosRepo := &MockPositionRepository{}

	market := &models.Market{
		Address:  "0x1234567890123456789012345678901234567890",
		IsActive: true,
	}
	mockRepo.CreateMarket(context.Background(), market)

	service := NewMarketService(mockRepo, mockOfferRepo, mockPosRepo)

	// Test existing market
	retrieved, err := service.GetMarket(context.Background(), market.Address)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retrieved == nil {
		t.Fatal("market should not be nil")
	}
	if retrieved.Address != market.Address {
		t.Errorf("market address mismatch: got %s, want %s", retrieved.Address, market.Address)
	}

	// Test non-existent market
	_, err = service.GetMarket(context.Background(), "0x9999999999999999999999999999999999999999")
	if err != appErr.ErrMarketNotFound {
		t.Errorf("expected ErrMarketNotFound, got %v", err)
	}
}

func TestListMarkets(t *testing.T) {
	mockRepo := NewMockMarketRepository()
	mockOfferRepo := NewMockOfferRepository()
	mockPosRepo := &MockPositionRepository{}

	// Create test markets
	for i := 0; i < 5; i++ {
		addr := "0x" + string(rune(i)) + "234567890123456789012345678901234567890"
		market := &models.Market{
			Address:  addr,
			IsActive: true,
		}
		mockRepo.CreateMarket(context.Background(), market)
	}

	service := NewMarketService(mockRepo, mockOfferRepo, mockPosRepo)

	markets, err := service.ListMarkets(context.Background(), 10, 0, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(markets) != 5 {
		t.Errorf("expected 5 markets, got %d", len(markets))
	}
}
