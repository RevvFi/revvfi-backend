/*@
file: 006_drop_payload_not_null.sql
version: 1.5
date: June 13, 2026

description:
Drops NOT NULL constraint from payload column in failed_events table
to allow storing failed events without requiring payload data.
*/

BEGIN;

ALTER TABLE failed_events ALTER COLUMN payload DROP NOT NULL;

COMMIT;

-- Verification
DO $$
BEGIN
    RAISE NOTICE '';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'MIGRATION 006 COMPLETE';
    RAISE NOTICE '========================================';
    RAISE NOTICE ' Dropped NOT NULL constraint from payload column';
    RAISE NOTICE '========================================';
END $$;