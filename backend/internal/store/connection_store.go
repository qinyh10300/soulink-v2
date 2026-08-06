package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type MatchingProfile struct {
	UserID             int64
	Bio                string
	Skills             []string
	Interests          []string
	Availability       string
	PreferredScenes    []string
	CollaborationStyle string
	Boundaries         string
	AIPortrait         string
	AIPortraitApproved bool
	RomanticOptIn      bool
	MatchingPaused     bool
	HiddenCategories   []string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type UpsertMatchingProfileParams struct {
	UserID             int64
	Bio                string
	Skills             []string
	Interests          []string
	Availability       string
	PreferredScenes    []string
	CollaborationStyle string
	Boundaries         string
	AIPortrait         string
	AIPortraitApproved bool
	RomanticOptIn      bool
	MatchingPaused     bool
	HiddenCategories   []string
	Now                time.Time
}

type ProjectMemo struct {
	Building         string   `json:"building"`
	Stage            string   `json:"stage"`
	WeeklyTime       string   `json:"weeklyTime"`
	Compensation     string   `json:"compensation"`
	RequiredSkills   []string `json:"requiredSkills"`
	NiceToHaveSkills []string `json:"niceToHaveSkills"`
	WorkingStyle     string   `json:"workingStyle"`
	RiskAppetite     string   `json:"riskAppetite"`
}

type ConnectionIntent struct {
	ID                       int64
	AuthorID                 int64
	AuthorNickname           string
	AuthorDepartment         string
	AuthorGrade              string
	AuthorBio                string
	AuthorSkills             []string
	AuthorAvailability       string
	AuthorCollaborationStyle string
	Category                 string
	Title                    string
	Description              string
	Goal                     string
	TimeWindow               string
	Location                 string
	Visibility               string
	Status                   string
	Constraints              []string
	DealBreakers             []string
	DesiredNextStep          string
	ProjectMemo              ProjectMemo
	ExpiresAt                sql.NullTime
	PublishedAt              sql.NullTime
	InterestCount            int
	InterestState            string
	SavedByMe                bool
	CreatedAt                time.Time
	UpdatedAt                time.Time
}

type CreateConnectionIntentParams struct {
	AuthorID        int64
	Category        string
	Title           string
	Description     string
	Goal            string
	TimeWindow      string
	Location        string
	Visibility      string
	Status          string
	Constraints     []string
	DealBreakers    []string
	DesiredNextStep string
	ProjectMemo     ProjectMemo
	ExpiresAt       sql.NullTime
	PublishedAt     sql.NullTime
	Now             time.Time
}

type UpdateConnectionIntentParams struct {
	IntentID        int64
	AuthorID        int64
	Category        string
	Title           string
	Description     string
	Goal            string
	TimeWindow      string
	Location        string
	Visibility      string
	Status          string
	Constraints     []string
	DealBreakers    []string
	DesiredNextStep string
	ProjectMemo     ProjectMemo
	ExpiresAt       sql.NullTime
	PublishedAt     sql.NullTime
	Now             time.Time
}

type IntentInterest struct {
	IntentID           int64
	UserID             int64
	Nickname           string
	Department         string
	Grade              string
	Bio                string
	Skills             []string
	CollaborationStyle string
	Status             string
	Message            string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type CreateMatchRoomParams struct {
	IntentID             int64
	OwnerID              int64
	CandidateID          int64
	FitReason            string
	Uncertainty          string
	SharedContext        []string
	ConversationStarters []string
	SuggestedPlaces      []string
	AssistantSource      string
	AssistantModel       string
	AssistantConfidence  float64
	AssistantPrompt      string
	Now                  time.Time
}

type MatchRoom struct {
	ID                     int64
	IntentID               int64
	IntentTitle            string
	IntentCategory         string
	OwnerID                int64
	OwnerNickname          string
	OwnerDepartment        string
	OwnerGrade             string
	CandidateID            int64
	CandidateNickname      string
	CandidateDepartment    string
	CandidateGrade         string
	Status                 string
	FitReason              string
	Uncertainty            string
	SharedContext          []string
	ConversationStarters   []string
	SuggestedPlaces        []string
	AssistantSource        string
	AssistantModel         string
	AssistantConfidence    float64
	AssistantPromptVersion string
	ScheduledAt            sql.NullTime
	SelectedPlace          string
	OwnerContactShared     bool
	CandidateContactShared bool
	OwnerContactType       string
	OwnerContactValue      string
	CandidateContactType   string
	CandidateContactValue  string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type MatchRoomMessage struct {
	ID             int64
	RoomID         int64
	SenderID       int64
	SenderNickname string
	Content        string
	CreatedAt      time.Time
}

func (s *Store) GetMatchingProfile(ctx context.Context, tx *sql.Tx, userID int64) (MatchingProfile, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT user_id, bio, skills::text, interests::text, availability,
		       preferred_scenes::text, collaboration_style, boundaries,
		       ai_portrait, ai_portrait_approved, romantic_opt_in, matching_paused,
		       hidden_categories::text, created_at, updated_at
		FROM matching_profiles
		WHERE user_id = $1
	`, userID)
	return scanMatchingProfile(row)
}

func (s *Store) UpsertMatchingProfile(ctx context.Context, tx *sql.Tx, params UpsertMatchingProfileParams) (MatchingProfile, error) {
	skills, err := marshalStringList(params.Skills)
	if err != nil {
		return MatchingProfile{}, err
	}
	interests, err := marshalStringList(params.Interests)
	if err != nil {
		return MatchingProfile{}, err
	}
	scenes, err := marshalStringList(params.PreferredScenes)
	if err != nil {
		return MatchingProfile{}, err
	}
	hiddenCategories, err := marshalStringList(params.HiddenCategories)
	if err != nil {
		return MatchingProfile{}, err
	}

	row := executor(s.db, tx).QueryRowContext(ctx, `
		INSERT INTO matching_profiles (
			user_id, bio, skills, interests, availability, preferred_scenes,
			collaboration_style, boundaries, ai_portrait, ai_portrait_approved,
			romantic_opt_in, matching_paused, hidden_categories, created_at, updated_at
		) VALUES (
			$1, $2, $3::jsonb, $4::jsonb, $5, $6::jsonb,
			$7, $8, $9, $10, $11, $12, $13::jsonb, $14, $14
		)
		ON CONFLICT (user_id) DO UPDATE SET
			bio = EXCLUDED.bio,
			skills = EXCLUDED.skills,
			interests = EXCLUDED.interests,
			availability = EXCLUDED.availability,
			preferred_scenes = EXCLUDED.preferred_scenes,
			collaboration_style = EXCLUDED.collaboration_style,
			boundaries = EXCLUDED.boundaries,
			ai_portrait = EXCLUDED.ai_portrait,
			ai_portrait_approved = EXCLUDED.ai_portrait_approved,
			romantic_opt_in = EXCLUDED.romantic_opt_in,
			matching_paused = EXCLUDED.matching_paused,
			hidden_categories = EXCLUDED.hidden_categories,
			updated_at = EXCLUDED.updated_at
		RETURNING user_id, bio, skills::text, interests::text, availability,
		          preferred_scenes::text, collaboration_style, boundaries,
		          ai_portrait, ai_portrait_approved, romantic_opt_in, matching_paused,
		          hidden_categories::text, created_at, updated_at
	`, params.UserID, params.Bio, skills, interests, params.Availability, scenes,
		params.CollaborationStyle, params.Boundaries, params.AIPortrait,
		params.AIPortraitApproved, params.RomanticOptIn, params.MatchingPaused,
		hiddenCategories, params.Now)

	profile, err := scanMatchingProfile(row)
	if err != nil {
		return MatchingProfile{}, fmt.Errorf("upsert matching profile: %w", err)
	}
	return profile, nil
}

func (s *Store) CreateConnectionIntent(ctx context.Context, tx *sql.Tx, params CreateConnectionIntentParams) (int64, error) {
	constraints, err := marshalStringList(params.Constraints)
	if err != nil {
		return 0, err
	}
	dealBreakers, err := marshalStringList(params.DealBreakers)
	if err != nil {
		return 0, err
	}
	projectMemo, err := marshalProjectMemo(params.ProjectMemo)
	if err != nil {
		return 0, err
	}

	row := executor(s.db, tx).QueryRowContext(ctx, `
		INSERT INTO connection_intents (
			author_id, category, title, description, goal, time_window, location,
			visibility, status, constraints, deal_breakers, desired_next_step,
			project_memo, expires_at, published_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10::jsonb, $11::jsonb,
			$12, $13::jsonb, $14, $15, $16, $16
		)
		RETURNING id
	`, params.AuthorID, params.Category, params.Title, params.Description, params.Goal,
		params.TimeWindow, params.Location, params.Visibility, params.Status,
		constraints, dealBreakers, params.DesiredNextStep, projectMemo,
		nullableTime(params.ExpiresAt), nullableTime(params.PublishedAt), params.Now)

	var intentID int64
	if err := row.Scan(&intentID); err != nil {
		return 0, fmt.Errorf("create connection intent: %w", err)
	}
	return intentID, nil
}

func (s *Store) FindConnectionIntentIDByAuthorAndTitle(
	ctx context.Context,
	tx *sql.Tx,
	authorID int64,
	title string,
) (int64, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT id
		FROM connection_intents
		WHERE author_id = $1 AND title = $2
		ORDER BY id ASC
		LIMIT 1
	`, authorID, title)

	var intentID int64
	if err := row.Scan(&intentID); err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrNotFound
		}
		return 0, fmt.Errorf("find connection intent: %w", err)
	}
	return intentID, nil
}

