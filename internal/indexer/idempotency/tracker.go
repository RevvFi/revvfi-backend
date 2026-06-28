package idempotency

import (
	"context"

	"github.com/Revvfi/revvfi-backend/internal/models"
	"github.com/Revvfi/revvfi-backend/internal/repository/postgres"
)

/*@
 * Tracker
 * @desc Thin indexer adapter around processed_events persistence.
 *
 * The processor owns blockchain log metadata, while EventRepository owns the
 * database schema. This adapter keeps processor code focused on flow control
 * and prevents duplicate event handling across restarts.
 */
type Tracker struct {
	eventRepo *postgres.EventRepository
}

/*@
 * NewTracker
 * @desc Creates an idempotency tracker backed by EventRepository.
 */
func NewTracker(eventRepo *postgres.EventRepository) *Tracker {
	return &Tracker{eventRepo: eventRepo}
}

/*@
 * IsProcessed
 * @desc Returns true when tx_hash/log_index already exists in processed_events.
 */
func (t *Tracker) IsProcessed(ctx context.Context, txHash string, logIndex int) (bool, error) {
	return t.eventRepo.IsProcessed(ctx, txHash, logIndex)
}

/*@
 * MarkProcessed
 * @desc Persists the processed-event marker after successful handling.
 */
func (t *Tracker) MarkProcessed(ctx context.Context, txHash string, logIndex int, blockNum uint64, eventName, contractAddress string) error {
	return t.eventRepo.MarkProcessed(ctx, &models.ProcessedEvent{
		TxHash:          txHash,
		LogIndex:        int32(logIndex),
		BlockNumber:     int64(blockNum),
		EventName:       eventName,
		ContractAddress: contractAddress,
	})
}
