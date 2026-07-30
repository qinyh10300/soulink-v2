package profile

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/meta"
	"soulink-backend/internal/store"
)

const maxContactValueLength = 255

type Service struct {
	store *store.Store
}

type UpdateProfileInput struct {
	Nickname     string
	Department   string
	Grade        string
	ContactType  string
	ContactValue string
}

type UpsertContactInput struct {
	ContactType  string
	ContactValue string
}

type Details struct {
	User     store.User
	Contacts []store.ContactMethod
}

type PublicDetails struct {
	User store.User
}

// NewService 创建个人资料服务实例。
func NewService(dataStore *store.Store) *Service {
	return &Service{store: dataStore}
}

// Get 返回当前用户的完整资料及联系方式列表。
func (s *Service) Get(ctx context.Context, userID int64) (Details, error) {
	if err := s.ensureReady(); err != nil {
		return Details{}, err
	}
	return s.loadDetails(ctx, nil, userID)
}

// GetPublic returns the fields that can be shown on another user's public profile.
func (s *Service) GetPublic(ctx context.Context, userID int64) (PublicDetails, error) {
	if err := s.ensureReady(); err != nil {
		return PublicDetails{}, err
	}
	user, err := s.store.GetUserByID(ctx, nil, userID)
	if err != nil {
		return PublicDetails{}, mapUserStoreError(err)
	}
	return PublicDetails{User: user}, nil
}

// Update 校验并更新当前用户的个人资料。
func (s *Service) Update(ctx context.Context, userID int64, input UpdateProfileInput) (Details, error) {
	if err := s.ensureReady(); err != nil {
		return Details{}, err
	}

	input.Nickname = strings.TrimSpace(input.Nickname)
	input.Department = strings.TrimSpace(input.Department)
	input.Grade = strings.TrimSpace(input.Grade)
	input.ContactType = strings.TrimSpace(input.ContactType)
	input.ContactValue = strings.TrimSpace(input.ContactValue)

	if input.Nickname == "" {
		return Details{}, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
	}
	if (input.ContactType == "") != (input.ContactValue == "") {
		return Details{}, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
	}
	if input.ContactType != "" {
		if err := validateContact(input.ContactType, input.ContactValue); err != nil {
			return Details{}, err
		}
	}

	var details Details
	now := time.Now()
	err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		currentUser, err := s.store.GetUserByID(ctx, tx, userID)
		if err != nil {
			return mapUserStoreError(err)
		}
		if !meta.IsValidMajorDescription(input.Department) {
			return httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "invalid major description")
		}
		if !meta.IsValidGrade(input.Grade) {
			return httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "invalid grade")
		}

		if _, err := s.store.UpdateUserProfile(ctx, tx, store.UpdateProfileParams{
			UserID:       userID,
			Nickname:     input.Nickname,
			Department:   input.Department,
			Grade:        input.Grade,
			ContactType:  currentUser.ContactType,
			ContactValue: currentUser.ContactValue,
			Now:          now,
		}); err != nil {
			if isDuplicateNicknameErr(err) {
				return httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, "nickname already exists")
			}
			return mapUserStoreError(err)
		}

		if input.ContactType != "" {
			if err := s.replaceUserContactMethod(ctx, tx, userID, input.ContactType, input.ContactValue, now); err != nil {
				return err
			}
		}

		loaded, err := s.loadDetails(ctx, tx, userID)
		if err != nil {
			return err
		}
		details = loaded
		return nil
	})
	if err != nil {
		return Details{}, err
	}
	return details, nil
}

