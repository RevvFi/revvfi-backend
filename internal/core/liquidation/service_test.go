package liquidation

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
Unit tests for liquidation service.
*/

/*
@struct MockAuctionRepository

@desc
Mock implementation for auction repository.
*/
type MockAuctionRepository struct {
	auctions map[string]*models.Auction
	err      error
}

func (m *MockAuctionRepository) CreateAuction(ctx context.Context, auction *models.Auction) error {
	if m.err != nil {
		return m.err
	}
	key := auction.MarketAddress
	m.auctions[key] = auction
	return nil
}

func (m *MockAuctionRepository) GetByID(ctx context.Context, auctionID string) (*models.Auction, error) {
	if m.err != nil {
		return nil, m.err
	}
	auction, exists := m.auctions[auctionID]
	if !exists {
		return nil, nil
	}
	return auction, nil
}

func (m *MockAuctionRepository) GetActiveByMarket(ctx context.Context, market string) ([]models.Auction, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []models.Auction
	for _, auction := range m.auctions {
		if auction.MarketAddress == market && auction.Status == "active" {
			result = append(result, *auction)
		}
	}
	return result, nil
}

func (m *MockAuctionRepository) UpdateAuction(ctx context.Context, auction *models.Auction) error {
	if m.err != nil {
		return m.err
	}
	m.auctions[auction.MarketAddress] = auction
	return nil
}

func (m *MockAuctionRepository) GetLiquidatableCount(ctx context.Context) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	count := int64(0)
	for _, auction := range m.auctions {
		if auction.Status == "active" {
			count++
		}
	}
	return count, nil
}

func (m *MockAuctionRepository) GetAllActive(ctx context.Context) ([]models.Auction, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make([]models.Auction, 0)
	for _, auction := range m.auctions {
		if auction.Status == "active" {
			result = append(result, *auction)
		}
	}
	return result, nil
}

func (m *MockAuctionRepository) GetBidsByAuction(ctx context.Context, auctionID int64) ([]models.Bid, error) {
	if m.err != nil {
		return nil, m.err
	}
	return []models.Bid{}, nil
}

/*
@struct MockMarketRepository

@desc
Mock implementation for market repository.
*/
type MockMarketRepository struct {
	markets map[string]*models.Market
	err     error
}

func (m *MockMarketRepository) GetByAddress(ctx context.Context, address string) (*models.Market, error) {
	if m.err != nil {
		return nil, m.err
	}
	market, exists := m.markets[address]
	if !exists {
		return nil, nil
	}
	return market, nil
}

func (m *MockMarketRepository) UpdateMarket(ctx context.Context, market *models.Market) error {
	if m.err != nil {
		return m.err
	}
	m.markets[market.Address] = market
	return nil
}

func (m *MockMarketRepository) GetAllActive(ctx context.Context) ([]models.Market, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []models.Market
	for _, market := range m.markets {
		if market.IsActive && !market.IsClosed {
			result = append(result, *market)
		}
	}
	return result, nil
}

/*
@struct MockBorrowerRepository

@desc
Mock implementation for borrower repository.
*/
type MockBorrowerRepository struct {
	borrowers map[string]*models.Borrower
	err       error
}

func (m *MockBorrowerRepository) GetByAddress(ctx context.Context, address string) (*models.Borrower, error) {
	if m.err != nil {
		return nil, m.err
	}
	borrower, exists := m.borrowers[address]
	if !exists {
		return nil, nil
	}
	return borrower, nil
}

func (m *MockBorrowerRepository) UpdateBorrower(ctx context.Context, borrower *models.Borrower) error {
	if m.err != nil {
		return m.err
	}
	m.borrowers[borrower.Address] = borrower
	return nil
}

/*
@function TestGetAuction

@desc
Tests auction retrieval.
*/
func TestGetAuction(t *testing.T) {
	auction := &models.Auction{
		MarketAddress: "0xMarket1",
		Status:        "active",
		StartTime:     time.Now(),
		EndTime:       time.Now().Add(24 * time.Hour),
		DebtAmount:    big.NewInt(1000000),
		CurrentPrice:  big.NewInt(1000000),
	}

	mockAuction := &MockAuctionRepository{
		auctions: map[string]*models.Auction{
			"0xMarket1": auction,
		},
	}
	mockMarket := &MockMarketRepository{markets: make(map[string]*models.Market)}
	mockBorrower := &MockBorrowerRepository{borrowers: make(map[string]*models.Borrower)}

	svc := NewLiquidationService(mockAuction, mockMarket, mockBorrower)

	retrieved, err := svc.GetAuction(context.Background(), "0xMarket1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if retrieved == nil {
		t.Error("expected auction, got nil")
	}
	if retrieved.MarketAddress != "0xMarket1" {
		t.Errorf("expected 0xMarket1, got %s", retrieved.MarketAddress)
	}
}

