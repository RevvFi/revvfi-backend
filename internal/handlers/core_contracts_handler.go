// internal/indexer/handlers/core_contracts_handler.go
package handlers

import (
	"context"
	"log"

	"github.com/Revvfi/revvfi-backend/internal/indexer/types"
	"github.com/Revvfi/revvfi-backend/internal/repository/postgres"
)

/*@
 * CoreContractsHandler
 * @desc Handles CoreContractsSet event from RevvFiFactory
 * @event
 *   - CoreContractsSet: Core protocol contracts configured (one-time setup)
 *
 * @importance
 *   This event is emitted only once during protocol initialization
 *   Stores the core contract addresses that the factory will use
 */
type CoreContractsHandler struct {
	eventRepo *postgres.EventRepository
}

/*@
 * NewCoreContractsHandler
 * @desc Creates a new core contracts handler instance
 * @param eventRepo Repository for database operations
 * @returns Configured CoreContractsHandler
 */
func NewCoreContractsHandler(eventRepo *postgres.EventRepository) *CoreContractsHandler {
	return &CoreContractsHandler{eventRepo: eventRepo}
}

/*@
 * Handle
 * @desc Processes CoreContractsSet event
 * @param ctx Context for cancellation and deadlines
 * @param event The decoded event from blockchain
 * @param blockNum Block number where event occurred
 * @returns error if handling fails
 */
func (h *CoreContractsHandler) Handle(ctx context.Context, event interface{}, blockNum uint64) error {
	e, ok := event.(*types.CoreContractsSetEvent)
	if !ok {
		return nil
	}

	log.Printf("CoreContractsSet: ArchController=%s, PositionNFT=%s, Liquidator=%s, ReputationRegistry=%s",
		e.ArchController.Hex(), e.PositionNFT.Hex(), e.Liquidator.Hex(), e.ReputationRegistry.Hex())

	// TODO: Save to a protocol_config table if needed for tracking
	// For now, just log as these addresses are already in config
	
	return nil
}