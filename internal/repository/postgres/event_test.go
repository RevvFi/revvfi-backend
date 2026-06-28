package postgres

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Revvfi/revvfi-backend/internal/models"
)

/*
@function TestEventRepositoryIdempotency

@desc
Tests processed event lookup and idempotent insert behavior.
*/
func TestEventRepositoryIdempotency(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sql mock: %v", err)
	}
	defer conn.Close()

	repo := NewEventRepository(Wrap(conn))
	ctx := context.Background()

	mock.ExpectQuery("(?i)select exists").WithArgs("0xabc", 3).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	processed, err := repo.IsProcessed(ctx, "0xabc", 3)
	if err != nil {
		t.Fatalf("check processed: %v", err)
	}
	if processed {
		t.Fatal("expected event to be unprocessed")
	}

	mock.ExpectExec("(?i)insert into processed_events").WithArgs("0xabc", int32(3), int64(10), "OfferSubmitted", "0xMarket", sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	if err := repo.MarkProcessed(ctx, &models.ProcessedEvent{TxHash: "0xabc", LogIndex: 3, BlockNumber: 10, EventName: "OfferSubmitted", ContractAddress: "0xMarket"}); err != nil {
		t.Fatalf("mark processed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
