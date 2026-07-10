/*@
file: 003_indexer_missing_columns.sql
version: 1.3
date: June 13, 2026

description:
Adds missing columns to existing tables for proper indexer functionality.

fixes:
- Adds event_name to failed_events (was missing)
- Adds error_message to failed_events  
- Adds created_at/updated_at to markets and borrowers
- Adds missing indexes for performance
- Fixes topics mapping for event decoding
*/

BEGIN;

/*@
section: 1. FIX FAILED_EVENTS TABLE
*/

-- Add missing columns to failed_events
ALTER TABLE failed_events ADD COLUMN IF NOT EXISTS event_name TEXT;
ALTER TABLE failed_events ADD COLUMN IF NOT EXISTS error_message TEXT;
ALTER TABLE failed_events ADD COLUMN IF NOT EXISTS tx_hash TEXT;
ALTER TABLE failed_events ADD COLUMN IF NOT EXISTS block_number BIGINT;

-- Update existing rows to set event_name from event_data if possible
UPDATE failed_events 
SET event_name = event_data->>'event_name' 
WHERE event_name IS NULL AND event_data IS NOT NULL;

-- Create indexes for failed_events queries
CREATE INDEX IF NOT EXISTS idx_failed_events_event_name ON failed_events(event_name);
CREATE INDEX IF NOT EXISTS idx_failed_events_tx_hash ON failed_events(tx_hash);
CREATE INDEX IF NOT EXISTS idx_failed_events_block_number ON failed_events(block_number);

COMMENT ON COLUMN failed_events.event_name IS 'Name of the event that failed processing';
COMMENT ON COLUMN failed_events.error_message IS 'Error message from the failed processing attempt';
COMMENT ON COLUMN failed_events.tx_hash IS 'Transaction hash of the failed event';
COMMENT ON COLUMN failed_events.block_number IS 'Block number where the event occurred';

/*@
section: 2. FIX MARKETS TABLE - ADD TIMESTAMP COLUMNS
*/

-- Add timestamp columns if they don't exist
ALTER TABLE markets ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE markets ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE markets ADD COLUMN IF NOT EXISTS closed_at TIMESTAMPTZ;

-- Create indexes for time-based queries
CREATE INDEX IF NOT EXISTS idx_markets_created_at ON markets(created_at);
CREATE INDEX IF NOT EXISTS idx_markets_updated_at ON markets(updated_at);
CREATE INDEX IF NOT EXISTS idx_markets_closed_at ON markets(closed_at) WHERE closed_at IS NOT NULL;

-- Update existing markets with default timestamps
UPDATE markets SET created_at = NOW() WHERE created_at IS NULL;
UPDATE markets SET updated_at = NOW() WHERE updated_at IS NULL;

/*@
section: 3. FIX BORROWERS TABLE - ADD TIMESTAMP COLUMNS
*/

-- Add timestamp columns if they don't exist
ALTER TABLE borrowers ADD COLUMN IF NOT EXISTS registered_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE borrowers ADD COLUMN IF NOT EXISTS last_activity TIMESTAMPTZ;
ALTER TABLE borrowers ADD COLUMN IF NOT EXISTS last_reputation_update TIMESTAMPTZ DEFAULT NOW();

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_borrowers_registered_at ON borrowers(registered_at);
CREATE INDEX IF NOT EXISTS idx_borrowers_last_activity ON borrowers(last_activity);
CREATE INDEX IF NOT EXISTS idx_borrowers_reputation_update ON borrowers(last_reputation_update);

-- Update existing borrowers
UPDATE borrowers SET registered_at = NOW() WHERE registered_at IS NULL;
UPDATE borrowers SET last_reputation_update = NOW() WHERE last_reputation_update IS NULL;

/*@
section: 4. FIX POSITIONS TABLE - ADD TIMESTAMP COLUMNS
*/

-- Add timestamp columns if they don't exist
ALTER TABLE positions ADD COLUMN IF NOT EXISTS minted_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE positions ADD COLUMN IF NOT EXISTS settled_at TIMESTAMPTZ;
ALTER TABLE positions ADD COLUMN IF NOT EXISTS last_updated TIMESTAMPTZ DEFAULT NOW();

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_positions_minted_at ON positions(minted_at);
CREATE INDEX IF NOT EXISTS idx_positions_settled_at ON positions(settled_at) WHERE settled_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_positions_last_updated ON positions(last_updated);

-- Update existing positions
UPDATE positions SET minted_at = NOW() WHERE minted_at IS NULL;
UPDATE positions SET last_updated = NOW() WHERE last_updated IS NULL;

/*@
section: 5. FIX OFFERS TABLE - ADD TIMESTAMP COLUMNS
*/

-- Add timestamp columns if they don't exist
ALTER TABLE offers ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE offers ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE offers ADD COLUMN IF NOT EXISTS expired_at TIMESTAMPTZ;
ALTER TABLE offers ADD COLUMN IF NOT EXISTS cancelled_at TIMESTAMPTZ;

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_offers_created_at ON offers(created_at);
CREATE INDEX IF NOT EXISTS idx_offers_expired_at ON offers(expired_at) WHERE expired_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_offers_cancelled_at ON offers(cancelled_at) WHERE cancelled_at IS NOT NULL;

