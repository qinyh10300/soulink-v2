package questionnaire

import (
	"strings"
	"testing"
	"time"
)

func TestLoadDefaultCatalogParsesMarkdownTemplate(t *testing.T) {
	catalog, err := LoadDefaultCatalog()
	if err != nil {
		t.Fatalf("LoadDefaultCatalog() error = %v", err)
	}

	if catalog.Title == "" {
		t.Fatal("catalog title is empty")
	}
	if catalog.Key != "thu-date" {
		t.Fatalf("catalog key = %q, want %q", catalog.Key, "thu-date")
	}
	if catalog.Version != "v0.6" {
		t.Fatalf("catalog version = %q, want %q", catalog.Version, "v0.6")
	}
	if catalog.TotalCount != 52 {
		t.Fatalf("catalog totalCount = %d, want %d", catalog.TotalCount, 52)
	}
	if len(catalog.Sections) == 0 {
		t.Fatal("catalog sections are empty")
	}
	if catalog.Sections[0].Title != "轮廓" {
		t.Fatalf("catalog first section title = %q, want %q", catalog.Sections[0].Title, "轮廓")
	}
	gender, ok := catalog.QuestionByID("1")
	if !ok {
		t.Fatal("question 1 not found")
	}
	if gender.Type != QuestionTypeSingleChoice {
		t.Fatalf("question 1 type = %q, want %q", gender.Type, QuestionTypeSingleChoice)
	}
	if len(gender.Options) != 2 {
		t.Fatalf("question 1 options = %d, want %d", len(gender.Options), 2)
	}
	if !gender.SupportsImportance {
		t.Fatal("question 1 should support importance")
	}
	interestQuestion, ok := catalog.QuestionByID("32")
	if !ok {
		t.Fatal("question 32 not found")
	}
	if len(interestQuestion.Options) == 0 {
		t.Fatal("question 32 options are empty, want parsed options")
	}

	ageRange, ok := catalog.QuestionByID("8")
	if !ok {
		t.Fatal("question 8 not found")
	}
	if ageRange.Type != QuestionTypeRange {
		t.Fatalf("question 8 type = %q, want %q", ageRange.Type, QuestionTypeRange)
	}
	if !ageRange.AllowUnlimited {
		t.Fatal("question 8 should allow unlimited")
	}

	scale, ok := catalog.QuestionByID("12")
	if !ok {
		t.Fatal("question 12 not found")
	}
	if scale.Type != QuestionTypeScale {
		t.Fatalf("question 12 type = %q, want %q", scale.Type, QuestionTypeScale)
	}
	if !scale.SupportsImportance {
		t.Fatal("question 12 should support importance")
	}

	location, ok := catalog.QuestionByID("6")
	if !ok {
		t.Fatal("question 6 not found")
	}
	if len(location.Options) != 15 {
		t.Fatalf("question 6 options = %d, want %d", len(location.Options), 15)
	}

	multiple, ok := catalog.QuestionByID("18")
	if !ok {
		t.Fatal("question 18 not found")
	}
	if multiple.Type != QuestionTypeMultipleChoice {
		t.Fatalf("question 18 type = %q, want %q", multiple.Type, QuestionTypeMultipleChoice)
	}
	if multiple.MaxSelections == nil || *multiple.MaxSelections != 5 {
		t.Fatal("question 18 maxSelections should be 5")
	}

	datePreference, ok := catalog.QuestionByID("23")
	if !ok {
		t.Fatal("question 23 not found")
	}
	if datePreference.Type != QuestionTypeMultipleChoice {
		t.Fatalf("question 23 type = %q, want %q", datePreference.Type, QuestionTypeMultipleChoice)
	}

	goal, ok := catalog.QuestionByID("52")
	if !ok {
		t.Fatal("question 52 not found")
	}
	if goal.Type != QuestionTypeMultipleChoice {
		t.Fatalf("question 52 type = %q, want %q", goal.Type, QuestionTypeMultipleChoice)
	}
	if goal.MaxSelections == nil || *goal.MaxSelections != 3 {
		t.Fatal("question 52 maxSelections should be 3")
	}
}

func TestParseCatalogSupportsStringVersionIsolation(t *testing.T) {
	markdown := strings.TrimSpace(`
# Campus Match 测试问卷

1. 模板标识：campus-match
2. 模板版本：v1.2.1

# 第一部分

## 1. 我的性别
1. 题型：单选题
2. 是否必答：是
3. 选项：
   1. 男
   2. 女`)

	catalog, err := ParseCatalog(markdown)
	if err != nil {
		t.Fatalf("ParseCatalog() error = %v", err)
	}

	if catalog.Key != "campus-match" {
		t.Fatalf("catalog key = %q, want %q", catalog.Key, "campus-match")
	}
	if catalog.Version != "v1.2.1" {
		t.Fatalf("catalog version = %q, want %q", catalog.Version, "v1.2.1")
	}
}

