package safety

import (
	"context"
	"database/sql"
	"encoding/json"
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

	now := s.now()
	resolved, err := s.resolveReportContext(ctx, reporterID, input, now)
	if err != nil {
		return ReportResult{}, err
	}
	var reportID int64
	err = s.store.WithTx(ctx, func(tx *sql.Tx) error {
		var err error
		reportID, err = s.store.CreateSafetyReport(ctx, tx, store.CreateSafetyReportParams{
			ReporterID:      reporterID,
			ReportedUserID:  resolved.reportedUserID,
			IntentID:        resolved.intentID,
			RoomID:          resolved.roomID,
			ConversationID:  resolved.conversationID,
			GroupID:         resolved.groupID,
			Category:        input.Category,
			Details:         input.Details,
			ContextSnapshot: resolved.snapshot,
			RiskSignals:     resolved.riskSignals,
			Now:             now,
		})
		if err != nil {
			return err
		}
		if err := s.store.CreateSafetyReviewEvent(ctx, tx, store.SafetyReviewEvent{
			EntityType: "report", EntityID: reportID, ActorUserID: reporterID,
			Action: "submitted", NextStatus: "open", CreatedAt: now,
		}); err != nil {
			return err
		}
		if !input.BlockUser {
			return nil
		}
		if err := s.store.UpsertUserBlock(
			ctx,
			tx,
			reporterID,
			resolved.reportedUserID,
			"report:"+input.Category,
			now,
		); err != nil {
			return err
		}
		return s.store.CloseMatchRoomsBetweenUsers(ctx, tx, reporterID, resolved.reportedUserID, now)
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
		view, err := s.reportView(ctx, record)
		if err != nil {
			return nil, err
		}
		views = append(views, view)
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
	now := s.now()
	err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		previous, err := s.store.GetSafetyReport(ctx, tx, reportID)
		if err != nil {
			return mapStoreError(err, "找不到这条举报")
		}
		if previous.ReportedUserID == adminID {
			return forbidden("管理员不能处理涉及自己的举报")
		}
		if err := s.store.UpdateSafetyReport(
			ctx,
			tx,
			store.UpdateSafetyReportParams{
				ReportID:    reportID,
				Status:      input.Status,
				ReviewNotes: input.ReviewNotes,
				AssignToMe:  input.AssignToMe,
				AdminID:     adminID,
				Now:         now,
			},
		); err != nil {
			return mapStoreError(err, "找不到这条举报")
		}
		return s.store.CreateSafetyReviewEvent(ctx, tx, store.SafetyReviewEvent{
			EntityType: "report", EntityID: reportID, ActorUserID: adminID,
			Action: "reviewed", PreviousStatus: previous.Status, NextStatus: input.Status,
			Note: input.ReviewNotes, CreatedAt: now,
		})
	})
	if err != nil {
		return ReportView{}, err
	}
	record, err := s.store.GetSafetyReport(ctx, nil, reportID)
	if err != nil {
		return ReportView{}, mapStoreError(err, "找不到这条举报")
	}
	return s.reportView(ctx, record)
}

func (s *Service) ListMyCases(ctx context.Context, userID int64) ([]SafetyCaseView, error) {
	if s == nil || s.store == nil {
		return nil, httpapi.FeatureNotReady("safety")
	}
	records, err := s.store.ListSafetyReportsForReportedUser(ctx, nil, userID, 50)
	if err != nil {
		return nil, httpapi.InternalError(err)
	}
	views := make([]SafetyCaseView, 0, len(records))
	for _, record := range records {
		appeal, err := s.latestAppealView(ctx, record.ID)
		if err != nil {
			return nil, err
		}
		contextType, _ := reportContext(record)
		views = append(views, SafetyCaseView{
			ID: record.ID, TicketID: ticketID(record.ID), Category: record.Category,
			ContextType: contextType, ContextTitle: record.ContextTitle, Status: record.Status,
			CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt, Appeal: appeal,
		})
	}
	return views, nil
}

