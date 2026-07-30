package match

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/meta"
	"soulink-backend/internal/school"
	"soulink-backend/internal/store"
)

const (
	ResultStatusMatched = "matched"
	ResultStatusFailed  = "failed"
)

type Option func(*Service)

type Service struct {
	store          *store.Store
	resultCache    ResultCache
	resultCacheTTL time.Duration
	nowFunc        func() time.Time
}

type ContactView struct {
	Type  string `json:"type"`
	Label string `json:"label"`
	Value string `json:"value"`
}

type PartnerView struct {
	UserID     int64         `json:"userId,omitempty"`
	Nickname   string        `json:"nickname,omitempty"`
	SchoolName string        `json:"schoolName,omitempty"`
	Department string        `json:"department,omitempty"`
	Grade      string        `json:"grade,omitempty"`
	Contacts   []ContactView `json:"contacts"`
}

type ResultView struct {
	MatchID      int64           `json:"matchId"`
	CycleID      int64           `json:"cycleId"`
	CycleKey     string          `json:"cycleKey"`
	CycleTitle   string          `json:"cycleTitle"`
	RevealStatus string          `json:"revealStatus"`
	RevealAt     string          `json:"revealAt,omitempty"`
	ResultStatus string          `json:"resultStatus"`
	Message      string          `json:"message"`
	Score        *float64        `json:"score,omitempty"`
	Reason       string          `json:"reason,omitempty"`
	ReasonJSON   json.RawMessage `json:"reasonJson,omitempty"`
	Partner      *PartnerView    `json:"partner,omitempty"`
}

type HistoryResult struct {
	Items []ResultView `json:"items"`
}

type CycleStatusResult struct {
	HasCurrentCycle    bool   `json:"hasCurrentCycle"`
	CycleID            int64  `json:"cycleId,omitempty"`
	CycleKey           string `json:"cycleKey,omitempty"`
	CycleTitle         string `json:"cycleTitle,omitempty"`
	RevealStatus       string `json:"revealStatus,omitempty"`
	RevealAt           string `json:"revealAt,omitempty"`
	ParticipantCount   int    `json:"participantCount,omitempty"`
	MatchedUserCount   int    `json:"matchedUserCount,omitempty"`
	SuccessRatePercent int    `json:"successRatePercent,omitempty"`
	HasUserResult      bool   `json:"hasUserResult"`
	UserResultStatus   string `json:"userResultStatus,omitempty"`
	UserRevealAt       string `json:"userRevealAt,omitempty"`
	CanViewUserResult  bool   `json:"canViewUserResult"`
}

type ContactResult struct {
	Action         string `json:"action"`
	FriendshipID   int64  `json:"friendshipId,omitempty"`
	ConversationID int64  `json:"conversationId,omitempty"`
	Message        string `json:"message,omitempty"`
}

type RecordCycleInput struct {
	CycleKey             string
	Title                string
	QuestionnaireKey     string
	QuestionnaireVersion string
	RevealStatus         string
	RevealAt             time.Time
	Results              []RecordUserResultInput
}

type RecordUserResultInput struct {
	UserID          int64
	ResultStatus    string
	PartnerUserID   *int64
	MatchScore      *float64
	MatchReason     string
	MatchReasonJSON json.RawMessage
	Message         string
}

type SimpleVersionMatchInput struct {
	QuestionnaireKey     string
	QuestionnaireVersion string
	CycleKey             string
	Title                string
	RevealStatus         string
	RevealAt             time.Time
	RandomSeed           *int64
}

type partnerContactSnapshot struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

func WithRevealManager(_ any) Option {
	return func(service *Service) {
		_ = service
	}
}

func WithResultCache(cache ResultCache, ttl time.Duration) Option {
	return func(service *Service) {
		service.resultCache = cache
		if ttl > 0 {
			service.resultCacheTTL = ttl
		}
	}
}

func WithNowFunc(nowFunc func() time.Time) Option {
	return func(service *Service) {
		if nowFunc != nil {
			service.nowFunc = nowFunc
		}
	}
}

