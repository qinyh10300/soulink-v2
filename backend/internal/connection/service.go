package connection

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"soulink-backend/internal/ai"
	"soulink-backend/internal/analytics"
	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/store"
)

const (
	maxIntentTitleLength       = 80
	maxIntentDescriptionLength = 1200
	maxIntentShortTextLength   = 200
	maxInterestMessageLength   = 500
	maxRoomMessageLength       = 1000
	connectionFeedSessionLimit = 12
)

type Service struct {
	store       *store.Store
	interpreter *Interpreter
	generator   ai.Generator
	recorder    analytics.EventRecorder
	nowFunc     func() time.Time
}

func (s *Service) WithEventRecorder(recorder analytics.EventRecorder) *Service {
	s.recorder = recorder
	return s
}

func NewService(dataStore *store.Store, interpreter *Interpreter) *Service {
	if interpreter == nil {
		interpreter = NewInterpreter()
	}
	return &Service{
		store:       dataStore,
		interpreter: interpreter,
		generator:   interpreter.generator,
		nowFunc:     time.Now,
	}
}

func (s *Service) Interpret(ctx context.Context, raw string) (IntentInterpretation, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || utf8.RuneCountInString(raw) > maxIntentDescriptionLength {
		return IntentInterpretation{}, invalidParams("请用一句话描述你想认识谁")
	}
	return s.interpreter.InterpretContext(ctx, raw), nil
}

func (s *Service) GetMatchingProfile(ctx context.Context, userID int64) (MatchingProfileView, error) {
	if err := s.ensureReady(); err != nil {
		return MatchingProfileView{}, err
	}

	profile, err := s.store.GetMatchingProfile(ctx, nil, userID)
	if errors.Is(err, store.ErrNotFound) {
		return MatchingProfileView{
			MatchingProfileInput: MatchingProfileInput{
				Skills:           []string{},
				Interests:        []string{},
				PreferredScenes:  []string{},
				HiddenCategories: []string{},
			},
		}, nil
	}
	if err != nil {
		return MatchingProfileView{}, httpapi.InternalError(err)
	}
	return toMatchingProfileView(profile), nil
}

func (s *Service) UpdateMatchingProfile(ctx context.Context, userID int64, input MatchingProfileInput) (MatchingProfileView, error) {
	if err := s.ensureReady(); err != nil {
		return MatchingProfileView{}, err
	}

	normalized, err := normalizeMatchingProfile(input)
	if err != nil {
		return MatchingProfileView{}, err
	}
	profile, err := s.store.UpsertMatchingProfile(ctx, nil, store.UpsertMatchingProfileParams{
		UserID:             userID,
		Bio:                normalized.Bio,
		Skills:             normalized.Skills,
		Interests:          normalized.Interests,
		Availability:       normalized.Availability,
		PreferredScenes:    normalized.PreferredScenes,
		CollaborationStyle: normalized.CollaborationStyle,
		Boundaries:         normalized.Boundaries,
		AIPortrait:         normalized.AIPortrait,
		AIPortraitApproved: normalized.AIPortraitApproved,
		RomanticOptIn:      normalized.RomanticOptIn,
		MatchingPaused:     normalized.MatchingPaused,
		HiddenCategories:   normalized.HiddenCategories,
		Now:                s.now(),
	})
	if err != nil {
		return MatchingProfileView{}, httpapi.InternalError(err)
	}
	return toMatchingProfileView(profile), nil
}

func (s *Service) CreateIntent(ctx context.Context, userID int64, input IntentInput) (IntentView, error) {
	if err := s.ensureReady(); err != nil {
		return IntentView{}, err
	}

	now := s.now()
	prepared, expiresAt, publishedAt, err := s.prepareIntent(ctx, userID, input, nil, now)
	if err != nil {
		return IntentView{}, err
	}

	intentID, err := s.store.CreateConnectionIntent(ctx, nil, store.CreateConnectionIntentParams{
		AuthorID:        userID,
		Category:        prepared.Category,
		Title:           prepared.Title,
		Description:     prepared.Description,
		Goal:            prepared.Goal,
		TimeWindow:      prepared.TimeWindow,
		Location:        prepared.Location,
		Visibility:      prepared.Visibility,
		Status:          prepared.Status,
		Constraints:     prepared.Constraints,
		DealBreakers:    prepared.DealBreakers,
		DesiredNextStep: prepared.DesiredNextStep,
		ProjectMemo:     toStoreProjectMemo(prepared.ProjectMemo),
		ExpiresAt:       expiresAt,
		PublishedAt:     publishedAt,
		Now:             now,
	})
	if err != nil {
		return IntentView{}, httpapi.InternalError(err)
	}
	return s.GetIntent(ctx, userID, intentID)
}

func (s *Service) UpdateIntent(ctx context.Context, userID, intentID int64, input IntentInput) (IntentView, error) {
	if err := s.ensureReady(); err != nil {
		return IntentView{}, err
	}

	current, err := s.store.GetConnectionIntent(ctx, nil, userID, intentID)
	if err != nil {
		return IntentView{}, mapStoreError(err, "找不到这个意图")
	}
	if current.AuthorID != userID {
		return IntentView{}, forbidden("只有发布者可以修改这个意图")
	}

	now := s.now()
	prepared, expiresAt, publishedAt, err := s.prepareIntent(ctx, userID, input, &current, now)
	if err != nil {
		return IntentView{}, err
	}
	err = s.store.UpdateConnectionIntent(ctx, nil, store.UpdateConnectionIntentParams{
		IntentID:        intentID,
		AuthorID:        userID,
		Category:        prepared.Category,
		Title:           prepared.Title,
		Description:     prepared.Description,
		Goal:            prepared.Goal,
		TimeWindow:      prepared.TimeWindow,
		Location:        prepared.Location,
		Visibility:      prepared.Visibility,
		Status:          prepared.Status,
		Constraints:     prepared.Constraints,
		DealBreakers:    prepared.DealBreakers,
		DesiredNextStep: prepared.DesiredNextStep,
		ProjectMemo:     toStoreProjectMemo(prepared.ProjectMemo),
		ExpiresAt:       expiresAt,
		PublishedAt:     publishedAt,
		Now:             now,
	})
	if err != nil {
		return IntentView{}, mapStoreError(err, "找不到这个意图")
	}
	return s.GetIntent(ctx, userID, intentID)
}

func (s *Service) ExpireIntent(ctx context.Context, userID, intentID int64) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	if err := s.store.UpdateConnectionIntentStatus(ctx, nil, intentID, userID, StatusExpired, s.now()); err != nil {
		return mapStoreError(err, "找不到这个意图")
	}
	return nil
}

