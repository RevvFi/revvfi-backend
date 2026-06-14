/*@
file: 005_failed_event_update.sql
version: 1.5
date: June 13, 2026

description:
Adds event_data and raw_log columns to failed_events table for storing
complete event debugging information when processing fails.

fixes:
- Adds event_data JSONB column (stores decoded event data)
- Adds raw_log JSONB column (stores raw blockchain log)
- Allows StoreFailedEvent method to work properly
*/

BEGIN;

-- Add missing columns to failed_events
ALTER TABLE failed_events ADD COLUMN IF NOT EXISTS event_data JSONB;
ALTER TABLE failed_events ADD COLUMN IF NOT EXISTS raw_log JSONB;

-- Add comments for documentation
COMMENT ON COLUMN failed_events.event_data IS 'Decoded event data in JSON format for debugging';
COMMENT ON COLUMN failed_events.raw_log IS 'Raw blockchain log data for reprocessing';

-- Create index on event_data for JSON queries (optional)
CREATE INDEX IF NOT EXISTS idx_failed_events_event_data ON failed_events USING GIN (event_data);

COMMIT;

-- Verification
DO $$
BEGIN
    RAISE NOTICE '';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'MIGRATION 005 COMPLETE';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'Added event_data column to failed_events';
    RAISE NOTICE 'Added raw_log column to failed_events';
    RAISE NOTICE 'Created GIN index on event_data';
    RAISE NOTICE '========================================';
END $$;