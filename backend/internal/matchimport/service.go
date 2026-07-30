package matchimport

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"soulink-backend/internal/match"
	"soulink-backend/internal/store"
)

type Config struct {
	QuestionnaireKey              string `json:"questionnaireKey"`
	QuestionnaireVersion          string `json:"questionnaireVersion"`
	CycleKey                      string `json:"cycleKey"`
	Title                         string `json:"title"`
	RevealStatus                  string `json:"revealStatus,omitempty"`
	RevealAt                      string `json:"revealAt,omitempty"`
	DryRun                        bool   `json:"dryRun,omitempty"`
	MatchedMessage                string `json:"matchedMessage,omitempty"`
	UnmatchedMessage              string `json:"unmatchedMessage,omitempty"`
	ReasonTemplate                string `json:"reasonTemplate,omitempty"`
	MarkUnmatchedUsers            bool   `json:"markUnmatchedUsers,omitempty"`
	RequireCompletedQuestionnaire *bool  `json:"requireCompletedQuestionnaire,omitempty"`
	RequireParticipatingInMatch   *bool  `json:"requireParticipatingInMatch,omitempty"`
}

type Plan struct {
	ConfigPath                    string
	QuestionnaireKey              string
	QuestionnaireVersion          string
	CycleKey                      string
	Title                         string
	RevealStatus                  string
	RevealAt                      time.Time
	DryRun                        bool
	MatchedMessage                string
	UnmatchedMessage              string
	ReasonTemplate                string
	MarkUnmatchedUsers            bool
	RequireCompletedQuestionnaire bool
	RequireParticipatingInMatch   bool
}

type PairRecord struct {
	MatchGroupNo     int
	Pool             string
	UserAID          int64
	UserANickname    string
	UserAEmail       string
	UserBID          int64
	UserBNickname    string
	UserBEmail       string
	ScoreAToB        float64
	ScoreBToA        float64
	MatchQuality     float64
	Grade            string
	AsymmetryWarning bool
	ExplanationText  string
	ExplanationJSON  json.RawMessage
}

type ExplanationRecord struct {
	MatchGroupNo    int
	GroupName       string
	ExplanationText string
	ExplanationJSON json.RawMessage
}

type Stats struct {
	CycleID              int64
	CycleKey             string
	QuestionnaireKey     string
	QuestionnaireVersion string
	MatchedPairCount     int
	MatchedUserCount     int
	UnmatchedUserCount   int
	ResultCount          int
	DryRun               bool
}

type Service struct {
	store        *store.Store
	matchService *match.Service
	nowFunc      func() time.Time
}

func NewService(dataStore *store.Store, matchService *match.Service) *Service {
	return &Service{
		store:        dataStore,
		matchService: matchService,
		nowFunc:      time.Now,
	}
}

func LoadPlan(path string) (*Plan, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("match result import config path is required")
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read match result import config %q: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("decode match result import config %q: %w", path, err)
	}

	return resolvePlan(path, cfg)
}

