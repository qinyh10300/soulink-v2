package chat

import (
	"testing"
)

func TestDeterministicMessageSafetyClassifiesRiskLayers(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		level    string
		category string
	}{
		{name: "clear", content: "明天晚上一起去紫操跑步吗？", level: "clear"},
		{name: "contact", content: "加微信 soulink_2026 再说", level: "caution", category: "contact_inducement"},
		{name: "spam", content: "兼职返利稳赚，访问 https://example.com", level: "caution", category: "spam"},
		{name: "threat", content: "你给我等着，我明天去堵你", level: "high", category: "threat"},
		{name: "privacy", content: "我会开盒并公布你的地址", level: "high", category: "privacy"},
		{name: "coercion", content: "你必须单独见我，不许告诉别人", level: "high", category: "coercion"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := deterministicMessageSafety(test.content)
			if result.Level != test.level {
				t.Fatalf("level = %q, want %q (%#v)", result.Level, test.level, result)
			}
			if test.category != "" && !containsMessageSafetyValue(result.Categories, test.category) {
				t.Fatalf("categories = %#v, want %q", result.Categories, test.category)
			}
			if result.RequiresConfirmation != (test.level == "high") {
				t.Fatalf("requires confirmation = %v for %q", result.RequiresConfirmation, test.level)
			}
		})
	}
}

func TestMessageSafetyModelCannotLowerRuleRisk(t *testing.T) {
	rule := deterministicMessageSafety("我会开盒并公布你的地址")
	merged := mergeMessageSafety(rule, messageSafetyModelOutput{
		Level: "clear", Confidence: 0.99,
	})
	if merged.Level != "high" || !merged.RequiresConfirmation {
		t.Fatalf("model lowered hard rule: %#v", merged)
	}
}

func TestMessageSafetyModelCanEscalateAndOnlyUsesKnownCategories(t *testing.T) {
	rule := deterministicMessageSafety("今晚聊聊")
	merged := mergeMessageSafety(rule, messageSafetyModelOutput{
		Level: "high", Categories: []string{"harassment"},
		Reasons: []string{"存在持续骚扰语境"}, Confidence: 0.88,
	})
	if merged.Level != "high" || !containsMessageSafetyValue(merged.Categories, "harassment") {
		t.Fatalf("model escalation missing: %#v", merged)
	}
	if validMessageSafetyModelOutput(messageSafetyModelOutput{
		Level: "high", Categories: []string{"auto_send"}, Confidence: 1,
	}) {
		t.Fatal("unknown model category should be rejected")
	}
}

func TestRepeatedMessageRunIsDetectedWithoutRegexBackreferences(t *testing.T) {
	if !hasRepeatedMessageRun("加群啦啦啦啦啦啦啦啦", 8) {
		t.Fatal("expected repeated run")
	}
	if hasRepeatedMessageRun("正常的校园对话", 8) {
		t.Fatal("normal text was marked repeated")
	}
}
