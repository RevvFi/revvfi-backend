package poller

import (
	"context"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/Revvfi/revvfi-backend/internal/indexer/settings"
	"github.com/Revvfi/revvfi-backend/internal/repository/postgres"
)

type BlockPoller struct {
	ethClient      *ethclient.Client
	checkpointRepo *postgres.CheckpointRepository
	config         *settings.Config
}

func NewBlockPoller(
	ethClient *ethclient.Client,
	checkpointRepo *postgres.CheckpointRepository,
	config *settings.Config,
) *BlockPoller {
	return &BlockPoller{
		ethClient:      ethClient,
		checkpointRepo: checkpointRepo,
		config:         config,
	}
}

func (p *BlockPoller) FetchLogs(ctx context.Context, fromBlock, toBlock uint64) ([]types.Log, error) {
	// Define contract addresses to monitor
	addresses := []common.Address{
		common.HexToAddress(p.config.FactoryAddress),
		common.HexToAddress(p.config.LiquidatorAddress),
		common.HexToAddress(p.config.PositionNFTAddress),
		common.HexToAddress(p.config.MarketAddress),        // ← ADD
        common.HexToAddress(p.config.OfferBookAddress),     // ← ADD
	}
 // Debug logging
    log.Printf(" Monitoring %d contract addresses", len(addresses))
    for _, addr := range addresses {
        log.Printf("   - %s", addr.Hex())
    }
	if p.config.ArchControllerAddress != "" {
		addresses = append(addresses, common.HexToAddress(p.config.ArchControllerAddress))
	}
	if p.config.ReputationRegistryAddress != "" {
		addresses = append(addresses, common.HexToAddress(p.config.ReputationRegistryAddress))
	}

	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(fromBlock),
		ToBlock:   new(big.Int).SetUint64(toBlock),
		Addresses: addresses,
	}

	// Diagnostic: log the filter being used
	log.Printf("Fetching logs from %d to %d (addresses=%d)", fromBlock, toBlock, len(addresses))
	logs, err := p.ethClient.FilterLogs(ctx, query)
	if err != nil {
		return nil, err
	}
	log.Printf("Fetched %d logs for blocks %d-%d", len(logs), fromBlock, toBlock)
	return logs, nil
}
