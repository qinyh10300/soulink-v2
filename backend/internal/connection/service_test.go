package connection

import (
	"database/sql"
	"testing"
	"time"

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
	if len(preview.DealBreakers) != 0 {
		t.Fatalf("preview deal breakers = %#v, want hidden", preview.DealBreakers)
	}

	ownerView := toIntentView(record, 42)
	if ownerView.Author.Nickname != "真实昵称" {
		t.Fatalf("owner nickname = %q, want real nickname", ownerView.Author.Nickname)
	}
	if len(ownerView.DealBreakers) != 1 {
		t.Fatalf("owner deal breakers = %#v, want visible", ownerView.DealBreakers)
	}
}

func TestDefaultIntentLifetimeKeepsUrgentIntentsShort(t *testing.T) {
	t.Parallel()

	if defaultIntentLifetime(CategoryMeal) >= defaultIntentLifetime(CategoryProject) {
		t.Fatal("meal intent lifetime should be shorter than project intent lifetime")
	}
}
