package reorg

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/Revvfi/revvfi-backend/internal/repository/postgres"
)

/*@
 * Detector
 * @desc Compares saved checkpoints with canonical RPC blocks to detect reorgs.
 */
type Detector struct {
	checkpointRepo *postgres.CheckpointRepository
	ethClient      *ethclient.Client
}

/*@
 * NewDetector
 * @desc Creates a reorg detector backed by checkpoint storage and Ethereum RPC.
 */
func NewDetector(checkpointRepo *postgres.CheckpointRepository, ethClient *ethclient.Client) *Detector {
	return &Detector{checkpointRepo: checkpointRepo, ethClient: ethClient}
}

/*@
 * Check
 * @desc Verifies the last processed checkpoint is still canonical.
 */
func (d *Detector) Check(ctx context.Context, lastProcessedBlock uint64) error {
	if lastProcessedBlock == 0 {
		return nil
	}

	checkpoint, err := d.checkpointRepo.GetLastCheckpoint(ctx)
	if err != nil || checkpoint == nil {
		return err
	}
	if uint64(checkpoint.BlockNumber) > lastProcessedBlock {
		return nil
	}

	header, err := d.ethClient.HeaderByNumber(ctx, new(big.Int).SetInt64(checkpoint.BlockNumber))
	if err != nil {
		return fmt.Errorf("fetch checkpoint header: %w", err)
	}
	if checkpoint.BlockHash != header.Hash().Hex() {
		return fmt.Errorf("reorg at block %d: checkpoint=%s canonical=%s", checkpoint.BlockNumber, checkpoint.BlockHash, header.Hash().Hex())
	}
	return nil
}