func resolvePlan(configPath string, cfg Config) (*Plan, error) {
	questionnaireKey := strings.TrimSpace(cfg.QuestionnaireKey)
	questionnaireVersion := strings.TrimSpace(cfg.QuestionnaireVersion)
	cycleKey := strings.TrimSpace(cfg.CycleKey)
	title := strings.TrimSpace(cfg.Title)
	if questionnaireKey == "" || questionnaireVersion == "" || cycleKey == "" || title == "" {
		return nil, fmt.Errorf("questionnaireKey, questionnaireVersion, cycleKey and title are required")
	}

	revealAt, err := parsePlanTime(cfg.RevealAt)
	if err != nil {
		return nil, err
	}
	if revealAt.IsZero() {
		revealAt = time.Now()
	}

	revealStatus := strings.TrimSpace(cfg.RevealStatus)
	if revealStatus == "" {
		revealStatus = "revealed"
	}

	matchedMessage := strings.TrimSpace(cfg.MatchedMessage)
	if matchedMessage == "" {
		matchedMessage = "匹配成功"
	}

	unmatchedMessage := strings.TrimSpace(cfg.UnmatchedMessage)
	if unmatchedMessage == "" {
		unmatchedMessage = "本轮未匹配成功"
	}

	reasonTemplate := strings.TrimSpace(cfg.ReasonTemplate)
	if reasonTemplate == "" {
		reasonTemplate = "本轮结果来自离线匹配算法。"
	}

	requireCompletedQuestionnaire := true
	if cfg.RequireCompletedQuestionnaire != nil {
		requireCompletedQuestionnaire = *cfg.RequireCompletedQuestionnaire
	}
	requireParticipatingInMatch := true
	if cfg.RequireParticipatingInMatch != nil {
		requireParticipatingInMatch = *cfg.RequireParticipatingInMatch
	}

	return &Plan{
		ConfigPath:                    configPath,
		QuestionnaireKey:              questionnaireKey,
		QuestionnaireVersion:          questionnaireVersion,
		CycleKey:                      cycleKey,
		Title:                         title,
		RevealStatus:                  revealStatus,
		RevealAt:                      revealAt,
		DryRun:                        cfg.DryRun,
		MatchedMessage:                matchedMessage,
		UnmatchedMessage:              unmatchedMessage,
		ReasonTemplate:                reasonTemplate,
		MarkUnmatchedUsers:            cfg.MarkUnmatchedUsers,
		RequireCompletedQuestionnaire: requireCompletedQuestionnaire,
		RequireParticipatingInMatch:   requireParticipatingInMatch,
	}, nil
}

func LoadPairs(path string) ([]PairRecord, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("match results csv path is required")
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open match results csv %q: %w", path, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read match results csv %q: %w", path, err)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("match results csv %q is empty", path)
	}

	header := make(map[string]int, len(records[0]))
	for index, column := range records[0] {
		header[cleanHeader(column)] = index
	}
	requiredColumns := []string{
		"user_a_id", "user_b_id", "score_a_to_b", "score_b_to_a", "match_quality",
	}
	for _, column := range requiredColumns {
		if _, ok := header[column]; !ok {
			return nil, fmt.Errorf("match results csv missing required column %q", column)
		}
	}

	pairs := make([]PairRecord, 0, len(records)-1)
	for rowIndex, row := range records[1:] {
		if isEmptyCSVRow(row) {
			continue
		}

		userAID, err := csvInt64(row, header, "user_a_id")
		if err != nil {
			return nil, fmt.Errorf("parse csv row %d user_a_id: %w", rowIndex+2, err)
		}
		userBID, err := csvInt64(row, header, "user_b_id")
		if err != nil {
			return nil, fmt.Errorf("parse csv row %d user_b_id: %w", rowIndex+2, err)
		}
		scoreAToB, err := csvFloat(row, header, "score_a_to_b")
		if err != nil {
			return nil, fmt.Errorf("parse csv row %d score_a_to_b: %w", rowIndex+2, err)
		}
		scoreBToA, err := csvFloat(row, header, "score_b_to_a")
		if err != nil {
			return nil, fmt.Errorf("parse csv row %d score_b_to_a: %w", rowIndex+2, err)
		}
		matchQuality, err := csvFloat(row, header, "match_quality")
		if err != nil {
			return nil, fmt.Errorf("parse csv row %d match_quality: %w", rowIndex+2, err)
		}

		pair := PairRecord{
			MatchGroupNo:     len(pairs) + 1,
			Pool:             csvString(row, header, "pool"),
			UserAID:          userAID,
			UserANickname:    csvString(row, header, "user_a_nickname"),
			UserAEmail:       csvString(row, header, "user_a_email"),
			UserBID:          userBID,
			UserBNickname:    csvString(row, header, "user_b_nickname"),
			UserBEmail:       csvString(row, header, "user_b_email"),
			ScoreAToB:        scoreAToB,
			ScoreBToA:        scoreBToA,
			MatchQuality:     matchQuality,
			Grade:            strings.ToUpper(csvString(row, header, "grade")),
			AsymmetryWarning: csvBool(row, header, "asymmetry_warning"),
		}
		if pair.UserAID == pair.UserBID {
			return nil, fmt.Errorf("csv row %d contains self match for user %d", rowIndex+2, pair.UserAID)
		}
		pairs = append(pairs, pair)
	}

	if len(pairs) == 0 {
		return nil, fmt.Errorf("match results csv %q contains no pair rows", path)
	}
	return pairs, nil
}

