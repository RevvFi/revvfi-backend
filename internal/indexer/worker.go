// internal/indexer/worker.go
package indexer

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/Revvfi/revvfi-backend/internal/handlers"
	"github.com/Revvfi/revvfi-backend/internal/indexer/decoder"
	"github.com/Revvfi/revvfi-backend/internal/indexer/helpers"
	"github.com/Revvfi/revvfi-backend/internal/indexer/idempotency"
	"github.com/Revvfi/revvfi-backend/internal/indexer/poller"
	"github.com/Revvfi/revvfi-backend/internal/indexer/processor"
	"github.com/Revvfi/revvfi-backend/internal/indexer/registry"
	"github.com/Revvfi/revvfi-backend/internal/indexer/reorg"
	"github.com/Revvfi/revvfi-backend/internal/logger"
	"github.com/Revvfi/revvfi-backend/internal/models"
	"github.com/Revvfi/revvfi-backend/internal/repository/postgres"
)

/*@
 * Worker
 * @desc Main indexer worker that orchestrates all components
 */
type Worker struct {
	config           *Config
	ethClient        *ethclient.Client
	eventRepo        *postgres.EventRepository
	checkpointRepo   *postgres.CheckpointRepository
	marketRepo       *postgres.MarketRepository
	contractRegistry *registry.ContractRegistry

	// Components
	poller         *poller.BlockPoller
	decoder        *decoder.EventDecoder
	idempotency    *idempotency.Tracker
	reorgDetector  *reorg.Detector
	eventProcessor *processor.EventProcessor
	dispatcher     *processor.EventDispatcher

	// State
	lastProcessedBlock uint64
	mu                 sync.RWMutex
	running            bool
	cancel             context.CancelFunc
	wg                 sync.WaitGroup
}

/*@
 * NewWorker
 * @desc Creates a new indexer worker
 */
func NewWorker(
	cfg *Config,
	ethClient *ethclient.Client,
	eventRepo *postgres.EventRepository,
	checkpointRepo *postgres.CheckpointRepository,
	marketRepo *postgres.MarketRepository,
) (*Worker, error) {
	// Initialize components
	eventDecoder, err := decoder.NewEventDecoder(cfg.ArtifactPath)
	if err != nil {
		return nil, err
	}

	// Create contract registry
	contractRegistry := registry.NewContractRegistry()

	// Add core contracts (always watch these)
	contractRegistry.AddContract(common.HexToAddress(cfg.FactoryAddress), "Factory")
	contractRegistry.AddContract(common.HexToAddress(cfg.ArchControllerAddress), "ArchController")
	contractRegistry.AddContract(common.HexToAddress(cfg.PositionNFTAddress), "PositionNFT")
	contractRegistry.AddContract(common.HexToAddress(cfg.LiquidatorAddress), "Liquidator")
	contractRegistry.AddContract(common.HexToAddress(cfg.ReputationRegistryAddress), "ReputationRegistry")

	log.Printf("Added %d core contracts to registry", contractRegistry.Count())

	// Create handler registry with contract registry
	handlerRegistry := handlers.NewRegistry(eventRepo, ethClient, contractRegistry)

	// Create idempotency tracker
	idempotencyTracker := idempotency.NewTracker(eventRepo)

	// Create reorg detector
	reorgDetector := reorg.NewDetector(checkpointRepo, ethClient)

	// Create event processor
	eventProcessor := processor.NewEventProcessor(
		eventDecoder,
		handlerRegistry,
		idempotencyTracker,
		eventRepo,
	)

	// Create dispatcher
	dispatcher := processor.NewEventDispatcher(
		eventDecoder,
		handlerRegistry,
		idempotencyTracker,
		eventRepo,
		5, // worker count
	)

	// Create block poller with contract registry
	blockPoller := poller.NewBlockPoller(ethClient, checkpointRepo, cfg, contractRegistry)

	worker := &Worker{
		config:             cfg,
		ethClient:          ethClient,
		eventRepo:          eventRepo,
		checkpointRepo:     checkpointRepo,
		marketRepo:         marketRepo,
		contractRegistry:   contractRegistry,
		poller:             blockPoller,
		decoder:            eventDecoder,
		idempotency:        idempotencyTracker,
		reorgDetector:      reorgDetector,
		eventProcessor:     eventProcessor,
		dispatcher:         dispatcher,
		lastProcessedBlock: uint64(cfg.StartBlock),
	}

	// Load existing markets from database
	ctx := context.Background()
	if err := worker.loadMarketsFromDatabase(ctx); err != nil {
		logger.WarnContext(ctx, "Failed to load markets from database",
			logger.WithError(err),
		)
		// Don't fail startup, just log warning
	}

	return worker, nil
}

