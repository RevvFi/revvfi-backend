package poller

import (
	"context"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/Revvfi/revvfi-backend/internal/indexer/registry"
	"github.com/Revvfi/revvfi-backend/internal/indexer/settings"
	"github.com/Revvfi/revvfi-backend/internal/repository/postgres"
)

type BlockPoller struct {
	ethClient        *ethclient.Client
	checkpointRepo   *postgres.CheckpointRepository
	config           *settings.Config
	contractRegistry *registry.ContractRegistry
}

func NewBlockPoller(
	ethClient *ethclient.Client,
	checkpointRepo *postgres.CheckpointRepository,
	config *settings.Config,
	contractRegistry *registry.ContractRegistry,
) *BlockPoller {
	return &BlockPoller{
		ethClient:        ethClient,
		checkpointRepo:   checkpointRepo,
		config:           config,
		contractRegistry: contractRegistry,
	}
}

func (p *BlockPoller) FetchLogs(ctx context.Context, fromBlock, toBlock uint64) ([]types.Log, error) {
	// Get dynamic list of addresses from contract registry
	addresses := p.contractRegistry.GetAddresses()

	log.Printf("Fetching logs from %d to %d (addresses=%d)", fromBlock, toBlock, len(addresses))

	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(fromBlock),
		ToBlock:   new(big.Int).SetUint64(toBlock),
		Addresses: addresses,
	}

	logs, err := p.ethClient.FilterLogs(ctx, query)
	if err != nil {
		return nil, err
	}
	log.Printf("Fetched %d logs for blocks %d-%d", len(logs), fromBlock, toBlock)
	return logs, nil
}
