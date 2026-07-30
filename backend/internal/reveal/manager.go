package reveal

import (
	"context"
	"database/sql"
	"fmt"
	"hash/fnv"
	"math/rand"
	"strings"
	"time"

	"soulink-backend/internal/store"
)

type Manager struct {
	store     *store.Store
	schedules *ScheduleSet
	nowFunc   func() time.Time
}

func NewManager(dataStore *store.Store, schedules *ScheduleSet) *Manager {
	return &Manager{store: dataStore, schedules: schedules, nowFunc: time.Now}
}

func (m *Manager) AssignOnFirstCompletion(ctx context.Context, tx *sql.Tx, userID int64, questionnaireKey, questionnaireVersion string, now time.Time) (*time.Time, error) {
	if !m.ready() {
		return nil, nil
	}
	window, ok := m.schedules.Window(questionnaireKey, questionnaireVersion)
	if !ok {
		return nil, nil
	}
	if now.IsZero() {
		now = m.currentTime()
	}
	if !window.EndAt.IsZero() && now.After(window.EndAt) {
		return nil, nil
	}

	assignedCount, err := m.store.CountQuestionnaireRevealAssignments(ctx, tx, questionnaireKey, questionnaireVersion)
	if err != nil {
		return nil, fmt.Errorf("count questionnaire reveal assignments: %w", err)
	}
	batchTimes := window.BatchTimes()
	if len(batchTimes) == 0 {
		return nil, nil
	}
	assigned := batchTimes[assignedCount%len(batchTimes)]
	return &assigned, nil
}

func (m *Manager) ResolveUserRevealAt(ctx context.Context, tx *sql.Tx, userID int64, questionnaireKey, questionnaireVersion string, now time.Time) (*time.Time, error) {
	if !m.ready() {
		return nil, nil
	}
	if now.IsZero() {
		now = m.currentTime()
	}

	window, hasWindow := m.schedules.Window(questionnaireKey, questionnaireVersion)
	if hasWindow && !window.EndAt.IsZero() && !now.Before(window.EndAt) {
		if err := m.store.ClearQuestionnaireRevealAssignments(ctx, tx, questionnaireKey, questionnaireVersion); err != nil {
			return nil, fmt.Errorf("clear questionnaire reveal assignments: %w", err)
		}
		return nil, nil
	}

	state, err := m.store.GetQuestionnaireState(ctx, tx, userID, questionnaireKey, questionnaireVersion)
	if err != nil {
		if err == store.ErrNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get questionnaire state: %w", err)
	}
	if !state.ResultRevealAt.Valid {
		return nil, nil
	}
	if !state.ResultRevealAt.Time.After(now) {
		if err := m.store.UpdateQuestionnaireStateRevealAt(ctx, tx, userID, questionnaireKey, questionnaireVersion, nil, now); err != nil {
			return nil, fmt.Errorf("clear user questionnaire reveal assignment: %w", err)
		}
		return nil, nil
	}

	revealAt := state.ResultRevealAt.Time
	return &revealAt, nil
}

