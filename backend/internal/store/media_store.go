package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type MediaAsset struct {
	ID         int64
	AssetKey   string
	OwnerID    int64
	Kind       string
	Filename   string
	MIMEType   string
	ByteSize   int
	Width      int
	Height     int
	IsAnimated bool
	Content    []byte
	CreatedAt  time.Time
}

type CreateMediaAssetParams struct {
	AssetKey   string
	OwnerID    int64
	Kind       string
	Filename   string
	MIMEType   string
	ByteSize   int
	Width      int
	Height     int
	IsAnimated bool
	Content    []byte
	CreatedAt  time.Time
}

func (s *Store) CreateMediaAsset(ctx context.Context, tx *sql.Tx, params CreateMediaAssetParams) (MediaAsset, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		INSERT INTO media_assets (
			asset_key, owner_id, kind, filename, mime_type, byte_size,
			width, height, is_animated, content, created_at
		) VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, asset_key::text, owner_id, kind, filename, mime_type, byte_size,
		          width, height, is_animated, content, created_at
	`, params.AssetKey, params.OwnerID, params.Kind, params.Filename, params.MIMEType,
		params.ByteSize, params.Width, params.Height, params.IsAnimated, params.Content, params.CreatedAt)

	asset, err := scanMediaAsset(row)
	if err != nil {
		return MediaAsset{}, fmt.Errorf("create media asset: %w", err)
	}
	return asset, nil
}

func (s *Store) GetMediaAssetByKey(ctx context.Context, tx *sql.Tx, assetKey string) (MediaAsset, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT id, asset_key::text, owner_id, kind, filename, mime_type, byte_size,
		       width, height, is_animated, content, created_at
		FROM media_assets
		WHERE asset_key = $1::uuid
		LIMIT 1
	`, assetKey)
	return scanMediaAsset(row)
}

func (s *Store) AttachPlaygroundMedia(ctx context.Context, tx *sql.Tx, postID, mediaID int64, sortOrder int, now time.Time) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		INSERT INTO playground_post_media (post_id, media_id, sort_order, created_at)
		SELECT $1, id, $3, $4
		FROM media_assets
		WHERE id = $2 AND kind = 'plaza'
		ON CONFLICT (media_id) DO NOTHING
	`, postID, mediaID, sortOrder, now)
	if err != nil {
		return fmt.Errorf("attach playground media: %w", err)
	}
	return requireRowsAffected(result)
}

func (s *Store) DeleteExpiredOrphanedPlazaMedia(ctx context.Context, tx *sql.Tx, ownerID int64, before time.Time) error {
	_, err := executor(s.db, tx).ExecContext(ctx, `
		DELETE FROM media_assets m
		WHERE m.owner_id = $1
		  AND m.kind = 'plaza'
		  AND m.created_at < $2
		  AND NOT EXISTS (
			SELECT 1 FROM playground_post_media pm WHERE pm.media_id = m.id
		  )
	`, ownerID, before)
	if err != nil {
		return fmt.Errorf("delete orphaned plaza media: %w", err)
	}
	return nil
}

func (s *Store) DeleteMediaAsset(ctx context.Context, tx *sql.Tx, mediaID int64) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `DELETE FROM media_assets WHERE id = $1`, mediaID)
	if err != nil {
		return fmt.Errorf("delete media asset: %w", err)
	}
	return requireRowsAffected(result)
}

func scanMediaAsset(scanner rowScanner) (MediaAsset, error) {
	var asset MediaAsset
	if err := scanner.Scan(
		&asset.ID,
		&asset.AssetKey,
		&asset.OwnerID,
		&asset.Kind,
		&asset.Filename,
		&asset.MIMEType,
		&asset.ByteSize,
		&asset.Width,
		&asset.Height,
		&asset.IsAnimated,
		&asset.Content,
		&asset.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return MediaAsset{}, ErrNotFound
		}
		return MediaAsset{}, fmt.Errorf("scan media asset: %w", err)
	}
	return asset, nil
}