func (s *Service) GetIntent(ctx context.Context, viewerID, intentID int64) (IntentView, error) {
	if err := s.ensureReady(); err != nil {
		return IntentView{}, err
	}
	intent, err := s.store.GetConnectionIntent(ctx, nil, viewerID, intentID)
	if err != nil {
		return IntentView{}, mapStoreError(err, "找不到这个意图")
	}
	if intent.AuthorID != viewerID {
		authorProfile, err := s.GetMatchingProfile(ctx, intent.AuthorID)
		if err != nil {
			return IntentView{}, err
		}
		if authorProfile.MatchingPaused ||
			(intent.Category == CategoryRomantic && !authorProfile.RomanticOptIn) {
			return IntentView{}, forbidden("这个意图当前不可见")
		}
		blocked, err := s.store.UsersBlocked(ctx, nil, viewerID, intent.AuthorID)
		if err != nil {
			return IntentView{}, httpapi.InternalError(err)
		}
		if blocked {
			return IntentView{}, forbidden("这个意图当前不可见")
		}
		reported, err := s.store.UsersHaveActiveSafetyReport(ctx, nil, viewerID, intent.AuthorID)
		if err != nil {
			return IntentView{}, httpapi.InternalError(err)
		}
		if reported {
			return IntentView{}, forbidden("这个意图当前不可见")
		}
		if intent.Status != StatusActive || intent.Visibility == VisibilityPrivate ||
			(intent.ExpiresAt.Valid && !intent.ExpiresAt.Time.After(s.now())) {
			return IntentView{}, forbidden("这个意图当前不可见")
		}
	}
	return toIntentView(intent, viewerID), nil
}

func (s *Service) GetContextProfile(
	ctx context.Context,
	viewerID,
	intentID,
	subjectID int64,
) (ContextProfileView, error) {
	if err := s.ensureReady(); err != nil {
		return ContextProfileView{}, err
	}

	intent, err := s.store.GetConnectionIntent(ctx, nil, viewerID, intentID)
	if err != nil {
		return ContextProfileView{}, mapStoreError(err, "找不到这个意图")
	}
	if intent.AuthorID != viewerID {
		if _, err := s.GetIntent(ctx, viewerID, intentID); err != nil {
			return ContextProfileView{}, err
		}
	}

	if subjectID == intent.AuthorID {
		if intent.Category == CategoryRomantic && viewerID != intent.AuthorID && intent.InterestState != "accepted" {
			return ContextProfileView{}, forbidden("双方同意前不会揭示身份资料")
		}
	} else {
		if viewerID != intent.AuthorID {
			return ContextProfileView{}, forbidden("只能查看当前连接对象的资料")
		}
		status, statusErr := s.store.GetIntentInterestStatus(ctx, nil, intentID, subjectID)
		if statusErr != nil || (status != "interested" && status != "accepted") {
			return ContextProfileView{}, forbidden("这个回应者的资料当前不可见")
		}
	}

	if viewerID != subjectID {
		blocked, blockErr := s.store.UsersBlocked(ctx, nil, viewerID, subjectID)
		if blockErr != nil {
			return ContextProfileView{}, httpapi.InternalError(blockErr)
		}
		reported, reportErr := s.store.UsersHaveActiveSafetyReport(ctx, nil, viewerID, subjectID)
		if reportErr != nil {
			return ContextProfileView{}, httpapi.InternalError(reportErr)
		}
		if blocked || reported {
			return ContextProfileView{}, forbidden("这个用户的资料当前不可见")
		}
	}

	user, err := s.store.GetUserByID(ctx, nil, subjectID)
	if err != nil {
		return ContextProfileView{}, mapStoreError(err, "找不到这个用户")
	}
	profile, err := s.store.GetMatchingProfile(ctx, nil, subjectID)
	if errors.Is(err, store.ErrNotFound) {
		profile = store.MatchingProfile{UserID: subjectID}
	} else if err != nil {
		return ContextProfileView{}, httpapi.InternalError(err)
	}

	return contextualProfileView(user, profile, intent), nil
}

func (s *Service) ListMyIntents(ctx context.Context, userID int64) ([]IntentView, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	intents, err := s.store.ListMyConnectionIntents(ctx, nil, userID, 50)
	if err != nil {
		return nil, httpapi.InternalError(err)
	}
	return mapIntentViews(intents, userID), nil
}

func (s *Service) ListFeed(ctx context.Context, userID int64, category string) ([]IntentView, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	category = strings.TrimSpace(category)
	if category != "" && !validCategory(category) {
		return nil, invalidParams("未知的意图类别")
	}
	intents, err := s.store.ListFeedConnectionIntents(ctx, nil, userID, category, connectionFeedSessionLimit)
	if err != nil {
		return nil, httpapi.InternalError(err)
	}
	return mapIntentViews(intents, userID), nil
}

func (s *Service) ListSavedIntents(ctx context.Context, userID int64, category string) ([]IntentView, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	category = strings.TrimSpace(category)
	if category != "" && !validCategory(category) {
		return nil, invalidParams("未知的意图类别")
	}
	intents, err := s.store.ListSavedConnectionIntents(ctx, nil, userID, category, 50)
	if err != nil {
		return nil, httpapi.InternalError(err)
	}
	return mapIntentViews(intents, userID), nil
}

func (s *Service) SetIntentSaved(ctx context.Context, userID, intentID int64, saved bool) (IntentView, error) {
	intent, err := s.GetIntent(ctx, userID, intentID)
	if err != nil {
		return IntentView{}, err
	}
	if intent.IsOwner {
		return IntentView{}, invalidParams("无需收藏自己的意图")
	}
	record, err := s.store.GetConnectionIntent(ctx, nil, userID, intentID)
	if err != nil {
		return IntentView{}, mapStoreError(err, "找不到这个意图")
	}
	now := s.now()
	if err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		if err := s.store.SetIntentSaved(ctx, tx, intentID, userID, saved, now); err != nil {
			return err
		}
		if saved {
			return s.recordRecommendationSignal(ctx, tx, userID, record, "saved", 1, now)
		}
		return s.store.RevokeRecommendationInteractionSignal(ctx, tx, userID, intentID, "saved", now)
	}); err != nil {
		return IntentView{}, httpapi.InternalError(err)
	}
	return s.GetIntent(ctx, userID, intentID)
}

func (s *Service) DismissIntent(ctx context.Context, userID, intentID int64) error {
	intent, err := s.GetIntent(ctx, userID, intentID)
	if err != nil {
		return err
	}
	if intent.IsOwner {
		return invalidParams("不能忽略自己的意图")
	}
	if intent.InterestState == "interested" || intent.InterestState == "accepted" {
		return invalidParams("已回应的意图不能忽略")
	}
	record, err := s.store.GetConnectionIntent(ctx, nil, userID, intentID)
	if err != nil {
		return mapStoreError(err, "找不到这个意图")
	}
	now := s.now()
	if err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		if err := s.store.DismissIntent(ctx, tx, intentID, userID, now); err != nil {
			return err
		}
		if err := s.store.SetIntentSaved(ctx, tx, intentID, userID, false, now); err != nil {
			return err
		}
		if err := s.store.RevokeRecommendationInteractionSignal(ctx, tx, userID, intentID, "saved", now); err != nil {
			return err
		}
		if err := s.recordRecommendationSignal(ctx, tx, userID, record, "dismissed", -2, now); err != nil {
			return err
		}
		return s.store.ExpireActiveDailyRecommendations(ctx, tx, userID, 0, now)
	}); err != nil {
		return httpapi.InternalError(err)
	}
	return nil
}

