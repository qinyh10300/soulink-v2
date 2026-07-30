package questionnaire

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/store"
)

type ResultRevealCoordinator interface {
	AssignOnFirstCompletion(ctx context.Context, tx *sql.Tx, userID int64, questionnaireKey, questionnaireVersion string, now time.Time) (*time.Time, error)
	ResolveUserRevealAt(ctx context.Context, tx *sql.Tx, userID int64, questionnaireKey, questionnaireVersion string, now time.Time) (*time.Time, error)
}

type ServiceOption func(*Service)

type Service struct {
	store                   *store.Store
	catalog                 *Catalog
	resultRevealCoordinator ResultRevealCoordinator
	nowFunc                 func() time.Time
	soulAgentEnabled        bool
}

type Answer struct {
	QuestionID string `json:"questionId"`
	Value      any    `json:"value"`
	Importance bool   `json:"importance,omitempty"`
}

type CurrentResult struct {
	QuestionnaireID           int           `json:"questionnaireId"`
	QuestionnaireKey          string        `json:"questionnaireKey"`
	Title                     string        `json:"title"`
	Version                   string        `json:"version"`
	CollectionStatus          string        `json:"collectionStatus"`
	AcceptingResponses        bool          `json:"acceptingResponses"`
	CollectionStartsAt        string        `json:"collectionStartsAt,omitempty"`
	CollectionEndsAt          string        `json:"collectionEndsAt,omitempty"`
	CollectionTimezone        string        `json:"collectionTimezone,omitempty"`
	ResultRevealAt            string        `json:"resultRevealAt,omitempty"`
	TotalCount                int           `json:"totalCount"`
	AnsweredCount             int           `json:"answeredCount"`
	ProgressPercent           int           `json:"progressPercent"`
	HasCompletedQuestionnaire bool          `json:"hasCompletedQuestionnaire"`
	ParticipatingInMatch      bool          `json:"participatingInMatch"`
	Sections                  []SectionMeta `json:"sections"`
	Questions                 []Question    `json:"questions"`
	Answers                   []Answer      `json:"answers"`
}

type ProgressResult struct {
	QuestionnaireID           int    `json:"questionnaireId"`
	QuestionnaireKey          string `json:"questionnaireKey"`
	Version                   string `json:"version"`
	CollectionStatus          string `json:"collectionStatus"`
	AcceptingResponses        bool   `json:"acceptingResponses"`
	CollectionStartsAt        string `json:"collectionStartsAt,omitempty"`
	CollectionEndsAt          string `json:"collectionEndsAt,omitempty"`
	CollectionTimezone        string `json:"collectionTimezone,omitempty"`
	ResultRevealAt            string `json:"resultRevealAt,omitempty"`
	TotalCount                int    `json:"totalCount"`
	AnsweredCount             int    `json:"answeredCount"`
	ProgressPercent           int    `json:"progressPercent"`
	HasCompletedQuestionnaire bool   `json:"hasCompletedQuestionnaire"`
	ParticipatingInMatch      bool   `json:"participatingInMatch"`
}

type SaveAnswerInput struct {
	QuestionID string
	Value      json.RawMessage
	Importance *bool
}

type SaveAnswersInput struct {
	Answers []SaveAnswerInput
	Action  string
}

type SaveAnswerResult struct {
	QuestionID                string `json:"questionId"`
	Value                     any    `json:"value"`
	Importance                bool   `json:"importance,omitempty"`
	ResultRevealAt            string `json:"resultRevealAt,omitempty"`
	AnsweredCount             int    `json:"answeredCount"`
	ProgressPercent           int    `json:"progressPercent"`
	HasCompletedQuestionnaire bool   `json:"hasCompletedQuestionnaire"`
	ParticipatingInMatch      bool   `json:"participatingInMatch"`
}

