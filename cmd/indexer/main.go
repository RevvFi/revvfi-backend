package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Revvfi/revvfi-backend/internal/config"
	"github.com/Revvfi/revvfi-backend/internal/indexer"
	"github.com/Revvfi/revvfi-backend/internal/repository/postgres"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	// Load config (reuses existing config)
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to DB (reuses existing DB connection)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := postgres.Open(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create repositories (extend existing ones)
	eventRepo := postgres.NewEventRepository(db)
	checkpointRepo := postgres.NewCheckpointRepository(db)

	// Get existing RPC client (reuse blockchain package)
	ethClient, err := ethclient.DialContext(ctx, cfg.Blockchain.RPCURL)
	if err != nil {
		log.Fatalf("Failed to create RPC client: %v", err)
	}
	defer ethClient.Close()

	// Diagnostic: print RPC URL and network ID to ensure we're connected to the expected node
	log.Printf("Indexer RPC URL: %s", cfg.Blockchain.RPCURL)
	cfg.Blockchain.StartBlock = 0
 // will remove it just for the testing purpose     
	log.Printf("Starting from block: %d", cfg.Blockchain.StartBlock)  

	if netID, err := ethClient.NetworkID(ctx); err == nil {
		log.Printf("Connected to node network ID: %s", netID.String())
	} else {
		log.Printf("Failed to query network ID: %v", err)
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
	indexerCfg.MarketAddress = cfg.Blockchain.MarketAddress           // ← ADD
    indexerCfg.OfferBookAddress = cfg.Blockchain.OfferBookAddress     // ← ADD
	indexerCfg.StartBlock = int64(cfg.Blockchain.StartBlock)  

	worker, err := indexer.NewWorker(indexerCfg, ethClient, eventRepo, checkpointRepo)
	if err != nil {
		log.Fatalf("Failed to create indexer worker: %v", err)
	}

	// Setup signal handling (only ONCE)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start indexer in background
	if err := worker.Start(ctx); err != nil {
		log.Fatalf("Indexer failed: %v", err)
	}

	log.Println("Indexer started successfully. Waiting for blocks...")

	// Block here until signal received
	<-sigChan

	log.Println("Shutting down indexer...")
	worker.Stop()
	cancel()
}