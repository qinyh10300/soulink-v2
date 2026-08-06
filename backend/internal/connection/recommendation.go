package connection

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"soulink-backend/internal/ai"
	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/store"
)

const (
	recommendationAlgorithmVersion = "soulink-reciprocal-v2"
	dailyRecommendationTTL         = 24 * time.Hour
)

var modelContactPattern = regexp.MustCompile(`(?i)(?:1[3-9][0-9]{9}|[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,})`)

type modelRecommendationInput struct {
	Source     modelIntentSummary   `json:"source"`
	Candidates []modelCandidateItem `json:"candidates"`
}

type modelIntentSummary struct {
	Category    string   `json:"category"`
	Goal        string   `json:"goal"`
	TimeWindow  string   `json:"timeWindow"`
	Location    string   `json:"location"`
	Constraints []string `json:"constraints"`
}

type modelCandidateItem struct {
	CandidateID   int64    `json:"candidateId"`
	Category      string   `json:"category"`
	Goal          string   `json:"goal"`
	TimeWindow    string   `json:"timeWindow"`
	Location      string   `json:"location"`
	SharedContext []string `json:"sharedContext"`
}

type modelRecommendationOutput struct {
	Items []struct {
		CandidateID int64   `json:"candidateId"`
		FitReason   string  `json:"fitReason"`
		Uncertainty string  `json:"uncertainty"`
		Confidence  float64 `json:"confidence"`
	} `json:"items"`
}

type modelConversationStartersOutput struct {
	Starters []string `json:"starters"`
}

func (s *Service) conversationStarterDrafts(
	ctx context.Context,
	intent store.ConnectionIntent,
) ([]string, InferenceView) {
	starters := startersForCategory(intent.Category, intent.Title)
	inference := InferenceView{
		Source:        "deterministic",
		Confidence:    0.74,
		PromptVersion: "deterministic-starters-v2",
		Editable:      true,
	}
	if s.generator == nil {
		return starters, inference
	}
	prompt, ok := ai.PromptFor(ai.TaskConversationStarters)
	if !ok {
		return starters, inference
	}
	contextItems := []string{categoryLabel(intent.Category)}
	if intent.TimeWindow != "" {
		contextItems = append(contextItems, intent.TimeWindow)
	}
	if intent.Location != "" {
		contextItems = append(contextItems, intent.Location)
	}
	var output modelConversationStartersOutput
	metadata, err := s.generator.Generate(ctx, prompt, map[string]any{
		"category":      intent.Category,
		"intentTitle":   modelSafeText(intent.Title, 120),
		"sharedContext": safeModelStrings(contextItems, 4, 80),
	}, &output)
	if err != nil || len(output.Starters) != 3 {
		return starters, inference
	}
	clean := safeModelStrings(output.Starters, 3, 160)
	if len(clean) != 3 {
		return starters, inference
	}
	for _, starter := range clean {
		if !validModelExplanation(starter, 160) {
			return starters, inference
		}
	}
	return clean, InferenceView{
		Source:        metadata.Source,
		Model:         metadata.Model,
		Confidence:    0.82,
		PromptVersion: metadata.PromptVersion,
		Editable:      true,
	}
}

