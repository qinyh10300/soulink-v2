package profile

import (
	"context"
	"database/sql"
	"errors"
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
	User        store.User
	Settings    PrivacySettings
	Preferences *PublicPreferences
}

type PrivacySettings struct {
	ProfilePublic     bool
	SchedulePublic    bool
	PreferencesPublic bool
	AvatarURL         string
}

type PublicPreferences struct {
	Bio                string
	Skills             []string
	Interests          []string
	Availability       string
	PreferredScenes    []string
	CollaborationStyle string
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

func (s *Service) GetPrivacy(ctx context.Context, userID int64) (PrivacySettings, error) {
	if err := s.ensureReady(); err != nil {
		return PrivacySettings{}, err
	}
	settings, err := s.store.GetUserProfileSettings(ctx, nil, userID)
	if errors.Is(err, store.ErrNotFound) {
		return PrivacySettings{}, nil
	}
	if err != nil {
		return PrivacySettings{}, httpapi.InternalError(err)
	}
	return privacySettingsFromStore(settings), nil
}

func (s *Service) UpdatePrivacy(ctx context.Context, userID int64, settings PrivacySettings) (PrivacySettings, error) {
	if err := s.ensureReady(); err != nil {
		return PrivacySettings{}, err
	}
	if _, err := s.store.GetUserByID(ctx, nil, userID); err != nil {
		return PrivacySettings{}, mapUserStoreError(err)
	}
	if !settings.ProfilePublic {
		settings.SchedulePublic = false
		settings.PreferencesPublic = false
	}
	updated, err := s.store.UpsertUserProfileSettings(ctx, nil, store.UpsertUserProfileSettingsParams{
		UserID:            userID,
		ProfilePublic:     settings.ProfilePublic,
		SchedulePublic:    settings.SchedulePublic,
		PreferencesPublic: settings.PreferencesPublic,
		Now:               time.Now(),
	})
	if err != nil {
		return PrivacySettings{}, httpapi.InternalError(err)
	}
	return privacySettingsFromStore(updated), nil
}

// GetPublic returns only the profile fields that the owner has explicitly shared.
func (s *Service) GetPublic(ctx context.Context, viewerID, userID int64) (PublicDetails, error) {
	if err := s.ensureReady(); err != nil {
		return PublicDetails{}, err
	}
	user, err := s.store.GetUserByID(ctx, nil, userID)
	if err != nil {
		return PublicDetails{}, mapUserStoreError(err)
	}
	settings, err := s.GetPrivacy(ctx, userID)
	if err != nil {
		return PublicDetails{}, err
	}
	isSelf := viewerID == userID
	if !isSelf && !settings.ProfilePublic {
		return PublicDetails{}, httpapi.NewError(http.StatusForbidden, httpapi.CodeForbidden, "该用户尚未公开个人主页")
	}

	details := PublicDetails{User: user, Settings: settings}
	if isSelf || settings.PreferencesPublic {
		matchingProfile, profileErr := s.store.GetMatchingProfile(ctx, nil, userID)
		if profileErr == nil {
			details.Preferences = &PublicPreferences{
				Bio:                matchingProfile.Bio,
				Skills:             matchingProfile.Skills,
				Interests:          matchingProfile.Interests,
				Availability:       matchingProfile.Availability,
				PreferredScenes:    matchingProfile.PreferredScenes,
				CollaborationStyle: matchingProfile.CollaborationStyle,
			}
		} else if !errors.Is(profileErr, store.ErrNotFound) {
			return PublicDetails{}, httpapi.InternalError(profileErr)
		}
	}
	return details, nil
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

func privacySettingsFromStore(settings store.UserProfileSettings) PrivacySettings {
	avatarURL := ""
	if settings.AvatarAssetKey != "" {
		avatarURL = "/media/" + settings.AvatarAssetKey
	}
	return PrivacySettings{
		ProfilePublic:     settings.ProfilePublic,
		SchedulePublic:    settings.SchedulePublic,
		PreferencesPublic: settings.PreferencesPublic,
		AvatarURL:         avatarURL,
	}
}