func (s *Store) UpdateConnectionIntent(ctx context.Context, tx *sql.Tx, params UpdateConnectionIntentParams) error {
	constraints, err := marshalStringList(params.Constraints)
	if err != nil {
		return err
	}
	dealBreakers, err := marshalStringList(params.DealBreakers)
	if err != nil {
		return err
	}
	projectMemo, err := marshalProjectMemo(params.ProjectMemo)
	if err != nil {
		return err
	}

	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE connection_intents
		SET category = $3,
		    title = $4,
		    description = $5,
		    goal = $6,
		    time_window = $7,
		    location = $8,
		    visibility = $9,
		    status = $10,
		    constraints = $11::jsonb,
		    deal_breakers = $12::jsonb,
		    desired_next_step = $13,
		    project_memo = $14::jsonb,
		    expires_at = $15,
		    published_at = $16,
		    updated_at = $17
		WHERE id = $1 AND author_id = $2
	`, params.IntentID, params.AuthorID, params.Category, params.Title, params.Description,
		params.Goal, params.TimeWindow, params.Location, params.Visibility, params.Status,
		constraints, dealBreakers, params.DesiredNextStep, projectMemo,
		nullableTime(params.ExpiresAt), nullableTime(params.PublishedAt), params.Now)
	if err != nil {
		return fmt.Errorf("update connection intent: %w", err)
	}
	return requireRowsAffected(result)
}

func (s *Store) UpdateConnectionIntentStatus(ctx context.Context, tx *sql.Tx, intentID, authorID int64, status string, now time.Time) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE connection_intents
		SET status = $3, updated_at = $4
		WHERE id = $1 AND author_id = $2
	`, intentID, authorID, status, now)
	if err != nil {
		return fmt.Errorf("update connection intent status: %w", err)
	}
	return requireRowsAffected(result)
}

