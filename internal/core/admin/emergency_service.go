package admin

import (
	"fmt"

	"github.com/Revvfi/revvfi-backend/internal/api/dto/response"
	"github.com/Revvfi/revvfi-backend/internal/config"
	"github.com/Revvfi/revvfi-backend/internal/repository/postgres"
)

/*
@file emergency_service.go

@desc
Service for emergency protocol controls.

@responsibilities
- Prepare calldata to pause all protocol markets
- Prepare calldata to unpause all protocol markets
- Prepare calldata to drain accumulated fees to a recipient
- Record all emergency actions in the audit log
*/

/*
@struct AdminEmergencyService

@desc
Implements admin emergency control operations.

@fields
- repo: admin repository for audit log writes and market status updates
- cfg: blockchain configuration for contract addresses
*/
type AdminEmergencyService struct {
	repo *postgres.AdminRepository
	cfg  config.BlockchainConfig
}

/*
@function NewAdminEmergencyService

@desc
Creates a new AdminEmergencyService.

@params
- repo: admin PostgreSQL repository
- cfg: blockchain configuration

@returns
- *AdminEmergencyService
*/
func NewAdminEmergencyService(repo *postgres.AdminRepository, cfg config.BlockchainConfig) *AdminEmergencyService {
	return &AdminEmergencyService{repo: repo, cfg: cfg}
}

/*
@method PrepareEmergencyPause

@desc
Marks all non-closed markets as inactive in the DB and prepares ABI-encoded
calldata to call the emergency pause function on the ArchController.

@params
- ctx: handler context
- adminAddress: admin wallet address performing the action
- reason: justification for the emergency pause

@returns
- *response.TransactionPreparation: encoded calldata targeting the ArchController
- error
*/
func (s *AdminEmergencyService) PrepareEmergencyPause(ctx interface{}, adminAddress, reason string) (*response.TransactionPreparation, error) {
	c := extractCtx(ctx)
	if err := s.repo.SetAllMarketsActive(c, false); err != nil {
		return nil, fmt.Errorf("pause all markets: %w", err)
	}
	_ = s.repo.InsertAuditLog(c, adminAddress, "emergency_pause", "protocol", s.cfg.ArchControllerAddress, "", reason)

	sel := functionSelector("emergencyPause()")
	calldata := buildCalldata(sel)
	return &response.TransactionPreparation{
		Target:      s.cfg.ArchControllerAddress,
		TxData:      calldata,
		Value:       "0",
		Description: fmt.Sprintf("Emergency pause of all markets — %s", reason),
	}, nil
}

/*
@method PrepareEmergencyUnpause

@desc
Re-activates all markets in the DB and prepares ABI-encoded calldata to
call the emergency unpause function on the ArchController.

@params
- ctx: handler context
- adminAddress: admin wallet address performing the action
- reason: justification for the unpause

@returns
- *response.TransactionPreparation: encoded calldata targeting the ArchController
- error
*/
func (s *AdminEmergencyService) PrepareEmergencyUnpause(ctx interface{}, adminAddress, reason string) (*response.TransactionPreparation, error) {
	c := extractCtx(ctx)
	if err := s.repo.SetAllMarketsActive(c, true); err != nil {
		return nil, fmt.Errorf("unpause all markets: %w", err)
	}
	_ = s.repo.InsertAuditLog(c, adminAddress, "emergency_unpause", "protocol", s.cfg.ArchControllerAddress, "", reason)

	sel := functionSelector("emergencyUnpause()")
	calldata := buildCalldata(sel)
	return &response.TransactionPreparation{
		Target:      s.cfg.ArchControllerAddress,
		TxData:      calldata,
		Value:       "0",
		Description: fmt.Sprintf("Emergency unpause of all markets — %s", reason),
	}, nil
}

/*
@method PrepareDrainFees

@desc
Prepares ABI-encoded calldata to drain accumulated protocol fees to a specified recipient.

@params
- ctx: handler context
- recipient: Ethereum address to receive the drained fees
- adminAddress: admin wallet address performing the action
- reason: justification for the drain

@returns
- *response.TransactionPreparation: encoded calldata targeting the ArchController
- error
*/
func (s *AdminEmergencyService) PrepareDrainFees(ctx interface{}, recipient, adminAddress, reason string) (*response.TransactionPreparation, error) {
	c := extractCtx(ctx)
	_ = s.repo.InsertAuditLog(c, adminAddress, "drain_fees", "protocol", s.cfg.ArchControllerAddress, "", reason)

	sel := functionSelector("drainFees(address)")
	calldata := buildCalldata(sel, padAddress(recipient))
	return &response.TransactionPreparation{
		Target:      s.cfg.ArchControllerAddress,
		TxData:      calldata,
		Value:       "0",
		Description: fmt.Sprintf("Drain fees to %s — %s", recipient, reason),
	}, nil
}
