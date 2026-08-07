package media

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/store"
)

const (
	KindPlaza  = "plaza"
	KindAvatar = "avatar"

	MaxUploadBytes = 8 << 20
	maxDimension   = 12000
)

type Service struct {
	store *store.Store
	now   func() time.Time
}

type UploadInput struct {
	Filename string
	Content  []byte
}

type AssetView struct {
	Key        string `json:"key"`
	URL        string `json:"url"`
	MIMEType   string `json:"mimeType"`
	ByteSize   int    `json:"byteSize"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	IsAnimated bool   `json:"isAnimated"`
}

func NewService(dataStore *store.Store) *Service {
	return &Service{store: dataStore, now: time.Now}
}

func (s *Service) UploadPlaza(ctx context.Context, ownerID int64, input UploadInput) (AssetView, error) {
	asset, err := s.createAsset(ctx, nil, ownerID, KindPlaza, input)
	if err != nil {
		return AssetView{}, err
	}
	_ = s.store.DeleteExpiredOrphanedPlazaMedia(ctx, nil, ownerID, s.now().Add(-24*time.Hour))
	return toAssetView(asset), nil
}

func (s *Service) UploadAvatar(ctx context.Context, ownerID int64, input UploadInput) (AssetView, error) {
	if err := s.ensureReady(); err != nil {
		return AssetView{}, err
	}

	var created store.MediaAsset
	err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		var previousID int64
		settings, err := s.store.GetUserProfileSettings(ctx, tx, ownerID)
		if err == nil && settings.AvatarMediaID.Valid {
			previousID = settings.AvatarMediaID.Int64
		} else if err != nil && !errors.Is(err, store.ErrNotFound) {
			return httpapi.InternalError(err)
		}

		asset, err := s.createAsset(ctx, tx, ownerID, KindAvatar, input)
		if err != nil {
			return err
		}
		if err := s.store.SetUserAvatarMedia(ctx, tx, ownerID, asset.ID, s.now()); err != nil {
			return httpapi.InternalError(err)
		}
		if previousID > 0 {
			if err := s.store.DeleteMediaAsset(ctx, tx, previousID); err != nil && !errors.Is(err, store.ErrNotFound) {
				return httpapi.InternalError(err)
			}
		}
		created = asset
		return nil
	})
	if err != nil {
		return AssetView{}, err
	}
	return toAssetView(created), nil
}

func (s *Service) DeleteAvatar(ctx context.Context, ownerID int64) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	return s.store.WithTx(ctx, func(tx *sql.Tx) error {
		settings, err := s.store.GetUserProfileSettings(ctx, tx, ownerID)
		if errors.Is(err, store.ErrNotFound) || (err == nil && !settings.AvatarMediaID.Valid) {
			return nil
		}
		if err != nil {
			return httpapi.InternalError(err)
		}
		if err := s.store.ClearUserAvatarMedia(ctx, tx, ownerID, s.now()); err != nil {
			return httpapi.InternalError(err)
		}
		if err := s.store.DeleteMediaAsset(ctx, tx, settings.AvatarMediaID.Int64); err != nil && !errors.Is(err, store.ErrNotFound) {
			return httpapi.InternalError(err)
		}
		return nil
	})
}

func (s *Service) GetReadable(ctx context.Context, viewerID int64, assetKey string) (store.MediaAsset, error) {
	if err := s.ensureReady(); err != nil {
		return store.MediaAsset{}, err
	}
	if _, err := uuid.Parse(assetKey); err != nil {
		return store.MediaAsset{}, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "媒体地址无效")
	}
	asset, err := s.store.GetMediaAssetByKey(ctx, nil, assetKey)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return store.MediaAsset{}, httpapi.NewError(http.StatusNotFound, httpapi.CodeUserNotFound, "媒体不存在")
		}
		return store.MediaAsset{}, httpapi.InternalError(err)
	}
	if asset.Kind == KindAvatar && asset.OwnerID != viewerID {
		settings, err := s.store.GetUserProfileSettings(ctx, nil, asset.OwnerID)
		if err != nil || !settings.ProfilePublic || settings.AvatarMediaID.Int64 != asset.ID {
			return store.MediaAsset{}, httpapi.NewError(http.StatusForbidden, httpapi.CodeForbidden, "无权查看该媒体")
		}
	}
	return asset, nil
}

func (s *Service) createAsset(ctx context.Context, tx *sql.Tx, ownerID int64, kind string, input UploadInput) (store.MediaAsset, error) {
	if err := s.ensureReady(); err != nil {
		return store.MediaAsset{}, err
	}
	if ownerID <= 0 || (kind != KindPlaza && kind != KindAvatar) {
		return store.MediaAsset{}, invalidMedia("上传参数无效")
	}
	mimeType, width, height, animated, err := inspectImage(input.Content)
	if err != nil {
		return store.MediaAsset{}, err
	}
	if kind == KindAvatar && animated {
		return store.MediaAsset{}, invalidMedia("头像暂不支持动图，请上传 JPEG、PNG 或 WebP")
	}
	filename := strings.TrimSpace(filepath.Base(input.Filename))
	if len(filename) > 255 {
		filename = filename[:255]
	}
	asset, err := s.store.CreateMediaAsset(ctx, tx, store.CreateMediaAssetParams{
		AssetKey:   uuid.NewString(),
		OwnerID:    ownerID,
		Kind:       kind,
		Filename:   filename,
		MIMEType:   mimeType,
		ByteSize:   len(input.Content),
		Width:      width,
		Height:     height,
		IsAnimated: animated,
		Content:    input.Content,
		CreatedAt:  s.now(),
	})
	if err != nil {
		return store.MediaAsset{}, httpapi.InternalError(err)
	}
	return asset, nil
}

func inspectImage(content []byte) (string, int, int, bool, error) {
	if len(content) == 0 || len(content) > MaxUploadBytes {
		return "", 0, 0, false, invalidMedia("单张图片需小于 8 MB")
	}
	mimeType := http.DetectContentType(content)
	var config image.Config
	var err error
	switch mimeType {
	case "image/jpeg":
		config, err = jpeg.DecodeConfig(bytes.NewReader(content))
	case "image/png":
		config, err = png.DecodeConfig(bytes.NewReader(content))
	case "image/gif":
		config, err = gif.DecodeConfig(bytes.NewReader(content))
	case "image/webp":
		if len(content) < 12 || string(content[:4]) != "RIFF" || string(content[8:12]) != "WEBP" {
			err = fmt.Errorf("invalid webp signature")
		}
	default:
		return "", 0, 0, false, invalidMedia("仅支持 JPEG、PNG、WebP 或 GIF 图片")
	}
	if err != nil || config.Width > maxDimension || config.Height > maxDimension {
		return "", 0, 0, false, invalidMedia("图片无法识别或尺寸过大")
	}
	return mimeType, config.Width, config.Height, mimeType == "image/gif", nil
}

func toAssetView(asset store.MediaAsset) AssetView {
	return AssetView{
		Key:        asset.AssetKey,
		URL:        "/media/" + asset.AssetKey,
		MIMEType:   asset.MIMEType,
		ByteSize:   asset.ByteSize,
		Width:      asset.Width,
		Height:     asset.Height,
		IsAnimated: asset.IsAnimated,
	}
}

func (s *Service) ensureReady() error {
	if s == nil || s.store == nil {
		return httpapi.FeatureNotReady("media")
	}
	return nil
}

func invalidMedia(message string) error {
	return httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, message)
}
