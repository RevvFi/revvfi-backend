create table if not exists auth_nonces (
    wallet_address text primary key,
    nonce text not null,
    expires_at timestamptz not null,
    created_at timestamptz not null default now()
);

create table if not exists auth_sessions (
    id bigserial primary key,
    wallet_address text not null,
    token text not null unique,
    expires_at timestamptz not null,
    revoked_at timestamptz,
    created_at timestamptz not null default now()
);

create table if not exists markets (
    id bigserial primary key,
    address text not null unique,
    borrower text not null,
    borrow_asset text not null,
    borrow_asset_symbol text,
    borrow_asset_decimals smallint not null default 18,
    collateral_asset text not null,
    collateral_asset_symbol text,
    collateral_asset_decimals smallint not null default 18,
    collateral_oracle text not null,
    min_collateral_ratio integer not null,
    liquidation_threshold integer not null,
    total_principal text not null default '0',
    total_accrued_interest text not null default '0',
    total_debt text not null default '0',
    total_liquidity text not null default '0',
    borrow_index text not null default '1000000000000000000',
    weighted_avg_apr integer not null default 0,
    utilization_rate double precision not null default 0,
    is_active boolean not null default true,
    is_liquidating boolean not null default false,
    is_closed boolean not null default false,
    current_auction_id bigint,
    created_at timestamptz not null default now(),
    closed_at timestamptz,
    last_interest_accrual timestamptz not null default now(),
    last_health_check timestamptz
);

create table if not exists offers (
    id bigserial primary key,
    offer_id bigint not null unique,
    lender text not null,
    market_address text not null,
    amount text not null,
    remaining_amount text not null,
    apr integer not null,
    seniority smallint not null,
    status text not null,
    expiry timestamptz,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    cancelled_at timestamptz,
    expired_at timestamptz,
    block_number bigint not null default 0,
    tx_hash text not null default ''
);

create table if not exists positions (
    id bigserial primary key,
    token_id bigint not null unique,
    lender text not null,
    market_address text not null,
    principal text not null,
    current_principal text not null,
    accrued_interest text not null default '0',
    claimable_amount text not null default '0',
    apr integer not null,
    seniority smallint not null,
    status text not null,
    is_active boolean not null default true,
    is_settled boolean not null default false,
    minted_at timestamptz not null default now(),
    settled_at timestamptz,
    last_updated timestamptz not null default now(),
    block_number bigint not null default 0,
    tx_hash text not null default '',
    log_index integer not null default 0
);

create table if not exists borrowers (
    address text primary key,
    total_borrowed text not null default '0',
    total_repaid text not null default '0',
    outstanding_debt text not null default '0',
    total_loans integer not null default 0,
    successful_loans integer not null default 0,
    defaulted_loans integer not null default 0,
    active_loans integer not null default 0,
    reputation_score integer not null default 500,
    risk_label text not null default 'B',
    success_rate double precision not null default 0,
    registered_at timestamptz not null default now(),
    last_activity timestamptz,
    last_reputation_update timestamptz not null default now()
);

create table if not exists repayments (
    id bigserial primary key,
    market_address text not null,
    borrower_address text not null,
    amount text not null,
    interest_paid text not null,
    principal_paid text not null,
    repayment_type text not null,
    block_number bigint not null,
    tx_hash text not null,
    created_at timestamptz not null default now()
);

create table if not exists auctions (
    id bigserial primary key,
    auction_id bigint not null unique,
    market_address text not null,
    borrower_address text not null default '',
    collateral_amount text not null default '0',
    collateral_value text not null default '0',
    debt_amount text not null default '0',
    highest_bid text not null default '0',
    highest_bidder text,
    current_price text not null default '0',
    start_time timestamptz not null,
    end_time timestamptz not null,
    settlement_time timestamptz,
    status text not null,
    winning_bid text not null default '0',
    winner text,
    recovery_rate double precision not null default 0,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table if not exists withdrawal_requests (
    id bigserial primary key,
    request_id bigserial unique,
    lender text not null,
    position_id bigint not null,
    market_address text not null default '',
    requested_amount text not null,
    fulfilled_amount text not null default '0',
    remaining_amount text not null default '0',
    epoch_number integer not null,
    status text not null,
    fulfillment_time timestamptz,
    fulfillment_tx_hash text,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table if not exists withdrawal_epochs (
    id bigserial primary key,
    epoch_number integer not null unique,
    start_time timestamptz not null,
    end_time timestamptz not null,
    processed_at timestamptz,
    total_requested text not null default '0',
    total_fulfilled text not null default '0',
    status text not null,
    created_at timestamptz not null default now()
);

create table if not exists processed_events (
    id bigserial primary key,
    tx_hash text not null,
    log_index integer not null,
    block_number bigint not null,
    event_name text not null,
    contract_address text not null,
    processed_at timestamptz not null default now(),
    unique(tx_hash, log_index)
);

create table if not exists failed_events (
    id bigserial primary key,
    payload jsonb not null,
    reason text not null,
    created_at timestamptz not null default now()
);

create table if not exists chain_checkpoints (
    id bigserial primary key,
    block_number bigint not null unique,
    block_hash text not null,
    parent_hash text not null,
    is_finalized boolean not null default false,
    confirmation_count integer not null default 0,
    processed_at timestamptz not null default now()
);

create index if not exists idx_auth_sessions_wallet on auth_sessions(wallet_address);
create index if not exists idx_markets_borrower on markets(borrower);
create index if not exists idx_markets_active on markets(is_active);
create index if not exists idx_offers_market_status on offers(market_address, status);
create index if not exists idx_offers_lender on offers(lender);
create index if not exists idx_positions_lender on positions(lender);
create index if not exists idx_positions_market on positions(market_address);
create index if not exists idx_borrowers_reputation on borrowers(reputation_score desc);
create index if not exists idx_auctions_market_status on auctions(market_address, status);
create index if not exists idx_withdrawals_lender on withdrawal_requests(lender);
create index if not exists idx_withdrawals_epoch_status on withdrawal_requests(epoch_number, status);
create index if not exists idx_processed_events_block on processed_events(block_number);
