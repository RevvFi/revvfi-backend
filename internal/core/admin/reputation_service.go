package admin

import (
	"fmt"

	"github.com/Revvfi/revvfi-backend/internal/api/dto/response"
	"github.com/Revvfi/revvfi-backend/internal/config"
	appErr "github.com/Revvfi/revvfi-backend/internal/pkg/errors"
	"github.com/Revvfi/revvfi-backend/internal/repository/postgres"
)

/*
@file reputation_service.go

@desc
Service for admin borrower reputation management.

@responsibilities
- Return detailed borrower reputation profile
- Prepare ABI-encoded calldata to update borrower reputation on-chain
- Return paginated list of defaulted borrowers
*/

/*
@struct AdminReputationService

@desc
Implements admin borrower reputation management.

@fields
- repo: admin repository for borrower and audit data
- cfg: blockchain configuration for the reputation registry address
*/
type AdminReputationService struct {
	repo *postgres.AdminRepository
	cfg  config.BlockchainConfig
}

/*
@function NewAdminReputationService

@desc
Creates a new AdminReputationService.

@params
- repo: admin PostgreSQL repository
- cfg: blockchain configuration

@returns
- *AdminReputationService
*/
func NewAdminReputationService(repo *postgres.AdminRepository, cfg config.BlockchainConfig) *AdminReputationService {
	return &AdminReputationService{repo: repo, cfg: cfg}
}

/*
@method GetBorrowerReputation

@desc
Returns the full reputation profile for a borrower address.

@params
- ctx: handler context
- address: borrower wallet address

@returns
- *response.BorrowerReputation: detailed reputation data
- error: ErrBorrowerNotFound if address is unknown
*/
func (s *AdminReputationService) GetBorrowerReputation(ctx interface{}, address string) (*response.BorrowerReputation, error) {
	c := extractCtx(ctx)
	if address == "" {
		return nil, appErr.ErrInvalidAddress
	}
	row, err := s.repo.GetBorrowerByAddress(c, address)
	if err != nil {
		return nil, fmt.Errorf("get borrower: %w", err)
	}
	if row == nil {
		return nil, appErr.ErrBorrowerNotFound
	}

	var lastActivity int64
	if row.LastActivity.Valid {
		lastActivity = row.LastActivity.Time.Unix()
	}
	return &response.BorrowerReputation{
		Address:         row.Address,
		ReputationScore: row.ReputationScore,
		RiskLabel:       row.RiskLabel,
		SuccessRate:     row.SuccessRate,
		TotalLoans:      row.TotalLoans,
		ActiveLoans:     row.ActiveLoans,
		DefaultedLoans:  row.DefaultedLoans,
		TotalBorrowed:   row.TotalBorrowed,
		LastActivity:    lastActivity,
	}, nil
}

/*
@method PrepareSetReputation

@desc
Updates the borrower reputation score in the DB and prepares ABI-encoded calldata
for the ReputationRegistry contract.

@params
- ctx: handler context
- address: borrower wallet address
- newScore: new reputation score (0-1000)
- adminAddress: admin wallet performing the action
- reason: justification for the override

@returns
- *response.TransactionPreparation: encoded calldata targeting the ReputationRegistry
- error
*/
func (s *AdminReputationService) PrepareSetReputation(ctx interface{}, address string, newScore int32, adminAddress, reason string) (*response.TransactionPreparation, error) {
	c := extractCtx(ctx)
	if address == "" {
		return nil, appErr.ErrInvalidAddress
	}
	if newScore < 0 || newScore > 1000 {
		return nil, appErr.ErrInvalidInput
	}

	riskLabel := scoreToRiskLabel(newScore)
	if err := s.repo.UpdateBorrowerReputation(c, address, newScore, riskLabel); err != nil {
		return nil, fmt.Errorf("update borrower reputation: %w", err)
	}
	_ = s.repo.InsertAuditLog(c, adminAddress, "update_reputation", "borrower", address, "", reason)

	sel := functionSelector("setReputationScore(address,uint256)")
	calldata := buildCalldata(sel, padAddress(address), padUint256FromInt(int64(newScore)))
	return &response.TransactionPreparation{
		Target:      s.cfg.ReputationRegistryAddress,
		TxData:      calldata,
		Value:       "0",
		Description: fmt.Sprintf("Set reputation score of %s to %d (%s)", address, newScore, riskLabel),
	}, nil
}

/*
@method GetDefaultedBorrowers

@desc
Returns a paginated list of borrowers with at least one defaulted loan.

@params
- ctx: handler context
- page: 1-indexed page number
- limit: entries per page

@returns
- *response.DefaultedBorrowersResponse: paginated defaulted borrower list
- error
*/
func (s *AdminReputationService) GetDefaultedBorrowers(ctx interface{}, page, limit int32) (*response.DefaultedBorrowersResponse, error) {
	c := extractCtx(ctx)
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	rows, total, err := s.repo.GetDefaultedBorrowers(c, page, limit)
	if err != nil {
		return nil, fmt.Errorf("get defaulted borrowers: %w", err)
	}

	borrowers := make([]response.BorrowerInfo, 0, len(rows))
	for _, row := range rows {
		borrowers = append(borrowers, response.BorrowerInfo{
			Address:         row.Address,
			TotalBorrowed:   row.TotalBorrowed,
			TotalRepaid:     row.TotalRepaid,
			ActiveOffers:    row.ActiveLoans,
			DefaultCount:    row.DefaultedLoans,
			ReputationScore: row.ReputationScore,
			AddedAt:         row.RegisteredAt.Unix(),
		})
	}

	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}
	return &response.DefaultedBorrowersResponse{
		Borrowers: borrowers,
		Pagination: &response.PaginationInfo{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
		TotalDefaulted: total,
	}, nil
}

/*
@function scoreToRiskLabel

@desc
Maps a reputation score (0-1000) to a risk tier label.
*/
func scoreToRiskLabel(score int32) string {
	switch {
	case score >= 950:
		return "AAA"
	case score >= 850:
		return "AA"
	case score >= 700:
		return "A"
	case score >= 550:
		return "B"
	case score >= 350:
		return "C"
	default:
		return "D"
	}
}