func (s *Service) recommendationsForSource(
	ctx context.Context,
	userID int64,
	source store.ConnectionIntent,
	limit int,
	preferFresh bool,
) ([]RecommendationView, error) {
	candidates, err := s.store.ListRecommendationCandidates(ctx, nil, userID, source.Category, 60)
	if err != nil {
		return nil, httpapi.InternalError(err)
	}

	learningEnabled, after, err := s.recommendationLearningState(ctx, nil, userID)
	if err != nil {
		return nil, err
	}
	snapshot := emptyLearningSnapshot()
	if learningEnabled {
		snapshot, err = s.store.GetRecommendationLearningSnapshot(ctx, nil, userID, after, 12)
		if err != nil {
			return nil, httpapi.InternalError(err)
		}
	}
	recentCandidates := map[int64]struct{}{}
	if preferFresh {
		recentCandidates, err = s.store.ListRecentDailyCandidateIDs(ctx, nil, userID, s.now().Add(-7*24*time.Hour))
		if err != nil {
			return nil, httpapi.InternalError(err)
		}
	}

	recommendations := make([]RecommendationView, 0, len(candidates))
	for _, candidate := range candidates {
		if !recommendationCategoryEligible(source.Category, candidate.Category) {
			continue
		}
		score, reason, uncertainty, shared := scoreRecommendation(source, candidate)
		if learningEnabled {
			score, reason = applyLearningToRecommendation(score, reason, candidate, snapshot)
		}
		if preferFresh {
			if _, seen := recentCandidates[candidate.ID]; seen {
				score -= 10
			} else {
				score += 4
			}
		}
		score = clampInt(score, 0, 99)
		recommendations = append(recommendations, RecommendationView{
			Intent:        toIntentView(candidate, userID),
			Score:         score,
			FitReason:     reason,
			Uncertainty:   uncertainty,
			SharedContext: shared,
			Inference: InferenceView{
				Source:        "deterministic",
				Confidence:    deterministicRecommendationConfidence(score),
				PromptVersion: "deterministic-match-v2",
				Editable:      true,
			},
		})
	}
	sort.SliceStable(recommendations, func(i, j int) bool {
		if recommendations[i].Score == recommendations[j].Score {
			return recommendations[i].Intent.UpdatedAt.After(recommendations[j].Intent.UpdatedAt)
		}
		return recommendations[i].Score > recommendations[j].Score
	})
	if limit <= 0 || limit > 3 {
		limit = 3
	}
	if len(recommendations) > limit {
		recommendations = recommendations[:limit]
	}
	return s.enhanceRecommendationExplanations(ctx, source, recommendations), nil
}

func (s *Service) enhanceRecommendationExplanations(
	ctx context.Context,
	source store.ConnectionIntent,
	recommendations []RecommendationView,
) []RecommendationView {
	if s.generator == nil || len(recommendations) == 0 {
		return recommendations
	}
	prompt, ok := ai.PromptFor(ai.TaskMatchExplanation)
	if !ok {
		return recommendations
	}
	input := modelRecommendationInput{
		Source: modelIntentSummary{
			Category:    source.Category,
			Goal:        modelSafeText(source.Goal, 240),
			TimeWindow:  modelSafeText(source.TimeWindow, 120),
			Location:    modelSafeText(source.Location, 120),
			Constraints: safeModelStrings(source.Constraints, 6, 80),
		},
		Candidates: make([]modelCandidateItem, 0, len(recommendations)),
	}
	allowed := make(map[int64]int, len(recommendations))
	for index, recommendation := range recommendations {
		allowed[recommendation.Intent.ID] = index
		input.Candidates = append(input.Candidates, modelCandidateItem{
			CandidateID:   recommendation.Intent.ID,
			Category:      recommendation.Intent.Category,
			Goal:          modelSafeText(recommendation.Intent.Goal, 240),
			TimeWindow:    modelSafeText(recommendation.Intent.TimeWindow, 120),
			Location:      modelSafeText(recommendation.Intent.Location, 120),
			SharedContext: safeModelStrings(recommendation.SharedContext, 4, 80),
		})
	}
	var output modelRecommendationOutput
	metadata, err := s.generator.Generate(ctx, prompt, input, &output)
	if err != nil {
		return recommendations
	}
	seen := make(map[int64]struct{}, len(output.Items))
	for _, item := range output.Items {
		index, exists := allowed[item.CandidateID]
		if !exists {
			continue
		}
		if _, duplicate := seen[item.CandidateID]; duplicate {
			continue
		}
		if !validModelExplanation(item.FitReason, 800) ||
			!validModelExplanation(item.Uncertainty, 500) ||
			item.Confidence < 0 || item.Confidence > 1 {
			continue
		}
		seen[item.CandidateID] = struct{}{}
		recommendations[index].FitReason = strings.TrimSpace(item.FitReason)
		recommendations[index].Uncertainty = strings.TrimSpace(item.Uncertainty)
		recommendations[index].Inference = InferenceView{
			Source:        metadata.Source,
			Model:         metadata.Model,
			Confidence:    item.Confidence,
			PromptVersion: metadata.PromptVersion,
			Editable:      true,
		}
	}
	return recommendations
}