// NewService 创建匹配结果服务实例。
func NewService(dataStore *store.Store, options ...Option) *Service {
	service := &Service{store: dataStore, resultCacheTTL: 24 * time.Hour, nowFunc: time.Now}
	for _, option := range options {
		if option != nil {
			option(service)
		}
	}
	return service
}

// RecordCycleResults 记录一个匹配周期内每位用户的结果，供算法产出落库使用。
func (s *Service) RecordCycleResults(ctx context.Context, input RecordCycleInput) (store.MatchCycle, error) {
	if err := s.ensureReady(); err != nil {
		return store.MatchCycle{}, err
	}

	input.CycleKey = strings.TrimSpace(input.CycleKey)
	input.Title = strings.TrimSpace(input.Title)
	input.QuestionnaireKey = strings.TrimSpace(input.QuestionnaireKey)
	input.QuestionnaireVersion = strings.TrimSpace(input.QuestionnaireVersion)
	input.RevealStatus = normalizeRevealStatus(input.RevealStatus)
	if input.CycleKey == "" || input.Title == "" || len(input.Results) == 0 {
		return store.MatchCycle{}, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
	}
	if input.RevealAt.IsZero() {
		input.RevealAt = s.currentTime()
	}

	seenUsers := make(map[int64]struct{}, len(input.Results))
	matchedUserCount := 0
	for _, result := range input.Results {
		if _, exists := seenUsers[result.UserID]; exists {
			return store.MatchCycle{}, httpapi.NewError(http.StatusBadRequest, httpapi.CodeConflict, "数据冲突")
		}
		seenUsers[result.UserID] = struct{}{}
		if normalizeResultStatus(result.ResultStatus) == ResultStatusMatched {
			matchedUserCount++
		}
	}

	var cycle store.MatchCycle
	now := s.currentTime()
	currentUserIDs := make([]int64, 0, len(input.Results))
	for _, result := range input.Results {
		currentUserIDs = append(currentUserIDs, result.UserID)
	}
	staleUserIDs := make([]int64, 0)
	err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		createdCycle, err := s.store.UpsertMatchCycle(ctx, tx, store.UpsertMatchCycleParams{
			CycleKey:             input.CycleKey,
			Title:                input.Title,
			QuestionnaireKey:     input.QuestionnaireKey,
			QuestionnaireVersion: input.QuestionnaireVersion,
			RevealStatus:         input.RevealStatus,
			RevealAt:             input.RevealAt,
			ParticipantCount:     len(input.Results),
			MatchedUserCount:     matchedUserCount,
			CreatedAt:            now,
			UpdatedAt:            now,
		})
		if err != nil {
			return httpapi.InternalError(err)
		}
		cycle = createdCycle

		deletedUserIDs, err := s.store.DeleteMatchResultsByCycleExceptUsers(ctx, tx, createdCycle.ID, currentUserIDs)
		if err != nil {
			return httpapi.InternalError(err)
		}
		staleUserIDs = deletedUserIDs

		for _, item := range input.Results {
			params, err := s.prepareResultParams(ctx, tx, createdCycle.ID, item, now)
			if err != nil {
				return err
			}
			if _, err := s.store.UpsertUserMatchResult(ctx, tx, params); err != nil {
				return httpapi.InternalError(err)
			}
		}
		return nil
	})
	if err != nil {
		return store.MatchCycle{}, err
	}

	for _, userID := range staleUserIDs {
		s.deleteCurrentResultFromCache(ctx, cycle.ID, userID)
	}
	s.prewarmCurrentResultCache(ctx, cycle)
	return cycle, nil
}

