package liquidation

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"time"

	"github.com/Revvfi/revvfi-backend/internal/models"
	appErr "github.com/Revvfi/revvfi-backend/internal/pkg/errors"
) /*
@file service.go

@desc
Liquidation service manages auction creation and bidding.

@responsibilities
- Detect liquidatable positions
- Create Dutch auctions
- Process bids
- Settle auctions
*/

/*
@struct LiquidationService

@desc
Handles liquidation auctions.

@dependencies
- AuctionRepository: auction data access
- MarketRepository: market data access
- BorrowerRepository: borrower data access
*/
type LiquidationService struct {
	auctionRepo  AuctionRepository
	marketRepo   MarketRepository
	borrowerRepo BorrowerRepository
	monitor      *Monitor
	auction      *AuctionManager
}

/*
@interface AuctionRepository

@desc
Repository for auction data access.
*/
type AuctionRepository interface {
	CreateAuction(ctx context.Context, auction *models.Auction) error
	GetByID(ctx context.Context, auctionID string) (*models.Auction, error)
	GetActiveByMarket(ctx context.Context, market string) ([]models.Auction, error)
	UpdateAuction(ctx context.Context, auction *models.Auction) error
	GetLiquidatableCount(ctx context.Context) (int64, error)
}

/*
@interface MarketRepository

@desc
Repository for market data access.
*/
type MarketRepository interface {
	GetByAddress(ctx context.Context, address string) (*models.Market, error)
	UpdateMarket(ctx context.Context, market *models.Market) error
	GetAllActive(ctx context.Context) ([]models.Market, error)
}

/*
@interface BorrowerRepository

@desc
Repository for borrower data access.
*/
type BorrowerRepository interface {
	GetByAddress(ctx context.Context, address string) (*models.Borrower, error)
	UpdateBorrower(ctx context.Context, borrower *models.Borrower) error
}

/*
@function NewLiquidationService

@desc
Creates new liquidation service.

@params
- auctionRepo: auction repository
- marketRepo: market repository
- borrowerRepo: borrower repository

@returns
- *LiquidationService
*/
func NewLiquidationService(
auctionRepo AuctionRepository,
marketRepo MarketRepository,
borrowerRepo BorrowerRepository,
) *LiquidationService {
	return &LiquidationService{
		auctionRepo:  auctionRepo,
		marketRepo:   marketRepo,
		borrowerRepo: borrowerRepo,
		monitor:      NewMonitor(),
		auction:      NewAuctionManager(),
	}
}

