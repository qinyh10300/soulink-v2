package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type EngagementPreferences struct {
	UserID                int64
	NotificationsEnabled  bool
	NewMessages           bool
	MeetupInvites         bool
	MeetupReminders       bool
	DailyPicks            bool
	PermissionState       string
	Timezone              string
	SelectedScenes        []string
	OnboardingCompletedAt sql.NullTime
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type UpsertEngagementPreferencesParams struct {
	UserID               int64
	NotificationsEnabled bool
	NewMessages          bool
	MeetupInvites        bool
	MeetupReminders      bool
	DailyPicks           bool
	PermissionState      string
	Timezone             string
	Now                  time.Time
}

type CampusTheme struct {
	Key           string
	Title         string
	Description   string
	Category      string
	DefaultPrompt string
	ScheduleHint  string
	IconName      string
	IsActive      bool
	DisplayOrder  int
	Selected      bool
	UpdatedAt     time.Time
}

type UpdateCampusThemeParams struct {
	Key           string
	Title         string
	Description   string
	DefaultPrompt string
	ScheduleHint  string
	IconName      string
	IsActive      bool
	DisplayOrder  int
	Now           time.Time
}

type UserNotification struct {
	ID           int64
	UserID       int64
	Kind         string
	Title        string
	Body         string
	ActionPath   string
	SourceType   string
	SourceID     sql.NullInt64
	ScheduledFor time.Time
	DeliveredAt  sql.NullTime
	ReadAt       sql.NullTime
	CreatedAt    time.Time
}

type CampusThemeMetric struct {
	Key              string
	Title            string
	Category         string
	SupplyUsers      int64
	DemandUsers      int64
	GroupIntents     int64
	GroupsFormed     int64
	MeetupsEligible  int64
	MeetupsCompleted int64
}

func (s *Store) GetEngagementPreferences(
	ctx context.Context,
	tx *sql.Tx,
	userID int64,
) (EngagementPreferences, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT user_id, notifications_enabled, new_messages, meetup_invites,
		       meetup_reminders, daily_picks, permission_state, timezone,
		       selected_scenes::text, onboarding_completed_at, created_at, updated_at
		FROM user_engagement_preferences
		WHERE user_id = $1
	`, userID)
	return scanEngagementPreferences(row)
}

func (s *Store) UpsertEngagementPreferences(
	ctx context.Context,
	tx *sql.Tx,
	params UpsertEngagementPreferencesParams,
) (EngagementPreferences, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		INSERT INTO user_engagement_preferences (
			user_id, notifications_enabled, new_messages, meetup_invites,
			meetup_reminders, daily_picks, permission_state, timezone,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
		ON CONFLICT (user_id) DO UPDATE SET
			notifications_enabled = EXCLUDED.notifications_enabled,
			new_messages = EXCLUDED.new_messages,
			meetup_invites = EXCLUDED.meetup_invites,
			meetup_reminders = EXCLUDED.meetup_reminders,
			daily_picks = EXCLUDED.daily_picks,
			permission_state = EXCLUDED.permission_state,
			timezone = EXCLUDED.timezone,
			updated_at = EXCLUDED.updated_at
		RETURNING user_id, notifications_enabled, new_messages, meetup_invites,
		          meetup_reminders, daily_picks, permission_state, timezone,
		          selected_scenes::text, onboarding_completed_at, created_at, updated_at
	`, params.UserID, params.NotificationsEnabled, params.NewMessages,
		params.MeetupInvites, params.MeetupReminders, params.DailyPicks,
		params.PermissionState, params.Timezone, params.Now)
	return scanEngagementPreferences(row)
}