// RefreshCycleRevealTimes 当前仅用于刷新缓存，已不再执行用户级揭晓时间同步。
func (s *Service) RefreshCycleRevealTimes(ctx context.Context, cycleID int64) error {
	if err := s.ensureReady(); err != nil {
		return err
	}

	cycle, err := s.store.GetMatchCycleByID(ctx, nil, cycleID)
	if err != nil {
		if err == store.ErrNotFound {
			return httpapi.NewError(http.StatusNotFound, httpapi.CodeNoCurrentMatch, "匹配周期不存在")
		}
		return httpapi.InternalError(err)
	}

	s.prewarmCurrentResultCache(ctx, cycle)
	return nil
}

// RunSimpleVersionMatch 按问卷版本筛选用户，执行极简随机两两匹配并落库结果。
func (s *Service) RunSimpleVersionMatch(ctx context.Context, input SimpleVersionMatchInput) (store.MatchCycle, error) {
	if err := s.ensureReady(); err != nil {
		return store.MatchCycle{}, err
	}

	input.QuestionnaireKey = strings.TrimSpace(input.QuestionnaireKey)
	input.QuestionnaireVersion = strings.TrimSpace(input.QuestionnaireVersion)
	input.CycleKey = strings.TrimSpace(input.CycleKey)
	input.Title = strings.TrimSpace(input.Title)
	input.RevealStatus = normalizeRevealStatus(input.RevealStatus)
	if input.QuestionnaireVersion == "" || input.CycleKey == "" || input.Title == "" {
		return store.MatchCycle{}, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
	}
	if input.RevealAt.IsZero() {
		input.RevealAt = s.currentTime()
	}

	participants, err := s.store.ListCompletedQuestionnaireUsers(ctx, nil, input.QuestionnaireKey, input.QuestionnaireVersion)
	if err != nil {
		return store.MatchCycle{}, httpapi.InternalError(err)
	}
	if len(participants) == 0 {
		return store.MatchCycle{}, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "没有可参与匹配的用户")
	}

	seed := s.currentTime().UnixNano()
	if input.RandomSeed != nil {
		seed = *input.RandomSeed
	}
	rng := rand.New(rand.NewSource(seed))

	return s.RecordCycleResults(ctx, RecordCycleInput{
		CycleKey:             input.CycleKey,
		Title:                input.Title,
		QuestionnaireKey:     input.QuestionnaireKey,
		QuestionnaireVersion: input.QuestionnaireVersion,
		RevealStatus:         input.RevealStatus,
		RevealAt:             input.RevealAt,
		Results:              buildSimpleVersionResults(participants, input.QuestionnaireVersion, rng),
	})
}

// GetCurrent 返回当前匹配周期下当前用户的匹配结果。
func (s *Service) GetCurrent(ctx context.Context, userID int64) (ResultView, error) {
	if err := s.ensureReady(); err != nil {
		return ResultView{}, err
	}

	cycle, err := s.store.GetLatestMatchCycle(ctx, nil)
	if err != nil {
		if err == store.ErrNotFound {
			return ResultView{}, httpapi.NewError(http.StatusNotFound, httpapi.CodeNoCurrentMatch, "当前暂无匹配结果")
		}
		return ResultView{}, httpapi.InternalError(err)
	}
	now := s.currentTime()
	cycle, err = s.resetNextRoundParticipationIfNeeded(ctx, cycle, now)
	if err != nil {
		return ResultView{}, err
	}
	if cycle.RevealAt.After(now) {
		return ResultView{}, httpapi.NewError(http.StatusNotFound, httpapi.CodeNoCurrentMatch, "当前暂无匹配结果")
	}

	if cachedView, ok := s.loadCurrentResultFromCache(ctx, cycle.ID, userID); ok {
		return cachedView, nil
	}

	record, err := s.store.GetUserMatchResultByCycle(ctx, nil, userID, cycle.ID)
	if err != nil {
		if err == store.ErrNotFound {
			return ResultView{}, httpapi.NewError(http.StatusNotFound, httpapi.CodeNoCurrentMatch, "当前暂无匹配结果")
		}
		return ResultView{}, httpapi.InternalError(err)
	}

	view, err := buildResultView(record)
	if err != nil {
		return ResultView{}, err
	}
	s.storeCurrentResultInCache(ctx, cycle.ID, userID, view)
	return view, nil
}

