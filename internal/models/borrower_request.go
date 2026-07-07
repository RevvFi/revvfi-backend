package models

import (
	"database/sql"
	"time"
)

/*
@struct BorrowerRequest

@desc
Off-chain request from a wallet asking to be approved as a borrower.
registerBorrower() on RevvFiArchController is onlyOwner, so this table is
purely a review queue for the admin — it has no on-chain effect by itself.
Status transitions to "approved" only when the indexer observes a matching
on-chain BorrowerAdded event; "rejected" is a pure off-chain admin decision.
*/
type BorrowerRequest struct {
	ID            int64
	WalletAddress string
	Status        string // pending, approved, rejected
	Note          sql.NullString
	RequestedAt   time.Time
	DecidedAt     sql.NullTime
	DecidedBy     sql.NullString
}
