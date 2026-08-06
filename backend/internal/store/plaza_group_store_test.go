package store

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestExpirePlazaGroupInvitesArchivesChatAndClosesOpenInvites(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sql mock: %v", err)
	}
	defer db.Close()

	now := time.Date(2026, time.August, 7, 12, 0, 0, 0, time.UTC)
	mock.ExpectExec(`WITH archived AS \(`).
		WithArgs(now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE plaza_group_invites`).
		WithArgs(now).
		WillReturnResult(sqlmock.NewResult(0, 2))

	if err := New(db).ExpirePlazaGroupInvites(context.Background(), nil, now); err != nil {
		t.Fatalf("expire plaza groups: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectation: %v", err)
	}
}
