package questionnaire

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/store"
)

const (
	portraitStatusPending    = "pending"
	portraitStatusGenerating = "generating"
	portraitStatusReady      = "ready"
	portraitStatusFailed     = "failed"
	soulAgentQuotaMessage    = "该功能已达到限额"
)

type PortraitResult struct {
	SoulSignature   string         `json:"soulSignature"`
	PersonalProfile string         `json:"personalProfile"`
	IdealResonance  string         `json:"idealResonance"`
	Traits          map[string]any `json:"traits"`
}

type soulAgentCLIResponse struct {
	Traits map[string]any `json:"traits"`
	Report struct {
		SoulSignature   string `json:"soul_signature"`
		PersonalProfile string `json:"personal_profile"`
		IdealResonance  string `json:"ideal_resonance"`
	} `json:"report"`
}

type soulAgentPayload struct {
	Answers              map[string]any `json:"answers"`
	ImportantQuestionIDs []int          `json:"important_question_ids,omitempty"`
}

var soulAgentRequiredQuestionIDs = []int{
	3, 5, 11, 12, 13, 14, 16, 17, 18, 19, 20, 22, 23, 24, 25, 26, 27, 30, 32,
	33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52,
}

func (s *Service) GetPortrait(ctx context.Context, userID int64) (PortraitResult, error) {
	if err := s.ensureReady(); err != nil {
		return PortraitResult{}, err
	}
	if !s.isSoulAgentEnabled() {
		return PortraitResult{}, s.soulAgentDisabledError()
	}

	state, err := s.store.GetQuestionnaireState(ctx, nil, userID, s.catalog.Key, s.catalog.Version)
	if err != nil {
		if err == store.ErrNotFound {
			return PortraitResult{}, httpapi.NewError(http.StatusForbidden, httpapi.CodeForbidden, "complete questionnaire before viewing portrait")
		}
		return PortraitResult{}, httpapi.InternalError(err)
	}
	if !state.HasCompletedQuestionnaire {
		return PortraitResult{}, httpapi.NewError(http.StatusForbidden, httpapi.CodeForbidden, "complete questionnaire before viewing portrait")
	}
	if !s.shouldGeneratePortrait(state.AnsweredCount, state.HasCompletedQuestionnaire) {
		return PortraitResult{}, httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, "portrait is unavailable until all questions are submitted")
	}

	portrait, err := s.store.GetQuestionnairePortrait(ctx, nil, userID, s.catalog.Key, s.catalog.Version)
	if err != nil {
		if err == store.ErrNotFound {
			s.ensurePortraitGeneration(ctx, userID, state.AnsweredCount, state.HasCompletedQuestionnaire, s.currentTime(), false)
			return PortraitResult{}, httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, "portrait is still generating")
		}
		return PortraitResult{}, httpapi.InternalError(err)
	}

	switch portrait.Status {
	case portraitStatusReady:
		var traits map[string]any
		if len(portrait.TraitsJSON) > 0 {
			if err := json.Unmarshal(portrait.TraitsJSON, &traits); err != nil {
				return PortraitResult{}, httpapi.InternalError(fmt.Errorf("decode portrait traits: %w", err))
			}
		}
		return PortraitResult{
			SoulSignature:   portrait.SoulSignature,
			PersonalProfile: portrait.PersonalProfile,
			IdealResonance:  portrait.IdealResonance,
			Traits:          traits,
		}, nil
	case portraitStatusPending, portraitStatusGenerating:
		return PortraitResult{}, httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, "portrait is still generating")
	case portraitStatusFailed:
		s.ensurePortraitGeneration(ctx, userID, state.AnsweredCount, state.HasCompletedQuestionnaire, s.currentTime(), true)
		return PortraitResult{}, httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, "portrait is still generating")
	default:
		return PortraitResult{}, httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, "portrait is still generating")
	}
}