func LoadExplanations(path string) (map[int]ExplanationRecord, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open match explanations csv %q: %w", path, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read match explanations csv %q: %w", path, err)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("match explanations csv %q is empty", path)
	}

	header := make(map[string]int, len(records[0]))
	for index, column := range records[0] {
		header[cleanHeader(column)] = index
	}
	if _, ok := header["match_group_no"]; !ok {
		return nil, fmt.Errorf("match explanations csv missing required column %q", "match_group_no")
	}
	if _, hasText := header["explanation_text"]; !hasText {
		if _, hasJSON := header["explanation_json"]; !hasJSON {
			return nil, fmt.Errorf("match explanations csv requires at least one of %q or %q", "explanation_text", "explanation_json")
		}
	}

	explanations := make(map[int]ExplanationRecord, len(records)-1)
	for rowIndex, row := range records[1:] {
		if isEmptyCSVRow(row) {
			continue
		}

		groupNo, err := csvInt(row, header, "match_group_no")
		if err != nil {
			return nil, fmt.Errorf("parse explanation csv row %d match_group_no: %w", rowIndex+2, err)
		}
		if groupNo <= 0 {
			return nil, fmt.Errorf("parse explanation csv row %d match_group_no: must be positive", rowIndex+2)
		}
		if _, exists := explanations[groupNo]; exists {
			return nil, fmt.Errorf("match explanations csv contains duplicate match_group_no %d", groupNo)
		}

		explanationJSON, err := csvJSON(row, header, "explanation_json")
		if err != nil {
			return nil, fmt.Errorf("parse explanation csv row %d explanation_json: %w", rowIndex+2, err)
		}

		explanations[groupNo] = ExplanationRecord{
			MatchGroupNo:    groupNo,
			GroupName:       csvString(row, header, "group_name"),
			ExplanationText: csvString(row, header, "explanation_text"),
			ExplanationJSON: explanationJSON,
		}
	}

	return explanations, nil
}

func (s *Service) Run(ctx context.Context, plan *Plan, csvPath string, explanationsPath string) (Stats, error) {
	if s == nil || s.store == nil || s.matchService == nil {
		return Stats{}, fmt.Errorf("match import service is not ready")
	}
	if plan == nil {
		return Stats{}, fmt.Errorf("match import plan is not ready")
	}

	pairs, err := LoadPairs(csvPath)
	if err != nil {
		return Stats{}, err
	}
	explanations, err := LoadExplanations(explanationsPath)
	if err != nil {
		return Stats{}, err
	}
	applyExplanations(pairs, explanations)

	results, matchedUserIDs, err := buildMatchedResults(plan, pairs)
	if err != nil {
		return Stats{}, err
	}

	unmatchedCount := 0
	if plan.MarkUnmatchedUsers {
		participants, err := s.store.ListQuestionnaireUsers(ctx, nil, plan.QuestionnaireKey, plan.QuestionnaireVersion, store.QuestionnaireUserFilter{
			RequireCompletedQuestionnaire: plan.RequireCompletedQuestionnaire,
			RequireParticipatingInMatch:   plan.RequireParticipatingInMatch,
		})
		if err != nil {
			return Stats{}, fmt.Errorf("list questionnaire users for unmatched completion: %w", err)
		}
		if len(participants) == 0 {
			return Stats{}, fmt.Errorf("no questionnaire users found for %s@%s", plan.QuestionnaireKey, plan.QuestionnaireVersion)
		}

		participantSet := make(map[int64]struct{}, len(participants))
		for _, participant := range participants {
			participantSet[participant.ID] = struct{}{}
		}
		for userID := range matchedUserIDs {
			if _, ok := participantSet[userID]; !ok {
				return Stats{}, fmt.Errorf("matched user %d is not an eligible participant of %s@%s", userID, plan.QuestionnaireKey, plan.QuestionnaireVersion)
			}
		}

		for _, participant := range participants {
			if _, ok := matchedUserIDs[participant.ID]; ok {
				continue
			}
			results = append(results, match.RecordUserResultInput{
				UserID:       participant.ID,
				ResultStatus: match.ResultStatusFailed,
				Message:      plan.UnmatchedMessage,
			})
			unmatchedCount++
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].UserID < results[j].UserID
	})

	stats := Stats{
		CycleKey:             plan.CycleKey,
		QuestionnaireKey:     plan.QuestionnaireKey,
		QuestionnaireVersion: plan.QuestionnaireVersion,
		MatchedPairCount:     len(pairs),
		MatchedUserCount:     len(matchedUserIDs),
		UnmatchedUserCount:   unmatchedCount,
		ResultCount:          len(results),
		DryRun:               plan.DryRun,
	}

	if plan.DryRun {
		return stats, nil
	}

	cycle, err := s.matchService.RecordCycleResults(ctx, match.RecordCycleInput{
		CycleKey:             plan.CycleKey,
		Title:                plan.Title,
		QuestionnaireKey:     plan.QuestionnaireKey,
		QuestionnaireVersion: plan.QuestionnaireVersion,
		RevealStatus:         plan.RevealStatus,
		RevealAt:             plan.RevealAt,
		Results:              results,
	})
	if err != nil {
		return Stats{}, fmt.Errorf("record imported match results: %w", err)
	}
	stats.CycleID = cycle.ID
	return stats, nil
}