// GetHistory 返回当前用户的历史匹配结果列表。
func (s *Service) GetHistory(ctx context.Context, userID int64, limit int) (HistoryResult, error) {
	if err := s.ensureReady(); err != nil {
		return HistoryResult{}, err
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	records, err := s.store.ListUserMatchResults(ctx, nil, userID, limit)
	if err != nil {
		return HistoryResult{}, httpapi.InternalError(err)
	}

	cutoff := s.currentTime()
	items := make([]ResultView, 0, len(records))
	for _, record := range records {
		if !record.CycleRevealAt.IsZero() && record.CycleRevealAt.After(cutoff) {
			continue
		}
		view, err := buildResultView(record)
		if err != nil {
			return HistoryResult{}, err
		}
		items = append(items, view)
	}
	return HistoryResult{Items: items}, nil
}

// GetDetail 返回当前用户的一条历史匹配结果详情。
func (s *Service) GetDetail(ctx context.Context, userID, matchID int64) (ResultView, error) {
	if err := s.ensureReady(); err != nil {
		return ResultView{}, err
	}

	record, err := s.store.GetUserMatchResultByID(ctx, nil, userID, matchID)
	if err != nil {
		if err == store.ErrNotFound {
			return ResultView{}, httpapi.NewError(http.StatusNotFound, httpapi.CodeNoCurrentMatch, "匹配结果不存在")
		}
		return ResultView{}, httpapi.InternalError(err)
	}
	if !record.CycleRevealAt.IsZero() && record.CycleRevealAt.After(s.currentTime()) {
		return ResultView{}, httpapi.NewError(http.StatusNotFound, httpapi.CodeNoCurrentMatch, "匹配结果不存在")
	}
	return buildResultView(record)
}

// GetCycleStatus 返回最新匹配周期的揭晓状态及当前用户是否已有结果。
const (
	ContactActionChatReady   = "chat_ready"
	ContactActionRequestSent = "request_sent"
)

// ContactMatchedPartner ensures a revealed matched result can open a chat with its partner.
// If users are not friends yet, it sends (or reuses) a friend request and asks caller to wait.
func (s *Service) ContactMatchedPartner(ctx context.Context, userID, matchID int64) (ContactResult, error) {
	if err := s.ensureReady(); err != nil {
		return ContactResult{}, err
	}
	if matchID <= 0 {
		return ContactResult{}, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "invalid match id")
	}

	now := s.currentTime()
	var result ContactResult
	err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		record, err := s.store.GetUserMatchResultByID(ctx, tx, userID, matchID)
		if err != nil {
			if err == store.ErrNotFound {
				return httpapi.NewError(http.StatusNotFound, httpapi.CodeNoCurrentMatch, "match result not found")
			}
			return httpapi.InternalError(err)
		}
		if !record.CycleRevealAt.IsZero() && record.CycleRevealAt.After(now) {
			return httpapi.NewError(http.StatusNotFound, httpapi.CodeNoCurrentMatch, "match result not found")
		}
		if record.ResultStatus != ResultStatusMatched {
			return httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, "match result is not matched")
		}
		if !record.PartnerUserID.Valid || record.PartnerUserID.Int64 == userID {
			return httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, "matched partner is unavailable")
		}

		partnerID := record.PartnerUserID.Int64
		if _, err := s.store.GetUserByID(ctx, tx, partnerID); err != nil {
			if err == store.ErrNotFound {
				return httpapi.NewError(http.StatusNotFound, httpapi.CodeUserNotFound, "matched partner not found")
			}
			return httpapi.InternalError(err)
		}

		existingFriendship, err := s.store.GetFriendshipBetween(ctx, tx, userID, partnerID)
		if err != nil && err != store.ErrNotFound {
			return httpapi.InternalError(err)
		}
		if err == nil && existingFriendship.Status == "blocked" {
			return httpapi.NewError(http.StatusForbidden, httpapi.CodeForbidden, "friendship is blocked")
		}

		if err == nil {
			switch existingFriendship.Status {
			case "accepted":
				conversation, err := s.store.CreateConversation(ctx, tx, userID, partnerID, now)
				if err != nil {
					return httpapi.InternalError(err)
				}
				result = ContactResult{
					Action:         ContactActionChatReady,
					FriendshipID:   existingFriendship.ID,
					ConversationID: conversation.ID,
				}
				return nil
			case "pending":
				message := "好友申请已发送，请等待对方通过后进入聊天。"
				if existingFriendship.RequesterID != userID {
					message = "对方已向你发来好友申请，请先在好友页面通过后进入聊天。"
				}
				result = ContactResult{
					Action:       ContactActionRequestSent,
					FriendshipID: existingFriendship.ID,
					Message:      message,
				}
				return nil
			case "blocked":
				return httpapi.NewError(http.StatusForbidden, httpapi.CodeForbidden, "friendship is blocked")
			default:
				// Reset stale friendship statuses (for example rejected) and create a fresh request.
				if err := s.store.DeleteFriendship(ctx, tx, existingFriendship.ID); err != nil {
					return httpapi.InternalError(err)
				}
			}
		}

		friendship, err := s.store.CreateFriendship(ctx, tx, userID, partnerID, now)
		if err != nil {
			return httpapi.InternalError(err)
		}
		result = ContactResult{
			Action:       ContactActionRequestSent,
			FriendshipID: friendship.ID,
			Message:      "好友申请已发送，请等待对方通过后进入聊天。",
		}
		return nil
	})
	if err != nil {
		return ContactResult{}, err
	}

	return result, nil
}

