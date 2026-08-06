package playground

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi/v5/middleware"

	"soulink-backend/internal/analytics"
	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/store"
)

const (
	maxPlazaGroupTitleLength   = 100
	maxPlazaGroupSummaryLength = 1000
	maxPlazaGroupMembers       = 12
	plazaGroupPromptVersion    = "plaza-group-draft-v1"
)

var (
	plazaEmailPattern    = regexp.MustCompile(`(?i)[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}`)
	plazaPhonePattern    = regexp.MustCompile(`(?:\+?86[- ]?)?1[3-9][0-9]{9}`)
	plazaContactPattern  = regexp.MustCompile(`(?i)(?:微信|vx|v信|qq|q号)\s*(?:号|[:：])?\s*[a-z0-9_-]{5,}`)
	plazaDangerPattern   = regexp.MustCompile(`(?:人肉(?:他|她|你|一下)?|开盒|弄死|杀了你|跟踪你|报复你|曝光隐私)`)
	plazaTimePattern     = regexp.MustCompile(`(?:今天|明天|后天|本周末|周末|周[一二三四五六日天]|星期[一二三四五六日天])(?:[^，。；;\n]{0,12}(?:上午|中午|下午|晚上|晚间|[0-9]{1,2}(?::[0-9]{2})?点?))?`)
	plazaLocationPattern = regexp.MustCompile(`(?:紫操|东操|西操|图书馆|FIT\s*楼|FIT|蒙民伟楼?|六教|三教|清芬园|听涛园|观畴园|桃李园|荷园|情人坡|大礼堂)`)
)

type CreatePlazaGroupInviteInput struct {
	PostID         int64
	InviteType     string
	Title          string
	Summary        string
	MinMembers     int
	MaxMembers     int
	RevealStrategy string
	Deadline       time.Time
	GroupExpiresAt time.Time
}

type PlazaGroupIdentityView struct {
	Alias      string `json:"alias"`
	Revealed   bool   `json:"revealed"`
	UserID     int64  `json:"userId,omitempty"`
	Nickname   string `json:"nickname,omitempty"`
	Department string `json:"department,omitempty"`
	Grade      string `json:"grade,omitempty"`
}

type PlazaGroupApplicationView struct {
	ID        int64                  `json:"id"`
	Status    string                 `json:"status"`
	Identity  PlazaGroupIdentityView `json:"identity"`
	CreatedAt string                 `json:"createdAt"`
}

type PlazaGroupInviteSummary struct {
	ID                      int64   `json:"id"`
	PostID                  int64   `json:"postId"`
	InviteType              string  `json:"inviteType"`
	Title                   string  `json:"title"`
	Summary                 string  `json:"summary"`
	ExtractedTime           string  `json:"extractedTime"`
	ExtractedLocation       string  `json:"extractedLocation"`
	MinMembers              int     `json:"minMembers"`
	MaxMembers              int     `json:"maxMembers"`
	MemberCount             int     `json:"memberCount"`
	ApplicationCount        int     `json:"applicationCount"`
	RevealStrategy          string  `json:"revealStrategy"`
	Deadline                string  `json:"deadline"`
	GroupExpiresAt          string  `json:"groupExpiresAt"`
	Status                  string  `json:"status"`
	IsOwner                 bool    `json:"isOwner"`
	CanApply                bool    `json:"canApply"`
	ViewerApplicationStatus string  `json:"viewerApplicationStatus,omitempty"`
	InferenceSource         string  `json:"inferenceSource"`
	InferenceConfidence     float64 `json:"inferenceConfidence"`
	PromptVersion           string  `json:"promptVersion"`
}

type PlazaGroupInviteView struct {
	PlazaGroupInviteSummary
	Owner             PlazaGroupIdentityView      `json:"owner"`
	Applications      []PlazaGroupApplicationView `json:"applications"`
	ChatGroupID       *int64                      `json:"chatGroupId,omitempty"`
	CanForm           bool                        `json:"canForm"`
	CanKeep           bool                        `json:"canKeep"`
	KeepVotes         int                         `json:"keepVotes"`
	RequiredKeepVotes int                         `json:"requiredKeepVotes"`
}