func (s *Service) DailyRecommendations(ctx context.Context, userID int64) (DailyRecommendationsView, error) {
	if err := s.ensureReady(); err != nil {
		return DailyRecommendationsView{}, err
	}
	intents, err := s.store.ListMyConnectionIntents(ctx, nil, userID, 50)
	if err != nil {
		return DailyRecommendationsView{}, httpapi.InternalError(err)
	}
	var source *store.ConnectionIntent
	for index := range intents {
		intent := &intents[index]
		if intent.Status == StatusActive &&
			(!intent.ExpiresAt.Valid || intent.ExpiresAt.Time.After(s.now())) {
			source = intent
			break
		}
	}
	if source == nil {
		return DailyRecommendationsView{
			Algorithm: recommendationAlgorithmVersion,
			Items:     []RecommendationView{},
		}, nil
	}

	if err := s.ensureRecommendationSourceEligible(ctx, userID, *source); err != nil {
		return DailyRecommendationsView{}, err
	}
	now := s.now()
	current, err := s.store.ListCurrentDailyRecommendations(
		ctx, nil, userID, source.ID, source.UpdatedAt, now,
	)
	if err != nil {
		return DailyRecommendationsView{}, httpapi.InternalError(err)
	}
	replaceCurrent := false
	if len(current) > 0 {
		view, complete, viewErr := s.dailyViewFromStored(ctx, userID, current)
		if viewErr != nil {
			return DailyRecommendationsView{}, viewErr
		}
		if complete {
			return view, nil
		}
		replaceCurrent = true
	}

	generated, err := s.recommendationsForSource(ctx, userID, *source, 3, true)
	if err != nil {
		return DailyRecommendationsView{}, err
	}
	if len(generated) == 0 {
		return DailyRecommendationsView{
			SourceIntentID: source.ID,
			Algorithm:      recommendationAlgorithmVersion,
			Items:          []RecommendationView{},
		}, nil
	}

	expiresAt := now.Add(dailyRecommendationTTL)
	var stored []store.DailyRecommendation
	if err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		if err := s.store.LockDailyRecommendations(ctx, tx, userID); err != nil {
			return err
		}
		existing, err := s.store.ListCurrentDailyRecommendations(
			ctx, tx, userID, source.ID, source.UpdatedAt, now,
		)
		if err != nil {
			return err
		}
		if len(existing) > 0 && !replaceCurrent {
			stored = existing
			return nil
		}
		if replaceCurrent {
			if err := s.store.ExpireActiveDailyRecommendations(ctx, tx, userID, source.ID, now); err != nil {
				return err
			}
		}
		for index, recommendation := range generated {
			if err := s.store.CreateDailyRecommendation(ctx, tx, store.CreateDailyRecommendationParams{
				UserID:                userID,
				SourceIntentID:        source.ID,
				CandidateIntentID:     recommendation.Intent.ID,
				Rank:                  index + 1,
				Score:                 recommendation.Score,
				FitReason:             recommendation.FitReason,
				Uncertainty:           recommendation.Uncertainty,
				SharedContext:         recommendation.SharedContext,
				AlgorithmVersion:      recommendationAlgorithmVersion,
				InferenceSource:       recommendation.Inference.Source,
				ModelName:             recommendation.Inference.Model,
				Confidence:            recommendation.Inference.Confidence,
				PromptVersion:         recommendation.Inference.PromptVersion,
				SourceIntentUpdatedAt: source.UpdatedAt,
				Now:                   now,
				ExpiresAt:             expiresAt,
			}); err != nil {
				return err
			}
		}
		stored, err = s.store.ListCurrentDailyRecommendations(
			ctx, tx, userID, source.ID, source.UpdatedAt, now,
		)
		return err
	}); err != nil {
		return DailyRecommendationsView{}, httpapi.InternalError(err)
	}
	view, _, err := s.dailyViewFromStored(ctx, userID, stored)
	return view, err
}

func (s *Service) GetRecommendationPreferences(ctx context.Context, userID int64) (RecommendationPreferencesView, error) {
	if err := s.ensureReady(); err != nil {
		return RecommendationPreferencesView{}, err
	}
	preferences, err := s.store.GetAIUserPreferences(ctx, nil, userID)
	if errors.Is(err, store.ErrNotFound) {
		preferences = store.AIUserPreferences{
			UserID:                        userID,
			RecommendationLearningEnabled: true,
			AssistantEnabled:              true,
		}
	} else if err != nil {
		return RecommendationPreferencesView{}, httpapi.InternalError(err)
	}
	return s.recommendationPreferencesView(ctx, userID, preferences)
}

