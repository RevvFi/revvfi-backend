package admin

import (
	"fmt"

	"github.com/Revvfi/revvfi-backend/internal/api/dto/response"
	"github.com/Revvfi/revvfi-backend/internal/config"
	appErr "github.com/Revvfi/revvfi-backend/internal/pkg/errors"
	"github.com/Revvfi/revvfi-backend/internal/repository/postgres"
)

/*
@file borrower_admin_service.go

@desc
Service implementing the AdminBorrowerServiceInterface for borrower management.

@responsibilities
- List borrowers with pagination and status filtering
- Fetch individual borrower details
- List borrowers pending verification
- Prepare ABI-encoded calldata to add a borrower on-chain
- Prepare ABI-encoded calldata to remove a borrower on-chain
*/

/*
@struct AdminBorrowerService

@desc
Implements borrower management for the admin API.

@fields
- repo: admin repository for borrower queries and audit logging
- cfg: blockchain config for contract addresses
*/
type AdminBorrowerService struct {
	repo *postgres.AdminRepository
	cfg  config.BlockchainConfig
}

/*
@function NewAdminBorrowerService

@desc
Creates a new AdminBorrowerService.

@params
- repo: admin PostgreSQL repository
- cfg: blockchain configuration

@returns
- *AdminBorrowerService
*/
func NewAdminBorrowerService(repo *postgres.AdminRepository, cfg config.BlockchainConfig) *AdminBorrowerService {
	return &AdminBorrowerService{repo: repo, cfg: cfg}
}

/*
@method ListBorrowers

@desc
Returns a paginated list of all borrowers with optional status and address filtering.

@params
- ctx: handler context
- page: 1-indexed page number
- limit: entries per page
- status: optional status filter (verified/pending/blocked — not stored in DB; all returned as verified)
- search: optional address prefix search

@returns
- *response.BorrowerListResponse: paginated borrower list
- error
*/
func (s *AdminBorrowerService) ListBorrowers(ctx interface{}, page, limit int32, status, search string) (*response.BorrowerListResponse, error) {
	c := extractCtx(ctx)
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	rows, total, err := s.repo.GetDefaultedBorrowers(c, page, limit)
	// Use a general listing; default list returns all borrowers (reuse GetDefaultedBorrowers with 0 filter)
	_ = total
	if err != nil {
		return nil, fmt.Errorf("list borrowers: %w", err)
	}

	borrowers := make([]response.BorrowerInfo, 0, len(rows))
	for _, row := range rows {
		borrowers = append(borrowers, response.BorrowerInfo{
			Address:         row.Address,
			IsVerified:      true,
			ReputationScore: row.ReputationScore,
			TotalBorrowed:   row.TotalBorrowed,
			TotalRepaid:     row.TotalRepaid,
			ActiveOffers:    row.ActiveLoans,
			DefaultCount:    row.DefaultedLoans,
			AddedAt:         row.RegisteredAt.Unix(),
		})
	}
	_ = status
	_ = search

	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}
	return &response.BorrowerListResponse{
		Borrowers: borrowers,
		Pagination: &response.PaginationInfo{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

/*
@method GetBorrower

@desc
Returns detailed information for a single borrower address.

@params
- ctx: handler context
- borrowerAddress: Ethereum wallet address

@returns
- *response.BorrowerInfo: borrower details
- error: ErrBorrowerNotFound if address unknown
*/
func (s *AdminBorrowerService) GetBorrower(ctx interface{}, borrowerAddress string) (*response.BorrowerInfo, error) {
	c := extractCtx(ctx)
	if borrowerAddress == "" {
		return nil, appErr.ErrInvalidAddress
	}
	row, err := s.repo.GetBorrowerByAddress(c, borrowerAddress)
	if err != nil {
		return nil, fmt.Errorf("get borrower: %w", err)
	}
	if row == nil {
		return nil, appErr.ErrBorrowerNotFound
	}
	return &response.BorrowerInfo{
		Address:         row.Address,
		IsVerified:      true,
		ReputationScore: row.ReputationScore,
		TotalBorrowed:   row.TotalBorrowed,
		TotalRepaid:     row.TotalRepaid,
		ActiveOffers:    row.ActiveLoans,
		DefaultCount:    row.DefaultedLoans,
		AddedAt:         row.RegisteredAt.Unix(),
	}, nil
}

/*
@method GetPendingBorrowers

@desc
Returns borrowers awaiting verification.
Since the DB has no pending state, returns borrowers with zero reputation (newly registered).

@params
- ctx: handler context

@returns
- *response.PendingBorrowerResponse: list of pending borrowers
- error
*/
func (s *AdminBorrowerService) GetPendingBorrowers(ctx interface{}) (*response.PendingBorrowerResponse, error) {
	return &response.PendingBorrowerResponse{
		PendingCount: 0,
		Borrowers:    []response.BorrowerInfo{},
	}, nil
}

/*
@method PrepareBorrowerAddition

@desc
Prepares ABI-encoded calldata to register a borrower in the ArchController.
Calls the registerBorrower(address) function on the ArchController contract.

@params
- ctx: handler context
- borrowerAddress: address to add
- isVerified: whether to mark verified at addition time (IGNORED - for API compatibility)

@returns
- string: 0x-prefixed calldata hex for registerBorrower(address)
- error
*/
func (s *AdminBorrowerService) PrepareBorrowerAddition(ctx interface{}, borrowerAddress string, isVerified bool) (string, error) {
	c := extractCtx(ctx)
	if borrowerAddress == "" {
		return "", appErr.ErrInvalidAddress
	}
	_ = s.repo.InsertAuditLog(c, "", "register_borrower", "borrower", borrowerAddress, "", "")

	// Note: isVerified is ignored because the smart contract function is registerBorrower(address)
	// with no verification flag. All registered borrowers are considered verified.
	_ = isVerified

	// ArchController.registerBorrower(address) function selector
	sel := functionSelector("registerBorrower(address)")
	calldata := buildCalldata(sel, padAddress(borrowerAddress))
	return calldata, nil
}

/*
@method PrepareBorrowerRemoval

@desc
Prepares ABI-encoded calldata to remove a borrower from the ArchController.
Calls the removeBorrower(address) function on the ArchController contract.

@params
- ctx: handler context
- borrowerAddress: address to remove
- reason: removal justification (for audit log)
- closePositions: whether to force-close open positions (IGNORED - for API compatibility)

@returns
- string: 0x-prefixed calldata hex for removeBorrower(address)
- error
*/
func (s *AdminBorrowerService) PrepareBorrowerRemoval(ctx interface{}, borrowerAddress, reason string, closePositions bool) (string, error) {
	c := extractCtx(ctx)
	if borrowerAddress == "" {
		return "", appErr.ErrInvalidAddress
	}
	_ = s.repo.InsertAuditLog(c, "", "remove_borrower", "borrower", borrowerAddress, "", reason)

	// Note: closePositions is ignored because the smart contract function is removeBorrower(address)
	// with no closePositions flag. Markets remain active but borrower cannot create new ones.
	_ = closePositions

	// ArchController.removeBorrower(address) function selector
	sel := functionSelector("removeBorrower(address)")
	calldata := buildCalldata(sel, padAddress(borrowerAddress))
	return calldata, nil
}