func (s *Service) CreatePlazaGroupInvite(
	ctx context.Context,
	ownerID int64,
	input CreatePlazaGroupInviteInput,
) (PlazaGroupInviteView, error) {
	if err := s.ensureReady(); err != nil {
		return PlazaGroupInviteView{}, err
	}
	if input.PostID <= 0 {
		return PlazaGroupInviteView{}, invalidPlazaParams("请选择要转换的树洞")
	}
	if err := s.expirePlazaGroups(ctx); err != nil {
		return PlazaGroupInviteView{}, err
	}

	now := s.now()
	var created store.PlazaGroupInvite
	err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		post, err := s.store.GetPlaygroundPost(ctx, tx, ownerID, input.PostID)
		if err != nil {
			return mapPlazaGroupStoreError(err)
		}
		if post.Author.UserID != ownerID {
			return plazaForbidden("只有洞主可以把讨论转换为活动或群组")
		}
		if _, err := s.store.GetPlazaGroupInviteByPost(ctx, tx, post.ID); err == nil {
			return plazaConflict("这个树洞已经创建过连接邀请")
		} else if !errors.Is(err, store.ErrNotFound) {
			return httpapi.InternalError(err)
		}

		comments, err := s.store.ListPlaygroundComments(ctx, tx, ownerID, post.ID)
		if err != nil {
			return httpapi.InternalError(err)
		}
		prepared, err := preparePlazaGroupInvite(post, comments, input, now)
		if err != nil {
			return err
		}
		invite, err := s.store.CreatePlazaGroupInvite(ctx, tx, prepared)
		if err != nil {
			return httpapi.InternalError(err)
		}
		created = invite
		return nil
	})
	if err != nil {
		return PlazaGroupInviteView{}, err
	}
	return s.getPlazaGroupInviteView(ctx, ownerID, created.ID)
}

func (s *Service) ListMyPlazaGroupInvites(ctx context.Context, userID int64) ([]PlazaGroupInviteView, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := s.expirePlazaGroups(ctx); err != nil {
		return nil, err
	}
	invites, err := s.store.ListPlazaGroupInvitesForUser(ctx, nil, userID)
	if err != nil {
		return nil, httpapi.InternalError(err)
	}
	views := make([]PlazaGroupInviteView, 0, len(invites))
	for _, invite := range invites {
		view, err := s.buildPlazaGroupInviteView(ctx, nil, userID, invite)
		if err != nil {
			return nil, err
		}
		views = append(views, view)
	}
	return views, nil
}

func (s *Service) GetPlazaGroupInvite(ctx context.Context, userID, inviteID int64) (PlazaGroupInviteView, error) {
	if err := s.ensureReady(); err != nil {
		return PlazaGroupInviteView{}, err
	}
	if err := s.expirePlazaGroups(ctx); err != nil {
		return PlazaGroupInviteView{}, err
	}
	return s.getPlazaGroupInviteView(ctx, userID, inviteID)
}

func (s *Service) ApplyToPlazaGroup(ctx context.Context, userID, inviteID int64) (PlazaGroupInviteView, error) {
	if err := s.ensureReady(); err != nil {
		return PlazaGroupInviteView{}, err
	}
	if err := s.expirePlazaGroups(ctx); err != nil {
		return PlazaGroupInviteView{}, err
	}

	now := s.now()
	err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		invite, err := s.store.LockPlazaGroupInvite(ctx, tx, inviteID)
		if err != nil {
			return mapPlazaGroupStoreError(err)
		}
		if _, err := s.store.GetPlazaGroupApplicationByUser(ctx, tx, invite.ID, userID); err == nil {
			return plazaConflict("你已经申请过这个邀请")
		} else if !errors.Is(err, store.ErrNotFound) {
			return httpapi.InternalError(err)
		}

		post, err := s.store.GetPlaygroundPost(ctx, tx, userID, invite.PostID)
		if err != nil {
			return mapPlazaGroupStoreError(err)
		}
		comments, err := s.store.ListPlaygroundComments(ctx, tx, userID, invite.PostID)
		if err != nil {
			return httpapi.InternalError(err)
		}
		comment, ok := latestCommentByUser(comments, userID)
		if err := validatePlazaGroupApplicationEligibility(invite, userID, ok, now); err != nil {
			return err
		}

		blocked, err := s.store.UsersBlocked(ctx, tx, invite.OwnerID, userID)
		if err != nil {
			return httpapi.InternalError(err)
		}
		reported, err := s.store.UsersHaveActiveSafetyReport(ctx, tx, invite.OwnerID, userID)
		if err != nil {
			return httpapi.InternalError(err)
		}
		if blocked || reported {
			return plazaForbidden("当前安全设置不允许加入这个邀请")
		}
		alias := newAnonymousContext(post.Author.UserID, comments).labelsByUser[userID]
		if alias == "" || alias == "洞主" {
			return plazaForbidden("只有回复者可以申请加入")
		}
		if _, err := s.store.CreatePlazaGroupApplication(
			ctx, tx, invite.ID, userID, comment.ID, alias, now,
		); err != nil {
			return httpapi.InternalError(err)
		}
		return nil
	})
	if err != nil {
		return PlazaGroupInviteView{}, err
	}
	return s.getPlazaGroupInviteView(ctx, userID, inviteID)
}

