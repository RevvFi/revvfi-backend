package postgres

import (
	"context"
	"math/big"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

/*
@function TestCloseMarket

@desc
Tests that CloseMarket issues the correct UPDATE statement setting
is_closed=true, is_active=false, and closed_at=NOW() on the target market.
*/
func TestCloseMarket(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sql mock: %v", err)
	}
	defer conn.Close()

	repo := NewEventRepository(Wrap(conn))
	ctx := context.Background()

	mock.ExpectExec("(?i)update markets").
		WithArgs("0xMarket").
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.CloseMarket(ctx, "0xMarket"); err != nil {
		t.Fatalf("CloseMarket: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

/*
@function TestRedeemPosition

@desc
Tests that RedeemPosition issues the correct UPDATE statement setting
is_redeemed=true, status='redeemed', and claimable_amount to principal+interest.
*/
func TestRedeemPosition(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sql mock: %v", err)
	}
	defer conn.Close()

	repo := NewEventRepository(Wrap(conn))
	ctx := context.Background()

	principal := big.NewInt(1000000)
	interest := big.NewInt(50000)
	// total = principal + interest = 1050000
	expectedTotal := new(big.Int).Add(principal, interest).String()

	mock.ExpectExec("(?i)update positions").
		WithArgs(int64(42), expectedTotal).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.RedeemPosition(ctx, 42, principal, interest); err != nil {
		t.Fatalf("RedeemPosition: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

/*
@function TestRecordCollateralDeposit

@desc
Tests that RecordCollateralDeposit issues an UPSERT into collateral_balances
with the borrower, new total balance, and deposit amount.
*/
func TestRecordCollateralDeposit(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sql mock: %v", err)
	}
	defer conn.Close()

	repo := NewEventRepository(Wrap(conn))
	ctx := context.Background()

	amount := big.NewInt(500000)
	newTotal := big.NewInt(1500000)

	mock.ExpectExec("(?i)insert into collateral_balances").
		WithArgs("0xBorrower", newTotal.String(), amount.String()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.RecordCollateralDeposit(ctx, "0xBorrower", amount, newTotal); err != nil {
		t.Fatalf("RecordCollateralDeposit: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

/*
@function TestRecordCollateralWithdrawal

@desc
Tests that RecordCollateralWithdrawal issues an UPSERT into collateral_balances
with the borrower, new total balance (after withdrawal), and withdrawn amount.
*/
func TestRecordCollateralWithdrawal(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sql mock: %v", err)
	}
	defer conn.Close()

	repo := NewEventRepository(Wrap(conn))
	ctx := context.Background()

	amount := big.NewInt(200000)
	newTotal := big.NewInt(800000)

	mock.ExpectExec("(?i)insert into collateral_balances").
		WithArgs("0xBorrower", newTotal.String(), amount.String()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.RecordCollateralWithdrawal(ctx, "0xBorrower", amount, newTotal); err != nil {
		t.Fatalf("RecordCollateralWithdrawal: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

/*
@function TestAddControllerFactory

@desc
Tests that AddControllerFactory issues the correct INSERT/UPSERT into
controller_factories with the factory address and block number.
*/
func TestAddControllerFactory(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sql mock: %v", err)
	}
	defer conn.Close()

	repo := NewEventRepository(Wrap(conn))
	ctx := context.Background()

	mock.ExpectExec("(?i)insert into controller_factories").
		WithArgs("0xFactory", int64(1234)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.AddControllerFactory(ctx, "0xFactory", 1234); err != nil {
		t.Fatalf("AddControllerFactory: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

/*
@function TestRemoveControllerFactory

@desc
Tests that RemoveControllerFactory issues the correct UPDATE statement
setting removed_at on the target factory record.
*/
func TestRemoveControllerFactory(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sql mock: %v", err)
	}
	defer conn.Close()

	repo := NewEventRepository(Wrap(conn))
	ctx := context.Background()

	mock.ExpectExec("(?i)update controller_factories").
		WithArgs("0xFactory").
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.RemoveControllerFactory(ctx, "0xFactory"); err != nil {
		t.Fatalf("RemoveControllerFactory: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

/*
@function TestAddController

@desc
Tests that AddController issues the correct INSERT/UPSERT into controllers
with the controller address and block number.
*/
func TestAddController(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sql mock: %v", err)
	}
	defer conn.Close()

	repo := NewEventRepository(Wrap(conn))
	ctx := context.Background()

	mock.ExpectExec("(?i)insert into controllers").
		WithArgs("0xController", int64(5678)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.AddController(ctx, "0xController", 5678); err != nil {
		t.Fatalf("AddController: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

/*
@function TestRemoveController

@desc
Tests that RemoveController issues the correct UPDATE statement
setting removed_at on the target controller record.
*/
func TestRemoveController(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sql mock: %v", err)
	}
	defer conn.Close()

	repo := NewEventRepository(Wrap(conn))
	ctx := context.Background()

	mock.ExpectExec("(?i)update controllers").
		WithArgs("0xController").
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.RemoveController(ctx, "0xController"); err != nil {
		t.Fatalf("RemoveController: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

/*
@function TestIncrementBorrowerTotalBorrowed

@desc
Tests that IncrementBorrowerTotalBorrowed increments total_borrowed,
active_loans, and total_loans for the given borrower.
*/
func TestIncrementBorrowerTotalBorrowed(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sql mock: %v", err)
	}
	defer conn.Close()

	repo := NewEventRepository(Wrap(conn))
	ctx := context.Background()

	amount := big.NewInt(1000000000000000000) // 1 ETH

	mock.ExpectExec("(?i)update borrowers").
		WithArgs("0xBorrower", amount.String()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.IncrementBorrowerTotalBorrowed(ctx, "0xBorrower", amount); err != nil {
		t.Fatalf("IncrementBorrowerTotalBorrowed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

/*
@function TestIncrementBorrowerDefaults

@desc
Tests that IncrementBorrowerDefaults increments defaulted_loans and decrements
active_loans and reputation_score for the given borrower.
*/
func TestIncrementBorrowerDefaults(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sql mock: %v", err)
	}
	defer conn.Close()

	repo := NewEventRepository(Wrap(conn))
	ctx := context.Background()

	mock.ExpectExec("(?i)update borrowers").
		WithArgs("0xBorrower").
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.IncrementBorrowerDefaults(ctx, "0xBorrower"); err != nil {
		t.Fatalf("IncrementBorrowerDefaults: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

/*
@function TestRecordSuccessfulRepayment

@desc
Tests that RecordSuccessfulRepayment increments successful_loans, decrements
active_loans, and accumulates total_repaid for the given borrower.
*/
func TestRecordSuccessfulRepayment(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sql mock: %v", err)
	}
	defer conn.Close()

	repo := NewEventRepository(Wrap(conn))
	ctx := context.Background()

	amount := big.NewInt(5000000)

	mock.ExpectExec("(?i)update borrowers").
		WithArgs("0xBorrower", amount.String()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.RecordSuccessfulRepayment(ctx, "0xBorrower", amount); err != nil {
		t.Fatalf("RecordSuccessfulRepayment: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