func buildSoulAgentPayload(records []store.QuestionnaireAnswer) (soulAgentPayload, error) {
	answers := make(map[string]any, len(records))
	important := make([]int, 0)
	seen := make(map[int]struct{}, len(records))

	for _, record := range records {
		questionNumber, ok := parseQuestionNumber(record.QuestionID)
		if !ok {
			continue
		}

		var value any
		if err := json.Unmarshal(record.ValueJSON, &value); err != nil {
			return soulAgentPayload{}, httpapi.InternalError(fmt.Errorf("decode portrait answer %s: %w", record.QuestionID, err))
		}

		answers[fmt.Sprintf("q%d", questionNumber)] = value
		if record.IsImportant {
			if _, exists := seen[questionNumber]; !exists {
				important = append(important, questionNumber)
				seen[questionNumber] = struct{}{}
			}
		}
	}

	missing := make([]string, 0)
	for _, questionID := range soulAgentRequiredQuestionIDs {
		key := fmt.Sprintf("q%d", questionID)
		if _, ok := answers[key]; !ok {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return soulAgentPayload{}, httpapi.NewError(
			http.StatusConflict,
			httpapi.CodeConflict,
			"portrait is unavailable until all required answers are submitted",
		)
	}

	slices.Sort(important)
	return soulAgentPayload{
		Answers:              answers,
		ImportantQuestionIDs: important,
	}, nil
}

func signatureForSoulAgentPayload(payload soulAgentPayload) (string, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", httpapi.InternalError(err)
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

func (s *Service) currentPortraitAnswerSignature(ctx context.Context, userID int64) (string, error) {
	answers, err := s.store.ListQuestionnaireAnswers(ctx, nil, userID, s.catalog.Key, s.catalog.Version)
	if err != nil {
		return "", httpapi.InternalError(err)
	}

	payload, err := buildSoulAgentPayload(answers)
	if err != nil {
		return "", err
	}

	return signatureForSoulAgentPayload(payload)
}

func runSoulAgentPortrait(ctx context.Context, payload soulAgentPayload) (soulAgentCLIResponse, error) {
	projectRoot := resolveSoulAgentRoot()
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return soulAgentCLIResponse{}, httpapi.InternalError(err)
	}

	command := exec.CommandContext(
		ctx,
		resolveSoulAgentPython(),
		"-m",
		"soulagent.cli",
		"portrait",
		"--input",
		"-",
		"--skill",
		"tsinghua_soulink",
	)
	command.Dir = projectRoot
	command.Stdin = bytes.NewReader(rawPayload)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	if err := command.Run(); err != nil {
		detail := strings.TrimSpace(stderr.String())
		if detail == "" {
			detail = err.Error()
		}
		return soulAgentCLIResponse{}, httpapi.InternalError(fmt.Errorf("failed to generate portrait: %s: %w", detail, err))
	}

	var result soulAgentCLIResponse
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		raw := strings.TrimSpace(stdout.String())
		if len(raw) > 400 {
			raw = raw[:400] + "..."
		}
		return soulAgentCLIResponse{}, httpapi.InternalError(fmt.Errorf("decode soulagent portrait output: %w; stdout=%q", err, raw))
	}
	return result, nil
}

func resolveSoulAgentRoot() string {
	if root := strings.TrimSpace(os.Getenv("SOULAGENT_ROOT")); root != "" {
		return root
	}
	return filepath.Clean(filepath.Join(".", "soulagent"))
}

func resolveSoulAgentPython() string {
	if python := strings.TrimSpace(os.Getenv("SOULAGENT_PYTHON")); python != "" {
		return python
	}
	return "python3"
}

func parseQuestionNumber(questionID string) (int, bool) {
	normalized := strings.TrimSpace(strings.ToLower(questionID))
	normalized = strings.TrimPrefix(normalized, "q")
	number, err := strconv.Atoi(normalized)
	if err != nil {
		return 0, false
	}
	return number, true
}

func (s *Service) shouldGeneratePortrait(answeredCount int, hasCompleted bool) bool {
	return s != nil && s.catalog != nil && s.soulAgentEnabled && hasCompleted && answeredCount == s.catalog.TotalCount
}

func (s *Service) isSoulAgentEnabled() bool {
	return s != nil && s.soulAgentEnabled
}

func (s *Service) soulAgentDisabledError() error {
	return httpapi.NewError(http.StatusForbidden, httpapi.CodeForbidden, soulAgentQuotaMessage)
}