func (s *Store) CompleteUserOnboarding(
	ctx context.Context,
	tx *sql.Tx,
	userID int64,
	scenes []string,
	romanticOptIn bool,
	now time.Time,
) (EngagementPreferences, error) {
	rawScenes, err := marshalStringList(scenes)
	if err != nil {
		return EngagementPreferences{}, err
	}
	exec := executor(s.db, tx)
	if _, err := exec.ExecContext(ctx, `
		INSERT INTO user_engagement_preferences (
			user_id, selected_scenes, onboarding_completed_at, created_at, updated_at
		) VALUES ($1, $2::jsonb, $3, $3, $3)
		ON CONFLICT (user_id) DO UPDATE SET
			selected_scenes = EXCLUDED.selected_scenes,
			onboarding_completed_at = EXCLUDED.onboarding_completed_at,
			updated_at = EXCLUDED.updated_at
	`, userID, rawScenes, now); err != nil {
		return EngagementPreferences{}, fmt.Errorf("complete user onboarding: %w", err)
	}

	if _, err := exec.ExecContext(ctx, `
		INSERT INTO matching_profiles (user_id, preferred_scenes, romantic_opt_in, created_at, updated_at)
		VALUES ($1, $2::jsonb, $3, $4, $4)
		ON CONFLICT (user_id) DO UPDATE SET
			preferred_scenes = EXCLUDED.preferred_scenes,
			romantic_opt_in = EXCLUDED.romantic_opt_in,
			updated_at = EXCLUDED.updated_at
	`, userID, rawScenes, romanticOptIn, now); err != nil {
		return EngagementPreferences{}, fmt.Errorf("sync onboarding matching profile: %w", err)
	}

	if _, err := exec.ExecContext(ctx, `
		DELETE FROM user_campus_themes
		WHERE user_id = $1
	`, userID); err != nil {
		return EngagementPreferences{}, fmt.Errorf("clear onboarding campus themes: %w", err)
	}
	if _, err := exec.ExecContext(ctx, `
		INSERT INTO user_campus_themes (user_id, theme_key, selected_at)
		SELECT $1, theme.theme_key, $3
		FROM campus_themes theme
		WHERE theme.is_active
		  AND EXISTS (
			SELECT 1
			FROM jsonb_array_elements_text($2::jsonb) selected(scene)
			WHERE selected.scene = theme.category
		  )
		ON CONFLICT (user_id, theme_key) DO UPDATE SET selected_at = EXCLUDED.selected_at
	`, userID, rawScenes, now); err != nil {
		return EngagementPreferences{}, fmt.Errorf("select onboarding campus themes: %w", err)
	}

	return s.GetEngagementPreferences(ctx, tx, userID)
}

func (s *Store) ListCampusThemes(
	ctx context.Context,
	tx *sql.Tx,
	userID int64,
	includeInactive bool,
) ([]CampusTheme, error) {
	rows, err := queryContext(executor(s.db, tx), ctx, `
		SELECT theme.theme_key, theme.title, theme.description, theme.category,
		       theme.default_prompt, theme.schedule_hint, theme.icon_name,
		       theme.is_active, theme.display_order,
		       (selected.user_id IS NOT NULL), theme.updated_at
		FROM campus_themes theme
		LEFT JOIN user_campus_themes selected
		  ON selected.theme_key = theme.theme_key AND selected.user_id = $1
		WHERE $2 OR theme.is_active
		ORDER BY theme.display_order, theme.theme_key
	`, userID, includeInactive)
	if err != nil {
		return nil, fmt.Errorf("list campus themes: %w", err)
	}
	defer rows.Close()

	themes := make([]CampusTheme, 0)
	for rows.Next() {
		var theme CampusTheme
		if err := rows.Scan(
			&theme.Key, &theme.Title, &theme.Description, &theme.Category,
			&theme.DefaultPrompt, &theme.ScheduleHint, &theme.IconName,
			&theme.IsActive, &theme.DisplayOrder, &theme.Selected, &theme.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan campus theme: %w", err)
		}
		themes = append(themes, theme)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate campus themes: %w", err)
	}
	return themes, nil
}

