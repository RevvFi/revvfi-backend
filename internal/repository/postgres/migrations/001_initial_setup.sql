
/*@
file: 000_initial_setup.sql
version: 1.0
date: June 5, 2026
environment: Development
description: Complete RevvFi database schema with all fixes applied
*/

/*@
fixes_applied:
- wallet_address changed from VARCHAR(42) to TEXT for multi-chain support
- GetNonce query filters by consumed=FALSE and expires_at > NOW()
- Verification block uses proper variable naming to avoid shadowing
- All tables have proper indexes and constraints
- EIP-4361 compliance for nonce table
*/

/*@
section: 1. AUTHENTICATION TABLES
*/

/*@
table: auth_nonces
description: Stores SIWE nonces for wallet authentication
design: Supports multiple nonces per wallet (no PRIMARY KEY on wallet)
security: consumed flag prevents replay attacks
eip4361: Minimum 8-character nonce, alphanumeric validation at application layer
*/
DROP TABLE IF EXISTS auth_nonces CASCADE;
CREATE TABLE auth_nonces (
    id BIGSERIAL PRIMARY KEY,
    wallet_address TEXT NOT NULL,
    nonce VARCHAR(64) NOT NULL,
    consumed BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    /*@
    constraint: unique_wallet_nonce
    description: Allows multiple nonces per wallet (different nonce values)
    */
    CONSTRAINT unique_wallet_nonce UNIQUE(wallet_address, nonce),
    
    /*@
    constraint: nonce_min_length
    description: EIP-4361 requires minimum 8-character nonce
    */
    CONSTRAINT nonce_min_length CHECK (char_length(nonce) >= 8)
);

/*@
indexes: auth_nonces
purpose: Performance optimization for common query patterns
*/
CREATE INDEX idx_auth_nonces_wallet ON auth_nonces(wallet_address);
CREATE INDEX idx_auth_nonces_wallet_nonce ON auth_nonces(wallet_address, nonce);
CREATE INDEX idx_auth_nonces_unconsumed ON auth_nonces(wallet_address, consumed) WHERE consumed = FALSE;
CREATE INDEX idx_auth_nonces_expires ON auth_nonces(expires_at) WHERE consumed = FALSE;

COMMENT ON TABLE auth_nonces IS 'Stores SIWE nonces for wallet authentication - supports multiple nonces per wallet';
COMMENT ON COLUMN auth_nonces.wallet_address IS 'Wallet address - TEXT type for multi-chain compatibility (Ethereum, Cosmos, Solana, etc.)';
COMMENT ON COLUMN auth_nonces.consumed IS 'Whether nonce has been used (prevents replay attacks)';
COMMENT ON CONSTRAINT unique_wallet_nonce ON auth_nonces IS 'Prevents duplicate nonce values for same wallet';
COMMENT ON CONSTRAINT nonce_min_length ON auth_nonces IS 'Enforces minimum 8-character nonce length per EIP-4361';

