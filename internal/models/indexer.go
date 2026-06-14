package models

import "time"

/*
@model ProcessedEvent

@desc
Tracks blockchain events that have already been processed.

Used for:
- Idempotency protection
- Preventing duplicate event handling
- Recovery after service restart

Unique key:
(tx_hash, log_index)
*/
type ProcessedEvent struct {
	// Internal database ID
	ID int64 `db:"id"`

	// Transaction hash containing the event
	TxHash string `db:"tx_hash"`

	// Event log index inside the transaction
	LogIndex int32 `db:"log_index"`

	// Block where event was emitted
	BlockNumber int64 `db:"block_number"`

	// Solidity event name
	EventName string `db:"event_name"`

	// Contract that emitted the event
	ContractAddress string `db:"contract_address"`

	// Timestamp when indexer processed this event
	ProcessedAt time.Time `db:"processed_at"`
}

/*
@model ChainCheckpoint

@desc
Stores processed blockchain checkpoints.

Used for:
- Reorg protection
- Finality tracking
- Recovery after restart
- Confirmation monitoring

One row per block.
*/
type ChainCheckpoint struct {
	// Internal database ID
	ID int64 `db:"id"`

	// Blockchain block number
	BlockNumber int64 `db:"block_number"`

	// Hash of current block
	BlockHash string `db:"block_hash"`

	// Hash of parent block
	ParentHash string `db:"parent_hash"`

	// Whether block is considered finalized
	IsFinalized bool `db:"is_finalized"`

	// Number of confirmations observed
	ConfirmationCount int `db:"confirmation_count"`

	// Timestamp when checkpoint was recorded
	ProcessedAt time.Time `db:"processed_at"`
}



