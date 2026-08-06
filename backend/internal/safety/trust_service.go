package safety

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/store"
)

type TrustProfileView struct {
	EmailVerified     bool   `json:"emailVerified"`
	ProfileComplete   bool   `json:"profileComplete"`
	PhotoStatus       string `json:"photoStatus"`
	RecentlyActive    bool   `json:"recentlyActive"`
	PhotoRequestID    *int64 `json:"photoRequestId,omitempty"`
	PhotoReviewNotes  string `json:"photoReviewNotes,omitempty"`
	PhotoAppealReason string `json:"photoAppealReason,omitempty"`
	PhotoRequestedAt  string `json:"photoRequestedAt,omitempty"`
	PhotoUpdatedAt    string `json:"photoUpdatedAt,omitempty"`
}

type PhotoVerificationView struct {
	ID           int64  `json:"id"`
	UserID       int64  `json:"userId"`
	UserNickname string `json:"userNickname"`
	UserEmail    string `json:"userEmail"`
	EvidenceMIME string `json:"evidenceMime"`
	Status       string `json:"status"`
	ReviewNotes  string `json:"reviewNotes"`
	AppealReason string `json:"appealReason,omitempty"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

type SafetyReviewEventView struct {
	ID             int64  `json:"id"`
	ActorNickname  string `json:"actorNickname"`
	Action         string `json:"action"`
	PreviousStatus string `json:"previousStatus"`
	NextStatus     string `json:"nextStatus"`
	Note           string `json:"note"`
	CreatedAt      string `json:"createdAt"`
}

func (s *Service) GetTrustProfile(ctx context.Context, userID int64, includePrivate bool) (TrustProfileView, error) {
	if s == nil || s.store == nil {
		return TrustProfileView{}, httpapi.FeatureNotReady("trust")
	}
	profile, err := s.store.GetUserTrustProfile(ctx, nil, userID, s.now())
	if err != nil {
		return TrustProfileView{}, mapStoreError(err, "找不到这个用户")
	}
	return trustProfileView(profile, includePrivate), nil
}

func (s *Service) SubmitPhotoVerification(
	ctx context.Context,
	userID int64,
	mimeType string,
	evidence []byte,
) (TrustProfileView, error) {
	if mimeType != "image/jpeg" && mimeType != "image/png" {
		return TrustProfileView{}, invalidParams("照片仅支持 JPEG 或 PNG")
	}
	if len(evidence) == 0 || len(evidence) > 2*1024*1024 {
		return TrustProfileView{}, invalidParams("照片大小需要在 2MB 以内")
	}
	now := s.now()
	err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		request, err := s.store.CreatePhotoVerificationRequest(ctx, tx, userID, mimeType, evidence, now)
		if err != nil {
			if strings.Contains(err.Error(), "ux_photo_verification_pending_user") {
				return httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, "已有照片正在等待审核")
			}
			return httpapi.InternalError(err)
		}
		return s.store.CreateSafetyReviewEvent(ctx, tx, store.SafetyReviewEvent{
			EntityType: "photo_verification", EntityID: request.ID,
			ActorUserID: userID, Action: "submitted", NextStatus: "pending", CreatedAt: now,
		})
	})
	if err != nil {
		return TrustProfileView{}, err
	}
	return s.GetTrustProfile(ctx, userID, true)
}

func (s *Service) AppealPhotoVerification(
	ctx context.Context,
	userID, requestID int64,
	reason string,
) (TrustProfileView, error) {
	reason = strings.TrimSpace(reason)
	if requestID <= 0 || utf8.RuneCountInString(reason) < 8 || utf8.RuneCountInString(reason) > 1000 {
		return TrustProfileView{}, invalidParams("请用 8 到 1000 个字说明申诉理由")
	}
	now := s.now()
	err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		request, err := s.store.GetPhotoVerificationRequest(ctx, tx, requestID)
		if err != nil || request.UserID != userID || request.Status != "rejected" {
			return httpapi.NewError(http.StatusNotFound, httpapi.CodeUserNotFound, "没有可申诉的照片审核")
		}
		if err := s.store.AppealPhotoVerificationRequest(ctx, tx, requestID, userID, reason, now); err != nil {
			return httpapi.InternalError(err)
		}
		return s.store.CreateSafetyReviewEvent(ctx, tx, store.SafetyReviewEvent{
			EntityType: "photo_verification", EntityID: requestID,
			ActorUserID: userID, Action: "appealed", PreviousStatus: "rejected",
			NextStatus: "pending", Note: reason, CreatedAt: now,
		})
	})
	if err != nil {
		return TrustProfileView{}, err
	}
	return s.GetTrustProfile(ctx, userID, true)
}

func (s *Service) ListPhotoVerifications(
	ctx context.Context,
	adminEmail, status string,
) ([]PhotoVerificationView, error) {
	if !s.isAdmin(adminEmail) {
		return nil, forbidden("仅安全管理员可以查看照片审核")
	}
	status = strings.TrimSpace(strings.ToLower(status))
	if status != "" && status != "pending" && status != "approved" && status != "rejected" {
		return nil, invalidParams("未知的审核状态")
	}
	records, err := s.store.ListPhotoVerificationRequests(ctx, nil, status, 100)
	if err != nil {
		return nil, httpapi.InternalError(err)
	}
	views := make([]PhotoVerificationView, 0, len(records))
	for _, record := range records {
		views = append(views, photoVerificationView(record))
	}
	return views, nil
}

func (s *Service) PhotoVerificationEvidence(
	ctx context.Context,
	adminEmail string,
	requestID int64,
) (string, []byte, error) {
	if !s.isAdmin(adminEmail) {
		return "", nil, forbidden("仅安全管理员可以查看审核照片")
	}
	if requestID <= 0 {
		return "", nil, invalidParams("无效的照片审核编号")
	}
	_, mimeType, evidence, err := s.store.GetPhotoVerificationEvidence(ctx, nil, requestID)
	if err != nil {
		return "", nil, mapStoreError(err, "找不到这条照片审核")
	}
	return mimeType, evidence, nil
}

func (s *Service) ReviewPhotoVerification(
	ctx context.Context,
	adminID int64,
	adminEmail string,
	requestID int64,
	status, notes string,
) (PhotoVerificationView, error) {
	if !s.isAdmin(adminEmail) {
		return PhotoVerificationView{}, forbidden("仅安全管理员可以审核照片")
	}
	status = strings.TrimSpace(strings.ToLower(status))
	notes = strings.TrimSpace(notes)
	if requestID <= 0 || (status != "approved" && status != "rejected") {
		return PhotoVerificationView{}, invalidParams("审核参数不正确")
	}
	if status == "rejected" && utf8.RuneCountInString(notes) < 4 {
		return PhotoVerificationView{}, invalidParams("拒绝时请填写明确的审核说明")
	}
	if utf8.RuneCountInString(notes) > 1000 {
		return PhotoVerificationView{}, invalidParams("审核说明不能超过 1000 个字")
	}
	now := s.now()
	err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		request, err := s.store.GetPhotoVerificationRequest(ctx, tx, requestID)
		if err != nil {
			return mapStoreError(err, "找不到这条照片审核")
		}
		if request.Status != "pending" {
			return httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, "这条照片审核已经处理")
		}
		if err := s.store.ReviewPhotoVerificationRequest(
			ctx, tx, requestID, adminID, status, notes, now,
		); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				return httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, "这条照片审核已经处理")
			}
			return httpapi.InternalError(err)
		}
		return s.store.CreateSafetyReviewEvent(ctx, tx, store.SafetyReviewEvent{
			EntityType: "photo_verification", EntityID: requestID,
			ActorUserID: adminID, Action: "reviewed", PreviousStatus: request.Status,
			NextStatus: status, Note: notes, CreatedAt: now,
		})
	})
	if err != nil {
		return PhotoVerificationView{}, err
	}
	record, err := s.store.GetPhotoVerificationRequest(ctx, nil, requestID)
	if err != nil {
		return PhotoVerificationView{}, httpapi.InternalError(err)
	}
	return photoVerificationView(record), nil
}

func (s *Service) ReviewEvents(
	ctx context.Context,
	adminEmail, entityType string,
	entityID int64,
) ([]SafetyReviewEventView, error) {
	if !s.isAdmin(adminEmail) {
		return nil, forbidden("仅安全管理员可以查看审核日志")
	}
	if entityID <= 0 || (entityType != "report" && entityType != "report_appeal" && entityType != "photo_verification") {
		return nil, invalidParams("审核日志参数不正确")
	}
	events, err := s.store.ListSafetyReviewEvents(ctx, nil, entityType, entityID)
	if err != nil {
		return nil, httpapi.InternalError(err)
	}
	views := make([]SafetyReviewEventView, 0, len(events))
	for _, event := range events {
		views = append(views, SafetyReviewEventView{
			ID: event.ID, ActorNickname: event.ActorNickname, Action: event.Action,
			PreviousStatus: event.PreviousStatus, NextStatus: event.NextStatus,
			Note: event.Note, CreatedAt: event.CreatedAt.Format(time.RFC3339),
		})
	}
	return views, nil
}

func trustProfileView(profile store.UserTrustProfile, includePrivate bool) TrustProfileView {
	view := TrustProfileView{
		EmailVerified: profile.EmailVerified, ProfileComplete: profile.ProfileComplete,
		PhotoStatus: profile.PhotoStatus, RecentlyActive: profile.RecentlyActive,
	}
	if !includePrivate {
		if view.PhotoStatus != "approved" {
			view.PhotoStatus = "not_verified"
		}
		return view
	}
	if profile.PhotoRequestID.Valid {
		requestID := profile.PhotoRequestID.Int64
		view.PhotoRequestID = &requestID
	}
	view.PhotoReviewNotes = profile.PhotoReviewNotes
	view.PhotoAppealReason = profile.PhotoAppealReason
	if profile.PhotoRequestAt.Valid {
		view.PhotoRequestedAt = profile.PhotoRequestAt.Time.Format(time.RFC3339)
	}
	if profile.PhotoRequestUpdate.Valid {
		view.PhotoUpdatedAt = profile.PhotoRequestUpdate.Time.Format(time.RFC3339)
	}
	return view
}

func photoVerificationView(record store.PhotoVerificationRequest) PhotoVerificationView {
	return PhotoVerificationView{
		ID: record.ID, UserID: record.UserID, UserNickname: record.UserNickname,
		UserEmail: record.UserEmail, EvidenceMIME: record.EvidenceMIME,
		Status: record.Status, ReviewNotes: record.ReviewNotes,
		AppealReason: record.AppealReason, CreatedAt: record.CreatedAt.Format(time.RFC3339),
		UpdatedAt: record.UpdatedAt.Format(time.RFC3339),
	}
}
