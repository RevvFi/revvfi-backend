// internal/indexer/handlers/market_handler.go
package handlers

import (
	"context"
	"strings"
	"time"

	"github.com/Revvfi/revvfi-backend/internal/indexer/helpers"
	"github.com/Revvfi/revvfi-backend/internal/indexer/registry"
	"github.com/Revvfi/revvfi-backend/internal/indexer/types"
	"github.com/Revvfi/revvfi-backend/internal/logger"
	"github.com/Revvfi/revvfi-backend/internal/models"
	"github.com/Revvfi/revvfi-backend/internal/repository/postgres"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
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
	eventRepo        *postgres.EventRepository
	ethClient        *ethclient.Client
	contractRegistry *registry.ContractRegistry
}

/*@
 * NewMarketHandler
 * @desc Creates a new market handler instance
 * @param eventRepo Repository for database operations
 * @param ethClient Ethereum client for blockchain calls
 * @param contractRegistry Registry for dynamically tracking contracts
 * @returns Configured MarketHandler
 */
func NewMarketHandler(
	eventRepo *postgres.EventRepository,
	ethClient *ethclient.Client,
	contractRegistry *registry.ContractRegistry,
) *MarketHandler {
	return &MarketHandler{
		eventRepo:        eventRepo,
		ethClient:        ethClient,
		contractRegistry: contractRegistry,
	}
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
	case *types.MarketClosedEvent:
		return h.handleMarketClosed(ctx, e, blockNum)
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
	start := time.Now()

	logger.InfoContext(ctx, "Processing MarketDeployed event",
		logger.WithEventName("MarketDeployed"),
		logger.WithBlockNumber(blockNum),
		logger.WithMarketAddress(event.MarketAddress.Hex()),
		logger.WithBorrower(event.Borrower.Hex()),
		logger.WithBorrowAsset(event.BorrowAsset.Hex()),
		logger.WithCollateralAsset(event.CollateralAsset.Hex()),
	)

	// Fetch actual decimals from token contracts
	borrowDecimals := h.getTokenDecimals(ctx, event.BorrowAsset)
	collateralDecimals := h.getTokenDecimals(ctx, event.CollateralAsset)

	logger.InfoContext(ctx, "Fetched token decimals",
		"borrow_asset", event.BorrowAsset.Hex(),
		"borrow_decimals", borrowDecimals,
		"collateral_asset", event.CollateralAsset.Hex(),
		"collateral_decimals", collateralDecimals,
	)

	market := &models.Market{
		Address:                 event.MarketAddress.Hex(),
		Borrower:                event.Borrower.Hex(),
		BorrowAsset:             event.BorrowAsset.Hex(),
		BorrowAssetDecimals:     borrowDecimals,
		CollateralAsset:         event.CollateralAsset.Hex(),
		CollateralAssetDecimals: collateralDecimals,
		CollateralOracle:        event.Oracle.Hex(),
		MinCollateralRatio:      int32(event.MinCollateralRatio.Int64()),
		// removed not required LiquidationThreshold: int32(event.LiquidationThreshold.Int64()),
		IsActive:            true,
		CreatedAt:           time.Now(),
		LastInterestAccrual: time.Now(),
	}

	if err := h.eventRepo.SaveMarket(ctx, market); err != nil {
		logger.ErrorContext(ctx, "Failed to save market to database",
			logger.WithError(err),
			logger.WithMarketAddress(event.MarketAddress.Hex()),
		)
		return err
	}

	logger.InfoContext(ctx, "Market saved to database successfully",
		logger.WithMarketAddress(event.MarketAddress.Hex()),
		logger.WithDuration(time.Since(start)),
	)

	// Add market to contract registry for dynamic watching
	if h.contractRegistry != nil {
		marketAddr := event.MarketAddress
		h.contractRegistry.AddContract(marketAddr, "Market")

		// Query and add child contracts
		children, err := helpers.GetMarketChildContracts(ctx, h.ethClient, marketAddr)
		if err != nil {
			logger.WarnContext(ctx, "Failed to query market children, will not watch child contracts",
				"market", marketAddr.Hex(),
				logger.WithError(err),
			)
		} else {
			h.contractRegistry.AddContract(children.OfferBook, "OfferBook")
			h.contractRegistry.AddContract(children.CollateralEscrow, "CollateralEscrow")

			// LiquidityQueue is optional
			if children.LiquidityQueue != (common.Address{}) {
				h.contractRegistry.AddContract(children.LiquidityQueue, "LiquidityQueue")
			}

			logger.InfoContext(ctx, "Added new market and children to indexer watch list",
				"market", marketAddr.Hex(),
				"offerBook", children.OfferBook.Hex(),
				"escrow", children.CollateralEscrow.Hex(),
				"total_watching", h.contractRegistry.Count(),
			)
		}
	}

	return nil
}

// getTokenDecimals fetches the decimals from an ERC20 token contract
func (h *MarketHandler) getTokenDecimals(ctx context.Context, tokenAddress common.Address) int16 {
	// ERC20 decimals() function ABI
	decimalsABI := `[{"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint8"}],"payable":false,"stateMutability":"view","type":"function"}]`

	// Parse ABI
	parsedABI, err := abi.JSON(strings.NewReader(decimalsABI))
	if err != nil {
		logger.WarnContext(ctx, "Failed to parse ERC20 ABI",
			logger.WithError(err),
			"token", tokenAddress.Hex(),
		)
		return 18 // Default to 18 if ABI parse fails
	}

	// Create bound contract
	contract := bind.NewBoundContract(tokenAddress, parsedABI, h.ethClient, h.ethClient, h.ethClient)

	// Call decimals()
	var result []interface{}
	err = contract.Call(&bind.CallOpts{Context: ctx}, &result, "decimals")
	if err != nil {
		logger.WarnContext(ctx, "Failed to fetch token decimals, using default 18",
			logger.WithError(err),
			"token", tokenAddress.Hex(),
		)
		return 18
	}

	if len(result) == 0 {
		return 18
	}

	// Convert result to uint8 then int16
	decimals := result[0].(uint8)
	return int16(decimals)
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

/*@
 * handleMarketClosed
 * @desc Processes MarketClosedEvent - permanently closes the market
 * @param event Decoded MarketClosed event containing borrower and timestamp
 * @param blockNum Block number where market was closed
 * @note Sets is_closed=true and is_active=false; closed markets cannot be reopened.
 *       The Borrower field identifies which borrower's market was closed.
 */
func (h *MarketHandler) handleMarketClosed(ctx context.Context, event *types.MarketClosedEvent, blockNum uint64) error {
	logger.InfoContext(ctx, "Processing MarketClosed event",
		logger.WithEventName("MarketClosed"),
		logger.WithBlockNumber(blockNum),
		logger.WithBorrower(event.Borrower.Hex()),
		"timestamp", event.Timestamp.Int64(),
	)

	if err := h.eventRepo.CloseMarket(ctx, event.Borrower.Hex()); err != nil {
		logger.ErrorContext(ctx, "Failed to close market",
			logger.WithError(err),
			logger.WithBorrower(event.Borrower.Hex()),
		)
		return err
	}

	logger.InfoContext(ctx, "Market closed successfully",
		logger.WithBorrower(event.Borrower.Hex()),
	)

	return nil
}