// GetCycleStatus returns the latest match cycle status for the current user.
func (s *Service) GetCycleStatus(ctx context.Context, userID int64) (CycleStatusResult, error) {
	if err := s.ensureReady(); err != nil {
		return CycleStatusResult{}, err
	}

	cycle, err := s.store.GetLatestMatchCycle(ctx, nil)
	if err != nil {
		if err == store.ErrNotFound {
			return CycleStatusResult{HasCurrentCycle: false}, nil
		}
		return CycleStatusResult{}, httpapi.InternalError(err)
	}
	now := s.currentTime()
	cycle, err = s.resetNextRoundParticipationIfNeeded(ctx, cycle, now)
	if err != nil {
		return CycleStatusResult{}, err
	}

	result := CycleStatusResult{
		HasCurrentCycle:  true,
		CycleID:          cycle.ID,
		CycleKey:         cycle.CycleKey,
		CycleTitle:       cycle.Title,
		RevealStatus:     cycle.RevealStatus,
		RevealAt:         cycle.RevealAt.Format(time.RFC3339),
		ParticipantCount: cycle.ParticipantCount,
		MatchedUserCount: cycle.MatchedUserCount,
	}
	if cycle.ParticipantCount > 0 {
		result.SuccessRatePercent = cycle.MatchedUserCount * 100 / cycle.ParticipantCount
	}

	record, err := s.store.GetUserMatchResultByCycle(ctx, nil, userID, cycle.ID)
	if err == nil {
		result.HasUserResult = true
		result.UserResultStatus = record.ResultStatus
		result.CanViewUserResult = !cycle.RevealAt.After(now)
		if !record.CycleRevealAt.IsZero() {
			result.UserRevealAt = record.CycleRevealAt.Format(time.RFC3339)
		}
		return result, nil
	}
	if err == store.ErrNotFound {
		return result, nil
	}
	return CycleStatusResult{}, httpapi.InternalError(err)
}

func (s *Service) ensureReady() error {
	if s == nil || s.store == nil {
		return httpapi.FeatureNotReady("match")
	}
	return nil
}

func (s *Service) currentTime() time.Time {
	if s != nil && s.nowFunc != nil {
		return s.nowFunc()
	}
	return time.Now()
}

