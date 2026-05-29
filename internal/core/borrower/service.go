package borrower

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
Borrower service handles reputation tracking and risk assessment.
Maintains borrower profiles and calculates trust scores.

@responsibilities
- Register borrowers
- Track loan statistics
- Calculate reputation scores
- Assess borrower risk levels
*/

/*
@struct BorrowerService

@desc
Handles borrower profile and reputation management.

@dependencies
- BorrowerRepository: borrower data access
- RepCalculator: reputation calculation logic
*/
type BorrowerService struct {
	borrowerRepo BorrowerRepository
	calc         *RepCalculator
}

/*
@interface BorrowerRepository

@desc
Repository for borrower data access.
*/
type BorrowerRepository interface {
	CreateBorrower(ctx context.Context, borrower *models.Borrower) error
	GetByAddress(ctx context.Context, address string) (*models.Borrower, error)
	UpdateBorrower(ctx context.Context, borrower *models.Borrower) error
}

/*
@function NewBorrowerService

@desc
Creates new borrower service.

@params
- borrowerRepo: borrower repository

@returns
- *BorrowerService
*/
func NewBorrowerService(borrowerRepo BorrowerRepository) *BorrowerService {
	return &BorrowerService{
		borrowerRepo: borrowerRepo,
		calc:         NewRepCalculator(),
	}
}

/*
@function RegisterBorrower

@desc
Registers new borrower with initial reputation.

@params
- ctx: request context
- address: borrower wallet address

@returns
- *models.Borrower: created borrower profile
- error: if registration fails
*/
func (s *BorrowerService) RegisterBorrower(
ctx context.Context,
address string,
) (*models.Borrower, error) {
	// Check if already exists
	existing, err := s.borrowerRepo.GetByAddress(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing borrower: %w", err)
	}

	if existing != nil {
		return existing, nil
	}

	// Create new borrower with starting reputation
	borrower := &models.Borrower{
		Address:         address,
		TotalBorrowed:   big.NewInt(0),
		TotalRepaid:     big.NewInt(0),
		OutstandingDebt: big.NewInt(0),
		TotalLoans:      0,
		SuccessfulLoans: 0,
		DefaultedLoans:  0,
		ActiveLoans:     0,
		ReputationScore: 500, // Starting score
		RiskLabel:       "B",
		SuccessRate:     0,
	}

	if err := s.borrowerRepo.CreateBorrower(ctx, borrower); err != nil {
		return nil, fmt.Errorf("failed to create borrower: %w", err)
	}

	return borrower, nil
}

/*
@function GetBorrower

@desc
Retrieves borrower profile.

@params
- ctx: request context
- address: borrower address

@returns
- *models.Borrower: borrower profile
- error: if not found
*/
func (s *BorrowerService) GetBorrower(
ctx context.Context,
address string,
) (*models.Borrower, error) {
	borrower, err := s.borrowerRepo.GetByAddress(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch borrower: %w", err)
	}

	if borrower == nil {
		return nil, appErr.ErrBorrowerNotFound
	}

	return borrower, nil
}

/*
@function RecordSuccessfulRepayment

@desc
Records successful loan repayment and updates reputation.

@params
- ctx: request context
- address: borrower address

@returns
- error: if update fails
*/
func (s *BorrowerService) RecordSuccessfulRepayment(
ctx context.Context,
address string,
) error {
	borrower, err := s.borrowerRepo.GetByAddress(ctx, address)
	if err != nil {
		return fmt.Errorf("failed to fetch borrower: %w", err)
	}

	if borrower == nil {
		return appErr.ErrBorrowerNotFound
	}

	// Update statistics
	borrower.SuccessfulLoans++
	if borrower.ActiveLoans > 0 {
		borrower.ActiveLoans--
	}

	// Recalculate reputation
	borrower.ReputationScore = s.calc.CalculateReputation(
borrower.TotalLoans,
borrower.SuccessfulLoans,
borrower.DefaultedLoans,
)
	borrower.RiskLabel = s.calc.CalculateRiskLabel(borrower.ReputationScore)

	if err := s.borrowerRepo.UpdateBorrower(ctx, borrower); err != nil {
		return fmt.Errorf("failed to update borrower: %w", err)
	}

	return nil
}

/*
@function RecordDefault

@desc
Records loan default and penalizes reputation.

@params
- ctx: request context
- address: borrower address

@returns
- error: if update fails
*/
func (s *BorrowerService) RecordDefault(
ctx context.Context,
address string,
) error {
	borrower, err := s.borrowerRepo.GetByAddress(ctx, address)
	if err != nil {
		return fmt.Errorf("failed to fetch borrower: %w", err)
	}

	if borrower == nil {
		return appErr.ErrBorrowerNotFound
	}

	// Update statistics
	borrower.DefaultedLoans++
	if borrower.ActiveLoans > 0 {
		borrower.ActiveLoans--
	}

	// Recalculate reputation
	borrower.ReputationScore = s.calc.CalculateReputation(
borrower.TotalLoans,
borrower.SuccessfulLoans,
borrower.DefaultedLoans,
)
	borrower.RiskLabel = s.calc.CalculateRiskLabel(borrower.ReputationScore)

	if err := s.borrowerRepo.UpdateBorrower(ctx, borrower); err != nil {
		return fmt.Errorf("failed to update borrower: %w", err)
	}

	return nil
}

/*
@function GetRiskAssessment

@desc
Returns borrower risk assessment.

@params
- ctx: request context
- address: borrower address

@returns
- map[string]interface{}: risk metrics
- error: if assessment fails
*/
func (s *BorrowerService) GetRiskAssessment(
ctx context.Context,
address string,
) (map[string]interface{}, error) {
	borrower, err := s.GetBorrower(ctx, address)
	if err != nil {
		return nil, err
	}

	healthFactor := s.calc.CalculateHealthFactor(borrower)
	riskLevel := s.calc.GetRiskLevel(borrower.ReputationScore)

	return map[string]interface{}{
		"reputation_score": borrower.ReputationScore,
		"risk_label":       borrower.RiskLabel,
		"risk_level":       riskLevel,
		"health_factor":    healthFactor,
		"success_rate":     borrower.SuccessRate,
		"default_rate":     float64(borrower.DefaultedLoans) / float64(borrower.TotalLoans) * 100,
	}, nil
}
