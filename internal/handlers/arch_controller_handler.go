// internal/indexer/handlers/arch_controller_handler.go
package handlers

import (
	"context"
	"log"

	"github.com/Revvfi/revvfi-backend/internal/indexer/types"
	"github.com/Revvfi/revvfi-backend/internal/repository/postgres"
)

/*@
 * ArchControllerHandler
 * @desc Handles ArchController events from RevvFiArchController and RevvFiFactory
 * @events
 *   - ArchControllerUpdateRequested: Update requested (2-day timelock starts)
 *   - ArchControllerUpdated: Update executed after timelock
 *   - BorrowerAdded: New borrower authorized
 *   - BorrowerRemoved: Borrower unauthorized
 *   - ControllerAdded: New controller contract added
 *   - ControllerRemoved: Controller contract removed
 *   - MarketRegistered: Market registered in ArchController
 *   - MarketUnregistered: Market unregistered
 *   - OwnerUpdated: ArchController ownership transferred
 *   - TimelockUpdated: Timelock duration changed
 */
type ArchControllerHandler struct {
	eventRepo *postgres.EventRepository
}

/*@
 * NewArchControllerHandler
 * @desc Creates a new arch controller handler instance
 * @param eventRepo Repository for database operations
 * @returns Configured ArchControllerHandler
 */
func NewArchControllerHandler(eventRepo *postgres.EventRepository) *ArchControllerHandler {
	return &ArchControllerHandler{eventRepo: eventRepo}
}

/*@
 * Handle
 * @desc Routes arch controller events to appropriate handler methods
 * @param ctx Context for cancellation and deadlines
 * @param event The decoded event from blockchain
 * @param blockNum Block number where event occurred
 * @returns error if handling fails
 */
func (h *ArchControllerHandler) Handle(ctx context.Context, event interface{}, blockNum uint64) error {
	switch e := event.(type) {
	case *types.ArchControllerUpdateRequestedEvent:
		return h.handleArchControllerUpdateRequested(ctx, e, blockNum)
	case *types.ArchControllerUpdatedEvent:
		return h.handleArchControllerUpdated(ctx, e, blockNum)
	case *types.ArchBorrowerAddedEvent:
		return h.handleBorrowerAdded(ctx, e, blockNum)
	case *types.ArchBorrowerRemovedEvent:
		return h.handleBorrowerRemoved(ctx, e, blockNum)
	case *types.ControllerAddedEvent:
		return h.handleControllerAdded(ctx, e, blockNum)
	case *types.ControllerRemovedEvent:
		return h.handleControllerRemoved(ctx, e, blockNum)
	case *types.MarketRegisteredEvent:
		return h.handleMarketRegistered(ctx, e, blockNum)
	case *types.MarketUnregisteredEvent:
		return h.handleMarketUnregistered(ctx, e, blockNum)
	case *types.ArchOwnerUpdatedEvent:
		return h.handleOwnerUpdated(ctx, e, blockNum)
	case *types.ArchTimelockUpdatedEvent:
		return h.handleTimelockUpdated(ctx, e, blockNum)
	}
	return nil
}

func (h *ArchControllerHandler) handleArchControllerUpdateRequested(ctx context.Context, event *types.ArchControllerUpdateRequestedEvent, blockNum uint64) error {
	log.Printf("ArchControllerUpdateRequested: NewArchController=%s (2-day timelock started)", event.NewArchController.Hex())
	return nil
}

func (h *ArchControllerHandler) handleArchControllerUpdated(ctx context.Context, event *types.ArchControllerUpdatedEvent, blockNum uint64) error {
	log.Printf("ArchControllerUpdated: Old=%s, New=%s", event.OldArchController.Hex(), event.NewArchController.Hex())
	return nil
}

func (h *ArchControllerHandler) handleBorrowerAdded(ctx context.Context, event *types.ArchBorrowerAddedEvent, blockNum uint64) error {
	log.Printf("BorrowerAdded: Borrower=%s, BlockNumber=%d", event.Borrower.Hex(), blockNum)
	// TODO: Update borrower status in borrowers table
	return nil
}

func (h *ArchControllerHandler) handleBorrowerRemoved(ctx context.Context, event *types.ArchBorrowerRemovedEvent, blockNum uint64) error {
	log.Printf("BorrowerRemoved: Borrower=%s, BlockNumber=%d", event.Borrower.Hex(), blockNum)
	// TODO: Update borrower status in borrowers table
	return nil
}

func (h *ArchControllerHandler) handleControllerAdded(ctx context.Context, event *types.ControllerAddedEvent, blockNum uint64) error {
	log.Printf("ControllerAdded: ControllerFactory=%s", event.ControllerFactory.Hex())
	return nil
}

func (h *ArchControllerHandler) handleControllerRemoved(ctx context.Context, event *types.ControllerRemovedEvent, blockNum uint64) error {
	log.Printf("ControllerRemoved: Controller=%s, BlockNumber=%d", event.Controller.Hex(), blockNum)
	return nil
}

func (h *ArchControllerHandler) handleMarketRegistered(ctx context.Context, event *types.MarketRegisteredEvent, blockNum uint64) error {
	log.Printf("MarketRegistered: Market=%s, BlockNumber=%d", event.Market.Hex(), blockNum)
	return nil
}

func (h *ArchControllerHandler) handleMarketUnregistered(ctx context.Context, event *types.MarketUnregisteredEvent, blockNum uint64) error {
	log.Printf("MarketUnregistered: Market=%s, BlockNumber=%d", event.Market.Hex(), blockNum)
	return nil
}

func (h *ArchControllerHandler) handleOwnerUpdated(ctx context.Context, event *types.ArchOwnerUpdatedEvent, blockNum uint64) error {
	log.Printf("ArchController OwnerUpdated: OldOwner=%s, NewOwner=%s", event.OldOwner.Hex(), event.NewOwner.Hex())
	return nil
}

func (h *ArchControllerHandler) handleTimelockUpdated(ctx context.Context, event *types.ArchTimelockUpdatedEvent, blockNum uint64) error {
	log.Printf("TimelockUpdated: OldDelay=%d seconds, NewDelay=%d seconds", event.OldDelay.Int64(), event.NewDelay.Int64())
	return nil
}