/*@
table: auth_sessions
description: Stores authenticated user sessions for JWT revocation
*/
DROP TABLE IF EXISTS auth_sessions CASCADE;
CREATE TABLE auth_sessions (
    id BIGSERIAL PRIMARY KEY,
    wallet_address TEXT NOT NULL,
    token TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    revoked_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_auth_sessions_wallet ON auth_sessions(wallet_address);
CREATE INDEX idx_auth_sessions_token ON auth_sessions(token);
CREATE INDEX idx_auth_sessions_expires ON auth_sessions(expires_at);
CREATE INDEX idx_auth_sessions_revoked ON auth_sessions(revoked_at) WHERE revoked_at IS NULL;

COMMENT ON TABLE auth_sessions IS 'Stores JWT sessions for token revocation and management';
COMMENT ON COLUMN auth_sessions.wallet_address IS 'Authenticated wallet address - TEXT type for multi-chain support';
COMMENT ON COLUMN auth_sessions.revoked_at IS 'Timestamp when session was revoked (NULL if active)';

/*@
section: 2. CORE PROTOCOL TABLES
*/

/*@
table: markets
description: Isolated lending pools created by borrowers
*/
DROP TABLE IF EXISTS markets CASCADE;
CREATE TABLE markets (
    id BIGSERIAL PRIMARY KEY,
    address TEXT NOT NULL UNIQUE,
    borrower TEXT NOT NULL,
    borrow_asset TEXT NOT NULL,
    borrow_asset_symbol TEXT,
    borrow_asset_decimals SMALLINT NOT NULL DEFAULT 18,
    collateral_asset TEXT NOT NULL,
    collateral_asset_symbol TEXT,
    collateral_asset_decimals SMALLINT NOT NULL DEFAULT 18,
    collateral_oracle TEXT NOT NULL,
    min_collateral_ratio INTEGER NOT NULL,
    liquidation_threshold INTEGER NOT NULL,
    total_principal TEXT NOT NULL DEFAULT '0',
    total_accrued_interest TEXT NOT NULL DEFAULT '0',
    total_debt TEXT NOT NULL DEFAULT '0',
    total_liquidity TEXT NOT NULL DEFAULT '0',
    borrow_index TEXT NOT NULL DEFAULT '1000000000000000000',
    weighted_avg_apr INTEGER NOT NULL DEFAULT 0,
    utilization_rate DOUBLE PRECISION NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_liquidating BOOLEAN NOT NULL DEFAULT FALSE,
    is_closed BOOLEAN NOT NULL DEFAULT FALSE,
    current_auction_id BIGINT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    closed_at TIMESTAMP WITH TIME ZONE,
    last_interest_accrual TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_health_check TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_markets_borrower ON markets(borrower);
CREATE INDEX idx_markets_borrow_asset ON markets(borrow_asset);
CREATE INDEX idx_markets_collateral_asset ON markets(collateral_asset);
CREATE INDEX idx_markets_is_active ON markets(is_active);
CREATE INDEX idx_markets_created_at ON markets(created_at);
CREATE INDEX idx_markets_is_liquidating ON markets(is_liquidating) WHERE is_liquidating = TRUE;

COMMENT ON TABLE markets IS 'Isolated lending pools created by borrowers';
COMMENT ON COLUMN markets.min_collateral_ratio IS 'Minimum collateral ratio in basis points (11000 = 110%)';
COMMENT ON COLUMN markets.liquidation_threshold IS 'Liquidation trigger in basis points (9500 = 95%)';
COMMENT ON COLUMN markets.utilization_rate IS 'Calculated as total_debt / total_liquidity';

/*@
table: offers
description: Lender liquidity commitments with APR and seniority tiers
*/
DROP TABLE IF EXISTS offers CASCADE;
CREATE TABLE offers (
    id BIGSERIAL PRIMARY KEY,
    offer_id BIGINT NOT NULL UNIQUE,
    lender TEXT NOT NULL,
    market_address TEXT NOT NULL,
    amount TEXT NOT NULL,
    remaining_amount TEXT NOT NULL,
    apr INTEGER NOT NULL,
    seniority SMALLINT NOT NULL,
    status TEXT NOT NULL,
    expiry TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    cancelled_at TIMESTAMP WITH TIME ZONE,
    expired_at TIMESTAMP WITH TIME ZONE,
    block_number BIGINT NOT NULL DEFAULT 0,
    tx_hash TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_offers_market_status ON offers(market_address, status);
CREATE INDEX idx_offers_lender ON offers(lender);
CREATE INDEX idx_offers_apr ON offers(apr) WHERE status = 'active';
CREATE INDEX idx_offers_expiry ON offers(expiry) WHERE status IN ('active', 'partially_filled');
CREATE INDEX idx_offers_seniority ON offers(seniority, status);

COMMENT ON TABLE offers IS 'Lender liquidity offers with APR and seniority tiers';
COMMENT ON COLUMN offers.apr IS 'Annual Percentage Rate in basis points (100 bps = 1%)';
COMMENT ON COLUMN offers.seniority IS '0 = Senior (first repayment), 1 = Junior (first loss)';

/*@
table: positions
description: ERC721 lender positions representing principal plus interest claims
*/
DROP TABLE IF EXISTS positions CASCADE;
CREATE TABLE positions (
    id BIGSERIAL PRIMARY KEY,
    token_id BIGINT NOT NULL UNIQUE,
    lender TEXT NOT NULL,
    market_address TEXT NOT NULL,
    principal TEXT NOT NULL,
    current_principal TEXT NOT NULL,
    accrued_interest TEXT NOT NULL DEFAULT '0',
    claimable_amount TEXT NOT NULL DEFAULT '0',
    apr INTEGER NOT NULL,
    seniority SMALLINT NOT NULL,
    status TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_settled BOOLEAN NOT NULL DEFAULT FALSE,
    minted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    settled_at TIMESTAMP WITH TIME ZONE,
    last_updated TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    block_number BIGINT NOT NULL DEFAULT 0,
    tx_hash TEXT NOT NULL DEFAULT '',
    log_index INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_positions_lender ON positions(lender);
CREATE INDEX idx_positions_market ON positions(market_address);
CREATE INDEX idx_positions_is_active ON positions(is_active);
CREATE INDEX idx_positions_status ON positions(status);
CREATE INDEX idx_positions_seniority ON positions(seniority, status);

COMMENT ON TABLE positions IS 'ERC721 positions representing lender claims on principal and interest';
COMMENT ON COLUMN positions.current_principal IS 'Principal after repayments';
COMMENT ON COLUMN positions.claimable_amount IS 'Amount ready to claim after settlement';

/*@
section: 3. BORROWER AND REPUTATION TABLES
*/

/*@
table: borrowers
description: Borrower profiles with reputation scoring (0-1000)
formula: reputation_score = successRate - (defaults * 50)
*/
DROP TABLE IF EXISTS borrowers CASCADE;
CREATE TABLE borrowers (
    address TEXT PRIMARY KEY,
    total_borrowed TEXT NOT NULL DEFAULT '0',
    total_repaid TEXT NOT NULL DEFAULT '0',
    outstanding_debt TEXT NOT NULL DEFAULT '0',
    total_loans INTEGER NOT NULL DEFAULT 0,
    successful_loans INTEGER NOT NULL DEFAULT 0,
    defaulted_loans INTEGER NOT NULL DEFAULT 0,
    active_loans INTEGER NOT NULL DEFAULT 0,
    reputation_score INTEGER NOT NULL DEFAULT 500,
    risk_label TEXT NOT NULL DEFAULT 'B',
    success_rate DOUBLE PRECISION NOT NULL DEFAULT 0,
    registered_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_activity TIMESTAMP WITH TIME ZONE,
    last_reputation_update TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_borrowers_reputation ON borrowers(reputation_score DESC);
CREATE INDEX idx_borrowers_active_loans ON borrowers(active_loans) WHERE active_loans > 0;
CREATE INDEX idx_borrowers_risk_label ON borrowers(risk_label);

COMMENT ON TABLE borrowers IS 'Borrower profiles with reputation scoring';
COMMENT ON COLUMN borrowers.reputation_score IS '0-1000 score where higher is better';
COMMENT ON COLUMN borrowers.risk_label IS 'AAA (900+), AA (800-899), A (700-799), B (500-699), C (300-499), D (0-299)';

/*@
section: 4. LIQUIDATION AND AUCTION TABLES
*/

/*@
table: auctions
description: Dutch auction liquidations (3-day duration, 5 percent hourly price drop)
*/
DROP TABLE IF EXISTS auctions CASCADE;
CREATE TABLE auctions (
    id BIGSERIAL PRIMARY KEY,
    auction_id BIGINT NOT NULL UNIQUE,
    market_address TEXT NOT NULL,
    borrower_address TEXT NOT NULL DEFAULT '',
    collateral_amount TEXT NOT NULL DEFAULT '0',
    collateral_value TEXT NOT NULL DEFAULT '0',
    debt_amount TEXT NOT NULL DEFAULT '0',
    highest_bid TEXT NOT NULL DEFAULT '0',
    highest_bidder TEXT,
    current_price TEXT NOT NULL DEFAULT '0',
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    settlement_time TIMESTAMP WITH TIME ZONE,
    status TEXT NOT NULL,
    winning_bid TEXT NOT NULL DEFAULT '0',
    winner TEXT,
    recovery_rate DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_auctions_market_status ON auctions(market_address, status);
CREATE INDEX idx_auctions_borrower ON auctions(borrower_address);
CREATE INDEX idx_auctions_end_time ON auctions(end_time) WHERE status = 'active';
CREATE INDEX idx_auctions_status ON auctions(status);

COMMENT ON TABLE auctions IS 'Dutch auction liquidations';
COMMENT ON COLUMN auctions.current_price IS 'Current Dutch auction price (decreases over time)';
COMMENT ON COLUMN auctions.recovery_rate IS 'Percentage of debt recovered from liquidation';

/*@
section: 5. WITHDRAWAL QUEUE TABLES (7-day epochs)
*/

/*@
table: withdrawal_requests
description: Lender withdrawal requests queued for epoch processing
*/
DROP TABLE IF EXISTS withdrawal_requests CASCADE;
CREATE TABLE withdrawal_requests (
    id BIGSERIAL PRIMARY KEY,
    request_id BIGSERIAL UNIQUE,
    lender TEXT NOT NULL,
    position_id BIGINT NOT NULL,
    market_address TEXT NOT NULL DEFAULT '',
    requested_amount TEXT NOT NULL,
    fulfilled_amount TEXT NOT NULL DEFAULT '0',
    remaining_amount TEXT NOT NULL DEFAULT '0',
    epoch_number INTEGER NOT NULL,
    status TEXT NOT NULL,
    fulfillment_time TIMESTAMP WITH TIME ZONE,
    fulfillment_tx_hash TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_withdrawals_lender ON withdrawal_requests(lender);
CREATE INDEX idx_withdrawals_position ON withdrawal_requests(position_id);
CREATE INDEX idx_withdrawals_epoch_status ON withdrawal_requests(epoch_number, status);
CREATE INDEX idx_withdrawals_status ON withdrawal_requests(status);

COMMENT ON TABLE withdrawal_requests IS 'Lender withdrawal requests queued for epoch processing';
COMMENT ON COLUMN withdrawal_requests.epoch_number IS '7-day epoch cycle number';

/*@
table: withdrawal_epochs
description: Withdrawal epoch tracking (7-day cycles)
*/
DROP TABLE IF EXISTS withdrawal_epochs CASCADE;
CREATE TABLE withdrawal_epochs (
    id BIGSERIAL PRIMARY KEY,
    epoch_number INTEGER NOT NULL UNIQUE,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE,
    total_requested TEXT NOT NULL DEFAULT '0',
    total_fulfilled TEXT NOT NULL DEFAULT '0',
    status TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_withdrawal_epochs_number ON withdrawal_epochs(epoch_number);
CREATE INDEX idx_withdrawal_epochs_status ON withdrawal_epochs(status);

COMMENT ON TABLE withdrawal_epochs IS 'Withdrawal epoch tracking for 7-day cycles';
COMMENT ON COLUMN withdrawal_epochs.status IS 'pending, processing, completed, failed';

/*@
section: 6. REPAYMENT HISTORY
*/

/*@
table: repayments
description: Loan repayment history with interest tracking
*/
DROP TABLE IF EXISTS repayments CASCADE;
CREATE TABLE repayments (
    id BIGSERIAL PRIMARY KEY,
    market_address TEXT NOT NULL,
    borrower_address TEXT NOT NULL,
    amount TEXT NOT NULL,
    interest_paid TEXT NOT NULL,
    principal_paid TEXT NOT NULL,
    repayment_type TEXT NOT NULL,
    block_number BIGINT NOT NULL,
    tx_hash TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_repayments_borrower ON repayments(borrower_address);
CREATE INDEX idx_repayments_market ON repayments(market_address);
CREATE INDEX idx_repayments_tx_hash ON repayments(tx_hash);
CREATE INDEX idx_repayments_created_at ON repayments(created_at);

COMMENT ON TABLE repayments IS 'Loan repayment history with interest tracking';
COMMENT ON COLUMN repayments.repayment_type IS 'partial, full, liquidation';

/*@
section: 7. INFRASTRUCTURE TABLES
*/

/*@
table: processed_events
description: Event idempotency tracking (prevents duplicate blockchain event processing)
*/
DROP TABLE IF EXISTS processed_events CASCADE;
CREATE TABLE processed_events (
    id BIGSERIAL PRIMARY KEY,
    tx_hash TEXT NOT NULL,
    log_index INTEGER NOT NULL,
    block_number BIGINT NOT NULL,
    event_name TEXT NOT NULL,
    contract_address TEXT NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(tx_hash, log_index)
);

CREATE INDEX idx_processed_events_block ON processed_events(block_number);
CREATE INDEX idx_processed_events_tx ON processed_events(tx_hash);
CREATE INDEX idx_processed_events_processed_at ON processed_events(processed_at);

COMMENT ON TABLE processed_events IS 'Tracks processed blockchain events for idempotency';
COMMENT ON COLUMN processed_events.log_index IS 'Transaction log index for unique event identification';

/*@
table: chain_checkpoints
description: Blockchain checkpoint tracking for reorg protection
*/
DROP TABLE IF EXISTS chain_checkpoints CASCADE;
CREATE TABLE chain_checkpoints (
    id BIGSERIAL PRIMARY KEY,
    block_number BIGINT NOT NULL UNIQUE,
    block_hash TEXT NOT NULL,
    parent_hash TEXT NOT NULL,
    is_finalized BOOLEAN NOT NULL DEFAULT FALSE,
    confirmation_count INTEGER NOT NULL DEFAULT 0,
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_chain_checkpoints_block ON chain_checkpoints(block_number DESC);
CREATE INDEX idx_chain_checkpoints_finalized ON chain_checkpoints(is_finalized);
CREATE INDEX idx_chain_checkpoints_processed ON chain_checkpoints(processed_at);

COMMENT ON TABLE chain_checkpoints IS 'Tracks blockchain checkpoints for reorg protection';

/*@
table: failed_events
description: Dead letter queue for failed event processing
*/
DROP TABLE IF EXISTS failed_events CASCADE;
CREATE TABLE failed_events (
    id BIGSERIAL PRIMARY KEY,
    payload JSONB NOT NULL,
    reason TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_failed_events_created ON failed_events(created_at);
CREATE INDEX idx_failed_events_reason ON failed_events(reason);

COMMENT ON TABLE failed_events IS 'Dead letter queue for events that failed processing';

/*@
section: 8. AUTO-CLEANUP FUNCTIONS
*/

/*@
function: cleanup_expired_nonces
purpose: Removes expired and consumed nonces periodically
schedule: Run every hour via cron
*/
CREATE OR REPLACE FUNCTION cleanup_expired_nonces()
RETURNS void AS $$
BEGIN
    DELETE FROM auth_nonces 
    WHERE expires_at < NOW() 
       OR (consumed = TRUE AND created_at < NOW() - INTERVAL '1 hour');
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION cleanup_expired_nonces() IS 'Removes expired and consumed nonces older than 1 hour';

/*@
function: refresh_expired_offers
purpose: Marks expired offers as expired
schedule: Run every minute via cron
*/
CREATE OR REPLACE FUNCTION refresh_expired_offers()
RETURNS void AS $$
BEGIN
    UPDATE offers 
    SET status = 'expired', 
        updated_at = NOW(),
        expired_at = NOW()
    WHERE status IN ('active', 'partially_filled')
      AND expiry < NOW();
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION refresh_expired_offers() IS 'Marks expired offers as expired for cleanup';

/*@
section: 9. NONCE RETRIEVAL FUNCTION (FIXED)
purpose: Retrieves valid nonce ignoring expired and consumed entries
*/

/*@
function: get_valid_nonce
purpose: Returns the most recent valid nonce for a wallet
fix: Filters by consumed = FALSE and expires_at > NOW()
*/
CREATE OR REPLACE FUNCTION get_valid_nonce(wallet_addr TEXT)
RETURNS TEXT AS $$
DECLARE
    valid_nonce TEXT;
BEGIN
    SELECT nonce INTO valid_nonce
    FROM auth_nonces
    WHERE wallet_address = wallet_addr
      AND consumed = FALSE
      AND expires_at > NOW()
    ORDER BY created_at DESC
    LIMIT 1;
    
    RETURN valid_nonce;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION get_valid_nonce(TEXT) IS 'Returns the most recent unconsumed, non-expired nonce for a wallet';

/*@
section: 10. VERIFICATION QUERIES (FIXED)
purpose: Validates all tables were created correctly
fix: Uses proper variable naming to avoid identifier shadowing
*/

DO $$
DECLARE
    current_table TEXT;
    expected_tables TEXT[] := ARRAY[
        'auth_nonces', 'auth_sessions', 'markets', 'offers', 'positions',
        'borrowers', 'auctions', 'withdrawal_requests', 'withdrawal_epochs',
        'repayments', 'processed_events', 'chain_checkpoints', 'failed_events'
    ];
    missing_tables TEXT[] := ARRAY[]::TEXT[];
    table_exists BOOLEAN;
BEGIN
    FOREACH current_table IN ARRAY expected_tables
    LOOP
        SELECT EXISTS (
            SELECT 1
            FROM information_schema.tables
            WHERE information_schema.tables.table_name = current_table
              AND information_schema.tables.table_schema = 'public'
        ) INTO table_exists;
        
        IF NOT table_exists THEN
            missing_tables := array_append(missing_tables, current_table);
        END IF;
    END LOOP;
    
    IF array_length(missing_tables, 1) > 0 THEN
        RAISE NOTICE 'WARNING: Missing tables: %', missing_tables;
    ELSE
        RAISE NOTICE 'SUCCESS: All 13 tables created successfully';
    END IF;
END $$;

/*@
verification: auth_nonces schema check
*/
DO $$
DECLARE
    has_consumed BOOLEAN;
    has_composite_constraint BOOLEAN;
    has_wallet_index BOOLEAN;
BEGIN
    SELECT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'auth_nonces' AND column_name = 'consumed'
    ) INTO has_consumed;
    
    SELECT EXISTS (
        SELECT 1 FROM information_schema.constraint_column_usage 
        WHERE constraint_name = 'unique_wallet_nonce'
    ) INTO has_composite_constraint;
    
    SELECT EXISTS (
        SELECT 1 FROM pg_indexes 
        WHERE tablename = 'auth_nonces' AND indexname = 'idx_auth_nonces_wallet'
    ) INTO has_wallet_index;
    
    IF has_consumed AND has_composite_constraint AND has_wallet_index THEN
        RAISE NOTICE 'SUCCESS: auth_nonces schema is correct';
    ELSE
        RAISE NOTICE 'WARNING: auth_nonces schema may have issues';
        RAISE NOTICE '  - has_consumed column: %', has_consumed;
        RAISE NOTICE '  - has unique_wallet_nonce constraint: %', has_composite_constraint;
        RAISE NOTICE '  - has idx_auth_nonces_wallet index: %', has_wallet_index;
    END IF;
END $$;

/*@
verification: get_valid_nonce function test
*/
DO $$
DECLARE
    func_exists BOOLEAN;
BEGIN
    SELECT EXISTS (
        SELECT 1 FROM pg_proc 
        WHERE proname = 'get_valid_nonce'
    ) INTO func_exists;
    
    IF func_exists THEN
        RAISE NOTICE 'SUCCESS: get_valid_nonce function exists';
    ELSE
        RAISE NOTICE 'WARNING: get_valid_nonce function not found';
    END IF;
END $$;

/*@
section: 11. MIGRATION COMPLETE
*/

DO $$
BEGIN
    RAISE NOTICE '';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'REVVFI DATABASE MIGRATION COMPLETE';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'Version: 1.0';
    RAISE NOTICE 'Date: June 5, 2026';
    RAISE NOTICE '========================================';
    RAISE NOTICE '';
    RAISE NOTICE 'Key Features:';
    RAISE NOTICE '  - Multi-chain wallet support (TEXT type)';
    RAISE NOTICE '  - Multiple nonces per wallet';
    RAISE NOTICE '  - Replay attack protection (consumed flag)';
    RAISE NOTICE '  - EIP-4361 compliance';
    RAISE NOTICE '  - 7-day withdrawal epochs';
    RAISE NOTICE '  - Dutch auction liquidations';
    RAISE NOTICE '  - Event idempotency tracking';
    RAISE NOTICE '  - Reorg protection (chain checkpoints)';
    RAISE NOTICE '========================================';
END $$;