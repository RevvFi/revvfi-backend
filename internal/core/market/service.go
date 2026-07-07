package market

import (
"context"
"fmt"
"math/big"

"github.com/Revvfi/revvfi-backend/internal/models"
appErr "github.com/Revvfi/revvfi-backend/internal/pkg/errors"
)

/*
@file service.go

@desc
Market service handles core lending market business logic.
Manages market creation, state tracking, and health monitoring.

@responsibilities
- Create and initialize markets
- Track market state (active/liquidating/closed)
- Monitor market health factors
- Calculate market metrics (TVL, APR, utilization)
- Validate market operations
*/

/*
@struct MarketService

@desc
Handles lending market business logic and protocol validation.

@dependencies
- MarketRepository: for persistent market data
- OfferRepository: for offer tracking
- PositionRepository: for position tracking
- OracleClient: for collateral valuation
- Logger: for structured logging
*/
type MarketService struct {
	marketRepo     MarketRepository
	offerRepo      OfferRepository
	positionRepo   PositionRepository
	calculator     *Calculator
	validator      *Validator
}

/*
@interface MarketRepository

@desc
Repository for market data access.
*/
type MarketRepository interface {
	CreateMarket(ctx context.Context, market *models.Market) error
	GetByAddress(ctx context.Context, address string) (*models.Market, error)
	ListActive(ctx context.Context, limit, offset int32, borrower string) ([]models.Market, error)
	UpdateMarket(ctx context.Context, market *models.Market) error
}

/*
@interface OfferRepository

@desc
Repository for offer data access.
*/
type OfferRepository interface {
	GetActiveByMarket(ctx context.Context, marketAddr string) ([]models.Offer, error)
	CountActive(ctx context.Context, marketAddr string) (int32, error)
}

/*
@interface PositionRepository

@desc
Repository for position data access.
*/
type PositionRepository interface {
	CountActiveByMarket(ctx context.Context, marketAddr string) (int32, error)
	GetTotalPrincipalByMarket(ctx context.Context, marketAddr string) (*big.Int, error)
}

/*
@function NewMarketService

@desc
Creates new market service with dependencies.

@params
- marketRepo: market repository
- offerRepo: offer repository
- positionRepo: position repository

@returns
- *MarketService
*/
func NewMarketService(
marketRepo MarketRepository,
offerRepo OfferRepository,
positionRepo PositionRepository,
) *MarketService {
	return &MarketService{
		marketRepo:   marketRepo,
		offerRepo:    offerRepo,
		positionRepo: positionRepo,
		calculator:   NewCalculator(),
		validator:    NewValidator(),
	}
}

/*
@function CreateMarket

@desc
Creates new lending market with validation.

@params
- ctx: request context
- borrower: borrower wallet address
- borrowAsset: ERC20 token address to borrow
- collateralAsset: ERC20 collateral token
- collateralOracle: Chainlink oracle address
- minCollateralRatio: minimum collateral ratio in bps
- liquidationThreshold: liquidation threshold in bps

@returns
- *models.Market: created market
- error: if validation or creation fails
*/
func (s *MarketService) CreateMarket(
ctx context.Context,
marketAddr string,
borrower string,
borrowAsset string,
collateralAsset string,
collateralOracle string,
minCollateralRatio int32,
liquidationThreshold int32,
) (*models.Market, error) {
	// Validate inputs
	if err := s.validator.ValidateMarketCreation(
borrower,
borrowAsset,
collateralAsset,
minCollateralRatio,
liquidationThreshold,
); err != nil {
		return nil, fmt.Errorf("market validation failed: %w", err)
	}

	market := &models.Market{
		Address:              marketAddr,
		Borrower:             borrower,
		BorrowAsset:          borrowAsset,
		CollateralAsset:      collateralAsset,
		CollateralOracle:     collateralOracle,
		MinCollateralRatio:   minCollateralRatio,
		LiquidationThreshold: liquidationThreshold,
		TotalPrincipal:       big.NewInt(0),
		TotalAccruedInterest: big.NewInt(0),
		TotalLiquidity:       big.NewInt(0),
		BorrowIndex:          big.NewInt(1000000000000000000), // 1e18
		IsActive:             true,
	}

	if err := s.marketRepo.CreateMarket(ctx, market); err != nil {
		return nil, fmt.Errorf("failed to create market: %w", err)
	}

	return market, nil
}

/*
@function GetMarket

@desc
Retrieves market details by address.

@params
- ctx: request context
- address: market contract address

@returns
- *models.Market: market details
- error: if market not found
*/
func (s *MarketService) GetMarket(ctx context.Context, address string) (*models.Market, error) {
	market, err := s.marketRepo.GetByAddress(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch market: %w", err)
	}

	if market == nil {
		return nil, appErr.ErrMarketNotFound
	}

	return market, nil
}

/*
@function ListMarkets

@desc
Lists active markets with pagination, optionally scoped to a single borrower.

@params
- ctx: request context
- limit: max results
- offset: pagination offset
- borrower: optional borrower address filter ("" = all borrowers)

@returns
- []models.Market: market list
- error: if query fails
*/
func (s *MarketService) ListMarkets(
ctx context.Context,
limit, offset int32,
borrower string,
) ([]models.Market, error) {
	markets, err := s.marketRepo.ListActive(ctx, limit, offset, borrower)
	if err != nil {
		return nil, fmt.Errorf("failed to list markets: %w", err)
	}

	return markets, nil
}

/*
@function CalculateMetrics

@desc
Calculates derived market metrics.

@params
- ctx: request context
- market: market to analyze

@returns
- map[string]interface{}: calculated metrics
- error: if calculation fails
*/
func (s *MarketService) CalculateMetrics(
ctx context.Context,
market *models.Market,
) (map[string]interface{}, error) {
	// Get market state
	activeOffers, err := s.offerRepo.GetActiveByMarket(ctx, market.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch offers: %w", err)
	}

	totalLiquidity := big.NewInt(0)
	for _, offer := range activeOffers {
		totalLiquidity.Add(totalLiquidity, offer.RemainingAmount)
	}

	// Calculate metrics
	utilization := s.calculator.CalculateUtilization(
market.TotalDebt,
totalLiquidity,
)

	avgAPR := s.calculator.CalculateAverageAPR(activeOffers)

	return map[string]interface{}{
		"tvl":              totalLiquidity.String(),
		"utilization_rate": utilization,
		"avg_apr":          avgAPR,
		"active_positions": 0, // TODO: from position repo
	}, nil
}
