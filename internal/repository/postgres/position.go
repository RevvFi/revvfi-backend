package postgres

import (
	"context"
	"database/sql"
	"math/big"

	"github.com/Revvfi/revvfi-backend/internal/models"
)

/*
@struct PositionRepository

@desc
PostgreSQL implementation for lender position persistence.
*/
type PositionRepository struct {
	db *DB
}

/*
@function NewPositionRepository

@desc
Creates a PostgreSQL position repository.
*/
func NewPositionRepository(db *DB) *PositionRepository {
	return &PositionRepository{db: db}
}

/*
@method CreatePosition

@desc
Inserts a lender position record.
*/
func (r *PositionRepository) CreatePosition(ctx context.Context, position *models.Position) error {
	_, err := r.db.conn.ExecContext(ctx, `insert into positions(token_id,lender,market_address,principal,current_principal,accrued_interest,claimable_amount,apr,seniority,status,is_active,is_settled,minted_at,last_updated,block_number,tx_hash,log_index) values($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		position.TokenID, position.Lender, position.MarketAddress, stringOrZeroDB(position.Principal), stringOrZeroDB(position.CurrentPrincipal), stringOrZeroDB(position.AccruedInterest), stringOrZeroDB(position.ClaimableAmount), position.APR, position.Seniority, position.Status, position.IsActive, position.IsSettled, now(), now(), position.BlockNumber, position.TxHash, position.LogIndex)
	return mapError(err)
}

/*
@method GetByTokenID

@desc
Fetches a position by NFT token ID.
*/
func (r *PositionRepository) GetByTokenID(ctx context.Context, tokenID int64) (*models.Position, error) {
	row := r.db.conn.QueryRowContext(ctx, `select id,token_id,lender,market_address,principal,current_principal,accrued_interest,claimable_amount,apr,seniority,status,is_active,is_settled,minted_at,settled_at,last_updated,block_number,tx_hash,log_index from positions where token_id=$1`, tokenID)
	position, err := scanPosition(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, mapError(err)
	}
	return position, nil
}

/*
@method GetByLender

@desc
Lists positions owned by a lender.
*/
func (r *PositionRepository) GetByLender(ctx context.Context, lender string, limit, offset int32) ([]models.Position, error) {
	rows, err := r.db.conn.QueryContext(ctx, `select id,token_id,lender,market_address,principal,current_principal,accrued_interest,claimable_amount,apr,seniority,status,is_active,is_settled,minted_at,settled_at,last_updated,block_number,tx_hash,log_index from positions where lender=$1 order by minted_at desc limit $2 offset $3`, lender, limit, offset)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	return scanPositions(rows)
}

/*
@method UpdatePosition

@desc
Updates mutable position state.
*/
func (r *PositionRepository) UpdatePosition(ctx context.Context, position *models.Position) error {
	_, err := r.db.conn.ExecContext(ctx, `update positions set current_principal=$2,accrued_interest=$3,claimable_amount=$4,status=$5,is_active=$6,is_settled=$7,settled_at=$8,last_updated=$9 where token_id=$1`,
		position.TokenID, stringOrZeroDB(position.CurrentPrincipal), stringOrZeroDB(position.AccruedInterest), stringOrZeroDB(position.ClaimableAmount), position.Status, position.IsActive, position.IsSettled, position.SettledAt, now())
	return mapError(err)
}

/*
@method CountActiveByLender

@desc
Counts active positions for a lender.
*/
func (r *PositionRepository) CountActiveByLender(ctx context.Context, lender string) (int32, error) {
	var count int32
	err := r.db.conn.QueryRowContext(ctx, `select count(*) from positions where lender=$1 and is_active=true`, lender).Scan(&count)
	return count, mapError(err)
}

/*
@method CountActiveByMarket

@desc
Counts active positions for a market.
*/
func (r *PositionRepository) CountActiveByMarket(ctx context.Context, marketAddr string) (int32, error) {
	var count int32
	err := r.db.conn.QueryRowContext(ctx, `select count(*) from positions where market_address=$1 and is_active=true`, marketAddr).Scan(&count)
	return count, mapError(err)
}

/*
@method GetTotalValueByLender

@desc
Sums principal and accrued interest for a lender.
*/
func (r *PositionRepository) GetTotalValueByLender(ctx context.Context, lender string) (*big.Int, error) {
	var total sql.NullString
	err := r.db.conn.QueryRowContext(ctx, `select coalesce(sum((current_principal::numeric + accrued_interest::numeric))::text,'0') from positions where lender=$1`, lender).Scan(&total)
	return parseInt(total.String), mapError(err)
}

/*
@method GetTotalPrincipalByMarket

@desc
Sums active principal for a market.
*/
func (r *PositionRepository) GetTotalPrincipalByMarket(ctx context.Context, marketAddr string) (*big.Int, error) {
	var total sql.NullString
	err := r.db.conn.QueryRowContext(ctx, `select coalesce(sum(principal::numeric)::text,'0') from positions where market_address=$1 and is_active=true`, marketAddr).Scan(&total)
	return parseInt(total.String), mapError(err)
}

/*
@method Update

@desc
Compatibility method for updating positions.
*/
func (r *PositionRepository) Update(ctx context.Context, position *models.Position) error {
	return r.UpdatePosition(ctx, position)
}

/*
@method Create

@desc
Compatibility method for creating positions.
*/
func (r *PositionRepository) Create(ctx context.Context, position *models.Position) error {
	return r.CreatePosition(ctx, position)
}

/*
@function scanPosition

@desc
Scans a SQL row into a position model.
*/
func scanPosition(scanner interface {
	Scan(dest ...interface{}) error
}) (*models.Position, error) {
	var position models.Position
	var principal, current, interest, claimable string
	err := scanner.Scan(&position.ID, &position.TokenID, &position.Lender, &position.MarketAddress, &principal, &current, &interest, &claimable, &position.APR, &position.Seniority, &position.Status, &position.IsActive, &position.IsSettled, &position.MintedAt, &position.SettledAt, &position.LastUpdated, &position.BlockNumber, &position.TxHash, &position.LogIndex)
	if err != nil {
		return nil, err
	}
	position.Principal = parseInt(principal)
	position.CurrentPrincipal = parseInt(current)
	position.AccruedInterest = parseInt(interest)
	position.ClaimableAmount = parseInt(claimable)
	return &position, nil
}

/*
@function scanPositions

@desc
Scans SQL rows into position models.
*/
func scanPositions(rows *sql.Rows) ([]models.Position, error) {
	items := make([]models.Position, 0)
	for rows.Next() {
		item, err := scanPosition(rows)
		if err != nil {
			return nil, mapError(err)
		}
		items = append(items, *item)
	}
	return items, mapError(rows.Err())
}
