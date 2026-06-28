package reorg

import (
	"context"
	"fmt"
)

/*@
 * Rollback
 * @desc Deletes checkpoints newer than the supplied canonical block.
 *
 * Event-level rollback is repository-specific and should be added alongside
 * domain tables; this checkpoint rollback keeps the worker from advancing on
 * a forked chain segment.
 */
func (d *Detector) Rollback(ctx context.Context, blockNumber uint64) error {
	if err := d.checkpointRepo.DeleteAfterBlock(ctx, blockNumber); err != nil {
		return fmt.Errorf("rollback checkpoints after block %d: %w", blockNumber, err)
	}
	return nil
}
