// internal/indexer/handlers/liquidation_handler.go
package handlers

import (
	"context"
	"log"

	"github.com/Revvfi/revvfi-backend/internal/indexer/types"
	"github.com/Revvfi/revvfi-backend/internal/repository/postgres"
)

/*@
 * LiquidationHandler
 * @desc Handles liquidation execution events from RevvFiMarket
 * @event
 *   - LiquidationExecuted: Borrower position liquidated (collateral seized)
 *
 * @importance
 *   Critical for protocol solvency - triggered when collateral ratio
 *   falls below liquidation threshold. Liquidator purchases collateral
 *   at a discount to cover bad debt.
 *
 * @trigger_conditions
 *   - Collateral ratio <= liquidationThreshold (default 95%)
 *   - Anyone can call liquidate() (third-party liquidators)
 *   - Dutch auction ensures fair market price
 *
 * @effects
 *   - Liquidator receives collateral at discount
 *   - Protocol repays debt from liquidation proceeds
 *   - Remaining bad debt (if any) is absorbed by junior tranche
 *   - Borrower's position is closed
 *
 * @note
 *   This is different from AuctionCreated event (which starts the auction).
 *   LiquidationExecuted is emitted AFTER the auction completes successfully.
 */
type LiquidationHandler struct {
	eventRepo *postgres.EventRepository
}

/*@
 * NewLiquidationHandler
 * @desc Creates a new liquidation handler instance
 * @param eventRepo Repository for database operations
 * @returns Configured LiquidationHandler
 */
func NewLiquidationHandler(eventRepo *postgres.EventRepository) *LiquidationHandler {
	return &LiquidationHandler{eventRepo: eventRepo}
}

/*@
 * Handle
 * @desc Processes LiquidationExecuted, LiquidationStartedMarket, and
 *       LiquidationEndedMarket events
 * @param ctx Context for cancellation and deadlines
 * @param event The decoded event from blockchain
 * @param blockNum Block number where event occurred
 * @returns error if handling fails
 *
 * @state_changes
 *   - LiquidationExecuted: marks borrower's positions as liquidated,
 *     updates reputation, records liquidation history, updates market debt
 *   - LiquidationStartedMarket/LiquidationEndedMarket: flips the market's
 *     is_liquidating flag - this was previously registered to this handler
 *     but silently dropped (the type switch below didn't have cases for
 *     them), so markets.is_liquidating never reflected real liquidation
 *     state. That let the frontend keep offering "Trigger Liquidation" on
 *     a market that was already mid-auction, guaranteeing a revert for
 *     the second caller.
 */
func (h *LiquidationHandler) Handle(ctx context.Context, event interface{}, blockNum uint64) error {
	switch e := event.(type) {
	case *types.LiquidationExecutedEvent:
		log.Printf(" LiquidationExecuted: Borrower=%s, DebtAmount=%s, CollateralSeized=%s",
			e.Borrower.Hex(), e.DebtAmount.String(), e.CollateralAmount.String())
		return h.eventRepo.RecordLiquidation(ctx, e.Borrower.Hex(), e.DebtAmount, e.CollateralAmount, blockNum)
	case *types.LiquidationStartedMarketEvent:
		return h.eventRepo.UpdateMarketLiquidating(ctx, e.Market.Hex(), true)
	case *types.LiquidationEndedMarketEvent:
		return h.eventRepo.UpdateMarketLiquidating(ctx, e.Market.Hex(), false)
	default:
		return nil
	}
}
