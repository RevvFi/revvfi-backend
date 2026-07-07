package postgres

import (
	"context"
	"database/sql"

	"github.com/Revvfi/revvfi-backend/internal/models"
)

/*
@struct AuctionRepository

@desc
PostgreSQL implementation for liquidation auction persistence.
*/
type AuctionRepository struct{ db *DB }

/*
@function NewAuctionRepository

@desc
Creates a PostgreSQL auction repository.
*/
func NewAuctionRepository(db *DB) *AuctionRepository { return &AuctionRepository{db: db} }

/*
@method CreateAuction

@desc
Inserts an auction record.
*/
func (r *AuctionRepository) CreateAuction(ctx context.Context, auction *models.Auction) error {
	_, err := r.db.conn.ExecContext(ctx, `insert into auctions(auction_id,market_address,borrower_address,collateral_amount,collateral_value,debt_amount,highest_bid,highest_bidder,current_price,start_time,end_time,status,winning_bid,winner,recovery_rate,created_at,updated_at) values($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		auction.AuctionID, auction.MarketAddress, auction.BorrowerAddress, stringOrZeroDB(auction.CollateralAmount), stringOrZeroDB(auction.CollateralValue), stringOrZeroDB(auction.DebtAmount), stringOrZeroDB(auction.HighestBid), auction.HighestBidder, stringOrZeroDB(auction.CurrentPrice), auction.StartTime, auction.EndTime, auction.Status, stringOrZeroDB(auction.WinningBid), auction.Winner, auction.RecoveryRate, now(), now())
	return mapError(err)
}

/*
@method GetByID

@desc
Fetches an auction by market address or auction ID string.
*/
func (r *AuctionRepository) GetByID(ctx context.Context, auctionID string) (*models.Auction, error) {
	row := r.db.conn.QueryRowContext(ctx, `select id,auction_id,market_address,borrower_address,collateral_amount,collateral_value,debt_amount,highest_bid,highest_bidder,current_price,start_time,end_time,settlement_time,status,winning_bid,winner,recovery_rate,created_at,updated_at from auctions where market_address=$1 or auction_id::text=$1 order by created_at desc limit 1`, auctionID)
	auction, err := scanAuction(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, mapError(err)
	}
	return auction, nil
}

/*
@method GetActiveByMarket

@desc
Lists active auctions for a market.
*/
func (r *AuctionRepository) GetActiveByMarket(ctx context.Context, market string) ([]models.Auction, error) {
	rows, err := r.db.conn.QueryContext(ctx, `select id,auction_id,market_address,borrower_address,collateral_amount,collateral_value,debt_amount,highest_bid,highest_bidder,current_price,start_time,end_time,settlement_time,status,winning_bid,winner,recovery_rate,created_at,updated_at from auctions where market_address=$1 and status='active'`, market)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	return scanAuctions(rows)
}

/*
@method GetAllActive

@desc
Lists all active auctions across all markets.
*/
func (r *AuctionRepository) GetAllActive(ctx context.Context) ([]models.Auction, error) {
	rows, err := r.db.conn.QueryContext(ctx, `select id,auction_id,market_address,borrower_address,collateral_amount,collateral_value,debt_amount,highest_bid,highest_bidder,current_price,start_time,end_time,settlement_time,status,winning_bid,winner,recovery_rate,created_at,updated_at from auctions where status='active' order by created_at desc`)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	return scanAuctions(rows)
}

/*
@method UpdateAuction

@desc
Updates auction state.
*/
func (r *AuctionRepository) UpdateAuction(ctx context.Context, auction *models.Auction) error {
	_, err := r.db.conn.ExecContext(ctx, `update auctions set highest_bid=$2,highest_bidder=$3,current_price=$4,settlement_time=$5,status=$6,winning_bid=$7,winner=$8,recovery_rate=$9,updated_at=$10 where auction_id=$1`,
		auction.AuctionID, stringOrZeroDB(auction.HighestBid), auction.HighestBidder, stringOrZeroDB(auction.CurrentPrice), auction.SettlementTime, auction.Status, stringOrZeroDB(auction.WinningBid), auction.Winner, auction.RecoveryRate, now())
	return mapError(err)
}

/*
@method GetLiquidatableCount

@desc
Counts active liquidation auctions.
*/
func (r *AuctionRepository) GetLiquidatableCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.conn.QueryRowContext(ctx, `select count(*) from auctions where status='active'`).Scan(&count)
	return count, mapError(err)
}

/*
@method GetBidsByAuction

@desc
Lists every bid placed on an auction, most recent first.
*/
func (r *AuctionRepository) GetBidsByAuction(ctx context.Context, auctionID int64) ([]models.Bid, error) {
	rows, err := r.db.conn.QueryContext(ctx, `select id,auction_id,bidder,bid_amount,placed_at,tx_hash,created_at from bids where auction_id=$1 order by placed_at desc`, auctionID)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()

	items := make([]models.Bid, 0)
	for rows.Next() {
		var b models.Bid
		var bidAmount string
		if err := rows.Scan(&b.ID, &b.AuctionID, &b.Bidder, &bidAmount, &b.PlacedAt, &b.TxHash, &b.CreatedAt); err != nil {
			return nil, mapError(err)
		}
		b.BidAmount = parseInt(bidAmount)
		items = append(items, b)
	}
	return items, mapError(rows.Err())
}

/*
@function scanAuction

@desc
Scans a SQL row into an auction model.
*/
func scanAuction(scanner interface {
	Scan(dest ...interface{}) error
}) (*models.Auction, error) {
	var auction models.Auction
	var collateral, collateralValue, debt, highestBid, currentPrice, winningBid string
	err := scanner.Scan(&auction.ID, &auction.AuctionID, &auction.MarketAddress, &auction.BorrowerAddress, &collateral, &collateralValue, &debt, &highestBid, &auction.HighestBidder, &currentPrice, &auction.StartTime, &auction.EndTime, &auction.SettlementTime, &auction.Status, &winningBid, &auction.Winner, &auction.RecoveryRate, &auction.CreatedAt, &auction.UpdatedAt)
	if err != nil {
		return nil, err
	}
	auction.CollateralAmount = parseInt(collateral)
	auction.CollateralValue = parseInt(collateralValue)
	auction.DebtAmount = parseInt(debt)
	auction.HighestBid = parseInt(highestBid)
	auction.CurrentPrice = parseInt(currentPrice)
	auction.WinningBid = parseInt(winningBid)
	return &auction, nil
}

/*
@function scanAuctions

@desc
Scans SQL rows into auction models.
*/
func scanAuctions(rows *sql.Rows) ([]models.Auction, error) {
	items := make([]models.Auction, 0)
	for rows.Next() {
		item, err := scanAuction(rows)
		if err != nil {
			return nil, mapError(err)
		}
		items = append(items, *item)
	}
	return items, mapError(rows.Err())
}
