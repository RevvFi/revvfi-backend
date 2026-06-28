// internal/indexer/handlers/position_handler.go
package handlers

import (
	"context"
	"log"
	"math/big"
	"time"

	"github.com/Revvfi/revvfi-backend/internal/indexer/types"
	"github.com/Revvfi/revvfi-backend/internal/models"
	"github.com/Revvfi/revvfi-backend/internal/repository/postgres"
)

/*@
 * PositionHandler
 * @desc Handles all position NFT-related blockchain events
 * @events
 *   - PositionMinted: New lender position created
 *   - PositionSettled: Position fully repaid and settled
 *   - PositionRedeemed: Lender redeems position NFT for principal + interest
 *   - Transfer: Position NFT transferred between wallets
 */
type PositionHandler struct {
	eventRepo *postgres.EventRepository
}

/*@
 * NewPositionHandler
 * @desc Creates a new position handler instance
 */
func NewPositionHandler(eventRepo *postgres.EventRepository) *PositionHandler {
	return &PositionHandler{eventRepo: eventRepo}
}

/*@
 * Handle
 * @desc Routes position events to appropriate handler methods
 */
func (h *PositionHandler) Handle(ctx context.Context, event interface{}, blockNum uint64) error {
	switch e := event.(type) {
	case *types.PositionMintedEvent:
		return h.handlePositionMinted(ctx, e, blockNum)
	case *types.PositionSettledEvent:
		return h.handlePositionSettled(ctx, e, blockNum)
	case *types.PositionRedeemedEvent:
		return h.handlePositionRedeemed(ctx, e, blockNum)
	case *types.PositionTransferEvent:
		return h.handlePositionTransfer(ctx, e, blockNum)
	}
	return nil
}

/*@
 * handlePositionMinted
 * @desc Processes PositionMinted event - creates new position record
 * @param event Decoded PositionMinted event
 * @param blockNum Block number where position was minted
 */
func (h *PositionHandler) handlePositionMinted(ctx context.Context, event *types.PositionMintedEvent, blockNum uint64) error {
	position := &models.Position{
		TokenID:          event.TokenID.Int64(),
		Lender:           event.Lender.Hex(),
		MarketAddress:    event.Market.Hex(),
		Principal:        event.Principal,
		CurrentPrincipal: event.Principal,
		AccruedInterest:  big.NewInt(0),
		ClaimableAmount:  big.NewInt(0),
		APR:              int32(event.APR.Int64()),
		Seniority:        int16(event.Seniority),
		Status:           "active",
		IsActive:         true,
		IsSettled:        false,
		MintedAt:         time.Now(),
		BlockNumber:      int64(blockNum),
	}
	return h.eventRepo.SavePosition(ctx, position)
}

/*@
 * handlePositionSettled
 * @desc Processes PositionSettled event - marks position as settled
 * @param event Decoded PositionSettled event
 * @param blockNum Block number where position was settled
 */
func (h *PositionHandler) handlePositionSettled(ctx context.Context, event *types.PositionSettledEvent, blockNum uint64) error {
	return h.eventRepo.SettlePosition(ctx, event.TokenID.Int64(), event.ClaimableAmount)
}

/*@
 * handlePositionRedeemed
 * @desc Processes PositionRedeemed event - marks position as redeemed and deactivates it
 * @param event Decoded PositionRedeemed event with token ID, principal, and interest
 * @param blockNum Block number where redemption occurred
 * @note Claimable amount is set to principal + interest so APIs can display the total payout.
 *       Once redeemed, the NFT is burned and the position record is permanently inactive.
 */
func (h *PositionHandler) handlePositionRedeemed(ctx context.Context, event *types.PositionRedeemedEvent, blockNum uint64) error {
	log.Printf("PositionRedeemed: TokenID=%d Principal=%s Interest=%s Block=%d",
		event.TokenID.Int64(), event.Principal.String(), event.Interest.String(), blockNum)
	return h.eventRepo.RedeemPosition(ctx, event.TokenID.Int64(), event.Principal, event.Interest)
}

/*@
 * handlePositionTransfer
 * @desc Processes Transfer event - updates position owner
 * @param event Decoded Transfer event
 * @param blockNum Block number where transfer occurred
 */
func (h *PositionHandler) handlePositionTransfer(ctx context.Context, event *types.PositionTransferEvent, blockNum uint64) error {
	// Only update if this is a position NFT (not other ERC721 transfers)
	// Check if token ID exists as position
	return h.eventRepo.UpdatePositionOwner(ctx, event.TokenID.Int64(), event.To.Hex())
}
