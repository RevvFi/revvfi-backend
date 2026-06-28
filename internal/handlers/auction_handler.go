// internal/indexer/handlers/auction_handler.go
package handlers

import (
	"context"

	"time"

	"github.com/Revvfi/revvfi-backend/internal/indexer/types"
	"github.com/Revvfi/revvfi-backend/internal/models"
	"github.com/Revvfi/revvfi-backend/internal/repository/postgres"
)

/*@
 * AuctionHandler
 * @desc Handles liquidation auction events from RevvFiLiquidator
 * @events
 *   - AuctionCreated: New liquidation auction started
 *   - BidPlaced: Bid placed on auction
 *   - AuctionSettled: Auction completed successfully
 *   - AuctionCancelled: Auction cancelled (no bids)
 *
 * @auction_lifecycle
 *   AuctionCreated → active
 *   BidPlaced → updates highest_bid and highest_bidder
 *   AuctionSettled → settled (winner receives collateral)
 *   AuctionCancelled → cancelled (retry or permanent failure)
 */
type AuctionHandler struct {
	eventRepo *postgres.EventRepository
}

/*@
 * NewAuctionHandler
 * @desc Creates a new auction handler instance
 * @param eventRepo Repository for database operations
 * @returns Configured AuctionHandler
 */
func NewAuctionHandler(eventRepo *postgres.EventRepository) *AuctionHandler {
	return &AuctionHandler{eventRepo: eventRepo}
}

/*@
 * Handle
 * @desc Routes auction events to appropriate handler methods
 * @param ctx Context for cancellation and deadlines
 * @param event The decoded event from blockchain
 * @param blockNum Block number where event occurred
 * @returns error if handling fails
 */
func (h *AuctionHandler) Handle(ctx context.Context, event interface{}, blockNum uint64) error {
	switch e := event.(type) {
	case *types.AuctionCreatedEvent:
		return h.handleAuctionCreated(ctx, e)
	case *types.BidPlacedEvent:
		return h.handleBidPlaced(ctx, e)
	case *types.AuctionSettledEvent:
		return h.handleAuctionSettled(ctx, e)
	case *types.AuctionCancelledEvent:
		return h.handleAuctionCancelled(ctx, e)
	}
	return nil
}

/*@
 * handleAuctionCreated
 * @desc Creates new auction record
 * @param event Decoded AuctionCreated event
 * @param blockNum Block number where auction was created
 *
 * @state
 *   - Status: "active"
 *   - CurrentPrice = DebtAmount (starting price)
 *   - HighestBid = 0
 *   - HighestBidder = nil
 */
func (h *AuctionHandler) handleAuctionCreated(ctx context.Context, event *types.AuctionCreatedEvent) error {
	auction := &models.Auction{
		AuctionID:        event.AuctionID.Int64(),
		MarketAddress:    event.Market.Hex(),
		BorrowerAddress:  event.Borrower.Hex(),
		CollateralAmount: event.CollateralAmount,
		DebtAmount:       event.DebtAmount,
		CurrentPrice:     event.DebtAmount, // Initial price = debt amount
		StartTime:        time.Unix(event.StartTime.Int64(), 0),
		EndTime:          time.Unix(event.EndTime.Int64(), 0),
		Status:           "active",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	return h.eventRepo.SaveAuction(ctx, auction)
}

/*@
 * handleBidPlaced
 * @desc Updates auction with highest bid
 * @param event Decoded BidPlaced event
 * @param blockNum Block number where bid was placed
 *
 * @validation
 *   - Bid must be >= current price (enforced by contract)
 *   - Bid must be >= highest_bid + min_increment
 *   - Previous highest bidder is refunded
 */
func (h *AuctionHandler) handleBidPlaced(ctx context.Context, event *types.BidPlacedEvent) error {
	return h.eventRepo.UpdateAuctionBid(ctx, event.AuctionID.Int64(), event.BidAmount, event.Bidder.Hex())
}

/*@
 * handleAuctionSettled
 * @desc Marks auction as settled with winner info
 * @param event Decoded AuctionSettled event
 *
 * @state_transition
 *   - Status: "settled"
 *   - Winner receives collateral
 *   - Funds distributed to protocol
 *   - Recovery rate provided directly by contract
 *
 * @note The RecoveryRate field is emitted by the contract and represents
 *       the percentage of debt recovered (e.g., 8500 = 85.00%)
 *       No off-chain calculation needed - trust the contract's math
 */
func (h *AuctionHandler) handleAuctionSettled(ctx context.Context, event *types.AuctionSettledEvent) error {
	// RecoveryRate is provided directly by the contract
	// The contract calculates: (winningBid * 10000) / debtAmount
	// Returned as basis points (1% = 100 bps, 100% = 10000 bps)
	var recoveryRatePercent float64

	if event.RecoveryRate != nil {
		// Convert from basis points to percentage
		// Example: 8500 basis points = 85.00%
		recoveryRatePercent = float64(event.RecoveryRate.Int64()) / 100
	}

	// Store the auction settlement with the recovery rate
	// The winning bid and winner are also recorded for transparency
	return h.eventRepo.SettleAuction(ctx,
		event.AuctionID.Int64(),
		event.Winner.Hex(),
		event.WinningBid,
		recoveryRatePercent,
	)
}

/*@
 * handleAuctionCancelled
 * @desc Marks auction as cancelled (no bids received)
 * @param event Decoded AuctionCancelled event
 * @param blockNum Block number where auction was cancelled
 *
 * @state_transition
 *   - Status: "cancelled"
 *   - Collateral remains locked (new auction will be created)
 *   - Retry mechanism will trigger new auction
 *
 * @note This is not a failure - it's a normal part of the liquidation process
 *       when no bids are received during the auction period.
 */
func (h *AuctionHandler) handleAuctionCancelled(ctx context.Context, event *types.AuctionCancelledEvent) error {
	return h.eventRepo.UpdateAuctionStatus(ctx, event.AuctionID.Int64(), "cancelled")
}
