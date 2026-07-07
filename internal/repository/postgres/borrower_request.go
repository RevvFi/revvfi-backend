package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Revvfi/revvfi-backend/internal/models"
	appErr "github.com/Revvfi/revvfi-backend/internal/pkg/errors"
)

/*
@struct BorrowerRequestRepository

@desc
PostgreSQL implementation for the off-chain borrower access request queue.
*/
type BorrowerRequestRepository struct{ db *DB }

/*
@function NewBorrowerRequestRepository

@desc
Creates a PostgreSQL borrower request repository.
*/
func NewBorrowerRequestRepository(db *DB) *BorrowerRequestRepository {
	return &BorrowerRequestRepository{db: db}
}

func scanBorrowerRequest(scanner interface{ Scan(...interface{}) error }) (*models.BorrowerRequest, error) {
	var req models.BorrowerRequest
	err := scanner.Scan(&req.ID, &req.WalletAddress, &req.Status, &req.Note, &req.RequestedAt, &req.DecidedAt, &req.DecidedBy)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

/*
@method Create

@desc
Inserts a new pending borrower request. Relies on the partial unique index
on (LOWER(wallet_address)) WHERE status='pending' to reject duplicates —
mapError translates the resulting unique-violation into ErrDuplicateEntry,
which the service layer turns into ErrBorrowerRequestAlreadyPending.
*/
func (r *BorrowerRequestRepository) Create(ctx context.Context, wallet string) (*models.BorrowerRequest, error) {
	row := r.db.conn.QueryRowContext(ctx, `
		insert into borrower_requests (wallet_address, status, requested_at)
		values ($1, 'pending', now())
		returning id, wallet_address, status, note, requested_at, decided_at, decided_by
	`, wallet)
	req, err := scanBorrowerRequest(row)
	if err != nil {
		return nil, mapError(err)
	}
	return req, nil
}

/*
@method GetLatestByWallet

@desc
Fetches the most recent borrower request for a wallet, if any (nil, nil if
none exists yet).
*/
func (r *BorrowerRequestRepository) GetLatestByWallet(ctx context.Context, wallet string) (*models.BorrowerRequest, error) {
	row := r.db.conn.QueryRowContext(ctx, `
		select id, wallet_address, status, note, requested_at, decided_at, decided_by
		from borrower_requests
		where lower(wallet_address) = lower($1)
		order by requested_at desc
		limit 1
	`, wallet)
	req, err := scanBorrowerRequest(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, mapError(err)
	}
	return req, nil
}

/*
@method GetByID

@desc
Fetches a single borrower request by its ID (nil, nil if not found).
*/
func (r *BorrowerRequestRepository) GetByID(ctx context.Context, id int64) (*models.BorrowerRequest, error) {
	row := r.db.conn.QueryRowContext(ctx, `
		select id, wallet_address, status, note, requested_at, decided_at, decided_by
		from borrower_requests where id = $1
	`, id)
	req, err := scanBorrowerRequest(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, mapError(err)
	}
	return req, nil
}

/*
@method List

@desc
Lists borrower requests, optionally filtered by status, oldest first (so
admins naturally review a FIFO queue).
*/
func (r *BorrowerRequestRepository) List(ctx context.Context, status string) ([]*models.BorrowerRequest, error) {
	var rows *sql.Rows
	var err error
	if status != "" {
		rows, err = r.db.conn.QueryContext(ctx, `
			select id, wallet_address, status, note, requested_at, decided_at, decided_by
			from borrower_requests where status = $1 order by requested_at asc
		`, status)
	} else {
		rows, err = r.db.conn.QueryContext(ctx, `
			select id, wallet_address, status, note, requested_at, decided_at, decided_by
			from borrower_requests order by requested_at asc
		`)
	}
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()

	var out []*models.BorrowerRequest
	for rows.Next() {
		req, err := scanBorrowerRequest(rows)
		if err != nil {
			return nil, mapError(err)
		}
		out = append(out, req)
	}
	return out, mapError(rows.Err())
}

/*
@method Reject

@desc
Marks a pending request as rejected. Guards against double-processing by
only updating rows still in 'pending' status — returns
ErrBorrowerRequestNotPending (via 0 rows affected) if the request was
already decided (approved/rejected) by a concurrent admin action.
*/
func (r *BorrowerRequestRepository) Reject(ctx context.Context, id int64, adminWallet, note string) error {
	res, err := r.db.conn.ExecContext(ctx, `
		update borrower_requests
		set status = 'rejected', decided_at = now(), decided_by = $2, note = $3
		where id = $1 and status = 'pending'
	`, id, adminWallet, sql.NullString{String: note, Valid: note != ""})
	if err != nil {
		return mapError(err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return mapError(err)
	}
	if n == 0 {
		return appErr.ErrBorrowerRequestNotPending
	}
	return nil
}

/*
@method ResolveApproved

@desc
Auto-resolves any pending request for a wallet to 'approved' when the
indexer observes an on-chain BorrowerAdded event. A no-op (no error) if no
pending request exists for that wallet — an admin can register a borrower
who never went through the request flow at all.
*/
func (r *BorrowerRequestRepository) ResolveApproved(ctx context.Context, wallet string) error {
	_, err := r.db.conn.ExecContext(ctx, `
		update borrower_requests
		set status = 'approved', decided_at = now(), decided_by = 'on-chain'
		where lower(wallet_address) = lower($1) and status = 'pending'
	`, wallet)
	return mapError(err)
}
