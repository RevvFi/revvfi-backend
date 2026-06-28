// internal/indexer/handlers/offer_handler.go
package handlers

import (
	"context"
	"math/big"
	"time"

	"github.com/Revvfi/revvfi-backend/internal/indexer/types"
	"github.com/Revvfi/revvfi-backend/internal/logger"
	"github.com/Revvfi/revvfi-backend/internal/models"
	"github.com/Revvfi/revvfi-backend/internal/repository/postgres"
)

/*@
 * OfferHandler
 * @desc Handles all offer-related blockchain events from RevvFiOfferBook
 * @events
 *   - OfferSubmitted: New lender offer created
 *   - OfferCancelled: Existing offer cancelled by lender
 *   - OfferFilled: Offer partially or fully filled during borrow
 *   - OfferModified: Offer amount, APR, or expiry changed by lender
 *
 * @lifecycle
 *   OfferSubmitted → active
 *   OfferModified → updates amount/APR/expiry while active
 *   OfferFilled → updates remaining_amount (may become partially_filled or filled)
 *   OfferCancelled → cancelled (funds returned to lender)
 */
type OfferHandler struct {
	eventRepo *postgres.EventRepository
}

/*@
 * NewOfferHandler
 * @desc Creates a new offer handler instance
 * @param eventRepo Repository for database operations
 * @returns Configured OfferHandler
 */
func NewOfferHandler(eventRepo *postgres.EventRepository) *OfferHandler {
	return &OfferHandler{eventRepo: eventRepo}
}

/*@
 * Handle
 * @desc Routes offer events to appropriate handler methods
 * @param ctx Context for cancellation and deadlines
 * @param event The decoded event from blockchain
 * @param blockNum Block number where event occurred
 * @returns error if handling fails
 */
func (h *OfferHandler) Handle(ctx context.Context, event interface{}, blockNum uint64) error {
	switch e := event.(type) {
	case *types.OfferSubmittedEvent:
		return h.handleOfferSubmitted(ctx, e, blockNum)
	case *types.OfferCancelledEvent:
		return h.handleOfferCancelled(ctx, e, blockNum)
	case *types.OfferFilledEvent:
		return h.handleOfferFilled(ctx, e, blockNum)
	case *types.OfferModifiedEvent:
		return h.handleOfferModified(ctx, e, blockNum)
	}
	return nil
}

/*@
 * handleOfferSubmitted
 * @desc Processes OfferSubmitted event - creates new offer record
 * @param event Decoded OfferSubmitted event
 * @param blockNum Block number where offer was submitted
 *
 * @state_transition
 *   - Status: "active"
 *   - RemainingAmount = Amount (fully available)
 *   - FilledAmount = 0
 */
func (h *OfferHandler) handleOfferSubmitted(ctx context.Context, event *types.OfferSubmittedEvent, blockNum uint64) error {
	start := time.Now()

	logger.InfoContext(ctx, "Processing OfferSubmitted event",
		logger.WithEventName("OfferSubmitted"),
		logger.WithBlockNumber(blockNum),
		logger.WithOfferID(event.OfferID.Int64()),
		logger.WithLender(event.Lender.Hex()),
		logger.WithMarketAddress(event.Market.Hex()),
		logger.WithAmount(event.Amount.String()),
		logger.WithAPR(int(event.APR.Int64())),
		logger.WithSeniority(int(event.Seniority)),
	)

	// Convert expiry from seconds to time.Time
	// If expiry is 0, set to far future (no expiry)
	var expiryTime time.Time
	if event.Expiry.Int64() == 0 {
		expiryTime = time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)
	} else {
		expiryTime = time.Unix(event.Expiry.Int64(), 0)
	}

	offer := &models.Offer{
		OfferID:         event.OfferID.Int64(),
		Lender:          event.Lender.Hex(),
		MarketAddress:   event.Market.Hex(),
		Amount:          new(big.Int).Set(event.Amount),
		RemainingAmount: new(big.Int).Set(event.Amount),
		APR:             int32(event.APR.Int64()),
		Seniority:       int16(event.Seniority),
		Status:          "active",
		Expiry:          expiryTime,
		BlockNumber:     int64(blockNum),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := h.eventRepo.SaveOffer(ctx, offer); err != nil {
		logger.ErrorContext(ctx, "Failed to save offer to database",
			logger.WithError(err),
			logger.WithOfferID(event.OfferID.Int64()),
			logger.WithMarketAddress(event.Market.Hex()),
		)
		return err
	}

	logger.InfoContext(ctx, "Offer saved to database successfully",
		logger.WithOfferID(event.OfferID.Int64()),
		logger.WithDuration(time.Since(start)),
	)

	return nil
}

