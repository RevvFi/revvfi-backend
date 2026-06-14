// internal/indexer/handlers/fee_handler.go
package handlers

import (
	"context"
	"log"

	"github.com/Revvfi/revvfi-backend/internal/indexer/types"
	"github.com/Revvfi/revvfi-backend/internal/repository/postgres"
)

/*@
 * FeeHandler
 * @desc Handles fee-related events from RevvFiFactory
 * @events
 *   - FeeUpdated: Deployment fee amount changed
 *   - DeploymentFeeUpdated: Alternative name for fee update (depending on contract version)
 *
 * @state
 *   - Tracks fee changes for audit and reporting
 */
type FeeHandler struct {
	eventRepo *postgres.EventRepository
}

/*@
 * NewFeeHandler
 * @desc Creates a new fee handler instance
 * @param eventRepo Repository for database operations
 * @returns Configured FeeHandler
 */
func NewFeeHandler(eventRepo *postgres.EventRepository) *FeeHandler {
	return &FeeHandler{eventRepo: eventRepo}
}

/*@
 * Handle
 * @desc Routes fee events to appropriate handler methods
 * @param ctx Context for cancellation and deadlines
 * @param event The decoded event from blockchain
 * @param blockNum Block number where event occurred
 * @returns error if handling fails
 */
func (h *FeeHandler) Handle(ctx context.Context, event interface{}, blockNum uint64) error {
	switch e := event.(type) {
	case *types.FeeUpdatedEvent:
		return h.handleFeeUpdated(ctx, e, blockNum)
	case *types.DeploymentFeeUpdatedEvent:
		return h.handleDeploymentFeeUpdated(ctx, e, blockNum)
	}
	return nil
}

/*@
 * handleFeeUpdated
 * @desc Processes FeeUpdated event
 * @param event Decoded FeeUpdated event
 * @param blockNum Block number where fee was updated
 *
 * @state_transition
 *   - Records fee change for audit trail
 *   - New fee applies to future market deployments
 */
func (h *FeeHandler) handleFeeUpdated(ctx context.Context, event *types.FeeUpdatedEvent, blockNum uint64) error {
	log.Printf(" FeeUpdated: OldFee=%s ETH, NewFee=%s ETH",
		event.OldFee.String(), event.NewFee.String())
	
	// TODO: Save to fee_audit_log table if tracking is needed
	// return h.eventRepo.RecordFeeChange(ctx, event.OldFee, event.NewFee, blockNum)
	
	return nil
}

/*@
 * handleDeploymentFeeUpdated
 * @desc Processes DeploymentFeeUpdated event (alternative name)
 * @param event Decoded DeploymentFeeUpdated event
 * @param blockNum Block number where fee was updated
 */
func (h *FeeHandler) handleDeploymentFeeUpdated(ctx context.Context, event *types.DeploymentFeeUpdatedEvent, blockNum uint64) error {
	log.Printf(" DeploymentFeeUpdated: OldFee=%s ETH, NewFee=%s ETH",
		event.OldFee.String(), event.NewFee.String())
	
	return nil
}