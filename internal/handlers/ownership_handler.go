// internal/indexer/handlers/ownership_handler.go
package handlers

import (
	"context"
	"log"

	"github.com/Revvfi/revvfi-backend/internal/indexer/types"
	"github.com/Revvfi/revvfi-backend/internal/repository/postgres"
)

/*@
 * OwnershipHandler
 * @desc Handles OwnershipTransferred event from RevvFiFactory
 * @event
 *   - OwnershipTransferred: Factory ownership transferred to new address
 *
 * @security
 *   Only the current owner can transfer ownership
 *   Critical for tracking who can upgrade protocol contracts
 */
type OwnershipHandler struct {
	eventRepo *postgres.EventRepository
}

/*@
 * NewOwnershipHandler
 * @desc Creates a new ownership handler instance
 * @param eventRepo Repository for database operations
 * @returns Configured OwnershipHandler
 */
func NewOwnershipHandler(eventRepo *postgres.EventRepository) *OwnershipHandler {
	return &OwnershipHandler{eventRepo: eventRepo}
}

/*@
 * Handle
 * @desc Processes OwnershipTransferred event
 * @param ctx Context for cancellation and deadlines
 * @param event The decoded event from blockchain
 * @param blockNum Block number where event occurred
 * @returns error if handling fails
 */
func (h *OwnershipHandler) Handle(ctx context.Context, event interface{}, blockNum uint64) error {
	e, ok := event.(*types.OwnershipTransferredEvent)
	if !ok {
		return nil
	}

	log.Printf("OwnershipTransferred: PreviousOwner=%s, NewOwner=%s",
		e.PreviousOwner.Hex(), e.NewOwner.Hex())

	// TODO: Update factory_owner in protocol_config table
	
	return nil
}