func (s *Service) RespondToPlazaGroupApplication(
	ctx context.Context,
	ownerID, inviteID, applicationID int64,
	status string,
) (PlazaGroupInviteView, error) {
	status = strings.TrimSpace(status)
	if status != "accepted" && status != "declined" {
		return PlazaGroupInviteView{}, invalidPlazaParams("请选择接受或拒绝")
	}
	if err := s.ensureReady(); err != nil {
		return PlazaGroupInviteView{}, err
	}
	if err := s.expirePlazaGroups(ctx); err != nil {
		return PlazaGroupInviteView{}, err
	}

	now := s.now()
	formed := false
	err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		invite, err := s.store.LockPlazaGroupInvite(ctx, tx, inviteID)
		if err != nil {
			return mapPlazaGroupStoreError(err)
		}
		if invite.OwnerID != ownerID {
			return plazaForbidden("只有洞主可以处理申请")
		}
		if invite.Status != "open" || !now.Before(invite.Deadline) {
			return plazaConflict("这个邀请已经停止处理申请")
		}
		application, err := s.store.LockPlazaGroupApplication(ctx, tx, applicationID)
		if err != nil {
			return mapPlazaGroupStoreError(err)
		}
		if application.InviteID != invite.ID {
			return plazaForbidden("这个申请不属于当前邀请")
		}
		if application.Status != "applied" && application.Status != "waitlisted" {
			return plazaConflict("这个申请已经处理过")
		}

		counts, err := s.store.CountPlazaGroupApplications(ctx, tx, invite.ID)
		if err != nil {
			return httpapi.InternalError(err)
		}
		targetStatus := status
		if status == "accepted" && counts.Accepted+1 >= invite.MaxMembers {
			targetStatus = "waitlisted"
		}
		if err := s.store.UpdatePlazaGroupApplicationStatus(
			ctx, tx, application.ID, targetStatus, now,
		); err != nil {
			return httpapi.InternalError(err)
		}
		counts, err = s.store.CountPlazaGroupApplications(ctx, tx, invite.ID)
		if err != nil {
			return httpapi.InternalError(err)
		}
		if shouldFormPlazaGroup(invite, counts) {
			if _, err := s.formPlazaGroupTx(ctx, tx, invite, counts, now); err != nil {
				return err
			}
			formed = true
		}
		return nil
	})
	if err != nil {
		return PlazaGroupInviteView{}, err
	}
	view, err := s.getPlazaGroupInviteView(ctx, ownerID, inviteID)
	if err == nil && formed {
		s.recordPlazaGroupCreated(ctx, ownerID, view)
	}
	return view, err
}