func (s *Service) SetIntentInterest(ctx context.Context, userID, intentID int64, message string) (IntentView, error) {
	intent, err := s.GetIntent(ctx, userID, intentID)
	if err != nil {
		return IntentView{}, err
	}
	if intent.IsOwner {
		return IntentView{}, invalidParams("不能回应自己的意图")
	}
	profile, err := s.GetMatchingProfile(ctx, userID)
	if err != nil {
		return IntentView{}, err
	}
	if profile.MatchingPaused {
		return IntentView{}, forbidden("你已暂停匹配，请先在资料中重新开启")
	}
	message = strings.TrimSpace(message)
	if utf8.RuneCountInString(message) > maxInterestMessageLength {
		return IntentView{}, invalidParams("附言不能超过 500 个字")
	}
	if intent.Category == CategoryRomantic {
		if !profile.RomanticOptIn {
			return IntentView{}, forbidden("请先在资料中主动开启认真关系匹配")
		}
	}
	record, err := s.store.GetConnectionIntent(ctx, nil, userID, intentID)
	if err != nil {
		return IntentView{}, mapStoreError(err, "找不到这个意图")
	}
	now := s.now()
	if err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		if err := s.store.SetIntentInterest(ctx, tx, intentID, userID, "interested", message, now); err != nil {
			return err
		}
		if err := s.store.RevokeRecommendationInteractionSignal(ctx, tx, userID, intentID, "withdrawn", now); err != nil {
			return err
		}
		return s.recordRecommendationSignal(ctx, tx, userID, record, "interested", 3, now)
	}); err != nil {
		return IntentView{}, httpapi.InternalError(err)
	}
	return s.GetIntent(ctx, userID, intentID)
}

func (s *Service) WithdrawIntentInterest(ctx context.Context, userID, intentID int64) (IntentView, error) {
	if _, err := s.GetIntent(ctx, userID, intentID); err != nil {
		return IntentView{}, err
	}
	record, err := s.store.GetConnectionIntent(ctx, nil, userID, intentID)
	if err != nil {
		return IntentView{}, mapStoreError(err, "找不到这个意图")
	}
	now := s.now()
	if err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		if err := s.store.SetIntentInterest(ctx, tx, intentID, userID, "withdrawn", "", now); err != nil {
			return err
		}
		if err := s.store.RevokeRecommendationInteractionSignal(ctx, tx, userID, intentID, "interested", now); err != nil {
			return err
		}
		return s.recordRecommendationSignal(ctx, tx, userID, record, "withdrawn", -1, now)
	}); err != nil {
		return IntentView{}, httpapi.InternalError(err)
	}
	return s.GetIntent(ctx, userID, intentID)
}

func (s *Service) ListIntentInterests(ctx context.Context, userID, intentID int64) ([]InterestView, error) {
	intent, err := s.store.GetConnectionIntent(ctx, nil, userID, intentID)
	if err != nil {
		return nil, mapStoreError(err, "找不到这个意图")
	}
	if intent.AuthorID != userID {
		return nil, forbidden("只有发布者可以查看回应")
	}
	records, err := s.store.ListIntentInterests(ctx, nil, intentID)
	if err != nil {
		return nil, httpapi.InternalError(err)
	}
	views := make([]InterestView, 0, len(records))
	for _, record := range records {
		views = append(views, InterestView{
			UserID:             record.UserID,
			Nickname:           record.Nickname,
			Department:         record.Department,
			Grade:              record.Grade,
			Bio:                record.Bio,
			Skills:             nonNilStrings(record.Skills),
			CollaborationStyle: record.CollaborationStyle,
			Status:             record.Status,
			Message:            record.Message,
			CreatedAt:          record.CreatedAt,
			UpdatedAt:          record.UpdatedAt,
		})
	}
	return views, nil
}

