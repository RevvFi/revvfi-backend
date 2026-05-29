package postgres

import (
	"context"
	"database/sql"

	"github.com/Revvfi/revvfi-backend/internal/models"
)

/*
@struct BorrowerRepository

@desc
PostgreSQL implementation for borrower profile persistence.
*/
type BorrowerRepository struct{ db *DB }

/*
@function NewBorrowerRepository

@desc
Creates a PostgreSQL borrower repository.
*/
func NewBorrowerRepository(db *DB) *BorrowerRepository { return &BorrowerRepository{db: db} }

/*
@method CreateBorrower

@desc
Inserts a borrower profile.
*/
func (r *BorrowerRepository) CreateBorrower(ctx context.Context, borrower *models.Borrower) error {
	_, err := r.db.conn.ExecContext(ctx, `insert into borrowers(address,total_borrowed,total_repaid,outstanding_debt,total_loans,successful_loans,defaulted_loans,active_loans,reputation_score,risk_label,success_rate,registered_at,last_reputation_update) values($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		borrower.Address, stringOrZeroDB(borrower.TotalBorrowed), stringOrZeroDB(borrower.TotalRepaid), stringOrZeroDB(borrower.OutstandingDebt), borrower.TotalLoans, borrower.SuccessfulLoans, borrower.DefaultedLoans, borrower.ActiveLoans, borrower.ReputationScore, borrower.RiskLabel, borrower.SuccessRate, now(), now())
	return mapError(err)
}

/*
@method GetByAddress

@desc
Fetches a borrower by wallet address.
*/
func (r *BorrowerRepository) GetByAddress(ctx context.Context, address string) (*models.Borrower, error) {
	row := r.db.conn.QueryRowContext(ctx, `select address,total_borrowed,total_repaid,outstanding_debt,total_loans,successful_loans,defaulted_loans,active_loans,reputation_score,risk_label,success_rate,registered_at,last_activity,last_reputation_update from borrowers where address=$1`, address)
	borrower, err := scanBorrower(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, mapError(err)
	}
	return borrower, nil
}

/*
@method UpdateBorrower

@desc
Updates borrower profile and reputation fields.
*/
func (r *BorrowerRepository) UpdateBorrower(ctx context.Context, borrower *models.Borrower) error {
	_, err := r.db.conn.ExecContext(ctx, `update borrowers set total_borrowed=$2,total_repaid=$3,outstanding_debt=$4,total_loans=$5,successful_loans=$6,defaulted_loans=$7,active_loans=$8,reputation_score=$9,risk_label=$10,success_rate=$11,last_activity=$12,last_reputation_update=$13 where address=$1`,
		borrower.Address, stringOrZeroDB(borrower.TotalBorrowed), stringOrZeroDB(borrower.TotalRepaid), stringOrZeroDB(borrower.OutstandingDebt), borrower.TotalLoans, borrower.SuccessfulLoans, borrower.DefaultedLoans, borrower.ActiveLoans, borrower.ReputationScore, borrower.RiskLabel, borrower.SuccessRate, borrower.LastActivity, now())
	return mapError(err)
}

/*
@method Create

@desc
Compatibility method for borrower creation.
*/
func (r *BorrowerRepository) Create(ctx context.Context, borrower *models.Borrower) error {
	return r.CreateBorrower(ctx, borrower)
}

/*
@method Update

@desc
Compatibility method for borrower updates.
*/
func (r *BorrowerRepository) Update(ctx context.Context, borrower *models.Borrower) error {
	return r.UpdateBorrower(ctx, borrower)
}

/*
@method UpdateReputation

@desc
Updates only borrower reputation score.
*/
func (r *BorrowerRepository) UpdateReputation(ctx context.Context, address string, score int32) error {
	_, err := r.db.conn.ExecContext(ctx, `update borrowers set reputation_score=$2,last_reputation_update=$3 where address=$1`, address, score, now())
	return mapError(err)
}

/*
@method IncrementLoan

@desc
Increments total and active loan counters.
*/
func (r *BorrowerRepository) IncrementLoan(ctx context.Context, address string) error {
	_, err := r.db.conn.ExecContext(ctx, `update borrowers set total_loans=total_loans+1, active_loans=active_loans+1, last_activity=$2 where address=$1`, address, now())
	return mapError(err)
}

/*
@method RecordRepayment

@desc
Stores a repayment event.
*/
func (r *BorrowerRepository) RecordRepayment(ctx context.Context, repayment *models.Repayment) error {
	_, err := r.db.conn.ExecContext(ctx, `insert into repayments(market_address,borrower_address,amount,interest_paid,principal_paid,repayment_type,block_number,tx_hash,created_at) values($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		repayment.MarketAddress, repayment.BorrowerAddress, stringOrZeroDB(repayment.Amount), stringOrZeroDB(repayment.InterestPaid), stringOrZeroDB(repayment.PrincipalPaid), repayment.RepaymentType, repayment.BlockNumber, repayment.TxHash, now())
	return mapError(err)
}

/*
@method GetLeaderboard

@desc
Lists borrowers by reputation score.
*/
func (r *BorrowerRepository) GetLeaderboard(ctx context.Context, limit int32) ([]models.Borrower, error) {
	rows, err := r.db.conn.QueryContext(ctx, `select address,total_borrowed,total_repaid,outstanding_debt,total_loans,successful_loans,defaulted_loans,active_loans,reputation_score,risk_label,success_rate,registered_at,last_activity,last_reputation_update from borrowers order by reputation_score desc limit $1`, limit)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	items := make([]models.Borrower, 0)
	for rows.Next() {
		item, err := scanBorrower(rows)
		if err != nil {
			return nil, mapError(err)
		}
		items = append(items, *item)
	}
	return items, mapError(rows.Err())
}

/*
@function scanBorrower

@desc
Scans a SQL row into a borrower model.
*/
func scanBorrower(scanner interface {
	Scan(dest ...interface{}) error
}) (*models.Borrower, error) {
	var borrower models.Borrower
	var borrowed, repaid, debt string
	err := scanner.Scan(&borrower.Address, &borrowed, &repaid, &debt, &borrower.TotalLoans, &borrower.SuccessfulLoans, &borrower.DefaultedLoans, &borrower.ActiveLoans, &borrower.ReputationScore, &borrower.RiskLabel, &borrower.SuccessRate, &borrower.RegisteredAt, &borrower.LastActivity, &borrower.LastReputationUpdate)
	if err != nil {
		return nil, err
	}
	borrower.TotalBorrowed = parseInt(borrowed)
	borrower.TotalRepaid = parseInt(repaid)
	borrower.OutstandingDebt = parseInt(debt)
	return &borrower, nil
}
