package authstore

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestListEconomyLedgerEventsEmpty(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	s := New(db, time.Hour, time.Hour)

	mock.ExpectQuery(`SELECT id, created_at, platform_user_id, event_type, amount_delta, currency_code`).
		WithArgs(50).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "platform_user_id", "event_type", "amount_delta", "currency_code",
			"reference_type", "reference_id", "meta_json",
		}))

	rows, err := s.ListEconomyLedgerEvents(context.Background(), EconomyLedgerQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("expected no rows, got %d", len(rows))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestListEconomyLedgerEventsWithFilters(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	s := New(db, time.Hour, time.Hour)

	uid := 42
	before := int64(100)
	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)

	mock.ExpectQuery(`SELECT id, created_at, platform_user_id, event_type, amount_delta, currency_code`).
		WithArgs(42, "grant", from.UTC(), to.UTC(), before, 10).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "platform_user_id", "event_type", "amount_delta", "currency_code",
			"reference_type", "reference_id", "meta_json",
		}).AddRow(99, from, 42, "grant", int64(5), "default", nil, nil, nil))

	rows, err := s.ListEconomyLedgerEvents(context.Background(), EconomyLedgerQuery{
		Limit:          10,
		BeforeID:       &before,
		PlatformUserID: &uid,
		EventType:      "grant",
		FromTime:       &from,
		ToTime:         &to,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].ID != 99 {
		t.Fatalf("unexpected rows: %+v", rows)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