func (s *Service) ensurePortraitGeneration(ctx context.Context, userID int64, answeredCount int, hasCompleted bool, now time.Time, force bool) {
	if !s.shouldGeneratePortrait(answeredCount, hasCompleted) {
		return
	}

	answerSignature, err := s.currentPortraitAnswerSignature(ctx, userID)
	if err != nil {
		return
	}

	if !force {
		shouldGenerate, checkErr := s.shouldRefreshPortrait(ctx, userID, answerSignature)
		if checkErr != nil || !shouldGenerate {
			return
		}
	}

	if err := s.enqueuePortraitGeneration(ctx, userID, now, answerSignature); err == nil {
		s.startPortraitGeneration(userID)
	}
}

func (s *Service) shouldRefreshPortrait(ctx context.Context, userID int64, answerSignature string) (bool, error) {
	portrait, err := s.store.GetQuestionnairePortrait(ctx, nil, userID, s.catalog.Key, s.catalog.Version)
	if err != nil {
		if err == store.ErrNotFound {
			return true, nil
		}
		return false, httpapi.InternalError(err)
	}

	return portrait.AnswerSignature != answerSignature, nil
}

func (s *Service) enqueuePortraitGeneration(ctx context.Context, userID int64, now time.Time, answerSignature string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	if !s.isSoulAgentEnabled() {
		return s.soulAgentDisabledError()
	}

	return s.store.WithTx(ctx, func(tx *sql.Tx) error {
		state, err := s.store.GetQuestionnaireState(ctx, tx, userID, s.catalog.Key, s.catalog.Version)
		if err != nil {
			if err == store.ErrNotFound {
				return httpapi.NewError(http.StatusForbidden, httpapi.CodeForbidden, "complete questionnaire before viewing portrait")
			}
			return httpapi.InternalError(err)
		}
		if !s.shouldGeneratePortrait(state.AnsweredCount, state.HasCompletedQuestionnaire) {
			return httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, "portrait is unavailable until all questions are submitted")
		}

		_, err = s.store.UpsertQuestionnairePortrait(ctx, tx, store.UpsertQuestionnairePortraitParams{
			UserID:               userID,
			QuestionnaireKey:     s.catalog.Key,
			QuestionnaireVersion: s.catalog.Version,
			Status:               portraitStatusPending,
			AnswerSignature:      answerSignature,
			SoulSignature:        "",
			PersonalProfile:      "",
			IdealResonance:       "",
			TraitsJSON:           nil,
			ErrorMessage:         "",
			GeneratedAt:          nil,
			CreatedAt:            now,
			UpdatedAt:            now,
		})
		if err != nil {
			return httpapi.InternalError(err)
		}
		return nil
	})
}

func (s *Service) startPortraitGeneration(userID int64) {
	if s == nil {
		return
	}
	go s.generatePortraitInBackground(userID)
}

