package chat

import (
	"context"
	"net/http"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/store"
)

func newChatServiceForTest(t *testing.T) (*Service, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error: %v", err)
	}

	return NewService(store.New(db)), mock, func() { _ = db.Close() }
}

func conversationRows(now time.Time, disabledBy any, disabledAt any) *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "user1_id", "user2_id", "user1_last_read_at", "user2_last_read_at",
		"friendship_disabled_at", "friendship_disabled_by_user_id", "created_at",
	}).AddRow(int64(10), int64(1), int64(2), nil, nil, disabledAt, disabledBy, now)
}

func TestServiceSendMessageRejectsDisabledConversationForRemover(t *testing.T) {
	service, mock, cleanup := newChatServiceForTest(t)
	defer cleanup()

	now := time.Now()
	mock.ExpectQuery(`(?s)SELECT id, user1_id, user2_id, user1_last_read_at, user2_last_read_at,.*friendship_disabled_at, friendship_disabled_by_user_id, created_at.*FROM conversations WHERE id = \$1`).
		WithArgs(int64(10)).
		WillReturnRows(conversationRows(now, int64(1), now))

	_, err := service.SendMessage(context.Background(), 1, 10, "hello", "text")
	if err == nil {
		t.Fatal("SendMessage() expected disabled conversation error")
	}

	appErr := httpapi.AsAppError(err)
	if appErr.HTTPStatus != http.StatusConflict || appErr.Message != "你已经删除对方好友" {
		t.Fatalf("SendMessage() error = %#v, want conflict with remover message", appErr)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestServiceSendMessageRejectsDisabledConversationForPeer(t *testing.T) {
	service, mock, cleanup := newChatServiceForTest(t)
	defer cleanup()

	now := time.Now()
	mock.ExpectQuery(`(?s)SELECT id, user1_id, user2_id, user1_last_read_at, user2_last_read_at,.*friendship_disabled_at, friendship_disabled_by_user_id, created_at.*FROM conversations WHERE id = \$1`).
		WithArgs(int64(10)).
		WillReturnRows(conversationRows(now, int64(1), now))

	_, err := service.SendMessage(context.Background(), 2, 10, "hello", "text")
	if err == nil {
		t.Fatal("SendMessage() expected disabled conversation error")
	}

	appErr := httpapi.AsAppError(err)
	if appErr.HTTPStatus != http.StatusConflict || appErr.Message != "对方已经将你删除好友" {
		t.Fatalf("SendMessage() error = %#v, want conflict with peer message", appErr)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
