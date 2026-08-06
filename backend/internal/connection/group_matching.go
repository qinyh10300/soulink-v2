package connection

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi/v5/middleware"

	"soulink-backend/internal/analytics"
	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/meetup"
	"soulink-backend/internal/store"
)

const (
	maxGroupIntentTitleLength       = 100
	maxGroupIntentDescriptionLength = 1200
	maxGroupIntentRoleLength        = 80
	maxGroupMatchCandidates         = 40
	groupMatchPromptVersion         = "group-summary-v1"
)

var groupMatchCategories = map[string]struct{}{
	CategoryMeal:    {},
	CategoryStudy:   {},
	CategorySport:   {},
	CategoryEvent:   {},
	CategoryProject: {},
}

type GroupMatchIntentInput struct {
	Category            string     `json:"category"`
	Title               string     `json:"title"`
	Description         string     `json:"description"`
	RoleOffer           string     `json:"roleOffer"`
	RolesNeeded         []string   `json:"rolesNeeded"`
	Constraints         []string   `json:"constraints"`
	TimeWindow          string     `json:"timeWindow"`
	Location            string     `json:"location"`
	MinMembers          int        `json:"minMembers"`
	MaxMembers          int        `json:"maxMembers"`
	StartsAt            *time.Time `json:"startsAt,omitempty"`
	EndsAt              *time.Time `json:"endsAt,omitempty"`
	CreateCalendarEvent bool       `json:"createCalendarEvent"`
	ExpiresAt           *time.Time `json:"expiresAt,omitempty"`
}

type GroupMatchOrganizerView struct {
	UserID     int64  `json:"userId"`
	Nickname   string `json:"nickname"`
	Department string `json:"department"`
	Grade      string `json:"grade"`
}

type GroupMatchCandidateIdentityView struct {
	UserID     int64  `json:"userId,omitempty"`
	Nickname   string `json:"nickname"`
	Department string `json:"department,omitempty"`
	Grade      string `json:"grade,omitempty"`
	Revealed   bool   `json:"revealed"`
}

type GroupMatchInvitationView struct {
	ID                int64                           `json:"id"`
	GroupIntentID     int64                           `json:"groupIntentId"`
	CandidateIntentID int64                           `json:"candidateIntentId"`
	Candidate         GroupMatchCandidateIdentityView `json:"candidate"`
	SuggestedRole     string                          `json:"suggestedRole"`
	Score             int                             `json:"score"`
	FitReason         string                          `json:"fitReason"`
	Uncertainty       string                          `json:"uncertainty"`
	SharedConstraints []string                        `json:"sharedConstraints"`
	Status            string                          `json:"status"`
	CommonFreeTimes   []meetup.TimeSuggestion         `json:"commonFreeTimes"`
	CreatedAt         time.Time                       `json:"createdAt"`
	UpdatedAt         time.Time                       `json:"updatedAt"`
}

type GroupMatchIntentView struct {
	ID                   int64                      `json:"id"`
	Organizer            GroupMatchOrganizerView    `json:"organizer"`
	Category             string                     `json:"category"`
	Title                string                     `json:"title"`
	Description          string                     `json:"description"`
	RoleOffer            string                     `json:"roleOffer"`
	RolesNeeded          []string                   `json:"rolesNeeded"`
	RoleGaps             []string                   `json:"roleGaps"`
	Constraints          []string                   `json:"constraints"`
	TimeWindow           string                     `json:"timeWindow"`
	Location             string                     `json:"location"`
	MinMembers           int                        `json:"minMembers"`
	MaxMembers           int                        `json:"maxMembers"`
	AcceptedMembers      int                        `json:"acceptedMembers"`
	PendingInvitations   int                        `json:"pendingInvitations"`
	WaitlistedMembers    int                        `json:"waitlistedMembers"`
	StartsAt             *time.Time                 `json:"startsAt,omitempty"`
	EndsAt               *time.Time                 `json:"endsAt,omitempty"`
	CreateCalendarEvent  bool                       `json:"createCalendarEvent"`
	Status               string                     `json:"status"`
	ChatGroupID          *int64                     `json:"chatGroupId,omitempty"`
	CalendarEventID      *int64                     `json:"calendarEventId,omitempty"`
	GroupIntro           string                     `json:"groupIntro,omitempty"`
	ConversationStarters []string                   `json:"conversationStarters"`
	Inference            InferenceView              `json:"inference"`
	CommonFreeTimes      []meetup.TimeSuggestion    `json:"commonFreeTimes"`
	Invitations          []GroupMatchInvitationView `json:"invitations"`
	CanForm              bool                       `json:"canForm"`
	CanLeave             bool                       `json:"canLeave"`
	ExpiresAt            time.Time                  `json:"expiresAt"`
	CreatedAt            time.Time                  `json:"createdAt"`
	UpdatedAt            time.Time                  `json:"updatedAt"`
}

type GroupMatchRecommendationView struct {
	CandidateIntentID int64                   `json:"candidateIntentId"`
	Alias             string                  `json:"alias"`
	RoleOffer         string                  `json:"roleOffer"`
	SuggestedRole     string                  `json:"suggestedRole"`
	Score             int                     `json:"score"`
	FitReason         string                  `json:"fitReason"`
	Uncertainty       string                  `json:"uncertainty"`
	SharedConstraints []string                `json:"sharedConstraints"`
	RemainingRoleGaps []string                `json:"remainingRoleGaps"`
	CommonFreeTimes   []meetup.TimeSuggestion `json:"commonFreeTimes"`
	Inference         InferenceView           `json:"inference"`
	updatedAt         time.Time
	userID            int64
}

