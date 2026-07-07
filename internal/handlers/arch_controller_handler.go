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
 * @desc Handles ArchController events from RevvFiArchController
 * @events
 *   - ArchControllerUpdateRequested: Two-day timelock started for controller migration
 *   - ArchControllerUpdated: Controller migration executed after timelock
 *   - BorrowerAdded: New borrower authorized (whitelisted) in protocol
 *   - BorrowerRemoved: Borrower deauthorized (de-whitelisted) from protocol
 *   - ControllerAdded: New controller contract registered
 *   - ControllerRemoved: Controller contract deregistered
 *   - ControllerFactoryAdded: New factory contract registered
 *   - ControllerFactoryRemoved: Factory contract deregistered
 *   - MarketAdded: Market registered in ArchController (activates market record)
 *   - MarketRemoved: Market deregistered from ArchController (deactivates market record)
 *   - MarketRegistered: Market registered in NFT contract (PositionNFT event)
 *   - MarketUnregistered: Market unregistered from NFT contract
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
	case *types.ControllerFactoryAddedEvent:
		return h.handleControllerFactoryAdded(ctx, e, blockNum)
	case *types.ControllerFactoryRemovedEvent:
		return h.handleControllerFactoryRemoved(ctx, e, blockNum)
	case *types.MarketAddedEvent:
		return h.handleMarketAdded(ctx, e, blockNum)
	case *types.MarketRemovedEvent:
		return h.handleMarketRemoved(ctx, e, blockNum)
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

/*@
 * handleArchControllerUpdateRequested
 * @desc Logs the start of a two-day timelock for an ArchController migration
 * @param event Decoded ArchControllerUpdateRequested event
 * @param blockNum Block number where request was made
 */
func (h *ArchControllerHandler) handleArchControllerUpdateRequested(ctx context.Context, event *types.ArchControllerUpdateRequestedEvent, blockNum uint64) error {
	log.Printf("ArchControllerUpdateRequested: NewArchController=%s Block=%d (2-day timelock started)",
		event.NewArchController.Hex(), blockNum)
	return nil
}

/*@
 * handleArchControllerUpdated
 * @desc Logs the completion of an ArchController migration after the timelock
 * @param event Decoded ArchControllerUpdated event
 * @param blockNum Block number where migration was executed
 */
func (h *ArchControllerHandler) handleArchControllerUpdated(ctx context.Context, event *types.ArchControllerUpdatedEvent, blockNum uint64) error {
	log.Printf("ArchControllerUpdated: Old=%s New=%s Block=%d",
		event.OldArchController.Hex(), event.NewArchController.Hex(), blockNum)
	return nil
}

/*@
 * handleBorrowerAdded
 * @desc Sets is_whitelisted=true for a borrower when authorized by ArchController
 * @param event Decoded ArchBorrowerAdded event
 * @param blockNum Block number where borrower was added
 * @note Creates a borrower record if one does not already exist (reputation score 500 / B).
 * @note Also auto-resolves any pending off-chain borrower_requests row for this
 *       wallet to "approved" — best-effort: a failure here must not fail the
 *       whitelist update itself, since on-chain state is already correct.
 */
func (h *ArchControllerHandler) handleBorrowerAdded(ctx context.Context, event *types.ArchBorrowerAddedEvent, blockNum uint64) error {
	log.Printf("BorrowerAdded: Borrower=%s Block=%d", event.Borrower.Hex(), blockNum)
	if err := h.eventRepo.UpdateBorrowerWhitelistStatus(ctx, event.Borrower.Hex(), true); err != nil {
		return err
	}
	if err := h.eventRepo.ResolveApprovedBorrowerRequest(ctx, event.Borrower.Hex()); err != nil {
		log.Printf("BorrowerAdded: failed to auto-resolve pending request for %s: %v", event.Borrower.Hex(), err)
	}
	return nil
}

/*@
 * handleBorrowerRemoved
 * @desc Sets is_whitelisted=false for a borrower when deauthorized by ArchController
 * @param event Decoded ArchBorrowerRemoved event
 * @param blockNum Block number where borrower was removed
 */
func (h *ArchControllerHandler) handleBorrowerRemoved(ctx context.Context, event *types.ArchBorrowerRemovedEvent, blockNum uint64) error {
	log.Printf("BorrowerRemoved: Borrower=%s Block=%d", event.Borrower.Hex(), blockNum)
	return h.eventRepo.UpdateBorrowerWhitelistStatus(ctx, event.Borrower.Hex(), false)
}

/*@
 * handleControllerAdded
 * @desc Inserts a controller record when a new controller is registered in ArchController
 * @param event Decoded ControllerAdded event (ControllerFactory field holds the address)
 * @param blockNum Block number of the registration
 */
func (h *ArchControllerHandler) handleControllerAdded(ctx context.Context, event *types.ControllerAddedEvent, blockNum uint64) error {
	log.Printf("ControllerAdded: Address=%s Block=%d", event.ControllerFactory.Hex(), blockNum)
	return h.eventRepo.AddController(ctx, event.ControllerFactory.Hex(), blockNum)
}

/*@
 * handleControllerRemoved
 * @desc Sets removed_at on a controller record when deregistered from ArchController
 * @param event Decoded ControllerRemoved event
 * @param blockNum Block number of the removal
 */