func TestParseCatalogParsesGroupedOptions(t *testing.T) {
	markdown := strings.TrimSpace(`
# Campus Match 测试问卷

1. 模板标识：campus-match
2. 模板版本：v1.2.1

# 第一部分

## 31. 你认可的特质
1. 题型：多选题
2. 是否必答：是
3. 选择限制：至少选择 2 项，最多选择 3 项
4. 选项：
   1. 生活方式：
      1. 规律作息
      2. 热爱运动
   2. 性格特点：
      1. 真诚直接
`)

	catalog, err := ParseCatalog(markdown)
	if err != nil {
		t.Fatalf("ParseCatalog() error = %v", err)
	}

	question, ok := catalog.QuestionByID("31")
	if !ok {
		t.Fatal("question 31 not found")
	}
	if question.Type != QuestionTypeMultipleChoice {
		t.Fatalf("question 31 type = %q, want %q", question.Type, QuestionTypeMultipleChoice)
	}
	if len(question.Options) != 3 {
		t.Fatalf("question 31 options = %d, want 3", len(question.Options))
	}
	if question.Options[0].Value != "规律作息" || question.Options[1].Value != "热爱运动" || question.Options[2].Value != "真诚直接" {
		t.Fatalf("question 31 options = %#v, want grouped option values", question.Options)
	}
}

func TestParseCatalogParsesChoiceOptionsWithoutOptionsLabel(t *testing.T) {
	markdown := strings.TrimSpace(`
# Campus Match 测试问卷

1. 模板标识：campus-match
2. 模板版本：v1.2.1

# 第一部分

## 32. 选出 3-6 个你最稳定、最真实的兴趣
1. 题型：多选题
2. 是否必答：是
3. 选择限制：最少 3 项，最多 6 项
4. 阅读
5. 电影 / 剧集
6. 音乐 / 演出
`)

	catalog, err := ParseCatalog(markdown)
	if err != nil {
		t.Fatalf("ParseCatalog() error = %v", err)
	}

	question, ok := catalog.QuestionByID("32")
	if !ok {
		t.Fatal("question 32 not found")
	}
	if question.Type != QuestionTypeMultipleChoice {
		t.Fatalf("question 32 type = %q, want %q", question.Type, QuestionTypeMultipleChoice)
	}
	if len(question.Options) != 3 {
		t.Fatalf("question 32 options = %d, want %d", len(question.Options), 3)
	}
	if question.Options[0].Value != "阅读" || question.Options[1].Value != "电影 / 剧集" || question.Options[2].Value != "音乐 / 演出" {
		t.Fatalf("question 32 options = %#v, want bare list option values", question.Options)
	}
}

func TestParseCatalogParsesCollectionWindow(t *testing.T) {
	markdown := strings.TrimSpace(`
# Campus Match 测试问卷

1. 模板标识：campus-match
2. 模板版本：v2.0.0
3. 开放开始时间：2026-03-26 09:00
4. 开放结束时间：2026-03-29 21:30
5. 开放时区：Asia/Shanghai

# 第一部分

## 1. 我的性别
1. 题型：单选题
2. 是否必答：是
3. 选项：
   1. 男
   2. 女`)

	catalog, err := ParseCatalog(markdown)
	if err != nil {
		t.Fatalf("ParseCatalog() error = %v", err)
	}
	if catalog.CollectionTimezone != "Asia/Shanghai" {
		t.Fatalf("catalog.CollectionTimezone = %q, want %q", catalog.CollectionTimezone, "Asia/Shanghai")
	}
	if catalog.CollectionStartsAt == nil || catalog.CollectionEndsAt == nil {
		t.Fatalf("collection window = %#v - %#v, want both timestamps", catalog.CollectionStartsAt, catalog.CollectionEndsAt)
	}
	start := catalog.CollectionStartsAt.In(time.FixedZone("CST", 8*3600))
	end := catalog.CollectionEndsAt.In(time.FixedZone("CST", 8*3600))
	if start.Format("2006-01-02 15:04") != "2026-03-26 09:00" {
		t.Fatalf("collection start = %s, want 2026-03-26 09:00", start.Format("2006-01-02 15:04"))
	}
	if end.Format("2006-01-02 15:04") != "2026-03-29 21:30" {
		t.Fatalf("collection end = %s, want 2026-03-29 21:30", end.Format("2006-01-02 15:04"))
	}
}

func TestLoadCatalogParsesThuMarkdownTemplate(t *testing.T) {
	catalog, err := LoadCatalog("thu_questionnaire_template.md")
	if err != nil {
		t.Fatalf("LoadCatalog() error = %v", err)
	}
	if catalog.Key != "thu-date" {
		t.Fatalf("catalog key = %q, want %q", catalog.Key, "thu-date")
	}
	if catalog.Version != "v0.6" {
		t.Fatalf("catalog version = %q, want %q", catalog.Version, "v0.6")
	}
	if catalog.CollectionTimezone != "Asia/Shanghai" {
		t.Fatalf("catalog.CollectionTimezone = %q, want %q", catalog.CollectionTimezone, "Asia/Shanghai")
	}
	if catalog.TotalCount != 52 {
		t.Fatalf("catalog totalCount = %d, want %d", catalog.TotalCount, 52)
	}
	interestQuestion, ok := catalog.QuestionByID("32")
	if !ok {
		t.Fatal("question 32 not found")
	}
	if len(interestQuestion.Options) == 0 {
		t.Fatal("question 32 options are empty, want parsed options")
	}
}
