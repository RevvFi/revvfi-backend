// internal/indexer/handlers/interest_handler.go
package handlers

import (
	"context"
	"log"

	"github.com/Revvfi/revvfi-backend/internal/indexer/types"
	"github.com/Revvfi/revvfi-backend/internal/repository/postgres"
)

/*@
 * InterestHandler
 * @desc Handles interest accrual events from RevvFiMarket
 * @event
 *   - InterestAccrued: Interest accrued on market debt (Compound-style)
 *
 * @importance
 *   Tracks the cumulative borrow index which is used to calculate
 *   accrued interest for all positions in the market.
 *   This event is emitted on every block where interest accrues.
 *
 * @formula
 *   borrowIndex_new = borrowIndex_old * (1 + interestRate * timeDelta)
 *   interestAccrued = totalPrincipal * (borrowIndex_new - borrowIndex_old) / SCALE
 *
 * @used_by
 *   - Calculating lender interest earnings
 *   - Updating position values
 *   - Health factor calculations
 */
type InterestHandler struct {
	eventRepo *postgres.EventRepository
}

/*@
 * NewInterestHandler
 * @desc Creates a new interest handler instance
 * @param eventRepo Repository for database operations
 * @returns Configured InterestHandler
 */
func NewInterestHandler(eventRepo *postgres.EventRepository) *InterestHandler {
	return &InterestHandler{eventRepo: eventRepo}
}

/*@
 * Handle
 * @desc Processes InterestAccrued event
 * @param ctx Context for cancellation and deadlines
 * @param event The decoded event from blockchain
 * @param blockNum Block number where event occurred
 * @returns error if handling fails
 *
 * @state_changes
 *   - Updates market.borrow_index
 *   - Updates market.last_interest_accrual
 *   - May trigger position value recalculations
 */
func (h *InterestHandler) Handle(ctx context.Context, event interface{}, blockNum uint64) error {
	e, ok := event.(*types.InterestAccruedEvent)
	if !ok {
		return nil
	}

	log.Printf("InterestAccrued: Market=%s, BorrowIndex=%s, BlockNumber=%d",
		e.Market.Hex(), e.BorrowIndex.String(), blockNum)

	// Update market's borrow index in database
	// This index is used to calculate accrued interest for all positions
	return h.eventRepo.UpdateMarketBorrowIndex(ctx, e.Market.Hex(), e.BorrowIndex)
}