func (s *Service) FormPlazaGroup(ctx context.Context, ownerID, inviteID int64) (PlazaGroupInviteView, error) {
	if err := s.ensureReady(); err != nil {
		return PlazaGroupInviteView{}, err
	}
	if err := s.expirePlazaGroups(ctx); err != nil {
		return PlazaGroupInviteView{}, err
	}

	now := s.now()
	created := false
	err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		invite, err := s.store.LockPlazaGroupInvite(ctx, tx, inviteID)
		if err != nil {
			return mapPlazaGroupStoreError(err)
		}
		if invite.OwnerID != ownerID {
			return plazaForbidden("只有洞主可以确认成组")
		}
		if invite.Status == "formed" || invite.Status == "kept" {
			return nil
		}
		if invite.Status != "open" || !now.Before(invite.Deadline) {
			return plazaConflict("这个邀请当前不能成组")
		}
		counts, err := s.store.CountPlazaGroupApplications(ctx, tx, invite.ID)
		if err != nil {
			return httpapi.InternalError(err)
		}
		if counts.Accepted+1 < invite.MinMembers {
			return plazaConflict(fmt.Sprintf("至少还需要 %d 位成员同意", invite.MinMembers-counts.Accepted-1))
		}
		if _, err := s.formPlazaGroupTx(ctx, tx, invite, counts, now); err != nil {
			return err
		}
		created = true
		return nil
	})
	if err != nil {
		return PlazaGroupInviteView{}, err
	}
	view, err := s.getPlazaGroupInviteView(ctx, ownerID, inviteID)
	if err == nil && created {
		s.recordPlazaGroupCreated(ctx, ownerID, view)
	}
	return view, err
}

func (s *Service) KeepPlazaGroup(ctx context.Context, userID, inviteID int64) (PlazaGroupInviteView, error) {
	if err := s.ensureReady(); err != nil {
		return PlazaGroupInviteView{}, err
	}
	if err := s.expirePlazaGroups(ctx); err != nil {
		return PlazaGroupInviteView{}, err
	}

	now := s.now()
	err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		invite, err := s.store.LockPlazaGroupInvite(ctx, tx, inviteID)
		if err != nil {
			return mapPlazaGroupStoreError(err)
		}
		if invite.Status == "kept" {
			return nil
		}
		if invite.Status != "formed" || !now.Before(invite.GroupExpiresAt) {
			return plazaConflict("临时群当前不能发起续留")
		}
		applications, err := s.store.ListPlazaGroupApplications(ctx, tx, invite.ID)
		if err != nil {
			return httpapi.InternalError(err)
		}
		if !isPlazaGroupMember(invite, applications, userID) {
			return plazaForbidden("只有临时群成员可以选择保留")
		}
		if _, err := s.store.CreatePlazaGroupKeepVote(ctx, tx, invite.ID, userID, now); err != nil {
			return httpapi.InternalError(err)
		}
		votes, err := s.store.CountPlazaGroupKeepVotes(ctx, tx, invite.ID)
		if err != nil {
			return httpapi.InternalError(err)
		}
		if votes >= plazaGroupRequiredVotes(applications) {
			if err := s.store.MarkPlazaGroupKept(ctx, tx, invite.ID, now); err != nil {
				return httpapi.InternalError(err)
			}
		}
		return nil
	})
	if err != nil {
		return PlazaGroupInviteView{}, err
	}
	return s.getPlazaGroupInviteView(ctx, userID, inviteID)
}

func (s *Service) formPlazaGroupTx(
	ctx context.Context,
	tx *sql.Tx,
	invite store.PlazaGroupInvite,
	counts store.PlazaGroupApplicationCounts,
	now time.Time,
) (int64, error) {
	if invite.Status == "formed" || invite.Status == "kept" {
		if invite.ChatGroupID.Valid {
			return invite.ChatGroupID.Int64, nil
		}
		return 0, plazaConflict("临时群状态异常，请刷新后重试")
	}
	if counts.Accepted+1 < invite.MinMembers {
		return 0, plazaConflict("同意加入的人数还没有达到成组条件")
	}
	applications, err := s.store.ListPlazaGroupApplications(ctx, tx, invite.ID)
	if err != nil {
		return 0, httpapi.InternalError(err)
	}
	if err := s.store.WaitlistPendingPlazaGroupApplications(ctx, tx, invite.ID, now); err != nil {
		return 0, httpapi.InternalError(err)
	}
	group, err := s.store.CreateChatGroup(ctx, tx, invite.OwnerID, invite.Title, now)
	if err != nil {
		return 0, httpapi.InternalError(err)
	}
	if err := s.store.UpsertChatGroupMember(ctx, tx, group.ID, invite.OwnerID, "owner", now); err != nil {
		return 0, httpapi.InternalError(err)
	}
	for _, application := range applications {
		if application.Status != "accepted" {
			continue
		}
		if err := s.store.UpsertChatGroupMember(ctx, tx, group.ID, application.UserID, "member", now); err != nil {
			return 0, httpapi.InternalError(err)
		}
	}
	if err := s.store.MarkPlazaGroupFormed(ctx, tx, invite.ID, group.ID, now); err != nil {
		return 0, httpapi.InternalError(err)
	}
	return group.ID, nil
}