/*
@function CreateAuction

@desc
Creates liquidation auction for market.

@params
- ctx: request context
- marketAddr: market address
- collateral: collateral amount in wei
- debt: debt amount in wei
- startPrice: initial Dutch price
- endPrice: final Dutch price
- duration: auction duration in seconds

@returns
- *models.Auction: created auction
- error: if creation fails
*/
func (s *LiquidationService) CreateAuction(
ctx context.Context,
marketAddr string,
collateral *big.Int,
debt *big.Int,
startPrice *big.Int,
endPrice *big.Int,
duration int64,
) (*models.Auction, error) {
	// Validate market exists
	market, err := s.marketRepo.GetByAddress(ctx, marketAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch market: %w", err)
	}

	if market == nil {
		return nil, appErr.ErrMarketNotFound
	}

	// Validate health factor
	healthOK, err := s.monitor.CheckLiquidationHealth(
		ctx,
		market,
		new(big.Float).SetInt64(int64(market.MinCollateralRatio)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check health: %w", err)
	}

	if !healthOK {
		return nil, appErr.ErrLiquidationNotRequired
	}

	// Create auction
	auction := &models.Auction{
		AuctionID:       int64(time.Now().UnixNano()),
		MarketAddress:   marketAddr,
		CollateralAmount: new(big.Int).Set(collateral),
		DebtAmount:      new(big.Int).Set(debt),
		CurrentPrice:    new(big.Int).Set(startPrice),
		StartTime:       time.Now(),
		EndTime:         time.Now().Add(time.Duration(duration) * time.Second),
		Status:          "active",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.auctionRepo.CreateAuction(ctx, auction); err != nil {
		return nil, fmt.Errorf("failed to create auction: %w", err)
	}

	return auction, nil
}

/*
@function GetAuction

@desc
Retrieves auction by ID.

@params
- ctx: request context
- auctionID: auction ID

@returns
- *models.Auction: auction details
- error: if not found
*/
func (s *LiquidationService) GetAuction(
ctx context.Context,
auctionID string,
) (*models.Auction, error) {
	auction, err := s.auctionRepo.GetByID(ctx, auctionID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch auction: %w", err)
	}

	if auction == nil {
		return nil, appErr.ErrAuctionNotFound
	}

	// Update current price
	currentPrice := s.auction.CalculateDutchPrice(auction)
	auction.CurrentPrice = currentPrice

	return auction, nil
}

/*
@function GetCurrentPrice

@desc
Gets current Dutch auction price.

@params
- ctx: request context
- auctionID: auction ID

@returns
- *big.Int: current price in wei
- error: if auction not found
*/
func (s *LiquidationService) GetCurrentPrice(
ctx context.Context,
auctionID string,
) (*big.Int, error) {
	auction, err := s.GetAuction(ctx, auctionID)
	if err != nil {
		return nil, err
	}

	return auction.CurrentPrice, nil
}

/*
@function PlaceBid

@desc
Places bid on auction.

@params
- ctx: request context
- auctionID: auction ID
- bidder: bidder address
- amount: bid amount

@returns
- error: if bid fails
*/
func (s *LiquidationService) PlaceBid(
	ctx context.Context,
	auctionID string,
	bidder string,
	amount *big.Int,
) error {
	auction, err := s.auctionRepo.GetByID(ctx, auctionID)
	if err != nil {
		return fmt.Errorf("failed to fetch auction: %w", err)
	}

	if auction == nil {
		return appErr.ErrAuctionNotFound
	}

	// Check auction not expired
	if s.auction.IsAuctionExpired(auction) {
		return appErr.ErrAuctionExpired
	}

	// Validate bid amount
	currentPrice := s.auction.CalculateDutchPrice(auction)
	if amount.Cmp(currentPrice) < 0 {
		return appErr.ErrInsufficientBidAmount
	}

	// Update auction with new bid
	auction.HighestBidder = sql.NullString{String: bidder, Valid: true}
	auction.HighestBid = new(big.Int).Set(amount)
	auction.UpdatedAt = time.Now()

	if err := s.auctionRepo.UpdateAuction(ctx, auction); err != nil {
		return fmt.Errorf("failed to update auction: %w", err)
	}

	return nil
}

/*
@function SettleAuction

@desc
Settles completed auction.

@params
- ctx: request context
- auctionID: auction ID

@returns
- error: if settlement fails
*/
func (s *LiquidationService) SettleAuction(
	ctx context.Context,
	auctionID string,
) error {
	auction, err := s.auctionRepo.GetByID(ctx, auctionID)
	if err != nil {
		return fmt.Errorf("failed to fetch auction: %w", err)
	}

	if auction == nil {
		return appErr.ErrAuctionNotFound
	}

	// Check auction has ended
	if !s.auction.IsAuctionExpired(auction) && auction.HighestBid == nil {
		return appErr.ErrAuctionNotEnded
	}

	// Check no existing settlement
	if auction.Status == "settled" {
		return appErr.ErrAuctionAlreadySettled
	}

	// Update auction status
	auction.Status = "settled"
	if auction.SettlementTime.Valid == false {
		auction.SettlementTime = sql.NullTime{Time: time.Now(), Valid: true}
	}
	auction.UpdatedAt = time.Now()

	if err := s.auctionRepo.UpdateAuction(ctx, auction); err != nil {
		return fmt.Errorf("failed to settle auction: %w", err)
	}

	return nil
}/*
@function GetLiquidatableMarkets

@desc
Gets all liquidatable markets.

@params
- ctx: request context

@returns
- []models.Market: markets ready for liquidation
- error: if query fails
*/
func (s *LiquidationService) GetLiquidatableMarkets(
ctx context.Context,
) ([]models.Market, error) {
	allMarkets, err := s.marketRepo.GetAllActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch markets: %w", err)
	}

	liquidatable := []models.Market{}
	for _, market := range allMarkets {
		isLiquidatable := s.monitor.IsMarketLiquidatable(&market)
		if isLiquidatable {
			liquidatable = append(liquidatable, market)
		}
	}

	return liquidatable, nil
}

/*
@function GetHealthFactor

@desc
Gets market health factor.

@params
- ctx: request context
- marketAddr: market address

@returns
- float64: health factor (>1 = healthy)
- error: if calculation fails
*/
func (s *LiquidationService) GetHealthFactor(
ctx context.Context,
marketAddr string,
) (float64, error) {
	market, err := s.marketRepo.GetByAddress(ctx, marketAddr)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch market: %w", err)
	}

	if market == nil {
		return 0, appErr.ErrMarketNotFound
	}

	healthFactor := s.monitor.CalculateHealthFactor(market)
	return healthFactor, nil
}

/*
@function DetectLiquidatable

@desc
Detects all liquidatable positions in market.

@params
- ctx: request context
- marketAddr: market address

@returns
- []map[string]interface{}: liquidatable positions
- error: if detection fails
*/
func (s *LiquidationService) DetectLiquidatable(
	ctx context.Context,
	marketAddr string,
) ([]map[string]interface{}, error) {
	market, err := s.marketRepo.GetByAddress(ctx, marketAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch market: %w", err)
	}

	if market == nil {
		return nil, appErr.ErrMarketNotFound
	}

	var result []map[string]interface{}

	if s.monitor.IsMarketLiquidatable(market) {
		result = append(result, map[string]interface{}{
			"market":        marketAddr,
			"health_factor": s.monitor.CalculateHealthFactor(market),
			"collateral":    market.TotalLiquidity.String(),
			"debt":          market.TotalDebt.String(),
		})
	}

	return result, nil
}

/*
@function CancelAuction

@desc
Cancels active auction.

@params
- ctx: request context
- auctionID: auction ID

@returns
- error: if cancellation fails
*/
func (s *LiquidationService) CancelAuction(
	ctx context.Context,
	auctionID string,
) error {
	auction, err := s.auctionRepo.GetByID(ctx, auctionID)
	if err != nil {
		return fmt.Errorf("failed to fetch auction: %w", err)
	}

	if auction == nil {
		return appErr.ErrAuctionNotFound
	}

	if auction.Status != "active" {
		return appErr.ErrAuctionNotActive
	}

	auction.Status = "cancelled"
	auction.UpdatedAt = time.Now()

	if err := s.auctionRepo.UpdateAuction(ctx, auction); err != nil {
		return fmt.Errorf("failed to cancel auction: %w", err)
	}

	return nil
}