/*@
 * Start
 * @desc Starts the indexer worker
 */
func (w *Worker) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	w.cancel = cancel
	w.running = true

	// Load last checkpoint
	if err := w.loadCheckpoint(ctx); err != nil {
		log.Printf("No checkpoint found, starting from block %d", w.config.StartBlock)
	}

	// Start dispatcher
	if err := w.dispatcher.Start(); err != nil {
		return err
	}

	// Start main loop
	w.wg.Add(1)
	go w.mainLoop(ctx)

	return nil
}

/*@
 * Stop
 * @desc Stops the indexer worker gracefully
 */
func (w *Worker) Stop() {
	log.Println("Stopping indexer worker...")
	w.running = false
	if w.cancel != nil {
		w.cancel()
	}
	w.dispatcher.Stop()
	w.wg.Wait()
	log.Println("Indexer worker stopped")
}

/*@
 * mainLoop
 * @desc Main processing loop
 */
func (w *Worker) mainLoop(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(w.config.PollInterval)
	defer ticker.Stop()

	for w.running {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.processNewBlocks(ctx); err != nil {
				log.Printf("Error processing blocks: %v", err)
			}
		}
	}
}

/*@
 * processNewBlocks
 * @desc Processes new blocks from the chain
 */
func (w *Worker) processNewBlocks(ctx context.Context) error {
	// Get current block number
	header, err := w.ethClient.HeaderByNumber(ctx, nil)
	if err != nil {
		return err
	}
	currentBlock := header.Number.Uint64()

	log.Printf("Worker state: lastProcessed=%d current=%d", w.lastProcessedBlock, currentBlock)
     log.Printf(
    "current=%d confirmations=%d",
    currentBlock,
    w.config.BlockConfirmations,
)
	// Calculate finalized block
	if currentBlock < uint64(w.config.BlockConfirmations) {
		return nil
	}
	finalizedBlock := currentBlock - uint64(w.config.BlockConfirmations)

	// Check for reorg
	if err := w.reorgDetector.Check(ctx, w.lastProcessedBlock); err != nil {
		log.Printf("Reorg detected, rolling back...")
		if err := w.reorgDetector.Rollback(ctx, w.lastProcessedBlock-10); err != nil {
			return err
		}
		w.lastProcessedBlock = w.lastProcessedBlock - 10
	}
       log.Printf(
    "Processing range %d -> %d",
    w.lastProcessedBlock+1,
    finalizedBlock,
)
	// Process blocks
	for blockNum := w.lastProcessedBlock + 1; blockNum <= finalizedBlock; blockNum++ {
		if err := w.processBlock(ctx, blockNum); err != nil {
			log.Printf("Failed to process block %d: %v", blockNum, err)
			return err
		}
		w.lastProcessedBlock = blockNum
		w.saveCheckpoint(ctx, blockNum)
	}

	return nil
}

/*@
 * processBlock
 * @desc Processes a single block with error handling
 */
func (w *Worker) processBlock(ctx context.Context, blockNum uint64) error {
    // Fetch logs
    log.Printf("Processing block %d", blockNum)
    logs, err := w.poller.FetchLogs(ctx, blockNum, blockNum)
    if err != nil {
        // Save block processing error to failed_events
        if saveErr := w.eventRepo.SaveFailedEvent(ctx, "BlockFetch", "", blockNum, err.Error()); saveErr != nil {
            log.Printf("Failed to store block fetch error: %v", saveErr)
        }
        return err
    }

    // Dispatch logs to workers
    for _, logEntry := range logs {
        // Get event name from topics
        if len(logEntry.Topics) == 0 {
            w.eventRepo.SaveFailedEvent(ctx, "UnknownEvent", logEntry.TxHash.Hex(), blockNum, "No topics in log")
            continue
        }
        
        eventName := decoder.GetEventName(logEntry.Topics[0].Hex())
        if eventName == "" {
            w.eventRepo.SaveFailedEvent(ctx, "UnknownEvent", logEntry.TxHash.Hex(), blockNum, "Unknown event topic: "+logEntry.Topics[0].Hex())
            continue
        }
        
        if err := w.dispatcher.Dispatch(logEntry, blockNum); err != nil {
            // Save failed event processing to database
            if saveErr := w.eventRepo.SaveFailedEvent(ctx, eventName, logEntry.TxHash.Hex(), blockNum, err.Error()); saveErr != nil {
                log.Printf("Failed to store failed event: %v", saveErr)
            }
            log.Printf("Worker: failed to process log: %v", err)
        }
    }

    return nil
}
/*@
 * loadCheckpoint
 * @desc Loads last processed block from database
 */
