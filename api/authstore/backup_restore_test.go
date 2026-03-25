package authstore

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestCreateBackupRestoreRequest(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	s := New(db, time.Hour, time.Hour)

	mock.ExpectExec(`INSERT INTO backup_restore_requests`).
		WithArgs(5, "platform_mysql", "snap-2026-03-01", "good reason here", BackupRestoreStatusPending, nil).
		WillReturnResult(sqlmock.NewResult(42, 1))

	id, err := s.CreateBackupRestoreRequest(context.Background(), 5, "platform_mysql", "snap-2026-03-01", "good reason here", "")
	if err != nil {
		t.Fatal(err)
	}
	if id != 42 {
		t.Fatalf("id %d want 42", id)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestApproveBackupRestoreRequestFirstSlot(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	s := New(db, time.Hour, time.Hour)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT requested_by_user_id, status, approval_1_user_id, approval_2_user_id`).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"requested_by_user_id", "status", "approval_1_user_id", "approval_2_user_id"}).
			AddRow(3, BackupRestoreStatusPending, nil, nil))
	mock.ExpectExec(`UPDATE backup_restore_requests SET approval_1_user_id`).
		WithArgs(9, sqlmock.AnyArg(), int64(7), BackupRestoreStatusPending).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := s.ApproveBackupRestoreRequest(context.Background(), 7, 9); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestApproveBackupRestoreRequestRequesterForbidden(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	s := New(db, time.Hour, time.Hour)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT requested_by_user_id, status, approval_1_user_id, approval_2_user_id`).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"requested_by_user_id", "status", "approval_1_user_id", "approval_2_user_id"}).
			AddRow(9, BackupRestoreStatusPending, nil, nil))
	mock.ExpectRollback()

	err = s.ApproveBackupRestoreRequest(context.Background(), 7, 9)
	if err != ErrBackupRestoreCannotApprove {
		t.Fatalf("got %v want ErrBackupRestoreCannotApprove", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
