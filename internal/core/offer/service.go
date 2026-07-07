package offer

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/Revvfi/revvfi-backend/internal/models"
	appErr "github.com/Revvfi/revvfi-backend/internal/pkg/errors"
)

/*
@file service.go

@desc
Offer service handles lender liquidity offers.
Manages offer lifecycle, matching, and quote generation.

@responsibilities
- Validate offer parameters
- Match offers with borrow requests
- Calculate effective borrowing rates
- Generate offer quotes
*/

/*
@struct OfferService

@desc
Handles lender offer business logic.

@dependencies
- OfferRepository: for offer storage
- Matcher: for offer matching logic
- Calculator: for quote calculations
*/
type OfferService struct {
	offerRepo OfferRepository
	matcher   *Matcher
	calc      *Calculator
}

/*
@interface OfferRepository

@desc
Repository for offer data access.
*/
type OfferRepository interface {
	CreateOffer(ctx context.Context, offer *models.Offer) error
	GetByID(ctx context.Context, offerID int64) (*models.Offer, error)
	GetActiveByMarket(ctx context.Context, marketAddr string, limit, offset int32) ([]models.Offer, error)
	GetByLender(ctx context.Context, lender string, page, pageSize int32) ([]models.Offer, int64, error)
	UpdateOffer(ctx context.Context, offer *models.Offer) error
	CancelOffer(ctx context.Context, offerID int64) error
}

/*
@function NewOfferService

@desc
Creates new offer service.

@params
- offerRepo: offer repository

@returns
- *OfferService
*/
func NewOfferService(offerRepo OfferRepository) *OfferService {
	return &OfferService{
		offerRepo: offerRepo,
		matcher:   NewMatcher(),
		calc:      NewCalculator(),
	}
}

/*
@function CreateOffer

@desc
Creates new lender offer.

@params
- ctx: request context
- lender: lender wallet address
- market: market address
- amount: liquidity amount in wei
- apr: annual percentage rate in bps
- seniority: 0=Senior, 1=Junior
- expiryDays: offer expiration in days

@returns
- *models.Offer: created offer
- error: if validation or creation fails
*/
func (s *OfferService) CreateOffer(
ctx context.Context,
offerID int64,
lender string,
market string,
amount *big.Int,
apr int32,
seniority int16,
expiryDays int32,
) (*models.Offer, error) {
	if err := ValidateOfferCreation(lender, amount, apr, seniority); err != nil {
		return nil, fmt.Errorf("offer validation failed: %w", err)
	}

	now := time.Now()
	expiry := now.AddDate(0, 0, int(expiryDays))
	if expiryDays <= 0 {
		expiry = now.AddDate(0, 1, 0) // default 30 days
	}

	offer := &models.Offer{
		OfferID:         offerID,
		Lender:          lender,
		MarketAddress:   market,
		Amount:          new(big.Int).Set(amount),
		RemainingAmount: new(big.Int).Set(amount),
		APR:             apr,
		Seniority:       seniority,
		Status:          "active",
		CreatedAt:       now,
		Expiry:          expiry,
	}

	if err := s.offerRepo.CreateOffer(ctx, offer); err != nil {
		return nil, fmt.Errorf("failed to create offer: %w", err)
	}

	return offer, nil
}

/*
@function GetOffer

@desc
Retrieves offer by ID.

@params
- ctx: request context
- offerID: offer identifier

@returns
- *models.Offer: offer details
- error: if not found
*/
func (s *OfferService) GetOffer(ctx context.Context, offerID int64) (*models.Offer, error) {
	offer, err := s.offerRepo.GetByID(ctx, offerID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch offer: %w", err)
	}

	if offer == nil {
		return nil, appErr.ErrOfferNotFound
	}

	return offer, nil
}

/*
@function GetMarketOffers

@desc
Gets all active offers for a market.

@params
- ctx: request context
- marketAddr: market address
- limit: max results
- offset: pagination offset

@returns
- []models.Offer: active offers
- error: if query fails
*/
func (s *OfferService) GetMarketOffers(
ctx context.Context,
marketAddr string,
limit, offset int32,
) ([]models.Offer, error) {
	offers, err := s.offerRepo.GetActiveByMarket(ctx, marketAddr, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch offers: %w", err)
	}

	return offers, nil
}