func (s *Service) getPlazaGroupInviteView(ctx context.Context, userID, inviteID int64) (PlazaGroupInviteView, error) {
	invite, err := s.store.GetPlazaGroupInvite(ctx, nil, inviteID)
	if err != nil {
		return PlazaGroupInviteView{}, mapPlazaGroupStoreError(err)
	}
	return s.buildPlazaGroupInviteView(ctx, nil, userID, invite)
}

func (s *Service) buildPlazaGroupInviteView(
	ctx context.Context,
	tx *sql.Tx,
	viewerID int64,
	invite store.PlazaGroupInvite,
) (PlazaGroupInviteView, error) {
	post, err := s.store.GetPlaygroundPost(ctx, tx, viewerID, invite.PostID)
	if err != nil {
		return PlazaGroupInviteView{}, mapPlazaGroupStoreError(err)
	}
	comments, err := s.store.ListPlaygroundComments(ctx, tx, viewerID, invite.PostID)
	if err != nil {
		return PlazaGroupInviteView{}, httpapi.InternalError(err)
	}
	applications, err := s.store.ListPlazaGroupApplications(ctx, tx, invite.ID)
	if err != nil {
		return PlazaGroupInviteView{}, httpapi.InternalError(err)
	}

	viewerApplication, hasApplication := plazaApplicationByUser(applications, viewerID)
	if viewerID != invite.OwnerID && !hasApplication {
		return PlazaGroupInviteView{}, plazaForbidden("只有洞主或申请者可以查看邀请详情")
	}
	viewerIsMember := isPlazaGroupMember(invite, applications, viewerID)
	revealBoundaryPassed := plazaRevealBoundaryPassed(invite.Status)
	if revealBoundaryPassed && viewerID != invite.OwnerID && viewerApplication.Status != "accepted" {
		viewerIsMember = false
	}

	summary := plazaGroupInviteSummary(invite, applications, comments, viewerID)
	ownerIdentity := PlazaGroupIdentityView{Alias: "洞主"}
	if viewerID == invite.OwnerID || (revealBoundaryPassed && viewerIsMember) {
		owner, loadErr := s.store.GetUserByID(ctx, tx, invite.OwnerID)
		if loadErr != nil {
			return PlazaGroupInviteView{}, httpapi.InternalError(loadErr)
		}
		ownerIdentity = PlazaGroupIdentityView{
			Alias: "洞主", Revealed: true, UserID: owner.ID,
			Nickname: owner.Nickname, Department: owner.Department, Grade: owner.Grade,
		}
	}

	visibleApplications := make([]PlazaGroupApplicationView, 0, len(applications))
	for _, application := range applications {
		if viewerID != invite.OwnerID {
			if revealBoundaryPassed && viewerIsMember {
				if application.Status != "accepted" {
					continue
				}
			} else if application.UserID != viewerID {
				continue
			}
		}
		visibleApplications = append(visibleApplications, plazaGroupApplicationView(
			invite, application, viewerID, viewerIsMember,
		))
	}

	keepVotes, err := s.store.CountPlazaGroupKeepVotes(ctx, tx, invite.ID)
	if err != nil {
		return PlazaGroupInviteView{}, httpapi.InternalError(err)
	}
	hasKeepVote, err := s.store.HasPlazaGroupKeepVote(ctx, tx, invite.ID, viewerID)
	if err != nil {
		return PlazaGroupInviteView{}, httpapi.InternalError(err)
	}
	memberCount := plazaGroupRequiredVotes(applications)
	view := PlazaGroupInviteView{
		PlazaGroupInviteSummary: summary,
		Owner:                   ownerIdentity,
		Applications:            visibleApplications,
		CanForm:                 viewerID == invite.OwnerID && invite.Status == "open" && memberCount >= invite.MinMembers,
		CanKeep:                 invite.Status == "formed" && viewerIsMember && !hasKeepVote,
		KeepVotes:               keepVotes,
		RequiredKeepVotes:       memberCount,
	}
	if invite.ChatGroupID.Valid && viewerIsMember && (invite.Status == "formed" || invite.Status == "kept") {
		chatGroupID := invite.ChatGroupID.Int64
		view.ChatGroupID = &chatGroupID
	}
	_ = post
	return view, nil
}

