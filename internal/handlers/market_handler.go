// internal/indexer/handlers/market_handler.go
package handlers

import (
    "context"
    "time"

    "github.com/Revvfi/revvfi-backend/internal/models"
    "github.com/Revvfi/revvfi-backend/internal/repository/postgres"
    "github.com/Revvfi/revvfi-backend/internal/indexer/types"
)

/*@
 * MarketHandler
 * @desc Handles all market-related blockchain events
 * @events
 *   - MarketDeployed: New lending market created
 *   - MarketPaused: Market emergency paused
 *   - MarketUnpaused: Market resumed
 *   - InterestAccrued: Interest accrued on market
 */
type MarketHandler struct {
    eventRepo *postgres.EventRepository
}

/*@
 * NewMarketHandler
 * @desc Creates a new market handler instance
 * @param eventRepo Repository for database operations
 * @returns Configured MarketHandler
 */
func NewMarketHandler(eventRepo *postgres.EventRepository) *MarketHandler {
    return &MarketHandler{eventRepo: eventRepo}
}

/*@
 * Handle
 * @desc Routes market events to appropriate handler methods
 * @param ctx Context for cancellation and deadlines
 * @param event The decoded event from blockchain
 * @param blockNum Block number where event occurred
 * @returns error if handling fails
 */
func (h *MarketHandler) Handle(ctx context.Context, event interface{}, blockNum uint64) error {
    switch e := event.(type) {
    case *types.MarketDeployedEvent:
        return h.handleMarketDeployed(ctx, e, blockNum)
    case *types.MarketPausedEvent:
        return h.handleMarketPaused(ctx, e, blockNum)
    case *types.MarketUnpausedEvent:
        return h.handleMarketUnpaused(ctx, e, blockNum)
    case *types.InterestAccruedEvent:
        return h.handleInterestAccrued(ctx, e, blockNum)
    }
    return nil
}

/*@
 * handleMarketDeployed
 * @desc Processes MarketDeployed event - creates new market record
 * @param event Decoded MarketDeployed event
 * @param blockNum Block number where market was deployed
 */
func (h *MarketHandler) handleMarketDeployed(ctx context.Context, event *types.MarketDeployedEvent, blockNum uint64) error {
    market := &models.Market{
        Address:              event.MarketAddress.Hex(),
        Borrower:             event.Borrower.Hex(),
        BorrowAsset:          event.BorrowAsset.Hex(),
        BorrowAssetDecimals:  18, // Will be updated from config
        CollateralAsset:      event.CollateralAsset.Hex(),
        CollateralAssetDecimals: 18,
        CollateralOracle:     event.Oracle.Hex(),
        MinCollateralRatio:   int32(event.MinCollateralRatio.Int64()),
        // removed not required LiquidationThreshold: int32(event.LiquidationThreshold.Int64()),
        IsActive:             true,
        CreatedAt:            time.Now(),
        LastInterestAccrual:  time.Now(),
    }
    
    return h.eventRepo.SaveMarket(ctx, market)
}

/*@
 * handleMarketPaused
 * @desc Processes MarketPaused event - updates market status
 * @param event Decoded MarketPaused event
 * @param blockNum Block number where pause occurred
 */
func (h *MarketHandler) handleMarketPaused(ctx context.Context, event *types.MarketPausedEvent, blockNum uint64) error {
    return h.eventRepo.UpdateMarketStatus(ctx, event.Market.Hex(), false, true)
}

/*@
 * handleMarketUnpaused
 * @desc Processes MarketUnpaused event - updates market status
 * @param event Decoded MarketUnpaused event
 * @param blockNum Block number where unpause occurred
 */
func (h *MarketHandler) handleMarketUnpaused(ctx context.Context, event *types.MarketUnpausedEvent, blockNum uint64) error {
    return h.eventRepo.UpdateMarketStatus(ctx, event.Market.Hex(), true, false)
}

/*@
 * handleInterestAccrued
 * @desc Processes InterestAccrued event - updates market borrow index
 * @param event Decoded InterestAccrued event
 * @param blockNum Block number where interest accrued
 */
func (h *MarketHandler) handleInterestAccrued(ctx context.Context, event *types.InterestAccruedEvent, blockNum uint64) error {
    return h.eventRepo.UpdateMarketBorrowIndex(ctx, event.Market.Hex())
}