/*@
 * handleOfferCancelled
 * @desc Processes OfferCancelled event - marks offer as cancelled
 * @param event Decoded OfferCancelled event
 * @param blockNum Block number where cancellation occurred
 *
 * @state_transition
 *   - Status: "cancelled"
 *   - Funds returned to lender (on-chain)
 *   - Offer no longer available for matching
 */
func (h *OfferHandler) handleOfferCancelled(ctx context.Context, event *types.OfferCancelledEvent, blockNum uint64) error {
	logger.InfoContext(ctx, "Processing OfferCancelled event",
		logger.WithEventName("OfferCancelled"),
		logger.WithBlockNumber(blockNum),
		logger.WithOfferID(event.OfferID.Int64()),
		logger.WithLender(event.Lender.Hex()),
	)

	if err := h.eventRepo.UpdateOfferStatus(ctx, event.OfferID.Int64(), "cancelled"); err != nil {
		logger.ErrorContext(ctx, "Failed to cancel offer",
			logger.WithError(err),
			logger.WithOfferID(event.OfferID.Int64()),
		)
		return err
	}

	logger.InfoContext(ctx, "Offer cancelled successfully",
		logger.WithOfferID(event.OfferID.Int64()),
	)

	return nil
}

/*@
 * handleOfferFilled
 * @desc Processes OfferFilled event - updates remaining amount when offer is used
 * @param event Decoded OfferFilled event
 * @param blockNum Block number where fill occurred
 *
 * @state_transition
 *   - If remaining_amount becomes 0: Status = "filled"
 *   - If remaining_amount > 0: Status = "partially_filled"
 *   - RemainingAmount decreases by fill amount
 *
 * @note This event is emitted during borrow() execution
 */
func (h *OfferHandler) handleOfferFilled(ctx context.Context, event *types.OfferFilledEvent, blockNum uint64) error {
	return h.eventRepo.DecrementOfferRemaining(ctx, event.OfferID.Int64(), event.Amount)
}

/*@
 * handleOfferModified
 * @desc Processes OfferModified event - updates offer parameters
 * @param event Decoded OfferModified event
 * @param blockNum Block number where modification occurred
 *
 * @state_transition
 *   - Amount may increase (lender adds more liquidity)
 *   - Amount may decrease (lender reduces offer)
 *   - APR may change
 *   - Expiry may be extended or reduced
 *   - Status remains "active" unless amount becomes 0
 *
 * @security Only the offer lender can modify
 * @note Cannot modify after offer is filled or cancelled
 */
func (h *OfferHandler) handleOfferModified(ctx context.Context, event *types.OfferModifiedEvent, blockNum uint64) error {
	// Convert expiry from seconds to time.Time
	var expiryTime time.Time
	if event.NewExpiry.Int64() == 0 {
		expiryTime = time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)
	} else {
		expiryTime = time.Unix(event.NewExpiry.Int64(), 0)
	}

	// Determine new status based on amount
	status := "active"
	if event.NewAmount.Int64() == 0 {
		status = "cancelled"
	}

	// Update offer with new parameters
	return h.eventRepo.UpdateOffer(
		ctx,
		event.OfferID.Int64(),
		event.NewAmount,
		event.NewAPR,
		expiryTime,
		status,
	)
}
