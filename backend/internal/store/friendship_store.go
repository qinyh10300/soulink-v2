package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Friendship struct {
	ID          int64
	RequesterID int64
	AddresseeID int64
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type FriendshipWithUser struct {
	Friendship
	PeerNickname   string
	PeerEmail      string
	PeerDepartment string
	PeerGrade      string
}

func (s *Store) CreateFriendship(ctx context.Context, tx *sql.Tx, requesterID, addresseeID int64, now time.Time) (Friendship, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		INSERT INTO friendships (requester_id, addressee_id, status, created_at, updated_at)
		VALUES ($1, $2, 'pending', $3, $3)
		RETURNING id, requester_id, addressee_id, status, created_at, updated_at
	`, requesterID, addresseeID, now)
	return scanFriendship(row)
}

func (s *Store) EnsureAcceptedFriendshipBetween(ctx context.Context, tx *sql.Tx, userA, userB int64, now time.Time) (Friendship, error) {
	existing, err := s.GetFriendshipBetween(ctx, tx, userA, userB)
	if err == nil {
		if existing.Status == "accepted" {
			return existing, nil
		}

		row := executor(s.db, tx).QueryRowContext(ctx, `
			UPDATE friendships
			SET status = 'accepted', updated_at = $2
			WHERE id = $1
			RETURNING id, requester_id, addressee_id, status, created_at, updated_at
		`, existing.ID, now)
		updated, scanErr := scanFriendship(row)
		if scanErr != nil {
			return Friendship{}, fmt.Errorf("accept friendship: %w", scanErr)
		}
		return updated, nil
	}
	if err != ErrNotFound {
		return Friendship{}, err
	}

	row := executor(s.db, tx).QueryRowContext(ctx, `
		INSERT INTO friendships (requester_id, addressee_id, status, created_at, updated_at)
		VALUES ($1, $2, 'accepted', $3, $3)
		RETURNING id, requester_id, addressee_id, status, created_at, updated_at
	`, userA, userB, now)
	friendship, scanErr := scanFriendship(row)
	if scanErr != nil {
		return Friendship{}, fmt.Errorf("create accepted friendship: %w", scanErr)
	}
	return friendship, nil
}

func (s *Store) GetFriendship(ctx context.Context, tx *sql.Tx, id int64) (Friendship, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT id, requester_id, addressee_id, status, created_at, updated_at
		FROM friendships
		WHERE id = $1
	`, id)
	return scanFriendship(row)
}

func (s *Store) GetFriendshipBetween(ctx context.Context, tx *sql.Tx, userA, userB int64) (Friendship, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT id, requester_id, addressee_id, status, created_at, updated_at
		FROM friendships
		WHERE (requester_id = $1 AND addressee_id = $2)
		   OR (requester_id = $2 AND addressee_id = $1)
		LIMIT 1
	`, userA, userB)
	return scanFriendship(row)
}

func (s *Store) UpdateFriendshipStatus(ctx context.Context, tx *sql.Tx, id int64, status string, now time.Time) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE friendships SET status = $2, updated_at = $3 WHERE id = $1
	`, id, status, now)
	if err != nil {
		return fmt.Errorf("update friendship status: %w", err)
	}
	return requireRowsAffected(result)
}

func (s *Store) DeleteFriendship(ctx context.Context, tx *sql.Tx, id int64) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		DELETE FROM friendships WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("delete friendship: %w", err)
	}
	return requireRowsAffected(result)
}

// ListAcceptedFriends returns all accepted friends with peer user info.
func (s *Store) ListAcceptedFriends(ctx context.Context, tx *sql.Tx, userID int64) ([]FriendshipWithUser, error) {
	rows, err := executor(s.db, tx).(interface {
		QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	}).QueryContext(ctx, `
		SELECT f.id, f.requester_id, f.addressee_id, f.status, f.created_at, f.updated_at,
		       u.nickname, u.email, u.department, u.grade
		FROM friendships f
		JOIN users u ON u.id = CASE WHEN f.requester_id = $1 THEN f.addressee_id ELSE f.requester_id END
		WHERE (f.requester_id = $1 OR f.addressee_id = $1)
		  AND f.status = 'accepted'
		  AND u.deleted_at IS NULL
		ORDER BY f.updated_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list accepted friends: %w", err)
	}
	defer rows.Close()
	return scanFriendshipWithUserRows(rows)
}

// ListReceivedRequests returns pending friend requests received by user.
func (s *Store) ListReceivedRequests(ctx context.Context, tx *sql.Tx, userID int64) ([]FriendshipWithUser, error) {
	rows, err := executor(s.db, tx).(interface {
		QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	}).QueryContext(ctx, `
		SELECT f.id, f.requester_id, f.addressee_id, f.status, f.created_at, f.updated_at,
		       u.nickname, u.email, u.department, u.grade
		FROM friendships f
		JOIN users u ON u.id = f.requester_id
		WHERE f.addressee_id = $1 AND f.status = 'pending' AND u.deleted_at IS NULL
		ORDER BY f.created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list received requests: %w", err)
	}
	defer rows.Close()
	return scanFriendshipWithUserRows(rows)
}

// ListSentRequests returns pending friend requests sent by user.
func (s *Store) ListSentRequests(ctx context.Context, tx *sql.Tx, userID int64) ([]FriendshipWithUser, error) {
	rows, err := executor(s.db, tx).(interface {
		QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	}).QueryContext(ctx, `
		SELECT f.id, f.requester_id, f.addressee_id, f.status, f.created_at, f.updated_at,
		       u.nickname, u.email, u.department, u.grade
		FROM friendships f
		JOIN users u ON u.id = f.addressee_id
		WHERE f.requester_id = $1 AND f.status = 'pending' AND u.deleted_at IS NULL
		ORDER BY f.created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list sent requests: %w", err)
	}
	defer rows.Close()
	return scanFriendshipWithUserRows(rows)
}

func scanFriendship(row *sql.Row) (Friendship, error) {
	var f Friendship
	err := row.Scan(&f.ID, &f.RequesterID, &f.AddresseeID, &f.Status, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return Friendship{}, ErrNotFound
		}
		return Friendship{}, fmt.Errorf("scan friendship: %w", err)
	}
	return f, nil
}

func scanFriendshipWithUserRows(rows *sql.Rows) ([]FriendshipWithUser, error) {
	results := make([]FriendshipWithUser, 0)
	for rows.Next() {
		var f FriendshipWithUser
		if err := rows.Scan(
			&f.ID, &f.RequesterID, &f.AddresseeID, &f.Status, &f.CreatedAt, &f.UpdatedAt,
			&f.PeerNickname, &f.PeerEmail, &f.PeerDepartment, &f.PeerGrade,
		); err != nil {
			return nil, fmt.Errorf("scan friendship with user: %w", err)
		}
		results = append(results, f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate friendships: %w", err)
	}
	return results, nil
}