type SaveAnswersResult struct {
	ResultRevealAt            string `json:"resultRevealAt,omitempty"`
	AnsweredCount             int    `json:"answeredCount"`
	ProgressPercent           int    `json:"progressPercent"`
	HasCompletedQuestionnaire bool   `json:"hasCompletedQuestionnaire"`
	ParticipatingInMatch      bool   `json:"participatingInMatch"`
}

type UpdateParticipationResult struct {
	ResultRevealAt            string `json:"resultRevealAt,omitempty"`
	HasCompletedQuestionnaire bool   `json:"hasCompletedQuestionnaire"`
	ParticipatingInMatch      bool   `json:"participatingInMatch"`
}

type RangeAnswerValue struct {
	Min       *int `json:"min,omitempty"`
	Max       *int `json:"max,omitempty"`
	Unlimited bool `json:"unlimited,omitempty"`
}

type validatedAnswer struct {
	Question   Question
	ValueJSON  json.RawMessage
	Importance bool
}

type stateMetrics struct {
	AnsweredCount             int
	RequiredAnsweredCount     int
	ProgressPercent           int
	HasCompletedQuestionnaire bool
	ParticipatingInMatch      bool
	ResultRevealAt            *time.Time
}

type collectionInfo struct {
	Status             string
	AcceptingResponses bool
	StartsAt           string
	EndsAt             string
	Timezone           string
}

func WithResultRevealCoordinator(coordinator ResultRevealCoordinator) ServiceOption {
	return func(service *Service) {
		service.resultRevealCoordinator = coordinator
	}
}

func WithNowFunc(nowFunc func() time.Time) ServiceOption {
	return func(service *Service) {
		if nowFunc != nil {
			service.nowFunc = nowFunc
		}
	}
}

func WithSoulAgentEnabled(enabled bool) ServiceOption {
	return func(service *Service) {
		service.soulAgentEnabled = enabled
	}
}

// NewService 创建问卷服务实例。
func NewService(dataStore *store.Store, catalog *Catalog, options ...ServiceOption) *Service {
	service := &Service{
		store:            dataStore,
		catalog:          catalog,
		nowFunc:          time.Now,
		soulAgentEnabled: true,
	}
	for _, option := range options {
		if option != nil {
			option(service)
		}
	}
	return service
}

// GetCurrent 返回当前问卷定义、进度和已保存答案。
func (s *Service) GetCurrent(ctx context.Context, userID int64) (CurrentResult, error) {
	if err := s.ensureReady(); err != nil {
		return CurrentResult{}, err
	}

	state, err := s.getQuestionnaireState(ctx, userID)
	if err != nil {
		return CurrentResult{}, err
	}

	answers, err := s.store.ListQuestionnaireAnswers(ctx, nil, userID, s.catalog.Key, s.catalog.Version)
	if err != nil {
		return CurrentResult{}, httpapi.InternalError(err)
	}

	answeredCount, requiredAnsweredCount, answerList, err := s.buildAnswerSnapshot(answers)
	if err != nil {
		return CurrentResult{}, httpapi.InternalError(err)
	}

	collection := s.buildCollectionInfo(s.currentTime())
	participating := questionnaireParticipation(state)
	return CurrentResult{
		QuestionnaireID:           s.catalog.ID,
		QuestionnaireKey:          s.catalog.Key,
		Title:                     s.catalog.Title,
		Version:                   s.catalog.Version,
		CollectionStatus:          collection.Status,
		AcceptingResponses:        collection.AcceptingResponses,
		CollectionStartsAt:        collection.StartsAt,
		CollectionEndsAt:          collection.EndsAt,
		CollectionTimezone:        collection.Timezone,
		ResultRevealAt:            formatOptionalTime(questionnaireResultRevealAt(state, s.currentTime(), participating)),
		TotalCount:                s.catalog.TotalCount,
		AnsweredCount:             answeredCount,
		ProgressPercent:           progressPercent(answeredCount, s.catalog.TotalCount),
		HasCompletedQuestionnaire: requiredAnsweredCount == s.catalog.RequiredCount,
		ParticipatingInMatch:      participating,
		Sections:                  append([]SectionMeta(nil), s.catalog.Sections...),
		Questions:                 append([]Question(nil), s.catalog.Questions...),
		Answers:                   answerList,
	}, nil
}

