/*@
file: 006_admin_features.sql
version: 1.6
date: June 17, 2026
environment: Development
description: Admin feature tables for RevvFi - audit logs, liquidator config, system config
*/

BEGIN;

/*@
section: 1. ADMIN AUDIT LOGS
table: admin_audit_logs
description: Tracks all admin actions for compliance and audit trail
design: Append-only log; never update rows
*/
CREATE TABLE IF NOT EXISTS admin_audit_logs (
    id BIGSERIAL PRIMARY KEY,
    admin_address TEXT NOT NULL,
    action TEXT NOT NULL,
    target_type TEXT NOT NULL,
    target_address TEXT,
    params JSONB,
    tx_hash TEXT,
    result TEXT NOT NULL DEFAULT 'success',
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_admin ON admin_audit_logs(admin_address);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON admin_audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_target ON admin_audit_logs(target_address);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created ON admin_audit_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_target_type ON admin_audit_logs(target_type);

COMMENT ON TABLE admin_audit_logs IS 'Tracks all admin actions for compliance audit trail';
COMMENT ON COLUMN admin_audit_logs.action IS 'Action: update_min_cr, update_oracle, stop_auction, etc.';
COMMENT ON COLUMN admin_audit_logs.target_type IS 'Type: market, borrower, protocol, system, liquidator';

/*@
section: 2. LIQUIDATOR CONFIG
table: liquidator_config
description: Dutch auction parameters (3-day duration, 5% hourly drop)
design: Single-row config; updates in place
*/
CREATE TABLE IF NOT EXISTS liquidator_config (
    id SERIAL PRIMARY KEY,
    auction_duration_seconds INTEGER NOT NULL DEFAULT 259200,
    price_drop_rate_bps INTEGER NOT NULL DEFAULT 500,
    min_bid_increment_bps INTEGER NOT NULL DEFAULT 100,
    liquidation_incentive_bps INTEGER NOT NULL DEFAULT 200,
    max_slippage_bps INTEGER NOT NULL DEFAULT 1000,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by TEXT
);

COMMENT ON TABLE liquidator_config IS 'Dutch auction parameters for the liquidator contract';
COMMENT ON COLUMN liquidator_config.auction_duration_seconds IS '3-day default: 259200 seconds';
COMMENT ON COLUMN liquidator_config.price_drop_rate_bps IS 'Hourly price drop in basis points (500 = 5%)';

INSERT INTO liquidator_config (auction_duration_seconds, price_drop_rate_bps, min_bid_increment_bps, liquidation_incentive_bps, max_slippage_bps)
SELECT 259200, 500, 100, 200, 1000
WHERE NOT EXISTS (SELECT 1 FROM liquidator_config LIMIT 1);

/*@
section: 3. SYSTEM CONFIG
table: system_config
description: Protocol-wide key-value configuration store
design: Key-unique rows; upsert on update
*/
CREATE TABLE IF NOT EXISTS system_config (
    id SERIAL PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    value TEXT NOT NULL,
    description TEXT,
    config_type TEXT NOT NULL DEFAULT 'string',
    is_sensitive BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by TEXT
);

CREATE INDEX IF NOT EXISTS idx_system_config_key ON system_config(key);

COMMENT ON TABLE system_config IS 'Protocol-wide system configuration key-value store';
COMMENT ON COLUMN system_config.config_type IS 'Data type: string, number, boolean, address';
COMMENT ON COLUMN system_config.is_sensitive IS 'Mask value in API responses when true';

INSERT INTO system_config (key, value, description, config_type) VALUES
    ('max_markets_per_borrower', '10', 'Maximum markets a single borrower can create', 'number'),
    ('max_offer_duration_days', '365', 'Maximum duration for a lending offer in days', 'number'),
    ('min_loan_amount_wei', '1000000000000000000', 'Minimum loan amount in wei (1 ETH)', 'number'),
    ('epoch_duration_days', '7', 'Withdrawal epoch duration in days', 'number'),
    ('protocol_paused', 'false', 'Whether the entire protocol is paused', 'boolean'),
    ('indexer_enabled', 'true', 'Whether the blockchain indexer is running', 'boolean'),
    ('oracle_staleness_threshold_seconds', '3600', 'Maximum age of oracle price data in seconds', 'number')
ON CONFLICT (key) DO NOTHING;

COMMIT;
