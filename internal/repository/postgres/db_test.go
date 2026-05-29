package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

/*
@function TestWithTxCommitAndRollback

@desc
Tests PostgreSQL transaction helper commit and rollback behavior.
*/
func TestWithTxCommitAndRollback(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sql mock: %v", err)
	}
	defer conn.Close()

	db := Wrap(conn)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectCommit()
	if err := db.WithTx(ctx, func(tx *sql.Tx) error { return nil }); err != nil {
		t.Fatalf("commit transaction: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectRollback()
	if err := db.WithTx(ctx, func(tx *sql.Tx) error { return errors.New("fail") }); err == nil {
		t.Fatal("expected rollback error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
