package postgres

import (
	"context"
	"database/sql"

	"github.com/Revvfi/revvfi-backend/internal/models"
)

/*
@struct WithdrawalRepository

@desc
PostgreSQL implementation for withdrawal request and epoch persistence.
*/
type WithdrawalRepository struct{ db *DB }

/*
@function NewWithdrawalRepository

@desc
Creates a PostgreSQL withdrawal repository.
*/
func NewWithdrawalRepository(db *DB) *WithdrawalRepository { return &WithdrawalRepository{db: db} }

/*
@method CreateRequest

@desc
Inserts a withdrawal request.
*/
func (r *WithdrawalRepository) CreateRequest(ctx context.Context, request *models.WithdrawalRequest) error {
	err := r.db.conn.QueryRowContext(ctx, `insert into withdrawal_requests(lender,position_id,market_address,requested_amount,fulfilled_amount,remaining_amount,epoch_number,status,created_at,updated_at) values($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) returning request_id`,
		request.Lender, request.PositionID, request.MarketAddress, stringOrZeroDB(request.RequestedAmount), stringOrZeroDB(request.FulfilledAmount), stringOrZeroDB(request.RemainingAmount), request.EpochNumber, request.Status, now(), now()).Scan(&request.RequestID)
	return mapError(err)
}

/*
@method GetRequestByID

@desc
Fetches a withdrawal request by ID.
*/
func (r *WithdrawalRepository) GetRequestByID(ctx context.Context, requestID int64) (*models.WithdrawalRequest, error) {
	row := r.db.conn.QueryRowContext(ctx, `select id,request_id,lender,position_id,market_address,requested_amount,fulfilled_amount,remaining_amount,epoch_number,status,fulfillment_time,fulfillment_tx_hash,created_at,updated_at from withdrawal_requests where request_id=$1`, requestID)
	request, err := scanWithdrawalRequest(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, mapError(err)
	}
	return request, nil
}

/*
@method GetRequestsByLender

@desc
Lists withdrawal requests by lender.
*/
func (r *WithdrawalRepository) GetRequestsByLender(ctx context.Context, lender string, limit, offset int32) ([]models.WithdrawalRequest, error) {
	rows, err := r.db.conn.QueryContext(ctx, `select id,request_id,lender,position_id,market_address,requested_amount,fulfilled_amount,remaining_amount,epoch_number,status,fulfillment_time,fulfillment_tx_hash,created_at,updated_at from withdrawal_requests where lender=$1 order by created_at desc limit $2 offset $3`, lender, limit, offset)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	return scanWithdrawalRequests(rows)
}

/*
@method UpdateRequest

@desc
Updates withdrawal request state.
*/
func (r *WithdrawalRepository) UpdateRequest(ctx context.Context, request *models.WithdrawalRequest) error {
	_, err := r.db.conn.ExecContext(ctx, `update withdrawal_requests set fulfilled_amount=$2,remaining_amount=$3,status=$4,fulfillment_time=$5,fulfillment_tx_hash=$6,updated_at=$7 where request_id=$1`,
		request.RequestID, stringOrZeroDB(request.FulfilledAmount), stringOrZeroDB(request.RemainingAmount), request.Status, request.FulfillmentTime, request.FulfillmentTxHash, now())
	return mapError(err)
}

/*
@method GetActiveRequests

@desc
Lists active withdrawal requests for an epoch.
*/
func (r *WithdrawalRepository) GetActiveRequests(ctx context.Context, epochNumber int64) ([]models.WithdrawalRequest, error) {
	rows, err := r.db.conn.QueryContext(ctx, `select id,request_id,lender,position_id,market_address,requested_amount,fulfilled_amount,remaining_amount,epoch_number,status,fulfillment_time,fulfillment_tx_hash,created_at,updated_at from withdrawal_requests where epoch_number=$1 and status in ('pending','ready')`, epochNumber)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	return scanWithdrawalRequests(rows)
}

/*
@method CreateEpoch

@desc
Inserts a withdrawal epoch.
*/
func (r *WithdrawalRepository) CreateEpoch(ctx context.Context, epoch *models.WithdrawalEpoch) error {
	_, err := r.db.conn.ExecContext(ctx, `insert into withdrawal_epochs(epoch_number,start_time,end_time,processed_at,total_requested,total_fulfilled,status,created_at) values($1,$2,$3,$4,$5,$6,$7,$8)`,
		epoch.EpochNumber, epoch.StartTime, epoch.EndTime, epoch.ProcessedAt, stringOrZeroDB(epoch.TotalRequested), stringOrZeroDB(epoch.TotalFulfilled), epoch.Status, now())
	return mapError(err)
}

