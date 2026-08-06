package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type PlazaGroupInvite struct {
	ID                  int64
	PostID              int64
	OwnerID             int64
	InviteType          string
	Title               string
	Summary             string
	ExtractedTime       string
	ExtractedLocation   string
	MinMembers          int
	MaxMembers          int
	RevealStrategy      string
	Deadline            time.Time
	GroupExpiresAt      time.Time
	Status              string
	ChatGroupID         sql.NullInt64
	InferenceSource     string
	InferenceModel      string
	InferenceConfidence float64
	PromptVersion       string
	FormedAt            sql.NullTime
	ArchivedAt          sql.NullTime
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type PlazaGroupApplication struct {
	ID                  int64
	InviteID            int64
	UserID              int64
	CommentID           int64
	AnonymousAlias      string
	Status              string
	ApplicantNickname   string
	ApplicantDepartment string
	ApplicantGrade      string
	RespondedAt         sql.NullTime
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type PlazaGroupApplicationCounts struct {
	Applied    int
	Accepted   int
	Declined   int
	Waitlisted int
	Left       int
}

type CreatePlazaGroupInviteParams struct {
	PostID              int64
	OwnerID             int64
	InviteType          string
	Title               string
	Summary             string
	ExtractedTime       string
	ExtractedLocation   string
	MinMembers          int
	MaxMembers          int
	RevealStrategy      string
	Deadline            time.Time
	GroupExpiresAt      time.Time
	InferenceSource     string
	InferenceModel      string
	InferenceConfidence float64
	PromptVersion       string
	Now                 time.Time
}

func (s *Store) CreatePlazaGroupInvite(ctx context.Context, tx *sql.Tx, params CreatePlazaGroupInviteParams) (PlazaGroupInvite, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		INSERT INTO plaza_group_invites (
			post_id, owner_id, invite_type, title, summary,
			extracted_time, extracted_location, min_members, max_members,
			reveal_strategy, deadline, group_expires_at, status,
			inference_source, inference_model, inference_confidence, prompt_version,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12, 'open',
			$13, $14, $15, $16,
			$17, $17
		)
		RETURNING `+plazaGroupInviteColumns+`
	`, params.PostID, params.OwnerID, params.InviteType, params.Title, params.Summary,
		params.ExtractedTime, params.ExtractedLocation, params.MinMembers, params.MaxMembers,
		params.RevealStrategy, params.Deadline, params.GroupExpiresAt,
		params.InferenceSource, params.InferenceModel, params.InferenceConfidence,
		params.PromptVersion, params.Now)
	invite, err := scanPlazaGroupInvite(row)
	if err != nil {
		return PlazaGroupInvite{}, fmt.Errorf("create plaza group invite: %w", err)
	}
	return invite, nil
}

func (s *Store) GetPlazaGroupInvite(ctx context.Context, tx *sql.Tx, inviteID int64) (PlazaGroupInvite, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT `+plazaGroupInviteColumns+`
		FROM plaza_group_invites
		WHERE id = $1
	`, inviteID)
	return scanPlazaGroupInvite(row)
}

func (s *Store) LockPlazaGroupInvite(ctx context.Context, tx *sql.Tx, inviteID int64) (PlazaGroupInvite, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT `+plazaGroupInviteColumns+`
		FROM plaza_group_invites
		WHERE id = $1
		FOR UPDATE
	`, inviteID)
	return scanPlazaGroupInvite(row)
}

func (s *Store) GetPlazaGroupInviteByPost(ctx context.Context, tx *sql.Tx, postID int64) (PlazaGroupInvite, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT `+plazaGroupInviteColumns+`
		FROM plaza_group_invites
		WHERE post_id = $1
	`, postID)
	return scanPlazaGroupInvite(row)
}

func (s *Store) ListPlazaGroupInvitesForUser(ctx context.Context, tx *sql.Tx, userID int64) ([]PlazaGroupInvite, error) {
	rows, err := executor(s.db, tx).(interface {
		QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	}).QueryContext(ctx, `
		SELECT `+plazaGroupInviteColumns+`
		FROM plaza_group_invites invite
		WHERE invite.owner_id = $1
		   OR EXISTS (
			SELECT 1
			FROM plaza_group_applications application
			WHERE application.invite_id = invite.id AND application.user_id = $1
		   )
		ORDER BY invite.updated_at DESC, invite.id DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list plaza group invites: %w", err)
	}
	defer rows.Close()

	invites := make([]PlazaGroupInvite, 0)
	for rows.Next() {
		invite, err := scanPlazaGroupInvite(rows)
		if err != nil {
			return nil, err
		}
		invites = append(invites, invite)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate plaza group invites: %w", err)
	}
	return invites, nil
}

func (s *Store) CreatePlazaGroupApplication(
	ctx context.Context,
	tx *sql.Tx,
	inviteID, userID, commentID int64,
	alias string,
	now time.Time,
) (PlazaGroupApplication, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		WITH inserted AS (
			INSERT INTO plaza_group_applications (
				invite_id, user_id, comment_id, anonymous_alias, status, created_at, updated_at
			) VALUES ($1, $2, $3, $4, 'applied', $5, $5)
			RETURNING id, invite_id, user_id, comment_id, anonymous_alias,
			          status, responded_at, created_at, updated_at
		)
		SELECT inserted.id, inserted.invite_id, inserted.user_id, inserted.comment_id,
		       inserted.anonymous_alias, inserted.status,
		       user_account.nickname, user_account.department, user_account.grade,
		       inserted.responded_at, inserted.created_at, inserted.updated_at
		FROM inserted
		JOIN users user_account ON user_account.id = inserted.user_id
	`, inviteID, userID, commentID, alias, now)
	application, err := scanPlazaGroupApplication(row)
	if err != nil {
		return PlazaGroupApplication{}, fmt.Errorf("create plaza group application: %w", err)
	}
	return application, nil
}

