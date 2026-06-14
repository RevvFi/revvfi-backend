// internal/indexer/handlers/repay_handler.go
package handlers

import (
    "context"
    "time"
     "math/big"
    "github.com/Revvfi/revvfi-backend/internal/models"
    "github.com/Revvfi/revvfi-backend/internal/repository/postgres"
    "github.com/Revvfi/revvfi-backend/internal/indexer/types"
)

/*@
 * RepayHandler
 * @desc Handles Repay event from RevvFiMarket
 * @event Repay: Borrower repays debt (partial or full)
 */
type RepayHandler struct {
    eventRepo *postgres.EventRepository
}

/*@
 * NewRepayHandler
 * @desc Creates a new repay handler instance
 */
func NewRepayHandler(eventRepo *postgres.EventRepository) *RepayHandler {
    return &RepayHandler{eventRepo: eventRepo}
}

/*@
 * Handle
 * @desc Processes Repay event
 * @param event Decoded Repay event
 * @param blockNum Block number where repayment occurred
 */
func (h *RepayHandler) Handle(ctx context.Context, event interface{}, blockNum uint64) error {
    e, ok := event.(*types.RepayEvent)
    if !ok {
        return nil
    }
    return h.handleRepay(ctx, e, blockNum)
}

/*@
 * handleRepay
 * @desc Records repayment and updates market debt
 * @param event Decoded Repay event
 * @param blockNum Block number where repayment occurred
 */
func (h *RepayHandler) handleRepay(ctx context.Context, event *types.RepayEvent, blockNum uint64) error {
    // Determine repayment type based on principal paid
    // If all principal is repaid, it's a full repayment
    repaymentType := "partial"
    if event.PrincipalPaid.Cmp(event.Amount) == 0 && event.InterestPaid.Cmp(big.NewInt(0)) > 0 {
        // If full amount is principal and there was interest, likely full repayment
        repaymentType = "full"
    }

    repayment := &models.Repayment{
        MarketAddress:   "", // Would need to get from context
        BorrowerAddress: event.Borrower.Hex(),
        Amount:          event.Amount,
        InterestPaid:    event.InterestPaid,
        PrincipalPaid:   event.PrincipalPaid,
        RepaymentType:   repaymentType,
        BlockNumber:     int64(blockNum),
        CreatedAt:       time.Now(),
    }

    if err := h.eventRepo.SaveRepayment(ctx, repayment); err != nil {
        return err
    }

    // Note: TotalDebt is no longer in RepayEvent (not emitted by contract)
    // Market debt tracking should be calculated from Borrow and Repay amounts
    return nil
}