func (s *Service) generatePortraitInBackground(userID int64) {
	ctx := context.Background()
	now := s.currentTime()

	if err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		portrait, err := s.store.GetQuestionnairePortrait(ctx, tx, userID, s.catalog.Key, s.catalog.Version)
		if err != nil {
			if err == store.ErrNotFound {
				return store.ErrNotFound
			}
			return err
		}
		if portrait.Status == portraitStatusGenerating {
			return nil
		}
		if _, err := s.store.UpsertQuestionnairePortrait(ctx, tx, store.UpsertQuestionnairePortraitParams{
			UserID:               userID,
			QuestionnaireKey:     s.catalog.Key,
			QuestionnaireVersion: s.catalog.Version,
			Status:               portraitStatusGenerating,
			AnswerSignature:      portrait.AnswerSignature,
			SoulSignature:        portrait.SoulSignature,
			PersonalProfile:      portrait.PersonalProfile,
			IdealResonance:       portrait.IdealResonance,
			TraitsJSON:           portrait.TraitsJSON,
			ErrorMessage:         "",
			GeneratedAt:          nil,
			CreatedAt:            portrait.CreatedAt,
			UpdatedAt:            now,
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return
	}

	result, err := s.buildPortrait(ctx, userID)
	if err != nil {
		_ = s.markPortraitFailed(ctx, userID, err)
		return
	}
	answerSignature, err := s.currentPortraitAnswerSignature(ctx, userID)
	if err != nil {
		_ = s.markPortraitFailed(ctx, userID, err)
		return
	}
	_ = s.markPortraitReady(ctx, userID, answerSignature, result)
}

func (s *Service) buildPortrait(ctx context.Context, userID int64) (PortraitResult, error) {
	answers, err := s.store.ListQuestionnaireAnswers(ctx, nil, userID, s.catalog.Key, s.catalog.Version)
	if err != nil {
		return PortraitResult{}, httpapi.InternalError(err)
	}

	payload, err := buildSoulAgentPayload(answers)
	if err != nil {
		return PortraitResult{}, err
	}

	result, err := runSoulAgentPortrait(ctx, payload)
	if err != nil {
		return PortraitResult{}, err
	}

	return PortraitResult{
		SoulSignature:   result.Report.SoulSignature,
		PersonalProfile: result.Report.PersonalProfile,
		IdealResonance:  result.Report.IdealResonance,
		Traits:          result.Traits,
	}, nil
}

func (s *Service) markPortraitReady(ctx context.Context, userID int64, answerSignature string, result PortraitResult) error {
	rawTraits, err := json.Marshal(result.Traits)
	if err != nil {
		return httpapi.InternalError(err)
	}
	now := s.currentTime()

	return s.store.WithTx(ctx, func(tx *sql.Tx) error {
		existing, err := s.store.GetQuestionnairePortrait(ctx, tx, userID, s.catalog.Key, s.catalog.Version)
		if err != nil {
			return httpapi.InternalError(err)
		}
		_, err = s.store.UpsertQuestionnairePortrait(ctx, tx, store.UpsertQuestionnairePortraitParams{
			UserID:               userID,
			QuestionnaireKey:     s.catalog.Key,
			QuestionnaireVersion: s.catalog.Version,
			Status:               portraitStatusReady,
			AnswerSignature:      answerSignature,
			SoulSignature:        result.SoulSignature,
			PersonalProfile:      result.PersonalProfile,
			IdealResonance:       result.IdealResonance,
			TraitsJSON:           rawTraits,
			ErrorMessage:         "",
			GeneratedAt:          &now,
			CreatedAt:            existing.CreatedAt,
			UpdatedAt:            now,
		})
		if err != nil {
			return httpapi.InternalError(err)
		}
		return nil
	})
}

func (s *Service) markPortraitFailed(ctx context.Context, userID int64, cause error) error {
	now := s.currentTime()
	message := "failed to generate portrait"
	if cause != nil {
		message = cause.Error()
	}

	return s.store.WithTx(ctx, func(tx *sql.Tx) error {
		portrait, err := s.store.GetQuestionnairePortrait(ctx, tx, userID, s.catalog.Key, s.catalog.Version)
		if err != nil {
			if err == store.ErrNotFound {
				_, upsertErr := s.store.UpsertQuestionnairePortrait(ctx, tx, store.UpsertQuestionnairePortraitParams{
					UserID:               userID,
					QuestionnaireKey:     s.catalog.Key,
					QuestionnaireVersion: s.catalog.Version,
					Status:               portraitStatusFailed,
					AnswerSignature:      "",
					SoulSignature:        "",
					PersonalProfile:      "",
					IdealResonance:       "",
					TraitsJSON:           nil,
					ErrorMessage:         message,
					GeneratedAt:          nil,
					CreatedAt:            now,
					UpdatedAt:            now,
				})
				if upsertErr != nil {
					return httpapi.InternalError(upsertErr)
				}
				return nil
			}
			return httpapi.InternalError(err)
		}

		_, err = s.store.UpsertQuestionnairePortrait(ctx, tx, store.UpsertQuestionnairePortraitParams{
			UserID:               userID,
			QuestionnaireKey:     s.catalog.Key,
			QuestionnaireVersion: s.catalog.Version,
			Status:               portraitStatusFailed,
			AnswerSignature:      portrait.AnswerSignature,
			SoulSignature:        "",
			PersonalProfile:      "",
			IdealResonance:       "",
			TraitsJSON:           nil,
			ErrorMessage:         message,
			GeneratedAt:          nil,
			CreatedAt:            portrait.CreatedAt,
			UpdatedAt:            now,
		})
		if err != nil {
			return httpapi.InternalError(err)
		}
		return nil
	})
}