func (s *Service) AppealReport(
	ctx context.Context,
	userID, reportID int64,
	input ReportAppealInput,
) (ReportAppealView, error) {
	input.Reason = strings.TrimSpace(input.Reason)
	reasonLength := utf8.RuneCountInString(input.Reason)
	if reportID <= 0 || reasonLength < 8 || reasonLength > 1000 {
		return ReportAppealView{}, invalidParams("请用 8 到 1000 个字说明申诉理由")
	}
	if s == nil || s.store == nil {
		return ReportAppealView{}, httpapi.FeatureNotReady("safety")
	}
	now := s.now()
	var appeal store.SafetyReportAppeal
	err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		report, err := s.store.GetSafetyReport(ctx, tx, reportID)
		if err != nil || report.ReportedUserID != userID {
			return httpapi.NewError(http.StatusNotFound, httpapi.CodeUserNotFound, "找不到可申诉的安全工单")
		}
		if report.Status != "resolved" {
			return httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, "只有已处理的安全工单可以申诉")
		}
		if _, err := s.store.GetLatestSafetyReportAppeal(ctx, tx, reportID); err == nil {
			return httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, "这条安全工单已经申诉过")
		} else if !errors.Is(err, store.ErrNotFound) {
			return httpapi.InternalError(err)
		}
		appeal, err = s.store.CreateSafetyReportAppeal(ctx, tx, reportID, userID, input.Reason, now)
		if err != nil {
			return httpapi.InternalError(err)
		}
		if err := s.store.CreateSafetyReviewEvent(ctx, tx, store.SafetyReviewEvent{
			EntityType: "report_appeal", EntityID: appeal.ID, ActorUserID: userID,
			Action: "submitted", NextStatus: "pending", Note: input.Reason, CreatedAt: now,
		}); err != nil {
			return err
		}
		return s.store.CreateSafetyReviewEvent(ctx, tx, store.SafetyReviewEvent{
			EntityType: "report", EntityID: reportID, ActorUserID: userID,
			Action: "appealed", PreviousStatus: report.Status, NextStatus: report.Status,
			CreatedAt: now,
		})
	})
	if err != nil {
		return ReportAppealView{}, err
	}
	return toReportAppealView(appeal), nil
}

func (s *Service) ReviewReportAppeal(
	ctx context.Context,
	adminID int64,
	adminEmail string,
	appealID int64,
	input ReportAppealReviewInput,
) (ReportAppealView, error) {
	if !s.isAdmin(adminEmail) {
		return ReportAppealView{}, forbidden("仅安全管理员可以处理申诉")
	}
	input.Status = strings.TrimSpace(strings.ToLower(input.Status))
	input.ReviewNotes = strings.TrimSpace(input.ReviewNotes)
	if appealID <= 0 || (input.Status != "accepted" && input.Status != "rejected") {
		return ReportAppealView{}, invalidParams("申诉审核参数不正确")
	}
	if utf8.RuneCountInString(input.ReviewNotes) > 1000 {
		return ReportAppealView{}, invalidParams("申诉审核说明不能超过 1000 个字")
	}
	if s == nil || s.store == nil {
		return ReportAppealView{}, httpapi.FeatureNotReady("safety")
	}
	now := s.now()
	err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		appeal, err := s.store.GetSafetyReportAppeal(ctx, tx, appealID)
		if err != nil {
			return mapStoreError(err, "找不到这条申诉")
		}
		if appeal.Status != "pending" {
			return httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, "这条申诉已经处理")
		}
		if appeal.AppellantID == adminID {
			return forbidden("管理员不能处理自己的申诉")
		}
		report, err := s.store.GetSafetyReport(ctx, tx, appeal.ReportID)
		if err != nil {
			return mapStoreError(err, "找不到关联的安全工单")
		}
		if report.Status != "resolved" {
			return httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, "关联工单已不在可申诉状态")
		}
		if err := s.store.ReviewSafetyReportAppeal(
			ctx, tx, appealID, adminID, input.Status, input.ReviewNotes, now,
		); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				return httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, "这条申诉已经处理")
			}
			return httpapi.InternalError(err)
		}
		if err := s.store.CreateSafetyReviewEvent(ctx, tx, store.SafetyReviewEvent{
			EntityType: "report_appeal", EntityID: appealID, ActorUserID: adminID,
			Action: "reviewed", PreviousStatus: "pending", NextStatus: input.Status,
			Note: input.ReviewNotes, CreatedAt: now,
		}); err != nil {
			return err
		}
		nextReportStatus := report.Status
		if input.Status == "accepted" {
			nextReportStatus = "dismissed"
			if err := s.store.UpdateSafetyReport(ctx, tx, store.UpdateSafetyReportParams{
				ReportID: report.ID, Status: nextReportStatus, ReviewNotes: report.ReviewNotes,
				AdminID: adminID, Now: now,
			}); err != nil {
				return httpapi.InternalError(err)
			}
		}
		return s.store.CreateSafetyReviewEvent(ctx, tx, store.SafetyReviewEvent{
			EntityType: "report", EntityID: report.ID, ActorUserID: adminID,
			Action: "appeal_" + input.Status, PreviousStatus: report.Status,
			NextStatus: nextReportStatus, Note: input.ReviewNotes, CreatedAt: now,
		})
	})
	if err != nil {
		return ReportAppealView{}, err
	}
	appeal, err := s.store.GetSafetyReportAppeal(ctx, nil, appealID)
	if err != nil {
		return ReportAppealView{}, mapStoreError(err, "找不到这条申诉")
	}
	return toReportAppealView(appeal), nil
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

