// internal/indexer/handlers/reputation_handler.go
package handlers

import (
	"context"

	"github.com/Revvfi/revvfi-backend/internal/indexer/types"
	"github.com/Revvfi/revvfi-backend/internal/repository/postgres"
)

/*@
 * ReputationHandler
 * @desc Handles reputation and borrower registry events from ReputationRegistry
 * @events
 *   - ReputationScoreUpdated: Borrower reputation score changed
 *   - BorrowerRegistered: New borrower registered on-chain
 *   - BorrowerAdded: Borrower whitelisted in ArchController (cross-emitted)
 *   - BorrowerRemoved: Borrower removed from ArchController whitelist (cross-emitted)
 *   - SuccessfulRepaymentRecorded: Borrower repaid a loan successfully
 *   - DefaultRecorded: Borrower defaulted on a loan
 *   - BorrowActivityRecorded: Borrower drew down funds from a market
 */
type ReputationHandler struct {
	eventRepo *postgres.EventRepository
}

/*@
 * NewReputationHandler
 * @desc Creates a new reputation handler instance
 * @param eventRepo Repository for database operations
 * @returns Configured ReputationHandler
 */
func NewReputationHandler(eventRepo *postgres.EventRepository) *ReputationHandler {
	return &ReputationHandler{eventRepo: eventRepo}
}

/*@
 * Handle
 * @desc Routes reputation events to appropriate handler methods
 * @param ctx Context for cancellation and deadlines
 * @param event The decoded event from blockchain
 * @param blockNum Block number where event occurred
 * @returns error if handling fails
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
	case *types.SuccessfulRepaymentRecordedEvent:
		return h.handleSuccessfulRepaymentRecorded(ctx, e, blockNum)
	case *types.DefaultRecordedEvent:
		return h.handleDefaultRecorded(ctx, e, blockNum)
	case *types.BorrowActivityRecordedEvent:
		return h.handleBorrowActivityRecorded(ctx, e, blockNum)
	}
	return nil
}

/*@
 * handleReputationScoreUpdated
 * @desc Updates borrower reputation score and derived risk_label
 * @param event Decoded ReputationScoreUpdated event
 * @param blockNum Block number where reputation was updated
 * @note risk_label is recalculated inside UpdateBorrowerReputation based on the new score.
 */
func (h *ReputationHandler) handleReputationScoreUpdated(ctx context.Context, event *types.ReputationScoreUpdatedEvent, blockNum uint64) error {
	return h.eventRepo.UpdateBorrowerReputation(ctx, event.Borrower.Hex(), int32(event.NewScore.Int64()))
}

/*@
 * handleBorrowerRegistered
 * @desc Creates new borrower record with default reputation score of 500 (B grade)
 * @param event Decoded BorrowerRegistered event
 * @param blockNum Block number where borrower was registered
 */
func (h *ReputationHandler) handleBorrowerRegistered(ctx context.Context, event *types.BorrowerRegisteredEvent, blockNum uint64) error {
	return h.eventRepo.RegisterBorrowerFromEvent(ctx, event.Borrower.Hex())
}

/*@
 * handleBorrowerAdded
 * @desc Updates borrower whitelist status to true (ArchController whitelisted)
 * @param event Decoded ArchBorrowerAdded event
 * @param blockNum Block number where borrower was added
 * @note This case is reached when BorrowerAdded events are dispatched to ReputationHandler.
 *       The primary handler is ArchControllerHandler; this exists as a defensive fallback.
 */
func (h *ReputationHandler) handleBorrowerAdded(ctx context.Context, event *types.ArchBorrowerAddedEvent, blockNum uint64) error {
	return h.eventRepo.UpdateBorrowerWhitelistStatus(ctx, event.Borrower.Hex(), true)
}

/*@
 * handleBorrowerRemoved
 * @desc Updates borrower whitelist status to false (ArchController removed)
 * @param event Decoded ArchBorrowerRemoved event
 * @param blockNum Block number where borrower was removed
 */
func (h *ReputationHandler) handleBorrowerRemoved(ctx context.Context, event *types.ArchBorrowerRemovedEvent, blockNum uint64) error {
	return h.eventRepo.UpdateBorrowerWhitelistStatus(ctx, event.Borrower.Hex(), false)
}

/*@
 * handleSuccessfulRepaymentRecorded
 * @desc Increments successful_loans and total_repaid when a borrower repays
 * @param event Decoded SuccessfulRepaymentRecorded event
 * @param blockNum Block number where repayment was recorded
 * @note active_loans is decremented because the loan lifecycle has ended.
 */
func (h *ReputationHandler) handleSuccessfulRepaymentRecorded(ctx context.Context, event *types.SuccessfulRepaymentRecordedEvent, blockNum uint64) error {
	return h.eventRepo.RecordSuccessfulRepayment(ctx, event.Borrower.Hex(), event.Amount)
}

/*@
 * handleDefaultRecorded
 * @desc Increments defaulted_loans and reduces reputation score by 50 points
 * @param event Decoded DefaultRecorded event
 * @param blockNum Block number where default was recorded
 * @note Reputation decreases by 50 points per default (minimum 0). Risk label is
 *       recalculated inside IncrementBorrowerDefaults to reflect the new score.
 */
func (h *ReputationHandler) handleDefaultRecorded(ctx context.Context, event *types.DefaultRecordedEvent, blockNum uint64) error {
	return h.eventRepo.IncrementBorrowerDefaults(ctx, event.Borrower.Hex())
}

/*@
 * handleBorrowActivityRecorded
 * @desc Increments total_borrowed and active/total loan counts when borrow activity is recorded
 * @param event Decoded BorrowActivityRecorded event
 * @param blockNum Block number where borrow activity was recorded
 * @note This tracks the ReputationRegistry's borrow activity recording rather than
 *       the market's Borrow event; both may fire for the same drawdown.
 */
func (h *ReputationHandler) handleBorrowActivityRecorded(ctx context.Context, event *types.BorrowActivityRecordedEvent, blockNum uint64) error {
	return h.eventRepo.IncrementBorrowerTotalBorrowed(ctx, event.Borrower.Hex(), event.Amount)
}