type GroupMatchInboxView struct {
	Invitation  GroupMatchInvitationView `json:"invitation"`
	Organizer   GroupMatchOrganizerView  `json:"organizer"`
	Category    string                   `json:"category"`
	Title       string                   `json:"title"`
	Description string                   `json:"description"`
	TimeWindow  string                   `json:"timeWindow"`
	Location    string                   `json:"location"`
	MinMembers  int                      `json:"minMembers"`
	MaxMembers  int                      `json:"maxMembers"`
	Status      string                   `json:"status"`
	ExpiresAt   time.Time                `json:"expiresAt"`
}

type groupMatchResponseDecision struct {
	InvitationStatus string
	ShouldForm       bool
}

func (s *Service) CreateGroupMatchIntent(
	ctx context.Context,
	userID int64,
	input GroupMatchIntentInput,
) (GroupMatchIntentView, error) {
	if err := s.ensureReady(); err != nil {
		return GroupMatchIntentView{}, err
	}
	params, err := s.normalizeGroupMatchInput(userID, input)
	if err != nil {
		return GroupMatchIntentView{}, err
	}
	intentID, err := s.store.CreateGroupMatchIntent(ctx, nil, params)
	if err != nil {
		return GroupMatchIntentView{}, httpapi.InternalError(err)
	}
	return s.GetGroupMatchIntent(ctx, userID, intentID)
}

func (s *Service) ListMyGroupMatchIntents(
	ctx context.Context,
	userID int64,
) ([]GroupMatchIntentView, error) {
	if err := s.cleanupExpiredGroupMatches(ctx); err != nil {
		return nil, err
	}
	items, err := s.store.ListGroupMatchIntentsByOwner(ctx, nil, userID, 30)
	if err != nil {
		return nil, httpapi.InternalError(err)
	}
	views := make([]GroupMatchIntentView, 0, len(items))
	for _, item := range items {
		view, err := s.buildGroupMatchIntentView(ctx, userID, item)
		if err != nil {
			return nil, err
		}
		views = append(views, view)
	}
	return views, nil
}

func (s *Service) GetGroupMatchIntent(
	ctx context.Context,
	viewerID int64,
	intentID int64,
) (GroupMatchIntentView, error) {
	if err := s.cleanupExpiredGroupMatches(ctx); err != nil {
		return GroupMatchIntentView{}, err
	}
	item, err := s.store.GetGroupMatchIntent(ctx, nil, intentID)
	if err != nil {
		return GroupMatchIntentView{}, mapStoreError(err, "组队意图不存在")
	}
	if item.OrganizerID != viewerID {
		invitations, listErr := s.store.ListGroupMatchInvitationsForIntent(ctx, nil, intentID)
		if listErr != nil {
			return GroupMatchIntentView{}, httpapi.InternalError(listErr)
		}
		allowed := false
		for _, invitation := range invitations {
			if invitation.CandidateUserID == viewerID {
				allowed = true
				break
			}
		}
		if !allowed {
			return GroupMatchIntentView{}, forbidden("只有组队参与者可以查看")
		}
	}
	return s.buildGroupMatchIntentView(ctx, viewerID, item)
}

func (s *Service) GroupMatchRecommendations(
	ctx context.Context,
	userID int64,
	intentID int64,
) ([]GroupMatchRecommendationView, error) {
	if err := s.cleanupExpiredGroupMatches(ctx); err != nil {
		return nil, err
	}
	intent, err := s.store.GetGroupMatchIntent(ctx, nil, intentID)
	if err != nil {
		return nil, mapStoreError(err, "组队意图不存在")
	}
	if intent.OrganizerID != userID {
		return nil, forbidden("只有组织者可以查看候选")
	}
	if intent.Status != "open" || !intent.ExpiresAt.After(s.now()) {
		return []GroupMatchRecommendationView{}, nil
	}
	profile, err := s.store.GetMatchingProfile(ctx, nil, userID)
	if errors.Is(err, store.ErrNotFound) {
		profile = store.MatchingProfile{Skills: []string{}, Interests: []string{}}
	} else if err != nil {
		return nil, httpapi.InternalError(err)
	}
	candidates, err := s.store.ListGroupMatchCandidates(
		ctx, nil, intentID, userID, intent.Category, maxGroupMatchCandidates,
	)
	if err != nil {
		return nil, httpapi.InternalError(err)
	}
	views := make([]GroupMatchRecommendationView, 0, len(candidates))
	for _, candidate := range candidates {
		common, commonErr := s.commonGroupTimes(ctx, []int64{userID, candidate.UserID})
		if commonErr != nil {
			return nil, commonErr
		}
		score, reason, uncertainty, shared, gaps, role := scoreGroupMatchCandidate(
			intent, profile, candidate, len(common) > 0,
		)
		views = append(views, GroupMatchRecommendationView{
			CandidateIntentID: candidate.IntentID,
			RoleOffer:         candidate.RoleOffer,
			SuggestedRole:     role,
			Score:             score,
			FitReason:         reason,
			Uncertainty:       uncertainty,
			SharedConstraints: shared,
			RemainingRoleGaps: gaps,
			CommonFreeTimes:   common,
			Inference: InferenceView{
				Source: "deterministic", Confidence: groupMatchConfidence(score),
				PromptVersion: "group-match-ranking-v1", Editable: true,
			},
			updatedAt: candidate.UpdatedAt,
			userID:    candidate.UserID,
		})
	}
	sort.SliceStable(views, func(left, right int) bool {
		if views[left].Score == views[right].Score {
			return views[left].updatedAt.After(views[right].updatedAt)
		}
		return views[left].Score > views[right].Score
	})
	for index := range views {
		views[index].Alias = fmt.Sprintf("候选同学 %c", 'A'+rune(index))
	}
	return views, nil
}