func (s *Store) ReplaceUserCampusThemes(
	ctx context.Context,
	tx *sql.Tx,
	userID int64,
	themeKeys []string,
	now time.Time,
) error {
	rawKeys, err := marshalStringList(themeKeys)
	if err != nil {
		return err
	}
	exec := executor(s.db, tx)
	if _, err := exec.ExecContext(ctx, `DELETE FROM user_campus_themes WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("clear user campus themes: %w", err)
	}
	result, err := exec.ExecContext(ctx, `
		INSERT INTO user_campus_themes (user_id, theme_key, selected_at)
		SELECT $1, theme.theme_key, $3
		FROM campus_themes theme
		WHERE theme.is_active
		  AND EXISTS (
			SELECT 1 FROM jsonb_array_elements_text($2::jsonb) selected(theme_key)
			WHERE selected.theme_key = theme.theme_key
		  )
	`, userID, rawKeys, now)
	if err != nil {
		return fmt.Errorf("replace user campus themes: %w", err)
	}
	inserted, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("count selected campus themes: %w", err)
	}
	if inserted != int64(len(themeKeys)) {
		return ErrNotFound
	}
	return nil
}

func (s *Store) UpdateCampusTheme(
	ctx context.Context,
	tx *sql.Tx,
	params UpdateCampusThemeParams,
) (CampusTheme, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		UPDATE campus_themes
		SET title = $2, description = $3, default_prompt = $4,
		    schedule_hint = $5, icon_name = $6, is_active = $7,
		    display_order = $8, updated_at = $9
		WHERE theme_key = $1
		RETURNING theme_key, title, description, category, default_prompt,
		          schedule_hint, icon_name, is_active, display_order, FALSE, updated_at
	`, params.Key, params.Title, params.Description, params.DefaultPrompt,
		params.ScheduleHint, params.IconName, params.IsActive, params.DisplayOrder, params.Now)
	var theme CampusTheme
	if err := row.Scan(
		&theme.Key, &theme.Title, &theme.Description, &theme.Category,
		&theme.DefaultPrompt, &theme.ScheduleHint, &theme.IconName,
		&theme.IsActive, &theme.DisplayOrder, &theme.Selected, &theme.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return CampusTheme{}, ErrNotFound
		}
		return CampusTheme{}, fmt.Errorf("update campus theme: %w", err)
	}
	return theme, nil
}

func (s *Store) MaterializeUserNotifications(
	ctx context.Context,
	tx *sql.Tx,
	userID int64,
	localDate string,
	now time.Time,
) error {
	exec := executor(s.db, tx)
	if _, err := exec.ExecContext(ctx, `
		INSERT INTO user_notifications (
			user_id, kind, title, body, action_path, source_type, source_id,
			dedupe_key, scheduled_for, created_at
		)
		SELECT $1, 'meetup_reminder',
		       CASE WHEN meetup.starts_at <= $2::timestamptz + INTERVAL '2 hours'
		            THEN '约见即将开始' ELSE '约见将在 24 小时内开始' END,
		       '时间和地点仅在登录 SouLink 后显示。',
		       '/calendar', 'meetup', meetup.id,
		       'meetup-reminder:' || meetup.id || ':' ||
		         CASE WHEN meetup.starts_at <= $2::timestamptz + INTERVAL '2 hours' THEN '2h' ELSE '24h' END,
		       $2::timestamptz, $2::timestamptz
		FROM meetups meetup
		JOIN meetup_participants participant
		  ON participant.meetup_id = meetup.id AND participant.user_id = $1
		JOIN user_engagement_preferences preference ON preference.user_id = $1
		WHERE meetup.status = 'confirmed'
		  AND participant.status = 'accepted'
		  AND meetup.starts_at > $2::timestamptz
		  AND meetup.starts_at <= $2::timestamptz + INTERVAL '24 hours'
		  AND preference.notifications_enabled
		  AND preference.meetup_reminders
		ON CONFLICT (user_id, dedupe_key) DO NOTHING
	`, userID, now); err != nil {
		return fmt.Errorf("materialize meetup reminders: %w", err)
	}

	if _, err := exec.ExecContext(ctx, `
		INSERT INTO user_notifications (
			user_id, kind, title, body, action_path, source_type, source_id,
			dedupe_key, scheduled_for, created_at
		)
		SELECT $1, 'daily_picks', '今天的 3 个精选连接已准备好',
		       '打开 SouLink 查看推荐理由和不确定项。',
		       '/connect', '', NULL,
		       'daily-picks:' || $2, $3::timestamptz, $3::timestamptz
		FROM user_engagement_preferences preference
		WHERE preference.user_id = $1
		  AND preference.notifications_enabled
		  AND preference.daily_picks
		  AND EXISTS (
			SELECT 1 FROM daily_recommendations recommendation
			WHERE recommendation.user_id = $1
			  AND recommendation.expires_at > $3::timestamptz
		  )
		ON CONFLICT (user_id, dedupe_key) DO NOTHING
	`, userID, localDate, now); err != nil {
		return fmt.Errorf("materialize daily picks notification: %w", err)
	}
	return nil
}