func (s *Service) RespondToInterest(ctx context.Context, ownerID, intentID, candidateID int64, accepted bool) (*MatchRoomView, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}

	starters := []string{}
	starterInference := InferenceView{
		Source:        "deterministic",
		Confidence:    0.74,
		PromptVersion: "deterministic-starters-v2",
		Editable:      true,
	}
	if accepted {
		if source, sourceErr := s.store.GetConnectionIntent(ctx, nil, ownerID, intentID); sourceErr == nil && source.AuthorID == ownerID {
			starters, starterInference = s.conversationStarterDrafts(ctx, source)
		}
	}
	var roomID int64
	err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		intent, err := s.store.GetConnectionIntent(ctx, tx, ownerID, intentID)
		if err != nil {
			return mapStoreError(err, "找不到这个意图")
		}
		if intent.AuthorID != ownerID {
			return forbidden("只有发布者可以处理回应")
		}
		if intent.Status != StatusActive ||
			(intent.ExpiresAt.Valid && !intent.ExpiresAt.Time.After(s.now())) {
			return forbidden("这个意图已经停止匹配")
		}
		if accepted {
			for _, participantID := range []int64{ownerID, candidateID} {
				profile, profileErr := s.store.GetMatchingProfile(ctx, tx, participantID)
				if errors.Is(profileErr, store.ErrNotFound) {
					if intent.Category == CategoryRomantic {
						return forbidden("双方都需要主动开启认真关系匹配")
					}
					continue
				}
				if profileErr != nil {
					return httpapi.InternalError(profileErr)
				}
				if profile.MatchingPaused {
					return forbidden("其中一方已经暂停新匹配")
				}
				if intent.Category == CategoryRomantic && !profile.RomanticOptIn {
					return forbidden("双方都需要主动开启认真关系匹配")
				}
			}
		}

		status := "declined"
		if accepted {
			status = "accepted"
		}
		if err := s.store.UpdateIntentInterestStatus(ctx, tx, intentID, candidateID, status, s.now()); err != nil {
			return mapStoreError(err, "这个回应已被处理或不存在")
		}
		signalType := "declined_candidate"
		signalWeight := -1
		if accepted {
			signalType = "accepted_candidate"
			signalWeight = 3
		}
		if err := s.recordRecommendationSignal(ctx, tx, ownerID, store.ConnectionIntent{
			ID:       intent.ID,
			AuthorID: candidateID,
			Category: intent.Category,
		}, signalType, signalWeight, s.now()); err != nil {
			return err
		}
		if !accepted {
			return nil
		}

		contextItems := []string{categoryLabel(intent.Category)}
		if intent.TimeWindow != "" {
			contextItems = append(contextItems, intent.TimeWindow)
		}
		if intent.Location != "" {
			contextItems = append(contextItems, intent.Location)
		}
		if len(starters) == 0 {
			starters = startersForCategory(intent.Category, intent.Title)
		}
		roomID, err = s.store.CreateMatchRoom(ctx, tx, store.CreateMatchRoomParams{
			IntentID:             intent.ID,
			OwnerID:              ownerID,
			CandidateID:          candidateID,
			FitReason:            "对方主动回应了你的「" + intent.Title + "」，而你也愿意进一步了解。这次连接来自明确的共同意愿。",
			Uncertainty:          "SouLink 只确认了双方愿意见面；具体节奏和边界仍需要你们亲自沟通。",
			SharedContext:        contextItems,
			ConversationStarters: starters,
			SuggestedPlaces:      placesForIntent(intent.Location),
			AssistantSource:      starterInference.Source,
			AssistantModel:       starterInference.Model,
			AssistantConfidence:  starterInference.Confidence,
			AssistantPrompt:      starterInference.PromptVersion,
			Now:                  s.now(),
		})
		if err != nil {
			return httpapi.InternalError(err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if !accepted {
		return nil, nil
	}
	room, err := s.GetMatchRoom(ctx, ownerID, roomID)
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func (s *Service) Recommendations(ctx context.Context, userID, intentID int64) ([]RecommendationView, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	source, err := s.store.GetConnectionIntent(ctx, nil, userID, intentID)
	if err != nil {
		return nil, mapStoreError(err, "找不到这个意图")
	}
	if source.AuthorID != userID {
		return nil, forbidden("只能为自己的意图生成推荐")
	}
	if source.Status != StatusActive {
		return []RecommendationView{}, nil
	}
	profile, err := s.GetMatchingProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile.MatchingPaused || stringListContains(profile.HiddenCategories, source.Category) {
		return []RecommendationView{}, nil
	}
	if source.Category == CategoryRomantic {
		if !profile.RomanticOptIn {
			return nil, forbidden("请先在资料中主动开启认真关系匹配")
		}
	}

	return s.recommendationsForSource(ctx, userID, source, 3, false)
}

func (s *Service) ListMatchRooms(ctx context.Context, userID int64) ([]MatchRoomView, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	rooms, err := s.store.ListMatchRooms(ctx, nil, userID, 50)
	if err != nil {
		return nil, httpapi.InternalError(err)
	}
	views := make([]MatchRoomView, 0, len(rooms))
	for _, room := range rooms {
		views = append(views, toMatchRoomView(room, userID))
	}
	return views, nil
}

func (s *Service) GetMatchRoom(ctx context.Context, userID, roomID int64) (MatchRoomView, error) {
	if err := s.ensureReady(); err != nil {
		return MatchRoomView{}, err
	}
	room, err := s.store.GetMatchRoom(ctx, nil, userID, roomID)
	if err != nil {
		return MatchRoomView{}, mapStoreError(err, "找不到这个匹配房间")
	}
	return toMatchRoomView(room, userID), nil
}

func (s *Service) UpdateMatchRoomPlan(ctx context.Context, userID, roomID int64, scheduledAt *time.Time, place string) (MatchRoomView, error) {
	place = strings.TrimSpace(place)
	if utf8.RuneCountInString(place) > 160 {
		return MatchRoomView{}, invalidParams("见面地点太长")
	}
	var scheduled sql.NullTime
	if scheduledAt != nil {
		if scheduledAt.Before(s.now().Add(-5 * time.Minute)) {
			return MatchRoomView{}, invalidParams("见面时间不能早于现在")
		}
		scheduled = sql.NullTime{Time: *scheduledAt, Valid: true}
	}
	if err := s.store.UpdateMatchRoomPlan(ctx, nil, userID, roomID, scheduled, place, s.now()); err != nil {
		return MatchRoomView{}, mapStoreError(err, "找不到这个匹配房间")
	}
	return s.GetMatchRoom(ctx, userID, roomID)
}

func (s *Service) UpdateMatchRoomStatus(ctx context.Context, userID, roomID int64, status string) (MatchRoomView, error) {
	if err := s.ensureReady(); err != nil {
		return MatchRoomView{}, err
	}
	status = strings.ToLower(strings.TrimSpace(status))
	if !validRoomStatusTarget(status) {
		return MatchRoomView{}, invalidParams("未知的匹配房间状态")
	}
	if err := s.store.UpdateMatchRoomStatus(ctx, nil, userID, roomID, status, s.now()); err != nil {
		return MatchRoomView{}, mapStoreError(err, "找不到可更新的匹配房间")
	}
	return s.GetMatchRoom(ctx, userID, roomID)
}

func (s *Service) SetMatchRoomContactShared(ctx context.Context, userID, roomID int64, shared bool) (MatchRoomView, error) {
	if err := s.store.SetMatchRoomContactShared(ctx, nil, userID, roomID, shared, s.now()); err != nil {
		return MatchRoomView{}, mapStoreError(err, "找不到这个匹配房间")
	}
	return s.GetMatchRoom(ctx, userID, roomID)
}

func (s *Service) ListMatchRoomMessages(ctx context.Context, userID, roomID int64) ([]MatchRoomMessageView, error) {
	if _, err := s.GetMatchRoom(ctx, userID, roomID); err != nil {
		return nil, err
	}
	messages, err := s.store.ListMatchRoomMessages(ctx, nil, roomID, userID, 200)
	if err != nil {
		return nil, httpapi.InternalError(err)
	}
	return mapMessageViews(messages), nil
}

func (s *Service) CreateMatchRoomMessage(ctx context.Context, userID, roomID int64, content string) (MatchRoomMessageView, error) {
	content = strings.TrimSpace(content)
	if content == "" || utf8.RuneCountInString(content) > maxRoomMessageLength {
		return MatchRoomMessageView{}, invalidParams("消息需要在 1 到 1000 个字之间")
	}
	message, err := s.store.CreateMatchRoomMessage(ctx, nil, roomID, userID, content, s.now())
	if err != nil {
		return MatchRoomMessageView{}, mapStoreError(err, "找不到这个匹配房间")
	}
	return toMessageView(message), nil
}

func (s *Service) prepareIntent(ctx context.Context, userID int64, input IntentInput, current *store.ConnectionIntent, now time.Time) (IntentInput, sql.NullTime, sql.NullTime, error) {
	input.Category = strings.TrimSpace(input.Category)
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	input.Goal = strings.TrimSpace(input.Goal)
	input.TimeWindow = strings.TrimSpace(input.TimeWindow)
	input.Location = strings.TrimSpace(input.Location)
	input.Visibility = strings.TrimSpace(input.Visibility)
	input.Status = strings.TrimSpace(input.Status)
	input.DesiredNextStep = strings.TrimSpace(input.DesiredNextStep)

	if !validCategory(input.Category) {
		return IntentInput{}, sql.NullTime{}, sql.NullTime{}, invalidParams("未知的意图类别")
	}
	if input.Title == "" || utf8.RuneCountInString(input.Title) > maxIntentTitleLength {
		return IntentInput{}, sql.NullTime{}, sql.NullTime{}, invalidParams("标题需要在 1 到 80 个字之间")
	}
	if input.Description == "" || utf8.RuneCountInString(input.Description) > maxIntentDescriptionLength {
		return IntentInput{}, sql.NullTime{}, sql.NullTime{}, invalidParams("描述需要在 1 到 1200 个字之间")
	}
	for _, value := range []string{input.Goal, input.TimeWindow, input.Location, input.DesiredNextStep} {
		if utf8.RuneCountInString(value) > maxIntentShortTextLength {
			return IntentInput{}, sql.NullTime{}, sql.NullTime{}, invalidParams("意图字段内容过长")
		}
	}
	if !validVisibility(input.Visibility) {
		input.Visibility = VisibilityCandidates
	}
	if input.Status == "" {
		input.Status = StatusActive
	}
	if !validWritableStatus(input.Status) {
		return IntentInput{}, sql.NullTime{}, sql.NullTime{}, invalidParams("未知的意图状态")
	}
	if input.Category == CategoryRomantic {
		input.Visibility = VisibilityCandidates
	}

	var err error
	input.Constraints, err = normalizeStringList(input.Constraints, 12, 40)
	if err != nil {
		return IntentInput{}, sql.NullTime{}, sql.NullTime{}, err
	}
	input.DealBreakers, err = normalizeStringList(input.DealBreakers, 8, 80)
	if err != nil {
		return IntentInput{}, sql.NullTime{}, sql.NullTime{}, err
	}
	if input.Category == CategoryProject || input.Category == CategoryCofounder {
		if input.ProjectMemo == nil {
			input.ProjectMemo = &ProjectMemo{}
		}
		normalizedMemo, err := normalizeProjectMemo(*input.ProjectMemo)
		if err != nil {
			return IntentInput{}, sql.NullTime{}, sql.NullTime{}, err
		}
		input.ProjectMemo = &normalizedMemo
	} else {
		input.ProjectMemo = nil
	}

	profile, err := s.GetMatchingProfile(ctx, userID)
	if err != nil {
		return IntentInput{}, sql.NullTime{}, sql.NullTime{}, err
	}
	if input.Status == StatusActive && profile.MatchingPaused {
		return IntentInput{}, sql.NullTime{}, sql.NullTime{}, forbidden("你已暂停匹配，请先在资料中重新开启")
	}
	if input.Category == CategoryRomantic && !profile.RomanticOptIn {
		return IntentInput{}, sql.NullTime{}, sql.NullTime{}, forbidden("请先在资料中主动开启认真关系匹配")
	}

	expiresAt := sql.NullTime{Valid: true}
	if input.ExpiresAt != nil {
		expiresAt.Time = *input.ExpiresAt
	} else {
		expiresAt.Time = now.Add(defaultIntentLifetime(input.Category))
	}
	if expiresAt.Time.Before(now.Add(30 * time.Minute)) {
		return IntentInput{}, sql.NullTime{}, sql.NullTime{}, invalidParams("意图至少需要保留 30 分钟")
	}
	if expiresAt.Time.After(now.Add(90 * 24 * time.Hour)) {
		return IntentInput{}, sql.NullTime{}, sql.NullTime{}, invalidParams("意图有效期不能超过 90 天")
	}

	publishedAt := sql.NullTime{}
	if input.Status == StatusActive {
		publishedAt = sql.NullTime{Time: now, Valid: true}
		if current != nil && current.PublishedAt.Valid {
			publishedAt = current.PublishedAt
		}
	}
	return input, expiresAt, publishedAt, nil
}

func normalizeMatchingProfile(input MatchingProfileInput) (MatchingProfileInput, error) {
	input.Bio = strings.TrimSpace(input.Bio)
	input.Availability = strings.TrimSpace(input.Availability)
	input.CollaborationStyle = strings.TrimSpace(input.CollaborationStyle)
	input.Boundaries = strings.TrimSpace(input.Boundaries)
	input.AIPortrait = strings.TrimSpace(input.AIPortrait)

	textFields := []string{
		input.Bio,
		input.Availability,
		input.CollaborationStyle,
		input.Boundaries,
		input.AIPortrait,
	}
	for _, value := range textFields {
		if utf8.RuneCountInString(value) > 1000 {
			return MatchingProfileInput{}, invalidParams("资料字段不能超过 1000 个字")
		}
	}

	var err error
	input.Skills, err = normalizeStringList(input.Skills, 20, 40)
	if err != nil {
		return MatchingProfileInput{}, err
	}
	input.Interests, err = normalizeStringList(input.Interests, 20, 40)
	if err != nil {
		return MatchingProfileInput{}, err
	}
	input.PreferredScenes, err = normalizeStringList(input.PreferredScenes, 12, 60)
	if err != nil {
		return MatchingProfileInput{}, err
	}
	input.HiddenCategories, err = normalizeCategoryList(input.HiddenCategories)
	if err != nil {
		return MatchingProfileInput{}, err
	}
	if input.AIPortrait == "" {
		input.AIPortraitApproved = false
	}
	return input, nil
}

func normalizeProjectMemo(input ProjectMemo) (ProjectMemo, error) {
	input.Building = strings.TrimSpace(input.Building)
	input.Stage = strings.TrimSpace(input.Stage)
	input.WeeklyTime = strings.TrimSpace(input.WeeklyTime)
	input.Compensation = strings.TrimSpace(input.Compensation)
	input.WorkingStyle = strings.TrimSpace(input.WorkingStyle)
	input.RiskAppetite = strings.TrimSpace(input.RiskAppetite)

	if utf8.RuneCountInString(input.Building) > 600 {
		return ProjectMemo{}, invalidParams("项目简介不能超过 600 个字")
	}
	for _, value := range []string{
		input.Stage,
		input.WeeklyTime,
		input.Compensation,
		input.WorkingStyle,
		input.RiskAppetite,
	} {
		if utf8.RuneCountInString(value) > 240 {
			return ProjectMemo{}, invalidParams("项目 memo 字段不能超过 240 个字")
		}
	}

	var err error
	input.RequiredSkills, err = normalizeStringList(input.RequiredSkills, 12, 60)
	if err != nil {
		return ProjectMemo{}, err
	}
	input.NiceToHaveSkills, err = normalizeStringList(input.NiceToHaveSkills, 12, 60)
	if err != nil {
		return ProjectMemo{}, err
	}
	return input, nil
}

func normalizeCategoryList(values []string) ([]string, error) {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" {
			continue
		}
		if !validCategory(value) {
			return nil, invalidParams("未知的隐藏类别")
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result, nil
}

func normalizeStringList(values []string, maxItems, maxLength int) ([]string, error) {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if utf8.RuneCountInString(value) > maxLength {
			return nil, invalidParams("标签内容过长")
		}
		key := strings.ToLower(value)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, value)
		if len(result) > maxItems {
			return nil, invalidParams("标签数量过多")
		}
	}
	return result, nil
}