func (s *Service) InviteGroupMatchCandidate(
	ctx context.Context,
	userID int64,
	intentID int64,
	candidateIntentID int64,
) (GroupMatchInvitationView, error) {
	recommendations, err := s.GroupMatchRecommendations(ctx, userID, intentID)
	if err != nil {
		return GroupMatchInvitationView{}, err
	}
	var selected *GroupMatchRecommendationView
	for index := range recommendations {
		if recommendations[index].CandidateIntentID == candidateIntentID {
			selected = &recommendations[index]
			break
		}
	}
	if selected == nil {
		return GroupMatchInvitationView{}, forbidden("这个候选当前不可邀请")
	}
	now := s.now()
	invitation, err := s.store.UpsertGroupMatchInvitation(ctx, nil, store.CreateGroupMatchInvitationParams{
		GroupIntentID:     intentID,
		CandidateIntentID: candidateIntentID,
		CandidateUserID:   selected.userID,
		Score:             selected.Score,
		FitReason:         selected.FitReason,
		Uncertainty:       selected.Uncertainty,
		SharedConstraints: selected.SharedConstraints,
		SuggestedRole:     selected.SuggestedRole,
		Now:               now,
	})
	if err != nil {
		return GroupMatchInvitationView{}, httpapi.InternalError(err)
	}
	return s.groupInvitationView(ctx, invitation, userID, 0, false)
}

func (s *Service) ListMyGroupMatchInvitations(
	ctx context.Context,
	userID int64,
) ([]GroupMatchInboxView, error) {
	if err := s.cleanupExpiredGroupMatches(ctx); err != nil {
		return nil, err
	}
	items, err := s.store.ListGroupMatchInvitationsForCandidate(ctx, nil, userID, 50)
	if err != nil {
		return nil, httpapi.InternalError(err)
	}
	views := make([]GroupMatchInboxView, 0, len(items))
	for _, item := range items {
		invitation, err := s.groupInvitationView(ctx, item, userID, 0, false)
		if err != nil {
			return nil, err
		}
		views = append(views, GroupMatchInboxView{
			Invitation: invitation,
			Organizer: GroupMatchOrganizerView{
				UserID: item.GroupOrganizerID, Nickname: item.GroupOrganizerNickname,
			},
			Category: item.GroupCategory, Title: item.GroupTitle,
			Description: item.GroupDescription, TimeWindow: item.GroupTimeWindow,
			Location: item.GroupLocation, MinMembers: item.GroupMinMembers,
			MaxMembers: item.GroupMaxMembers, Status: item.GroupStatus,
			ExpiresAt: item.GroupExpiresAt,
		})
	}
	return views, nil
}

func (s *Service) RespondGroupMatchInvitation(
	ctx context.Context,
	userID int64,
	invitationID int64,
	response string,
) (GroupMatchIntentView, error) {
	response = strings.ToLower(strings.TrimSpace(response))
	if response != "accepted" && response != "declined" {
		return GroupMatchIntentView{}, invalidParams("请选择接受或拒绝")
	}
	preview, err := s.store.GetGroupMatchInvitation(ctx, nil, invitationID)
	if err != nil {
		return GroupMatchIntentView{}, mapStoreError(err, "邀请不存在")
	}
	if preview.CandidateUserID != userID {
		return GroupMatchIntentView{}, forbidden("只能处理发给自己的邀请")
	}
	now := s.now()
	formed := false
	err = s.store.WithTx(ctx, func(tx *sql.Tx) error {
		lockedIntents, lockErr := s.store.ListGroupMatchIntentsForUpdate(
			ctx, tx, []int64{preview.GroupIntentID, preview.CandidateIntentID},
		)
		if lockErr != nil {
			return lockErr
		}
		if len(lockedIntents) != 2 {
			return store.ErrNotFound
		}
		var groupIntent store.GroupMatchIntent
		var candidateIntent store.GroupMatchIntent
		for _, item := range lockedIntents {
			if item.ID == preview.GroupIntentID {
				groupIntent = item
			} else if item.ID == preview.CandidateIntentID {
				candidateIntent = item
			}
		}
		invitation, lockErr := s.store.GetGroupMatchInvitationForUpdate(ctx, tx, invitationID)
		if lockErr != nil {
			return lockErr
		}
		if invitation.CandidateUserID != userID {
			return forbidden("只能处理发给自己的邀请")
		}
		if groupIntent.Status != "open" || !groupIntent.ExpiresAt.After(now) {
			return httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, "这个组队邀请已经结束")
		}
		if invitation.Status != "invited" {
			return httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, "这个邀请已经处理过")
		}
		if response == "accepted" && candidateIntent.Status != "open" {
			return httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, "你的组队状态已经变化，请刷新后重试")
		}
		counts, countErr := s.store.GetGroupMatchCounts(ctx, tx, groupIntent.ID)
		if countErr != nil {
			return countErr
		}
		decision := decideGroupMatchResponse(
			response, groupIntent.MinMembers, groupIntent.MaxMembers,
			counts.Accepted, counts.Pending,
		)
		if updateErr := s.store.UpdateGroupMatchInvitationStatus(
			ctx, tx, invitationID, decision.InvitationStatus, now,
		); updateErr != nil {
			return updateErr
		}
		if decision.InvitationStatus == "accepted" {
			if updateErr := s.store.UpdateGroupMatchIntentStatus(
				ctx, tx, candidateIntent.ID, userID, "joined", now,
			); updateErr != nil {
				return updateErr
			}
			if updateErr := s.store.DeclineOtherGroupMatchInvitations(
				ctx, tx, userID, invitationID, now,
			); updateErr != nil {
				return updateErr
			}
		}
		if decision.ShouldForm {
			if formErr := s.formGroupMatchIntentTx(ctx, tx, groupIntent, now); formErr != nil {
				return formErr
			}
			formed = true
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return GroupMatchIntentView{}, httpapi.NewError(http.StatusNotFound, httpapi.CodeUserNotFound, "邀请不存在")
		}
		var apiErr *httpapi.AppError
		if errors.As(err, &apiErr) {
			return GroupMatchIntentView{}, err
		}
		return GroupMatchIntentView{}, httpapi.InternalError(err)
	}
	view, err := s.GetGroupMatchIntent(ctx, userID, preview.GroupIntentID)
	if err == nil && formed {
		s.recordGroupMatchCreated(ctx, userID, view)
	}
	return view, err
}