func (s *Store) ListUserNotifications(
	ctx context.Context,
	tx *sql.Tx,
	userID int64,
	beforeID int64,
	limit int,
) ([]UserNotification, error) {
	rows, err := queryContext(executor(s.db, tx), ctx, `
		SELECT id, user_id, kind, title, body, action_path, source_type, source_id,
		       scheduled_for, delivered_at, read_at, created_at
		FROM user_notifications
		WHERE user_id = $1
		  AND scheduled_for <= NOW()
		  AND ($2 = 0 OR id < $2)
		ORDER BY id DESC
		LIMIT $3
	`, userID, beforeID, normalizeNotificationLimit(limit))
	if err != nil {
		return nil, fmt.Errorf("list user notifications: %w", err)
	}
	defer rows.Close()

	items := make([]UserNotification, 0)
	for rows.Next() {
		var item UserNotification
		if err := rows.Scan(
			&item.ID, &item.UserID, &item.Kind, &item.Title, &item.Body,
			&item.ActionPath, &item.SourceType, &item.SourceID,
			&item.ScheduledFor, &item.DeliveredAt, &item.ReadAt, &item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user notification: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user notifications: %w", err)
	}
	return items, nil
}

func (s *Store) CountUnreadUserNotifications(
	ctx context.Context,
	tx *sql.Tx,
	userID int64,
) (int64, error) {
	var count int64
	if err := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT COUNT(*) FROM user_notifications
		WHERE user_id = $1 AND read_at IS NULL AND scheduled_for <= NOW()
	`, userID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count unread user notifications: %w", err)
	}
	return count, nil
}

func (s *Store) MarkUserNotificationRead(
	ctx context.Context,
	tx *sql.Tx,
	userID int64,
	notificationID int64,
	now time.Time,
) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE user_notifications
		SET read_at = COALESCE(read_at, $3)
		WHERE id = $1 AND user_id = $2
	`, notificationID, userID, now)
	if err != nil {
		return fmt.Errorf("mark user notification read: %w", err)
	}
	return requireRowsAffected(result)
}

func (s *Store) MarkAllUserNotificationsRead(
	ctx context.Context,
	tx *sql.Tx,
	userID int64,
	now time.Time,
) error {
	_, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE user_notifications SET read_at = $2
		WHERE user_id = $1 AND read_at IS NULL AND scheduled_for <= $2
	`, userID, now)
	if err != nil {
		return fmt.Errorf("mark all user notifications read: %w", err)
	}
	return nil
}

func (s *Store) MarkUserNotificationDelivered(
	ctx context.Context,
	tx *sql.Tx,
	userID int64,
	notificationID int64,
	now time.Time,
) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE user_notifications
		SET delivered_at = COALESCE(delivered_at, $3)
		WHERE id = $1 AND user_id = $2
	`, notificationID, userID, now)
	if err != nil {
		return fmt.Errorf("mark user notification delivered: %w", err)
	}
	return requireRowsAffected(result)
}

