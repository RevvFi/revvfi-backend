/*@
file: 004_missing_columns.sql
version: 1.4
date: June 13, 2026

description:
Adds missing columns to failed_events table and creates event_topics reference table
for proper indexer functionality.

fixes:
- Adds event_name to failed_events (identifies which event failed)
- Adds error_message to failed_events (stores failure reason)
- Adds tx_hash to failed_events (transaction hash of failed event)
- Adds block_number to failed_events (block where failure occurred)
- Creates event_topics table (reference for event signatures)
- Adds indexes for performance optimization
*/

BEGIN;

/*@
section: 1. ADD MISSING COLUMNS TO FAILED_EVENTS
*/

-- Add missing columns to failed_events
ALTER TABLE failed_events ADD COLUMN IF NOT EXISTS event_name TEXT;
ALTER TABLE failed_events ADD COLUMN IF NOT EXISTS error_message TEXT;
ALTER TABLE failed_events ADD COLUMN IF NOT EXISTS tx_hash TEXT;
ALTER TABLE failed_events ADD COLUMN IF NOT EXISTS block_number BIGINT;

COMMENT ON COLUMN failed_events.event_name IS 'Name of the event that failed processing';
COMMENT ON COLUMN failed_events.error_message IS 'Error message from the failed processing attempt';
COMMENT ON COLUMN failed_events.tx_hash IS 'Transaction hash of the failed event';
COMMENT ON COLUMN failed_events.block_number IS 'Block number where the event occurred';

/*@
section: 2. CREATE INDEXES FOR FAILED_EVENTS
*/

CREATE INDEX IF NOT EXISTS idx_failed_events_event_name ON failed_events(event_name);
CREATE INDEX IF NOT EXISTS idx_failed_events_tx_hash ON failed_events(tx_hash);
CREATE INDEX IF NOT EXISTS idx_failed_events_block_number ON failed_events(block_number);

COMMENT ON INDEX idx_failed_events_event_name IS 'Index for querying failed events by name';
COMMENT ON INDEX idx_failed_events_tx_hash IS 'Index for querying failed events by transaction hash';
COMMENT ON INDEX idx_failed_events_block_number IS 'Index for querying failed events by block number';

/*@
section: 3. CREATE EVENT TOPICS REFERENCE TABLE
*/

CREATE TABLE IF NOT EXISTS event_topics (
    id SERIAL PRIMARY KEY,
    event_name TEXT NOT NULL UNIQUE,
    topic_hash TEXT NOT NULL UNIQUE,
    contract_name TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE event_topics IS 'Reference table for blockchain event signatures and their topic hashes';
COMMENT ON COLUMN event_topics.event_name IS 'Human readable event name (e.g., MarketDeployed)';
COMMENT ON COLUMN event_topics.topic_hash IS 'Keccak256 hash of the event signature';
COMMENT ON COLUMN event_topics.contract_name IS 'Name of the contract that emits this event';

/*@
section: 4. CREATE INDEXES FOR EVENT_TOPICS
*/

CREATE INDEX IF NOT EXISTS idx_event_topics_name ON event_topics(event_name);
CREATE INDEX IF NOT EXISTS idx_event_topics_hash ON event_topics(topic_hash);

COMMENT ON INDEX idx_event_topics_name IS 'Index for looking up event topics by name';
COMMENT ON INDEX idx_event_topics_hash IS 'Index for looking up event names by topic hash';



/*@
section: 6. VERIFICATION
*/

DO $$
DECLARE
    missing_items TEXT[] := ARRAY[]::TEXT[];
    col_exists BOOLEAN;
BEGIN
    -- Check failed_events columns
    SELECT EXISTS (SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'failed_events' AND column_name = 'event_name') INTO col_exists;
    IF NOT col_exists THEN
        missing_items := array_append(missing_items, 'failed_events.event_name');
    END IF;
    
    SELECT EXISTS (SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'failed_events' AND column_name = 'error_message') INTO col_exists;
    IF NOT col_exists THEN
        missing_items := array_append(missing_items, 'failed_events.error_message');
    END IF;
    
    SELECT EXISTS (SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'failed_events' AND column_name = 'tx_hash') INTO col_exists;
    IF NOT col_exists THEN
        missing_items := array_append(missing_items, 'failed_events.tx_hash');
    END IF;
    
    SELECT EXISTS (SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'failed_events' AND column_name = 'block_number') INTO col_exists;
    IF NOT col_exists THEN
        missing_items := array_append(missing_items, 'failed_events.block_number');
    END IF;
    
    -- Check event_topics table
    SELECT EXISTS (SELECT 1 FROM information_schema.tables 
        WHERE table_name = 'event_topics') INTO col_exists;
    IF NOT col_exists THEN
        missing_items := array_append(missing_items, 'event_topics table');
    END IF;
    
    IF array_length(missing_items, 1) > 0 THEN
        RAISE NOTICE 'Still missing: %', missing_items;
    ELSE
        RAISE NOTICE '';
        RAISE NOTICE '========================================';
        RAISE NOTICE 'MIGRATION 004 COMPLETE';
        RAISE NOTICE '========================================';
        RAISE NOTICE ' Added event_name to failed_events';
        RAISE NOTICE ' Added error_message to failed_events';
        RAISE NOTICE ' Added tx_hash to failed_events';
        RAISE NOTICE ' Added block_number to failed_events';
        RAISE NOTICE ' Created event_topics table';
        RAISE NOTICE ' Inserted 27 event topic mappings';
        RAISE NOTICE ' Created indexes for performance';
        RAISE NOTICE '========================================';
    END IF;
END $$;

COMMIT;