func (s *Service) FormGroupMatchIntent(
	ctx context.Context,
	userID int64,
	intentID int64,
) (GroupMatchIntentView, error) {
	now := s.now()
	err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		intent, lockErr := s.store.GetGroupMatchIntentForUpdate(ctx, tx, intentID)
		if lockErr != nil {
			return lockErr
		}
		if intent.OrganizerID != userID {
			return forbidden("只有组织者可以确认成组")
		}
		if intent.Status != "open" || !intent.ExpiresAt.After(now) {
			return httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, "这个组队意图已经结束")
		}
		counts, countErr := s.store.GetGroupMatchCounts(ctx, tx, intentID)
		if countErr != nil {
			return countErr
		}
		if 1+counts.Accepted < intent.MinMembers {
			return httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, "还没有达到最低成组人数")
		}
		return s.formGroupMatchIntentTx(ctx, tx, intent, now)
	})
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return GroupMatchIntentView{}, httpapi.NewError(http.StatusNotFound, httpapi.CodeUserNotFound, "组队意图不存在")
		}
		var apiErr *httpapi.AppError
		if errors.As(err, &apiErr) {
			return GroupMatchIntentView{}, err
		}
		return GroupMatchIntentView{}, httpapi.InternalError(err)
	}
	view, err := s.GetGroupMatchIntent(ctx, userID, intentID)
	if err == nil {
		s.recordGroupMatchCreated(ctx, userID, view)
	}
	return view, err
}

func (s *Service) LeaveGroupMatchIntent(
	ctx context.Context,
	userID int64,
	intentID int64,
) (GroupMatchIntentView, error) {
	now := s.now()
	err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		intent, lockErr := s.store.GetGroupMatchIntentForUpdate(ctx, tx, intentID)
		if lockErr != nil {
			return lockErr
		}
		if intent.OrganizerID == userID {
			if intent.Status == "formed" {
				return httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, "群主请先在群聊中完成成员安排")
			}
			if intent.Status != "open" {
				return httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, "这个组队意图已经结束")
			}
			return s.store.UpdateGroupMatchIntentStatus(ctx, tx, intent.ID, userID, "cancelled", now)
		}
		invitations, listErr := s.store.ListGroupMatchInvitationsForIntent(ctx, tx, intent.ID)
		if listErr != nil {
			return listErr
		}
		var mine *store.GroupMatchInvitation
		for index := range invitations {
			if invitations[index].CandidateUserID == userID {
				mine = &invitations[index]
				break
			}
		}
		if mine == nil || (mine.Status != "accepted" && mine.Status != "waitlisted") {
			return forbidden("你当前不在这个组队中")
		}
		if updateErr := s.store.UpdateGroupMatchInvitationStatus(ctx, tx, mine.ID, "left", now); updateErr != nil {
			return updateErr
		}
		candidateStatus := "open"
		if !mine.GroupExpiresAt.After(now) {
			candidateStatus = "expired"
		}
		if mine.Status == "accepted" {
			if updateErr := s.store.UpdateGroupMatchIntentStatus(
				ctx, tx, mine.CandidateIntentID, userID, candidateStatus, now,
			); updateErr != nil {
				return updateErr
			}
		}
		if intent.Status == "formed" && mine.Status == "accepted" && intent.ChatGroupID.Valid {
			if updateErr := s.store.RemoveChatGroupMember(ctx, tx, intent.ChatGroupID.Int64, userID, now); updateErr != nil {
				return updateErr
			}
			if promoteErr := s.promoteGroupMatchWaitlist(ctx, tx, intent, now); promoteErr != nil {
				return promoteErr
			}
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return GroupMatchIntentView{}, httpapi.NewError(http.StatusNotFound, httpapi.CodeUserNotFound, "组队意图不存在")
		}
		var apiErr *httpapi.AppError
		if errors.As(err, &apiErr) {
			return GroupMatchIntentView{}, err
		}
		return GroupMatchIntentView{}, httpapi.InternalError(err)
	}
	return s.GetGroupMatchIntent(ctx, userID, intentID)
}

