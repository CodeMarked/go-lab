package authstore

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionRevoked  = errors.New("session revoked")
	ErrSessionExpired  = errors.New("session expired")
)

// Store performs migration-backed session and audit persistence (no runtime DDL).
type Store struct {
	db        *sql.DB
	idleTTL   time.Duration
	absTTL    time.Duration
}

func New(db *sql.DB, idleTTL, absoluteTTL time.Duration) *Store {
	return &Store{db: db, idleTTL: idleTTL, absTTL: absoluteTTL}
}

func HashOpaqueToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// NewOpaqueToken returns a high-entropy opaque string suitable for HttpOnly cookies (hex-encoded random bytes).
func NewOpaqueToken() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

type SessionRow struct {
	ID                 int64
	UserID             int
	ExpiresAt          time.Time
	AbsoluteExpiresAt  time.Time
	RevokedAt          sql.NullTime
}

func (s *Store) CreateSession(ctx context.Context, userID int, rawToken, ip, ua string) error {
	if s.db == nil {
		return errors.New("nil db")
	}
	now := time.Now().UTC()
	abs := now.Add(s.absTTL)
	slide := now.Add(s.idleTTL)
	if slide.After(abs) {
		slide = abs
	}
	hash := HashOpaqueToken(rawToken)
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO auth_sessions (user_id, token_hash, expires_at, absolute_expires_at, ip, user_agent)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		userID, hash, slide.UTC(), abs.UTC(), nullStr(ip), nullStr(ua),
	)
	return err
}

func (s *Store) RevokeSessionByRawToken(ctx context.Context, rawToken string) error {
	if s.db == nil {
		return errors.New("nil db")
	}
	hash := HashOpaqueToken(rawToken)
	_, err := s.db.ExecContext(ctx,
		`UPDATE auth_sessions SET revoked_at = UTC_TIMESTAMP() WHERE token_hash = ? AND revoked_at IS NULL`,
		hash,
	)
	return err
}

// ValidateSession checks idle/absolute bounds and extends the sliding expiry. Does not rotate the token.
func (s *Store) ValidateSession(ctx context.Context, rawToken string) (userID int, err error) {
	if s.db == nil {
		return 0, errors.New("nil db")
	}
	hash := HashOpaqueToken(rawToken)
	var row SessionRow
	err = s.db.QueryRowContext(ctx,
		`SELECT id, user_id, expires_at, absolute_expires_at, revoked_at
		 FROM auth_sessions WHERE token_hash = ?`,
		hash,
	).Scan(&row.ID, &row.UserID, &row.ExpiresAt, &row.AbsoluteExpiresAt, &row.RevokedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrSessionNotFound
	}
	if err != nil {
		return 0, err
	}
	if row.RevokedAt.Valid {
		return 0, ErrSessionRevoked
	}
	now := time.Now().UTC()
	if !now.Before(row.AbsoluteExpiresAt.UTC()) || !now.Before(row.ExpiresAt.UTC()) {
		return 0, ErrSessionExpired
	}
	slide := now.Add(s.idleTTL)
	if slide.After(row.AbsoluteExpiresAt.UTC()) {
		slide = row.AbsoluteExpiresAt.UTC()
	}
	_, err = s.db.ExecContext(ctx,
		`UPDATE auth_sessions SET last_seen_at = UTC_TIMESTAMP(), expires_at = ? WHERE id = ? AND revoked_at IS NULL`,
		slide, row.ID,
	)
	if err != nil {
		return 0, err
	}
	return row.UserID, nil
}

// RefreshSession extends sliding expiry the same way as ValidateSession (explicit keep-alive).
func (s *Store) RefreshSession(ctx context.Context, rawToken string) (userID int, err error) {
	return s.ValidateSession(ctx, rawToken)
}

func (s *Store) InsertAudit(ctx context.Context, eventType string, userID *int, ip, ua, subjectHint string, metaJSON []byte) error {
	if s.db == nil {
		return errors.New("nil db")
	}
	var uid any
	if userID != nil {
		uid = *userID
	} else {
		uid = nil
	}
	var meta any
	if len(metaJSON) > 0 {
		meta = metaJSON
	} else {
		meta = nil
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO auth_audit_events (event_type, user_id, ip, user_agent, subject_hint, meta_json)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		eventType, uid, nullStr(ip), nullStr(ua), nullStr(subjectHint), meta,
	)
	return err
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}