func (s *Service) resetNextRoundParticipationIfNeeded(ctx context.Context, cycle store.MatchCycle, now time.Time) (store.MatchCycle, error) {
	if cycle.ID == 0 || strings.ToLower(strings.TrimSpace(cycle.RevealStatus)) != "revealed" || cycle.RevealAt.After(now) {
		return cycle, nil
	}

	err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		key := strings.TrimSpace(cycle.QuestionnaireKey)
		version := strings.TrimSpace(cycle.QuestionnaireVersion)
		if key != "" && version != "" {
			if _, err := s.store.ResetQuestionnaireMatchParticipationByVersion(ctx, tx, key, version, now); err != nil {
				return httpapi.InternalError(err)
			}
		}
		if err := s.store.UpdateMatchCycleRevealStatus(ctx, tx, cycle.ID, "archived", now); err != nil {
			return httpapi.InternalError(err)
		}
		return nil
	})
	if err != nil {
		return store.MatchCycle{}, err
	}

	cycle.RevealStatus = "archived"
	return cycle, nil
}

func (s *Service) prewarmCurrentResultCache(ctx context.Context, cycle store.MatchCycle) {
	if s.resultCache == nil || cycle.ID == 0 {
		return
	}
	results, err := s.store.ListMatchResultsByCycle(ctx, nil, cycle.ID)
	if err != nil {
		return
	}
	for _, result := range results {
		record, err := s.store.GetUserMatchResultByCycle(ctx, nil, result.UserID, cycle.ID)
		if err != nil {
			continue
		}
		view, err := buildResultView(record)
		if err != nil {
			continue
		}
		s.storeCurrentResultInCache(ctx, cycle.ID, result.UserID, view)
	}
}

func (s *Service) loadCurrentResultFromCache(ctx context.Context, cycleID, userID int64) (ResultView, bool) {
	if s.resultCache == nil {
		return ResultView{}, false
	}
	view, err := s.resultCache.LoadCurrent(ctx, cycleID, userID)
	if err != nil {
		return ResultView{}, false
	}
	return view, true
}

func (s *Service) storeCurrentResultInCache(ctx context.Context, cycleID, userID int64, view ResultView) {
	if s.resultCache == nil {
		return
	}
	_ = s.resultCache.StoreCurrent(ctx, cycleID, userID, view, s.cacheTTLForView(view))
}

func (s *Service) deleteCurrentResultFromCache(ctx context.Context, cycleID, userID int64) {
	if s.resultCache == nil {
		return
	}
	_ = s.resultCache.DeleteCurrent(ctx, cycleID, userID)
}

func (s *Service) cacheTTLForView(view ResultView) time.Duration {
	ttl := s.resultCacheTTL
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	if strings.TrimSpace(view.RevealAt) == "" {
		return ttl
	}
	revealAt, err := time.Parse(time.RFC3339, view.RevealAt)
	if err != nil {
		return ttl
	}
	candidate := time.Until(revealAt) + ttl
	if candidate > ttl {
		return candidate
	}
	return ttl
}