func (s *Service) buildGroupMatchIntentView(
	ctx context.Context,
	viewerID int64,
	item store.GroupMatchIntent,
) (GroupMatchIntentView, error) {
	invitations, err := s.store.ListGroupMatchInvitationsForIntent(ctx, nil, item.ID)
	if err != nil {
		return GroupMatchIntentView{}, httpapi.InternalError(err)
	}
	counts := store.GroupMatchCounts{}
	acceptedUserIDs := []int64{item.OrganizerID}
	acceptedRoles := []string{item.RoleOffer}
	myStatus := ""
	for _, invitation := range invitations {
		switch invitation.Status {
		case "accepted":
			counts.Accepted++
			acceptedUserIDs = append(acceptedUserIDs, invitation.CandidateUserID)
			acceptedRoles = append(acceptedRoles, invitation.SuggestedRole, invitation.CandidateRoleOffer)
		case "invited":
			counts.Pending++
		case "waitlisted":
			counts.Waitlisted++
		case "declined":
			counts.Declined++
		}
		if invitation.CandidateUserID == viewerID {
			myStatus = invitation.Status
		}
	}
	common := []meetup.TimeSuggestion{}
	if len(acceptedUserIDs) >= 2 {
		common, err = s.commonGroupTimes(ctx, acceptedUserIDs)
		if err != nil {
			return GroupMatchIntentView{}, err
		}
	}
	visibleInvitations := make([]GroupMatchInvitationView, 0, len(invitations))
	for index, invitation := range invitations {
		if viewerID != item.OrganizerID {
			if item.Status != "formed" || myStatus != "accepted" {
				if invitation.CandidateUserID != viewerID {
					continue
				}
			} else if invitation.Status != "accepted" && invitation.CandidateUserID != viewerID {
				continue
			}
		}
		view, buildErr := s.groupInvitationView(ctx, invitation, viewerID, index, item.Status == "formed")
		if buildErr != nil {
			return GroupMatchIntentView{}, buildErr
		}
		visibleInvitations = append(visibleInvitations, view)
	}
	view := GroupMatchIntentView{
		ID: item.ID,
		Organizer: GroupMatchOrganizerView{
			UserID: item.OrganizerID, Nickname: item.OrganizerNickname,
			Department: item.OrganizerDepartment, Grade: item.OrganizerGrade,
		},
		Category: item.Category, Title: item.Title, Description: item.Description,
		RoleOffer: item.RoleOffer, RolesNeeded: nonNilStrings(item.RolesNeeded),
		RoleGaps:    remainingGroupRoles(item.RolesNeeded, acceptedRoles),
		Constraints: nonNilStrings(item.Constraints), TimeWindow: item.TimeWindow,
		Location: item.Location, MinMembers: item.MinMembers, MaxMembers: item.MaxMembers,
		AcceptedMembers: 1 + counts.Accepted, PendingInvitations: counts.Pending,
		WaitlistedMembers: counts.Waitlisted, StartsAt: nullableTimePointer(item.StartsAt),
		EndsAt: nullableTimePointer(item.EndsAt), CreateCalendarEvent: item.CreateCalendarEvent,
		Status: item.Status, ChatGroupID: nullableGroupInt64Pointer(item.ChatGroupID),
		CalendarEventID: nullableGroupInt64Pointer(item.CalendarEventID), GroupIntro: item.GroupIntro,
		ConversationStarters: nonNilStrings(item.ConversationStarters),
		Inference: InferenceView{
			Source: item.InferenceSource, Model: item.InferenceModel,
			Confidence: item.InferenceConfidence, PromptVersion: item.PromptVersion,
			Editable: true,
		},
		CommonFreeTimes: common, Invitations: visibleInvitations,
		CanForm: viewerID == item.OrganizerID && item.Status == "open" &&
			1+counts.Accepted >= item.MinMembers,
		CanLeave: (viewerID == item.OrganizerID && item.Status == "open") ||
			(viewerID != item.OrganizerID && (myStatus == "accepted" || myStatus == "waitlisted")),
		ExpiresAt: item.ExpiresAt, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
	}
	return view, nil
}

func (s *Service) groupInvitationView(
	ctx context.Context,
	item store.GroupMatchInvitation,
	viewerID int64,
	index int,
	groupFormed bool,
) (GroupMatchInvitationView, error) {
	identity := groupMatchCandidateIdentity(item, viewerID, index, groupFormed)
	common, err := s.commonGroupTimes(ctx, []int64{item.GroupOrganizerID, item.CandidateUserID})
	if err != nil {
		return GroupMatchInvitationView{}, err
	}
	return GroupMatchInvitationView{
		ID: item.ID, GroupIntentID: item.GroupIntentID,
		CandidateIntentID: item.CandidateIntentID, Candidate: identity,
		SuggestedRole: item.SuggestedRole, Score: item.Score,
		FitReason: item.FitReason, Uncertainty: item.Uncertainty,
		SharedConstraints: nonNilStrings(item.SharedConstraints), Status: item.Status,
		CommonFreeTimes: common, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
	}, nil
}

func groupMatchCandidateIdentity(
	item store.GroupMatchInvitation,
	viewerID int64,
	index int,
	groupFormed bool,
) GroupMatchCandidateIdentityView {
	revealed := item.CandidateUserID == viewerID ||
		(item.GroupOrganizerID == viewerID && item.Status == "accepted") ||
		(groupFormed && item.Status == "accepted")
	if !revealed {
		return GroupMatchCandidateIdentityView{
			Nickname: fmt.Sprintf("候选同学 %c", 'A'+rune(index)), Revealed: false,
		}
	}
	return GroupMatchCandidateIdentityView{
		UserID: item.CandidateUserID, Nickname: item.CandidateNickname,
		Department: item.CandidateDepartment, Grade: item.CandidateGrade, Revealed: true,
	}
}

