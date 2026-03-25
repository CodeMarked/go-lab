package authstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Backup restore request workflow (Phase C): two distinct approvers, neither the requester.

const (
	BackupRestoreStatusPending   = "pending"
	BackupRestoreStatusApproved  = "approved"
	BackupRestoreStatusRejected  = "rejected"
	BackupRestoreStatusFulfilled = "fulfilled"
	BackupRestoreStatusCancelled = "cancelled"
)

var (
	ErrBackupRestoreNotFound       = errors.New("backup restore request not found")
	ErrBackupRestoreInvalidState   = errors.New("backup restore request is not in a valid state for this action")
	ErrBackupRestoreCannotApprove  = errors.New("requester cannot approve own restore request")
	ErrBackupRestoreDuplicateSlot  = errors.New("same approver cannot fill both approval slots")
	ErrBackupRestoreAlreadyFilled  = errors.New("approval slots already complete")
	ErrBackupRestoreNotRequester   = errors.New("only the requester may cancel this request")
)

// BackupRestoreRequestRow is one row from backup_restore_requests.
type BackupRestoreRequestRow struct {
	ID                  int64
	CreatedAt           time.Time
	UpdatedAt           time.Time
	RequestedByUserID   int
	Scope               string
	RestorePointLabel   string
	Reason              string
	Status              string
	RejectionReason     sql.NullString
	Approval1UserID     sql.NullInt64
	Approval1At         sql.NullTime
	Approval2UserID     sql.NullInt64
	Approval2At         sql.NullTime
	FulfilledAt         sql.NullTime
	FulfilledNote       sql.NullString
	RequestID           sql.NullString
}

