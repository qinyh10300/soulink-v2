package friend

import (
	"context"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"soulink-backend/internal/store"
)

func newFriendServiceForTest(t *testing.T) (*Service, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error: %v", err)
	}

	return NewService(store.New(db), nil), mock, func() { _ = db.Close() }
}

func friendUserRows(now time.Time, userID int64, nickname string) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "email", "school_key", "password_hash", "nickname", "department", "grade", "contact_type", "contact_value", "has_completed_questionnaire", "created_at", "updated_at", "deleted_at"}).
		AddRow(userID, "friend@example.com", "thu", "hash", nickname, "dept", "grade", nil, nil, false, now, now, nil)
}

func friendshipRows(now time.Time, friendshipID, requesterID, addresseeID int64, status string) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "requester_id", "addressee_id", "status", "created_at", "updated_at"}).
		AddRow(friendshipID, requesterID, addresseeID, status, now, now)
}

func conversationRows(now time.Time, user1ID, user2ID, disabledBy int64) *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "user1_id", "user2_id", "user1_last_read_at", "user2_last_read_at",
		"friendship_disabled_at", "friendship_disabled_by_user_id", "created_at",
	}).AddRow(int64(20), user1ID, user2ID, nil, nil, now, disabledBy, now)
}

func TestServiceSendRequestByNicknameSuccess(t *testing.T) {
	service, mock, cleanup := newFriendServiceForTest(t)
	defer cleanup()

	now := time.Now()
	mock.ExpectQuery(`(?s)SELECT id, email, school_key, password_hash, nickname, department, grade, contact_type, contact_value,.*FROM users.*WHERE LOWER\(nickname\) = LOWER\(\$1\).*LIMIT 1`).
		WithArgs("target-name").
		WillReturnRows(friendUserRows(now, 2, "target-name"))
	mock.ExpectQuery(`(?s)SELECT id, requester_id, addressee_id, status, created_at, updated_at.*FROM friendships.*LIMIT 1`).
		WithArgs(int64(1), int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "requester_id", "addressee_id", "status", "created_at", "updated_at"}))
	mock.ExpectQuery(`(?s)INSERT INTO friendships .*RETURNING id, requester_id, addressee_id, status, created_at, updated_at`).
		WithArgs(int64(1), int64(2), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "requester_id", "addressee_id", "status", "created_at", "updated_at"}).
			AddRow(int64(10), int64(1), int64(2), "pending", now, now))

	friendship, err := service.SendRequest(context.Background(), 1, "target-name")
	if err != nil {
		t.Fatalf("SendRequest() unexpected error: %v", err)
	}
	if friendship.ID != 10 {
		t.Fatalf("SendRequest() friendship id = %d, want 10", friendship.ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestServiceSendRequestRejectsSelfByNickname(t *testing.T) {
	service, mock, cleanup := newFriendServiceForTest(t)
	defer cleanup()

	now := time.Now()
	mock.ExpectQuery(`(?s)SELECT id, email, school_key, password_hash, nickname, department, grade, contact_type, contact_value,.*FROM users.*WHERE LOWER\(nickname\) = LOWER\(\$1\).*LIMIT 1`).
		WithArgs("self-name").
		WillReturnRows(friendUserRows(now, 1, "self-name"))

	_, err := service.SendRequest(context.Background(), 1, "self-name")
	if err == nil {
		t.Fatal("SendRequest() expected self-add error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestServiceRemoveFriendDisablesConversationAndNotifiesBothSides(t *testing.T) {
	service, mock, cleanup := newFriendServiceForTest(t)
	defer cleanup()

	notified := make([]struct {
		userID int64
		event  string
	}, 0, 2)
	service.notify = func(userID int64, eventType string, data any) {
		notified = append(notified, struct {
			userID int64
			event  string
		}{userID: userID, event: eventType})
	}

	now := time.Now()
	mock.ExpectQuery(`(?s)SELECT id, requester_id, addressee_id, status, created_at, updated_at.*FROM friendships.*WHERE id = \$1`).
		WithArgs(int64(9)).
		WillReturnRows(friendshipRows(now, 9, 1, 2, "accepted"))
	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)UPDATE conversations.*SET friendship_disabled_at = \$4,.*friendship_disabled_by_user_id = \$3.*RETURNING id, user1_id, user2_id, user1_last_read_at, user2_last_read_at,.*created_at`).
		WithArgs(int64(1), int64(2), int64(1), sqlmock.AnyArg()).
		WillReturnRows(conversationRows(now, 1, 2, 1))
	mock.ExpectExec(`DELETE FROM friendships WHERE id = \$1`).
		WithArgs(int64(9)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := service.RemoveFriend(context.Background(), 1, 9); err != nil {
		t.Fatalf("RemoveFriend() unexpected error: %v", err)
	}
	if len(notified) != 2 {
		t.Fatalf("notifications = %d, want 2", len(notified))
	}
	if notified[0].event != "conversation_updated" || notified[1].event != "conversation_updated" {
		t.Fatalf("unexpected notifications: %#v", notified)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