func (s *Service) prepareResultParams(ctx context.Context, tx *sql.Tx, cycleID int64, input RecordUserResultInput, now time.Time) (store.UpsertMatchResultParams, error) {
	status := normalizeResultStatus(input.ResultStatus)
	if status == "" {
		return store.UpsertMatchResultParams{}, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
	}

	if _, err := s.store.GetUserByID(ctx, tx, input.UserID); err != nil {
		if err == store.ErrNotFound {
			return store.UpsertMatchResultParams{}, httpapi.NewError(http.StatusNotFound, httpapi.CodeUserNotFound, "用户不存在")
		}
		return store.UpsertMatchResultParams{}, httpapi.InternalError(err)
	}

	params := store.UpsertMatchResultParams{
		CycleID:         cycleID,
		UserID:          input.UserID,
		ResultStatus:    status,
		MatchReason:     strings.TrimSpace(input.MatchReason),
		MatchReasonJSON: append(json.RawMessage(nil), input.MatchReasonJSON...),
		ResultMessage:   strings.TrimSpace(input.Message),
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	switch status {
	case ResultStatusMatched:
		if input.PartnerUserID == nil || *input.PartnerUserID == input.UserID {
			return store.UpsertMatchResultParams{}, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
		}
		partner, err := s.store.GetUserByID(ctx, tx, *input.PartnerUserID)
		if err != nil {
			if err == store.ErrNotFound {
				return store.UpsertMatchResultParams{}, httpapi.NewError(http.StatusNotFound, httpapi.CodeUserNotFound, "用户不存在")
			}
			return store.UpsertMatchResultParams{}, httpapi.InternalError(err)
		}
		partnerContacts, err := s.store.ListUserContactMethods(ctx, tx, partner.ID)
		if err != nil {
			return store.UpsertMatchResultParams{}, httpapi.InternalError(err)
		}
		contactsJSON, err := json.Marshal(toPartnerContactSnapshots(partner, partnerContacts))
		if err != nil {
			return store.UpsertMatchResultParams{}, httpapi.InternalError(err)
		}

		params.PartnerUserID = input.PartnerUserID
		params.MatchScore = input.MatchScore
		params.PartnerNickname = partner.Nickname
		params.PartnerDepartment = partner.Department
		params.PartnerGrade = partner.Grade
		params.PartnerContactsJSON = contactsJSON
		if params.ResultMessage == "" {
			params.ResultMessage = "匹配成功"
		}
	case ResultStatusFailed:
		params.PartnerContactsJSON = json.RawMessage(`[]`)
		params.MatchScore = nil
		params.MatchReason = ""
		params.MatchReasonJSON = nil
		params.RevealedAt = nil
		if params.ResultMessage == "" {
			params.ResultMessage = "本轮未匹配成功"
		}
	default:
		return store.UpsertMatchResultParams{}, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
	}

	return params, nil
}

func buildResultView(record store.MatchResultRecord) (ResultView, error) {
	view := ResultView{
		MatchID:      record.ID,
		CycleID:      record.CycleID,
		CycleKey:     record.CycleKey,
		CycleTitle:   record.CycleTitle,
		RevealStatus: record.RevealStatus,
		ResultStatus: record.ResultStatus,
		Message:      record.ResultMessage,
		Reason:       record.MatchReason,
		ReasonJSON:   append(json.RawMessage(nil), record.MatchReasonJSON...),
	}
	if !record.CycleRevealAt.IsZero() {
		view.RevealAt = record.CycleRevealAt.Format(time.RFC3339)
	}
	if record.MatchScore.Valid {
		score := record.MatchScore.Float64
		view.Score = &score
	}
	if record.ResultStatus != ResultStatusMatched {
		view.Reason = ""
		view.ReasonJSON = nil
		return view, nil
	}

	contacts := make([]partnerContactSnapshot, 0)
	if len(record.PartnerContactsJSON) != 0 {
		if err := json.Unmarshal(record.PartnerContactsJSON, &contacts); err != nil {
			return ResultView{}, httpapi.InternalError(fmt.Errorf("decode partner contacts: %w", err))
		}
	}

	partner := &PartnerView{
		SchoolName: schoolNameForResult(record),
		Nickname:   record.PartnerNickname,
		Department: record.PartnerDepartment,
		Grade:      record.PartnerGrade,
		Contacts:   make([]ContactView, 0, len(contacts)),
	}
	if record.PartnerUserID.Valid {
		partner.UserID = record.PartnerUserID.Int64
	}
	for _, contact := range contacts {
		partner.Contacts = append(partner.Contacts, ContactView{
			Type:  contact.Type,
			Label: meta.ContactTypeLabel(contact.Type),
			Value: contact.Value,
		})
	}
	view.Partner = partner
	return view, nil
}

func schoolNameForResult(record store.MatchResultRecord) string {
	if record.PartnerSchoolKey != "" {
		if name, ok := schoolNameForKey(record.PartnerSchoolKey); ok {
			return name
		}
	}
	if record.PartnerEmail != "" {
		if name, ok := school.DisplayNameByEmail(record.PartnerEmail); ok {
			return name
		}
	}
	return ""
}

func schoolNameForKey(key string) (string, bool) {
	switch strings.TrimSpace(strings.ToLower(key)) {
	case "thu":
		return school.DisplayName(school.Tsinghua), true
	case "pku":
		return school.DisplayName(school.Peking), true
	case "ruc":
		return school.DisplayName(school.Renmin), true
	case "cuc":
		return school.DisplayName(school.Communication), true
	case "bfsu":
		return school.DisplayName(school.Beiwai), true
	case "bnu":
		return school.DisplayName(school.BeijingNormal), true
	case "zuel":
		return school.DisplayName(school.ZUEL), true
	case "nus":
		return school.DisplayName(school.NUS), true
	case "ntu":
		return school.DisplayName(school.NTU), true
	case "sjtu", "sjdu":
		return school.DisplayName(school.ShanghaiJiao), true
	case "fdu":
		return school.DisplayName(school.Fudan), true
	default:
		return "", false
	}
}

func normalizeRevealStatus(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "", "revealed":
		return "revealed"
	case "draft", "pending", "archived":
		return value
	default:
		return "revealed"
	}
}