func scoreRecommendation(source, candidate store.ConnectionIntent) (int, string, string, []string) {
	score := 46
	shared := make([]string, 0, 4)
	reasons := make([]string, 0, 4)

	if source.Category == candidate.Category {
		score += 24
		shared = append(shared, categoryLabel(source.Category))
		reasons = append(reasons, "你们正在寻找相近类型的连接")
	} else if categoriesCompatible(source.Category, candidate.Category) {
		score += 18
		reasons = append(reasons, "你们的目标具有互补性")
	} else {
		score += 4
	}

	if source.Location != "" && candidate.Location != "" && strings.EqualFold(source.Location, candidate.Location) {
		score += 10
		shared = append(shared, source.Location)
		reasons = append(reasons, "偏好的校内地点一致")
	}
	if source.TimeWindow != "" && candidate.TimeWindow != "" && strings.EqualFold(source.TimeWindow, candidate.TimeWindow) {
		score += 10
		shared = append(shared, source.TimeWindow)
		reasons = append(reasons, "时间窗口能够直接对上")
	}

	commonConstraints := intersection(source.Constraints, candidate.Constraints)
	if len(commonConstraints) > 0 {
		added := len(commonConstraints) * 4
		if added > 12 {
			added = 12
		}
		score += added
		shared = append(shared, commonConstraints...)
		reasons = append(reasons, "你们对见面方式有共同偏好")
	}

	if source.Category == CategoryProject || source.Category == CategoryCofounder {
		requiredSkillMatches := fuzzyIntersection(
			source.ProjectMemo.RequiredSkills,
			candidate.AuthorSkills,
		)
		if len(requiredSkillMatches) > 0 {
			added := len(requiredSkillMatches) * 6
			if added > 18 {
				added = 18
			}
			score += added
			shared = append(shared, requiredSkillMatches...)
			reasons = append(reasons, "对方的显式技能与项目必备能力有直接交集")
		}

		reciprocalSkillMatches := fuzzyIntersection(
			candidate.ProjectMemo.RequiredSkills,
			source.AuthorSkills,
		)
		if len(reciprocalSkillMatches) > 0 {
			score += min(len(reciprocalSkillMatches)*4, 12)
			shared = append(shared, reciprocalSkillMatches...)
			reasons = append(reasons, "你的技能也能回应对方正在寻找的能力")
		}

		if source.ProjectMemo.WeeklyTime != "" &&
			candidate.AuthorAvailability != "" &&
			hasSharedTimeSignal(
				source.ProjectMemo.WeeklyTime,
				candidate.AuthorAvailability,
			) {
			score += 4
			reasons = append(reasons, "投入节奏存在可以进一步对齐的线索")
		}
	}

	commonSignals := sharedDescriptionSignals(source.Description, candidate.Description)
	if len(commonSignals) > 0 {
		added := len(commonSignals) * 3
		if added > 9 {
			added = 9
		}
		score += added
		shared = append(shared, commonSignals...)
		reasons = append(reasons, "描述里出现了相近的兴趣或能力线索")
	}
	if score > 96 {
		score = 96
	}
	shared = uniqueStrings(shared, 4)

	if len(reasons) == 0 {
		reasons = append(reasons, "对方的需求与这次意图存在可进一步确认的交集")
	}
	reason := strings.Join(uniqueStrings(reasons, 3), "；") + "。"

	uncertainties := make([]string, 0, 2)
	if candidate.TimeWindow == "" {
		uncertainties = append(uncertainties, "对方尚未明确时间")
	}
	if candidate.Location == "" {
		uncertainties = append(uncertainties, "地点偏好还未确认")
	}
	if source.Category == CategoryProject || source.Category == CategoryCofounder {
		if source.ProjectMemo.WeeklyTime == "" || candidate.AuthorAvailability == "" {
			uncertainties = append(uncertainties, "实际投入时间仍需确认")
		}
		if source.ProjectMemo.Compensation == "" {
			uncertainties = append(uncertainties, "权益或补偿方式尚未说明")
		}
		if len(uncertainties) == 0 {
			uncertainties = append(uncertainties, "分工、决策方式和风险偏好仍需当面确认")
		}
	}
	if len(uncertainties) == 0 {
		uncertainties = append(uncertainties, "文字意图无法替代一次真实交流")
	}

	return score, reason, strings.Join(uniqueStrings(uncertainties, 2), "；") + "。", shared
}