/*
@function GetLenderOffers

@desc
Lists every offer (any market, any status) placed by a specific lender -
the read path behind "My Offers" in the frontend's Portfolio/Lend pages.

@params
- ctx: request context
- lender: lender's wallet address
- page: 1-indexed page number
- pageSize: results per page

@returns
- []models.Offer: the lender's offers
- error: if query fails
*/
func (s *OfferService) GetLenderOffers(
ctx context.Context,
lender string,
page, pageSize int32,
) ([]models.Offer, error) {
	offers, _, err := s.offerRepo.GetByLender(ctx, lender, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch lender offers: %w", err)
	}

	return offers, nil
}

/*
@function CalculateQuote

@desc
Calculates optimal offer matching for borrow request.

@params
- ctx: request context
- marketAddr: market address
- borrowAmount: amount to borrow in wei
- maxAPR: maximum acceptable APR in bps

@returns
- map[string]interface{}: quote details
- error: if calculation fails
*/
func (s *OfferService) CalculateQuote(
ctx context.Context,
marketAddr string,
borrowAmount *big.Int,
maxAPR int32,
) (map[string]interface{}, error) {
	// Get active offers
	offers, err := s.offerRepo.GetActiveByMarket(ctx, marketAddr, 100, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch offers: %w", err)
	}

	if len(offers) == 0 {
		return nil, appErr.ErrInsufficientLiquidity
	}

	// Match offers
	matched := s.matcher.MatchOffers(offers, borrowAmount, maxAPR)
	if len(matched) == 0 {
		return nil, appErr.ErrInsufficientLiquidity
	}

	// Calculate metrics
	avgAPR := s.calc.CalculateWeightedAPR(matched)
	totalMatched := big.NewInt(0)
	for _, o := range matched {
		totalMatched.Add(totalMatched, o.RemainingAmount)
	}

	return map[string]interface{}{
		"matched_offers":    matched,
		"total_liquidity":   totalMatched.String(),
		"weighted_apr":      avgAPR,
		"is_fully_matched":  totalMatched.Cmp(borrowAmount) >= 0,
	}, nil
}

/*
@function CancelOffer

@desc
Cancels active offer.

@params
- ctx: request context
- offerID: offer to cancel
- lender: lender address (for authorization)

@returns
- error: if cancellation fails
*/
func (s *OfferService) CancelOffer(
ctx context.Context,
offerID int64,
lender string,
) error {
	offer, err := s.offerRepo.GetByID(ctx, offerID)
	if err != nil {
		return fmt.Errorf("failed to fetch offer: %w", err)
	}

	if offer == nil {
		return appErr.ErrOfferNotFound
	}

	if offer.Lender != lender {
		return appErr.ErrUnauthorized
	}

	if offer.Status != "active" && offer.Status != "partially_filled" {
		return appErr.ErrOfferNotActive
	}

	if err := s.offerRepo.CancelOffer(ctx, offerID); err != nil {
		return fmt.Errorf("failed to cancel offer: %w", err)
	}

	return nil
}

/*
@function ValidateOfferCreation

@desc
Validates offer creation parameters.

@params
- lender: lender address
- amount: offer amount
- apr: annual percentage rate
- seniority: seniority level

@returns
- error: if validation fails
*/
func ValidateOfferCreation(lender string, amount *big.Int, apr int32, seniority int16) error {
	if lender == "" || len(lender) != 42 {
		return appErr.ErrInvalidAddress
	}

	if amount == nil || amount.Sign() <= 0 {
		return appErr.ErrInvalidAmount
	}

	if apr <= 0 || apr > 5000 {
		return appErr.ErrInvalidAPR
	}

	if seniority != 0 && seniority != 1 {
		return appErr.ErrInvalidInput
	}

	return nil
}