func (s *Store) GetPlazaGroupApplication(ctx context.Context, tx *sql.Tx, applicationID int64) (PlazaGroupApplication, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, plazaGroupApplicationSelect+`
		WHERE application.id = $1
	`, applicationID)
	return scanPlazaGroupApplication(row)
}

func (s *Store) LockPlazaGroupApplication(ctx context.Context, tx *sql.Tx, applicationID int64) (PlazaGroupApplication, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, plazaGroupApplicationSelect+`
		WHERE application.id = $1
		FOR UPDATE OF application
	`, applicationID)
	return scanPlazaGroupApplication(row)
}

func (s *Store) GetPlazaGroupApplicationByUser(ctx context.Context, tx *sql.Tx, inviteID, userID int64) (PlazaGroupApplication, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, plazaGroupApplicationSelect+`
		WHERE application.invite_id = $1 AND application.user_id = $2
	`, inviteID, userID)
	return scanPlazaGroupApplication(row)
}

func (s *Store) ListPlazaGroupApplications(ctx context.Context, tx *sql.Tx, inviteID int64) ([]PlazaGroupApplication, error) {
	rows, err := executor(s.db, tx).(interface {
		QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	}).QueryContext(ctx, plazaGroupApplicationSelect+`
		WHERE application.invite_id = $1
		ORDER BY application.created_at, application.id
	`, inviteID)
	if err != nil {
		return nil, fmt.Errorf("list plaza group applications: %w", err)
	}
	defer rows.Close()

	applications := make([]PlazaGroupApplication, 0)
	for rows.Next() {
		application, err := scanPlazaGroupApplication(rows)
		if err != nil {
			return nil, err
		}
		applications = append(applications, application)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate plaza group applications: %w", err)
	}
	return applications, nil
}

func (s *Store) CountPlazaGroupApplications(ctx context.Context, tx *sql.Tx, inviteID int64) (PlazaGroupApplicationCounts, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE status = 'applied'),
			COUNT(*) FILTER (WHERE status = 'accepted'),
			COUNT(*) FILTER (WHERE status = 'declined'),
			COUNT(*) FILTER (WHERE status = 'waitlisted'),
			COUNT(*) FILTER (WHERE status = 'left')
		FROM plaza_group_applications
		WHERE invite_id = $1
	`, inviteID)
	var counts PlazaGroupApplicationCounts
	if err := row.Scan(&counts.Applied, &counts.Accepted, &counts.Declined, &counts.Waitlisted, &counts.Left); err != nil {
		return PlazaGroupApplicationCounts{}, fmt.Errorf("count plaza group applications: %w", err)
	}
	return counts, nil
}

func (s *Store) UpdatePlazaGroupApplicationStatus(
	ctx context.Context,
	tx *sql.Tx,
	applicationID int64,
	status string,
	now time.Time,
) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE plaza_group_applications
		SET status = $2, responded_at = $3, updated_at = $3
		WHERE id = $1
	`, applicationID, status, now)
	if err != nil {
		return fmt.Errorf("update plaza group application status: %w", err)
	}
	return requireRowsAffected(result)
}

func (s *Store) WaitlistPendingPlazaGroupApplications(ctx context.Context, tx *sql.Tx, inviteID int64, now time.Time) error {
	if _, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE plaza_group_applications
		SET status = 'waitlisted', responded_at = $2, updated_at = $2
		WHERE invite_id = $1 AND status = 'applied'
	`, inviteID, now); err != nil {
		return fmt.Errorf("waitlist pending plaza group applications: %w", err)
	}
	return nil
}

func (s *Store) MarkPlazaGroupFormed(
	ctx context.Context,
	tx *sql.Tx,
	inviteID, chatGroupID int64,
	now time.Time,
) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE plaza_group_invites
		SET status = 'formed', chat_group_id = $2, formed_at = $3, updated_at = $3
		WHERE id = $1 AND status = 'open'
	`, inviteID, chatGroupID, now)
	if err != nil {
		return fmt.Errorf("mark plaza group formed: %w", err)
	}
	return requireRowsAffected(result)
}