func (s *Store) GetCampusThemeMetrics(
	ctx context.Context,
	tx *sql.Tx,
	from time.Time,
	to time.Time,
) ([]CampusThemeMetric, error) {
	rows, err := queryContext(executor(s.db, tx), ctx, `
		SELECT theme.theme_key, theme.title, theme.category,
		       COALESCE(supply.user_count, 0), COALESCE(demand.user_count, 0),
		       COALESCE(groups.intent_count, 0), COALESCE(groups.formed_count, 0),
		       COALESCE(meetings.eligible_count, 0), COALESCE(meetings.completed_count, 0)
		FROM campus_themes theme
		LEFT JOIN LATERAL (
			SELECT COUNT(DISTINCT offered.user_id) AS user_count
			FROM (
				SELECT intent.author_id AS user_id
				FROM connection_intents intent
				WHERE intent.category = theme.category AND intent.status = 'active'
				UNION ALL
				SELECT group_intent.organizer_id AS user_id
				FROM group_match_intents group_intent
				WHERE group_intent.category = theme.category AND group_intent.status = 'open'
			) offered
		) supply ON TRUE
		LEFT JOIN LATERAL (
			SELECT COUNT(DISTINCT wanted.user_id) AS user_count
			FROM (
				SELECT selected.user_id
				FROM user_campus_themes selected
				WHERE selected.theme_key = theme.theme_key
				  AND selected.selected_at >= $1 AND selected.selected_at < $2
				UNION ALL
				SELECT interest.user_id
				FROM intent_interests interest
				JOIN connection_intents intent ON intent.id = interest.intent_id
				WHERE intent.category = theme.category
				  AND interest.created_at >= $1 AND interest.created_at < $2
			) wanted
		) demand ON TRUE
		LEFT JOIN LATERAL (
			SELECT COUNT(*) AS intent_count,
			       COUNT(*) FILTER (WHERE group_intent.formed_at IS NOT NULL) AS formed_count
			FROM group_match_intents group_intent
			WHERE group_intent.category = theme.category
			  AND group_intent.created_at >= $1 AND group_intent.created_at < $2
		) groups ON TRUE
		LEFT JOIN LATERAL (
			SELECT COUNT(*) FILTER (
			           WHERE meetup.status IN ('confirmed', 'completed', 'no_show')
			       ) AS eligible_count,
			       COUNT(*) FILTER (WHERE meetup.status = 'completed') AS completed_count
			FROM meetups meetup
			WHERE meetup.category = theme.category
			  AND meetup.starts_at >= $1 AND meetup.starts_at < $2
		) meetings ON TRUE
		ORDER BY theme.display_order, theme.theme_key
	`, from, to)
	if err != nil {
		return nil, fmt.Errorf("get campus theme metrics: %w", err)
	}
	defer rows.Close()

	metrics := make([]CampusThemeMetric, 0)
	for rows.Next() {
		var metric CampusThemeMetric
		if err := rows.Scan(
			&metric.Key, &metric.Title, &metric.Category,
			&metric.SupplyUsers, &metric.DemandUsers,
			&metric.GroupIntents, &metric.GroupsFormed,
			&metric.MeetupsEligible, &metric.MeetupsCompleted,
		); err != nil {
			return nil, fmt.Errorf("scan campus theme metric: %w", err)
		}
		metrics = append(metrics, metric)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate campus theme metrics: %w", err)
	}
	return metrics, nil
}

func scanEngagementPreferences(row *sql.Row) (EngagementPreferences, error) {
	var preferences EngagementPreferences
	var scenes string
	if err := row.Scan(
		&preferences.UserID, &preferences.NotificationsEnabled,
		&preferences.NewMessages, &preferences.MeetupInvites,
		&preferences.MeetupReminders, &preferences.DailyPicks,
		&preferences.PermissionState, &preferences.Timezone, &scenes,
		&preferences.OnboardingCompletedAt, &preferences.CreatedAt, &preferences.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return EngagementPreferences{}, ErrNotFound
		}
		return EngagementPreferences{}, fmt.Errorf("scan engagement preferences: %w", err)
	}
	if err := unmarshalStringList(scenes, &preferences.SelectedScenes); err != nil {
		return EngagementPreferences{}, err
	}
	return preferences, nil
}

func normalizeNotificationLimit(limit int) int {
	if limit <= 0 {
		return 21
	}
	if limit > 51 {
		return 51
	}
	return limit
}