/*
@function TestPlaceBid

@desc
Tests bid placement.
*/
func TestPlaceBid(t *testing.T) {
	mockAuction := &MockAuctionRepository{
		auctions: map[string]*models.Auction{
			"0xMarket1": {
				MarketAddress: "0xMarket1",
				Status:        "active",
				StartTime:     time.Now().Add(-1 * time.Hour),
				EndTime:       time.Now().Add(23 * time.Hour),
				DebtAmount:    big.NewInt(1000000),
				CurrentPrice:  big.NewInt(900000),
				CreatedAt:     time.Now().Add(-1 * time.Hour),
			},
		},
	}
	mockMarket := &MockMarketRepository{markets: make(map[string]*models.Market)}
	mockBorrower := &MockBorrowerRepository{borrowers: make(map[string]*models.Borrower)}

	svc := NewLiquidationService(mockAuction, mockMarket, mockBorrower)

	err := svc.PlaceBid(context.Background(), "0xMarket1", "0xBidder1", big.NewInt(1000000))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	updated, _ := mockAuction.GetByID(context.Background(), "0xMarket1")
	if updated.HighestBidder.String != "0xBidder1" {
		t.Errorf("expected bidder 0xBidder1, got %s", updated.HighestBidder.String)
	}
}

/*
@function TestSettleAuction

@desc
Tests auction settlement.
*/
func TestSettleAuction(t *testing.T) {
	mockAuction := &MockAuctionRepository{
		auctions: map[string]*models.Auction{
			"0xMarket1": {
				MarketAddress: "0xMarket1",
				Status:        "active",
				StartTime:     time.Now().Add(-2 * time.Hour),
				EndTime:       time.Now().Add(-1 * time.Hour),
				HighestBid:    big.NewInt(950000),
			},
		},
	}
	mockMarket := &MockMarketRepository{markets: make(map[string]*models.Market)}
	mockBorrower := &MockBorrowerRepository{borrowers: make(map[string]*models.Borrower)}

	svc := NewLiquidationService(mockAuction, mockMarket, mockBorrower)

	err := svc.SettleAuction(context.Background(), "0xMarket1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	settled, _ := mockAuction.GetByID(context.Background(), "0xMarket1")
	if settled.Status != "settled" {
		t.Errorf("expected status settled, got %s", settled.Status)
	}
}

/*
@function TestGetHealthFactor

@desc
Tests health factor calculation.
*/
func TestGetHealthFactor(t *testing.T) {
	mockAuction := &MockAuctionRepository{auctions: make(map[string]*models.Auction)}
	mockMarket := &MockMarketRepository{
		markets: map[string]*models.Market{
			"0xMarket1": {
				Address:        "0xMarket1",
				TotalLiquidity: big.NewInt(1500000),
				TotalDebt:      big.NewInt(1000000),
				IsActive:       true,
			},
		},
	}
	mockBorrower := &MockBorrowerRepository{borrowers: make(map[string]*models.Borrower)}

	svc := NewLiquidationService(mockAuction, mockMarket, mockBorrower)

	hf, err := svc.GetHealthFactor(context.Background(), "0xMarket1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expectedHF := 1.5
	if hf < expectedHF-0.01 || hf > expectedHF+0.01 {
		t.Errorf("expected HF around 1.5, got %f", hf)
	}
}

/*
@function TestCancelAuction

@desc
Tests auction cancellation.
*/
func TestCancelAuction(t *testing.T) {
	mockAuction := &MockAuctionRepository{
		auctions: map[string]*models.Auction{
			"0xMarket1": {
				MarketAddress: "0xMarket1",
				Status:        "active",
			},
		},
	}
	mockMarket := &MockMarketRepository{markets: make(map[string]*models.Market)}
	mockBorrower := &MockBorrowerRepository{borrowers: make(map[string]*models.Borrower)}

	svc := NewLiquidationService(mockAuction, mockMarket, mockBorrower)

	err := svc.CancelAuction(context.Background(), "0xMarket1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	auction, _ := mockAuction.GetByID(context.Background(), "0xMarket1")
	if auction.Status != "cancelled" {
		t.Errorf("expected status cancelled, got %s", auction.Status)
	}
}