func applyExplanations(pairs []PairRecord, explanations map[int]ExplanationRecord) {
	if len(pairs) == 0 || len(explanations) == 0 {
		return
	}
	for index := range pairs {
		record, ok := explanations[pairs[index].MatchGroupNo]
		if !ok {
			continue
		}
		pairs[index].ExplanationText = strings.TrimSpace(record.ExplanationText)
		pairs[index].ExplanationJSON = append(json.RawMessage(nil), record.ExplanationJSON...)
	}
}

func buildMatchedResults(plan *Plan, pairs []PairRecord) ([]match.RecordUserResultInput, map[int64]struct{}, error) {
	usedUserIDs := make(map[int64]struct{}, len(pairs)*2)
	results := make([]match.RecordUserResultInput, 0, len(pairs)*2)

	for _, pair := range pairs {
		if _, exists := usedUserIDs[pair.UserAID]; exists {
			return nil, nil, fmt.Errorf("user %d appears in multiple pairs", pair.UserAID)
		}
		if _, exists := usedUserIDs[pair.UserBID]; exists {
			return nil, nil, fmt.Errorf("user %d appears in multiple pairs", pair.UserBID)
		}
		usedUserIDs[pair.UserAID] = struct{}{}
		usedUserIDs[pair.UserBID] = struct{}{}

		leftPartnerID := pair.UserBID
		rightPartnerID := pair.UserAID
		leftScore := normalizeCSVScore(pair.MatchQuality)
		rightScore := normalizeCSVScore(pair.MatchQuality)

		reason := buildReason(plan, pair, true)
		if explanation := strings.TrimSpace(pair.ExplanationText); explanation != "" {
			reason = explanation
		}
		reasonJSON := append(json.RawMessage(nil), pair.ExplanationJSON...)

		results = append(results,
			match.RecordUserResultInput{
				UserID:          pair.UserAID,
				ResultStatus:    match.ResultStatusMatched,
				PartnerUserID:   &leftPartnerID,
				MatchScore:      &leftScore,
				MatchReason:     reason,
				MatchReasonJSON: reasonJSON,
				Message:         plan.MatchedMessage,
			},
			match.RecordUserResultInput{
				UserID:          pair.UserBID,
				ResultStatus:    match.ResultStatusMatched,
				PartnerUserID:   &rightPartnerID,
				MatchScore:      &rightScore,
				MatchReason:     reason,
				MatchReasonJSON: reasonJSON,
				Message:         plan.MatchedMessage,
			},
		)
	}

	return results, usedUserIDs, nil
}

