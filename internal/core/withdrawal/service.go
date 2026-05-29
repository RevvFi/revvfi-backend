package withdrawal

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"time"

	"github.com/Revvfi/revvfi-backend/internal/models"
	appErr "github.com/Revvfi/revvfi-backend/internal/pkg/errors"
)

/*
@file service.go

@desc
Withdrawal service manages epoch-based withdrawal requests.
Handles request lifecycle, epoch processing, and fulfillment tracking.

@responsibilities
- Create withdrawal requests
- Process withdrawal epochs
- Calculate fulfillment amounts
- Manage withdrawal queue
*/

/*
@struct WithdrawalService

@desc
Manages withdrawal requests and epoch processing.

@dependencies
- WithdrawalRepository: withdrawal data access
- PositionRepository: position data access
- EpochCalculator: epoch calculation logic
*/
type WithdrawalService struct {
	withdrawalRepo WithdrawalRepository
	positionRepo   PositionRepository
	calculator     *EpochCalculator
}

/*
@interface WithdrawalRepository

@desc
Repository for withdrawal data access.
*/
type WithdrawalRepository interface {
	CreateRequest(ctx context.Context, request *models.WithdrawalRequest) error
	GetRequestByID(ctx context.Context, requestID int64) (*models.WithdrawalRequest, error)
	GetRequestsByLender(ctx context.Context, lender string, limit, offset int32) ([]models.WithdrawalRequest, error)
	UpdateRequest(ctx context.Context, request *models.WithdrawalRequest) error
	GetActiveRequests(ctx context.Context, epochNumber int64) ([]models.WithdrawalRequest, error)
	CreateEpoch(ctx context.Context, epoch *models.WithdrawalEpoch) error
	GetEpochByNumber(ctx context.Context, epochNumber int64) (*models.WithdrawalEpoch, error)
	UpdateEpoch(ctx context.Context, epoch *models.WithdrawalEpoch) error
	GetCurrentEpoch(ctx context.Context) (*models.WithdrawalEpoch, error)
}

/*
@interface PositionRepository

@desc
Repository for position data access.
*/
type PositionRepository interface {
	GetByTokenID(ctx context.Context, tokenID int64) (*models.Position, error)
	UpdatePosition(ctx context.Context, position *models.Position) error
}

/*
@function NewWithdrawalService

@desc
Creates new withdrawal service.

@params
- withdrawalRepo: withdrawal repository
- positionRepo: position repository

@returns
- *WithdrawalService
*/
func NewWithdrawalService(
	withdrawalRepo WithdrawalRepository,
	positionRepo PositionRepository,
) *WithdrawalService {
	return &WithdrawalService{
		withdrawalRepo: withdrawalRepo,
		positionRepo:   positionRepo,
		calculator:     NewEpochCalculator(),
	}
}

/*
@function RequestWithdrawal

@desc
Creates new withdrawal request for a position.

@params
- ctx: request context
- lender: lender address
- positionID: position token ID
- amount: withdrawal amount in wei

@returns
- *models.WithdrawalRequest: created request
- error: if creation fails
*/
func (s *WithdrawalService) RequestWithdrawal(
	ctx context.Context,
	lender string,
	positionID int64,
	amount *big.Int,
) (*models.WithdrawalRequest, error) {
	// Validate position exists
	position, err := s.positionRepo.GetByTokenID(ctx, positionID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch position: %w", err)
	}

	if position == nil {
		return nil, appErr.ErrPositionNotFound
	}

	// Verify ownership
	if position.Lender != lender {
		return nil, appErr.ErrUnauthorized
	}

	// Validate amount
	if amount.Cmp(big.NewInt(0)) <= 0 {
		return nil, appErr.ErrInvalidAmount
	}

	if amount.Cmp(position.ClaimableAmount) > 0 {
		return nil, fmt.Errorf("insufficient claimable amount: %w", appErr.ErrInvalidAmount)
	}

	// Get current epoch
	currentEpoch, err := s.withdrawalRepo.GetCurrentEpoch(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current epoch: %w", err)
	}

	// Create request for next epoch
	request := &models.WithdrawalRequest{
		Lender:        lender,
		PositionID:    positionID,
		RequestedAmount: new(big.Int).Set(amount),
		Status:        "pending",
		CreatedAt:     time.Now(),
		EpochNumber:   currentEpoch.EpochNumber + 1,
		FulfilledAmount: big.NewInt(0),
	}

	if err := s.withdrawalRepo.CreateRequest(ctx, request); err != nil {
		return nil, fmt.Errorf("failed to create withdrawal request: %w", err)
	}

	return request, nil
}

/*
@function CancelWithdrawalRequest

@desc
Cancels pending withdrawal request.

@params
- ctx: request context
- requestID: withdrawal request ID
- lender: lender address (for authorization)

@returns
- error: if cancellation fails
*/
func (s *WithdrawalService) CancelWithdrawalRequest(
	ctx context.Context,
	requestID int64,
	lender string,
) error {
	request, err := s.withdrawalRepo.GetRequestByID(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to fetch request: %w", err)
	}

	if request == nil {
		return appErr.ErrWithdrawalNotFound
	}

	// Verify ownership
	if request.Lender != lender {
		return appErr.ErrUnauthorized
	}

	// Check if cancellable
	if request.Status != "pending" {
		return fmt.Errorf("cannot cancel non-pending request: %w", appErr.ErrInvalidInput)
	}

	request.Status = "cancelled"
	if err := s.withdrawalRepo.UpdateRequest(ctx, request); err != nil {
		return fmt.Errorf("failed to cancel request: %w", err)
	}

	return nil
}