func (m *Manager) SyncCycleResults(ctx context.Context, tx *sql.Tx, cycle store.MatchCycle, now time.Time) error {
	if !m.ready() {
		return nil
	}
	if strings.TrimSpace(cycle.QuestionnaireKey) == "" || strings.TrimSpace(cycle.QuestionnaireVersion) == "" {
		return nil
	}
	if now.IsZero() {
		now = m.currentTime()
	}

	results, err := m.store.ListMatchResultsByCycle(ctx, tx, cycle.ID)
	if err != nil {
		return fmt.Errorf("list cycle match results: %w", err)
	}
	if len(results) == 0 {
		return nil
	}

	window, hasWindow := m.schedules.Window(cycle.QuestionnaireKey, cycle.QuestionnaireVersion)
	if hasWindow && !window.EndAt.IsZero() && !now.Before(window.EndAt) {
		for _, result := range results {
			if err := m.store.UpdateMatchResultRevealAt(ctx, tx, cycle.ID, result.UserID, nil, now); err != nil {
				return fmt.Errorf("clear expired match reveal at for user %d: %w", result.UserID, err)
			}
		}
		if err := m.store.ClearQuestionnaireRevealAssignments(ctx, tx, cycle.QuestionnaireKey, cycle.QuestionnaireVersion); err != nil {
			return fmt.Errorf("clear expired questionnaire reveal assignments: %w", err)
		}
		return nil
	}

	stateRevealAtByUser := make(map[int64]*time.Time, len(results))
	for _, result := range results {
		state, err := m.store.GetQuestionnaireState(ctx, tx, result.UserID, cycle.QuestionnaireKey, cycle.QuestionnaireVersion)
		if err != nil {
			if err == store.ErrNotFound {
				continue
			}
			return fmt.Errorf("get questionnaire state for user %d: %w", result.UserID, err)
		}
		if state.ResultRevealAt.Valid {
			revealAt := state.ResultRevealAt.Time
			stateRevealAtByUser[result.UserID] = &revealAt
		}
	}

	processedPairs := make(map[string]struct{}, len(results))
	for _, result := range results {
		switch result.ResultStatus {
		case "failed":
			if err := m.store.UpdateMatchResultRevealAt(ctx, tx, cycle.ID, result.UserID, nil, now); err != nil {
				return fmt.Errorf("set failed result reveal at for user %d: %w", result.UserID, err)
			}
			if err := m.store.UpdateQuestionnaireStateRevealAt(ctx, tx, result.UserID, cycle.QuestionnaireKey, cycle.QuestionnaireVersion, nil, now); err != nil {
				return fmt.Errorf("clear failed questionnaire reveal assignment for user %d: %w", result.UserID, err)
			}
		case "matched":
			if !result.PartnerUserID.Valid {
				continue
			}
			pairKey, leftUserID, rightUserID := pairIdentity(result.UserID, result.PartnerUserID.Int64)
			if _, exists := processedPairs[pairKey]; exists {
				continue
			}
			processedPairs[pairKey] = struct{}{}

			leftRevealAt := cloneTimePointer(stateRevealAtByUser[leftUserID])
			rightRevealAt := cloneTimePointer(stateRevealAtByUser[rightUserID])
			pairedRevealAt := choosePairedRevealAt(cycle.ID, leftUserID, rightUserID, leftRevealAt, rightRevealAt, window, hasWindow)

			if err := m.store.UpdateMatchResultRevealAt(ctx, tx, cycle.ID, leftUserID, pairedRevealAt, now); err != nil {
				return fmt.Errorf("update paired reveal at for user %d: %w", leftUserID, err)
			}
			if err := m.store.UpdateMatchResultRevealAt(ctx, tx, cycle.ID, rightUserID, pairedRevealAt, now); err != nil {
				return fmt.Errorf("update paired reveal at for user %d: %w", rightUserID, err)
			}
			if err := m.store.UpdateQuestionnaireStateRevealAt(ctx, tx, leftUserID, cycle.QuestionnaireKey, cycle.QuestionnaireVersion, pairedRevealAt, now); err != nil {
				return fmt.Errorf("update questionnaire reveal assignment for user %d: %w", leftUserID, err)
			}
			if err := m.store.UpdateQuestionnaireStateRevealAt(ctx, tx, rightUserID, cycle.QuestionnaireKey, cycle.QuestionnaireVersion, pairedRevealAt, now); err != nil {
				return fmt.Errorf("update questionnaire reveal assignment for user %d: %w", rightUserID, err)
			}
		}
	}

	return nil
}

func (m *Manager) ready() bool {
	return m != nil && m.store != nil && m.schedules != nil
}

func (m *Manager) currentTime() time.Time {
	if m != nil && m.nowFunc != nil {
		return m.nowFunc()
	}
	return time.Now()
}

func pairIdentity(leftUserID, rightUserID int64) (string, int64, int64) {
	if leftUserID > rightUserID {
		leftUserID, rightUserID = rightUserID, leftUserID
	}
	return fmt.Sprintf("%d:%d", leftUserID, rightUserID), leftUserID, rightUserID
}

func choosePairedRevealAt(cycleID, leftUserID, rightUserID int64, leftRevealAt, rightRevealAt *time.Time, window Window, hasWindow bool) *time.Time {
	if leftRevealAt == nil && rightRevealAt == nil {
		if !hasWindow {
			return nil
		}
		batchTimes := window.BatchTimes()
		if len(batchTimes) == 0 {
			return nil
		}
		index := hashedBatchIndex(cycleID, leftUserID, rightUserID, len(batchTimes))
		revealAt := batchTimes[index]
		return &revealAt
	}
	if leftRevealAt == nil {
		return rightRevealAt
	}
	if rightRevealAt == nil {
		return leftRevealAt
	}
	if leftRevealAt.Equal(*rightRevealAt) {
		return leftRevealAt
	}

	seed := cycleID ^ leftUserID ^ rightUserID
	rng := rand.New(rand.NewSource(seed))
	if rng.Intn(2) == 0 {
		return leftRevealAt
	}
	return rightRevealAt
}

func hashedBatchIndex(cycleID, leftUserID, rightUserID int64, batchCount int) int {
	if batchCount <= 0 {
		return 0
	}
	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(fmt.Sprintf("%d:%d:%d", cycleID, leftUserID, rightUserID)))
	return int(hasher.Sum32() % uint32(batchCount))
}

func cloneTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}