/*
@method GetEpochByNumber

@desc
Fetches a withdrawal epoch by sequence number.
*/
func (r *WithdrawalRepository) GetEpochByNumber(ctx context.Context, epochNumber int64) (*models.WithdrawalEpoch, error) {
	row := r.db.conn.QueryRowContext(ctx, `select id,epoch_number,start_time,end_time,processed_at,total_requested,total_fulfilled,status,created_at from withdrawal_epochs where epoch_number=$1`, epochNumber)
	epoch, err := scanWithdrawalEpoch(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, mapError(err)
	}
	return epoch, nil
}

/*
@method UpdateEpoch

@desc
Updates withdrawal epoch state.
*/
func (r *WithdrawalRepository) UpdateEpoch(ctx context.Context, epoch *models.WithdrawalEpoch) error {
	_, err := r.db.conn.ExecContext(ctx, `update withdrawal_epochs set processed_at=$2,total_requested=$3,total_fulfilled=$4,status=$5 where epoch_number=$1`, epoch.EpochNumber, epoch.ProcessedAt, stringOrZeroDB(epoch.TotalRequested), stringOrZeroDB(epoch.TotalFulfilled), epoch.Status)
	return mapError(err)
}

/*
@method GetCurrentEpoch

@desc
Fetches the current active withdrawal epoch.
*/
func (r *WithdrawalRepository) GetCurrentEpoch(ctx context.Context) (*models.WithdrawalEpoch, error) {
	row := r.db.conn.QueryRowContext(ctx, `select id,epoch_number,start_time,end_time,processed_at,total_requested,total_fulfilled,status,created_at from withdrawal_epochs where start_time <= now() and end_time >= now() order by epoch_number desc limit 1`)
	epoch, err := scanWithdrawalEpoch(row)
	if err == sql.ErrNoRows {
		return &models.WithdrawalEpoch{EpochNumber: 0, Status: "pending", TotalRequested: parseInt("0"), TotalFulfilled: parseInt("0")}, nil
	}
	if err != nil {
		return nil, mapError(err)
	}
	return epoch, nil
}

/*
@function scanWithdrawalRequest

@desc
Scans a SQL row into a withdrawal request model.
*/
func scanWithdrawalRequest(scanner interface {
	Scan(dest ...interface{}) error
}) (*models.WithdrawalRequest, error) {
	var request models.WithdrawalRequest
	var requested, fulfilled, remaining string
	err := scanner.Scan(&request.ID, &request.RequestID, &request.Lender, &request.PositionID, &request.MarketAddress, &requested, &fulfilled, &remaining, &request.EpochNumber, &request.Status, &request.FulfillmentTime, &request.FulfillmentTxHash, &request.CreatedAt, &request.UpdatedAt)
	if err != nil {
		return nil, err
	}
	request.RequestedAmount = parseInt(requested)
	request.FulfilledAmount = parseInt(fulfilled)
	request.RemainingAmount = parseInt(remaining)
	return &request, nil
}

/*
@function scanWithdrawalRequests

@desc
Scans SQL rows into withdrawal request models.
*/
func scanWithdrawalRequests(rows *sql.Rows) ([]models.WithdrawalRequest, error) {
	items := make([]models.WithdrawalRequest, 0)
	for rows.Next() {
		item, err := scanWithdrawalRequest(rows)
		if err != nil {
			return nil, mapError(err)
		}
		items = append(items, *item)
	}
	return items, mapError(rows.Err())
}

/*
@function scanWithdrawalEpoch

@desc
Scans a SQL row into a withdrawal epoch model.
*/
func scanWithdrawalEpoch(scanner interface {
	Scan(dest ...interface{}) error
}) (*models.WithdrawalEpoch, error) {
	var epoch models.WithdrawalEpoch
	var requested, fulfilled string
	err := scanner.Scan(&epoch.ID, &epoch.EpochNumber, &epoch.StartTime, &epoch.EndTime, &epoch.ProcessedAt, &requested, &fulfilled, &epoch.Status, &epoch.CreatedAt)
	if err != nil {
		return nil, err
	}
	epoch.TotalRequested = parseInt(requested)
	epoch.TotalFulfilled = parseInt(fulfilled)
	return &epoch, nil
}