func (w *Worker) loadCheckpoint(ctx context.Context) error {
	checkpoint, err := w.checkpointRepo.GetLastCheckpoint(ctx)
	if err != nil {
		return err
	}
	if checkpoint != nil {
		// If the checkpoint in DB is ahead of the node's current block (e.g. different node
		// or node reset), we should avoid using that larger value because the poller will
		// skip processing (lastProcessed > current). Fetch the node's current header and
		// clamp the checkpoint to the node's current block if necessary.
		currentHeader, err := w.ethClient.HeaderByNumber(ctx, nil)
		if err != nil {
			// Couldn't query node; fall back to stored checkpoint but log the error.
			w.lastProcessedBlock = uint64(checkpoint.BlockNumber)
			log.Printf("Loaded checkpoint: block %d (failed to query node current block: %v)", w.lastProcessedBlock, err)
			return nil
		}
		currentBlock := currentHeader.Number.Uint64()
		if uint64(checkpoint.BlockNumber) > currentBlock {
			// Stored checkpoint is in the future relative to this node. Reset to currentBlock
			// so the poller can make progress. This is a safe recovery when switching nodes
			// or after node reorgs/resets in dev environments.
			log.Printf("WARNING: stored checkpoint (%d) > node current block (%d). Resetting lastProcessed to %d", checkpoint.BlockNumber, currentBlock, currentBlock)
			w.lastProcessedBlock = currentBlock
		} else {
			w.lastProcessedBlock = uint64(checkpoint.BlockNumber)
		}
		log.Printf("Loaded checkpoint: block %d", w.lastProcessedBlock)
	}
	return nil
}

/*@
 * saveCheckpoint
 * @desc Saves current checkpoint to database
 */
func (w *Worker) saveCheckpoint(ctx context.Context, blockNum uint64) error {
	block, err := w.ethClient.BlockByNumber(ctx, new(big.Int).SetUint64(blockNum))
	if err != nil {
		return err
	}

	return w.checkpointRepo.SaveCheckpoint(ctx, &models.ChainCheckpoint{
		BlockNumber: int64(blockNum),
		BlockHash:   block.Hash().Hex(),
		ParentHash:  block.ParentHash().Hex(),
		ProcessedAt: time.Now(),
	})
}

/*
 * loadMarketsFromDatabase
 * @desc Loads all known markets from database and adds them to the registry
 */
func (w *Worker) loadMarketsFromDatabase(ctx context.Context) error {
	// Get all active markets from database
	markets, err := w.marketRepo.GetAllActive(ctx)
	if err != nil {
		return fmt.Errorf("failed to get markets: %w", err)
	}

	log.Printf("Loading %d markets from database", len(markets))

	// Add each market and query its child contracts
	for _, market := range markets {
		marketAddr := common.HexToAddress(market.Address)

		// Add market itself
		w.contractRegistry.AddContract(marketAddr, "Market")

		// Query and add child contracts
		children, err := helpers.GetMarketChildContracts(ctx, w.ethClient, marketAddr)
		if err != nil {
			log.Printf("Failed to query market %s children, skipping: %v", market.Address, err)
			continue
		}

		w.contractRegistry.AddContract(children.OfferBook, "OfferBook")
		w.contractRegistry.AddContract(children.CollateralEscrow, "CollateralEscrow")

		// LiquidityQueue is optional
		if children.LiquidityQueue != (common.Address{}) {
			w.contractRegistry.AddContract(children.LiquidityQueue, "LiquidityQueue")
		}

		log.Printf("Added market %s and children to registry (OfferBook: %s)",
			market.Address, children.OfferBook.Hex())
	}

	log.Printf("Finished loading markets, now watching %d contracts total", w.contractRegistry.Count())

	return nil
}
