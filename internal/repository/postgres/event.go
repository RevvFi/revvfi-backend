package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Revvfi/revvfi-backend/internal/models"
)

/*
@struct EventRepository

@desc
PostgreSQL implementation for event idempotency persistence.
*/
type EventRepository struct{ db *DB }

/*
@function NewEventRepository

@desc
Creates a PostgreSQL event repository.
*/
func NewEventRepository(db *DB) *EventRepository { return &EventRepository{db: db} }

/*
@method IsProcessed

@desc
Checks whether a blockchain log has already been processed.
*/
func (r *EventRepository) IsProcessed(ctx context.Context, txHash string, logIndex int32) (bool, error) {
	var exists bool
	err := r.db.conn.QueryRowContext(ctx, `select exists(select 1 from processed_events where tx_hash=$1 and log_index=$2)`, txHash, logIndex).Scan(&exists)
	return exists, mapError(err)
}

/*
@method MarkProcessed

@desc
Marks a blockchain event as processed.
*/
func (r *EventRepository) MarkProcessed(ctx context.Context, event *models.ProcessedEvent) error {
	_, err := r.db.conn.ExecContext(ctx, `insert into processed_events(tx_hash,log_index,block_number,event_name,contract_address,processed_at) values($1,$2,$3,$4,$5,$6) on conflict(tx_hash,log_index) do nothing`,
		event.TxHash, event.LogIndex, event.BlockNumber, event.EventName, event.ContractAddress, now())
	return mapError(err)
}

/*
@method StoreFailedEvent

@desc
Stores a failed event payload for retry.
*/
func (r *EventRepository) StoreFailedEvent(ctx context.Context, event map[string]interface{}, reason string) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = r.db.conn.ExecContext(ctx, `insert into failed_events(payload,reason,created_at) values($1,$2,$3)`, string(payload), reason, now())
	return mapError(err)
}

/*
@method GetFailedEvents

@desc
Lists failed events for retry.
*/
func (r *EventRepository) GetFailedEvents(ctx context.Context, limit int32) ([]map[string]interface{}, error) {
	rows, err := r.db.conn.QueryContext(ctx, `select payload from failed_events order by created_at asc limit $1`, limit)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, mapError(err)
		}
		var item map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, mapError(rows.Err())
}

/*
@method CleanupOldEvents

@desc
Deletes processed events older than a retention timestamp.
*/
func (r *EventRepository) CleanupOldEvents(ctx context.Context, olderThan time.Time) error {
	_, err := r.db.conn.ExecContext(ctx, `delete from processed_events where processed_at < $1`, olderThan)
	return mapError(err)
}
