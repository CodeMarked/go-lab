package authstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// EconomyLedgerEvent is one append-only row from economy_ledger_events.
type EconomyLedgerEvent struct {
	ID             int64
	CreatedAt      time.Time
	PlatformUserID int
	EventType      string
	AmountDelta    int64
	CurrencyCode   string
	ReferenceType  sql.NullString
	ReferenceID    sql.NullString
	MetaJSON       []byte
}

// EconomyLedgerQuery filters and paginates ledger reads (newest first by id).
type EconomyLedgerQuery struct {
	Limit          int
	BeforeID       *int64
	PlatformUserID *int
	EventType      string
	FromTime       *time.Time
	ToTime         *time.Time
}

// ListEconomyLedgerEvents returns ledger rows matching the query.
func (s *Store) ListEconomyLedgerEvents(ctx context.Context, q EconomyLedgerQuery) ([]EconomyLedgerEvent, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("nil db")
	}
	if q.Limit <= 0 {
		q.Limit = 50
	}
	if q.Limit > 100 {
		q.Limit = 100
	}

	var b strings.Builder
	b.WriteString(`SELECT id, created_at, platform_user_id, event_type, amount_delta, currency_code,
		reference_type, reference_id, meta_json
		FROM economy_ledger_events WHERE 1=1`)
	args := make([]any, 0, 8)
	if q.PlatformUserID != nil {
		b.WriteString(` AND platform_user_id = ?`)
		args = append(args, *q.PlatformUserID)
	}
	if q.EventType != "" {
		b.WriteString(` AND event_type = ?`)
		args = append(args, q.EventType)
	}
	if q.FromTime != nil {
		b.WriteString(` AND created_at >= ?`)
		args = append(args, q.FromTime.UTC())
	}
	if q.ToTime != nil {
		b.WriteString(` AND created_at <= ?`)
		args = append(args, q.ToTime.UTC())
	}
	if q.BeforeID != nil {
		b.WriteString(` AND id < ?`)
		args = append(args, *q.BeforeID)
	}
	b.WriteString(` ORDER BY id DESC LIMIT ?`)
	args = append(args, q.Limit)

	rows, err := s.db.QueryContext(ctx, b.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("economy ledger query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []EconomyLedgerEvent
	for rows.Next() {
		var r EconomyLedgerEvent
		var meta []byte
		if err := rows.Scan(&r.ID, &r.CreatedAt, &r.PlatformUserID, &r.EventType, &r.AmountDelta,
			&r.CurrencyCode, &r.ReferenceType, &r.ReferenceID, &meta); err != nil {
			return nil, fmt.Errorf("economy ledger scan: %w", err)
		}
		if len(meta) > 0 {
			r.MetaJSON = meta
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if out == nil {
		out = []EconomyLedgerEvent{}
	}
	return out, nil
}
