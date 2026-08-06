package connection

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"soulink-backend/internal/ai"
	"soulink-backend/internal/store"
)

func TestNormalizeStringListDeduplicatesAndTrims(t *testing.T) {
	t.Parallel()

	got, err := normalizeStringList([]string{" AI ", "产品", "ai", ""}, 4, 10)
	if err != nil {
		t.Fatalf("normalizeStringList() unexpected error: %v", err)
	}
	if len(got) != 2 || got[0] != "AI" || got[1] != "产品" {
		t.Fatalf("normalizeStringList() = %#v, want [AI 产品]", got)
	}
}

func TestNormalizeCategoryListDeduplicatesAndValidates(t *testing.T) {
	result, err := normalizeCategoryList([]string{" meal ", "PROJECT", "meal", ""})
	if err != nil {
		t.Fatalf("normalizeCategoryList() error = %v", err)
	}
	if len(result) != 2 || result[0] != CategoryMeal || result[1] != CategoryProject {
		t.Fatalf("normalizeCategoryList() = %#v", result)
	}

	if _, err := normalizeCategoryList([]string{"people-search"}); err == nil {
		t.Fatal("normalizeCategoryList() accepted an unknown category")
	}
}

func TestRecommendationScoringRewardsConcreteOverlap(t *testing.T) {
	t.Parallel()

	source := store.ConnectionIntent{
		Category:    CategoryProject,
		Description: "做一个 AI 校园产品，需要前端伙伴",
		TimeWindow:  "本周",
		Location:    "FIT楼",
		Constraints: []string{"长期", "认真"},
	}
	closeCandidate := store.ConnectionIntent{
		Category:    CategoryProject,
		Description: "对 AI 产品和前端开发感兴趣",
		TimeWindow:  "本周",
		Location:    "FIT楼",
		Constraints: []string{"长期"},
	}
	looseCandidate := store.ConnectionIntent{
		Category:    CategoryMeal,
		Description: "找人吃饭",
		TimeWindow:  "明天",
		Location:    "紫荆",
	}

	closeScore, _, _, shared := scoreRecommendation(source, closeCandidate)
	looseScore, _, _, _ := scoreRecommendation(source, looseCandidate)

	if closeScore <= looseScore {
		t.Fatalf("close score = %d, loose score = %d, want close score higher", closeScore, looseScore)
	}
	if len(shared) == 0 {
		t.Fatal("shared context is empty, want concrete overlap")
	}
}

func TestRecommendationScoringUsesProjectMemoAndProfileSkills(t *testing.T) {
	t.Parallel()

	source := store.ConnectionIntent{
		Category: CategoryCofounder,
		ProjectMemo: store.ProjectMemo{
			RequiredSkills: []string{"React", "工程架构"},
			WeeklyTime:     "每周 10 小时",
		},
	}
	candidate := store.ConnectionIntent{
		Category:           CategoryProject,
		AuthorSkills:       []string{"React", "前端工程"},
		AuthorAvailability: "每周工作日晚间",
	}

	score, reason, _, shared := scoreRecommendation(source, candidate)

	if score < 70 {
		t.Fatalf("score = %d, want strong project skill signal", score)
	}
	if !strings.Contains(reason, "显式技能") {
		t.Fatalf("reason = %q, want skill-specific explanation", reason)
	}
	if len(shared) == 0 || shared[0] != "React" {
		t.Fatalf("shared = %#v, want React", shared)
	}
}

func TestRecommendationCategoryEligibilityKeepsProjectPoolFocused(t *testing.T) {
	t.Parallel()

	if !recommendationCategoryEligible(CategoryCofounder, CategoryProject) {
		t.Fatal("cofounder intent should accept project candidates")
	}
	if recommendationCategoryEligible(CategoryCofounder, CategoryMeal) {
		t.Fatal("cofounder intent should reject meal candidates")
	}
	if recommendationCategoryEligible(CategoryStudy, CategorySport) {
		t.Fatal("study intent should reject sport candidates")
	}
	if !recommendationCategoryEligible(CategoryMeal, CategoryConversation) {
		t.Fatal("meal intent should accept low-pressure conversation candidates")
	}
}

