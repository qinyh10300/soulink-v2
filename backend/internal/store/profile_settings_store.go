package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type UserProfileSettings struct {
	UserID            int64
	ProfilePublic     bool
	SchedulePublic    bool
	PreferencesPublic bool
	AvatarMediaID     sql.NullInt64
	AvatarAssetKey    string
	UpdatedAt         time.Time
}

type UpsertUserProfileSettingsParams struct {
	UserID            int64
	ProfilePublic     bool
	SchedulePublic    bool
	PreferencesPublic bool
	Now               time.Time
}

func (s *Store) GetUserProfileSettings(ctx context.Context, tx *sql.Tx, userID int64) (UserProfileSettings, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT s.user_id, s.profile_public, s.schedule_public, s.preferences_public,
		       s.avatar_media_id, COALESCE(m.asset_key::text, ''), s.updated_at
		FROM user_profile_settings s
		LEFT JOIN media_assets m ON m.id = s.avatar_media_id
		WHERE s.user_id = $1
		LIMIT 1
	`, userID)
	return scanUserProfileSettings(row)
}

func (s *Store) UpsertUserProfileSettings(ctx context.Context, tx *sql.Tx, params UpsertUserProfileSettingsParams) (UserProfileSettings, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		INSERT INTO user_profile_settings (
			user_id, profile_public, schedule_public, preferences_public, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $5)
		ON CONFLICT (user_id) DO UPDATE SET
			profile_public = EXCLUDED.profile_public,
			schedule_public = EXCLUDED.schedule_public,
			preferences_public = EXCLUDED.preferences_public,
			updated_at = EXCLUDED.updated_at
		RETURNING user_id, profile_public, schedule_public, preferences_public,
		          avatar_media_id, updated_at
	`, params.UserID, params.ProfilePublic, params.SchedulePublic, params.PreferencesPublic, params.Now)

	var settings UserProfileSettings
	if err := row.Scan(
		&settings.UserID,
		&settings.ProfilePublic,
		&settings.SchedulePublic,
		&settings.PreferencesPublic,
		&settings.AvatarMediaID,
		&settings.UpdatedAt,
	); err != nil {
		return UserProfileSettings{}, fmt.Errorf("upsert user profile settings: %w", err)
	}
	if settings.AvatarMediaID.Valid {
		asset, err := s.GetMediaAssetByID(ctx, tx, settings.AvatarMediaID.Int64)
		if err != nil {
			return UserProfileSettings{}, err
		}
		settings.AvatarAssetKey = asset.AssetKey
	}
	return settings, nil
}

func (s *Store) SetUserAvatarMedia(ctx context.Context, tx *sql.Tx, userID, mediaID int64, now time.Time) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		INSERT INTO user_profile_settings (user_id, avatar_media_id, created_at, updated_at)
		VALUES ($1, $2, $3, $3)
		ON CONFLICT (user_id) DO UPDATE SET
			avatar_media_id = EXCLUDED.avatar_media_id,
			updated_at = EXCLUDED.updated_at
	`, userID, mediaID, now)
	if err != nil {
		return fmt.Errorf("set user avatar media: %w", err)
	}
	return requireRowsAffected(result)
}

func (s *Store) ClearUserAvatarMedia(ctx context.Context, tx *sql.Tx, userID int64, now time.Time) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE user_profile_settings
		SET avatar_media_id = NULL, updated_at = $2
		WHERE user_id = $1
	`, userID, now)
	if err != nil {
		return fmt.Errorf("clear user avatar media: %w", err)
	}
	return requireRowsAffected(result)
}

func (s *Store) GetMediaAssetByID(ctx context.Context, tx *sql.Tx, mediaID int64) (MediaAsset, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT id, asset_key::text, owner_id, kind, filename, mime_type, byte_size,
		       width, height, is_animated, content, created_at
		FROM media_assets
		WHERE id = $1
		LIMIT 1
	`, mediaID)
	return scanMediaAsset(row)
}

func scanUserProfileSettings(scanner rowScanner) (UserProfileSettings, error) {
	var settings UserProfileSettings
	if err := scanner.Scan(
		&settings.UserID,
		&settings.ProfilePublic,
		&settings.SchedulePublic,
		&settings.PreferencesPublic,
		&settings.AvatarMediaID,
		&settings.AvatarAssetKey,
		&settings.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return UserProfileSettings{}, ErrNotFound
		}
		return UserProfileSettings{}, fmt.Errorf("scan user profile settings: %w", err)
	}
	return settings, nil
}
