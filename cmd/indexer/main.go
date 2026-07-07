package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Revvfi/revvfi-backend/internal/config"
	"github.com/Revvfi/revvfi-backend/internal/indexer"
	"github.com/Revvfi/revvfi-backend/internal/logger"
	"github.com/Revvfi/revvfi-backend/internal/repository/postgres"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	// Load config (reuses existing config)
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize structured logger FIRST
	env := logger.EnvDevelopment
	if cfg.Environment == "production" {
		env = logger.EnvProduction
	} else if cfg.Environment == "staging" {
		env = logger.EnvStaging
	}
	logger.Init(logger.ServiceIndexer, env)

	logger.Info("RevvFi Indexer starting",
		"environment", env,
		"version", "1.0.0",
	)

	// Connect to DB (reuses existing DB connection)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := postgres.Open(ctx, cfg.Database)
	if err != nil {
		logger.Error("Failed to connect to database",
			logger.WithError(err),
		)
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	logger.Info("Database connection established")

	// Create repositories (extend existing ones)
	eventRepo := postgres.NewEventRepository(db)
	checkpointRepo := postgres.NewCheckpointRepository(db)
	marketRepo := postgres.NewMarketRepository(db)

	// Get existing RPC client (reuse blockchain package)
	logger.Info("Connecting to Ethereum RPC",
		"rpc_url", cfg.Blockchain.RPCURL,
	)
	ethClient, err := ethclient.DialContext(ctx, cfg.Blockchain.RPCURL)
	if err != nil {
		logger.Error("Failed to create RPC client",
			logger.WithError(err),
			"rpc_url", cfg.Blockchain.RPCURL,
		)
		log.Fatalf("Failed to create RPC client: %v", err)
	}
	defer ethClient.Close()

	logger.Info("RPC connection established",
		"rpc_url", cfg.Blockchain.RPCURL,
		"start_block", cfg.Blockchain.StartBlock,
	)

	if netID, err := ethClient.NetworkID(ctx); err == nil {
		logger.Info("Network ID verified",
			"network_id", netID.String(),
			"chain_id", cfg.Blockchain.ChainID,
		)
	} else {
		logger.Warn("Failed to query network ID",
			logger.WithError(err),
		)
	}

	// Create indexer worker (new)
	indexerCfg := indexer.DefaultConfig()
	indexerCfg.RPCURL = cfg.Blockchain.RPCURL
	indexerCfg.ChainID = cfg.Blockchain.ChainID
	indexerCfg.ArtifactPath = cfg.Blockchain.ArtifactPath
	indexerCfg.FactoryAddress = cfg.Blockchain.FactoryAddress
	indexerCfg.ArchControllerAddress = cfg.Blockchain.ArchControllerAddress
	indexerCfg.PositionNFTAddress = cfg.Blockchain.PositionNFTAddress
	indexerCfg.LiquidatorAddress = cfg.Blockchain.LiquidatorAddress
	indexerCfg.ReputationRegistryAddress = cfg.Blockchain.ReputationRegistryAddress
	indexerCfg.MarketAddress = cfg.Blockchain.MarketAddress
	indexerCfg.OfferBookAddress = cfg.Blockchain.OfferBookAddress
	indexerCfg.CollateralEscrowAddress = cfg.Blockchain.CollateralEscrowAddress
	indexerCfg.StartBlock = int64(cfg.Blockchain.StartBlock)
	indexerCfg.BlockConfirmations = cfg.Blockchain.BlockConfirmations
	indexerCfg.PollInterval = cfg.Blockchain.SyncBlockInterval
	indexerCfg.BatchSize = cfg.Blockchain.MaxBlockBatchSize

	logger.Info("Creating indexer worker",
		logger.WithContractAddress(cfg.Blockchain.FactoryAddress),
		"arch_controller", cfg.Blockchain.ArchControllerAddress,
		"position_nft", cfg.Blockchain.PositionNFTAddress,
		"liquidator", cfg.Blockchain.LiquidatorAddress,
	)

	worker, err := indexer.NewWorker(indexerCfg, ethClient, eventRepo, checkpointRepo, marketRepo)
	if err != nil {
		logger.Error("Failed to create indexer worker",
			logger.WithError(err),
		)
		log.Fatalf("Failed to create indexer worker: %v", err)
	}

	// Setup signal handling (only ONCE)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start indexer in background
	logger.Info("Starting indexer worker")
	if err := worker.Start(ctx); err != nil {
		logger.Error("Indexer failed to start",
			logger.WithError(err),
		)
		log.Fatalf("Indexer failed: %v", err)
	}

	logger.Info("Indexer started successfully, waiting for blocks...",
		logger.WithFromBlock(uint64(cfg.Blockchain.StartBlock)),
	)

	// Block here until signal received
	<-sigChan

	logger.Info("Shutdown signal received, stopping indexer...")
	worker.Stop()
	cancel()
	logger.Info("Indexer shutdown complete")
}