func toMatchingProfileView(profile store.MatchingProfile) MatchingProfileView {
	updatedAt := profile.UpdatedAt
	return MatchingProfileView{
		MatchingProfileInput: MatchingProfileInput{
			Bio:                profile.Bio,
			Skills:             nonNilStrings(profile.Skills),
			Interests:          nonNilStrings(profile.Interests),
			Availability:       profile.Availability,
			PreferredScenes:    nonNilStrings(profile.PreferredScenes),
			CollaborationStyle: profile.CollaborationStyle,
			Boundaries:         profile.Boundaries,
			AIPortrait:         profile.AIPortrait,
			AIPortraitApproved: profile.AIPortraitApproved,
			RomanticOptIn:      profile.RomanticOptIn,
			MatchingPaused:     profile.MatchingPaused,
			HiddenCategories:   nonNilStrings(profile.HiddenCategories),
		},
		UpdatedAt: &updatedAt,
	}
}

func contextualProfileView(
	user store.User,
	profile store.MatchingProfile,
	intent store.ConnectionIntent,
) ContextProfileView {
	view := ContextProfileView{
		UserID:        user.ID,
		Nickname:      user.Nickname,
		SchoolName:    "清华大学",
		Department:    user.Department,
		Grade:         user.Grade,
		Verified:      true,
		Bio:           strings.TrimSpace(profile.Bio),
		Availability:  strings.TrimSpace(profile.Availability),
		VisibleFields: []string{},
		Intent: ContextProfileIntent{
			ID:              intent.ID,
			Category:        intent.Category,
			Title:           intent.Title,
			Goal:            intent.Goal,
			TimeWindow:      intent.TimeWindow,
			Location:        intent.Location,
			DesiredNextStep: intent.DesiredNextStep,
		},
	}
	if view.Department != "" {
		view.VisibleFields = append(view.VisibleFields, "department")
	}
	if view.Grade != "" {
		view.VisibleFields = append(view.VisibleFields, "grade")
	}
	if view.Bio != "" {
		view.VisibleFields = append(view.VisibleFields, "bio")
	}
	if view.Availability != "" {
		view.VisibleFields = append(view.VisibleFields, "availability")
	}

	switch intent.Category {
	case CategoryProject, CategoryCofounder:
		view.Skills = nonNilStrings(profile.Skills)
		view.CollaborationStyle = strings.TrimSpace(profile.CollaborationStyle)
		if len(view.Skills) > 0 {
			view.VisibleFields = append(view.VisibleFields, "skills")
		}
		if view.CollaborationStyle != "" {
			view.VisibleFields = append(view.VisibleFields, "collaborationStyle")
		}
	case CategoryStudy:
		view.Skills = nonNilStrings(profile.Skills)
		view.Interests = nonNilStrings(profile.Interests)
		view.PreferredScenes = nonNilStrings(profile.PreferredScenes)
		view.CollaborationStyle = strings.TrimSpace(profile.CollaborationStyle)
		if len(view.Skills) > 0 {
			view.VisibleFields = append(view.VisibleFields, "skills")
		}
		if len(view.Interests) > 0 {
			view.VisibleFields = append(view.VisibleFields, "interests")
		}
		if len(view.PreferredScenes) > 0 {
			view.VisibleFields = append(view.VisibleFields, "preferredScenes")
		}
		if view.CollaborationStyle != "" {
			view.VisibleFields = append(view.VisibleFields, "collaborationStyle")
		}
	default:
		view.Interests = nonNilStrings(profile.Interests)
		view.PreferredScenes = nonNilStrings(profile.PreferredScenes)
		if len(view.Interests) > 0 {
			view.VisibleFields = append(view.VisibleFields, "interests")
		}
		if len(view.PreferredScenes) > 0 {
			view.VisibleFields = append(view.VisibleFields, "preferredScenes")
		}
	}

	return view
}

