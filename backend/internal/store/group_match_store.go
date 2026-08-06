package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type GroupMatchIntent struct {
	ID                   int64
	OrganizerID          int64
	OrganizerNickname    string
	OrganizerDepartment  string
	OrganizerGrade       string
	Category             string
	Title                string
	Description          string
	RoleOffer            string
	RolesNeeded          []string
	Constraints          []string
	TimeWindow           string
	Location             string
	MinMembers           int
	MaxMembers           int
	StartsAt             sql.NullTime
	EndsAt               sql.NullTime
	CreateCalendarEvent  bool
	Status               string
	ChatGroupID          sql.NullInt64
	CalendarEventID      sql.NullInt64
	GroupIntro           string
	ConversationStarters []string
	InferenceSource      string
	InferenceModel       string
	InferenceConfidence  float64
	PromptVersion        string
	ExpiresAt            time.Time
	FormedAt             sql.NullTime
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type CreateGroupMatchIntentParams struct {
	OrganizerID         int64
	Category            string
	Title               string
	Description         string
	RoleOffer           string
	RolesNeeded         []string
	Constraints         []string
	TimeWindow          string
	Location            string
	MinMembers          int
	MaxMembers          int
	StartsAt            sql.NullTime
	EndsAt              sql.NullTime
	CreateCalendarEvent bool
	ExpiresAt           time.Time
	Now                 time.Time
}

type GroupMatchCandidate struct {
	IntentID           int64
	UserID             int64
	Nickname           string
	Department         string
	Grade              string
	Category           string
	Title              string
	Description        string
	RoleOffer          string
	RolesNeeded        []string
	Constraints        []string
	TimeWindow         string
	Location           string
	MinMembers         int
	MaxMembers         int
	Bio                string
	Skills             []string
	Interests          []string
	Availability       string
	PreferredScenes    []string
	CollaborationStyle string
	ExpiresAt          time.Time
	UpdatedAt          time.Time
}

type GroupMatchInvitation struct {
	ID                        int64
	GroupIntentID             int64
	CandidateIntentID         int64
	CandidateUserID           int64
	CandidateNickname         string
	CandidateDepartment       string
	CandidateGrade            string
	CandidateRoleOffer        string
	Score                     int
	FitReason                 string
	Uncertainty               string
	SharedConstraints         []string
	SuggestedRole             string
	Status                    string
	RespondedAt               sql.NullTime
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
	GroupOrganizerID          int64
	GroupOrganizerNickname    string
	GroupCategory             string
	GroupTitle                string
	GroupDescription          string
	GroupTimeWindow           string
	GroupLocation             string
	GroupMinMembers           int
	GroupMaxMembers           int
	GroupStatus               string
	GroupChatID               sql.NullInt64
	GroupCalendarEventID      sql.NullInt64
	GroupIntro                string
	GroupConversationStarters []string
	GroupExpiresAt            time.Time
}

type CreateGroupMatchInvitationParams struct {
	GroupIntentID     int64
	CandidateIntentID int64
	CandidateUserID   int64
	Score             int
	FitReason         string
	Uncertainty       string
	SharedConstraints []string
	SuggestedRole     string
	Now               time.Time
}

type GroupMatchCounts struct {
	Accepted   int
	Pending    int
	Waitlisted int
	Declined   int
}

func (s *Store) CreateGroupMatchIntent(
	ctx context.Context,
	tx *sql.Tx,
	params CreateGroupMatchIntentParams,
) (int64, error) {
	rolesNeeded, err := marshalStringList(params.RolesNeeded)
	if err != nil {
		return 0, err
	}
	constraints, err := marshalStringList(params.Constraints)
	if err != nil {
		return 0, err
	}
	row := executor(s.db, tx).QueryRowContext(ctx, `
		INSERT INTO group_match_intents (
			organizer_id, category, title, description, role_offer, roles_needed,
			constraints, time_window, location, min_members, max_members,
			starts_at, ends_at, create_calendar_event, expires_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6::jsonb, $7::jsonb, $8, $9, $10, $11,
			$12, $13, $14, $15, $16, $16
		)
		RETURNING id
	`, params.OrganizerID, params.Category, params.Title, params.Description,
		params.RoleOffer, rolesNeeded, constraints, params.TimeWindow, params.Location,
		params.MinMembers, params.MaxMembers, nullableTime(params.StartsAt),
		nullableTime(params.EndsAt), params.CreateCalendarEvent, params.ExpiresAt, params.Now)
	var intentID int64
	if err := row.Scan(&intentID); err != nil {
		return 0, fmt.Errorf("create group match intent: %w", err)
	}
	return intentID, nil
}

func (s *Store) GetGroupMatchIntent(
	ctx context.Context,
	tx *sql.Tx,
	intentID int64,
) (GroupMatchIntent, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, groupMatchIntentSelect+`
		WHERE intent.id = $1 AND organizer.deleted_at IS NULL
	`, intentID)
	return scanGroupMatchIntent(row)
}

