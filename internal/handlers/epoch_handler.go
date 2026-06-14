// internal/indexer/handlers/epoch_handler.go
package handlers

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/Revvfi/revvfi-backend/internal/indexer/types"
	"github.com/Revvfi/revvfi-backend/internal/models"
	"github.com/Revvfi/revvfi-backend/internal/repository/postgres"
)

/*@
 * EpochHandler
 * @desc Handles withdrawal epoch processing events from RevvFiLiquidityQueue
 * @event
 *   - EpochProcessed: Withdrawal epoch processed (every 7 days)
 *
 * @importance
 *   Implements the epoch-based withdrawal queue system.
 *   Withdrawals are processed in batches (epochs) to manage liquidity.
 *   Each epoch lasts 7 days and processes all withdrawal requests
 *   submitted during that period.
 *
 * @epoch_flow
 *   1. Lenders request withdrawals → requests added to current epoch
 *   2. Epoch ends after 7 days
 *   3. Market processes epoch (determines available liquidity)
 *   4. EpochProcessed event emitted with fulfillment results
 *   5. Lenders can claim their fulfilled amounts
 *
 * @fulfillment_rule
 *   If total requested > available liquidity:
 *     - Fulfilled pro-rata based on request amounts
 *     - Unfulfilled amounts roll over to next epoch
 *
 * @used_by
 *   - Updating withdrawal request statuses
 *   - Notifying lenders of available claims
 *   - Tracking epoch liquidity metrics
 */
type EpochHandler struct {
	eventRepo *postgres.EventRepository
}

/*@
 * NewEpochHandler
 * @desc Creates a new epoch handler instance
 * @param eventRepo Repository for database operations
 * @returns Configured EpochHandler
 */
func NewEpochHandler(eventRepo *postgres.EventRepository) *EpochHandler {
	return &EpochHandler{eventRepo: eventRepo}
}

/*@
 * Handle
 * @desc Processes EpochProcessed event
 * @param ctx Context for cancellation and deadlines
 * @param event The decoded event from blockchain
 * @param blockNum Block number where event occurred
 * @returns error if handling fails
 *
 * @state_changes
 *   - Updates withdrawal request fulfillment amounts
 *   - Marks epoch as completed
 *   - Updates lender claimable balances
 *   - Creates rollover requests for next epoch (if unfulfilled)
 */
func (h *EpochHandler) Handle(ctx context.Context, event interface{}, blockNum uint64) error {
	e, ok := event.(*types.EpochProcessedEvent)
	if !ok {
		return nil
	}

	log.Printf("EpochProcessed: Epoch=%d, TotalRequested=%s, TotalFulfilled=%s",
		e.EpochNumber.Int64(), e.TotalRequested.String(), e.TotalFulfilled.String())

	// Create epoch record in database
	epoch := &models.WithdrawalEpoch{
		EpochNumber:    int32(e.EpochNumber.Int64()),
		TotalRequested: e.TotalRequested,
		TotalFulfilled: e.TotalFulfilled,
		Status:         "completed",
		ProcessedAt:    sql.NullTime{
        Time:  time.Now(),
        Valid: true,
    },
		CreatedAt:      time.Now(),
	}

	// Save epoch and update all withdrawal requests in this epoch
	return h.eventRepo.CompleteEpoch(ctx, epoch, blockNum)
}