// GetProgress 返回当前问卷的轻量进度信息。
func (s *Service) GetProgress(ctx context.Context, userID int64) (ProgressResult, error) {
	if err := s.ensureReady(); err != nil {
		return ProgressResult{}, err
	}

	state, err := s.getQuestionnaireState(ctx, userID)
	if err != nil {
		return ProgressResult{}, err
	}

	answers, err := s.store.ListQuestionnaireAnswers(ctx, nil, userID, s.catalog.Key, s.catalog.Version)
	if err != nil {
		return ProgressResult{}, httpapi.InternalError(err)
	}

	answeredCount, requiredAnsweredCount, _, err := s.buildAnswerSnapshot(answers)
	if err != nil {
		return ProgressResult{}, httpapi.InternalError(err)
	}

	collection := s.buildCollectionInfo(s.currentTime())
	participating := questionnaireParticipation(state)
	return ProgressResult{
		QuestionnaireID:           s.catalog.ID,
		QuestionnaireKey:          s.catalog.Key,
		Version:                   s.catalog.Version,
		CollectionStatus:          collection.Status,
		AcceptingResponses:        collection.AcceptingResponses,
		CollectionStartsAt:        collection.StartsAt,
		CollectionEndsAt:          collection.EndsAt,
		CollectionTimezone:        collection.Timezone,
		ResultRevealAt:            formatOptionalTime(questionnaireResultRevealAt(state, s.currentTime(), participating)),
		TotalCount:                s.catalog.TotalCount,
		AnsweredCount:             answeredCount,
		ProgressPercent:           progressPercent(answeredCount, s.catalog.TotalCount),
		HasCompletedQuestionnaire: requiredAnsweredCount == s.catalog.RequiredCount,
		ParticipatingInMatch:      participating,
	}, nil
}

// SaveAnswer 保存单题答案，并刷新问卷状态。
func (s *Service) SaveAnswer(ctx context.Context, userID int64, input SaveAnswerInput) (SaveAnswerResult, error) {
	if err := s.ensureReady(); err != nil {
		return SaveAnswerResult{}, err
	}

	now := s.currentTime()
	if err := s.ensureCollectionOpen(now); err != nil {
		return SaveAnswerResult{}, err
	}

	question, ok := s.catalog.QuestionByID(input.QuestionID)
	if !ok {
		return SaveAnswerResult{}, httpapi.NewError(http.StatusNotFound, httpapi.CodeQuestionnaireNotFound, "问卷不存在")
	}

	valueJSON, value, importance, err := validateAndNormalizeAnswer(question, input.Value, input.Importance)
	if err != nil {
		return SaveAnswerResult{}, err
	}

	var metrics stateMetrics
	if err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		if _, err := s.store.UpsertQuestionnaireAnswer(ctx, tx, store.UpsertQuestionnaireAnswerParams{
			UserID:               userID,
			QuestionnaireKey:     s.catalog.Key,
			QuestionnaireVersion: s.catalog.Version,
			QuestionID:           question.ID,
			ValueJSON:            valueJSON,
			IsImportant:          importance,
			AnsweredAt:           now,
			CreatedAt:            now,
			UpdatedAt:            now,
		}); err != nil {
			return httpapi.InternalError(err)
		}

		currentAnswers, err := s.store.ListQuestionnaireAnswers(ctx, tx, userID, s.catalog.Key, s.catalog.Version)
		if err != nil {
			return httpapi.InternalError(err)
		}

		computed, err := s.persistState(ctx, tx, userID, currentAnswers, "save", now)
		if err != nil {
			return err
		}
		metrics = computed
		return nil
	}); err != nil {
		return SaveAnswerResult{}, err
	}

	s.ensurePortraitGeneration(ctx, userID, metrics.AnsweredCount, metrics.HasCompletedQuestionnaire, now, false)

	return SaveAnswerResult{
		QuestionID:                question.ID,
		Value:                     value,
		Importance:                importance,
		ResultRevealAt:            formatOptionalTime(questionnaireResultRevealAtFromMetrics(metrics, now)),
		AnsweredCount:             metrics.AnsweredCount,
		ProgressPercent:           metrics.ProgressPercent,
		HasCompletedQuestionnaire: metrics.HasCompletedQuestionnaire,
		ParticipatingInMatch:      metrics.ParticipatingInMatch,
	}, nil
}

