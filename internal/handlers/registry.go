// internal/indexer/handlers/registry.go
package handlers

import (
    "context"
    
    "github.com/Revvfi/revvfi-backend/internal/repository/postgres"
)

/*@
 * EventHandler interface
 * @desc All event handlers must implement this interface
 */
type EventHandler interface {
    Handle(ctx context.Context, event interface{}, blockNum uint64) error
}

/*@
 * Registry
 * @desc Central registry for all event handlers
 * @maps event names to their corresponding handlers
 */
type Registry struct {
    handlers map[string]EventHandler
    eventRepo *postgres.EventRepository
}

/*@
 * NewRegistry
 * @desc Creates a new registry and registers all event handlers
 * @param eventRepo Repository for database operations
 * @returns Configured Registry with all handlers registered
 */
func NewRegistry(eventRepo *postgres.EventRepository) *Registry {
    r := &Registry{
        handlers: make(map[string]EventHandler),
        eventRepo: eventRepo,
    }
    
// ============================================
// FACTORY EVENTS (RevvFiFactory) - 8 events
// ============================================
r.Register("MarketDeployed", NewMarketHandler(eventRepo))
r.Register("CoreContractsSet", NewCoreContractsHandler(eventRepo))
r.Register("ArchControllerUpdateRequested", NewArchControllerHandler(eventRepo))
r.Register("ArchControllerUpdated", NewArchControllerHandler(eventRepo))
r.Register("FeeUpdated", NewFeeHandler(eventRepo))
r.Register("DeploymentFeeUpdated", NewFeeHandler(eventRepo))  // ← ADD THIS
r.Register("ImplementationsSet", NewImplementationHandler(eventRepo))
r.Register("OwnershipTransferred", NewOwnershipHandler(eventRepo))

// ============================================
// MARKET EVENTS (RevvFiMarket) - 9 events
// ============================================
r.Register("Borrow", NewBorrowHandler(eventRepo))
r.Register("Repay", NewRepayHandler(eventRepo))
r.Register("InterestAccrued", NewInterestHandler(eventRepo))
r.Register("CollateralDeposited", NewCollateralHandler(eventRepo))
r.Register("CollateralWithdrawn", NewCollateralHandler(eventRepo))
r.Register("CollateralLiquidated", NewCollateralHandler(eventRepo))
r.Register("MarketClosedEvent", NewMarketHandler(eventRepo)) // contract emits "MarketClosedEvent"
r.Register("MarketPaused", NewMarketHandler(eventRepo))
r.Register("MarketUnpaused", NewMarketHandler(eventRepo))
r.Register("LiquidationExecuted", NewLiquidationHandler(eventRepo))
r.Register("LiquidationStartedMarket", NewLiquidationHandler(eventRepo))
r.Register("LiquidationEndedMarket", NewLiquidationHandler(eventRepo))
r.Register("ContractsSet", NewMarketHandler(eventRepo))    // signal-only, no DB write
r.Register("DrawdownExecuted", NewMarketHandler(eventRepo)) // complex array type, no DB write

// ============================================
// OFFERBOOK EVENTS (RevvFiOfferBook) - 5 events
// ============================================
r.Register("OfferSubmitted", NewOfferHandler(eventRepo))
r.Register("OfferCancelled", NewOfferHandler(eventRepo))
r.Register("OfferFilled", NewOfferHandler(eventRepo))
r.Register("OfferModified", NewOfferHandler(eventRepo))
r.Register("DrawdownExecutedOffer", NewOfferHandler(eventRepo))

// ============================================
// POSITION NFT EVENTS (RevvFiPositionNFT) - 5 events
// ============================================
r.Register("PositionMinted", NewPositionHandler(eventRepo))
r.Register("PositionSettled", NewPositionHandler(eventRepo))
r.Register("PositionRedeemed", NewPositionHandler(eventRepo))
r.Register("PositionClaimed", NewPositionHandler(eventRepo))
r.Register("Transfer", NewPositionHandler(eventRepo))

// ============================================
// LIQUIDATOR EVENTS (RevvFiLiquidator) - 4 events
// ============================================
r.Register("AuctionCreated", NewAuctionHandler(eventRepo))
r.Register("BidPlaced", NewAuctionHandler(eventRepo))
r.Register("AuctionSettled", NewAuctionHandler(eventRepo))
r.Register("AuctionCancelled", NewAuctionHandler(eventRepo))

// ============================================
// LIQUIDITY QUEUE EVENTS (RevvFiLiquidityQueue) - 4 events
// ============================================
r.Register("WithdrawalRequested", NewWithdrawalHandler(eventRepo))
r.Register("WithdrawalClaimed", NewWithdrawalHandler(eventRepo))
r.Register("WithdrawalCancelled", NewWithdrawalHandler(eventRepo))
r.Register("EpochProcessed", NewEpochHandler(eventRepo))

// ============================================
// REPUTATION REGISTRY EVENTS (ReputationRegistry) - 5 events
// ============================================
r.Register("ReputationScoreUpdated", NewReputationHandler(eventRepo))
r.Register("BorrowerRegistered", NewReputationHandler(eventRepo))
r.Register("SuccessfulRepaymentRecorded", NewReputationHandler(eventRepo))
r.Register("DefaultRecorded", NewReputationHandler(eventRepo))
r.Register("BorrowActivityRecorded", NewReputationHandler(eventRepo))

// ============================================
// ARCH CONTROLLER EVENTS (RevvFiArchController) - 14 events
// ============================================
r.Register("BorrowerAdded", NewArchControllerHandler(eventRepo))
r.Register("BorrowerRemoved", NewArchControllerHandler(eventRepo))
r.Register("ControllerAdded", NewArchControllerHandler(eventRepo))
r.Register("ControllerRemoved", NewArchControllerHandler(eventRepo))
r.Register("ControllerFactoryAdded", NewArchControllerHandler(eventRepo))
r.Register("ControllerFactoryRemoved", NewArchControllerHandler(eventRepo))
r.Register("MarketAdded", NewArchControllerHandler(eventRepo))
r.Register("MarketRemoved", NewArchControllerHandler(eventRepo))
r.Register("MarketRegistered", NewArchControllerHandler(eventRepo))
r.Register("MarketUnregistered", NewArchControllerHandler(eventRepo))
r.Register("OwnerUpdated", NewArchControllerHandler(eventRepo))
r.Register("TimelockUpdated", NewArchControllerHandler(eventRepo))
r.Register("AssetBlacklisted", NewArchControllerHandler(eventRepo))
r.Register("AssetPermitted", NewArchControllerHandler(eventRepo))

// ============================================
// COLLATERAL ESCROW EVENTS (RevvFiCollateralEscrow)
// ============================================
r.Register("MinCollateralRatioUpdated", NewCollateralHandler(eventRepo))
r.Register("LiquidationThresholdUpdated", NewCollateralHandler(eventRepo))

// ============================================
// COMMON / MISC EVENTS
// ============================================
r.Register("Initialized", NewMarketHandler(eventRepo)) // OZ initializer signal, no DB write
    return r
}

/*@
 * Register
 * @desc Registers an event handler for a specific event name
 * @param eventName Name of the event (must match contract event)
 * @param handler EventHandler implementation
 */
func (r *Registry) Register(eventName string, handler EventHandler) {
    r.handlers[eventName] = handler
}

/*@
 * Get
 * @desc Retrieves the handler for a given event name
 * @param eventName Name of the event
 * @returns EventHandler and boolean indicating if found
 */
func (r *Registry) Get(eventName string) (EventHandler, bool) {
    handler, ok := r.handlers[eventName]
    return handler, ok
}