func (s *Service) formGroupMatchIntentTx(
	ctx context.Context,
	tx *sql.Tx,
	intent store.GroupMatchIntent,
	now time.Time,
) error {
	invitations, err := s.store.ListGroupMatchInvitationsForIntent(ctx, tx, intent.ID)
	if err != nil {
		return err
	}
	accepted := make([]store.GroupMatchInvitation, 0)
	for _, invitation := range invitations {
		if invitation.Status == "accepted" {
			accepted = append(accepted, invitation)
		}
	}
	if 1+len(accepted) < intent.MinMembers {
		return httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, "还没有达到最低成组人数")
	}
	intro, starters := deterministicGroupIntro(intent, accepted)
	group, err := s.store.CreateChatGroup(ctx, tx, intent.OrganizerID, intent.Title, now)
	if err != nil {
		return err
	}
	if err := s.store.UpsertChatGroupMember(ctx, tx, group.ID, intent.OrganizerID, "owner", now); err != nil {
		return err
	}
	for _, invitation := range accepted {
		if err := s.store.UpsertChatGroupMember(
			ctx, tx, group.ID, invitation.CandidateUserID, "member", now,
		); err != nil {
			return err
		}
	}
	calendarEventID := sql.NullInt64{}
	if intent.CreateCalendarEvent && intent.StartsAt.Valid && intent.EndsAt.Valid {
		event, err := s.store.CreateCalendarEvent(ctx, tx, store.CreateCalendarEventParams{
			OrganizerID: intent.OrganizerID, EventType: calendarEventTypeForGroup(intent.Category),
			Title: intent.Title, Description: intro, Location: intent.Location,
			StartsAt: intent.StartsAt.Time, EndsAt: intent.EndsAt.Time,
			Visibility: "busy_only", Now: now,
		})
		if err != nil {
			return err
		}
		calendarEventID = sql.NullInt64{Int64: event.ID, Valid: true}
		for _, invitation := range accepted {
			if err := s.store.AddCalendarAttendee(ctx, tx, event.ID, invitation.CandidateUserID, now); err != nil {
				return err
			}
			if err := s.store.UpdateCalendarAttendeeStatus(
				ctx, tx, event.ID, invitation.CandidateUserID, "accepted", now,
			); err != nil {
				return err
			}
		}
	}
	return s.store.FormGroupMatchIntent(
		ctx, tx, intent.ID, group.ID, calendarEventID, intro, starters,
		"deterministic", "", 0.82, groupMatchPromptVersion, now,
	)
}

func (s *Service) promoteGroupMatchWaitlist(
	ctx context.Context,
	tx *sql.Tx,
	intent store.GroupMatchIntent,
	now time.Time,
) error {
	waitlisted, err := s.store.GetOldestWaitlistedGroupMatchInvitation(ctx, tx, intent.ID)
	if errors.Is(err, store.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	candidateIntent, err := s.store.GetGroupMatchIntentForUpdate(ctx, tx, waitlisted.CandidateIntentID)
	if err != nil {
		return err
	}
	if candidateIntent.Status != "open" || !candidateIntent.ExpiresAt.After(now) {
		return s.store.UpdateGroupMatchInvitationStatus(ctx, tx, waitlisted.ID, "declined", now)
	}
	if err := s.store.UpdateGroupMatchInvitationStatus(ctx, tx, waitlisted.ID, "accepted", now); err != nil {
		return err
	}
	if err := s.store.UpdateGroupMatchIntentStatus(
		ctx, tx, candidateIntent.ID, candidateIntent.OrganizerID, "joined", now,
	); err != nil {
		return err
	}
	if err := s.store.DeclineOtherGroupMatchInvitations(
		ctx, tx, candidateIntent.OrganizerID, waitlisted.ID, now,
	); err != nil {
		return err
	}
	return s.store.UpsertChatGroupMember(
		ctx, tx, intent.ChatGroupID.Int64, candidateIntent.OrganizerID, "member", now,
	)
}

func (s *Service) commonGroupTimes(
	ctx context.Context,
	userIDs []int64,
) ([]meetup.TimeSuggestion, error) {
	if len(userIDs) < 2 {
		return []meetup.TimeSuggestion{}, nil
	}
	now := s.now()
	to := now.Add(14 * 24 * time.Hour)
	windows, err := s.store.ListAvailabilityWindows(ctx, nil, userIDs)
	if err != nil {
		return nil, httpapi.InternalError(err)
	}
	busy, err := s.store.ListBusyIntervals(ctx, nil, userIDs, now, to, 0)
	if err != nil {
		return nil, httpapi.InternalError(err)
	}
	suggestions, _ := meetup.BuildCommonSuggestions(
		windows, busy, userIDs, now, to, 60*time.Minute,
	)
	return suggestions, nil
}

func (s *Service) cleanupExpiredGroupMatches(ctx context.Context) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	if err := s.store.ExpireGroupMatchIntents(ctx, nil, s.now()); err != nil {
		return httpapi.InternalError(err)
	}
	return nil
}