func plazaGroupInviteSummary(
	invite store.PlazaGroupInvite,
	applications []store.PlazaGroupApplication,
	comments []store.PlaygroundComment,
	viewerID int64,
) PlazaGroupInviteSummary {
	viewerApplication, hasApplication := plazaApplicationByUser(applications, viewerID)
	applicationCount := 0
	acceptedCount := 0
	for _, application := range applications {
		if application.Status != "left" {
			applicationCount++
		}
		if application.Status == "accepted" {
			acceptedCount++
		}
	}
	viewerCommented := false
	if _, ok := latestCommentByUser(comments, viewerID); ok {
		viewerCommented = true
	}
	summary := PlazaGroupInviteSummary{
		ID: invite.ID, PostID: invite.PostID, InviteType: invite.InviteType,
		Title: invite.Title, Summary: invite.Summary,
		ExtractedTime: invite.ExtractedTime, ExtractedLocation: invite.ExtractedLocation,
		MinMembers: invite.MinMembers, MaxMembers: invite.MaxMembers,
		MemberCount: acceptedCount + 1, ApplicationCount: applicationCount,
		RevealStrategy: invite.RevealStrategy,
		Deadline:       invite.Deadline.Format(time.RFC3339),
		GroupExpiresAt: invite.GroupExpiresAt.Format(time.RFC3339),
		Status:         invite.Status, IsOwner: invite.OwnerID == viewerID,
		CanApply:            invite.Status == "open" && invite.OwnerID != viewerID && viewerCommented && !hasApplication,
		InferenceSource:     invite.InferenceSource,
		InferenceConfidence: invite.InferenceConfidence,
		PromptVersion:       invite.PromptVersion,
	}
	if hasApplication {
		summary.ViewerApplicationStatus = viewerApplication.Status
	}
	return summary
}

func plazaGroupApplicationView(
	invite store.PlazaGroupInvite,
	application store.PlazaGroupApplication,
	viewerID int64,
	viewerIsMember bool,
) PlazaGroupApplicationView {
	revealed := application.UserID == viewerID ||
		(invite.OwnerID == viewerID && invite.RevealStrategy == "on_accept" && application.Status == "accepted") ||
		(plazaRevealBoundaryPassed(invite.Status) && viewerIsMember && application.Status == "accepted")
	identity := PlazaGroupIdentityView{Alias: application.AnonymousAlias, Revealed: revealed}
	if revealed {
		identity.UserID = application.UserID
		identity.Nickname = application.ApplicantNickname
		identity.Department = application.ApplicantDepartment
		identity.Grade = application.ApplicantGrade
	}
	return PlazaGroupApplicationView{
		ID: application.ID, Status: application.Status,
		Identity: identity, CreatedAt: application.CreatedAt.Format(time.RFC3339),
	}
}