func toIntentView(intent store.ConnectionIntent, viewerID int64) IntentView {
	isOwner := intent.AuthorID == viewerID
	author := AuthorPreview{
		UserID:     intent.AuthorID,
		Nickname:   intent.AuthorNickname,
		Department: intent.AuthorDepartment,
		Grade:      intent.AuthorGrade,
		Verified:   true,
	}
	if intent.Category == CategoryRomantic && !isOwner && intent.InterestState != "accepted" {
		author = AuthorPreview{
			Nickname:   "一位同学",
			Department: "清华大学",
			Grade:      "身份已验证",
			Verified:   true,
		}
	}

	view := IntentView{
		ID:              intent.ID,
		Author:          author,
		Category:        intent.Category,
		Title:           intent.Title,
		Description:     intent.Description,
		Goal:            intent.Goal,
		TimeWindow:      intent.TimeWindow,
		Location:        intent.Location,
		Visibility:      intent.Visibility,
		Status:          intent.Status,
		Constraints:     nonNilStrings(intent.Constraints),
		DesiredNextStep: intent.DesiredNextStep,
		InterestCount:   intent.InterestCount,
		InterestState:   intent.InterestState,
		SavedByMe:       intent.SavedByMe,
		IsOwner:         isOwner,
		CreatedAt:       intent.CreatedAt,
		UpdatedAt:       intent.UpdatedAt,
	}
	if intent.Category == CategoryProject || intent.Category == CategoryCofounder {
		memo := fromStoreProjectMemo(intent.ProjectMemo)
		view.ProjectMemo = &memo
	}
	if isOwner {
		view.DealBreakers = nonNilStrings(intent.DealBreakers)
	}
	if intent.ExpiresAt.Valid {
		value := intent.ExpiresAt.Time
		view.ExpiresAt = &value
	}
	if intent.PublishedAt.Valid {
		value := intent.PublishedAt.Time
		view.PublishedAt = &value
	}
	return view
}

func toStoreProjectMemo(memo *ProjectMemo) store.ProjectMemo {
	if memo == nil {
		return store.ProjectMemo{}
	}
	return store.ProjectMemo{
		Building:         memo.Building,
		Stage:            memo.Stage,
		WeeklyTime:       memo.WeeklyTime,
		Compensation:     memo.Compensation,
		RequiredSkills:   nonNilStrings(memo.RequiredSkills),
		NiceToHaveSkills: nonNilStrings(memo.NiceToHaveSkills),
		WorkingStyle:     memo.WorkingStyle,
		RiskAppetite:     memo.RiskAppetite,
	}
}

func fromStoreProjectMemo(memo store.ProjectMemo) ProjectMemo {
	return ProjectMemo{
		Building:         memo.Building,
		Stage:            memo.Stage,
		WeeklyTime:       memo.WeeklyTime,
		Compensation:     memo.Compensation,
		RequiredSkills:   nonNilStrings(memo.RequiredSkills),
		NiceToHaveSkills: nonNilStrings(memo.NiceToHaveSkills),
		WorkingStyle:     memo.WorkingStyle,
		RiskAppetite:     memo.RiskAppetite,
	}
}

func mapIntentViews(intents []store.ConnectionIntent, viewerID int64) []IntentView {
	views := make([]IntentView, 0, len(intents))
	for _, intent := range intents {
		views = append(views, toIntentView(intent, viewerID))
	}
	return views
}

func toMatchRoomView(room store.MatchRoom, viewerID int64) MatchRoomView {
	owner := RoomParticipantView{
		UserID:     room.OwnerID,
		Nickname:   room.OwnerNickname,
		Department: room.OwnerDepartment,
		Grade:      room.OwnerGrade,
	}
	candidate := RoomParticipantView{
		UserID:     room.CandidateID,
		Nickname:   room.CandidateNickname,
		Department: room.CandidateDepartment,
		Grade:      room.CandidateGrade,
	}

	me := owner
	other := candidate
	myShared := room.OwnerContactShared
	otherShared := room.CandidateContactShared
	if viewerID == room.CandidateID {
		me = candidate
		other = owner
		myShared = room.CandidateContactShared
		otherShared = room.OwnerContactShared
		if room.OwnerContactShared && room.OwnerContactValue != "" {
			other.Contact = &ContactView{Type: room.OwnerContactType, Value: room.OwnerContactValue}
		}
	} else if room.CandidateContactShared && room.CandidateContactValue != "" {
		other.Contact = &ContactView{Type: room.CandidateContactType, Value: room.CandidateContactValue}
	}

	view := MatchRoomView{
		ID:                   room.ID,
		IntentID:             room.IntentID,
		IntentTitle:          room.IntentTitle,
		IntentCategory:       room.IntentCategory,
		Me:                   me,
		Other:                other,
		Status:               room.Status,
		FitReason:            room.FitReason,
		Uncertainty:          room.Uncertainty,
		SharedContext:        nonNilStrings(room.SharedContext),
		ConversationStarters: nonNilStrings(room.ConversationStarters),
		Assistant: InferenceView{
			Source:        room.AssistantSource,
			Model:         room.AssistantModel,
			Confidence:    room.AssistantConfidence,
			PromptVersion: room.AssistantPromptVersion,
			Editable:      true,
		},
		SuggestedPlaces:    nonNilStrings(room.SuggestedPlaces),
		SelectedPlace:      room.SelectedPlace,
		MyContactShared:    myShared,
		OtherContactShared: otherShared,
		CreatedAt:          room.CreatedAt,
		UpdatedAt:          room.UpdatedAt,
	}
	if room.ScheduledAt.Valid {
		value := room.ScheduledAt.Time
		view.ScheduledAt = &value
	}
	return view
}

