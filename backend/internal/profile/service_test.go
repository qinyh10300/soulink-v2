package profile

import (
	"context"
	"errors"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"soulink-backend/internal/store"
)

func newProfileServiceForTest(t *testing.T) (*Service, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error: %v", err)
	}

	return NewService(store.New(db)), mock, func() { _ = db.Close() }
}

func userRows(now time.Time, contactType, contactValue any) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "email", "school_key", "password_hash", "nickname", "department", "grade", "contact_type", "contact_value", "has_completed_questionnaire", "created_at", "updated_at", "deleted_at"}).
		AddRow(int64(1), "test@mails.tsinghua.edu.cn", "thu", "hash", "小张", "计算机系", "本科三年级", contactType, contactValue, false, now, now, nil)
}

func TestServiceUpdateRejectsIncompleteContact(t *testing.T) {
	service, _, cleanup := newProfileServiceForTest(t)
	defer cleanup()

	_, err := service.Update(context.Background(), 1, UpdateProfileInput{
		Nickname:     "小张",
		Department:   "计算机系",
		Grade:        "本科三年级",
		ContactType:  "wechat",
		ContactValue: "",
	})
	if err == nil {
		t.Fatal("Update() expected validation error for incomplete contact info")
	}
}

func TestServiceUpdateSuccessKeepsContacts(t *testing.T) {
	service, mock, cleanup := newProfileServiceForTest(t)
	defer cleanup()

	now := time.Now()
	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)SELECT id, email, school_key, password_hash, nickname, department, grade, contact_type, contact_value,.*FROM users.*WHERE id = \$1.*LIMIT 1`).
		WithArgs(int64(1)).
		WillReturnRows(userRows(now, "wechat", "zhangsan"))
	mock.ExpectQuery(`(?s)UPDATE users.*RETURNING id, email, school_key, password_hash, nickname, department, grade, contact_type, contact_value,.*deleted_at`).
		WithArgs(int64(1), "小张", "计算机系", "本科三年级", "wechat", "zhangsan", sqlmock.AnyArg()).
		WillReturnRows(userRows(now, "wechat", "zhangsan"))
	mock.ExpectQuery(`(?s)SELECT id, email, school_key, password_hash, nickname, department, grade, contact_type, contact_value,.*FROM users.*WHERE id = \$1.*LIMIT 1`).
		WithArgs(int64(1)).
		WillReturnRows(userRows(now, "wechat", "zhangsan"))
	mock.ExpectQuery(`(?s)SELECT user_id, contact_type, contact_value, created_at, updated_at.*FROM user_contact_methods.*WHERE user_id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "contact_type", "contact_value", "created_at", "updated_at"}).
			AddRow(int64(1), "wechat", "zhangsan", now, now))
	mock.ExpectCommit()

	details, err := service.Update(context.Background(), 1, UpdateProfileInput{
		Nickname:   "小张",
		Department: "计算机系",
		Grade:      "本科三年级",
	})
	if err != nil {
		t.Fatalf("Update() unexpected error: %v", err)
	}
	if len(details.Contacts) != 1 || details.Contacts[0].ContactType != "wechat" {
		t.Fatalf("Update() returned unexpected contacts: %#v", details.Contacts)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestServiceUpdateWithContactReplacesPreviousContact(t *testing.T) {
	service, mock, cleanup := newProfileServiceForTest(t)
	defer cleanup()

	now := time.Now()
	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)SELECT id, email, school_key, password_hash, nickname, department, grade, contact_type, contact_value,.*FROM users.*WHERE id = \$1.*LIMIT 1`).
		WithArgs(int64(1)).
		WillReturnRows(userRows(now, "wechat", "old_wechat"))
	mock.ExpectQuery(`(?s)UPDATE users.*RETURNING id, email, school_key, password_hash, nickname, department, grade, contact_type, contact_value,.*deleted_at`).
		WithArgs(int64(1), "小张", "计算机系", "本科三年级", "wechat", "old_wechat", sqlmock.AnyArg()).
		WillReturnRows(userRows(now, "wechat", "old_wechat"))
	mock.ExpectExec(`DELETE FROM user_contact_methods WHERE user_id = \$1`).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`(?s)INSERT INTO user_contact_methods .* ON CONFLICT .* RETURNING user_id, contact_type, contact_value, created_at, updated_at`).
		WithArgs(int64(1), "phone", "13800000000", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "contact_type", "contact_value", "created_at", "updated_at"}).
			AddRow(int64(1), "phone", "13800000000", now, now))
	mock.ExpectQuery(`(?s)SELECT contact_type, contact_value.*FROM user_contact_methods.*WHERE user_id = \$1.*LIMIT 1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"contact_type", "contact_value"}).
			AddRow("phone", "13800000000"))
	mock.ExpectExec(`(?s)UPDATE users.*SET contact_type = \$2,.*contact_value = \$3,.*updated_at = \$4.*WHERE id = \$1`).
		WithArgs(int64(1), "phone", "13800000000", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`(?s)SELECT id, email, school_key, password_hash, nickname, department, grade, contact_type, contact_value,.*FROM users.*WHERE id = \$1.*LIMIT 1`).
		WithArgs(int64(1)).
		WillReturnRows(userRows(now, "phone", "13800000000"))
	mock.ExpectQuery(`(?s)SELECT user_id, contact_type, contact_value, created_at, updated_at.*FROM user_contact_methods.*WHERE user_id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "contact_type", "contact_value", "created_at", "updated_at"}).
			AddRow(int64(1), "phone", "13800000000", now, now))
	mock.ExpectCommit()

	details, err := service.Update(context.Background(), 1, UpdateProfileInput{
		Nickname:     "小张",
		Department:   "计算机系",
		Grade:        "本科三年级",
		ContactType:  "phone",
		ContactValue: "13800000000",
	})
	if err != nil {
		t.Fatalf("Update() unexpected error: %v", err)
	}
	if details.User.ContactType != "phone" || details.User.ContactValue != "13800000000" {
		t.Fatalf("Update() returned unexpected summary contact: %#v", details.User)
	}
	if len(details.Contacts) != 1 || details.Contacts[0].ContactType != "phone" {
		t.Fatalf("Update() returned unexpected contacts: %#v", details.Contacts)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestServiceUpsertContactSuccessReplacesPreviousContact(t *testing.T) {
	service, mock, cleanup := newProfileServiceForTest(t)
	defer cleanup()

	now := time.Now()
	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)SELECT id, email, school_key, password_hash, nickname, department, grade, contact_type, contact_value,.*FROM users.*WHERE id = \$1.*LIMIT 1`).
		WithArgs(int64(1)).
		WillReturnRows(userRows(now, "wechat", "zhangsan"))
	mock.ExpectExec(`DELETE FROM user_contact_methods WHERE user_id = \$1`).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectQuery(`(?s)INSERT INTO user_contact_methods .* ON CONFLICT .* RETURNING user_id, contact_type, contact_value, created_at, updated_at`).
		WithArgs(int64(1), "phone", "13800000000", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "contact_type", "contact_value", "created_at", "updated_at"}).
			AddRow(int64(1), "phone", "13800000000", now, now))
	mock.ExpectQuery(`(?s)SELECT contact_type, contact_value.*FROM user_contact_methods.*WHERE user_id = \$1.*LIMIT 1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"contact_type", "contact_value"}).
			AddRow("phone", "13800000000"))
	mock.ExpectExec(`(?s)UPDATE users.*SET contact_type = \$2,.*contact_value = \$3,.*updated_at = \$4.*WHERE id = \$1`).
		WithArgs(int64(1), "phone", "13800000000", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`(?s)SELECT id, email, school_key, password_hash, nickname, department, grade, contact_type, contact_value,.*FROM users.*WHERE id = \$1.*LIMIT 1`).
		WithArgs(int64(1)).
		WillReturnRows(userRows(now, "phone", "13800000000"))
	mock.ExpectQuery(`(?s)SELECT user_id, contact_type, contact_value, created_at, updated_at.*FROM user_contact_methods.*WHERE user_id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "contact_type", "contact_value", "created_at", "updated_at"}).
			AddRow(int64(1), "phone", "13800000000", now, now))
	mock.ExpectCommit()

	details, err := service.UpsertContact(context.Background(), 1, UpsertContactInput{
		ContactType:  "phone",
		ContactValue: "13800000000",
	})
	if err != nil {
		t.Fatalf("UpsertContact() unexpected error: %v", err)
	}
	if details.User.ContactType != "phone" || details.User.ContactValue != "13800000000" {
		t.Fatalf("UpsertContact() returned unexpected summary contact: %#v", details.User)
	}
	if len(details.Contacts) != 1 || details.Contacts[0].ContactType != "phone" {
		t.Fatalf("UpsertContact() returned unexpected contacts: %#v", details.Contacts)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestServiceDeleteContactSuccess(t *testing.T) {
	service, mock, cleanup := newProfileServiceForTest(t)
	defer cleanup()

	now := time.Now()
	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)SELECT id, email, school_key, password_hash, nickname, department, grade, contact_type, contact_value,.*FROM users.*WHERE id = \$1.*LIMIT 1`).
		WithArgs(int64(1)).
		WillReturnRows(userRows(now, "wechat", "zhangsan"))
	mock.ExpectExec(`(?s)DELETE FROM user_contact_methods.*WHERE user_id = \$1 AND contact_type = \$2`).
		WithArgs(int64(1), "wechat").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`(?s)SELECT contact_type, contact_value.*FROM user_contact_methods.*WHERE user_id = \$1.*LIMIT 1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"contact_type", "contact_value"}))
	mock.ExpectExec(`(?s)UPDATE users.*SET contact_type = \$2,.*contact_value = \$3,.*updated_at = \$4.*WHERE id = \$1`).
		WithArgs(int64(1), nil, nil, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`(?s)SELECT id, email, school_key, password_hash, nickname, department, grade, contact_type, contact_value,.*FROM users.*WHERE id = \$1.*LIMIT 1`).
		WithArgs(int64(1)).
		WillReturnRows(userRows(now, nil, nil))
	mock.ExpectQuery(`(?s)SELECT user_id, contact_type, contact_value, created_at, updated_at.*FROM user_contact_methods.*WHERE user_id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "contact_type", "contact_value", "created_at", "updated_at"}))
	mock.ExpectCommit()

	details, err := service.DeleteContact(context.Background(), 1, "wechat")
	if err != nil {
		t.Fatalf("DeleteContact() unexpected error: %v", err)
	}
	if details.User.ContactType != "" || details.User.ContactValue != "" {
		t.Fatalf("DeleteContact() should clear summary contact: %#v", details.User)
	}
	if len(details.Contacts) != 0 {
		t.Fatalf("DeleteContact() returned unexpected contacts: %#v", details.Contacts)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestServiceUpdateRejectsDuplicateNickname(t *testing.T) {
	service, mock, cleanup := newProfileServiceForTest(t)
	defer cleanup()

	now := time.Now()
	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)SELECT id, email, school_key, password_hash, nickname, department, grade, contact_type, contact_value,.*FROM users.*WHERE id = \$1.*LIMIT 1`).
		WithArgs(int64(1)).
		WillReturnRows(userRows(now, "wechat", "zhangsan"))
	mock.ExpectQuery(`(?s)UPDATE users.*RETURNING id, email, school_key, password_hash, nickname, department, grade, contact_type, contact_value,.*deleted_at`).
		WithArgs(int64(1), "duplicate-name", "计算机系", "本科三年级", "wechat", "zhangsan", sqlmock.AnyArg()).
		WillReturnError(errors.New(`duplicate key value violates unique constraint "ux_users_nickname_active"`))
	mock.ExpectRollback()

	_, err := service.Update(context.Background(), 1, UpdateProfileInput{
		Nickname:   "duplicate-name",
		Department: "计算机系",
		Grade:      "本科三年级",
	})
	if err == nil {
		t.Fatal("Update() expected duplicate nickname error")
	}
	if err.Error() != "nickname already exists" {
		t.Fatalf("Update() error = %v, want duplicate nickname error", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestGetPrivacyDefaultsToPrivate(t *testing.T) {
	service, mock, cleanup := newProfileServiceForTest(t)
	defer cleanup()

	mock.ExpectQuery(`(?s)SELECT s.user_id, s.profile_public, s.schedule_public, s.preferences_public.*FROM user_profile_settings s.*WHERE s.user_id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"user_id", "profile_public", "schedule_public", "preferences_public",
			"avatar_media_id", "asset_key", "updated_at",
		}))

	privacy, err := service.GetPrivacy(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetPrivacy() error = %v", err)
	}
	if privacy.ProfilePublic || privacy.SchedulePublic || privacy.PreferencesPublic || privacy.AvatarURL != "" {
		t.Fatalf("GetPrivacy() = %#v, want private defaults", privacy)
	}
}

func TestUpdatePrivacyClosesDependentSharingWhenProfileIsPrivate(t *testing.T) {
	service, mock, cleanup := newProfileServiceForTest(t)
	defer cleanup()
	now := time.Now()

	mock.ExpectQuery(`(?s)SELECT id, email, school_key, password_hash, nickname, department, grade, contact_type, contact_value,.*FROM users.*WHERE id = \$1.*LIMIT 1`).
		WithArgs(int64(1)).
		WillReturnRows(userRows(now, nil, nil))
	mock.ExpectQuery(`(?s)INSERT INTO user_profile_settings.*ON CONFLICT.*RETURNING user_id, profile_public, schedule_public, preferences_public`).
		WithArgs(int64(1), false, false, false, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"user_id", "profile_public", "schedule_public", "preferences_public", "avatar_media_id", "updated_at",
		}).AddRow(int64(1), false, false, false, nil, now))

	privacy, err := service.UpdatePrivacy(context.Background(), 1, PrivacySettings{
		ProfilePublic:     false,
		SchedulePublic:    true,
		PreferencesPublic: true,
	})
	if err != nil {
		t.Fatalf("UpdatePrivacy() error = %v", err)
	}
	if privacy.SchedulePublic || privacy.PreferencesPublic {
		t.Fatalf("UpdatePrivacy() = %#v, dependent settings should be private", privacy)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
