// internal/indexer/handlers/reputation_handler.go
package handlers

import (
    "context"

    "github.com/Revvfi/revvfi-backend/internal/repository/postgres"
    "github.com/Revvfi/revvfi-backend/internal/indexer/types"
)

/*@
 * ReputationHandler
 * @desc Handles reputation and borrower registry events
 * @events
 *   - ReputationScoreUpdated: Borrower reputation score changes
 *   - BorrowerRegistered: New borrower registered
 *   - BorrowerAdded: Admin whitelists borrower (ArchController)
 *   - BorrowerRemoved: Admin removes borrower (ArchController)
 */
type ReputationHandler struct {
    eventRepo *postgres.EventRepository
}

/*@
 * NewReputationHandler
 * @desc Creates a new reputation handler instance
 */
func NewReputationHandler(eventRepo *postgres.EventRepository) *ReputationHandler {
    return &ReputationHandler{eventRepo: eventRepo}
}

/*@
 * Handle
 * @desc Routes reputation events to appropriate handler methods
 */
func (h *ReputationHandler) Handle(ctx context.Context, event interface{}, blockNum uint64) error {
    switch e := event.(type) {
    case *types.ReputationScoreUpdatedEvent:
        return h.handleReputationScoreUpdated(ctx, e, blockNum)
    case *types.BorrowerRegisteredEvent:
        return h.handleBorrowerRegistered(ctx, e, blockNum)
    case *types.ArchBorrowerAddedEvent:
        return h.handleBorrowerAdded(ctx, e, blockNum)
    case *types.ArchBorrowerRemovedEvent:
        return h.handleBorrowerRemoved(ctx, e, blockNum)
    }
    return nil
}

/*@
 * handleReputationScoreUpdated
 * @desc Updates borrower reputation score
 * @param event Decoded ReputationScoreUpdated event
 * @param blockNum Block number where reputation was updated
 */
func (h *ReputationHandler) handleReputationScoreUpdated(ctx context.Context, event *types.ReputationScoreUpdatedEvent, blockNum uint64) error {
    return h.eventRepo.UpdateBorrowerReputation(ctx, event.Borrower.Hex(), int32(event.NewScore.Int64()))
}

/*@
 * handleBorrowerRegistered
 * @desc Creates new borrower record
 * @param event Decoded BorrowerRegistered event
 * @param blockNum Block number where borrower was registered
 */
func (h *ReputationHandler) handleBorrowerRegistered(ctx context.Context, event *types.BorrowerRegisteredEvent, blockNum uint64) error {
    return h.eventRepo.RegisterBorrowerFromEvent(ctx, event.Borrower.Hex())
}

/*@
 * handleBorrowerAdded
 * @desc Updates borrower whitelist status (ArchController)
 * @param event Decoded ArchBorrowerAdded event
 * @param blockNum Block number where borrower was added
 */
func (h *ReputationHandler) handleBorrowerAdded(ctx context.Context, event *types.ArchBorrowerAddedEvent, blockNum uint64) error {
    return h.eventRepo.UpdateBorrowerWhitelistStatus(ctx, event.Borrower.Hex(), true)
}

/*@
 * handleBorrowerRemoved
 * @desc Updates borrower whitelist status (ArchController)
 * @param event Decoded ArchBorrowerRemoved event
 * @param blockNum Block number where borrower was removed
 */
func (h *ReputationHandler) handleBorrowerRemoved(ctx context.Context, event *types.ArchBorrowerRemovedEvent, blockNum uint64) error {
    return h.eventRepo.UpdateBorrowerWhitelistStatus(ctx, event.Borrower.Hex(), false)
}