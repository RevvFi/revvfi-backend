// internal/handlers/collateral_handler.go
package handlers

import (
	"context"
	"log"
	"math/big"

	"github.com/Revvfi/revvfi-backend/internal/indexer/types"
	"github.com/Revvfi/revvfi-backend/internal/repository/postgres"
)

/*@
 * CollateralHandler
 * @desc Handles all collateral-related events from RevvFiCollateralEscrow
 * @events
 *   - CollateralDeposited: Borrower deposited collateral into the escrow
 *   - CollateralWithdrawn: Borrower withdrew excess collateral from the escrow
 *   - CollateralLiquidated: Collateral was seized during auction settlement
 *   - MinCollateralRatioUpdated: Minimum collateral ratio changed by admin
 *   - LiquidationThresholdUpdated: Liquidation threshold changed by admin
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
	case *types.CollateralLiquidatedEvent:
		return h.handleCollateralLiquidated(ctx, e, blockNum)
	case *types.MinCollateralRatioUpdatedEvent:
		return h.handleMinCollateralRatioUpdated(ctx, e, blockNum)
	case *types.LiquidationThresholdUpdatedEvent:
		return h.handleLiquidationThresholdUpdated(ctx, e, blockNum)
	}
	return nil
}

/*@
 * handleCollateralDeposited
 * @desc Processes CollateralDeposited event - upserts borrower's collateral balance
 * @param event Decoded CollateralDeposited event with borrower, amount, and new total
 * @param blockNum Block number where deposit occurred
 * @note Uses event.Total as the authoritative current balance; event.Amount is
 *       accumulated into total_deposited for historical analysis.
 */
func (h *CollateralHandler) handleCollateralDeposited(ctx context.Context, event *types.CollateralDepositedEvent, blockNum uint64) error {
	log.Printf("CollateralDeposited: Borrower=%s Amount=%s Total=%s Block=%d",
		event.Borrower.Hex(), event.Amount.String(), event.Total.String(), blockNum)
	return h.eventRepo.RecordCollateralDeposit(ctx, event.Borrower.Hex(), event.Amount, event.Total)
}

/*@
 * handleCollateralWithdrawn
 * @desc Processes CollateralWithdrawn event - upserts borrower's collateral balance
 * @param event Decoded CollateralWithdrawn event with borrower, amount, and new total
 * @param blockNum Block number where withdrawal occurred
 * @note Uses event.Total as the authoritative current balance; event.Amount is
 *       accumulated into total_withdrawn for historical analysis.
 */
func (h *CollateralHandler) handleCollateralWithdrawn(ctx context.Context, event *types.CollateralWithdrawnEvent, blockNum uint64) error {
	log.Printf("CollateralWithdrawn: Borrower=%s Amount=%s Total=%s Block=%d",
		event.Borrower.Hex(), event.Amount.String(), event.Total.String(), blockNum)
	return h.eventRepo.RecordCollateralWithdrawal(ctx, event.Borrower.Hex(), event.Amount, event.Total)
}

/*@
 * handleCollateralLiquidated
 * @desc Processes CollateralLiquidated event - zeroes the borrower's collateral balance
 * @param event Decoded CollateralLiquidated event
 * @param blockNum Block number where liquidation occurred
 * @note The auction handler records settlement details; this handler reflects that
 *       all collateral was seized by recording a withdrawal that sets balance to zero.
 */
func (h *CollateralHandler) handleCollateralLiquidated(ctx context.Context, event *types.CollateralLiquidatedEvent, blockNum uint64) error {
	log.Printf("CollateralLiquidated: Borrower=%s Amount=%s Proceeds=%s Block=%d",
		event.Borrower.Hex(), event.Amount.String(), event.Proceeds.String(), blockNum)
	// After liquidation the borrower holds zero collateral; Total is set to 0.
	return h.eventRepo.RecordCollateralWithdrawal(ctx, event.Borrower.Hex(), event.Amount, big.NewInt(0))
}

/*@
 * handleMinCollateralRatioUpdated
 * @desc Logs MinCollateralRatioUpdated event from CollateralEscrow
 * @param event Decoded MinCollateralRatioUpdated event with old/new ratio in bps
 * @param blockNum Block number where update occurred
 * @note Updating the markets table requires resolving the escrow contract address
 *       to the market address via the ContractsSet event mapping, which is not yet
 *       stored. Currently logs the change for observability.
 */
func (h *CollateralHandler) handleMinCollateralRatioUpdated(ctx context.Context, event *types.MinCollateralRatioUpdatedEvent, blockNum uint64) error {
	log.Printf("MinCollateralRatioUpdated: OldRatio=%s NewRatio=%s Block=%d (TODO: resolve escrow->market)",
		event.OldRatio.String(), event.NewRatio.String(), blockNum)
	return nil
}

/*@
 * handleLiquidationThresholdUpdated
 * @desc Logs LiquidationThresholdUpdated event from CollateralEscrow
 * @param event Decoded LiquidationThresholdUpdated event with old/new threshold in bps
 * @param blockNum Block number where update occurred
 * @note Updating the markets table requires resolving the escrow contract address
 *       to the market address via the ContractsSet event mapping, which is not yet
 *       stored. Currently logs the change for observability.
 */
func (h *CollateralHandler) handleLiquidationThresholdUpdated(ctx context.Context, event *types.LiquidationThresholdUpdatedEvent, blockNum uint64) error {
	log.Printf("LiquidationThresholdUpdated: OldThreshold=%s NewThreshold=%s Block=%d (TODO: resolve escrow->market)",
		event.OldThreshold.String(), event.NewThreshold.String(), blockNum)
	return nil
}