func (s *Service) UpdateRecommendationPreferences(
	ctx context.Context,
	userID int64,
	input RecommendationPreferenceInput,
) (RecommendationPreferencesView, error) {
	if err := s.ensureReady(); err != nil {
		return RecommendationPreferencesView{}, err
	}
	preferences, err := s.store.SetRecommendationLearningEnabled(
		ctx, nil, userID, input.LearningEnabled, s.now(),
	)
	if err != nil {
		return RecommendationPreferencesView{}, httpapi.InternalError(err)
	}
	return s.recommendationPreferencesView(ctx, userID, preferences)
}

func (s *Service) ResetRecommendationPreferences(ctx context.Context, userID int64) (RecommendationPreferencesView, error) {
	if err := s.ensureReady(); err != nil {
		return RecommendationPreferencesView{}, err
	}
	now := s.now()
	if err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		return s.store.ResetRecommendationLearning(ctx, tx, userID, now)
	}); err != nil {
		return RecommendationPreferencesView{}, httpapi.InternalError(err)
	}
	return s.GetRecommendationPreferences(ctx, userID)
}

func (s *Service) recommendationPreferencesView(
	ctx context.Context,
	userID int64,
	preferences store.AIUserPreferences,
) (RecommendationPreferencesView, error) {
	after := time.Unix(0, 0)
	if preferences.LearningResetAt.Valid {
		after = preferences.LearningResetAt.Time
	}
	snapshot, err := s.store.GetRecommendationLearningSnapshot(ctx, nil, userID, after, 12)
	if err != nil {
		return RecommendationPreferencesView{}, httpapi.InternalError(err)
	}
	view := RecommendationPreferencesView{
		LearningEnabled: preferences.RecommendationLearningEnabled,
		Affinities:      make([]RecommendationAffinityView, 0, len(snapshot.Affinities)),
		RecentSignals:   make([]RecommendationSignalView, 0, len(snapshot.RecentSignals)),
	}
	if preferences.LearningResetAt.Valid {
		value := preferences.LearningResetAt.Time
		view.LearningResetAt = &value
	}
	if !preferences.UpdatedAt.IsZero() {
		value := preferences.UpdatedAt
		view.UpdatedAt = &value
	}
	for _, affinity := range snapshot.Affinities {
		view.Affinities = append(view.Affinities, RecommendationAffinityView{
			Category: affinity.Category,
			Weight:   affinity.Weight,
			Signals:  affinity.Count,
			Summary:  affinitySummary(affinity.Weight),
		})
	}
	for _, signal := range snapshot.RecentSignals {
		view.RecentSignals = append(view.RecentSignals, RecommendationSignalView{
			Type:       signal.SignalType,
			Category:   signal.Category,
			Impact:     signal.Weight,
			Source:     signal.Source,
			OccurredAt: signal.CreatedAt,
		})
	}
	return view, nil
}

func (s *Service) dailyViewFromStored(
	ctx context.Context,
	userID int64,
	records []store.DailyRecommendation,
) (DailyRecommendationsView, bool, error) {
	view := DailyRecommendationsView{
		Algorithm: recommendationAlgorithmVersion,
		Items:     []RecommendationView{},
	}
	if len(records) == 0 {
		return view, false, nil
	}
	view.SourceIntentID = records[0].SourceIntentID
	generatedAt := records[0].CreatedAt
	refreshAt := records[0].ExpiresAt
	view.GeneratedAt = &generatedAt
	view.RefreshAt = &refreshAt
	view.Algorithm = records[0].AlgorithmVersion
	for _, record := range records {
		intent, err := s.GetIntent(ctx, userID, record.CandidateIntentID)
		if err != nil {
			appErr := httpapi.AsAppError(err)
			if appErr.HTTPStatus == http.StatusForbidden || appErr.HTTPStatus == http.StatusNotFound {
				continue
			}
			return DailyRecommendationsView{}, false, err
		}
		view.Items = append(view.Items, RecommendationView{
			Intent:        intent,
			Score:         record.Score,
			FitReason:     record.FitReason,
			Uncertainty:   record.Uncertainty,
			SharedContext: nonNilStrings(record.SharedContext),
			Inference: InferenceView{
				Source:        record.InferenceSource,
				Model:         record.ModelName,
				Confidence:    record.Confidence,
				PromptVersion: record.PromptVersion,
				Editable:      true,
			},
		})
	}
	return view, len(view.Items) == len(records), nil
}

func (s *Service) ensureRecommendationSourceEligible(ctx context.Context, userID int64, source store.ConnectionIntent) error {
	profile, err := s.GetMatchingProfile(ctx, userID)
	if err != nil {
		return err
	}
	if profile.MatchingPaused || stringListContains(profile.HiddenCategories, source.Category) {
		return forbidden("当前匹配设置已暂停这类推荐")
	}
	if source.Category == CategoryRomantic && !profile.RomanticOptIn {
		return forbidden("请先在资料中主动开启认真关系匹配")
	}
	return nil
}