func preparePlazaGroupInvite(
	post store.PlaygroundPost,
	comments []store.PlaygroundComment,
	input CreatePlazaGroupInviteInput,
	now time.Time,
) (store.CreatePlazaGroupInviteParams, error) {
	inviteType := strings.TrimSpace(input.InviteType)
	if inviteType != "event" && inviteType != "group" {
		return store.CreatePlazaGroupInviteParams{}, invalidPlazaParams("请选择临时活动或同频群组")
	}
	minMembers := input.MinMembers
	if minMembers == 0 {
		minMembers = 2
	}
	maxMembers := input.MaxMembers
	if maxMembers == 0 {
		maxMembers = 6
	}
	if minMembers < 2 || maxMembers < minMembers || maxMembers > maxPlazaGroupMembers {
		return store.CreatePlazaGroupInviteParams{}, invalidPlazaParams("人数需在 2 到 12 人之间")
	}
	revealStrategy := strings.TrimSpace(input.RevealStrategy)
	if revealStrategy == "" {
		revealStrategy = "on_form"
	}
	if revealStrategy != "on_accept" && revealStrategy != "on_form" {
		return store.CreatePlazaGroupInviteParams{}, invalidPlazaParams("请选择有效的身份揭示策略")
	}
	deadline := input.Deadline
	if deadline.IsZero() {
		deadline = now.Add(72 * time.Hour)
	}
	if deadline.Before(now.Add(15*time.Minute)) || deadline.After(now.Add(30*24*time.Hour)) {
		return store.CreatePlazaGroupInviteParams{}, invalidPlazaParams("申请截止时间需在 15 分钟到 30 天内")
	}
	groupExpiresAt := input.GroupExpiresAt
	if groupExpiresAt.IsZero() {
		groupExpiresAt = deadline.Add(7 * 24 * time.Hour)
	}
	if !groupExpiresAt.After(deadline.Add(time.Hour)) || groupExpiresAt.After(deadline.Add(30*24*time.Hour)) {
		return store.CreatePlazaGroupInviteParams{}, invalidPlazaParams("临时群到期时间需晚于截止时间 1 小时，且不超过 30 天")
	}

	discussion := strings.Builder{}
	discussion.WriteString(post.Content)
	for _, comment := range comments {
		discussion.WriteByte('\n')
		discussion.WriteString(comment.Content)
	}
	if err := validatePlazaContent(discussion.String()); err != nil {
		return store.CreatePlazaGroupInviteParams{}, err
	}
	draft := derivePlazaGroupDraft(post, comments, inviteType)
	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = draft.Title
	}
	summary := strings.TrimSpace(input.Summary)
	if summary == "" {
		summary = draft.Summary
	}
	if title == "" || utf8.RuneCountInString(title) > maxPlazaGroupTitleLength ||
		utf8.RuneCountInString(summary) > maxPlazaGroupSummaryLength {
		return store.CreatePlazaGroupInviteParams{}, invalidPlazaParams("活动标题或说明过长")
	}
	if err := validatePlazaContent(title + "\n" + summary); err != nil {
		return store.CreatePlazaGroupInviteParams{}, err
	}

	return store.CreatePlazaGroupInviteParams{
		PostID: post.ID, OwnerID: post.Author.UserID, InviteType: inviteType,
		Title: title, Summary: summary,
		ExtractedTime: draft.ExtractedTime, ExtractedLocation: draft.ExtractedLocation,
		MinMembers: minMembers, MaxMembers: maxMembers,
		RevealStrategy: revealStrategy, Deadline: deadline, GroupExpiresAt: groupExpiresAt,
		InferenceSource: "deterministic", InferenceConfidence: draft.Confidence,
		PromptVersion: plazaGroupPromptVersion, Now: now,
	}, nil
}

type plazaGroupDraft struct {
	Title             string
	Summary           string
	ExtractedTime     string
	ExtractedLocation string
	Confidence        float64
}

func derivePlazaGroupDraft(post store.PlaygroundPost, comments []store.PlaygroundComment, inviteType string) plazaGroupDraft {
	allText := strings.Builder{}
	allText.WriteString(post.Content)
	for _, comment := range comments {
		allText.WriteByte('\n')
		allText.WriteString(comment.Content)
	}
	text := allText.String()
	extractedTime := strings.TrimSpace(plazaTimePattern.FindString(text))
	extractedLocation := strings.TrimSpace(plazaLocationPattern.FindString(text))
	label := "同频群"
	if inviteType == "event" {
		label = "临时活动"
	}
	core := truncateRunes(strings.Join(strings.Fields(post.Content), " "), 24)
	if core == "" {
		core = fmt.Sprintf("树洞 #%d", post.ID)
	}
	title := truncateRunes(core+" · "+label, maxPlazaGroupTitleLength)
	summary := fmt.Sprintf("围绕树洞 #%d 的讨论整理为%s，已有 %d 条匿名回复。成组后请先确认时间、地点和彼此边界。", post.ID, label, len(comments))
	confidence := 0.58
	if extractedTime != "" {
		confidence += 0.1
	}
	if extractedLocation != "" {
		confidence += 0.1
	}
	return plazaGroupDraft{
		Title: title, Summary: summary,
		ExtractedTime: extractedTime, ExtractedLocation: extractedLocation,
		Confidence: confidence,
	}
}