-- Update existing offers
UPDATE offers SET created_at = NOW() WHERE created_at IS NULL;
UPDATE offers SET updated_at = NOW() WHERE updated_at IS NULL;

/*@
section: 6. ADD EVENT TOPICS MAPPING TABLE (for reference)
*/

CREATE TABLE IF NOT EXISTS event_topics (
    id SERIAL PRIMARY KEY,
    event_name TEXT NOT NULL UNIQUE,
    topic_hash TEXT NOT NULL UNIQUE,
    contract_name TEXT,
    abi_signature TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_event_topics_name ON event_topics(event_name);
CREATE INDEX IF NOT EXISTS idx_event_topics_hash ON event_topics(topic_hash);

COMMENT ON TABLE event_topics IS 'Reference table for event signature topics';

/*@
section: 7. CREATE VIEW FOR ACTIVE MARKETS WITH METRICS
*/

CREATE OR REPLACE VIEW active_markets_view AS
SELECT 
    m.address,
    m.borrower,
    m.borrow_asset,
    m.borrow_asset_symbol,
    m.collateral_asset,
    m.collateral_asset_symbol,
    m.min_collateral_ratio,
    m.liquidation_threshold,
    m.total_debt,
    m.total_liquidity,
    m.utilization_rate,
    m.weighted_avg_apr,
    m.is_active,
    m.is_liquidating,
    COUNT(p.id) as active_positions_count,
    COUNT(DISTINCT p.lender) as unique_lenders_count,
    COALESCE(SUM(p.principal::NUMERIC), 0) as total_principal_outstanding,
    COALESCE(SUM(p.accrued_interest::NUMERIC), 0) as total_accrued_interest
FROM markets m
LEFT JOIN positions p ON p.market_address = m.address AND p.is_active = TRUE
WHERE m.is_active = TRUE AND m.is_closed = FALSE
GROUP BY m.id;

COMMENT ON VIEW active_markets_view IS 'Aggregated view of active markets with metrics';

/*@
section: 8. CREATE FUNCTION TO RETRY FAILED EVENTS
*/

CREATE OR REPLACE FUNCTION retry_failed_events(max_retries_per_event INTEGER DEFAULT 5)
RETURNS TABLE(
    retried_count BIGINT,
    failed_count BIGINT
) AS $$
DECLARE
    retried BIGINT := 0;
    failed BIGINT := 0;
BEGIN
    -- Mark events for retry that haven't exceeded max retries
    UPDATE failed_events
    SET 
        retry_count = retry_count + 1,
        next_retry_at = NOW() + INTERVAL '5 minutes',
        processed = FALSE,
        resolved_at = NULL
    WHERE 
        processed = FALSE 
        AND retry_count < max_retries_per_event
        AND (next_retry_at IS NULL OR next_retry_at <= NOW());
    
    GET DIAGNOSTICS retried = ROW_COUNT;
    
    -- Count permanently failed events
    SELECT COUNT(*) INTO failed
    FROM failed_events
    WHERE processed = FALSE AND retry_count >= max_retries_per_event;
    
    RETURN QUERY SELECT retried, failed;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION retry_failed_events(INTEGER) IS 'Retries failed events that haven''t exceeded max retry count';

/*@
section: 9. VERIFICATION
*/

DO $$
DECLARE
    missing_columns TEXT[] := ARRAY[]::TEXT[];
    col_exists BOOLEAN;
BEGIN
    -- Check failed_events columns
    SELECT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'failed_events' AND column_name = 'event_name'
    ) INTO col_exists;
    IF NOT col_exists THEN
        missing_columns := array_append(missing_columns, 'failed_events.event_name');
    END IF;
    
    SELECT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'failed_events' AND column_name = 'error_message'
    ) INTO col_exists;
    IF NOT col_exists THEN
        missing_columns := array_append(missing_columns, 'failed_events.error_message');
    END IF;
    
    -- Check markets columns
    SELECT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'markets' AND column_name = 'created_at'
    ) INTO col_exists;
    IF NOT col_exists THEN
        missing_columns := array_append(missing_columns, 'markets.created_at');
    END IF;
    
    IF array_length(missing_columns, 1) > 0 THEN
        RAISE NOTICE 'Still missing columns: %', missing_columns;
    ELSE
        RAISE NOTICE '';
        RAISE NOTICE '========================================';
        RAISE NOTICE 'MIGRATION 003 COMPLETE';
        RAISE NOTICE '========================================';
        RAISE NOTICE 'Added event_name to failed_events';
        RAISE NOTICE 'Added error_message to failed_events';
        RAISE NOTICE 'Added timestamps to markets';
        RAISE NOTICE 'Added timestamps to borrowers';
        RAISE NOTICE 'Added timestamps to positions';
        RAISE NOTICE 'Added timestamps to offers';
        RAISE NOTICE 'Created event_topics reference table';
        RAISE NOTICE 'Created active_markets_view';
        RAISE NOTICE 'Created retry_failed_events function';
        RAISE NOTICE '========================================';
    END IF;
END $$;

COMMIT;