// SaveAnswers 批量保存问卷答案，并刷新问卷状态。
func (s *Service) SaveAnswers(ctx context.Context, userID int64, input SaveAnswersInput) (SaveAnswersResult, error) {
	if err := s.ensureReady(); err != nil {
		return SaveAnswersResult{}, err
	}
	if len(input.Answers) == 0 {
		return SaveAnswersResult{}, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
	}

	now := s.currentTime()
	if err := s.ensureCollectionOpen(now); err != nil {
		return SaveAnswersResult{}, err
	}

	action := strings.TrimSpace(input.Action)
	if action == "" {
		action = "save"
	}
	switch action {
	case "save", "save_exit", "submit":
	default:
		return SaveAnswersResult{}, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
	}

	seen := make(map[string]struct{}, len(input.Answers))
	normalized := make([]validatedAnswer, 0, len(input.Answers))
	for _, answer := range input.Answers {
		if _, exists := seen[answer.QuestionID]; exists {
			return SaveAnswersResult{}, httpapi.NewError(http.StatusBadRequest, httpapi.CodeConflict, "数据冲突")
		}
		seen[answer.QuestionID] = struct{}{}

		question, ok := s.catalog.QuestionByID(answer.QuestionID)
		if !ok {
			return SaveAnswersResult{}, httpapi.NewError(http.StatusNotFound, httpapi.CodeQuestionnaireNotFound, "问卷不存在")
		}

		valueJSON, _, importance, err := validateAndNormalizeAnswer(question, answer.Value, answer.Importance)
		if err != nil {
			return SaveAnswersResult{}, err
		}
		normalized = append(normalized, validatedAnswer{
			Question:   question,
			ValueJSON:  valueJSON,
			Importance: importance,
		})
	}

	var metrics stateMetrics
	if err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		for _, answer := range normalized {
			if _, err := s.store.UpsertQuestionnaireAnswer(ctx, tx, store.UpsertQuestionnaireAnswerParams{
				UserID:               userID,
				QuestionnaireKey:     s.catalog.Key,
				QuestionnaireVersion: s.catalog.Version,
				QuestionID:           answer.Question.ID,
				ValueJSON:            answer.ValueJSON,
				IsImportant:          answer.Importance,
				AnsweredAt:           now,
				CreatedAt:            now,
				UpdatedAt:            now,
			}); err != nil {
				return httpapi.InternalError(err)
			}
		}

		currentAnswers, err := s.store.ListQuestionnaireAnswers(ctx, tx, userID, s.catalog.Key, s.catalog.Version)
		if err != nil {
			return httpapi.InternalError(err)
		}

		computed, err := s.persistState(ctx, tx, userID, currentAnswers, action, now)
		if err != nil {
			return err
		}
		metrics = computed
		return nil
	}); err != nil {
		return SaveAnswersResult{}, err
	}

	s.ensurePortraitGeneration(ctx, userID, metrics.AnsweredCount, metrics.HasCompletedQuestionnaire, now, false)
	return SaveAnswersResult{
		ResultRevealAt:            formatOptionalTime(questionnaireResultRevealAtFromMetrics(metrics, now)),
		AnsweredCount:             metrics.AnsweredCount,
		ProgressPercent:           metrics.ProgressPercent,
		HasCompletedQuestionnaire: metrics.HasCompletedQuestionnaire,
		ParticipatingInMatch:      metrics.ParticipatingInMatch,
	}, nil
}

