package questionnaire

import (
	"encoding/json"
	"testing"
	"time"

	"soulink-backend/internal/store"
)

func TestResolveAnswerMigrationPlanBuildsPlanFromConfig(t *testing.T) {
	loader := func(fileName string) (*Catalog, error) {
		switch fileName {
		case "old.md":
			return testCatalog("test-date", "v1.0.0", []Question{{ID: "1", Order: 1, Title: "旧题 1", Type: QuestionTypeSingleChoice, Required: true, SupportsImportance: true, Options: []Option{{Label: "男", Value: "男"}}}}), nil
		case "new.md":
			return testCatalog("test-date", "v1.1.0", []Question{{ID: "10", Order: 1, Title: "新题 10", Type: QuestionTypeSingleChoice, Required: true, SupportsImportance: true, Options: []Option{{Label: "男", Value: "男"}}}}), nil
		default:
			t.Fatalf("unexpected file name %s", fileName)
			return nil, nil
		}
	}

	plan, err := resolveAnswerMigrationPlan(AnswerMigrationConfig{
		SourceCatalogFile:          "old.md",
		SourceQuestionnaireKey:     "test-date",
		SourceQuestionnaireVersion: "v1.0.0",
		TargetCatalogFile:          "new.md",
		TargetQuestionnaireKey:     "test-date",
		TargetQuestionnaireVersion: "v1.1.0",
		Mappings:                   []AnswerMigrationMapping{{SourceQuestionID: "1", TargetQuestionID: "10"}},
	}, loader)
	if err != nil {
		t.Fatalf("resolveAnswerMigrationPlan() error = %v", err)
	}
	if plan.SourceCatalog.Key != "test-date" || plan.SourceCatalog.Version != "v1.0.0" {
		t.Fatalf("unexpected source catalog: %+v", plan.SourceCatalog)
	}
	if plan.TargetCatalog.Key != "test-date" || plan.TargetCatalog.Version != "v1.1.0" {
		t.Fatalf("unexpected target catalog: %+v", plan.TargetCatalog)
	}
	if len(plan.Mappings) != 1 || plan.Mappings[0].TargetQuestionID != "10" {
		t.Fatalf("unexpected mappings: %+v", plan.Mappings)
	}
}

func TestResolveAnswerMigrationPlanRejectsDuplicateTargetMappings(t *testing.T) {
	catalog := testCatalog("test-date", "v1.0.0", []Question{{ID: "1", Order: 1, Title: "题 1", Type: QuestionTypeSingleChoice, Required: true, Options: []Option{{Label: "男", Value: "男"}}}, {ID: "2", Order: 2, Title: "题 2", Type: QuestionTypeSingleChoice, Required: true, Options: []Option{{Label: "男", Value: "男"}}}})
	loader := func(fileName string) (*Catalog, error) {
		if fileName == "old.md" {
			return catalog, nil
		}
		return testCatalog("test-date", "v1.1.0", []Question{{ID: "10", Order: 1, Title: "新题 10", Type: QuestionTypeSingleChoice, Required: true, Options: []Option{{Label: "男", Value: "男"}}}}), nil
	}

	_, err := resolveAnswerMigrationPlan(AnswerMigrationConfig{
		SourceCatalogFile: "old.md",
		TargetCatalogFile: "new.md",
		Mappings: []AnswerMigrationMapping{
			{SourceQuestionID: "1", TargetQuestionID: "10"},
			{SourceQuestionID: "2", TargetQuestionID: "10"},
		},
	}, loader)
	if err == nil {
		t.Fatal("expected duplicate target mapping error")
	}
}