func (s *Store) GetGroupMatchIntentForUpdate(
	ctx context.Context,
	tx *sql.Tx,
	intentID int64,
) (GroupMatchIntent, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, groupMatchIntentSelect+`
		WHERE intent.id = $1 AND organizer.deleted_at IS NULL
		FOR UPDATE OF intent
	`, intentID)
	return scanGroupMatchIntent(row)
}

func (s *Store) ListGroupMatchIntentsForUpdate(
	ctx context.Context,
	tx *sql.Tx,
	intentIDs []int64,
) ([]GroupMatchIntent, error) {
	if len(intentIDs) == 0 {
		return []GroupMatchIntent{}, nil
	}
	rows, err := queryContext(executor(s.db, tx), ctx, groupMatchIntentSelect+`
		WHERE intent.id = ANY($1::bigint[]) AND organizer.deleted_at IS NULL
		ORDER BY intent.id
		FOR UPDATE OF intent
	`, intentIDs)
	if err != nil {
		return nil, fmt.Errorf("lock group match intents: %w", err)
	}
	defer rows.Close()
	result := make([]GroupMatchIntent, 0, len(intentIDs))
	for rows.Next() {
		item, err := scanGroupMatchIntent(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate locked group match intents: %w", err)
	}
	return result, nil
}

func (s *Store) ListGroupMatchIntentsByOwner(
	ctx context.Context,
	tx *sql.Tx,
	ownerID int64,
	limit int,
) ([]GroupMatchIntent, error) {
	rows, err := queryContext(executor(s.db, tx), ctx, groupMatchIntentSelect+`
		WHERE intent.organizer_id = $1 AND organizer.deleted_at IS NULL
		ORDER BY intent.updated_at DESC, intent.id DESC
		LIMIT $2
	`, ownerID, normalizeConnectionLimit(limit))
	if err != nil {
		return nil, fmt.Errorf("list group match intents: %w", err)
	}
	defer rows.Close()
	result := make([]GroupMatchIntent, 0)
	for rows.Next() {
		item, err := scanGroupMatchIntent(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate group match intents: %w", err)
	}
	return result, nil
}

func (s *Store) ListGroupMatchCandidates(
	ctx context.Context,
	tx *sql.Tx,
	groupIntentID int64,
	organizerID int64,
	category string,
	limit int,
) ([]GroupMatchCandidate, error) {
	rows, err := queryContext(executor(s.db, tx), ctx, `
		SELECT candidate.id, candidate.organizer_id, candidate_user.nickname,
		       candidate_user.department, candidate_user.grade, candidate.category,
		       candidate.title, candidate.description, candidate.role_offer,
		       candidate.roles_needed::text, candidate.constraints::text,
		       candidate.time_window, candidate.location,
		       candidate.min_members, candidate.max_members,
		       COALESCE(profile.bio, ''), COALESCE(profile.skills, '[]'::jsonb)::text,
		       COALESCE(profile.interests, '[]'::jsonb)::text,
		       COALESCE(profile.availability, ''),
		       COALESCE(profile.preferred_scenes, '[]'::jsonb)::text,
		       COALESCE(profile.collaboration_style, ''),
		       candidate.expires_at, candidate.updated_at
		FROM group_match_intents candidate
		JOIN users candidate_user ON candidate_user.id = candidate.organizer_id
		LEFT JOIN matching_profiles profile ON profile.user_id = candidate.organizer_id
		WHERE candidate.organizer_id <> $2
		  AND candidate.category = $3
		  AND candidate.status = 'open'
		  AND candidate.expires_at > NOW()
		  AND candidate_user.deleted_at IS NULL
		  AND COALESCE(profile.matching_paused, FALSE) = FALSE
		  AND NOT EXISTS (
			SELECT 1 FROM group_match_invitations invitation
			WHERE invitation.group_intent_id = $1
			  AND invitation.candidate_user_id = candidate.organizer_id
		  )
		  AND NOT EXISTS (
			SELECT 1 FROM user_blocks block
			WHERE (block.blocker_id = $2 AND block.blocked_id = candidate.organizer_id)
			   OR (block.blocker_id = candidate.organizer_id AND block.blocked_id = $2)
		  )
		  AND NOT EXISTS (
			SELECT 1 FROM safety_reports report
			WHERE report.status IN ('open', 'reviewing')
			  AND ((report.reporter_id = $2 AND report.reported_user_id = candidate.organizer_id)
			    OR (report.reporter_id = candidate.organizer_id AND report.reported_user_id = $2))
		  )
		ORDER BY candidate.updated_at DESC, candidate.id DESC
		LIMIT $4
	`, groupIntentID, organizerID, category, normalizeConnectionLimit(limit))
	if err != nil {
		return nil, fmt.Errorf("list group match candidates: %w", err)
	}
	defer rows.Close()
	result := make([]GroupMatchCandidate, 0)
	for rows.Next() {
		item, err := scanGroupMatchCandidate(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate group match candidates: %w", err)
	}
	return result, nil
}

func (s *Store) UpsertGroupMatchInvitation(
	ctx context.Context,
	tx *sql.Tx,
	params CreateGroupMatchInvitationParams,
) (GroupMatchInvitation, error) {
	shared, err := marshalStringList(params.SharedConstraints)
	if err != nil {
		return GroupMatchInvitation{}, err
	}
	row := executor(s.db, tx).QueryRowContext(ctx, `
		INSERT INTO group_match_invitations (
			group_intent_id, candidate_intent_id, candidate_user_id, score,
			fit_reason, uncertainty, shared_constraints, suggested_role,
			status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8, 'invited', $9, $9)
		ON CONFLICT (group_intent_id, candidate_user_id) DO UPDATE SET
			candidate_intent_id = EXCLUDED.candidate_intent_id,
			score = EXCLUDED.score,
			fit_reason = EXCLUDED.fit_reason,
			uncertainty = EXCLUDED.uncertainty,
			shared_constraints = EXCLUDED.shared_constraints,
			suggested_role = EXCLUDED.suggested_role,
			status = CASE
				WHEN group_match_invitations.status IN ('declined', 'left') THEN 'invited'
				ELSE group_match_invitations.status
			END,
			responded_at = CASE
				WHEN group_match_invitations.status IN ('declined', 'left') THEN NULL
				ELSE group_match_invitations.responded_at
			END,
			updated_at = EXCLUDED.updated_at
		RETURNING id
	`, params.GroupIntentID, params.CandidateIntentID, params.CandidateUserID,
		params.Score, params.FitReason, params.Uncertainty, shared,
		params.SuggestedRole, params.Now)
	var invitationID int64
	if err := row.Scan(&invitationID); err != nil {
		return GroupMatchInvitation{}, fmt.Errorf("upsert group match invitation: %w", err)
	}
	return s.GetGroupMatchInvitation(ctx, tx, invitationID)
}

func (s *Store) GetGroupMatchInvitation(
	ctx context.Context,
	tx *sql.Tx,
	invitationID int64,
) (GroupMatchInvitation, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, groupMatchInvitationSelect+`
		WHERE invitation.id = $1
	`, invitationID)
	return scanGroupMatchInvitation(row)
}

func (s *Store) GetGroupMatchInvitationForUpdate(
	ctx context.Context,
	tx *sql.Tx,
	invitationID int64,
) (GroupMatchInvitation, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, groupMatchInvitationSelect+`
		WHERE invitation.id = $1
		FOR UPDATE OF invitation
	`, invitationID)
	return scanGroupMatchInvitation(row)
}

func (s *Store) ListGroupMatchInvitationsForIntent(
	ctx context.Context,
	tx *sql.Tx,
	groupIntentID int64,
) ([]GroupMatchInvitation, error) {
	return s.listGroupMatchInvitations(ctx, tx, `
		WHERE invitation.group_intent_id = $1
		ORDER BY invitation.created_at, invitation.id
	`, groupIntentID)
}

func (s *Store) ListGroupMatchInvitationsForCandidate(
	ctx context.Context,
	tx *sql.Tx,
	candidateUserID int64,
	limit int,
) ([]GroupMatchInvitation, error) {
	return s.listGroupMatchInvitations(ctx, tx, `
		WHERE invitation.candidate_user_id = $1
		ORDER BY CASE invitation.status WHEN 'invited' THEN 0 ELSE 1 END,
		         invitation.updated_at DESC, invitation.id DESC
		LIMIT $2
	`, candidateUserID, normalizeConnectionLimit(limit))
}

func (s *Store) listGroupMatchInvitations(
	ctx context.Context,
	tx *sql.Tx,
	clause string,
	args ...any,
) ([]GroupMatchInvitation, error) {
	rows, err := queryContext(executor(s.db, tx), ctx, groupMatchInvitationSelect+clause, args...)
	if err != nil {
		return nil, fmt.Errorf("list group match invitations: %w", err)
	}
	defer rows.Close()
	result := make([]GroupMatchInvitation, 0)
	for rows.Next() {
		item, err := scanGroupMatchInvitation(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate group match invitations: %w", err)
	}
	return result, nil
}

func (s *Store) GetGroupMatchCounts(
	ctx context.Context,
	tx *sql.Tx,
	groupIntentID int64,
) (GroupMatchCounts, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT COUNT(*) FILTER (WHERE status = 'accepted'),
		       COUNT(*) FILTER (WHERE status = 'invited'),
		       COUNT(*) FILTER (WHERE status = 'waitlisted'),
		       COUNT(*) FILTER (WHERE status = 'declined')
		FROM group_match_invitations
		WHERE group_intent_id = $1
	`, groupIntentID)
	var counts GroupMatchCounts
	if err := row.Scan(&counts.Accepted, &counts.Pending, &counts.Waitlisted, &counts.Declined); err != nil {
		return GroupMatchCounts{}, fmt.Errorf("count group match invitations: %w", err)
	}
	return counts, nil
}

func (s *Store) UpdateGroupMatchInvitationStatus(
	ctx context.Context,
	tx *sql.Tx,
	invitationID int64,
	status string,
	now time.Time,
) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE group_match_invitations
		SET status = $2, responded_at = $3, updated_at = $3
		WHERE id = $1
	`, invitationID, status, now)
	if err != nil {
		return fmt.Errorf("update group match invitation status: %w", err)
	}
	return requireRowsAffected(result)
}

func (s *Store) UpdateGroupMatchIntentStatus(
	ctx context.Context,
	tx *sql.Tx,
	intentID int64,
	ownerID int64,
	status string,
	now time.Time,
) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE group_match_intents
		SET status = $3, updated_at = $4
		WHERE id = $1 AND organizer_id = $2
	`, intentID, ownerID, status, now)
	if err != nil {
		return fmt.Errorf("update group match intent status: %w", err)
	}
	return requireRowsAffected(result)
}

func (s *Store) ExpireGroupMatchIntents(ctx context.Context, tx *sql.Tx, now time.Time) error {
	_, err := executor(s.db, tx).ExecContext(ctx, `
		WITH expired AS (
			UPDATE group_match_intents
			SET status = 'expired', updated_at = $1
			WHERE status = 'open' AND expires_at <= $1
			RETURNING id
		), released AS (
			UPDATE group_match_intents candidate
			SET status = CASE WHEN candidate.expires_at > $1 THEN 'open' ELSE 'expired' END,
			    updated_at = $1
			FROM group_match_invitations invitation
			WHERE invitation.group_intent_id IN (SELECT id FROM expired)
			  AND invitation.candidate_intent_id = candidate.id
			  AND invitation.status = 'accepted'
			  AND candidate.status = 'joined'
			RETURNING candidate.id
		)
		UPDATE group_match_invitations
		SET status = 'left', responded_at = $1, updated_at = $1
		WHERE group_intent_id IN (SELECT id FROM expired)
		  AND status IN ('invited', 'accepted', 'waitlisted')
	`, now)
	if err != nil {
		return fmt.Errorf("expire group match intents: %w", err)
	}
	return nil
}

func (s *Store) FormGroupMatchIntent(
	ctx context.Context,
	tx *sql.Tx,
	intentID int64,
	chatGroupID int64,
	calendarEventID sql.NullInt64,
	intro string,
	starters []string,
	inferenceSource string,
	inferenceModel string,
	inferenceConfidence float64,
	promptVersion string,
	now time.Time,
) error {
	rawStarters, err := marshalStringList(starters)
	if err != nil {
		return err
	}
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE group_match_intents
		SET status = 'formed', chat_group_id = $2, calendar_event_id = $3,
		    group_intro = $4, conversation_starters = $5::jsonb,
		    inference_source = $6, inference_model = $7,
		    inference_confidence = $8, prompt_version = $9,
		    formed_at = $10, updated_at = $10
		WHERE id = $1 AND status = 'open'
	`, intentID, chatGroupID, nullableInt64Value(calendarEventID), intro, rawStarters,
		inferenceSource, inferenceModel, inferenceConfidence, promptVersion, now)
	if err != nil {
		return fmt.Errorf("form group match intent: %w", err)
	}
	return requireRowsAffected(result)
}

