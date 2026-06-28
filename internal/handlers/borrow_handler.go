// internal/indexer/handlers/borrow_handler.go
package handlers

import (
	"context"
	"time"

	"github.com/Revvfi/revvfi-backend/internal/indexer/types"
	"github.com/Revvfi/revvfi-backend/internal/logger"
	"github.com/Revvfi/revvfi-backend/internal/repository/postgres"
)

/*@
 * BorrowHandler
 * @desc Handles Borrow event from RevvFiMarket
 * @event Borrow: Borrower accepts offers and receives funds
 */
type BorrowHandler struct {
	eventRepo *postgres.EventRepository
}

/*@
 * NewBorrowHandler
 * @desc Creates a new borrow handler instance
 */
func NewBorrowHandler(eventRepo *postgres.EventRepository) *BorrowHandler {
	return &BorrowHandler{eventRepo: eventRepo}
}

/*@
 * Handle
 * @desc Processes Borrow event
 * @param event Decoded Borrow event
 * @param blockNum Block number where borrow occurred
 */
func (h *BorrowHandler) Handle(ctx context.Context, event interface{}, blockNum uint64) error {
	e, ok := event.(*types.BorrowEvent)
	if !ok {
		return nil
	}
	return h.handleBorrow(ctx, e, blockNum)
}

/*@
 * handleBorrow
 * @desc Updates market debt after borrow
 * @param event Decoded Borrow event
 * @param blockNum Block number where borrow occurred
 */
func (h *BorrowHandler) handleBorrow(ctx context.Context, event *types.BorrowEvent, blockNum uint64) error {
	start := time.Now()

	logger.InfoContext(ctx, "Processing Borrow event",
		logger.WithEventName("Borrow"),
		logger.WithBlockNumber(blockNum),
		logger.WithBorrower(event.Borrower.Hex()),
		logger.WithAmount(event.Amount.String()),
		"weighted_apr", event.WeightedAPR,
	)

	// Update market total debt
	// TODO: Add actual database update logic here

	logger.InfoContext(ctx, "Borrow event processed successfully",
		logger.WithBorrower(event.Borrower.Hex()),
		logger.WithDuration(time.Since(start)),
	)

	return nil
}