// BackupRestoreStatusCounts aggregates rows by status for DataOps dashboard.
func (s *Store) BackupRestoreStatusCounts(ctx context.Context) (map[string]int64, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("nil db")
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT status, COUNT(*) FROM backup_restore_requests GROUP BY status`)
	if err != nil {
		return nil, fmt.Errorf("backup restore counts: %w", err)
	}
	defer func() { _ = rows.Close() }()
	out := make(map[string]int64)
	for rows.Next() {
		var st string
		var n int64
		if err := rows.Scan(&st, &n); err != nil {
			return nil, err
		}
		out[st] = n
	}
	return out, rows.Err()
}

// CreateBackupRestoreRequest inserts a new pending request.
func (s *Store) CreateBackupRestoreRequest(ctx context.Context, requestedByUserID int, scope, restorePointLabel, reason, reqID string) (int64, error) {
	if s == nil || s.db == nil {
		return 0, errors.New("nil db")
	}
	if requestedByUserID <= 0 || scope == "" || restorePointLabel == "" || reason == "" {
		return 0, errors.New("invalid backup restore request fields")
	}
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO backup_restore_requests
		 (requested_by_user_id, scope, restore_point_label, reason, status, request_id)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		requestedByUserID, scope, restorePointLabel, reason, BackupRestoreStatusPending, nullStr(reqID),
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// ListBackupRestoreRequests returns newest first. statusFilter empty = all statuses.
func (s *Store) ListBackupRestoreRequests(ctx context.Context, limit int, statusFilter string) ([]BackupRestoreRequestRow, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("nil db")
	}
	if limit < 1 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	q := `SELECT id, created_at, updated_at, requested_by_user_id, scope, restore_point_label, reason, status,
		rejection_reason, approval_1_user_id, approval_1_at, approval_2_user_id, approval_2_at,
		fulfilled_at, fulfilled_note, request_id
		FROM backup_restore_requests WHERE 1=1`
	args := []any{}
	if statusFilter != "" {
		q += ` AND status = ?`
		args = append(args, statusFilter)
	}
	q += ` ORDER BY id DESC LIMIT ?`
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list backup restore: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []BackupRestoreRequestRow
	for rows.Next() {
		var r BackupRestoreRequestRow
		if err := rows.Scan(
			&r.ID, &r.CreatedAt, &r.UpdatedAt, &r.RequestedByUserID, &r.Scope, &r.RestorePointLabel, &r.Reason, &r.Status,
			&r.RejectionReason, &r.Approval1UserID, &r.Approval1At, &r.Approval2UserID, &r.Approval2At,
			&r.FulfilledAt, &r.FulfilledNote, &r.RequestID,
		); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// GetBackupRestoreRequest loads one row by id.
func (s *Store) GetBackupRestoreRequest(ctx context.Context, id int64) (BackupRestoreRequestRow, error) {
	if s == nil || s.db == nil {
		return BackupRestoreRequestRow{}, errors.New("nil db")
	}
	var r BackupRestoreRequestRow
	err := s.db.QueryRowContext(ctx,
		`SELECT id, created_at, updated_at, requested_by_user_id, scope, restore_point_label, reason, status,
			rejection_reason, approval_1_user_id, approval_1_at, approval_2_user_id, approval_2_at,
			fulfilled_at, fulfilled_note, request_id
		 FROM backup_restore_requests WHERE id = ?`, id,
	).Scan(
		&r.ID, &r.CreatedAt, &r.UpdatedAt, &r.RequestedByUserID, &r.Scope, &r.RestorePointLabel, &r.Reason, &r.Status,
		&r.RejectionReason, &r.Approval1UserID, &r.Approval1At, &r.Approval2UserID, &r.Approval2At,
		&r.FulfilledAt, &r.FulfilledNote, &r.RequestID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return BackupRestoreRequestRow{}, ErrBackupRestoreNotFound
	}
	return r, err
}

// ApproveBackupRestoreRequest records one approval; when two distinct non-requester approvers are set, status becomes approved.
func (s *Store) ApproveBackupRestoreRequest(ctx context.Context, id int64, approverUserID int) error {
	if s == nil || s.db == nil {
		return errors.New("nil db")
	}
	if approverUserID <= 0 {
		return errors.New("invalid approver")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var reqBy int
	var st string
	var a1, a2 sql.NullInt64
	err = tx.QueryRowContext(ctx,
		`SELECT requested_by_user_id, status, approval_1_user_id, approval_2_user_id
		 FROM backup_restore_requests WHERE id = ? FOR UPDATE`, id,
	).Scan(&reqBy, &st, &a1, &a2)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrBackupRestoreNotFound
	}
	if err != nil {
		return err
	}
	if st != BackupRestoreStatusPending {
		return ErrBackupRestoreInvalidState
	}
	if reqBy == approverUserID {
		return ErrBackupRestoreCannotApprove
	}
	now := time.Now().UTC()
	if !a1.Valid {
		_, err = tx.ExecContext(ctx,
			`UPDATE backup_restore_requests SET approval_1_user_id = ?, approval_1_at = ? WHERE id = ? AND status = ?`,
			approverUserID, now, id, BackupRestoreStatusPending)
		if err != nil {
			return err
		}
	} else if int(a1.Int64) == approverUserID {
		return ErrBackupRestoreDuplicateSlot
	} else if !a2.Valid {
		_, err = tx.ExecContext(ctx,
			`UPDATE backup_restore_requests SET approval_2_user_id = ?, approval_2_at = ?, status = ? WHERE id = ? AND status = ?`,
			approverUserID, now, BackupRestoreStatusApproved, id, BackupRestoreStatusPending)
		if err != nil {
			return err
		}
	} else {
		return ErrBackupRestoreAlreadyFilled
	}
	return tx.Commit()
}

// RejectBackupRestoreRequest rejects a pending request (approver must not be the requester).
func (s *Store) RejectBackupRestoreRequest(ctx context.Context, id int64, actorUserID int, rejectionReason string) error {
	if s == nil || s.db == nil {
		return errors.New("nil db")
	}
	if actorUserID <= 0 {
		return errors.New("invalid actor")
	}
	res, err := s.db.ExecContext(ctx,
		`UPDATE backup_restore_requests SET status = ?, rejection_reason = ?
		 WHERE id = ? AND status = ? AND requested_by_user_id <> ?`,
		BackupRestoreStatusRejected, rejectionReason, id, BackupRestoreStatusPending, actorUserID,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		row, err := s.GetBackupRestoreRequest(ctx, id)
		if err != nil {
			return err
		}
		if row.RequestedByUserID == actorUserID {
			return ErrBackupRestoreCannotApprove
		}
		return ErrBackupRestoreInvalidState
	}
	return nil
}

// CancelBackupRestoreRequest allows the requester to cancel a pending request.
func (s *Store) CancelBackupRestoreRequest(ctx context.Context, id int64, requesterUserID int) error {
	if s == nil || s.db == nil {
		return errors.New("nil db")
	}
	res, err := s.db.ExecContext(ctx,
		`UPDATE backup_restore_requests SET status = ?
		 WHERE id = ? AND status = ? AND requested_by_user_id = ?`,
		BackupRestoreStatusCancelled, id, BackupRestoreStatusPending, requesterUserID,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		row, gerr := s.GetBackupRestoreRequest(ctx, id)
		if gerr != nil {
			return gerr
		}
		if row.RequestedByUserID != requesterUserID {
			return ErrBackupRestoreNotRequester
		}
		return ErrBackupRestoreInvalidState
	}
	return nil
}

// FulfillBackupRestoreRequest marks an approved request fulfilled after operators run restore out of band.
func (s *Store) FulfillBackupRestoreRequest(ctx context.Context, id int64, actorUserID int, note string) error {
	if s == nil || s.db == nil {
		return errors.New("nil db")
	}
	res, err := s.db.ExecContext(ctx,
		`UPDATE backup_restore_requests SET status = ?, fulfilled_at = UTC_TIMESTAMP(), fulfilled_note = ?
		 WHERE id = ? AND status = ?`,
		BackupRestoreStatusFulfilled, note, id, BackupRestoreStatusApproved,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrBackupRestoreInvalidState
	}
	_ = actorUserID // reserved for future audit column
	return nil
}