func buildReason(plan *Plan, pair PairRecord, leftPerspective bool) string {
	quality := normalizeCSVScore(pair.MatchQuality)
	selfScore := normalizeCSVScore(pair.ScoreAToB)
	partnerScore := normalizeCSVScore(pair.ScoreBToA)
	if !leftPerspective {
		selfScore, partnerScore = partnerScore, selfScore
	}

	grade := normalizedGrade(pair.Grade, quality)
	parts := make([]string, 0, 4)
	if template := strings.TrimSpace(plan.ReasonTemplate); template != "" {
		parts = append(parts, template)
	}
	parts = append(parts, fmt.Sprintf(
		"综合匹配度 %.1f 分（你对对方 %.1f 分，对方对你 %.1f 分），质量等级 %s。",
		quality,
		selfScore,
		partnerScore,
		grade,
	))
	parts = append(parts, gradeHint(grade))
	if pair.AsymmetryWarning {
		parts = append(parts, "双方偏好强度存在一定差异，建议在聊天初期多观察对方反馈，避免推进过快。")
	}
	return strings.Join(parts, " ")
}

func gradeHint(grade string) string {
	switch grade {
	case "S":
		return "整体契合度较高，建议尽快安排一次轻松但有深度的交流，主动分享近期最投入的话题。"
	case "A":
		return "你们在多个关键维度上较为契合，建议先从共同兴趣、校园生活和日常节奏切入，逐步建立熟悉感。"
	default:
		return "这是一组有潜力的匹配，建议循序渐进地了解彼此，先确认双方的舒适区和交流节奏。"
	}
}

func normalizedGrade(raw string, quality float64) string {
	grade := strings.ToUpper(strings.TrimSpace(raw))
	if grade != "" {
		return grade
	}
	switch {
	case quality >= 70:
		return "S"
	case quality >= 55:
		return "A"
	default:
		return "B"
	}
}

func normalizeCSVScore(value float64) float64 {
	if value <= 1 {
		value *= 100
	}
	if value < 0 {
		value = 0
	}
	if value > 100 {
		value = 100
	}
	return math.Round(value*10) / 10
}

func parsePlanTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed, nil
	}

	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.Time{}, fmt.Errorf("load timezone: %w", err)
	}

	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
	}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, value, location); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported revealAt format %q", value)
}

func cleanHeader(value string) string {
	return strings.TrimSpace(strings.TrimPrefix(value, "\uFEFF"))
}

func isEmptyCSVRow(row []string) bool {
	for _, value := range row {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}

func csvString(row []string, header map[string]int, key string) string {
	index, ok := header[key]
	if !ok || index >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[index])
}

func csvInt64(row []string, header map[string]int, key string) (int64, error) {
	value := csvString(row, header, key)
	if value == "" {
		return 0, fmt.Errorf("empty value")
	}
	return strconv.ParseInt(value, 10, 64)
}

func csvInt(row []string, header map[string]int, key string) (int, error) {
	value := csvString(row, header, key)
	if value == "" {
		return 0, fmt.Errorf("empty value")
	}
	return strconv.Atoi(value)
}

func csvFloat(row []string, header map[string]int, key string) (float64, error) {
	value := csvString(row, header, key)
	if value == "" {
		return 0, fmt.Errorf("empty value")
	}
	return strconv.ParseFloat(value, 64)
}

func csvJSON(row []string, header map[string]int, key string) (json.RawMessage, error) {
	value := csvString(row, header, key)
	if value == "" {
		return nil, nil
	}
	var payload any
	if err := json.Unmarshal([]byte(value), &payload); err != nil {
		return nil, err
	}
	normalized, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(normalized), nil
}

func csvBool(row []string, header map[string]int, key string) bool {
	switch strings.ToLower(csvString(row, header, key)) {
	case "1", "true", "yes", "y":
		return true
	default:
		return false
	}
}