func (s *Service) recommendationLearningState(
	ctx context.Context,
	tx *sql.Tx,
	userID int64,
) (bool, time.Time, error) {
	preferences, err := s.store.GetAIUserPreferences(ctx, tx, userID)
	if errors.Is(err, store.ErrNotFound) {
		return true, time.Unix(0, 0), nil
	}
	if err != nil {
		return false, time.Time{}, httpapi.InternalError(err)
	}
	after := time.Unix(0, 0)
	if preferences.LearningResetAt.Valid {
		after = preferences.LearningResetAt.Time
	}
	return preferences.RecommendationLearningEnabled, after, nil
}

func (s *Service) recordRecommendationSignal(
	ctx context.Context,
	tx *sql.Tx,
	userID int64,
	intent store.ConnectionIntent,
	signalType string,
	weight int,
	now time.Time,
) error {
	enabled, _, err := s.recommendationLearningState(ctx, tx, userID)
	if err != nil {
		return err
	}
	if !enabled || intent.AuthorID == userID {
		return nil
	}
	if err := s.store.UpsertRecommendationInteractionSignal(ctx, tx, store.UpsertRecommendationInteractionSignalParams{
		UserID:      userID,
		IntentID:    intent.ID,
		OtherUserID: intent.AuthorID,
		Category:    intent.Category,
		SignalType:  signalType,
		Weight:      weight,
		Now:         now,
	}); err != nil {
		return httpapi.InternalError(err)
	}
	return nil
}

func applyLearningToRecommendation(
	score int,
	reason string,
	candidate store.ConnectionIntent,
	snapshot store.RecommendationLearningSnapshot,
) (int, string) {
	categoryWeight := clampInt(snapshot.CategoryWeights[candidate.Category], -8, 8)
	userWeight := clampInt(snapshot.UserWeights[candidate.AuthorID], -10, 10)
	score += categoryWeight + userWeight
	if categoryWeight >= 2 {
		reason = appendReason(reason, "你最近对这类连接有正向反馈")
	}
	if userWeight >= 2 {
		reason = appendReason(reason, "你曾对与这位同学的连接给过正向反馈")
	}
	return score, reason
}

func appendReason(reason, addition string) string {
	reason = strings.TrimSpace(strings.TrimSuffix(reason, "。"))
	if reason == "" {
		return addition + "。"
	}
	return reason + "；" + addition + "。"
}

func emptyLearningSnapshot() store.RecommendationLearningSnapshot {
	return store.RecommendationLearningSnapshot{
		CategoryWeights: map[string]int{},
		UserWeights:     map[int64]int{},
		Affinities:      []store.RecommendationAffinity{},
		RecentSignals:   []store.RecommendationRecentSignal{},
	}
}

func deterministicRecommendationConfidence(score int) float64 {
	confidence := 0.55 + float64(score)/250
	if confidence > 0.9 {
		confidence = 0.9
	}
	return confidence
}

func modelSafeText(value string, maxLength int) string {
	value = modelContactPattern.ReplaceAllString(strings.TrimSpace(value), "[已隐藏联系方式]")
	runes := []rune(value)
	if len(runes) > maxLength {
		value = string(runes[:maxLength])
	}
	return value
}

func safeModelStrings(values []string, maxItems, maxLength int) []string {
	result := make([]string, 0, min(len(values), maxItems))
	for _, value := range values {
		value = modelSafeText(value, maxLength)
		if value != "" {
			result = append(result, value)
		}
		if len(result) == maxItems {
			break
		}
	}
	return result
}

func validModelExplanation(value string, maxLength int) bool {
	value = strings.TrimSpace(value)
	return value != "" && utf8.RuneCountInString(value) <= maxLength &&
		!modelContactPattern.MatchString(value)
}

func affinitySummary(weight int) string {
	switch {
	case weight >= 4:
		return "近期反馈明显偏好"
	case weight >= 1:
		return "近期反馈略有偏好"
	case weight <= -4:
		return "近期反馈建议减少"
	case weight <= -1:
		return "近期反馈略微减少"
	default:
		return "近期反馈中性"
	}
}

func clampInt(value, minimum, maximum int) int {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}
