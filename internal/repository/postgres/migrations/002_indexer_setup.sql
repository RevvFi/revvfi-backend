/*@
file: 002_indexer_setup.sql
version: 1.2
date: June 6, 2026

description:
Indexer infrastructure tables for RevvFi.

improvements:
- Added parent_hash and depth to reorg_events
- Added resolved_at to failed_events
- Added block_timestamp to processed_events
- Added processed flag for archive strategy
*/

/*@
section: 1. PROCESSED EVENTS (Idempotency)
*/

CREATE TABLE IF NOT EXISTS processed_events (
    id BIGSERIAL PRIMARY KEY,
    tx_hash TEXT NOT NULL,
    log_index INTEGER NOT NULL,
    block_number BIGINT NOT NULL,
    block_timestamp TIMESTAMPTZ,  -- ← Added
    event_name TEXT NOT NULL,
    contract_address TEXT NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT unique_processed_event UNIQUE(tx_hash, log_index)
);

CREATE INDEX IF NOT EXISTS idx_processed_events_tx ON processed_events(tx_hash);
CREATE INDEX IF NOT EXISTS idx_processed_events_block ON processed_events(block_number);
CREATE INDEX IF NOT EXISTS idx_processed_events_name ON processed_events(event_name);
CREATE INDEX IF NOT EXISTS idx_processed_events_processed_at ON processed_events(processed_at);

COMMENT ON TABLE processed_events IS 'Tracks processed blockchain events for idempotency protection';

/*@
section: 2. CHAIN CHECKPOINTS (Reorg Protection)
*/

CREATE TABLE IF NOT EXISTS chain_checkpoints (
    id BIGSERIAL PRIMARY KEY,
    block_number BIGINT NOT NULL UNIQUE,
    block_hash TEXT NOT NULL,
    parent_hash TEXT NOT NULL,
    is_finalized BOOLEAN NOT NULL DEFAULT FALSE,
    confirmation_count INTEGER NOT NULL DEFAULT 0,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chain_checkpoints_block ON chain_checkpoints(block_number DESC);
CREATE INDEX IF NOT EXISTS idx_chain_checkpoints_finalized ON chain_checkpoints(is_finalized);
CREATE INDEX IF NOT EXISTS idx_chain_checkpoints_processed ON chain_checkpoints(processed_at);

COMMENT ON TABLE chain_checkpoints IS 'Tracks blockchain checkpoints for reorg protection';

/*@
section: 3. FAILED EVENTS (Dead Letter Queue)
*/

-- NOTE: 001_initial_setup.sql already creates failed_events with a minimal
-- schema (id, payload, reason, created_at), so this CREATE TABLE is a no-op
-- on any database that ran 001 first. The columns this migration actually
-- needs are added via ALTER TABLE ... ADD COLUMN IF NOT EXISTS below so they
-- land regardless of whether the table pre-existed.
CREATE TABLE IF NOT EXISTS failed_events (
    id BIGSERIAL PRIMARY KEY,
    event_name TEXT NOT NULL,
    event_data JSONB NOT NULL,
    raw_log JSONB,
    failure_reason TEXT NOT NULL,
    retry_count INTEGER NOT NULL DEFAULT 0,
    max_retries INTEGER NOT NULL DEFAULT 5,
    next_retry_at TIMESTAMPTZ,
    processed BOOLEAN DEFAULT FALSE,  -- ← Added
    resolved_at TIMESTAMPTZ,  -- ← Added
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE failed_events ADD COLUMN IF NOT EXISTS event_name TEXT;
ALTER TABLE failed_events ADD COLUMN IF NOT EXISTS event_data JSONB;
ALTER TABLE failed_events ADD COLUMN IF NOT EXISTS raw_log JSONB;
ALTER TABLE failed_events ADD COLUMN IF NOT EXISTS failure_reason TEXT;
ALTER TABLE failed_events ADD COLUMN IF NOT EXISTS retry_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE failed_events ADD COLUMN IF NOT EXISTS max_retries INTEGER NOT NULL DEFAULT 5;
ALTER TABLE failed_events ADD COLUMN IF NOT EXISTS next_retry_at TIMESTAMPTZ;
ALTER TABLE failed_events ADD COLUMN IF NOT EXISTS processed BOOLEAN DEFAULT FALSE;
ALTER TABLE failed_events ADD COLUMN IF NOT EXISTS resolved_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_failed_events_retry ON failed_events(next_retry_at) WHERE processed = FALSE;
CREATE INDEX IF NOT EXISTS idx_failed_events_created ON failed_events(created_at);
CREATE INDEX IF NOT EXISTS idx_failed_events_name ON failed_events(event_name);

COMMENT ON TABLE failed_events IS 'Dead letter queue for failed blockchain event processing';

/*@
section: 4. INDEXER SYNC STATE
*/

CREATE TABLE IF NOT EXISTS indexer_sync_state (
    id BIGSERIAL PRIMARY KEY,
    service_name TEXT NOT NULL UNIQUE,
    last_processed_block BIGINT NOT NULL,
    last_processed_block_hash TEXT,
    last_processed_timestamp TIMESTAMPTZ,
    sync_status TEXT NOT NULL DEFAULT 'syncing',
    lag_blocks INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_indexer_sync_service ON indexer_sync_state(service_name);
CREATE INDEX IF NOT EXISTS idx_indexer_sync_status ON indexer_sync_state(sync_status);

COMMENT ON TABLE indexer_sync_state IS 'Tracks indexer synchronization progress';

/*@
section: 5. REORG AUDIT TABLE
*/

CREATE TABLE IF NOT EXISTS reorg_events (
    id BIGSERIAL PRIMARY KEY,
    old_block_number BIGINT NOT NULL,
    old_block_hash TEXT NOT NULL,
    new_block_hash TEXT NOT NULL,
    parent_hash TEXT,  -- ← Added
    depth INTEGER DEFAULT 1,  -- ← Added
    detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_reorg_events_block ON reorg_events(old_block_number);
CREATE INDEX IF NOT EXISTS idx_reorg_events_detected_at ON reorg_events(detected_at);

COMMENT ON TABLE reorg_events IS 'Audit log of detected blockchain reorganizations';

/*@
section: 6. VERIFICATION
*/

DO $$
DECLARE
    expected_tables TEXT[] := ARRAY[
        'processed_events',
        'chain_checkpoints',
        'failed_events',
        'indexer_sync_state',
        'reorg_events'
    ];
    current_table TEXT;
    missing_tables TEXT[] := ARRAY[]::TEXT[];
BEGIN
    FOREACH current_table IN ARRAY expected_tables
    LOOP
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.tables 
            WHERE table_name = current_table AND table_schema = 'public'
        ) THEN
            missing_tables := array_append(missing_tables, current_table);
        END IF;
    END LOOP;
    
    IF array_length(missing_tables, 1) > 0 THEN
        RAISE NOTICE 'Missing tables: %', missing_tables;
    ELSE
        RAISE NOTICE '';
        RAISE NOTICE '========================================';
        RAISE NOTICE 'INDEXER SETUP COMPLETE';
        RAISE NOTICE '========================================';
        RAISE NOTICE 'processed_events';
        RAISE NOTICE 'chain_checkpoints';
        RAISE NOTICE 'failed_events';
        RAISE NOTICE 'indexer_sync_state';
        RAISE NOTICE 'reorg_events';
        RAISE NOTICE '========================================';
    END IF;
END $$;