func (s *Service) normalizeGroupMatchInput(
	userID int64,
	input GroupMatchIntentInput,
) (store.CreateGroupMatchIntentParams, error) {
	now := s.now()
	input.Category = strings.ToLower(strings.TrimSpace(input.Category))
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	input.RoleOffer = strings.TrimSpace(input.RoleOffer)
	input.TimeWindow = strings.TrimSpace(input.TimeWindow)
	input.Location = strings.TrimSpace(input.Location)
	if _, ok := groupMatchCategories[input.Category]; !ok {
		return store.CreateGroupMatchIntentParams{}, invalidParams("请选择饭局、自习、运动、活动或项目团队")
	}
	if input.Title == "" || utf8.RuneCountInString(input.Title) > maxGroupIntentTitleLength {
		return store.CreateGroupMatchIntentParams{}, invalidParams("组队标题需要在 1 到 100 个字之间")
	}
	if utf8.RuneCountInString(input.Description) > maxGroupIntentDescriptionLength ||
		utf8.RuneCountInString(input.RoleOffer) > maxGroupIntentRoleLength ||
		utf8.RuneCountInString(input.TimeWindow) > 200 ||
		utf8.RuneCountInString(input.Location) > 160 {
		return store.CreateGroupMatchIntentParams{}, invalidParams("组队信息内容过长")
	}
	if input.RoleOffer == "" {
		input.RoleOffer = defaultGroupRole(input.Category)
	}
	rolesNeeded, err := normalizeStringList(input.RolesNeeded, 10, 80)
	if err != nil {
		return store.CreateGroupMatchIntentParams{}, err
	}
	constraints, err := normalizeStringList(input.Constraints, 12, 100)
	if err != nil {
		return store.CreateGroupMatchIntentParams{}, err
	}
	if input.MinMembers == 0 {
		input.MinMembers = 3
	}
	if input.MaxMembers == 0 {
		input.MaxMembers = 5
	}
	if input.MinMembers < 2 || input.MaxMembers > 12 || input.MaxMembers < input.MinMembers {
		return store.CreateGroupMatchIntentParams{}, invalidParams("成组人数需要在 2 到 12 人之间")
	}
	startsAt := sql.NullTime{}
	endsAt := sql.NullTime{}
	if (input.StartsAt == nil) != (input.EndsAt == nil) {
		return store.CreateGroupMatchIntentParams{}, invalidParams("请同时填写活动开始和结束时间")
	}
	if input.StartsAt != nil {
		if input.StartsAt.Before(now.Add(-5*time.Minute)) || !input.EndsAt.After(*input.StartsAt) ||
			input.EndsAt.Sub(*input.StartsAt) > 12*time.Hour {
			return store.CreateGroupMatchIntentParams{}, invalidParams("活动时间需要在未来且不超过 12 小时")
		}
		startsAt = sql.NullTime{Time: *input.StartsAt, Valid: true}
		endsAt = sql.NullTime{Time: *input.EndsAt, Valid: true}
	}
	if input.CreateCalendarEvent && !startsAt.Valid {
		return store.CreateGroupMatchIntentParams{}, invalidParams("创建日程前需要填写活动时间")
	}
	expiresAt := now.Add(14 * 24 * time.Hour)
	if input.ExpiresAt != nil {
		expiresAt = *input.ExpiresAt
	}
	if !expiresAt.After(now.Add(time.Hour)) || expiresAt.After(now.Add(30*24*time.Hour)) {
		return store.CreateGroupMatchIntentParams{}, invalidParams("组队截止时间需要在未来 1 小时到 30 天内")
	}
	return store.CreateGroupMatchIntentParams{
		OrganizerID: userID, Category: input.Category, Title: input.Title,
		Description: input.Description, RoleOffer: input.RoleOffer,
		RolesNeeded: rolesNeeded, Constraints: constraints,
		TimeWindow: input.TimeWindow, Location: input.Location,
		MinMembers: input.MinMembers, MaxMembers: input.MaxMembers,
		StartsAt: startsAt, EndsAt: endsAt,
		CreateCalendarEvent: input.CreateCalendarEvent,
		ExpiresAt:           expiresAt, Now: now,
	}, nil
}

func scoreGroupMatchCandidate(
	source store.GroupMatchIntent,
	sourceProfile store.MatchingProfile,
	candidate store.GroupMatchCandidate,
	hasCommonCalendarTime bool,
) (int, string, string, []string, []string, string) {
	score := 36
	reasons := []string{"同一组队场景"}
	uncertainties := make([]string, 0, 3)
	available := append([]string{candidate.RoleOffer}, candidate.Skills...)
	roleMatches := fuzzyIntersection(source.RolesNeeded, available)
	if len(roleMatches) > 0 {
		score += 24 + minInt(len(roleMatches)-1, 2)*4
		reasons = append(reasons, "可补位 "+strings.Join(roleMatches, "、"))
	} else if len(source.RolesNeeded) > 0 {
		uncertainties = append(uncertainties, "具体角色补位仍需确认")
	}
	reverseAvailable := append([]string{source.RoleOffer}, sourceProfile.Skills...)
	reverseMatches := fuzzyIntersection(candidate.RolesNeeded, reverseAvailable)
	if len(reverseMatches) > 0 {
		score += 8
		reasons = append(reasons, "双方需求互补")
	}
	shared := fuzzyIntersection(source.Constraints, candidate.Constraints)
	if len(shared) > 0 {
		score += minInt(len(shared), 3) * 4
		reasons = append(reasons, "共同约束 "+strings.Join(shared, "、"))
	}
	if source.Location != "" && candidate.Location != "" &&
		textSignalsOverlap(source.Location, candidate.Location) {
		score += 8
		reasons = append(reasons, "地点方向一致")
	} else if source.Location != "" && candidate.Location != "" {
		uncertainties = append(uncertainties, "集合地点需要进一步协调")
	}
	if hasCommonCalendarTime {
		score += 10
		reasons = append(reasons, "未来两周存在共同空闲")
	} else if hasSharedTimeSignal(source.TimeWindow, candidate.TimeWindow) ||
		textSignalsOverlap(source.TimeWindow, candidate.TimeWindow) {
		score += 5
		reasons = append(reasons, "声明时间有交集")
		uncertainties = append(uncertainties, "请用日历候选时段最终确认")
	} else {
		uncertainties = append(uncertainties, "共同空闲时间尚未确认")
	}
	if rangesOverlap(source.MinMembers, source.MaxMembers, candidate.MinMembers, candidate.MaxMembers) {
		score += 5
	} else {
		uncertainties = append(uncertainties, "双方预期人数不同")
	}
	if score > 100 {
		score = 100
	}
	if len(uncertainties) == 0 {
		uncertainties = append(uncertainties, "仍需本人确认节奏与边界")
	}
	suggestedRole := candidate.RoleOffer
	if len(roleMatches) > 0 {
		suggestedRole = roleMatches[0]
	}
	if suggestedRole == "" {
		suggestedRole = defaultGroupRole(source.Category)
	}
	return score, strings.Join(uniqueStrings(reasons, 4), "；"),
		strings.Join(uniqueStrings(uncertainties, 3), "；"), nonNilStrings(shared),
		remainingGroupRoles(source.RolesNeeded, available), suggestedRole
}

