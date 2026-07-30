package matchimport

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func boolPtr(value bool) *bool {
	return &value
}

func TestLoadPairsParsesUTF8BOMCSV(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "match_results.csv")
	content := "\uFEFFpool,user_a_id,user_a_nickname,user_a_email,user_b_id,user_b_nickname,user_b_email,score_a_to_b,score_b_to_a,match_quality,grade,asymmetry_warning\n" +
		"MF,101,Alice,alice@example.com,202,Bob,bob@example.com,0.7123,0.7345,0.7123,S,\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	pairs, err := LoadPairs(path)
	if err != nil {
		t.Fatalf("LoadPairs() error = %v", err)
	}
	if len(pairs) != 1 {
		t.Fatalf("len(pairs) = %d, want 1", len(pairs))
	}
	if pairs[0].UserAID != 101 || pairs[0].UserBID != 202 {
		t.Fatalf("unexpected pair IDs: %#v", pairs[0])
	}
	if pairs[0].Grade != "S" {
		t.Fatalf("pairs[0].Grade = %q, want %q", pairs[0].Grade, "S")
	}
	if pairs[0].MatchGroupNo != 1 {
		t.Fatalf("pairs[0].MatchGroupNo = %d, want 1", pairs[0].MatchGroupNo)
	}
}

func TestLoadExplanationsParsesCSV(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "match_explanations.csv")
	content := "\uFEFFmatch_group_no,group_name,explanation_text,explanation_json\n" +
		"1,MF,这是一段解释。," + "\"{\"\"match_summary\"\":\"\"总结一\"\"}\"\n" +
		"2,MM,另一段解释。,\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	records, err := LoadExplanations(path)
	if err != nil {
		t.Fatalf("LoadExplanations() error = %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("len(records) = %d, want 2", len(records))
	}
	if records[1].ExplanationText != "这是一段解释。" {
		t.Fatalf("records[1].ExplanationText = %q", records[1].ExplanationText)
	}
	if string(records[1].ExplanationJSON) != `{"match_summary":"总结一"}` {
		t.Fatalf("records[1].ExplanationJSON = %s", string(records[1].ExplanationJSON))
	}
}

func TestBuildMatchedResultsRejectsDuplicateUser(t *testing.T) {
	plan := &Plan{MatchedMessage: "匹配成功", ReasonTemplate: "模板"}
	_, _, err := buildMatchedResults(plan, []PairRecord{
		{UserAID: 1, UserBID: 2, MatchQuality: 0.72, ScoreAToB: 0.72, ScoreBToA: 0.75, Grade: "S"},
		{UserAID: 2, UserBID: 3, MatchQuality: 0.68, ScoreAToB: 0.70, ScoreBToA: 0.68, Grade: "A"},
	})
	if err == nil || !strings.Contains(err.Error(), "appears in multiple pairs") {
		t.Fatalf("buildMatchedResults() error = %v, want duplicate user error", err)
	}
}

func TestBuildReasonIncludesScoresAndAsymmetryHint(t *testing.T) {
	plan := &Plan{ReasonTemplate: "本轮结果来自离线匹配算法。"}
	reason := buildReason(plan, PairRecord{
		ScoreAToB:        0.66,
		ScoreBToA:        0.82,
		MatchQuality:     0.66,
		Grade:            "A",
		AsymmetryWarning: true,
	}, true)

	for _, want := range []string{
		"本轮结果来自离线匹配算法。",
		"综合匹配度 66.0 分",
		"质量等级 A",
		"双方偏好强度存在一定差异",
	} {
		if !strings.Contains(reason, want) {
			t.Fatalf("buildReason() = %q, want substring %q", reason, want)
		}
	}
}

func TestBuildMatchedResultsUsesExplanationTextWhenPresent(t *testing.T) {
	plan := &Plan{MatchedMessage: "匹配成功", ReasonTemplate: "模板"}
	results, _, err := buildMatchedResults(plan, []PairRecord{
		{
			UserAID:         1,
			UserBID:         2,
			MatchQuality:    0.72,
			ScoreAToB:       0.72,
			ScoreBToA:       0.75,
			Grade:           "S",
			ExplanationText: "这是写回数据库的解释文本。",
			ExplanationJSON: []byte(`{"match_summary":"总结","match_reasons":["原因"]}`),
		},
	})
	if err != nil {
		t.Fatalf("buildMatchedResults() error = %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	for _, result := range results {
		if result.MatchReason != "这是写回数据库的解释文本。" {
			t.Fatalf("result.MatchReason = %q", result.MatchReason)
		}
		if string(result.MatchReasonJSON) != `{"match_summary":"总结","match_reasons":["原因"]}` {
			t.Fatalf("result.MatchReasonJSON = %s", string(result.MatchReasonJSON))
		}
	}
}

func TestResolvePlanDefaultsParticipantFiltersToTrue(t *testing.T) {
	plan, err := resolvePlan("test.json", Config{
		QuestionnaireKey:     "thu-date",
		QuestionnaireVersion: "v0.3",
		CycleKey:             "cycle-1",
		Title:                "Cycle 1",
	})
	if err != nil {
		t.Fatalf("resolvePlan() error = %v", err)
	}
	if !plan.RequireCompletedQuestionnaire {
		t.Fatal("RequireCompletedQuestionnaire = false, want true")
	}
	if !plan.RequireParticipatingInMatch {
		t.Fatal("RequireParticipatingInMatch = false, want true")
	}
}

func TestResolvePlanAllowsParticipantFiltersToBeDisabled(t *testing.T) {
	plan, err := resolvePlan("test.json", Config{
		QuestionnaireKey:              "thu-date",
		QuestionnaireVersion:          "v0.3",
		CycleKey:                      "cycle-1",
		Title:                         "Cycle 1",
		RequireCompletedQuestionnaire: boolPtr(false),
		RequireParticipatingInMatch:   boolPtr(false),
	})
	if err != nil {
		t.Fatalf("resolvePlan() error = %v", err)
	}
	if plan.RequireCompletedQuestionnaire {
		t.Fatal("RequireCompletedQuestionnaire = true, want false")
	}
	if plan.RequireParticipatingInMatch {
		t.Fatal("RequireParticipatingInMatch = true, want false")
	}
}