func normalizeResultStatus(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case ResultStatusMatched, ResultStatusFailed:
		return value
	default:
		return ""
	}
}

func toPartnerContactSnapshots(user store.User, contacts []store.ContactMethod) []partnerContactSnapshot {
	if len(contacts) == 0 {
		email := strings.TrimSpace(user.Email)
		if email == "" {
			return []partnerContactSnapshot{}
		}
		return []partnerContactSnapshot{{
			Type:  "email",
			Value: email,
		}}
	}

	latest := contacts[0]
	for _, contact := range contacts[1:] {
		if contact.UpdatedAt.After(latest.UpdatedAt) {
			latest = contact
			continue
		}
		if contact.UpdatedAt.Equal(latest.UpdatedAt) && contact.ContactType < latest.ContactType {
			latest = contact
		}
	}

	return []partnerContactSnapshot{{
		Type:  latest.ContactType,
		Value: latest.ContactValue,
	}}
}

func buildSimpleVersionResults(users []store.User, questionnaireVersion string, rng *rand.Rand) []RecordUserResultInput {
	participants := append([]store.User(nil), users...)
	rng.Shuffle(len(participants), func(i, j int) {
		participants[i], participants[j] = participants[j], participants[i]
	})

	results := make([]RecordUserResultInput, 0, len(participants))
	matchReason := fmt.Sprintf("本轮采用极简随机匹配方案，在完成版本 %s 问卷的用户中随机为你配对。", questionnaireVersion)
	for i := 0; i+1 < len(participants); i += 2 {
		left := participants[i]
		right := participants[i+1]
		score := 60 + float64(rng.Intn(31))

		results = append(results,
			RecordUserResultInput{
				UserID:        left.ID,
				ResultStatus:  ResultStatusMatched,
				PartnerUserID: int64Ptr(right.ID),
				MatchScore:    float64Ptr(score),
				MatchReason:   matchReason,
				Message:       "匹配成功",
			},
			RecordUserResultInput{
				UserID:        right.ID,
				ResultStatus:  ResultStatusMatched,
				PartnerUserID: int64Ptr(left.ID),
				MatchScore:    float64Ptr(score),
				MatchReason:   matchReason,
				Message:       "匹配成功",
			},
		)
	}

	if len(participants)%2 == 1 {
		leftover := participants[len(participants)-1]
		results = append(results, RecordUserResultInput{
			UserID:       leftover.ID,
			ResultStatus: ResultStatusFailed,
			Message:      "本轮未匹配成功，剩余人数不足形成新的两两配对。",
		})
	}

	return results
}

func int64Ptr(value int64) *int64 {
	return &value
}

func float64Ptr(value float64) *float64 {
	return &value
}