func (s *Service) UpdateParticipation(ctx context.Context, userID int64, participating bool) (UpdateParticipationResult, error) {
	if err := s.ensureReady(); err != nil {
		return UpdateParticipationResult{}, err
	}

	now := s.currentTime()
	if err := s.ensureCollectionOpen(now); err != nil {
		return UpdateParticipationResult{}, err
	}
	state, err := s.store.GetQuestionnaireState(ctx, nil, userID, s.catalog.Key, s.catalog.Version)
	if err != nil {
		if err == store.ErrNotFound {
			return UpdateParticipationResult{}, httpapi.NewError(http.StatusForbidden, httpapi.CodeForbidden, "complete questionnaire before updating match participation")
		}
		return UpdateParticipationResult{}, httpapi.InternalError(err)
	}
	if !state.HasCompletedQuestionnaire {
		return UpdateParticipationResult{}, httpapi.NewError(http.StatusForbidden, httpapi.CodeForbidden, "complete questionnaire before updating match participation")
	}

	if state.ParticipatingInMatch != participating {
		if err := s.store.UpdateQuestionnaireMatchParticipation(ctx, nil, userID, s.catalog.Key, s.catalog.Version, participating, now); err != nil {
			return UpdateParticipationResult{}, httpapi.InternalError(err)
		}
		state.ParticipatingInMatch = participating
		state.UpdatedAt = now
	}

	return UpdateParticipationResult{
		ResultRevealAt:            formatOptionalTime(questionnaireResultRevealAt(&state, now, participating)),
		HasCompletedQuestionnaire: true,
		ParticipatingInMatch:      participating,
	}, nil
}

func (s *Service) ensureReady() error {
	if s == nil || s.store == nil || s.catalog == nil {
		return httpapi.FeatureNotReady("questionnaire")
	}
	return nil
}

func (s *Service) currentTime() time.Time {
	if s != nil && s.nowFunc != nil {
		return s.nowFunc()
	}
	return time.Now()
}

func (s *Service) buildCollectionInfo(now time.Time) collectionInfo {
	status, accepting := s.catalog.CollectionWindow(now)
	return collectionInfo{
		Status:             string(status),
		AcceptingResponses: accepting,
		StartsAt:           formatCollectionTime(s.catalog.CollectionStartsAt),
		EndsAt:             formatCollectionTime(s.catalog.CollectionEndsAt),
		Timezone:           s.catalog.CollectionTimezone,
	}
}

func (s *Service) ensureCollectionOpen(now time.Time) error {
	status, accepting := s.catalog.CollectionWindow(now)
	if accepting {
		return nil
	}

	switch status {
	case CollectionStatusScheduled:
		return httpapi.NewError(http.StatusForbidden, httpapi.CodeForbidden, "问卷暂未开放")
	case CollectionStatusClosed:
		return httpapi.NewError(http.StatusForbidden, httpapi.CodeForbidden, "问卷收集已结束")
	default:
		return httpapi.NewError(http.StatusForbidden, httpapi.CodeForbidden, "当前不在问卷开放时间内")
	}
}

