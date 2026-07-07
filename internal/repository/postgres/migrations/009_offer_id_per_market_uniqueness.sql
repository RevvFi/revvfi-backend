-- Migration: Fix offer_id uniqueness scope
-- Date: 2026-07-03
--
-- Bug: RevvFiOfferBook assigns offer IDs per-contract (each market's OfferBook
-- starts its own counter at 1). The `offers` table treated offer_id as
-- globally unique across ALL markets, so as soon as a second market existed,
-- its offer #1 collided with the first market's offer #1 on insert/update
-- (via `ON CONFLICT (offer_id) DO UPDATE`), silently corrupting the older
-- market's offer status and remaining_amount with the newer market's data.
--
-- Fix: offer_id is only meaningful together with market_address. Replace the
-- single-column unique constraint with a composite one.

BEGIN;

ALTER TABLE offers DROP CONSTRAINT IF EXISTS offers_offer_id_key;
ALTER TABLE offers ADD CONSTRAINT offers_offer_id_market_key UNIQUE (offer_id, market_address);

-- The old single-column index is no longer needed as a uniqueness guard, but
-- lookups by offer_id alone are still common (list views), so keep a plain
-- non-unique index for those.
CREATE INDEX IF NOT EXISTS idx_offers_offer_id ON offers(offer_id);

COMMIT;

DO $$
BEGIN
    RAISE NOTICE '';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'MIGRATION 009 COMPLETE';
    RAISE NOTICE '========================================';
    RAISE NOTICE '  offer_id uniqueness is now scoped per (offer_id, market_address)';
    RAISE NOTICE '========================================';
END $$;
