package borrower

import (
	"context"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Revvfi/revvfi-backend/internal/models"
	appErr "github.com/Revvfi/revvfi-backend/internal/pkg/errors"
)

/*
@file request_service.go

@desc
BorrowerRequestService implements the off-chain "request to become a
borrower" review queue. registerBorrower() on RevvFiArchController is
onlyOwner, so an arbitrary wallet can never self-register on-chain — this
service lets any signed-in wallet ask, and lets an admin review the queue
from the frontend (approve by sending the real on-chain tx from their own
wallet, or reject with no on-chain effect at all).
*/

/*
@interface BorrowerRequestRepository

@desc
Repository for borrower request persistence.
*/
type BorrowerRequestRepository interface {
	Create(ctx context.Context, wallet string) (*models.BorrowerRequest, error)
	GetLatestByWallet(ctx context.Context, wallet string) (*models.BorrowerRequest, error)
	GetByID(ctx context.Context, id int64) (*models.BorrowerRequest, error)
	List(ctx context.Context, status string) ([]*models.BorrowerRequest, error)
	Reject(ctx context.Context, id int64, adminWallet, note string) error
}

/*
@interface ExistingBorrowerChecker

@desc
Narrow view of BorrowerRepository — just enough to check whether a wallet
is already a registered borrower before accepting a new request for it.
*/
type ExistingBorrowerChecker interface {
	GetByAddress(ctx context.Context, address string) (*models.Borrower, error)
}

/*
@struct BorrowerRequestService
*/
type BorrowerRequestService struct {
	requestRepo BorrowerRequestRepository
	borrowerRepo ExistingBorrowerChecker
}

/*
@function NewBorrowerRequestService
*/
func NewBorrowerRequestService(requestRepo BorrowerRequestRepository, borrowerRepo ExistingBorrowerChecker) *BorrowerRequestService {
	return &BorrowerRequestService{requestRepo: requestRepo, borrowerRepo: borrowerRepo}
}

/*
@function Create

@desc
Submits a new borrower access request for the given wallet.

@params
- ctx: request context
- wallet: the requesting wallet address (must come from an authenticated
  session — never trust a client-supplied address for identity)

@returns
- *models.BorrowerRequest: the created request
- error: ErrInvalidAddress, ErrBorrowerAlreadyRegistered, or
  ErrBorrowerRequestAlreadyPending
*/
func (s *BorrowerRequestService) Create(ctx context.Context, wallet string) (*models.BorrowerRequest, error) {
	if !common.IsHexAddress(wallet) {
		return nil, appErr.ErrInvalidAddress
	}

	existing, err := s.borrowerRepo.GetByAddress(ctx, wallet)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing borrower: %w", err)
	}
	if existing != nil {
		return nil, appErr.ErrBorrowerAlreadyRegistered
	}

	req, err := s.requestRepo.Create(ctx, wallet)
	if err != nil {
		if errors.Is(err, appErr.ErrDuplicateEntry) {
			return nil, appErr.ErrBorrowerRequestAlreadyPending
		}
		return nil, fmt.Errorf("failed to create borrower request: %w", err)
	}
	return req, nil
}

/*
@function GetMine

@desc
Returns the latest borrower request for a wallet, or nil if it has never
submitted one.
*/
func (s *BorrowerRequestService) GetMine(ctx context.Context, wallet string) (*models.BorrowerRequest, error) {
	if !common.IsHexAddress(wallet) {
		return nil, appErr.ErrInvalidAddress
	}
	return s.requestRepo.GetLatestByWallet(ctx, wallet)
}

/*
@function List

@desc
Lists borrower requests for the admin review queue, optionally filtered by
status ("pending", "approved", "rejected"; empty = all).
*/
func (s *BorrowerRequestService) List(ctx context.Context, status string) ([]*models.BorrowerRequest, error) {
	if status != "" && status != "pending" && status != "approved" && status != "rejected" {
		return nil, appErr.ErrInvalidInput
	}
	return s.requestRepo.List(ctx, status)
}

/*
@function Reject

@desc
Rejects a pending borrower request. Pure off-chain decision — no on-chain
transaction is involved, since nothing was ever granted on-chain.

@params
- ctx: request context
- id: request ID
- adminWallet: the admin wallet making the decision (for the audit trail)
- note: optional rejection reason

@returns
- error: ErrBorrowerRequestNotFound, ErrBorrowerRequestNotPending
*/
func (s *BorrowerRequestService) Reject(ctx context.Context, id int64, adminWallet, note string) error {
	existing, err := s.requestRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to fetch borrower request: %w", err)
	}
	if existing == nil {
		return appErr.ErrBorrowerRequestNotFound
	}
	return s.requestRepo.Reject(ctx, id, adminWallet, note)
}
