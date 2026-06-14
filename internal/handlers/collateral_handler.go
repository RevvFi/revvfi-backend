// internal/handlers/collateral_handler.go
package handlers

import (
	"context"
	"log"

	"github.com/Revvfi/revvfi-backend/internal/indexer/types"
	"github.com/Revvfi/revvfi-backend/internal/repository/postgres"
)

/*@
 * CollateralHandler
 * @desc Handles collateral-related events from RevvFiMarket
 * @events
 *   - CollateralDeposited: Borrower deposited collateral
 *   - CollateralWithdrawn: Borrower withdrew collateral
 */
type CollateralHandler struct {
	eventRepo *postgres.EventRepository
}

/*@
 * NewCollateralHandler
 * @desc Creates a new collateral handler instance
 * @param eventRepo Repository for database operations
 * @returns Configured CollateralHandler
 */
func NewCollateralHandler(eventRepo *postgres.EventRepository) *CollateralHandler {
	return &CollateralHandler{eventRepo: eventRepo}
}

/*@
 * Handle
 * @desc Routes collateral events to appropriate handler methods
 * @param ctx Context for cancellation and deadlines
 * @param event The decoded event from blockchain
 * @param blockNum Block number where event occurred
 * @returns error if handling fails
 */
func (h *CollateralHandler) Handle(ctx context.Context, event interface{}, blockNum uint64) error {
	switch e := event.(type) {
	case *types.CollateralDepositedEvent:
		return h.handleCollateralDeposited(ctx, e, blockNum)
	case *types.CollateralWithdrawnEvent:
		return h.handleCollateralWithdrawn(ctx, e, blockNum)
	}
	return nil
}

func (h *CollateralHandler) handleCollateralDeposited(ctx context.Context, event *types.CollateralDepositedEvent, blockNum uint64) error {
	log.Printf(" CollateralDeposited: Borrower=%s, Amount=%s", event.Borrower.Hex(), event.Amount.String())
	// TODO: Update collateral balance in database
	return nil
}

func (h *CollateralHandler) handleCollateralWithdrawn(ctx context.Context, event *types.CollateralWithdrawnEvent, blockNum uint64) error {
	log.Printf("CollateralWithdrawn: Borrower=%s, Amount=%s", event.Borrower.Hex(), event.Amount.String())
	// TODO: Update collateral balance in database
	return nil
}