package connection

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"soulink-backend/internal/ai"
)

type interpreterGenerator struct {
	response string
	err      error
}

func (generator interpreterGenerator) Generate(
	_ context.Context,
	prompt ai.Prompt,
	_ any,
	output any,
) (ai.Metadata, error) {
	if generator.err != nil {
		return ai.Metadata{Source: "model", Model: "test-model", PromptVersion: prompt.Version}, generator.err
	}
	return ai.Metadata{Source: "model", Model: "test-model", PromptVersion: prompt.Version}, json.Unmarshal([]byte(generator.response), output)
}

func TestInterpreterUnderstandsCampusMealIntent(t *testing.T) {
	t.Parallel()

	result := NewInterpreter().Interpret("今晚想在紫荆找个饭搭子，轻松聊聊 AI")

	if result.Category != CategoryMeal {
		t.Fatalf("category = %q, want %q", result.Category, CategoryMeal)
	}
	if result.TimeWindow != "今晚" {
		t.Fatalf("timeWindow = %q, want 今晚", result.TimeWindow)
	}
	if result.Location != "紫荆" {
		t.Fatalf("location = %q, want 紫荆", result.Location)
	}
	if result.Visibility != VisibilityFeed {
		t.Fatalf("visibility = %q, want %q", result.Visibility, VisibilityFeed)
	}
	if len(result.Questions) != 0 {
		t.Fatalf("questions = %#v, want no clarification questions", result.Questions)
	}
	if result.Confidence < 0.9 {
		t.Fatalf("confidence = %.2f, want at least 0.9", result.Confidence)
	}
}

func TestInterpreterKeepsRomanticIntentPrivate(t *testing.T) {
	t.Parallel()

	result := NewInterpreter().Interpret("希望认真认识一个耐心、有好奇心的人，想发展长期关系")

	if result.Category != CategoryRomantic {
		t.Fatalf("category = %q, want %q", result.Category, CategoryRomantic)
	}
	if result.Visibility != VisibilityCandidates {
		t.Fatalf("visibility = %q, want %q", result.Visibility, VisibilityCandidates)
	}
	if !strings.Contains(strings.Join(result.Inferences, " "), "仅向推荐候选人可见") {
		t.Fatalf("inferences = %#v, want romantic visibility explanation", result.Inferences)
	}
}

func TestInterpreterAsksOnlyMaterialProjectQuestions(t *testing.T) {
	t.Parallel()

	result := NewInterpreter().Interpret("想找一个 CTO 一起做校园 AI 产品")

	if result.Category != CategoryCofounder {
		t.Fatalf("category = %q, want %q", result.Category, CategoryCofounder)
	}
	if len(result.Questions) == 0 || len(result.Questions) > 3 {
		t.Fatalf("questions count = %d, want between 1 and 3", len(result.Questions))
	}

	foundSkills := false
	for _, question := range result.Questions {
		if question.ID == "skills" {
			foundSkills = true
		}
	}
	if !foundSkills {
		t.Fatalf("questions = %#v, want a skills clarification", result.Questions)
	}
}

func TestInterpreterBuildsProjectMemoFromConcreteSignals(t *testing.T) {
	t.Parallel()

	result := NewInterpreter().Interpret(
		"已有第一版 AI 产品，想找 React 前端 CTO，每周投入 10 小时，可以讨论联合创始人股权",
	)

	if result.ProjectMemo == nil {
		t.Fatal("project memo = nil")
	}
	if result.ProjectMemo.Stage != "已有可用原型" {
		t.Fatalf("project stage = %q", result.ProjectMemo.Stage)
	}
	if result.ProjectMemo.WeeklyTime == "" {
		t.Fatal("weekly time is empty")
	}
	if len(result.ProjectMemo.RequiredSkills) == 0 {
		t.Fatal("required skills are empty")
	}
	if result.ProjectMemo.Compensation == "" {
		t.Fatal("compensation is empty")
	}
}

func TestInterpreterFlagsPublicContactDetails(t *testing.T) {
	t.Parallel()

	result := NewInterpreter().Interpret("找跑步搭子，可以直接加微信号 13800138000")

	if result.Safety.Level != "review" {
		t.Fatalf("safety level = %q, want review", result.Safety.Level)
	}
	if !strings.Contains(result.Safety.Notice, "联系方式") {
		t.Fatalf("safety notice = %q, want contact warning", result.Safety.Notice)
	}
}

func TestInterpreterUsesValidatedModelOutputWithoutGrantingVisibility(t *testing.T) {
	t.Parallel()

	interpreter := NewInterpreterWithGenerator(interpreterGenerator{response: `{
		"category":"romantic",
		"title":"认真认识一位新朋友",
		"goal":"先在双方愿意的情况下轻松交流",
		"timeWindow":"本周",
		"location":"图书馆",
		"constraints":["低压力"],
		"desiredNextStep":"双方同意后再开聊",
		"questions":[],
		"inferences":["用户明确表达了认真认识的目标"],
		"confidence":0.88
	}`})
	result := interpreter.Interpret("本周想在图书馆认真认识一个人")

	if result.Source != "model" || result.Model != "test-model" {
		t.Fatalf("model metadata = %#v", result)
	}
	if result.Visibility != VisibilityCandidates {
		t.Fatalf("visibility = %q, want candidates enforced by service", result.Visibility)
	}
	if !result.Editable || result.PromptVersion != "intent-clarification-v2" {
		t.Fatalf("inference controls = %#v", result)
	}
}

func TestInterpreterFallsBackWhenModelFailsOrReturnsInvalidCategory(t *testing.T) {
	t.Parallel()

	for _, generator := range []interpreterGenerator{
		{err: errors.New("provider down")},
		{response: `{"category":"bypass","title":"x","goal":"x","timeWindow":"","location":"","constraints":[],"desiredNextStep":"x","questions":[],"inferences":[],"confidence":0.9}`},
	} {
		result := NewInterpreterWithGenerator(generator).Interpret("今晚在紫荆找饭搭子")
		if result.Source != "deterministic" || result.Category != CategoryMeal {
			t.Fatalf("fallback result = %#v", result)
		}
		if result.FallbackReason == "" {
			t.Fatal("fallback reason is empty")
		}
	}
}
