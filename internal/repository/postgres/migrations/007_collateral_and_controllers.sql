/*@
file: 006_collateral_and_controllers.sql
version: 1.6
date: June 17, 2026

description:
Adds new tables and missing columns required for complete indexer coverage.

changes:
- Creates collateral_balances to track per-borrower collateral from CollateralEscrow events
- Creates controller_factories to track ControllerFactoryAdded/Removed events
- Creates controllers to track ControllerAdded/Removed events
- Adds is_whitelisted to borrowers (required by UpdateBorrowerWhitelistStatus)
- Adds is_redeemed and redeemed_at to positions (required by PositionRedeemed handler)
*/

BEGIN;

/*@
section: 1. COLLATERAL BALANCES
purpose: Tracks per-borrower collateral balance derived from CollateralEscrow events.
         Uses the Total field from each event as the authoritative current balance
         and accumulates deposit/withdrawal deltas for historical analysis.
*/

CREATE TABLE IF NOT EXISTS collateral_balances (
    id              BIGSERIAL PRIMARY KEY,
    borrower        TEXT NOT NULL UNIQUE,
    balance         TEXT NOT NULL DEFAULT '0',
    total_deposited TEXT NOT NULL DEFAULT '0',
    total_withdrawn TEXT NOT NULL DEFAULT '0',
    last_deposit_at    TIMESTAMPTZ,
    last_withdrawal_at TIMESTAMPTZ,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_collateral_balances_borrower
    ON collateral_balances(borrower);

COMMENT ON TABLE collateral_balances IS
    'Per-borrower collateral balance derived from RevvFiCollateralEscrow events';
COMMENT ON COLUMN collateral_balances.balance IS
    'Current total collateral (set from Total field in CollateralDeposited/CollateralWithdrawn events)';
COMMENT ON COLUMN collateral_balances.total_deposited IS
    'Cumulative sum of all deposit Amount fields';
COMMENT ON COLUMN collateral_balances.total_withdrawn IS
    'Cumulative sum of all withdrawal Amount fields';

/*@
section: 2. CONTROLLER FACTORIES
purpose: Tracks ControllerFactory registrations in the ArchController.
         ControllerFactoryAdded inserts a row; ControllerFactoryRemoved sets removed_at.
*/

CREATE TABLE IF NOT EXISTS controller_factories (
    id           BIGSERIAL PRIMARY KEY,
    address      TEXT NOT NULL UNIQUE,
    added_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    removed_at   TIMESTAMPTZ,
    block_number BIGINT NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_controller_factories_address
    ON controller_factories(address);
CREATE INDEX IF NOT EXISTS idx_controller_factories_active
    ON controller_factories(address) WHERE removed_at IS NULL;

COMMENT ON TABLE controller_factories IS
    'Controller factories registered in RevvFiArchController';

/*@
section: 3. CONTROLLERS
purpose: Tracks Controller registrations in the ArchController.
         ControllerAdded inserts a row; ControllerRemoved sets removed_at.
*/

CREATE TABLE IF NOT EXISTS controllers (
    id           BIGSERIAL PRIMARY KEY,
    address      TEXT NOT NULL UNIQUE,
    added_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    removed_at   TIMESTAMPTZ,
    block_number BIGINT NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_controllers_address
    ON controllers(address);
CREATE INDEX IF NOT EXISTS idx_controllers_active
    ON controllers(address) WHERE removed_at IS NULL;

COMMENT ON TABLE controllers IS
    'Controllers registered in RevvFiArchController';

/*@
section: 4. BORROWERS — ADD MISSING COLUMNS
purpose: Adds is_whitelisted which is referenced by UpdateBorrowerWhitelistStatus
         but was absent from the original schema.
*/

ALTER TABLE borrowers
    ADD COLUMN IF NOT EXISTS is_whitelisted BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_borrowers_whitelisted
    ON borrowers(is_whitelisted) WHERE is_whitelisted = TRUE;

COMMENT ON COLUMN borrowers.is_whitelisted IS
    'Whether the borrower has been whitelisted in the ArchController';

/*@
section: 5. POSITIONS — ADD REDEMPTION TRACKING
purpose: Allows the PositionRedeemed handler to mark NFTs that have been
         fully redeemed (principal + interest collected by lender).
*/

ALTER TABLE positions
    ADD COLUMN IF NOT EXISTS is_redeemed BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE positions
    ADD COLUMN IF NOT EXISTS redeemed_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_positions_redeemed
    ON positions(is_redeemed) WHERE is_redeemed = TRUE;

COMMENT ON COLUMN positions.is_redeemed IS
    'Whether the position NFT has been redeemed by the lender';
COMMENT ON COLUMN positions.redeemed_at IS
    'Timestamp of the PositionRedeemed event';

/*@
section: VERIFICATION
*/

DO $$
DECLARE
    missing TEXT[] := ARRAY[]::TEXT[];
    ok      BOOLEAN;
BEGIN
    SELECT EXISTS (SELECT 1 FROM information_schema.tables
        WHERE table_name = 'collateral_balances') INTO ok;
    IF NOT ok THEN missing := array_append(missing, 'collateral_balances'); END IF;

    SELECT EXISTS (SELECT 1 FROM information_schema.tables
        WHERE table_name = 'controller_factories') INTO ok;
    IF NOT ok THEN missing := array_append(missing, 'controller_factories'); END IF;

    SELECT EXISTS (SELECT 1 FROM information_schema.tables
        WHERE table_name = 'controllers') INTO ok;
    IF NOT ok THEN missing := array_append(missing, 'controllers'); END IF;

    SELECT EXISTS (SELECT 1 FROM information_schema.columns
        WHERE table_name = 'borrowers' AND column_name = 'is_whitelisted') INTO ok;
    IF NOT ok THEN missing := array_append(missing, 'borrowers.is_whitelisted'); END IF;

    SELECT EXISTS (SELECT 1 FROM information_schema.columns
        WHERE table_name = 'positions' AND column_name = 'is_redeemed') INTO ok;
    IF NOT ok THEN missing := array_append(missing, 'positions.is_redeemed'); END IF;

    IF array_length(missing, 1) > 0 THEN
        RAISE NOTICE 'MIGRATION 006 — still missing: %', missing;
    ELSE
        RAISE NOTICE '========================================';
        RAISE NOTICE 'MIGRATION 006 COMPLETE';
        RAISE NOTICE '========================================';
        RAISE NOTICE ' Created collateral_balances';
        RAISE NOTICE ' Created controller_factories';
        RAISE NOTICE ' Created controllers';
        RAISE NOTICE ' Added borrowers.is_whitelisted';
        RAISE NOTICE ' Added positions.is_redeemed, redeemed_at';
        RAISE NOTICE '========================================';
    END IF;
END $$;

COMMIT;
