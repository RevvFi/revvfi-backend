package postgres

import (
	"context"
	"database/sql"

	"github.com/Revvfi/revvfi-backend/internal/models"
)

/*
@struct OfferRepository

@desc
PostgreSQL implementation for offer persistence.
*/
type OfferRepository struct {
	db *DB
}

/*
@function NewOfferRepository

@desc
Creates a PostgreSQL offer repository.
*/
func NewOfferRepository(db *DB) *OfferRepository {
	return &OfferRepository{db: db}
}

/*
@method CreateOffer

@desc
Inserts an offer record.
*/
func (r *OfferRepository) CreateOffer(ctx context.Context, offer *models.Offer) error {
	_, err := r.db.conn.ExecContext(ctx, `insert into offers(offer_id,lender,market_address,amount,remaining_amount,apr,seniority,status,expiry,created_at,updated_at,block_number,tx_hash) values($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		offer.OfferID, offer.Lender, offer.MarketAddress, stringOrZeroDB(offer.Amount), stringOrZeroDB(offer.RemainingAmount), offer.APR, offer.Seniority, offer.Status, offer.Expiry, now(), now(), offer.BlockNumber, offer.TxHash)
	return mapError(err)
}

/*
@method GetByID

@desc
Fetches an offer by offer ID.
*/
func (r *OfferRepository) GetByID(ctx context.Context, offerID int64) (*models.Offer, error) {
	row := r.db.conn.QueryRowContext(ctx, `select id,offer_id,lender,market_address,amount,remaining_amount,apr,seniority,status,expiry,created_at,updated_at,cancelled_at,expired_at,block_number,tx_hash from offers where offer_id=$1`, offerID)
	offer, err := scanOffer(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, mapError(err)
	}
	return offer, nil
}

/*
@method GetActiveByMarket

@desc
Lists active offers for a market.
*/
func (r *OfferRepository) GetActiveByMarket(ctx context.Context, marketAddr string, limit, offset int32) ([]models.Offer, error) {
	rows, err := r.db.conn.QueryContext(ctx, `select id,offer_id,lender,market_address,amount,remaining_amount,apr,seniority,status,expiry,created_at,updated_at,cancelled_at,expired_at,block_number,tx_hash from offers where market_address=$1 and status in ('active','partially_filled') order by apr asc, seniority asc limit $2 offset $3`, marketAddr, limit, offset)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	return scanOffers(rows)
}

/*
@method UpdateOffer

@desc
Updates mutable offer state.
*/
func (r *OfferRepository) UpdateOffer(ctx context.Context, offer *models.Offer) error {
	_, err := r.db.conn.ExecContext(ctx, `update offers set remaining_amount=$2,status=$3,updated_at=$4,cancelled_at=$5,expired_at=$6 where offer_id=$1`, offer.OfferID, stringOrZeroDB(offer.RemainingAmount), offer.Status, now(), offer.CancelledAt, offer.ExpiredAt)
	return mapError(err)
}

/*
@method CancelOffer

@desc
Marks an offer as cancelled.
*/
func (r *OfferRepository) CancelOffer(ctx context.Context, offerID int64) error {
	_, err := r.db.conn.ExecContext(ctx, `update offers set status='cancelled', cancelled_at=$2, updated_at=$2 where offer_id=$1`, offerID, now())
	return mapError(err)
}

/*
@method CountActive

@desc
Counts active offers for a market.
*/
func (r *OfferRepository) CountActive(ctx context.Context, marketAddr string) (int32, error) {
	var count int32
	err := r.db.conn.QueryRowContext(ctx, `select count(*) from offers where market_address=$1 and status in ('active','partially_filled')`, marketAddr).Scan(&count)
	return count, mapError(err)
}

/*
@method ListActive

@desc
Lists active offers by market and seniority.
*/
func (r *OfferRepository) ListActive(ctx context.Context, market string, seniority int16) ([]models.Offer, error) {
	rows, err := r.db.conn.QueryContext(ctx, `select id,offer_id,lender,market_address,amount,remaining_amount,apr,seniority,status,expiry,created_at,updated_at,cancelled_at,expired_at,block_number,tx_hash from offers where market_address=$1 and seniority=$2 and status in ('active','partially_filled') order by apr asc`, market, seniority)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	return scanOffers(rows)
}

/*
@method GetByMarket

@desc
Compatibility method listing offers for a market with count.
*/
func (r *OfferRepository) GetByMarket(ctx context.Context, market string, page, pageSize int32) ([]models.Offer, int64, error) {
	limit, offset := pageSize, (page-1)*pageSize
	items, err := r.GetActiveByMarket(ctx, market, limit, offset)
	return items, int64(len(items)), err
}

/*
@method GetByLender

@desc
Lists offers by lender with count.
*/
func (r *OfferRepository) GetByLender(ctx context.Context, lender string, page, pageSize int32) ([]models.Offer, int64, error) {
	rows, err := r.db.conn.QueryContext(ctx, `select id,offer_id,lender,market_address,amount,remaining_amount,apr,seniority,status,expiry,created_at,updated_at,cancelled_at,expired_at,block_number,tx_hash from offers where lender=$1 order by created_at desc limit $2 offset $3`, lender, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, mapError(err)
	}
	defer rows.Close()
	items, err := scanOffers(rows)
	return items, int64(len(items)), err
}

/*
@method Create

@desc
Compatibility method for creating offers.
*/
func (r *OfferRepository) Create(ctx context.Context, offer *models.Offer) error {
	return r.CreateOffer(ctx, offer)
}

/*
@method Update

@desc
Compatibility method for updating offers.
*/
func (r *OfferRepository) Update(ctx context.Context, offer *models.Offer) error {
	return r.UpdateOffer(ctx, offer)
}

/*
@method Cancel

@desc
Compatibility method for cancelling offers.
*/
func (r *OfferRepository) Cancel(ctx context.Context, offerID int64) error {
	return r.CancelOffer(ctx, offerID)
}

/*
@method GetBestOffers

@desc
Lists best active offers by APR.
*/
func (r *OfferRepository) GetBestOffers(ctx context.Context, market string, seniority int16, limit int32) ([]models.Offer, error) {
	rows, err := r.db.conn.QueryContext(ctx, `select id,offer_id,lender,market_address,amount,remaining_amount,apr,seniority,status,expiry,created_at,updated_at,cancelled_at,expired_at,block_number,tx_hash from offers where market_address=$1 and seniority=$2 and status in ('active','partially_filled') order by apr asc limit $3`, market, seniority, limit)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	return scanOffers(rows)
}

/*
@function scanOffer

@desc
Scans a SQL row into an offer model.
*/
func scanOffer(scanner interface {
	Scan(dest ...interface{}) error
}) (*models.Offer, error) {
	var offer models.Offer
	var amount, remaining string
	err := scanner.Scan(&offer.ID, &offer.OfferID, &offer.Lender, &offer.MarketAddress, &amount, &remaining, &offer.APR, &offer.Seniority, &offer.Status, &offer.Expiry, &offer.CreatedAt, &offer.UpdatedAt, &offer.CancelledAt, &offer.ExpiredAt, &offer.BlockNumber, &offer.TxHash)
	if err != nil {
		return nil, err
	}
	offer.Amount = parseInt(amount)
	offer.RemainingAmount = parseInt(remaining)
	return &offer, nil
}

/*
@function scanOffers

@desc
Scans SQL rows into offer models.
*/
func scanOffers(rows *sql.Rows) ([]models.Offer, error) {
	items := make([]models.Offer, 0)
	for rows.Next() {
		item, err := scanOffer(rows)
		if err != nil {
			return nil, mapError(err)
		}
		items = append(items, *item)
	}
	return items, mapError(rows.Err())
}