func (s *Service) reportView(ctx context.Context, record store.SafetyReport) (ReportView, error) {
	view := toReportView(record)
	appeal, err := s.latestAppealView(ctx, record.ID)
	if err != nil {
		return ReportView{}, err
	}
	view.Appeal = appeal
	return view, nil
}

func (s *Service) latestAppealView(ctx context.Context, reportID int64) (*ReportAppealView, error) {
	appeal, err := s.store.GetLatestSafetyReportAppeal(ctx, nil, reportID)
	if errors.Is(err, store.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, httpapi.InternalError(err)
	}
	view := toReportAppealView(appeal)
	return &view, nil
}

func toReportAppealView(appeal store.SafetyReportAppeal) ReportAppealView {
	return ReportAppealView{
		ID: appeal.ID, Reason: appeal.Reason, Status: appeal.Status,
		ReviewNotes: appeal.ReviewNotes, CreatedAt: appeal.CreatedAt, UpdatedAt: appeal.UpdatedAt,
	}
}

func toReportView(record store.SafetyReport) ReportView {
	contextType, contextID := reportContext(record)
	var snapshot any = map[string]any{}
	if len(record.ContextSnapshot) > 0 {
		_ = json.Unmarshal(record.ContextSnapshot, &snapshot)
	}
	riskSignals := make([]string, 0)
	if len(record.RiskSignals) > 0 {
		_ = json.Unmarshal(record.RiskSignals, &riskSignals)
	}
	view := ReportView{
		ID:                   record.ID,
		TicketID:             ticketID(record.ID),
		ReporterID:           record.ReporterID,
		ReporterNickname:     record.ReporterNickname,
		ReporterEmail:        record.ReporterEmail,
		ReportedUserID:       record.ReportedUserID,
		ReportedUserNickname: record.ReportedUserNickname,
		ReportedUserEmail:    record.ReportedUserEmail,
		ContextType:          contextType,
		ContextID:            contextID,
		ContextTitle:         record.ContextTitle,
		Category:             record.Category,
		Details:              record.Details,
		Status:               record.Status,
		AssignedToNickname:   record.AssignedToNickname,
		ReviewNotes:          record.ReviewNotes,
		ContextSnapshot:      snapshot,
		RiskSignals:          riskSignals,
		CreatedAt:            record.CreatedAt,
		UpdatedAt:            record.UpdatedAt,
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

func reportContext(record store.SafetyReport) (string, int64) {
	switch {
	case record.IntentID.Valid:
		return ContextIntent, record.IntentID.Int64
	case record.RoomID.Valid:
		return ContextMatchRoom, record.RoomID.Int64
	case record.ConversationID.Valid:
		return ContextDirect, record.ConversationID.Int64
	case record.GroupID.Valid:
		return ContextGroup, record.GroupID.Int64
	default:
		return "", 0
	}
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
