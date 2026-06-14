package postgres

import (
	"context"
	"database/sql"
	"github.com/Revvfi/revvfi-backend/internal/models"
)

type CheckpointRepository struct {
    db *DB
}

func NewCheckpointRepository(db *DB) *CheckpointRepository {
    return &CheckpointRepository{db: db}
}

func (r *CheckpointRepository) SaveCheckpoint(ctx context.Context, checkpoint *models.ChainCheckpoint) error {
    _, err := r.db.conn.ExecContext(ctx, `
        INSERT INTO chain_checkpoints(block_number, block_hash, parent_hash, processed_at)
        VALUES($1, $2, $3, $4)
        ON CONFLICT (block_number) DO UPDATE SET
            block_hash = EXCLUDED.block_hash,
            processed_at = EXCLUDED.processed_at
    `, checkpoint.BlockNumber, checkpoint.BlockHash, checkpoint.ParentHash, checkpoint.ProcessedAt)
    return mapError(err)
}

func (r *CheckpointRepository) GetLastCheckpoint(ctx context.Context) (*models.ChainCheckpoint, error) {
    var checkpoint models.ChainCheckpoint
    err := r.db.conn.QueryRowContext(ctx, `
        SELECT block_number, block_hash, processed_at
        FROM chain_checkpoints
        ORDER BY block_number DESC
        LIMIT 1
    `).Scan(&checkpoint.BlockNumber, &checkpoint.BlockHash, &checkpoint.ProcessedAt)
    
    if err == sql.ErrNoRows {
        return nil, nil
    }
    return &checkpoint, mapError(err)
}

func (r *CheckpointRepository) DeleteAfterBlock(ctx context.Context, blockNumber uint64) error {
    _, err := r.db.conn.ExecContext(ctx, `
        DELETE FROM chain_checkpoints WHERE block_number > $1
    `, blockNumber)
    
    return mapError(err)
}