func mapMessageViews(messages []store.MatchRoomMessage) []MatchRoomMessageView {
	views := make([]MatchRoomMessageView, 0, len(messages))
	for _, message := range messages {
		views = append(views, toMessageView(message))
	}
	return views
}

func toMessageView(message store.MatchRoomMessage) MatchRoomMessageView {
	return MatchRoomMessageView{
		ID:             message.ID,
		RoomID:         message.RoomID,
		SenderID:       message.SenderID,
		SenderNickname: message.SenderNickname,
		Content:        message.Content,
		CreatedAt:      message.CreatedAt,
	}
}

func startersForCategory(category, title string) []string {
	switch category {
	case CategoryProject, CategoryCofounder:
		return []string{
			"先各用三分钟介绍自己正在做的事情，可以吗？",
			"这件事未来四周最需要解决的问题是什么？",
			"你理想中的协作节奏和分工是什么样？",
		}
	case CategoryRomantic:
		return []string{
			"最近有什么事情让你觉得生活很有意思？",
			"你喜欢怎样低压力地认识一个新朋友？",
			"如果在校园里散步，你最愿意从哪里开始？",
		}
	default:
		return []string{
			"你为什么会发出或回应「" + title + "」？",
			"这次见面怎样安排会让你更舒服？",
			"最近在清华最想分享的一件小事是什么？",
		}
	}
}

func placesForIntent(location string) []string {
	places := make([]string, 0, 4)
	if location != "" {
		places = append(places, location)
	}
	places = append(places, "图书馆北馆咖啡厅", "清华学堂前广场", "紫荆园")
	return uniqueStrings(places, 4)
}

func validCategory(value string) bool {
	switch value {
	case CategoryMeal, CategoryStudy, CategorySport, CategoryEvent,
		CategoryConversation, CategoryRomantic, CategoryProject,
		CategoryCofounder, CategoryOther:
		return true
	default:
		return false
	}
}

func validVisibility(value string) bool {
	switch value {
	case VisibilityPrivate, VisibilityCandidates, VisibilityFeed:
		return true
	default:
		return false
	}
}

func validWritableStatus(value string) bool {
	switch value {
	case StatusDraft, StatusActive, StatusPaused, StatusFulfilled:
		return true
	default:
		return false
	}
}

func validRoomStatusTarget(value string) bool {
	return value == RoomStatusMet || value == RoomStatusClosed
}

func defaultIntentLifetime(category string) time.Duration {
	switch category {
	case CategoryMeal, CategorySport, CategoryEvent:
		return 7 * 24 * time.Hour
	case CategoryProject, CategoryCofounder:
		return 30 * 24 * time.Hour
	default:
		return 14 * 24 * time.Hour
	}
}

func categoryLabel(category string) string {
	labels := map[string]string{
		CategoryMeal:         "饭搭子",
		CategoryStudy:        "学习搭子",
		CategorySport:        "运动搭子",
		CategoryEvent:        "活动同伴",
		CategoryConversation: "认真聊天",
		CategoryRomantic:     "认真关系",
		CategoryProject:      "项目伙伴",
		CategoryCofounder:    "CTO / 联合创始人",
		CategoryOther:        "新的连接",
	}
	return labels[category]
}

func categoriesCompatible(left, right string) bool {
	if left == right {
		return true
	}
	project := map[string]bool{CategoryProject: true, CategoryCofounder: true}
	if project[left] && project[right] {
		return true
	}
	casual := map[string]bool{
		CategoryMeal: true, CategoryEvent: true, CategoryConversation: true,
	}
	return casual[left] && casual[right]
}

func recommendationCategoryEligible(source, candidate string) bool {
	return categoriesCompatible(source, candidate)
}

func intersection(left, right []string) []string {
	rightSet := make(map[string]string, len(right))
	for _, value := range right {
		rightSet[strings.ToLower(value)] = value
	}
	result := make([]string, 0)
	for _, value := range left {
		if matched, ok := rightSet[strings.ToLower(value)]; ok {
			result = append(result, matched)
		}
	}
	return uniqueStrings(result, 4)
}

func fuzzyIntersection(required, available []string) []string {
	result := make([]string, 0, 4)
	for _, requirement := range required {
		requirementKey := strings.ToLower(strings.TrimSpace(requirement))
		if requirementKey == "" {
			continue
		}
		for _, skill := range available {
			skillKey := strings.ToLower(strings.TrimSpace(skill))
			if skillKey == "" {
				continue
			}
			if requirementKey == skillKey ||
				strings.Contains(requirementKey, skillKey) ||
				strings.Contains(skillKey, requirementKey) {
				result = append(result, skill)
				break
			}
		}
	}
	return uniqueStrings(result, 4)
}

func sharedDescriptionSignals(left, right string) []string {
	signals := []string{
		"AI", "创业", "前端", "后端", "算法", "产品", "设计", "音乐", "电影",
		"跑步", "羽毛球", "阅读", "摄影", "安静", "长期", "认真", "低压力",
	}
	result := make([]string, 0, 4)
	for _, signal := range signals {
		if containsAny(left, []string{signal}) && containsAny(right, []string{signal}) {
			result = append(result, signal)
		}
	}
	return result
}

func hasSharedTimeSignal(left, right string) bool {
	signals := []string{"每周", "工作日", "周末", "晚间", "晚上", "白天", "小时"}
	for _, signal := range signals {
		if containsAny(left, []string{signal}) && containsAny(right, []string{signal}) {
			return true
		}
	}
	return false
}

func uniqueStrings(values []string, limit int) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, value)
		if limit > 0 && len(result) == limit {
			break
		}
	}
	return result
}

func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func stringListContains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func (s *Service) ensureReady() error {
	if s == nil || s.store == nil {
		return httpapi.FeatureNotReady("connection matching")
	}
	return nil
}

func (s *Service) now() time.Time {
	if s.nowFunc != nil {
		return s.nowFunc()
	}
	return time.Now()
}

func invalidParams(message string) error {
	return httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, message)
}

func forbidden(message string) error {
	return httpapi.NewError(http.StatusForbidden, httpapi.CodeForbidden, message)
}

func mapStoreError(err error, notFoundMessage string) error {
	if errors.Is(err, store.ErrNotFound) {
		return httpapi.NewError(http.StatusNotFound, httpapi.CodeUserNotFound, notFoundMessage)
	}
	return httpapi.InternalError(err)
}
