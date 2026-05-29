package position

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
Position service handles lender position lifecycle.
Manages position creation, valuation, and settlement.

@responsibilities
- Track lender positions (NFTs)
- Calculate position values (principal + interest)
- Manage position settlement
- Aggregate portfolio data
*/

/*
@struct PositionService

@desc
Handles lender position management.

@dependencies
- PositionRepository: position data access
- Valulator: position valuation calculations
*/
type PositionService struct {
	positionRepo PositionRepository
	valuator     *Valuator
}

/*
@interface PositionRepository

@desc
Repository for position data access.
*/
type PositionRepository interface {
	CreatePosition(ctx context.Context, position *models.Position) error
	GetByTokenID(ctx context.Context, tokenID int64) (*models.Position, error)
	GetByLender(ctx context.Context, lender string, limit, offset int32) ([]models.Position, error)
	UpdatePosition(ctx context.Context, position *models.Position) error
	CountActiveByLender(ctx context.Context, lender string) (int32, error)
	GetTotalValueByLender(ctx context.Context, lender string) (*big.Int, error)
}

/*
@function NewPositionService

@desc
Creates new position service.

@params
- positionRepo: position repository

@returns
- *PositionService
*/
func NewPositionService(positionRepo PositionRepository) *PositionService {
	return &PositionService{
		positionRepo: positionRepo,
		valuator:     NewValuator(),
	}
}

/*
@function CreatePosition

@desc
Creates new lender position.

@params
- ctx: request context
- tokenID: NFT token ID
- lender: position owner
- market: market address
- principal: principal amount in wei
- apr: annual percentage rate
- seniority: seniority level

@returns
- *models.Position: created position
- error: if creation fails
*/
func (s *PositionService) CreatePosition(
ctx context.Context,
tokenID int64,
lender string,
market string,
principal *big.Int,
apr int32,
seniority int16,
) (*models.Position, error) {
	position := &models.Position{
		TokenID:         tokenID,
		Lender:          lender,
		MarketAddress:   market,
		Principal:       new(big.Int).Set(principal),
		CurrentPrincipal: new(big.Int).Set(principal),
		AccruedInterest: big.NewInt(0),
		ClaimableAmount: big.NewInt(0),
		APR:             apr,
		Seniority:       seniority,
		Status:          "active",
		IsActive:        true,
	}

	if err := s.positionRepo.CreatePosition(ctx, position); err != nil {
		return nil, fmt.Errorf("failed to create position: %w", err)
	}

	return position, nil
}

/*
@function GetPosition

@desc
Retrieves position by token ID.

@params
- ctx: request context
- tokenID: position token ID

@returns
- *models.Position: position details
- error: if not found
*/
func (s *PositionService) GetPosition(
ctx context.Context,
tokenID int64,
) (*models.Position, error) {
	position, err := s.positionRepo.GetByTokenID(ctx, tokenID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch position: %w", err)
	}

	if position == nil {
		return nil, appErr.ErrPositionNotFound
	}

	return position, nil
}

/*
@function GetLenderPositions

@desc
Retrieves all positions for a lender.

@params
- ctx: request context
- lender: lender address
- limit: max results
- offset: pagination offset

@returns
- []models.Position: lender positions
- error: if query fails
*/
func (s *PositionService) GetLenderPositions(
ctx context.Context,
lender string,
limit, offset int32,
) ([]models.Position, error) {
	positions, err := s.positionRepo.GetByLender(ctx, lender, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch positions: %w", err)
	}

	return positions, nil
}

/*
@function GetPortfolioSummary

@desc
Calculates aggregated portfolio summary.

@params
- ctx: request context
- lender: lender address

@returns
- map[string]interface{}: portfolio metrics
- error: if calculation fails
*/
func (s *PositionService) GetPortfolioSummary(
ctx context.Context,
lender string,
) (map[string]interface{}, error) {
	positions, err := s.positionRepo.GetByLender(ctx, lender, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch positions: %w", err)
	}

	totalSupplied := big.NewInt(0)
	totalValue := big.NewInt(0)
	totalInterest := big.NewInt(0)
	activeCount := int32(0)
	settledCount := int32(0)

	for _, pos := range positions {
		totalSupplied.Add(totalSupplied, pos.Principal)
		// Calculate current value: Principal + Accrued Interest
		currentValue := new(big.Int).Add(pos.CurrentPrincipal, pos.AccruedInterest)
		totalValue.Add(totalValue, currentValue)
		totalInterest.Add(totalInterest, pos.AccruedInterest)

		if pos.IsActive {
			activeCount++
		} else {
			settledCount++
		}
	}

	avgAPR := s.valuator.CalculatePortfolioAPR(positions)

	return map[string]interface{}{
		"total_supplied":     totalSupplied.String(),
		"total_value":        totalValue.String(),
		"earned_interest":    totalInterest.String(),
		"avg_apr":            avgAPR,
		"active_positions":   activeCount,
		"settled_positions":  settledCount,
		"position_count":     len(positions),
	}, nil
}

/*
@function UpdatePositionValue

@desc
Updates position value with accrued interest.

@params
- ctx: request context
- tokenID: position token ID
- newAccruedInterest: new accrued interest amount

@returns
- error: if update fails
*/
func (s *PositionService) UpdatePositionValue(
ctx context.Context,
tokenID int64,
newAccruedInterest *big.Int,
) error {
	position, err := s.positionRepo.GetByTokenID(ctx, tokenID)
	if err != nil {
		return fmt.Errorf("failed to fetch position: %w", err)
	}

	if position == nil {
		return appErr.ErrPositionNotFound
	}

	position.AccruedInterest = newAccruedInterest
	// Update position in repository
	if err := s.positionRepo.UpdatePosition(ctx, position); err != nil {
		return fmt.Errorf("failed to update position: %w", err)
	}

	return nil
}

/*
@function SettlePosition

@desc
Marks position as settled and claimable.

@params
- ctx: request context
- tokenID: position token ID

@returns
- error: if settlement fails
*/
func (s *PositionService) SettlePosition(
ctx context.Context,
tokenID int64,
) error {
	position, err := s.positionRepo.GetByTokenID(ctx, tokenID)
	if err != nil {
		return fmt.Errorf("failed to fetch position: %w", err)
	}

	if position == nil {
		return appErr.ErrPositionNotFound
	}

	if position.IsSettled {
		return appErr.ErrPositionAlreadySettled
	}

	position.IsSettled = true
	position.IsActive = false
	position.Status = "settled"
	// ClaimableAmount = CurrentPrincipal + AccruedInterest
	position.ClaimableAmount = new(big.Int).Add(position.CurrentPrincipal, position.AccruedInterest)

	if err := s.positionRepo.UpdatePosition(ctx, position); err != nil {
		return fmt.Errorf("failed to settle position: %w", err)
	}

	return nil
}