func TestPrepareMigratedAnswersMigratesOnlyMappedAnswers(t *testing.T) {
	now := time.Date(2026, 3, 28, 10, 0, 0, 0, time.UTC)
	targetCatalog := testCatalog("test-date", "v1.1.0", []Question{
		{ID: "10", Order: 1, Title: "新题 10", Type: QuestionTypeSingleChoice, Required: true, SupportsImportance: true, Options: []Option{{Label: "男", Value: "男"}, {Label: "女", Value: "女"}}},
		{ID: "20", Order: 2, Title: "新题 20", Type: QuestionTypeMultipleChoice, Required: false, SupportsImportance: false, Options: []Option{{Label: "运动", Value: "运动"}, {Label: "阅读", Value: "阅读"}}, MinSelections: intPtr(1), MaxSelections: intPtr(2)},
	})
	sourceAnswers := []store.QuestionnaireAnswer{
		{QuestionID: "1", ValueJSON: mustJSON(t, "男"), IsImportant: true, AnsweredAt: now},
		{QuestionID: "2", ValueJSON: mustJSON(t, []string{"运动", "阅读"}), IsImportant: true, AnsweredAt: now.Add(time.Minute)},
		{QuestionID: "3", ValueJSON: mustJSON(t, "不会迁移"), IsImportant: false, AnsweredAt: now.Add(2 * time.Minute)},
	}

	prepared, skipped, err := prepareMigratedAnswers(sourceAnswers, targetCatalog, []AnswerMigrationMapping{
		{SourceQuestionID: "1", TargetQuestionID: "10"},
		{SourceQuestionID: "2", TargetQuestionID: "20"},
	})
	if err != nil {
		t.Fatalf("prepareMigratedAnswers() error = %v", err)
	}
	if skipped != 0 {
		t.Fatalf("prepareMigratedAnswers() skipped = %d, want 0", skipped)
	}
	if len(prepared) != 2 {
		t.Fatalf("prepareMigratedAnswers() len = %d, want 2", len(prepared))
	}
	if prepared[0].QuestionID != "10" || !prepared[0].Importance {
		t.Fatalf("unexpected first migrated answer: %+v", prepared[0])
	}
	if prepared[1].QuestionID != "20" || prepared[1].Importance {
		t.Fatalf("unexpected second migrated answer: %+v", prepared[1])
	}
}

func TestPrepareMigratedAnswersSkipsInvalidTargetAnswer(t *testing.T) {
	targetCatalog := testCatalog("test-date", "v1.1.0", []Question{{ID: "10", Order: 1, Title: "新题 10", Type: QuestionTypeSingleChoice, Required: true, SupportsImportance: true, Options: []Option{{Label: "女", Value: "女"}}}})
	sourceAnswers := []store.QuestionnaireAnswer{{QuestionID: "1", ValueJSON: mustJSON(t, "男"), IsImportant: true, AnsweredAt: time.Now()}}

	prepared, skipped, err := prepareMigratedAnswers(sourceAnswers, targetCatalog, []AnswerMigrationMapping{{SourceQuestionID: "1", TargetQuestionID: "10"}})
	if err != nil {
		t.Fatalf("prepareMigratedAnswers() error = %v", err)
	}
	if len(prepared) != 0 {
		t.Fatalf("prepareMigratedAnswers() len = %d, want 0", len(prepared))
	}
	if skipped != 1 {
		t.Fatalf("prepareMigratedAnswers() skipped = %d, want 1", skipped)
	}
}

func testCatalog(key, version string, questions []Question) *Catalog {
	byID := make(map[string]Question, len(questions))
	requiredCount := 0
	for index, question := range questions {
		if question.Order == 0 {
			question.Order = index + 1
		}
		if question.Required {
			requiredCount++
		}
		byID[question.ID] = question
		questions[index] = question
	}
	return &Catalog{
		ID:            1,
		Key:           key,
		Title:         "测试问卷",
		Version:       version,
		Questions:     questions,
		TotalCount:    len(questions),
		RequiredCount: requiredCount,
		byID:          byID,
	}
}

func mustJSON(t *testing.T, value any) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	return raw
}

func intPtr(value int) *int {
	return &value
}