func validatePlazaContent(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	if plazaEmailPattern.MatchString(value) || plazaPhonePattern.MatchString(value) || plazaContactPattern.MatchString(value) {
		return invalidPlazaParams("请移除邮箱、手机号或社交账号，成组后再自愿交换联系方式")
	}
	if plazaDangerPattern.MatchString(value) {
		return invalidPlazaParams("内容包含威胁或侵犯隐私的表达，请修改后再发布")
	}
	return nil
}

func validatePlazaGroupApplicationEligibility(
	invite store.PlazaGroupInvite,
	userID int64,
	hasComment bool,
	now time.Time,
) error {
	if invite.OwnerID == userID {
		return plazaForbidden("洞主已经在临时群中，无需申请")
	}
	if invite.Status != "open" || !now.Before(invite.Deadline) {
		return plazaConflict("这个邀请已经停止接受申请")
	}
	if !hasComment {
		return plazaForbidden("先匿名回复这个树洞，再申请加入")
	}
	return nil
}

func shouldFormPlazaGroup(invite store.PlazaGroupInvite, counts store.PlazaGroupApplicationCounts) bool {
	members := counts.Accepted + 1
	return members >= invite.MaxMembers || (members >= invite.MinMembers && counts.Applied == 0)
}

func plazaRevealBoundaryPassed(status string) bool {
	return status == "formed" || status == "kept" || status == "archived"
}

func isPlazaGroupMember(invite store.PlazaGroupInvite, applications []store.PlazaGroupApplication, userID int64) bool {
	if invite.OwnerID == userID {
		return true
	}
	application, ok := plazaApplicationByUser(applications, userID)
	return ok && application.Status == "accepted"
}

func plazaGroupRequiredVotes(applications []store.PlazaGroupApplication) int {
	count := 1
	for _, application := range applications {
		if application.Status == "accepted" {
			count++
		}
	}
	return count
}

func plazaApplicationByUser(applications []store.PlazaGroupApplication, userID int64) (store.PlazaGroupApplication, bool) {
	for _, application := range applications {
		if application.UserID == userID {
			return application, true
		}
	}
	return store.PlazaGroupApplication{}, false
}

func latestCommentByUser(comments []store.PlaygroundComment, userID int64) (store.PlaygroundComment, bool) {
	for index := len(comments) - 1; index >= 0; index-- {
		if comments[index].Author.UserID == userID {
			return comments[index], true
		}
	}
	return store.PlaygroundComment{}, false
}

func truncateRunes(value string, limit int) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) <= limit {
		return string(runes)
	}
	return string(runes[:limit-1]) + "…"
}

func (s *Service) expirePlazaGroups(ctx context.Context) error {
	return s.store.WithTx(ctx, func(tx *sql.Tx) error {
		if err := s.store.ExpirePlazaGroupInvites(ctx, tx, s.now()); err != nil {
			return httpapi.InternalError(err)
		}
		return nil
	})
}

func (s *Service) now() time.Time {
	if s.nowFunc != nil {
		return s.nowFunc()
	}
	return time.Now()
}

func (s *Service) recordPlazaGroupCreated(ctx context.Context, actorID int64, view PlazaGroupInviteView) {
	if s.recorder == nil {
		return
	}
	_ = s.recorder.Record(ctx, analytics.EventInput{
		Name: analytics.EventPlazaGroupCreated, ActorUserID: actorID,
		SubjectType: "plaza_group", SubjectID: view.ID,
		RequestID: middleware.GetReqID(ctx),
		Properties: map[string]any{
			"source": "api", "inviteType": view.InviteType,
			"memberCount": view.MemberCount, "status": view.Status,
		},
	})
}

func mapPlazaGroupStoreError(err error) error {
	if errors.Is(err, store.ErrNotFound) {
		return httpapi.NewError(http.StatusNotFound, httpapi.CodeUserNotFound, "resource not found")
	}
	return httpapi.InternalError(err)
}

func invalidPlazaParams(message string) error {
	return httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, message)
}

func plazaForbidden(message string) error {
	return httpapi.NewError(http.StatusForbidden, httpapi.CodeForbidden, message)
}

func plazaConflict(message string) error {
	return httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, message)
}
