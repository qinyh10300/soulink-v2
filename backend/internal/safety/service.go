package safety

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/store"
)

var reportCategories = map[string]struct{}{
	"harassment":    {},
	"privacy":       {},
	"impersonation": {},
	"unsafe":        {},
	"spam":          {},
	"other":         {},
}

var reportStatuses = map[string]struct{}{
	"open":      {},
	"reviewing": {},
	"resolved":  {},
	"dismissed": {},
}

type Service struct {
	store       *store.Store
	adminEmails map[string]struct{}
	nowFunc     func() time.Time
}

func NewService(dataStore *store.Store, adminEmails []string) *Service {
	admins := make(map[string]struct{}, len(adminEmails))
	for _, email := range adminEmails {
		normalized := strings.ToLower(strings.TrimSpace(email))
		if normalized != "" {
			admins[normalized] = struct{}{}
		}
	}
	return &Service{
		store:       dataStore,
		adminEmails: admins,
		nowFunc:     time.Now,
	}
}

func (s *Service) Capabilities(email string) Capabilities {
	return Capabilities{IsAdmin: s.isAdmin(email)}
}

func (s *Service) SubmitReport(
	ctx context.Context,
	reporterID int64,
	input ReportInput,
) (ReportResult, error) {
	input.ContextType = strings.TrimSpace(strings.ToLower(input.ContextType))
	input.Category = strings.TrimSpace(strings.ToLower(input.Category))
	input.Details = strings.TrimSpace(input.Details)

	if input.ContextID <= 0 {
		return ReportResult{}, invalidParams("请选择需要举报的内容")
	}
	if _, ok := reportCategories[input.Category]; !ok {
		return ReportResult{}, invalidParams("请选择举报原因")
	}
	detailLength := utf8.RuneCountInString(input.Details)
	if detailLength < 8 || detailLength > 2000 {
		return ReportResult{}, invalidParams("请用 8 到 2000 个字说明发生了什么")
	}
	if s == nil || s.store == nil {
		return ReportResult{}, httpapi.FeatureNotReady("safety")
	}

	var reportedUserID int64
	var intentID sql.NullInt64
	var roomID sql.NullInt64

	switch input.ContextType {
	case ContextIntent:
		intent, err := s.store.GetConnectionIntent(ctx, nil, reporterID, input.ContextID)
		if err != nil {
			return ReportResult{}, mapStoreError(err, "找不到这个意图")
		}
		reportedUserID = intent.AuthorID
		intentID = sql.NullInt64{Int64: intent.ID, Valid: true}
	case ContextMatchRoom:
		room, err := s.store.GetMatchRoom(ctx, nil, reporterID, input.ContextID)
		if err != nil {
			return ReportResult{}, mapStoreError(err, "找不到这个匹配房间")
		}
		reportedUserID = room.OwnerID
		if reportedUserID == reporterID {
			reportedUserID = room.CandidateID
		}
		roomID = sql.NullInt64{Int64: room.ID, Valid: true}
	default:
		return ReportResult{}, invalidParams("未知的举报场景")
	}

	if reportedUserID == reporterID {
		return ReportResult{}, invalidParams("不能举报自己的内容")
	}

	now := s.now()
	var reportID int64
	err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		var err error
		reportID, err = s.store.CreateSafetyReport(ctx, tx, store.CreateSafetyReportParams{
			ReporterID:     reporterID,
			ReportedUserID: reportedUserID,
			IntentID:       intentID,
			RoomID:         roomID,
			Category:       input.Category,
			Details:        input.Details,
			Now:            now,
		})
		if err != nil {
			return err
		}
		if !input.BlockUser {
			return nil
		}
		if err := s.store.UpsertUserBlock(
			ctx,
			tx,
			reporterID,
			reportedUserID,
			"report:"+input.Category,
			now,
		); err != nil {
			return err
		}
		return s.store.CloseMatchRoomsBetweenUsers(ctx, tx, reporterID, reportedUserID, now)
	})
	if err != nil {
		return ReportResult{}, httpapi.InternalError(err)
	}

	return ReportResult{
		ID:        reportID,
		TicketID:  ticketID(reportID),
		Blocked:   input.BlockUser,
		CreatedAt: now,
	}, nil
}