func TestRomanticIntentPreviewHidesIdentityBeforeConsent(t *testing.T) {
	t.Parallel()

	now := time.Now()
	record := store.ConnectionIntent{
		ID:               7,
		AuthorID:         42,
		AuthorNickname:   "真实昵称",
		AuthorDepartment: "计算机系",
		AuthorGrade:      "博士二年级",
		Category:         CategoryRomantic,
		Title:            "认真认识一个人",
		Status:           StatusActive,
		Constraints:      []string{},
		DealBreakers:     []string{"不尊重边界"},
		ExpiresAt:        sql.NullTime{Time: now.Add(time.Hour), Valid: true},
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	preview := toIntentView(record, 99)
	if preview.Author.Nickname != "一位同学" {
		t.Fatalf("preview nickname = %q, want anonymized label", preview.Author.Nickname)
	}
	if preview.Author.UserID != 0 {
		t.Fatalf("preview author id = %d, want hidden", preview.Author.UserID)
	}
	if len(preview.DealBreakers) != 0 {
		t.Fatalf("preview deal breakers = %#v, want hidden", preview.DealBreakers)
	}

	accepted := record
	accepted.InterestState = "accepted"
	acceptedView := toIntentView(accepted, 99)
	if acceptedView.Author.UserID != 42 || acceptedView.Author.Nickname != "真实昵称" {
		t.Fatalf("accepted author = %#v, want consented identity", acceptedView.Author)
	}

	ownerView := toIntentView(record, 42)
	if ownerView.Author.Nickname != "真实昵称" {
		t.Fatalf("owner nickname = %q, want real nickname", ownerView.Author.Nickname)
	}
	if len(ownerView.DealBreakers) != 1 {
		t.Fatalf("owner deal breakers = %#v, want visible", ownerView.DealBreakers)
	}
}

func TestContextualProfileOnlyIncludesCategoryRelevantFields(t *testing.T) {
	t.Parallel()

	user := store.User{
		ID:         42,
		Nickname:   "项目同学",
		Department: "计算机系",
		Grade:      "博士二年级",
	}
	profile := store.MatchingProfile{
		Bio:                "做 AI 产品",
		Skills:             []string{"Go", "React"},
		Interests:          []string{"跑步", "电影"},
		Availability:       "工作日晚间",
		PreferredScenes:    []string{"紫荆", "图书馆"},
		CollaborationStyle: "先对齐目标再动手",
	}
	projectIntent := store.ConnectionIntent{
		ID:              7,
		Category:        CategoryProject,
		Title:           "一起做校园 AI 产品",
		Goal:            "两周完成原型",
		TimeWindow:      "本月",
		Location:        "FIT 楼",
		DesiredNextStep: "先聊 30 分钟",
	}

	projectView := contextualProfileView(user, profile, projectIntent)
	if len(projectView.Skills) != 2 || projectView.CollaborationStyle == "" {
		t.Fatalf("project context missing relevant fields: %#v", projectView)
	}
	if len(projectView.Interests) != 0 || len(projectView.PreferredScenes) != 0 {
		t.Fatalf("project context leaked unrelated fields: %#v", projectView)
	}

	mealIntent := projectIntent
	mealIntent.Category = CategoryMeal
	mealView := contextualProfileView(user, profile, mealIntent)
	if len(mealView.Interests) != 2 || len(mealView.PreferredScenes) != 2 {
		t.Fatalf("meal context missing relevant fields: %#v", mealView)
	}
	if len(mealView.Skills) != 0 || mealView.CollaborationStyle != "" {
		t.Fatalf("meal context leaked project fields: %#v", mealView)
	}
}

func TestDefaultIntentLifetimeKeepsUrgentIntentsShort(t *testing.T) {
	t.Parallel()

	if defaultIntentLifetime(CategoryMeal) >= defaultIntentLifetime(CategoryProject) {
		t.Fatal("meal intent lifetime should be shorter than project intent lifetime")
	}
}

func TestValidRoomStatusTargetRejectsReopening(t *testing.T) {
	t.Parallel()

	if !validRoomStatusTarget(RoomStatusMet) || !validRoomStatusTarget(RoomStatusClosed) {
		t.Fatal("met and closed should be valid room status targets")
	}
	if validRoomStatusTarget(RoomStatusOpen) {
		t.Fatal("clients must not be able to reopen a room")
	}
}

func TestLearningAdjustmentRewardsPositiveFeedbackAndPenalizesNegativeFeedback(t *testing.T) {
	t.Parallel()
	candidate := store.ConnectionIntent{AuthorID: 42, Category: CategoryStudy}
	positive := store.RecommendationLearningSnapshot{
		CategoryWeights: map[string]int{CategoryStudy: 3},
		UserWeights:     map[int64]int{42: 4},
	}
	negative := store.RecommendationLearningSnapshot{
		CategoryWeights: map[string]int{CategoryStudy: -3},
		UserWeights:     map[int64]int{42: -4},
	}

	positiveScore, positiveReason := applyLearningToRecommendation(70, "基础匹配。", candidate, positive)
	negativeScore, _ := applyLearningToRecommendation(70, "基础匹配。", candidate, negative)
	if positiveScore <= negativeScore {
		t.Fatalf("positive score = %d, negative score = %d", positiveScore, negativeScore)
	}
	if !strings.Contains(positiveReason, "最近") || !strings.Contains(positiveReason, "这位同学") {
		t.Fatalf("positive reason = %q", positiveReason)
	}
}

type recommendationGenerator struct {
	response string
}

func (generator recommendationGenerator) Generate(
	_ context.Context,
	prompt ai.Prompt,
	_ any,
	output any,
) (ai.Metadata, error) {
	return ai.Metadata{Source: "model", Model: "test-model", PromptVersion: prompt.Version}, json.Unmarshal([]byte(generator.response), output)
}

func TestModelRecommendationCannotInjectUnauthorizedCandidate(t *testing.T) {
	t.Parallel()
	service := &Service{generator: recommendationGenerator{response: `{
		"items":[
			{"candidateId":999,"fitReason":"越权候选","uncertainty":"无","confidence":1},
			{"candidateId":7,"fitReason":"共同时间明确。","uncertainty":"仍需确认见面节奏。","confidence":0.84}
		]
	}`}}
	recommendations := []RecommendationView{{
		Intent:      IntentView{ID: 7},
		FitReason:   "fallback",
		Uncertainty: "fallback",
		Inference:   InferenceView{Source: "deterministic"},
	}}

	result := service.enhanceRecommendationExplanations(t.Context(), store.ConnectionIntent{}, recommendations)
	if len(result) != 1 || result[0].Intent.ID != 7 {
		t.Fatalf("result = %#v, model must not add candidates", result)
	}
	if result[0].Inference.Source != "model" || result[0].FitReason != "共同时间明确。" {
		t.Fatalf("allowed explanation was not applied: %#v", result[0])
	}
}

func TestModelInputHidesContactDetails(t *testing.T) {
	t.Parallel()
	result := modelSafeText("微信 13800138000，邮箱 demo@example.com", 120)
	if strings.Contains(result, "13800138000") || strings.Contains(result, "demo@example.com") {
		t.Fatalf("model input leaked contact details: %q", result)
	}
}