func (s *Service) persistState(ctx context.Context, tx *sql.Tx, userID int64, answers []store.QuestionnaireAnswer, action string, now time.Time) (stateMetrics, error) {
	previousState, previousStateErr := s.store.GetQuestionnaireState(ctx, tx, userID, s.catalog.Key, s.catalog.Version)
	if previousStateErr != nil && previousStateErr != store.ErrNotFound {
		return stateMetrics{}, httpapi.InternalError(previousStateErr)
	}

	answeredCount, requiredAnsweredCount, _, err := s.buildAnswerSnapshot(answers)
	if err != nil {
		return stateMetrics{}, httpapi.InternalError(err)
	}

	var resultRevealAt *time.Time
	if previousStateErr == nil && previousState.ResultRevealAt.Valid {
		revealAt := previousState.ResultRevealAt.Time
		resultRevealAt = &revealAt
	}

	metrics := stateMetrics{
		AnsweredCount:             answeredCount,
		RequiredAnsweredCount:     requiredAnsweredCount,
		ProgressPercent:           progressPercent(answeredCount, s.catalog.TotalCount),
		HasCompletedQuestionnaire: requiredAnsweredCount == s.catalog.RequiredCount,
		ResultRevealAt:            resultRevealAt,
	}
	if previousStateErr == nil && previousState.HasCompletedQuestionnaire && previousState.ParticipatingInMatch {
		metrics.ParticipatingInMatch = true
	}
	if !metrics.HasCompletedQuestionnaire {
		metrics.ParticipatingInMatch = false
	}

	justCompleted := metrics.HasCompletedQuestionnaire && (previousStateErr == store.ErrNotFound || !previousState.HasCompletedQuestionnaire || resultRevealAt == nil)
	if justCompleted && s.resultRevealCoordinator != nil {
		assignedRevealAt, err := s.resultRevealCoordinator.AssignOnFirstCompletion(ctx, tx, userID, s.catalog.Key, s.catalog.Version, now)
		if err != nil {
			return stateMetrics{}, httpapi.InternalError(err)
		}
		metrics.ResultRevealAt = assignedRevealAt
	}

	if _, err := s.store.UpsertQuestionnaireState(ctx, tx, store.UpsertQuestionnaireStateParams{
		UserID:                    userID,
		QuestionnaireKey:          s.catalog.Key,
		QuestionnaireTitle:        s.catalog.Title,
		QuestionnaireVersion:      s.catalog.Version,
		TotalCount:                s.catalog.TotalCount,
		RequiredCount:             s.catalog.RequiredCount,
		AnsweredCount:             metrics.AnsweredCount,
		RequiredAnsweredCount:     metrics.RequiredAnsweredCount,
		ProgressPercent:           metrics.ProgressPercent,
		HasCompletedQuestionnaire: metrics.HasCompletedQuestionnaire,
		ParticipatingInMatch:      metrics.ParticipatingInMatch,
		LastAction:                action,
		LastSavedAt:               now,
		ResultRevealAt:            metrics.ResultRevealAt,
		CreatedAt:                 now,
		UpdatedAt:                 now,
	}); err != nil {
		return stateMetrics{}, httpapi.InternalError(err)
	}

	if err := s.store.SetUserQuestionnaireCompleted(ctx, tx, userID, metrics.HasCompletedQuestionnaire, now); err != nil {
		return stateMetrics{}, httpapi.InternalError(err)
	}

	return metrics, nil
}

func (s *Service) getQuestionnaireState(ctx context.Context, userID int64) (*store.QuestionnaireState, error) {
	state, err := s.store.GetQuestionnaireState(ctx, nil, userID, s.catalog.Key, s.catalog.Version)
	if err != nil {
		if err == store.ErrNotFound {
			return nil, nil
		}
		return nil, httpapi.InternalError(err)
	}
	return &state, nil
}

func (s *Service) buildAnswerSnapshot(records []store.QuestionnaireAnswer) (int, int, []Answer, error) {
	return buildAnswerSnapshotForCatalog(s.catalog, records)
}

func progressPercent(answeredCount, totalCount int) int {
	if totalCount <= 0 {
		return 0
	}
	return answeredCount * 100 / totalCount
}