func (s *Store) GetConnectionIntent(ctx context.Context, tx *sql.Tx, viewerID, intentID int64) (ConnectionIntent, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT i.id, i.author_id, u.nickname, u.department, u.grade,
		       COALESCE(mp.bio, ''), COALESCE(mp.skills, '[]'::jsonb)::text,
		       COALESCE(mp.availability, ''), COALESCE(mp.collaboration_style, ''),
		       i.category, i.title, i.description, i.goal, i.time_window, i.location,
		       i.visibility, i.status, i.constraints::text, i.deal_breakers::text,
		       i.desired_next_step, i.project_memo::text, i.expires_at, i.published_at,
		       (SELECT COUNT(*) FROM intent_interests ii
		        WHERE ii.intent_id = i.id AND ii.status IN ('interested', 'accepted')) AS interest_count,
		       COALESCE((SELECT ii.status FROM intent_interests ii
		                 WHERE ii.intent_id = i.id AND ii.user_id = $1), '') AS interest_state,
		       EXISTS(SELECT 1 FROM intent_saves s WHERE s.intent_id = i.id AND s.user_id = $1) AS saved_by_me,
		       i.created_at, i.updated_at
		FROM connection_intents i
		JOIN users u ON u.id = i.author_id
		LEFT JOIN matching_profiles mp ON mp.user_id = i.author_id
		WHERE i.id = $2 AND u.deleted_at IS NULL
	`, viewerID, intentID)
	return scanConnectionIntent(row)
}

func (s *Store) ListMyConnectionIntents(ctx context.Context, tx *sql.Tx, userID int64, limit int) ([]ConnectionIntent, error) {
	return s.listConnectionIntents(ctx, tx, `
		WHERE i.author_id = $1 AND u.deleted_at IS NULL
		ORDER BY i.updated_at DESC, i.id DESC
		LIMIT $2
	`, userID, normalizeConnectionLimit(limit))
}

func (s *Store) ListFeedConnectionIntents(ctx context.Context, tx *sql.Tx, viewerID int64, category string, limit int) ([]ConnectionIntent, error) {
	return s.listConnectionIntents(ctx, tx, `
		WHERE i.author_id <> $1
		  AND i.status = 'active'
		  AND i.visibility = 'feed'
		  AND i.category <> 'romantic'
		  AND (i.expires_at IS NULL OR i.expires_at > NOW())
		  AND ($2 = '' OR i.category = $2)
		  AND COALESCE(mp.matching_paused, FALSE) = FALSE
		  AND NOT EXISTS (
			SELECT 1 FROM matching_profiles viewer_profile
			WHERE viewer_profile.user_id = $1
			  AND viewer_profile.hidden_categories ? i.category
		  )
		  AND NOT EXISTS (
			SELECT 1 FROM intent_dismissals d
			WHERE d.intent_id = i.id AND d.user_id = $1
		  )
		  AND u.deleted_at IS NULL
		  AND NOT EXISTS (
			SELECT 1
			FROM user_blocks b
			WHERE (b.blocker_id = $1 AND b.blocked_id = i.author_id)
			   OR (b.blocker_id = i.author_id AND b.blocked_id = $1)
		  )
		  AND NOT EXISTS (
			SELECT 1
			FROM safety_reports report
			WHERE report.status <> 'dismissed'
			  AND (
			    (report.reporter_id = $1 AND report.reported_user_id = i.author_id)
			    OR (report.reporter_id = i.author_id AND report.reported_user_id = $1)
			  )
		  )
		ORDER BY i.published_at DESC NULLS LAST, i.id DESC
		LIMIT $3
	`, viewerID, category, normalizeConnectionLimit(limit))
}

func (s *Store) ListSavedConnectionIntents(ctx context.Context, tx *sql.Tx, viewerID int64, category string, limit int) ([]ConnectionIntent, error) {
	return s.listConnectionIntents(ctx, tx, `
		JOIN intent_saves saved ON saved.intent_id = i.id
		WHERE saved.user_id = $1
		  AND i.author_id <> $1
		  AND i.status = 'active'
		  AND i.visibility <> 'private'
		  AND i.category <> 'romantic'
		  AND (i.expires_at IS NULL OR i.expires_at > NOW())
		  AND ($2 = '' OR i.category = $2)
		  AND COALESCE(mp.matching_paused, FALSE) = FALSE
		  AND u.deleted_at IS NULL
		  AND NOT EXISTS (
			SELECT 1
			FROM user_blocks b
			WHERE (b.blocker_id = $1 AND b.blocked_id = i.author_id)
			   OR (b.blocker_id = i.author_id AND b.blocked_id = $1)
		  )
		  AND NOT EXISTS (
			SELECT 1
			FROM safety_reports report
			WHERE report.status <> 'dismissed'
			  AND (
			    (report.reporter_id = $1 AND report.reported_user_id = i.author_id)
			    OR (report.reporter_id = i.author_id AND report.reported_user_id = $1)
			  )
		  )
		ORDER BY saved.created_at DESC, i.id DESC
		LIMIT $3
	`, viewerID, category, normalizeConnectionLimit(limit))
}

func (s *Store) ListRecommendationCandidates(ctx context.Context, tx *sql.Tx, viewerID int64, sourceCategory string, limit int) ([]ConnectionIntent, error) {
	return s.listConnectionIntents(ctx, tx, `
		WHERE i.author_id <> $1
		  AND i.status = 'active'
		  AND i.visibility IN ('candidates', 'feed')
		  AND (i.expires_at IS NULL OR i.expires_at > NOW())
		  AND (
		    ($2 = 'romantic' AND i.category = 'romantic')
		    OR ($2 <> 'romantic' AND i.category <> 'romantic')
		  )
		  AND COALESCE(mp.matching_paused, FALSE) = FALSE
		  AND (i.category <> 'romantic' OR COALESCE(mp.romantic_opt_in, FALSE))
		  AND NOT EXISTS (
			SELECT 1 FROM matching_profiles viewer_profile
			WHERE viewer_profile.user_id = $1
			  AND viewer_profile.hidden_categories ? i.category
		  )
		  AND NOT EXISTS (
			SELECT 1 FROM intent_dismissals d
			WHERE d.intent_id = i.id AND d.user_id = $1
		  )
		  AND u.deleted_at IS NULL
		  AND NOT EXISTS (
			SELECT 1
			FROM user_blocks b
			WHERE (b.blocker_id = $1 AND b.blocked_id = i.author_id)
			   OR (b.blocker_id = i.author_id AND b.blocked_id = $1)
		  )
		  AND NOT EXISTS (
			SELECT 1
			FROM safety_reports report
			WHERE report.status <> 'dismissed'
			  AND (
			    (report.reporter_id = $1 AND report.reported_user_id = i.author_id)
			    OR (report.reporter_id = i.author_id AND report.reported_user_id = $1)
			  )
		  )
		ORDER BY i.updated_at DESC, i.id DESC
		LIMIT $3
	`, viewerID, sourceCategory, normalizeConnectionLimit(limit))
}

func (s *Store) listConnectionIntents(ctx context.Context, tx *sql.Tx, clause string, args ...any) ([]ConnectionIntent, error) {
	query := `
		SELECT i.id, i.author_id, u.nickname, u.department, u.grade,
		       COALESCE(mp.bio, ''), COALESCE(mp.skills, '[]'::jsonb)::text,
		       COALESCE(mp.availability, ''), COALESCE(mp.collaboration_style, ''),
		       i.category, i.title, i.description, i.goal, i.time_window, i.location,
		       i.visibility, i.status, i.constraints::text, i.deal_breakers::text,
		       i.desired_next_step, i.project_memo::text, i.expires_at, i.published_at,
		       (SELECT COUNT(*) FROM intent_interests ii
		        WHERE ii.intent_id = i.id AND ii.status IN ('interested', 'accepted')) AS interest_count,
		       COALESCE((SELECT ii.status FROM intent_interests ii
		                 WHERE ii.intent_id = i.id AND ii.user_id = $1), '') AS interest_state,
		       EXISTS(SELECT 1 FROM intent_saves s WHERE s.intent_id = i.id AND s.user_id = $1) AS saved_by_me,
		       i.created_at, i.updated_at
		FROM connection_intents i
		JOIN users u ON u.id = i.author_id
		LEFT JOIN matching_profiles mp ON mp.user_id = i.author_id
	` + clause

	rows, err := queryContext(executor(s.db, tx), ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list connection intents: %w", err)
	}
	defer rows.Close()

	intents := make([]ConnectionIntent, 0)
	for rows.Next() {
		intent, err := scanConnectionIntent(rows)
		if err != nil {
			return nil, err
		}
		intents = append(intents, intent)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate connection intents: %w", err)
	}
	return intents, nil
}

func (s *Store) SetIntentSaved(ctx context.Context, tx *sql.Tx, intentID, userID int64, saved bool, now time.Time) error {
	if saved {
		_, err := executor(s.db, tx).ExecContext(ctx, `
			INSERT INTO intent_saves (intent_id, user_id, created_at)
			VALUES ($1, $2, $3)
			ON CONFLICT (intent_id, user_id) DO NOTHING
		`, intentID, userID, now)
		if err != nil {
			return fmt.Errorf("save intent: %w", err)
		}
		return nil
	}

	_, err := executor(s.db, tx).ExecContext(ctx, `
		DELETE FROM intent_saves WHERE intent_id = $1 AND user_id = $2
	`, intentID, userID)
	if err != nil {
		return fmt.Errorf("unsave intent: %w", err)
	}
	return nil
}

func (s *Store) DismissIntent(ctx context.Context, tx *sql.Tx, intentID, userID int64, now time.Time) error {
	_, err := executor(s.db, tx).ExecContext(ctx, `
		INSERT INTO intent_dismissals (intent_id, user_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (intent_id, user_id) DO NOTHING
	`, intentID, userID, now)
	if err != nil {
		return fmt.Errorf("dismiss intent: %w", err)
	}
	return nil
}

func (s *Store) SetIntentInterest(ctx context.Context, tx *sql.Tx, intentID, userID int64, status, message string, now time.Time) error {
	_, err := executor(s.db, tx).ExecContext(ctx, `
		INSERT INTO intent_interests (intent_id, user_id, status, message, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
		ON CONFLICT (intent_id, user_id) DO UPDATE SET
			status = EXCLUDED.status,
			message = EXCLUDED.message,
			updated_at = EXCLUDED.updated_at
	`, intentID, userID, status, message, now)
	if err != nil {
		return fmt.Errorf("set intent interest: %w", err)
	}
	return nil
}

func (s *Store) GetIntentInterestStatus(
	ctx context.Context,
	tx *sql.Tx,
	intentID int64,
	userID int64,
) (string, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT status
		FROM intent_interests
		WHERE intent_id = $1 AND user_id = $2
	`, intentID, userID)

	var status string
	if err := row.Scan(&status); err != nil {
		if err == sql.ErrNoRows {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("get intent interest status: %w", err)
	}
	return status, nil
}

func (s *Store) UpdateIntentInterestStatus(ctx context.Context, tx *sql.Tx, intentID, userID int64, status string, now time.Time) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE intent_interests
		SET status = $3, updated_at = $4
		WHERE intent_id = $1 AND user_id = $2 AND status = 'interested'
	`, intentID, userID, status, now)
	if err != nil {
		return fmt.Errorf("update intent interest: %w", err)
	}
	return requireRowsAffected(result)
}

func (s *Store) ListIntentInterests(ctx context.Context, tx *sql.Tx, intentID int64) ([]IntentInterest, error) {
	rows, err := queryContext(executor(s.db, tx), ctx, `
		SELECT ii.intent_id, ii.user_id, u.nickname, u.department, u.grade,
		       COALESCE(mp.bio, ''), COALESCE(mp.skills, '[]'::jsonb)::text,
		       COALESCE(mp.collaboration_style, ''), ii.status, ii.message,
		       ii.created_at, ii.updated_at
		FROM intent_interests ii
		JOIN users u ON u.id = ii.user_id
		LEFT JOIN matching_profiles mp ON mp.user_id = ii.user_id
		WHERE ii.intent_id = $1
		  AND ii.status <> 'withdrawn'
		  AND u.deleted_at IS NULL
		ORDER BY
		  CASE ii.status WHEN 'interested' THEN 0 WHEN 'accepted' THEN 1 ELSE 2 END,
		  ii.updated_at DESC
	`, intentID)
	if err != nil {
		return nil, fmt.Errorf("list intent interests: %w", err)
	}
	defer rows.Close()

	interests := make([]IntentInterest, 0)
	for rows.Next() {
		var interest IntentInterest
		var rawSkills string
		if err := rows.Scan(
			&interest.IntentID, &interest.UserID, &interest.Nickname,
			&interest.Department, &interest.Grade, &interest.Bio, &rawSkills,
			&interest.CollaborationStyle, &interest.Status, &interest.Message,
			&interest.CreatedAt, &interest.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan intent interest: %w", err)
		}
		if err := unmarshalStringList(rawSkills, &interest.Skills); err != nil {
			return nil, err
		}
		interests = append(interests, interest)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate intent interests: %w", err)
	}
	return interests, nil
}

func (s *Store) CreateMatchRoom(ctx context.Context, tx *sql.Tx, params CreateMatchRoomParams) (int64, error) {
	sharedContext, err := marshalStringList(params.SharedContext)
	if err != nil {
		return 0, err
	}
	starters, err := marshalStringList(params.ConversationStarters)
	if err != nil {
		return 0, err
	}
	places, err := marshalStringList(params.SuggestedPlaces)
	if err != nil {
		return 0, err
	}

	row := executor(s.db, tx).QueryRowContext(ctx, `
		INSERT INTO match_rooms (
			intent_id, owner_id, candidate_id, status, fit_reason, uncertainty,
			shared_context, conversation_starters, suggested_places,
			assistant_source, assistant_model, assistant_confidence,
			assistant_prompt_version, created_at, updated_at
		) VALUES (
			$1, $2, $3, 'open', $4, $5, $6::jsonb, $7::jsonb, $8::jsonb,
			$9, $10, $11, $12, $13, $13
		)
		ON CONFLICT (intent_id, candidate_id) DO UPDATE SET updated_at = EXCLUDED.updated_at
		RETURNING id
	`, params.IntentID, params.OwnerID, params.CandidateID, params.FitReason,
		params.Uncertainty, sharedContext, starters, places, params.AssistantSource,
		params.AssistantModel, params.AssistantConfidence, params.AssistantPrompt, params.Now)

	var roomID int64
	if err := row.Scan(&roomID); err != nil {
		return 0, fmt.Errorf("create match room: %w", err)
	}
	return roomID, nil
}

func (s *Store) FindMatchRoomID(
	ctx context.Context,
	tx *sql.Tx,
	intentID int64,
	candidateID int64,
) (int64, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT id
		FROM match_rooms
		WHERE intent_id = $1 AND candidate_id = $2
		LIMIT 1
	`, intentID, candidateID)

	var roomID int64
	if err := row.Scan(&roomID); err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrNotFound
		}
		return 0, fmt.Errorf("find match room: %w", err)
	}
	return roomID, nil
}

