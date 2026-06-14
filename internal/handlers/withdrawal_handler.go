// internal/indexer/handlers/withdrawal_handler.go
package handlers

import (
	"context"
	"math/big"
	"time"

	"github.com/Revvfi/revvfi-backend/internal/indexer/types"
	"github.com/Revvfi/revvfi-backend/internal/models"
	"github.com/Revvfi/revvfi-backend/internal/repository/postgres"
)

/*@
 * WithdrawalHandler
 * @desc Handles withdrawal queue events from RevvFiLiquidityQueue
 * @events
 *   - WithdrawalRequested: Lender requests withdrawal
 *   - WithdrawalClaimed: Lender claims fulfilled withdrawal
 *   - WithdrawalCancelled: Lender cancels request
 *   - EpochProcessed: Withdrawal epoch processed
 */
type WithdrawalHandler struct {
    eventRepo *postgres.EventRepository
}

/*@
 * NewWithdrawalHandler
 * @desc Creates a new withdrawal handler instance
 */
func NewWithdrawalHandler(eventRepo *postgres.EventRepository) *WithdrawalHandler {
    return &WithdrawalHandler{eventRepo: eventRepo}
}

/*@
 * Handle
 * @desc Routes withdrawal events to appropriate handler methods
 */
func (h *WithdrawalHandler) Handle(ctx context.Context, event interface{}, blockNum uint64) error {
    switch e := event.(type) {
    case *types.WithdrawalRequestedEvent:
        return h.handleWithdrawalRequested(ctx, e, blockNum)
    case *types.WithdrawalClaimedEvent:
        return h.handleWithdrawalClaimed(ctx, e, blockNum)
    case *types.WithdrawalCancelledEvent:
        return h.handleWithdrawalCancelled(ctx, e, blockNum)
    case *types.EpochProcessedEvent:
        return h.handleEpochProcessed(ctx, e, blockNum)
    }
    return nil
}

/*@
 * handleWithdrawalRequested
 * @desc Creates new withdrawal request record
 * @param event Decoded WithdrawalRequested event
 * @param blockNum Block number where request was made
 */
func (h *WithdrawalHandler) handleWithdrawalRequested(ctx context.Context, event *types.WithdrawalRequestedEvent, blockNum uint64) error {
    request := &models.WithdrawalRequest{
        RequestID:       event.RequestID.Int64(),
        Lender:          event.Lender.Hex(),
        PositionID:      event.PositionID.Int64(),
        MarketAddress:   event.Market.Hex(),
        RequestedAmount: event.Amount,
        FulfilledAmount: big.NewInt(0),
        EpochNumber:     int32(event.EpochNumber.Int64()),
        Status:          "pending",
        CreatedAt:       time.Now(),
        UpdatedAt:       time.Now(),
    }
    
    return h.eventRepo.SaveWithdrawalRequest(ctx, request)
}

/*@
 * handleWithdrawalClaimed
 * @desc Updates withdrawal request as claimed
 * @param event Decoded WithdrawalClaimed event
 * @param blockNum Block number where claim occurred
 */
func (h *WithdrawalHandler) handleWithdrawalClaimed(ctx context.Context, event *types.WithdrawalClaimedEvent, blockNum uint64) error {
    return h.eventRepo.MarkWithdrawalClaimed(ctx, event.RequestID.Int64())
}

/*@
 * handleWithdrawalCancelled
 * @desc Updates withdrawal request as cancelled
 * @param event Decoded WithdrawalCancelled event
 * @param blockNum Block number where cancellation occurred
 */
func (h *WithdrawalHandler) handleWithdrawalCancelled(ctx context.Context, event *types.WithdrawalCancelledEvent, blockNum uint64) error {
    return h.eventRepo.UpdateWithdrawalStatus(ctx, event.RequestID.Int64(), "cancelled")
}

/*@
 * handleEpochProcessed
 * @desc Updates withdrawal epoch with fulfillment amounts
 * @param event Decoded EpochProcessed event
 * @param blockNum Block number where epoch was processed
 */
func (h *WithdrawalHandler) handleEpochProcessed(ctx context.Context, event *types.EpochProcessedEvent, blockNum uint64) error {
    return h.eventRepo.UpdateEpochFulfillment(ctx, int32(event.EpochNumber.Int64()), event.TotalRequested, event.TotalFulfilled)
}