func (h *ArchControllerHandler) handleControllerRemoved(ctx context.Context, event *types.ControllerRemovedEvent, blockNum uint64) error {
	log.Printf("ControllerRemoved: Address=%s Block=%d", event.Controller.Hex(), blockNum)
	return h.eventRepo.RemoveController(ctx, event.Controller.Hex())
}

/*@
 * handleControllerFactoryAdded
 * @desc Inserts a controller factory record when a factory is registered in ArchController
 * @param event Decoded ControllerFactoryAdded event
 * @param blockNum Block number of the registration
 */
func (h *ArchControllerHandler) handleControllerFactoryAdded(ctx context.Context, event *types.ControllerFactoryAddedEvent, blockNum uint64) error {
	log.Printf("ControllerFactoryAdded: Factory=%s Block=%d", event.Factory.Hex(), blockNum)
	return h.eventRepo.AddControllerFactory(ctx, event.Factory.Hex(), blockNum)
}

/*@
 * handleControllerFactoryRemoved
 * @desc Sets removed_at on a controller factory record when deregistered from ArchController
 * @param event Decoded ControllerFactoryRemoved event
 * @param blockNum Block number of the removal
 */
func (h *ArchControllerHandler) handleControllerFactoryRemoved(ctx context.Context, event *types.ControllerFactoryRemovedEvent, blockNum uint64) error {
	log.Printf("ControllerFactoryRemoved: Factory=%s Block=%d", event.Factory.Hex(), blockNum)
	return h.eventRepo.RemoveControllerFactory(ctx, event.Factory.Hex())
}

/*@
 * handleMarketAdded
 * @desc Sets a market as active when it is registered in the ArchController
 * @param event Decoded MarketAdded event with market and controller addresses
 * @param blockNum Block number where market was added
 * @note MarketDeployed creates the initial market record; MarketAdded activates it in
 *       the protocol registry (is_active=true, is_liquidating=false).
 */
func (h *ArchControllerHandler) handleMarketAdded(ctx context.Context, event *types.MarketAddedEvent, blockNum uint64) error {
	log.Printf("MarketAdded: Market=%s Controller=%s Block=%d",
		event.Market.Hex(), event.Controller.Hex(), blockNum)
	return h.eventRepo.UpdateMarketStatus(ctx, event.Market.Hex(), true, false)
}

/*@
 * handleMarketRemoved
 * @desc Deactivates a market when it is deregistered from the ArchController
 * @param event Decoded MarketRemoved event
 * @param blockNum Block number where market was removed
 * @note Sets is_active=false and is_liquidating=false; does not close the market
 *       (MarketClosedEvent is responsible for permanent closure).
 */
func (h *ArchControllerHandler) handleMarketRemoved(ctx context.Context, event *types.MarketRemovedEvent, blockNum uint64) error {
	log.Printf("MarketRemoved: Market=%s Block=%d", event.Market.Hex(), blockNum)
	return h.eventRepo.UpdateMarketStatus(ctx, event.Market.Hex(), false, false)
}

/*@
 * handleMarketRegistered
 * @desc Logs MarketRegistered event emitted by PositionNFT when a market is registered
 * @param event Decoded MarketRegistered event
 * @param blockNum Block number where registration occurred
 */
func (h *ArchControllerHandler) handleMarketRegistered(ctx context.Context, event *types.MarketRegisteredEvent, blockNum uint64) error {
	log.Printf("MarketRegistered (PositionNFT): Market=%s Block=%d", event.Market.Hex(), blockNum)
	return nil
}

/*@
 * handleMarketUnregistered
 * @desc Logs MarketUnregistered event emitted by PositionNFT when a market is removed
 * @param event Decoded MarketUnregistered event
 * @param blockNum Block number where unregistration occurred
 */
func (h *ArchControllerHandler) handleMarketUnregistered(ctx context.Context, event *types.MarketUnregisteredEvent, blockNum uint64) error {
	log.Printf("MarketUnregistered (PositionNFT): Market=%s Block=%d", event.Market.Hex(), blockNum)
	return nil
}

/*@
 * handleOwnerUpdated
 * @desc Logs ArchController ownership transfer
 * @param event Decoded ArchOwnerUpdated event
 * @param blockNum Block number where ownership changed
 */
func (h *ArchControllerHandler) handleOwnerUpdated(ctx context.Context, event *types.ArchOwnerUpdatedEvent, blockNum uint64) error {
	log.Printf("ArchController OwnerUpdated: OldOwner=%s NewOwner=%s Block=%d",
		event.OldOwner.Hex(), event.NewOwner.Hex(), blockNum)
	return nil
}

/*@
 * handleTimelockUpdated
 * @desc Logs ArchController timelock duration change
 * @param event Decoded ArchTimelockUpdated event
 * @param blockNum Block number where timelock was changed
 */
func (h *ArchControllerHandler) handleTimelockUpdated(ctx context.Context, event *types.ArchTimelockUpdatedEvent, blockNum uint64) error {
	log.Printf("TimelockUpdated: OldDelay=%d seconds NewDelay=%d seconds Block=%d",
		event.OldDelay.Int64(), event.NewDelay.Int64(), blockNum)
	return nil
}