func (s *Store) GetMatchRoom(ctx context.Context, tx *sql.Tx, viewerID, roomID int64) (MatchRoom, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, matchRoomSelect+`
		WHERE r.id = $2 AND ($1 = r.owner_id OR $1 = r.candidate_id)
	`, viewerID, roomID)
	return scanMatchRoom(row)
}

func (s *Store) ListMatchRooms(ctx context.Context, tx *sql.Tx, viewerID int64, limit int) ([]MatchRoom, error) {
	rows, err := queryContext(executor(s.db, tx), ctx, matchRoomSelect+`
		WHERE $1 = r.owner_id OR $1 = r.candidate_id
		ORDER BY r.updated_at DESC, r.id DESC
		LIMIT $2
	`, viewerID, normalizeConnectionLimit(limit))
	if err != nil {
		return nil, fmt.Errorf("list match rooms: %w", err)
	}
	defer rows.Close()

	rooms := make([]MatchRoom, 0)
	for rows.Next() {
		room, err := scanMatchRoom(rows)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate match rooms: %w", err)
	}
	return rooms, nil
}

func (s *Store) UpdateMatchRoomPlan(ctx context.Context, tx *sql.Tx, viewerID, roomID int64, scheduledAt sql.NullTime, place string, now time.Time) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE match_rooms
		SET scheduled_at = $3, selected_place = $4, updated_at = $5
		WHERE id = $2 AND ($1 = owner_id OR $1 = candidate_id) AND status = 'open'
	`, viewerID, roomID, nullableTime(scheduledAt), place, now)
	if err != nil {
		return fmt.Errorf("update match room plan: %w", err)
	}
	return requireRowsAffected(result)
}

func (s *Store) UpdateMatchRoomStatus(ctx context.Context, tx *sql.Tx, viewerID, roomID int64, status string, now time.Time) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE match_rooms
		SET status = CAST($3 AS VARCHAR(24)),
		    owner_contact_shared = CASE WHEN CAST($3 AS VARCHAR(24)) = 'closed' THEN FALSE ELSE owner_contact_shared END,
		    candidate_contact_shared = CASE WHEN CAST($3 AS VARCHAR(24)) = 'closed' THEN FALSE ELSE candidate_contact_shared END,
		    updated_at = $4
		WHERE id = $2
		  AND ($1 = owner_id OR $1 = candidate_id)
		  AND status <> 'closed'
	`, viewerID, roomID, status, now)
	if err != nil {
		return fmt.Errorf("update match room status: %w", err)
	}
	return requireRowsAffected(result)
}