/*
@function GetWithdrawalRequests

@desc
Retrieves withdrawal requests for a lender.

@params
- ctx: request context
- lender: lender address
- limit: max results
- offset: pagination offset

@returns
- []models.WithdrawalRequest: lender requests
- error: if query fails
*/
func (s *WithdrawalService) GetWithdrawalRequests(
	ctx context.Context,
	lender string,
	limit, offset int32,
) ([]models.WithdrawalRequest, error) {
	requests, err := s.withdrawalRepo.GetRequestsByLender(ctx, lender, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch requests: %w", err)
	}

	return requests, nil
}

/*
@function ProcessEpoch

@desc
Processes withdrawal epoch.

@params
- ctx: request context
- epochNumber: epoch to process

@returns
- error: if processing fails
*/
func (s *WithdrawalService) ProcessEpoch(
	ctx context.Context,
	epochNumber int64,
) error {
	epoch, err := s.withdrawalRepo.GetEpochByNumber(ctx, epochNumber)
	if err != nil {
		return fmt.Errorf("failed to fetch epoch: %w", err)
	}

	if epoch == nil {
		return fmt.Errorf("epoch not found: %w", appErr.ErrInternal)
	}

	if epoch.Status != "pending" {
		return fmt.Errorf("epoch already processed: %w", appErr.ErrInvalidInput)
	}

	// Get active requests for epoch
	requests, err := s.withdrawalRepo.GetActiveRequests(ctx, epochNumber)
	if err != nil {
		return fmt.Errorf("failed to fetch requests: %w", err)
	}

	totalRequested := big.NewInt(0)
	for _, req := range requests {
		totalRequested.Add(totalRequested, req.RequestedAmount)
	}

	// Calculate fulfillment
	fulfillmentRatio := s.calculator.CalculateFulfillmentRatio(totalRequested, epoch.TotalRequested)

	// Update requests with fulfilled amounts
	currentTime := time.Now().Unix()
	for _, req := range requests {
		fulfilled := new(big.Int).Mul(req.RequestedAmount, big.NewInt(int64(fulfillmentRatio*1000)))
		fulfilled.Div(fulfilled, big.NewInt(1000))

		req.Status = "fulfilled"
		req.FulfilledAmount = fulfilled
		req.FulfillmentTime = sql.NullTime{Time: time.Unix(currentTime, 0), Valid: true}

		if err := s.withdrawalRepo.UpdateRequest(ctx, &req); err != nil {
			return fmt.Errorf("failed to update request: %w", err)
		}
	}

	// Mark epoch as completed
	epoch.Status = "completed"
	epoch.ProcessedAt = sql.NullTime{Time: time.Now(), Valid: true}

	if err := s.withdrawalRepo.UpdateEpoch(ctx, epoch); err != nil {
		return fmt.Errorf("failed to update epoch: %w", err)
	}

	return nil
}

/*
@function GetCurrentEpoch

@desc
Retrieves current withdrawal epoch.

@params
- ctx: request context

@returns
- *models.WithdrawalEpoch: current epoch
- error: if fetch fails
*/
func (s *WithdrawalService) GetCurrentEpoch(
	ctx context.Context,
) (*models.WithdrawalEpoch, error) {
	epoch, err := s.withdrawalRepo.GetCurrentEpoch(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch current epoch: %w", err)
	}

	if epoch == nil {
		return nil, appErr.ErrInternal
	}

	return epoch, nil
}
/*
@function GetEpochStatus

@desc
Retrieves status of specific epoch.

@params
- ctx: request context
- epochNumber: epoch number

@returns
- map[string]interface{}: epoch status data
- error: if fetch fails
*/
func (s *WithdrawalService) GetEpochStatus(
	ctx context.Context,
	epochNumber int64,
) (map[string]interface{}, error) {
	epoch, err := s.withdrawalRepo.GetEpochByNumber(ctx, epochNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch epoch: %w", err)
	}

	if epoch == nil {
		return nil, appErr.ErrInternal
	}

	requests, err := s.withdrawalRepo.GetActiveRequests(ctx, epochNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch requests: %w", err)
	}

	totalFulfilled := big.NewInt(0)
	for _, req := range requests {
		totalFulfilled.Add(totalFulfilled, req.FulfilledAmount)
	}

	return map[string]interface{}{
		"epoch_number":       epoch.EpochNumber,
		"status":             epoch.Status,
		"total_requested":    epoch.TotalRequested.String(),
		"total_fulfilled":    epoch.TotalFulfilled.String(),
		"request_count":      len(requests),
		"start_time":         epoch.StartTime.Unix(),
		"end_time":           epoch.EndTime.Unix(),
	}, nil
}
