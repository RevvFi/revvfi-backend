-- Borrower access requests: off-chain queue that lets any signed-in wallet
-- ask to become a borrower, and lets an admin approve (by sending the
-- on-chain registerBorrower tx from their own wallet) or reject the request.
--
-- registerBorrower() on RevvFiArchController is onlyOwner, so there is no
-- way for an arbitrary wallet to self-register on-chain. This table exists
-- purely as an off-chain request/review queue; on-chain registration
-- (via ArchController.registerBorrower) remains the sole source of truth for
-- who is actually an approved borrower. When the indexer observes a
-- BorrowerAdded event for a wallet with a pending request, it auto-resolves
-- that request to 'approved' (see ArchControllerHandler.handleBorrowerAdded).

CREATE TABLE IF NOT EXISTS borrower_requests (
    id BIGSERIAL PRIMARY KEY,
    wallet_address TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    note TEXT,
    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    decided_at TIMESTAMPTZ,
    decided_by TEXT
);

-- Only one outstanding pending request per wallet at a time.
CREATE UNIQUE INDEX IF NOT EXISTS idx_borrower_requests_wallet_pending
    ON borrower_requests (LOWER(wallet_address))
    WHERE status = 'pending';

CREATE INDEX IF NOT EXISTS idx_borrower_requests_status ON borrower_requests (status);
CREATE INDEX IF NOT EXISTS idx_borrower_requests_wallet ON borrower_requests (LOWER(wallet_address));