func decideGroupMatchResponse(
	response string,
	minMembers int,
	maxMembers int,
	acceptedInvitations int,
	pendingInvitations int,
) groupMatchResponseDecision {
	if response == "declined" {
		return groupMatchResponseDecision{
			InvitationStatus: "declined",
			ShouldForm:       1+acceptedInvitations >= minMembers && pendingInvitations-1 <= 0,
		}
	}
	if 1+acceptedInvitations >= maxMembers {
		return groupMatchResponseDecision{InvitationStatus: "waitlisted"}
	}
	acceptedAfter := acceptedInvitations + 1
	pendingAfter := pendingInvitations - 1
	memberCount := 1 + acceptedAfter
	return groupMatchResponseDecision{
		InvitationStatus: "accepted",
		ShouldForm: memberCount >= minMembers &&
			(memberCount >= maxMembers || pendingAfter <= 0),
	}
}

func deterministicGroupIntro(
	intent store.GroupMatchIntent,
	accepted []store.GroupMatchInvitation,
) (string, []string) {
	roles := []string{intent.RoleOffer}
	for _, invitation := range accepted {
		roles = append(roles, invitation.SuggestedRole)
	}
	roles = uniqueStrings(roles, 8)
	intro := fmt.Sprintf(
		"这是一个围绕“%s”组成的 %d 人%s小组。当前可互补的角色包括：%s。",
		intent.Title, 1+len(accepted), groupCategoryLabel(intent.Category), strings.Join(roles, "、"),
	)
	if intent.TimeWindow != "" || intent.Location != "" {
		intro += " 初步安排为"
		if intent.TimeWindow != "" {
			intro += intent.TimeWindow
		}
		if intent.Location != "" {
			intro += "，地点倾向 " + intent.Location
		}
		intro += "，以成员最终确认版本为准。"
	}
	starters := []string{
		"大家先用一句话介绍自己能贡献的角色，以及这次最希望得到什么。",
		"一起确认时间、地点和无法妥协的边界，未确认内容先保留为草案。",
		"选出一个最小下一步，并约定负责人和下一次同步时间。",
	}
	return intro, starters
}

func remainingGroupRoles(needed []string, available []string) []string {
	result := make([]string, 0, len(needed))
	for _, role := range needed {
		if len(fuzzyIntersection([]string{role}, available)) == 0 {
			result = append(result, role)
		}
	}
	return uniqueStrings(result, 10)
}

func textSignalsOverlap(left string, right string) bool {
	left = strings.ToLower(strings.TrimSpace(left))
	right = strings.ToLower(strings.TrimSpace(right))
	if left == "" || right == "" {
		return false
	}
	if strings.Contains(left, right) || strings.Contains(right, left) {
		return true
	}
	for _, token := range strings.FieldsFunc(left, func(value rune) bool {
		return value == ' ' || value == '、' || value == '，' || value == ',' || value == '/' || value == '或'
	}) {
		if utf8.RuneCountInString(token) >= 2 && strings.Contains(right, token) {
			return true
		}
	}
	return false
}

func rangesOverlap(leftMin, leftMax, rightMin, rightMax int) bool {
	return leftMin <= rightMax && rightMin <= leftMax
}

func groupMatchConfidence(score int) float64 {
	confidence := 0.5 + float64(score)/250
	if confidence > 0.9 {
		return 0.9
	}
	return confidence
}

func defaultGroupRole(category string) string {
	switch category {
	case CategoryProject:
		return "项目协作者"
	case CategoryStudy:
		return "学习伙伴"
	case CategorySport:
		return "运动伙伴"
	case CategoryMeal:
		return "饭局成员"
	default:
		return "活动成员"
	}
}

func groupCategoryLabel(category string) string {
	labels := map[string]string{
		CategoryMeal: "饭局", CategoryStudy: "自习", CategorySport: "运动",
		CategoryEvent: "活动", CategoryProject: "项目",
	}
	return labels[category]
}

func calendarEventTypeForGroup(category string) string {
	switch category {
	case CategoryMeal:
		return "meal"
	case CategoryStudy:
		return "study"
	case CategorySport:
		return "sports"
	case CategoryProject:
		return "project"
	default:
		return "other"
	}
}

func nullableTimePointer(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time
	return &result
}

func nullableGroupInt64Pointer(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	result := value.Int64
	return &result
}

func minInt(left int, right int) int {
	if left < right {
		return left
	}
	return right
}

func (s *Service) recordGroupMatchCreated(
	ctx context.Context,
	actorID int64,
	view GroupMatchIntentView,
) {
	if s.recorder == nil {
		return
	}
	_ = s.recorder.Record(ctx, analytics.EventInput{
		Name: analytics.EventGroupMatchCreated, ActorUserID: actorID,
		SubjectType: "group_match", SubjectID: view.ID,
		RequestID: middleware.GetReqID(ctx),
		Properties: map[string]any{
			"source": "api", "category": view.Category,
			"memberCount": view.AcceptedMembers, "status": view.Status,
		},
	})
}