// UpsertContact 保存当前用户最后一次主动选择的联系方式。
func (s *Service) UpsertContact(ctx context.Context, userID int64, input UpsertContactInput) (Details, error) {
	if err := s.ensureReady(); err != nil {
		return Details{}, err
	}
	if err := validateContact(input.ContactType, input.ContactValue); err != nil {
		return Details{}, err
	}

	var details Details
	now := time.Now()
	err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		if _, err := s.store.GetUserByID(ctx, tx, userID); err != nil {
			return mapUserStoreError(err)
		}
		if err := s.replaceUserContactMethod(ctx, tx, userID, input.ContactType, input.ContactValue, now); err != nil {
			return err
		}
		loaded, err := s.loadDetails(ctx, tx, userID)
		if err != nil {
			return err
		}
		details = loaded
		return nil
	})
	if err != nil {
		return Details{}, err
	}
	return details, nil
}

// DeleteContact 删除指定渠道的联系方式。
func (s *Service) DeleteContact(ctx context.Context, userID int64, contactType string) (Details, error) {
	if err := s.ensureReady(); err != nil {
		return Details{}, err
	}
	contactType = strings.TrimSpace(contactType)
	if !meta.IsValidContactType(contactType) {
		return Details{}, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
	}

	var details Details
	now := time.Now()
	err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		if _, err := s.store.GetUserByID(ctx, tx, userID); err != nil {
			return mapUserStoreError(err)
		}
		if err := s.store.DeleteUserContactMethod(ctx, tx, userID, contactType); err != nil {
			if err == store.ErrNotFound {
				return httpapi.NewError(http.StatusNotFound, httpapi.CodeContactNotFound, "联系方式不存在")
			}
			return httpapi.InternalError(err)
		}
		if err := s.store.SyncUserPrimaryContact(ctx, tx, userID, now); err != nil {
			return mapUserStoreError(err)
		}
		loaded, err := s.loadDetails(ctx, tx, userID)
		if err != nil {
			return err
		}
		details = loaded
		return nil
	})
	if err != nil {
		return Details{}, err
	}
	return details, nil
}

func (s *Service) ensureReady() error {
	if s == nil || s.store == nil {
		return httpapi.FeatureNotReady("profile")
	}
	return nil
}

func (s *Service) loadDetails(ctx context.Context, tx *sql.Tx, userID int64) (Details, error) {
	user, err := s.store.GetUserByID(ctx, tx, userID)
	if err != nil {
		return Details{}, mapUserStoreError(err)
	}
	contacts, err := s.store.ListUserContactMethods(ctx, tx, userID)
	if err != nil {
		return Details{}, httpapi.InternalError(err)
	}
	return Details{User: user, Contacts: contacts}, nil
}

func (s *Service) replaceUserContactMethod(ctx context.Context, tx *sql.Tx, userID int64, contactType, contactValue string, now time.Time) error {
	if err := s.store.DeleteAllUserContactMethods(ctx, tx, userID); err != nil {
		return httpapi.InternalError(err)
	}
	if _, err := s.store.UpsertUserContactMethod(ctx, tx, store.UpsertContactMethodParams{
		UserID:       userID,
		ContactType:  contactType,
		ContactValue: contactValue,
		CreatedAt:    now,
		UpdatedAt:    now,
	}); err != nil {
		return httpapi.InternalError(err)
	}
	if err := s.store.SyncUserPrimaryContact(ctx, tx, userID, now); err != nil {
		return mapUserStoreError(err)
	}
	return nil
}

func validateContact(contactType, contactValue string) error {
	contactType = strings.TrimSpace(contactType)
	contactValue = strings.TrimSpace(contactValue)
	if !meta.IsValidContactType(contactType) || contactValue == "" {
		return httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
	}
	if utf8.RuneCountInString(contactValue) > maxContactValueLength {
		return httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
	}
	return nil
}

func mapUserStoreError(err error) error {
	if err == nil {
		return nil
	}
	if err == store.ErrNotFound {
		return httpapi.NewError(http.StatusNotFound, httpapi.CodeUserNotFound, "用户不存在")
	}
	return httpapi.InternalError(err)
}

func isDuplicateNicknameErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "ux_users_nickname_active")
}