func (s *Store) CreatePlazaGroupKeepVote(ctx context.Context, tx *sql.Tx, inviteID, userID int64, now time.Time) (bool, error) {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		INSERT INTO plaza_group_keep_votes (invite_id, user_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (invite_id, user_id) DO NOTHING
	`, inviteID, userID, now)
	if err != nil {
		return false, fmt.Errorf("create plaza group keep vote: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("plaza group keep vote rows affected: %w", err)
	}
	return rowsAffected > 0, nil
}

func (s *Store) CountPlazaGroupKeepVotes(ctx context.Context, tx *sql.Tx, inviteID int64) (int, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM plaza_group_keep_votes vote
		WHERE vote.invite_id = $1
		  AND (
			EXISTS (
				SELECT 1 FROM plaza_group_invites invite
				WHERE invite.id = vote.invite_id AND invite.owner_id = vote.user_id
			)
			OR EXISTS (
				SELECT 1 FROM plaza_group_applications application
				WHERE application.invite_id = vote.invite_id
				  AND application.user_id = vote.user_id
				  AND application.status = 'accepted'
			)
		  )
	`, inviteID)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("count plaza group keep votes: %w", err)
	}
	return count, nil
}

func (s *Store) HasPlazaGroupKeepVote(ctx context.Context, tx *sql.Tx, inviteID, userID int64) (bool, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM plaza_group_keep_votes WHERE invite_id = $1 AND user_id = $2
		)
	`, inviteID, userID)
	var exists bool
	if err := row.Scan(&exists); err != nil {
		return false, fmt.Errorf("check plaza group keep vote: %w", err)
	}
	return exists, nil
}

func (s *Store) MarkPlazaGroupKept(ctx context.Context, tx *sql.Tx, inviteID int64, now time.Time) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE plaza_group_invites
		SET status = 'kept', updated_at = $2
		WHERE id = $1 AND status = 'formed'
	`, inviteID, now)
	if err != nil {
		return fmt.Errorf("mark plaza group kept: %w", err)
	}
	return requireRowsAffected(result)
}

func (s *Store) ExpirePlazaGroupInvites(ctx context.Context, tx *sql.Tx, now time.Time) error {
	if _, err := executor(s.db, tx).ExecContext(ctx, `
		WITH archived AS (
			UPDATE plaza_group_invites
			SET status = 'archived', archived_at = $1, updated_at = $1
			WHERE status = 'formed' AND group_expires_at <= $1
			RETURNING chat_group_id
		)
		UPDATE chat_groups
		SET deleted_at = $1, updated_at = $1
		WHERE id IN (SELECT chat_group_id FROM archived WHERE chat_group_id IS NOT NULL)
		  AND deleted_at IS NULL
	`, now); err != nil {
		return fmt.Errorf("archive expired plaza groups: %w", err)
	}

	if _, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE plaza_group_invites
		SET status = 'expired', updated_at = $1
		WHERE status = 'open' AND deadline <= $1
	`, now); err != nil {
		return fmt.Errorf("expire plaza group invites: %w", err)
	}
	return nil
}

const plazaGroupInviteColumns = `
	id, post_id, owner_id, invite_type, title, summary,
	extracted_time, extracted_location, min_members, max_members,
	reveal_strategy, deadline, group_expires_at, status, chat_group_id,
	inference_source, inference_model, inference_confidence, prompt_version,
	formed_at, archived_at, created_at, updated_at`

const plazaGroupApplicationSelect = `
	SELECT application.id, application.invite_id, application.user_id, application.comment_id,
	       application.anonymous_alias, application.status,
	       user_account.nickname, user_account.department, user_account.grade,
	       application.responded_at, application.created_at, application.updated_at
	FROM plaza_group_applications application
	JOIN users user_account ON user_account.id = application.user_id AND user_account.deleted_at IS NULL
`

func scanPlazaGroupInvite(scanner rowScanner) (PlazaGroupInvite, error) {
	var invite PlazaGroupInvite
	if err := scanner.Scan(
		&invite.ID, &invite.PostID, &invite.OwnerID, &invite.InviteType,
		&invite.Title, &invite.Summary, &invite.ExtractedTime, &invite.ExtractedLocation,
		&invite.MinMembers, &invite.MaxMembers, &invite.RevealStrategy,
		&invite.Deadline, &invite.GroupExpiresAt, &invite.Status, &invite.ChatGroupID,
		&invite.InferenceSource, &invite.InferenceModel, &invite.InferenceConfidence,
		&invite.PromptVersion, &invite.FormedAt, &invite.ArchivedAt,
		&invite.CreatedAt, &invite.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return PlazaGroupInvite{}, ErrNotFound
		}
		return PlazaGroupInvite{}, fmt.Errorf("scan plaza group invite: %w", err)
	}
	return invite, nil
}

func scanPlazaGroupApplication(scanner rowScanner) (PlazaGroupApplication, error) {
	var application PlazaGroupApplication
	if err := scanner.Scan(
		&application.ID, &application.InviteID, &application.UserID,
		&application.CommentID, &application.AnonymousAlias, &application.Status,
		&application.ApplicantNickname, &application.ApplicantDepartment,
		&application.ApplicantGrade, &application.RespondedAt,
		&application.CreatedAt, &application.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return PlazaGroupApplication{}, ErrNotFound
		}
		return PlazaGroupApplication{}, fmt.Errorf("scan plaza group application: %w", err)
	}
	return application, nil
}
