package match

import (
	"context"
	"database/sql"
	"encoding/json"
	"math/rand"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"soulink-backend/internal/store"
)

func TestBuildResultViewMatched(t *testing.T) {
	contactsJSON, err := json.Marshal([]partnerContactSnapshot{{Type: "wechat", Value: "alice_wechat"}, {Type: "phone", Value: "13800000000"}})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	view, err := buildResultView(store.MatchResultRecord{
		MatchResult: store.MatchResult{
			ID:                  101,
			CycleID:             9,
			ResultStatus:        ResultStatusMatched,
			PartnerUserID:       sql.NullInt64{Int64: 2, Valid: true},
			MatchScore:          sql.NullFloat64{Float64: 92.5, Valid: true},
			MatchReason:         "你们在生活方式和价值观上高度一致",
			ResultMessage:       "匹配成功",
			PartnerNickname:     "Alice",
			PartnerDepartment:   "计算机学院",
			PartnerGrade:        "本科三年级",
			PartnerContactsJSON: contactsJSON,
		},
		CycleKey:      "2026-w13",
		CycleTitle:    "2026 第 13 周匹配",
		RevealStatus:  "revealed",
		CycleRevealAt: time.Date(2026, 3, 26, 20, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("buildResultView() error = %v", err)
	}

	if view.ResultStatus != ResultStatusMatched {
		t.Fatalf("view.ResultStatus = %q, want %q", view.ResultStatus, ResultStatusMatched)
	}
	if view.Score == nil || *view.Score != 92.5 {
		t.Fatalf("view.Score = %#v, want 92.5", view.Score)
	}
	if view.Partner == nil {
		t.Fatal("view.Partner is nil")
	}
	if len(view.Partner.Contacts) != 2 {
		t.Fatalf("len(view.Partner.Contacts) = %d, want 2", len(view.Partner.Contacts))
	}
	if view.Partner.Contacts[0].Label != "微信" {
		t.Fatalf("first contact label = %q, want %q", view.Partner.Contacts[0].Label, "微信")
	}
}

func TestContactMatchedPartnerCreatesPendingFriendRequestWhenNotFriends(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 4, 2, 20, 0, 0, 0, time.UTC)
	service := NewService(store.New(db), WithNowFunc(func() time.Time { return now }))

	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)SELECT r.id, r.cycle_id, r.user_id, r.result_status, r.partner_user_id.*WHERE r.user_id = \$1 AND r.id = \$2`).
		WithArgs(int64(1), int64(10)).
		WillReturnRows(matchResultRecordRows(now, 10, 1, 2, ResultStatusMatched))
	mock.ExpectQuery(`(?s)SELECT id, email, school_key, password_hash, nickname, department, grade, contact_type, contact_value,.*FROM users.*WHERE id = \$1 AND deleted_at IS NULL`).
		WithArgs(int64(2)).
		WillReturnRows(userRows(now, 2, "partner@example.com", "partner"))
	mock.ExpectQuery(`(?s)SELECT id, requester_id, addressee_id, status, created_at, updated_at.*FROM friendships.*LIMIT 1`).
		WithArgs(int64(1), int64(2)).
		WillReturnRows(emptyFriendshipRows())
	mock.ExpectQuery(`(?s)INSERT INTO friendships .*VALUES \(\$1, \$2, 'pending', \$3, \$3\).*RETURNING id, requester_id, addressee_id, status, created_at, updated_at`).
		WithArgs(int64(1), int64(2), now).
		WillReturnRows(friendshipRows(now, 5, 1, 2, "pending"))
	mock.ExpectCommit()

	result, err := service.ContactMatchedPartner(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("ContactMatchedPartner() error = %v", err)
	}
	if result.Action != ContactActionRequestSent || result.FriendshipID != 5 || result.ConversationID != 0 {
		t.Fatalf("ContactMatchedPartner() = %#v, want action=request_sent friendship=5", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestContactMatchedPartnerOpensChatWhenAlreadyFriends(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 4, 2, 20, 0, 0, 0, time.UTC)
	service := NewService(store.New(db), WithNowFunc(func() time.Time { return now }))

	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)SELECT r.id, r.cycle_id, r.user_id, r.result_status, r.partner_user_id.*WHERE r.user_id = \$1 AND r.id = \$2`).
		WithArgs(int64(1), int64(10)).
		WillReturnRows(matchResultRecordRows(now, 10, 1, 2, ResultStatusMatched))
	mock.ExpectQuery(`(?s)SELECT id, email, school_key, password_hash, nickname, department, grade, contact_type, contact_value,.*FROM users.*WHERE id = \$1 AND deleted_at IS NULL`).
		WithArgs(int64(2)).
		WillReturnRows(userRows(now, 2, "partner@example.com", "partner"))
	mock.ExpectQuery(`(?s)SELECT id, requester_id, addressee_id, status, created_at, updated_at.*FROM friendships.*LIMIT 1`).
		WithArgs(int64(1), int64(2)).
		WillReturnRows(friendshipRows(now, 5, 1, 2, "accepted"))
	mock.ExpectQuery(`(?s)INSERT INTO conversations .*VALUES \(\$1, \$2, \$3\).*RETURNING id, user1_id, user2_id, user1_last_read_at, user2_last_read_at,.*created_at`).
		WithArgs(int64(1), int64(2), now).
		WillReturnRows(conversationRows(now, 21, 1, 2))
	mock.ExpectCommit()

	result, err := service.ContactMatchedPartner(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("ContactMatchedPartner() error = %v", err)
	}
	if result.Action != ContactActionChatReady || result.FriendshipID != 5 || result.ConversationID != 21 {
		t.Fatalf("ContactMatchedPartner() = %#v, want action=chat_ready friendship=5 conversation=21", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestBuildResultViewFailed(t *testing.T) {
	view, err := buildResultView(store.MatchResultRecord{
		MatchResult: store.MatchResult{
			ID:            102,
			CycleID:       9,
			ResultStatus:  ResultStatusFailed,
			MatchReason:   "这条说明不应该暴露",
			ResultMessage: "本轮未匹配成功",
		},
		CycleKey:      "2026-w13",
		CycleTitle:    "2026 第 13 周匹配",
		RevealStatus:  "revealed",
		CycleRevealAt: time.Date(2026, 3, 26, 20, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("buildResultView() error = %v", err)
	}

	if view.ResultStatus != ResultStatusFailed {
		t.Fatalf("view.ResultStatus = %q, want %q", view.ResultStatus, ResultStatusFailed)
	}
	if view.Score != nil {
		t.Fatalf("view.Score = %#v, want nil", view.Score)
	}
	if view.Reason != "" {
		t.Fatalf("view.Reason = %q, want empty", view.Reason)
	}
	if view.Partner != nil {
		t.Fatalf("view.Partner = %#v, want nil", view.Partner)
	}
}

func TestToPartnerContactSnapshotsFallsBackToEmail(t *testing.T) {
	snapshots := toPartnerContactSnapshots(store.User{Email: "partner@example.com"}, nil)
	if len(snapshots) != 1 {
		t.Fatalf("len(snapshots) = %d, want 1", len(snapshots))
	}
	if snapshots[0].Type != "email" || snapshots[0].Value != "partner@example.com" {
		t.Fatalf("unexpected fallback snapshot: %#v", snapshots[0])
	}
}

func TestToPartnerContactSnapshotsUsesLatestSavedContact(t *testing.T) {
	now := time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC)
	snapshots := toPartnerContactSnapshots(store.User{Email: "partner@example.com"}, []store.ContactMethod{
		{ContactType: "wechat", ContactValue: "wx_old", UpdatedAt: now.Add(-time.Hour)},
		{ContactType: "phone", ContactValue: "13800000000", UpdatedAt: now},
	})
	if len(snapshots) != 1 {
		t.Fatalf("len(snapshots) = %d, want 1", len(snapshots))
	}
	if snapshots[0].Type != "phone" || snapshots[0].Value != "13800000000" {
		t.Fatalf("unexpected latest snapshot: %#v", snapshots[0])
	}
}

func TestBuildSimpleVersionResultsLeavesOddUserFailed(t *testing.T) {
	rng := rand.New(rand.NewSource(20260325))
	results := buildSimpleVersionResults([]store.User{{ID: 101}, {ID: 102}, {ID: 103}}, "v0.1", rng)
	if len(results) != 3 {
		t.Fatalf("len(results) = %d, want 3", len(results))
	}

	matched := make(map[int64]RecordUserResultInput)
	failedCount := 0
	for _, result := range results {
		switch result.ResultStatus {
		case ResultStatusMatched:
			matched[result.UserID] = result
			if result.PartnerUserID == nil {
				t.Fatalf("matched user %d partnerUserID is nil", result.UserID)
			}
			if *result.PartnerUserID == result.UserID {
				t.Fatalf("matched user %d paired with self", result.UserID)
			}
			if result.MatchScore == nil {
				t.Fatalf("matched user %d score is nil", result.UserID)
			}
			if !strings.Contains(result.MatchReason, "v0.1") {
				t.Fatalf("matched user %d reason = %q, want version mention", result.UserID, result.MatchReason)
			}
		case ResultStatusFailed:
			failedCount++
			if result.PartnerUserID != nil {
				t.Fatalf("failed user %d partnerUserID = %#v, want nil", result.UserID, result.PartnerUserID)
			}
			if result.MatchScore != nil {
				t.Fatalf("failed user %d matchScore = %#v, want nil", result.UserID, result.MatchScore)
			}
		default:
			t.Fatalf("unexpected result status %q", result.ResultStatus)
		}
	}

	if len(matched) != 2 {
		t.Fatalf("matched user count = %d, want 2", len(matched))
	}
	if failedCount != 1 {
		t.Fatalf("failed count = %d, want 1", failedCount)
	}

	for userID, result := range matched {
		partnerResult, ok := matched[*result.PartnerUserID]
		if !ok {
			t.Fatalf("partner result for user %d not found", userID)
		}
		if partnerResult.PartnerUserID == nil || *partnerResult.PartnerUserID != userID {
			t.Fatalf("partner result for user %d is not symmetric: %#v", userID, partnerResult)
		}
	}
}

func matchResultRecordRows(now time.Time, matchID, userID, partnerID int64, status string) *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "cycle_id", "user_id", "result_status", "partner_user_id", "match_score",
		"match_reason", "match_reason_json", "result_message", "partner_nickname", "partner_department",
		"partner_grade", "partner_contacts_json", "revealed_at", "created_at", "updated_at",
		"cycle_key", "title", "questionnaire_key", "questionnaire_version", "reveal_status", "reveal_at",
		"school_key", "email",
	}).AddRow(
		matchID, int64(3), userID, status, partnerID, 88.0,
		"", []byte(`{}`), "matched", "partner", "dept",
		"grade", []byte(`[]`), now, now, now,
		"2026-w14", "2026 W14", "questionnaire", "v1", "revealed", now,
		"thu", "partner@example.com",
	)
}

func userRows(now time.Time, userID int64, email, nickname string) *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "email", "school_key", "password_hash", "nickname", "department", "grade",
		"contact_type", "contact_value", "has_completed_questionnaire", "created_at", "updated_at", "deleted_at",
	}).AddRow(userID, email, "thu", "hash", nickname, "dept", "grade", nil, nil, false, now, now, nil)
}

func emptyFriendshipRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "requester_id", "addressee_id", "status", "created_at", "updated_at"})
}

func friendshipRows(now time.Time, friendshipID, requesterID, addresseeID int64, status string) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "requester_id", "addressee_id", "status", "created_at", "updated_at"}).
		AddRow(friendshipID, requesterID, addresseeID, status, now, now)
}

func conversationRows(now time.Time, conversationID, user1ID, user2ID int64) *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "user1_id", "user2_id", "user1_last_read_at", "user2_last_read_at",
		"friendship_disabled_at", "friendship_disabled_by_user_id", "created_at",
	}).AddRow(conversationID, user1ID, user2ID, nil, nil, nil, nil, now)
}