func (s *Service) ListReports(
	ctx context.Context,
	adminEmail string,
	status string,
) ([]ReportView, error) {
	if !s.isAdmin(adminEmail) {
		return nil, forbidden("仅安全管理员可以查看举报")
	}
	status = strings.TrimSpace(strings.ToLower(status))
	if status != "" {
		if _, ok := reportStatuses[status]; !ok {
			return nil, invalidParams("未知的处理状态")
		}
	}
	if s.store == nil {
		return nil, httpapi.FeatureNotReady("safety")
	}
	records, err := s.store.ListSafetyReports(ctx, nil, status, 100)
	if err != nil {
		return nil, httpapi.InternalError(err)
	}
	views := make([]ReportView, 0, len(records))
	for _, record := range records {
		views = append(views, toReportView(record))
	}
	return views, nil
}

func (s *Service) UpdateReport(
	ctx context.Context,
	adminID int64,
	adminEmail string,
	reportID int64,
	input ReportUpdateInput,
) (ReportView, error) {
	if !s.isAdmin(adminEmail) {
		return ReportView{}, forbidden("仅安全管理员可以处理举报")
	}
	input.Status = strings.TrimSpace(strings.ToLower(input.Status))
	input.ReviewNotes = strings.TrimSpace(input.ReviewNotes)
	if _, ok := reportStatuses[input.Status]; !ok {
		return ReportView{}, invalidParams("未知的处理状态")
	}
	if reportID <= 0 {
		return ReportView{}, invalidParams("无效的举报编号")
	}
	if utf8.RuneCountInString(input.ReviewNotes) > 2000 {
		return ReportView{}, invalidParams("审核备注不能超过 2000 个字")
	}
	if s.store == nil {
		return ReportView{}, httpapi.FeatureNotReady("safety")
	}
	if err := s.store.UpdateSafetyReport(
		ctx,
		nil,
		store.UpdateSafetyReportParams{
			ReportID:    reportID,
			Status:      input.Status,
			ReviewNotes: input.ReviewNotes,
			AssignToMe:  input.AssignToMe,
			AdminID:     adminID,
			Now:         s.now(),
		},
	); err != nil {
		return ReportView{}, mapStoreError(err, "找不到这条举报")
	}
	record, err := s.store.GetSafetyReport(ctx, nil, reportID)
	if err != nil {
		return ReportView{}, mapStoreError(err, "找不到这条举报")
	}
	return toReportView(record), nil
}

func (s *Service) isAdmin(email string) bool {
	if s == nil {
		return false
	}
	_, ok := s.adminEmails[strings.ToLower(strings.TrimSpace(email))]
	return ok
}

func (s *Service) now() time.Time {
	if s.nowFunc == nil {
		return time.Now().UTC()
	}
	return s.nowFunc().UTC()
}

func toReportView(record store.SafetyReport) ReportView {
	view := ReportView{
		ID:                   record.ID,
		TicketID:             ticketID(record.ID),
		ReporterID:           record.ReporterID,
		ReporterNickname:     record.ReporterNickname,
		ReporterEmail:        record.ReporterEmail,
		ReportedUserID:       record.ReportedUserID,
		ReportedUserNickname: record.ReportedUserNickname,
		ReportedUserEmail:    record.ReportedUserEmail,
		ContextTitle:         record.ContextTitle,
		Category:             record.Category,
		Details:              record.Details,
		Status:               record.Status,
		AssignedToNickname:   record.AssignedToNickname,
		ReviewNotes:          record.ReviewNotes,
		CreatedAt:            record.CreatedAt,
		UpdatedAt:            record.UpdatedAt,
	}
	if record.IntentID.Valid {
		view.ContextType = ContextIntent
		view.ContextID = record.IntentID.Int64
	} else {
		view.ContextType = ContextMatchRoom
		view.ContextID = record.RoomID.Int64
	}
	if record.ResolvedAt.Valid {
		resolvedAt := record.ResolvedAt.Time
		view.ResolvedAt = &resolvedAt
	}
	if record.ResolvedBy.Valid {
		resolvedBy := record.ResolvedBy.Int64
		view.ResolvedBy = &resolvedBy
	}
	if record.AssignedTo.Valid {
		assignedTo := record.AssignedTo.Int64
		view.AssignedTo = &assignedTo
	}
	return view
}

func ticketID(reportID int64) string {
	return fmt.Sprintf("SR-%06d", reportID)
}

func invalidParams(message string) error {
	return httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, message)
}

func forbidden(message string) error {
	return httpapi.NewError(http.StatusForbidden, httpapi.CodeForbidden, message)
}

func mapStoreError(err error, message string) error {
	if errors.Is(err, store.ErrNotFound) {
		return httpapi.NewError(http.StatusNotFound, httpapi.CodeUserNotFound, message)
	}
	return httpapi.InternalError(err)
}