func (s *Store) SetMatchRoomContactShared(ctx context.Context, tx *sql.Tx, viewerID, roomID int64, shared bool, now time.Time) error {
	room, err := s.GetMatchRoom(ctx, tx, viewerID, roomID)
	if err != nil {
		return err
	}

	var result sql.Result
	if viewerID == room.OwnerID {
		result, err = executor(s.db, tx).ExecContext(ctx, `
			UPDATE match_rooms SET owner_contact_shared = $3, updated_at = $4
			WHERE id = $2 AND owner_id = $1 AND status <> 'closed'
		`, viewerID, roomID, shared, now)
	} else {
		result, err = executor(s.db, tx).ExecContext(ctx, `
			UPDATE match_rooms SET candidate_contact_shared = $3, updated_at = $4
			WHERE id = $2 AND candidate_id = $1 AND status <> 'closed'
		`, viewerID, roomID, shared, now)
	}
	if err != nil {
		return fmt.Errorf("update match room contact sharing: %w", err)
	}
	return requireRowsAffected(result)
}

func (s *Store) CreateMatchRoomMessage(ctx context.Context, tx *sql.Tx, roomID, senderID int64, content string, now time.Time) (MatchRoomMessage, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		INSERT INTO match_room_messages (room_id, sender_id, content, created_at)
		SELECT r.id, $2, $3, $4
		FROM match_rooms r
		WHERE r.id = $1
		  AND ($2 = r.owner_id OR $2 = r.candidate_id)
		  AND r.status IN ('open', 'met')
		RETURNING id, room_id, sender_id, content, created_at
	`, roomID, senderID, content, now)

	var message MatchRoomMessage
	if err := row.Scan(&message.ID, &message.RoomID, &message.SenderID, &message.Content, &message.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return MatchRoomMessage{}, ErrNotFound
		}
		return MatchRoomMessage{}, fmt.Errorf("create match room message: %w", err)
	}

	user, err := s.GetUserByID(ctx, tx, senderID)
	if err != nil {
		return MatchRoomMessage{}, err
	}
	message.SenderNickname = user.Nickname
	return message, nil
}

func (s *Store) ListMatchRoomMessages(ctx context.Context, tx *sql.Tx, roomID, viewerID int64, limit int) ([]MatchRoomMessage, error) {
	rows, err := queryContext(executor(s.db, tx), ctx, `
		SELECT m.id, m.room_id, m.sender_id, u.nickname, m.content, m.created_at
		FROM match_room_messages m
		JOIN match_rooms r ON r.id = m.room_id
		JOIN users u ON u.id = m.sender_id
		WHERE m.room_id = $1 AND ($2 = r.owner_id OR $2 = r.candidate_id)
		ORDER BY m.created_at ASC, m.id ASC
		LIMIT $3
	`, roomID, viewerID, normalizeMessageLimit(limit))
	if err != nil {
		return nil, fmt.Errorf("list match room messages: %w", err)
	}
	defer rows.Close()

	messages := make([]MatchRoomMessage, 0)
	for rows.Next() {
		var message MatchRoomMessage
		if err := rows.Scan(
			&message.ID, &message.RoomID, &message.SenderID, &message.SenderNickname,
			&message.Content, &message.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan match room message: %w", err)
		}
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate match room messages: %w", err)
	}
	return messages, nil
}

const matchRoomSelect = `
	SELECT r.id, r.intent_id, i.title, i.category,
	       r.owner_id, owner.nickname, owner.department, owner.grade,
	       r.candidate_id, candidate.nickname, candidate.department, candidate.grade,
	       r.status, r.fit_reason, r.uncertainty, r.shared_context::text,
	       r.conversation_starters::text, r.suggested_places::text,
	       r.assistant_source, r.assistant_model, r.assistant_confidence::float8,
	       r.assistant_prompt_version,
	       r.scheduled_at, r.selected_place,
	       r.owner_contact_shared, r.candidate_contact_shared,
	       COALESCE(owner.contact_type, ''), COALESCE(owner.contact_value, ''),
	       COALESCE(candidate.contact_type, ''), COALESCE(candidate.contact_value, ''),
	       r.created_at, r.updated_at
	FROM match_rooms r
	JOIN connection_intents i ON i.id = r.intent_id
	JOIN users owner ON owner.id = r.owner_id
	JOIN users candidate ON candidate.id = r.candidate_id
`

func scanMatchingProfile(row rowScanner) (MatchingProfile, error) {
	var profile MatchingProfile
	var rawSkills, rawInterests, rawScenes, rawHiddenCategories string
	if err := row.Scan(
		&profile.UserID, &profile.Bio, &rawSkills, &rawInterests, &profile.Availability,
		&rawScenes, &profile.CollaborationStyle, &profile.Boundaries, &profile.AIPortrait,
		&profile.AIPortraitApproved, &profile.RomanticOptIn, &profile.MatchingPaused,
		&rawHiddenCategories, &profile.CreatedAt, &profile.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return MatchingProfile{}, ErrNotFound
		}
		return MatchingProfile{}, fmt.Errorf("scan matching profile: %w", err)
	}
	if err := unmarshalStringList(rawSkills, &profile.Skills); err != nil {
		return MatchingProfile{}, err
	}
	if err := unmarshalStringList(rawInterests, &profile.Interests); err != nil {
		return MatchingProfile{}, err
	}
	if err := unmarshalStringList(rawScenes, &profile.PreferredScenes); err != nil {
		return MatchingProfile{}, err
	}
	if err := unmarshalStringList(rawHiddenCategories, &profile.HiddenCategories); err != nil {
		return MatchingProfile{}, err
	}
	return profile, nil
}

func scanConnectionIntent(row rowScanner) (ConnectionIntent, error) {
	var intent ConnectionIntent
	var rawAuthorSkills, rawConstraints, rawDealBreakers, rawProjectMemo string
	if err := row.Scan(
		&intent.ID, &intent.AuthorID, &intent.AuthorNickname, &intent.AuthorDepartment,
		&intent.AuthorGrade, &intent.AuthorBio, &rawAuthorSkills,
		&intent.AuthorAvailability, &intent.AuthorCollaborationStyle,
		&intent.Category, &intent.Title, &intent.Description,
		&intent.Goal, &intent.TimeWindow, &intent.Location, &intent.Visibility,
		&intent.Status, &rawConstraints, &rawDealBreakers, &intent.DesiredNextStep,
		&rawProjectMemo, &intent.ExpiresAt, &intent.PublishedAt, &intent.InterestCount,
		&intent.InterestState, &intent.SavedByMe, &intent.CreatedAt, &intent.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return ConnectionIntent{}, ErrNotFound
		}
		return ConnectionIntent{}, fmt.Errorf("scan connection intent: %w", err)
	}
	if err := unmarshalStringList(rawConstraints, &intent.Constraints); err != nil {
		return ConnectionIntent{}, err
	}
	if err := unmarshalStringList(rawAuthorSkills, &intent.AuthorSkills); err != nil {
		return ConnectionIntent{}, err
	}
	if err := unmarshalStringList(rawDealBreakers, &intent.DealBreakers); err != nil {
		return ConnectionIntent{}, err
	}
	if err := json.Unmarshal([]byte(rawProjectMemo), &intent.ProjectMemo); err != nil {
		return ConnectionIntent{}, fmt.Errorf("unmarshal project memo: %w", err)
	}
	intent.ProjectMemo.RequiredSkills = nonNilStringList(intent.ProjectMemo.RequiredSkills)
	intent.ProjectMemo.NiceToHaveSkills = nonNilStringList(intent.ProjectMemo.NiceToHaveSkills)
	return intent, nil
}

func scanMatchRoom(row rowScanner) (MatchRoom, error) {
	var room MatchRoom
	var rawContext, rawStarters, rawPlaces string
	if err := row.Scan(
		&room.ID, &room.IntentID, &room.IntentTitle, &room.IntentCategory,
		&room.OwnerID, &room.OwnerNickname, &room.OwnerDepartment, &room.OwnerGrade,
		&room.CandidateID, &room.CandidateNickname, &room.CandidateDepartment,
		&room.CandidateGrade, &room.Status, &room.FitReason, &room.Uncertainty,
		&rawContext, &rawStarters, &rawPlaces, &room.AssistantSource,
		&room.AssistantModel, &room.AssistantConfidence, &room.AssistantPromptVersion,
		&room.ScheduledAt, &room.SelectedPlace,
		&room.OwnerContactShared, &room.CandidateContactShared,
		&room.OwnerContactType, &room.OwnerContactValue,
		&room.CandidateContactType, &room.CandidateContactValue,
		&room.CreatedAt, &room.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return MatchRoom{}, ErrNotFound
		}
		return MatchRoom{}, fmt.Errorf("scan match room: %w", err)
	}
	if err := unmarshalStringList(rawContext, &room.SharedContext); err != nil {
		return MatchRoom{}, err
	}
	if err := unmarshalStringList(rawStarters, &room.ConversationStarters); err != nil {
		return MatchRoom{}, err
	}
	if err := unmarshalStringList(rawPlaces, &room.SuggestedPlaces); err != nil {
		return MatchRoom{}, err
	}
	return room, nil
}

func marshalStringList(values []string) (string, error) {
	if values == nil {
		values = []string{}
	}
	raw, err := json.Marshal(values)
	if err != nil {
		return "", fmt.Errorf("marshal string list: %w", err)
	}
	return string(raw), nil
}

func unmarshalStringList(raw string, target *[]string) error {
	if err := json.Unmarshal([]byte(raw), target); err != nil {
		return fmt.Errorf("unmarshal string list: %w", err)
	}
	if *target == nil {
		*target = []string{}
	}
	return nil
}

func marshalProjectMemo(memo ProjectMemo) (string, error) {
	memo.RequiredSkills = nonNilStringList(memo.RequiredSkills)
	memo.NiceToHaveSkills = nonNilStringList(memo.NiceToHaveSkills)
	raw, err := json.Marshal(memo)
	if err != nil {
		return "", fmt.Errorf("marshal project memo: %w", err)
	}
	return string(raw), nil
}

func nonNilStringList(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func nullableTime(value sql.NullTime) any {
	if value.Valid {
		return value.Time
	}
	return nil
}

func queryContext(target queryer, ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	querier, ok := target.(interface {
		QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	})
	if !ok {
		return nil, fmt.Errorf("store executor does not support QueryContext")
	}
	return querier.QueryContext(ctx, query, args...)
}

func normalizeConnectionLimit(limit int) int {
	if limit <= 0 {
		return 30
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func normalizeMessageLimit(limit int) int {
	if limit <= 0 {
		return 100
	}
	if limit > 300 {
		return 300
	}
	return limit
}