func (s *Store) DeclineOtherGroupMatchInvitations(
	ctx context.Context,
	tx *sql.Tx,
	candidateUserID int64,
	acceptedInvitationID int64,
	now time.Time,
) error {
	_, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE group_match_invitations
		SET status = 'declined', responded_at = $3, updated_at = $3
		WHERE candidate_user_id = $1 AND id <> $2 AND status = 'invited'
	`, candidateUserID, acceptedInvitationID, now)
	if err != nil {
		return fmt.Errorf("decline other group match invitations: %w", err)
	}
	return nil
}

func (s *Store) GetOldestWaitlistedGroupMatchInvitation(
	ctx context.Context,
	tx *sql.Tx,
	groupIntentID int64,
) (GroupMatchInvitation, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, groupMatchInvitationSelect+`
		WHERE invitation.group_intent_id = $1 AND invitation.status = 'waitlisted'
		ORDER BY invitation.responded_at, invitation.id
		LIMIT 1
		FOR UPDATE OF invitation
	`, groupIntentID)
	return scanGroupMatchInvitation(row)
}

const groupMatchIntentSelect = `
	SELECT intent.id, intent.organizer_id, organizer.nickname,
	       organizer.department, organizer.grade, intent.category, intent.title,
	       intent.description, intent.role_offer, intent.roles_needed::text,
	       intent.constraints::text, intent.time_window, intent.location,
	       intent.min_members, intent.max_members, intent.starts_at, intent.ends_at,
	       intent.create_calendar_event, intent.status, intent.chat_group_id,
	       intent.calendar_event_id, intent.group_intro,
	       intent.conversation_starters::text, intent.inference_source,
	       intent.inference_model, intent.inference_confidence,
	       intent.prompt_version, intent.expires_at, intent.formed_at,
	       intent.created_at, intent.updated_at
	FROM group_match_intents intent
	JOIN users organizer ON organizer.id = intent.organizer_id
`

const groupMatchInvitationSelect = `
	SELECT invitation.id, invitation.group_intent_id, invitation.candidate_intent_id,
	       invitation.candidate_user_id, candidate_user.nickname,
	       candidate_user.department, candidate_user.grade, candidate_intent.role_offer,
	       invitation.score, invitation.fit_reason, invitation.uncertainty,
	       invitation.shared_constraints::text, invitation.suggested_role,
	       invitation.status, invitation.responded_at, invitation.created_at,
	       invitation.updated_at, group_intent.organizer_id, organizer.nickname,
	       group_intent.category, group_intent.title, group_intent.description,
	       group_intent.time_window, group_intent.location,
	       group_intent.min_members, group_intent.max_members, group_intent.status,
	       group_intent.chat_group_id, group_intent.calendar_event_id,
	       group_intent.group_intro, group_intent.conversation_starters::text,
	       group_intent.expires_at
	FROM group_match_invitations invitation
	JOIN group_match_intents group_intent ON group_intent.id = invitation.group_intent_id
	JOIN group_match_intents candidate_intent ON candidate_intent.id = invitation.candidate_intent_id
	JOIN users candidate_user ON candidate_user.id = invitation.candidate_user_id
	JOIN users organizer ON organizer.id = group_intent.organizer_id
`

func scanGroupMatchIntent(row rowScanner) (GroupMatchIntent, error) {
	var item GroupMatchIntent
	var rolesNeeded string
	var constraints string
	var starters string
	if err := row.Scan(
		&item.ID, &item.OrganizerID, &item.OrganizerNickname,
		&item.OrganizerDepartment, &item.OrganizerGrade, &item.Category,
		&item.Title, &item.Description, &item.RoleOffer, &rolesNeeded,
		&constraints, &item.TimeWindow, &item.Location, &item.MinMembers,
		&item.MaxMembers, &item.StartsAt, &item.EndsAt, &item.CreateCalendarEvent,
		&item.Status, &item.ChatGroupID, &item.CalendarEventID, &item.GroupIntro,
		&starters, &item.InferenceSource, &item.InferenceModel,
		&item.InferenceConfidence, &item.PromptVersion, &item.ExpiresAt,
		&item.FormedAt, &item.CreatedAt, &item.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return GroupMatchIntent{}, ErrNotFound
		}
		return GroupMatchIntent{}, fmt.Errorf("scan group match intent: %w", err)
	}
	if err := unmarshalGroupStringLists(
		[]string{rolesNeeded, constraints, starters},
		[]*[]string{&item.RolesNeeded, &item.Constraints, &item.ConversationStarters},
	); err != nil {
		return GroupMatchIntent{}, err
	}
	return item, nil
}

func scanGroupMatchCandidate(row rowScanner) (GroupMatchCandidate, error) {
	var item GroupMatchCandidate
	var rolesNeeded string
	var constraints string
	var skills string
	var interests string
	var scenes string
	if err := row.Scan(
		&item.IntentID, &item.UserID, &item.Nickname, &item.Department,
		&item.Grade, &item.Category, &item.Title, &item.Description,
		&item.RoleOffer, &rolesNeeded, &constraints, &item.TimeWindow,
		&item.Location, &item.MinMembers, &item.MaxMembers, &item.Bio,
		&skills, &interests, &item.Availability,
		&scenes, &item.CollaborationStyle, &item.ExpiresAt, &item.UpdatedAt,
	); err != nil {
		return GroupMatchCandidate{}, fmt.Errorf("scan group match candidate: %w", err)
	}
	if err := unmarshalGroupStringLists(
		[]string{rolesNeeded, constraints, skills, interests, scenes},
		[]*[]string{&item.RolesNeeded, &item.Constraints, &item.Skills, &item.Interests, &item.PreferredScenes},
	); err != nil {
		return GroupMatchCandidate{}, err
	}
	return item, nil
}

func scanGroupMatchInvitation(row rowScanner) (GroupMatchInvitation, error) {
	var item GroupMatchInvitation
	var shared string
	var starters string
	if err := row.Scan(
		&item.ID, &item.GroupIntentID, &item.CandidateIntentID,
		&item.CandidateUserID, &item.CandidateNickname,
		&item.CandidateDepartment, &item.CandidateGrade,
		&item.CandidateRoleOffer, &item.Score, &item.FitReason,
		&item.Uncertainty, &shared, &item.SuggestedRole, &item.Status,
		&item.RespondedAt, &item.CreatedAt, &item.UpdatedAt,
		&item.GroupOrganizerID, &item.GroupOrganizerNickname,
		&item.GroupCategory, &item.GroupTitle, &item.GroupDescription,
		&item.GroupTimeWindow, &item.GroupLocation, &item.GroupMinMembers,
		&item.GroupMaxMembers, &item.GroupStatus, &item.GroupChatID,
		&item.GroupCalendarEventID, &item.GroupIntro, &starters,
		&item.GroupExpiresAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return GroupMatchInvitation{}, ErrNotFound
		}
		return GroupMatchInvitation{}, fmt.Errorf("scan group match invitation: %w", err)
	}
	if err := unmarshalGroupStringLists(
		[]string{shared, starters},
		[]*[]string{&item.SharedConstraints, &item.GroupConversationStarters},
	); err != nil {
		return GroupMatchInvitation{}, err
	}
	return item, nil
}

func unmarshalGroupStringLists(raw []string, targets []*[]string) error {
	for index := range raw {
		if err := json.Unmarshal([]byte(raw[index]), targets[index]); err != nil {
			return fmt.Errorf("decode group match list: %w", err)
		}
		if *targets[index] == nil {
			*targets[index] = []string{}
		}
	}
	return nil
}