func validateAndNormalizeAnswer(question Question, raw json.RawMessage, importance *bool) (json.RawMessage, any, bool, error) {
	if len(raw) == 0 {
		return nil, nil, false, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
	}

	important := false
	if importance != nil {
		important = *importance
	}
	if important && !question.SupportsImportance {
		return nil, nil, false, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
	}

	switch question.Type {
	case QuestionTypeSingleChoice:
		var value string
		if err := json.Unmarshal(raw, &value); err != nil {
			return nil, nil, false, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
		}
		value = strings.TrimSpace(value)
		if value == "" || !containsOption(question.Options, value) {
			return nil, nil, false, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
		}
		normalized, _ := json.Marshal(value)
		return normalized, value, important, nil
	case QuestionTypeMultipleChoice:
		var value []string
		if err := json.Unmarshal(raw, &value); err != nil {
			return nil, nil, false, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
		}
		normalizedValues := make([]string, 0, len(value))
		seen := make(map[string]struct{}, len(value))
		for _, item := range value {
			item = strings.TrimSpace(item)
			if item == "" || !containsOption(question.Options, item) {
				return nil, nil, false, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
			}
			if _, ok := seen[item]; ok {
				continue
			}
			seen[item] = struct{}{}
			normalizedValues = append(normalizedValues, item)
		}
		if len(normalizedValues) == 0 {
			return nil, nil, false, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
		}
		if question.MinSelections != nil && len(normalizedValues) < *question.MinSelections {
			return nil, nil, false, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
		}
		if question.MaxSelections != nil && len(normalizedValues) > *question.MaxSelections {
			return nil, nil, false, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
		}
		normalized, _ := json.Marshal(normalizedValues)
		return normalized, normalizedValues, important, nil
	case QuestionTypeNumber, QuestionTypeScale:
		var value int
		if err := json.Unmarshal(raw, &value); err != nil {
			return nil, nil, false, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
		}
		if question.Min != nil && value < *question.Min {
			return nil, nil, false, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
		}
		if question.Max != nil && value > *question.Max {
			return nil, nil, false, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
		}
		normalized, _ := json.Marshal(value)
		return normalized, value, important, nil
	case QuestionTypeRange:
		var value RangeAnswerValue
		if err := json.Unmarshal(raw, &value); err != nil {
			return nil, nil, false, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
		}
		if value.Unlimited {
			if !question.AllowUnlimited {
				return nil, nil, false, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
			}
			value.Min = nil
			value.Max = nil
		} else {
			if value.Min == nil || value.Max == nil {
				return nil, nil, false, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
			}
			if question.Min != nil && *value.Min < *question.Min {
				return nil, nil, false, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
			}
			if question.Max != nil && *value.Max > *question.Max {
				return nil, nil, false, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
			}
			if *value.Min > *value.Max {
				return nil, nil, false, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
			}
		}
		normalized, _ := json.Marshal(value)
		return normalized, value, important, nil
	case QuestionTypeText:
		var value string
		if err := json.Unmarshal(raw, &value); err != nil {
			return nil, nil, false, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
		}
		value = strings.TrimSpace(value)
		if question.Required && value == "" {
			return nil, nil, false, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
		}
		if question.MaxLength != nil && len([]rune(value)) > *question.MaxLength {
			return nil, nil, false, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
		}
		normalized, _ := json.Marshal(value)
		return normalized, value, important, nil
	default:
		return nil, nil, false, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
	}
}

func containsOption(options []Option, value string) bool {
	for _, option := range options {
		if option.Value == value {
			return true
		}
	}
	return false
}

func formatCollectionTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format(time.RFC3339)
}

func formatOptionalTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format(time.RFC3339)
}

func questionnaireParticipation(state *store.QuestionnaireState) bool {
	return state != nil && state.HasCompletedQuestionnaire && state.ParticipatingInMatch
}

func questionnaireResultRevealAt(state *store.QuestionnaireState, now time.Time, participating bool) *time.Time {
	if !participating || state == nil || !state.ResultRevealAt.Valid || !state.ResultRevealAt.Time.After(now) {
		return nil
	}
	revealAt := state.ResultRevealAt.Time
	return &revealAt
}

func questionnaireResultRevealAtFromMetrics(metrics stateMetrics, now time.Time) *time.Time {
	if !metrics.ParticipatingInMatch || metrics.ResultRevealAt == nil || !metrics.ResultRevealAt.After(now) {
		return nil
	}
	revealAt := *metrics.ResultRevealAt
	return &revealAt
}
