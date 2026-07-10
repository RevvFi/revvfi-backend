-- Migration: Fill the gap left by 002_indexer_setup.sql
-- Date: 2026-07-08
--
-- 001_initial_setup.sql already creates processed_events, chain_checkpoints,
-- and failed_events (with a minimal failed_events schema: id, payload,
-- reason, created_at) with hardcoded index names. 002_indexer_setup.sql
-- later tried to (re)create the same three tables via `CREATE TABLE IF NOT
-- EXISTS` (a no-op against the pre-existing tables) and then ran bare
-- `CREATE INDEX` statements reusing several of 001's exact index names
-- (idx_processed_events_tx, idx_processed_events_block,
-- idx_processed_events_processed_at, idx_chain_checkpoints_block/finalized/
-- processed, idx_failed_events_created) with no `IF NOT EXISTS` guard.
--
-- On a fresh database, 002 fails on the first such collision. Because the
-- whole pasted script runs as one implicit transaction, everything in the
-- file after that point never committed: indexer_sync_state, reorg_events,
-- and failed_events' retry_count/max_retries/next_retry_at/processed/
-- resolved_at columns and their indexes were silently never created.
--
-- This migration is purely additive/idempotent and fills exactly that gap,
-- regardless of how far 002 actually got.
--
-- It also adds `failure_reason`, a column 002 defined that is distinct
-- from 001's `reason` column: internal/repository/postgres/event.go uses
-- BOTH columns from different functions (SaveFailedEvent/
-- SaveFailedEventWithData write `reason`; GetUnresolvedFailedEvents/
-- MarkFailedEventResolved/IncrementFailedEventRetry/StoreFailedEvent read
-- and write `failure_reason`), so both must exist.

BEGIN;

/*@
section: 1. FAILED EVENTS - MISSING RETRY/QUEUE COLUMNS
*/
ALTER TABLE failed_events ADD COLUMN IF NOT EXISTS failure_reason TEXT;
ALTER TABLE failed_events ADD COLUMN IF NOT EXISTS retry_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE failed_events ADD COLUMN IF NOT EXISTS max_retries INTEGER NOT NULL DEFAULT 5;
ALTER TABLE failed_events ADD COLUMN IF NOT EXISTS next_retry_at TIMESTAMPTZ;
ALTER TABLE failed_events ADD COLUMN IF NOT EXISTS processed BOOLEAN DEFAULT FALSE;
ALTER TABLE failed_events ADD COLUMN IF NOT EXISTS resolved_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_failed_events_retry ON failed_events(next_retry_at) WHERE processed = FALSE;
CREATE INDEX IF NOT EXISTS idx_failed_events_name ON failed_events(event_name);

/*@
section: 2. PROCESSED EVENTS - MISSING INDEX
*/
CREATE INDEX IF NOT EXISTS idx_processed_events_name ON processed_events(event_name);

/*@
section: 3. INDEXER SYNC STATE (never created if 002 rolled back)
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
section: 4. REORG AUDIT TABLE (never created if 002 rolled back)
*/
CREATE TABLE IF NOT EXISTS reorg_events (
    id BIGSERIAL PRIMARY KEY,
    old_block_number BIGINT NOT NULL,
    old_block_hash TEXT NOT NULL,
    new_block_hash TEXT NOT NULL,
    parent_hash TEXT,
    depth INTEGER DEFAULT 1,
    detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_reorg_events_block ON reorg_events(old_block_number);
CREATE INDEX IF NOT EXISTS idx_reorg_events_detected_at ON reorg_events(detected_at);

COMMENT ON TABLE reorg_events IS 'Audit log of detected blockchain reorganizations';

/*@
section: 5. VERIFICATION
*/
DO $$
DECLARE
    missing TEXT[] := ARRAY[]::TEXT[];
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name = 'failed_events' AND column_name = 'retry_count') THEN
        missing := array_append(missing, 'failed_events.retry_count');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name = 'failed_events' AND column_name = 'failure_reason') THEN
        missing := array_append(missing, 'failed_events.failure_reason');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables
                   WHERE table_schema = 'public' AND table_name = 'indexer_sync_state') THEN
        missing := array_append(missing, 'indexer_sync_state');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables
                   WHERE table_schema = 'public' AND table_name = 'reorg_events') THEN
        missing := array_append(missing, 'reorg_events');
    END IF;

    IF array_length(missing, 1) > 0 THEN
        RAISE NOTICE 'Still missing: %', missing;
    ELSE
        RAISE NOTICE 'MIGRATION 011 COMPLETE — indexer gap filled';
    END IF;
END $$;

COMMIT;
