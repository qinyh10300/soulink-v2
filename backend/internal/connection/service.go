package connection

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/store"
)

const (
	maxIntentTitleLength       = 80
	maxIntentDescriptionLength = 1200
	maxIntentShortTextLength   = 200
	maxInterestMessageLength   = 500
	maxRoomMessageLength       = 1000
)

type Service struct {
	store       *store.Store
	interpreter *Interpreter
	nowFunc     func() time.Time
}

func NewService(dataStore *store.Store, interpreter *Interpreter) *Service {
	if interpreter == nil {
		interpreter = NewInterpreter()
	}
	return &Service{
		store:       dataStore,
		interpreter: interpreter,
		nowFunc:     time.Now,
	}
}

func (s *Service) Interpret(raw string) (IntentInterpretation, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || utf8.RuneCountInString(raw) > maxIntentDescriptionLength {
		return IntentInterpretation{}, invalidParams("请用一句话描述你想认识谁")
	}
	return s.interpreter.Interpret(raw), nil
}

func (s *Service) GetMatchingProfile(ctx context.Context, userID int64) (MatchingProfileView, error) {
	if err := s.ensureReady(); err != nil {
		return MatchingProfileView{}, err
	}

	profile, err := s.store.GetMatchingProfile(ctx, nil, userID)
	if errors.Is(err, store.ErrNotFound) {
		return MatchingProfileView{
			MatchingProfileInput: MatchingProfileInput{
				Skills:          []string{},
				Interests:       []string{},
				PreferredScenes: []string{},
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
		if intent.Status != StatusActive || intent.Visibility == VisibilityPrivate {
			return IntentView{}, forbidden("这个意图当前不可见")
		}
	}
	return toIntentView(intent, viewerID), nil
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
	intents, err := s.store.ListFeedConnectionIntents(ctx, nil, userID, category, 24)
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
	if err := s.store.SetIntentSaved(ctx, nil, intentID, userID, saved, s.now()); err != nil {
		return IntentView{}, httpapi.InternalError(err)
	}
	return s.GetIntent(ctx, userID, intentID)
}

func (s *Service) SetIntentInterest(ctx context.Context, userID, intentID int64, message string) (IntentView, error) {
	intent, err := s.GetIntent(ctx, userID, intentID)
	if err != nil {
		return IntentView{}, err
	}
	if intent.IsOwner {
		return IntentView{}, invalidParams("不能回应自己的意图")
	}
	message = strings.TrimSpace(message)
	if utf8.RuneCountInString(message) > maxInterestMessageLength {
		return IntentView{}, invalidParams("附言不能超过 500 个字")
	}
	if intent.Category == CategoryRomantic {
		profile, err := s.GetMatchingProfile(ctx, userID)
		if err != nil {
			return IntentView{}, err
		}
		if !profile.RomanticOptIn {
			return IntentView{}, forbidden("请先在资料中主动开启认真关系匹配")
		}
	}
	if err := s.store.SetIntentInterest(ctx, nil, intentID, userID, "interested", message, s.now()); err != nil {
		return IntentView{}, httpapi.InternalError(err)
	}
	return s.GetIntent(ctx, userID, intentID)
}

func (s *Service) WithdrawIntentInterest(ctx context.Context, userID, intentID int64) (IntentView, error) {
	if _, err := s.GetIntent(ctx, userID, intentID); err != nil {
		return IntentView{}, err
	}
	if err := s.store.SetIntentInterest(ctx, nil, intentID, userID, "withdrawn", "", s.now()); err != nil {
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

	var roomID int64
	err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		intent, err := s.store.GetConnectionIntent(ctx, tx, ownerID, intentID)
		if err != nil {
			return mapStoreError(err, "找不到这个意图")
		}
		if intent.AuthorID != ownerID {
			return forbidden("只有发布者可以处理回应")
		}

		status := "declined"
		if accepted {
			status = "accepted"
		}
		if err := s.store.UpdateIntentInterestStatus(ctx, tx, intentID, candidateID, status, s.now()); err != nil {
			return mapStoreError(err, "这个回应已被处理或不存在")
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
		roomID, err = s.store.CreateMatchRoom(ctx, tx, store.CreateMatchRoomParams{
			IntentID:             intent.ID,
			OwnerID:              ownerID,
			CandidateID:          candidateID,
			FitReason:            "对方主动回应了你的「" + intent.Title + "」，而你也愿意进一步了解。这次连接来自明确的共同意愿。",
			Uncertainty:          "SouLink 只确认了双方愿意见面；具体节奏和边界仍需要你们亲自沟通。",
			SharedContext:        contextItems,
			ConversationStarters: startersForCategory(intent.Category, intent.Title),
			SuggestedPlaces:      placesForIntent(intent.Location),
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
	if source.Category == CategoryRomantic {
		profile, err := s.GetMatchingProfile(ctx, userID)
		if err != nil {
			return nil, err
		}
		if !profile.RomanticOptIn {
			return nil, forbidden("请先在资料中主动开启认真关系匹配")
		}
	}

	candidates, err := s.store.ListRecommendationCandidates(ctx, nil, userID, source.Category, 60)
	if err != nil {
		return nil, httpapi.InternalError(err)
	}
	recommendations := make([]RecommendationView, 0, len(candidates))
	for _, candidate := range candidates {
		score, reason, uncertainty, shared := scoreRecommendation(source, candidate)
		recommendations = append(recommendations, RecommendationView{
			Intent:        toIntentView(candidate, userID),
			Score:         score,
			FitReason:     reason,
			Uncertainty:   uncertainty,
			SharedContext: shared,
		})
	}
	sort.SliceStable(recommendations, func(i, j int) bool {
		if recommendations[i].Score == recommendations[j].Score {
			return recommendations[i].Intent.UpdatedAt.After(recommendations[j].Intent.UpdatedAt)
		}
		return recommendations[i].Score > recommendations[j].Score
	})
	if len(recommendations) > 3 {
		recommendations = recommendations[:3]
	}
	return recommendations, nil
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
	if input.AIPortrait == "" {
		input.AIPortraitApproved = false
	}
	return input, nil
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
		uncertainties = append(uncertainties, "实际投入时间与分工仍需当面确认")
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
		},
		UpdatedAt: &updatedAt,
	}
}

func toIntentView(intent store.ConnectionIntent, viewerID int64) IntentView {
	isOwner := intent.AuthorID == viewerID
	author := AuthorPreview{
		Nickname:   intent.AuthorNickname,
		Department: intent.AuthorDepartment,
		Grade:      intent.AuthorGrade,
		Verified:   true,
	}
	if intent.Category == CategoryRomantic && !isOwner {
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
		SuggestedPlaces:      nonNilStrings(room.SuggestedPlaces),
		SelectedPlace:        room.SelectedPlace,
		MyContactShared:      myShared,
		OtherContactShared:   otherShared,
		CreatedAt:            room.CreatedAt,
		UpdatedAt:            room.UpdatedAt,
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
