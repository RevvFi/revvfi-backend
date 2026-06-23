-- Migration: Add auctions and bids tables for liquidation system
-- Date: 2026-06-22

-- Auctions table
CREATE TABLE IF NOT EXISTS auctions (
    id SERIAL PRIMARY KEY,
    auction_id INTEGER NOT NULL UNIQUE,
    market_address VARCHAR(42) NOT NULL,
    borrower VARCHAR(42) NOT NULL,
    borrow_asset VARCHAR(42) NOT NULL,
    collateral_asset VARCHAR(42) NOT NULL,
    collateral_amount NUMERIC(78) NOT NULL,
    debt_amount NUMERIC(78) NOT NULL,
    reserve_price NUMERIC(78) NOT NULL,
    start_time BIGINT NOT NULL,
    end_time BIGINT NOT NULL,
    highest_bid NUMERIC(78) DEFAULT 0,
    highest_bidder VARCHAR(42),
    active BOOLEAN DEFAULT true,
    settled BOOLEAN DEFAULT false,
    collateral_transferred BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Bids table
CREATE TABLE IF NOT EXISTS bids (
    id SERIAL PRIMARY KEY,
    auction_id INTEGER NOT NULL REFERENCES auctions(auction_id) ON DELETE CASCADE,
    bidder VARCHAR(42) NOT NULL,
    bid_amount NUMERIC(78) NOT NULL,
    placed_at BIGINT NOT NULL,
    tx_hash VARCHAR(66),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_auctions_market ON auctions(market_address);
CREATE INDEX IF NOT EXISTS idx_auctions_borrower ON auctions(borrower);
CREATE INDEX IF NOT EXISTS idx_auctions_active ON auctions(active) WHERE active = true;
CREATE INDEX IF NOT EXISTS idx_auctions_auction_id ON auctions(auction_id);
CREATE INDEX IF NOT EXISTS idx_bids_auction_id ON bids(auction_id);
CREATE INDEX IF NOT EXISTS idx_bids_bidder ON bids(bidder);

-- Comments
COMMENT ON TABLE auctions IS 'Liquidation auctions for undercollateralized positions';
COMMENT ON TABLE bids IS 'Bid history for liquidation auctions';
COMMENT ON COLUMN auctions.auction_id IS 'On-chain auction ID from RevvFiLiquidator contract';
COMMENT ON COLUMN auctions.reserve_price IS 'Minimum acceptable bid (80% of debt amount)';
COMMENT ON COLUMN auctions.highest_bid IS 'Current highest bid amount';
COMMENT ON COLUMN bids.placed_at IS 'Block timestamp when bid was placed';
