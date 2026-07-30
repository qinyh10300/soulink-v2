package match_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"soulink-backend/internal/app"
	"soulink-backend/internal/auth"
	"soulink-backend/internal/config"
	"soulink-backend/internal/mail"
	"soulink-backend/internal/match"
	"soulink-backend/internal/meta"
	"soulink-backend/internal/migrate"
	"soulink-backend/internal/platform/database"
	"soulink-backend/internal/profile"
	"soulink-backend/internal/questionnaire"
	"soulink-backend/internal/reveal"
	"soulink-backend/internal/store"
)

var stagedRevealCodePattern = regexp.MustCompile(`\b(\d{6})\b`)

type stagedRevealSender struct {
	mu       sync.Mutex
	messages []mail.Message
}

type stagedRevealCache struct {
	mu    sync.Mutex
	items map[string]match.ResultView
}

type stagedRevealEnv struct {
	server               *httptest.Server
	client               *http.Client
	adminDB              *sql.DB
	db                   *sql.DB
	dbName               string
	sender               *stagedRevealSender
	cache                *stagedRevealCache
	matchService         *match.Service
	questionnaireCatalog *questionnaire.Catalog
	now                  *time.Time
}

type stagedProgressData struct {
	Version                   string `json:"version"`
	ResultRevealAt            string `json:"resultRevealAt"`
	AnsweredCount             int    `json:"answeredCount"`
	HasCompletedQuestionnaire bool   `json:"hasCompletedQuestionnaire"`
}

type stagedCurrentData struct {
	MatchID      int64    `json:"matchId"`
	ResultStatus string   `json:"resultStatus"`
	RevealAt     string   `json:"revealAt"`
	Message      string   `json:"message"`
	Score        *float64 `json:"score"`
	Partner      *struct {
		Nickname string `json:"nickname"`
		Contacts []struct {
			Type  string `json:"type"`
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"contacts"`
	} `json:"partner"`
}

type stagedCycleStatusData struct {
	HasCurrentCycle   bool   `json:"hasCurrentCycle"`
	HasUserResult     bool   `json:"hasUserResult"`
	UserResultStatus  string `json:"userResultStatus"`
	UserRevealAt      string `json:"userRevealAt"`
	CanViewUserResult bool   `json:"canViewUserResult"`
}

type stagedAuthData struct {
	Token string `json:"token"`
	User  struct {
		ID       int64  `json:"id"`
		Email    string `json:"email"`
		Nickname string `json:"nickname"`
	} `json:"user"`
}

type stagedUser struct {
	UserID   int64
	Email    string
	Nickname string
	Token    string
}

func TestMatchIntegration_StagedRevealLifecycle(t *testing.T) {
	if os.Getenv(matchRunIntegrationEnv) != "1" {
		t.Skipf("set %s=1 to run match integration tests", matchRunIntegrationEnv)
	}

	env := newStagedRevealEnv(t)
	users := []stagedUser{
		registerStagedUser(t, env, fmt.Sprintf("staged-a-%d@mails.tsinghua.edu.cn", time.Now().UnixNano()), "用户甲", "计算机学院", "本科三年级", "Start1234"),
		registerStagedUser(t, env, fmt.Sprintf("staged-b-%d@pku.edu.cn", time.Now().UnixNano()+1), "用户乙", "经济学院", "本科二年级", "Start1234"),
		registerStagedUser(t, env, fmt.Sprintf("staged-c-%d@stu.pku.edu.cn", time.Now().UnixNano()+2), "用户丙", "管理学院", "本科四年级", "Start1234"),
	}

	initialRevealAt := make(map[int64]string, len(users))
	for index, user := range users {
		upsertStagedContact(t, env, user.Token, "wechat", fmt.Sprintf("staged_wechat_%d", index+1))
		progress := submitStagedQuestionnaire(t, env, user.Token)
		initialRevealAt[user.UserID] = progress.ResultRevealAt
	}
	if initialRevealAt[users[0].UserID] == "" || initialRevealAt[users[1].UserID] == "" || initialRevealAt[users[2].UserID] == "" {
		t.Fatalf("initial reveal times = %#v, want all assigned", initialRevealAt)
	}
	if initialRevealAt[users[0].UserID] == initialRevealAt[users[1].UserID] {
		t.Fatalf("first two reveal times are equal before matching: %#v", initialRevealAt)
	}

	matchedScore := 88.5
	partnerA := users[1].UserID
	partnerB := users[0].UserID
	cycle, err := env.matchService.RecordCycleResults(context.Background(), match.RecordCycleInput{
		CycleKey:             "staged-v0-1",
		Title:                "v0.1 分批公布匹配",
		QuestionnaireKey:     env.questionnaireCatalog.Key,
		QuestionnaireVersion: env.questionnaireCatalog.Version,
		RevealStatus:         "revealed",
		RevealAt:             time.Date(2026, 3, 27, 20, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		Results: []match.RecordUserResultInput{
			{UserID: users[0].UserID, ResultStatus: match.ResultStatusMatched, PartnerUserID: &partnerA, MatchScore: &matchedScore, MatchReason: "测试匹配成功", Message: "匹配成功"},
			{UserID: users[1].UserID, ResultStatus: match.ResultStatusMatched, PartnerUserID: &partnerB, MatchScore: &matchedScore, MatchReason: "测试匹配成功", Message: "匹配成功"},
			{UserID: users[2].UserID, ResultStatus: match.ResultStatusFailed, Message: "本轮未匹配成功"},
		},
	})
	if err != nil {
		t.Fatalf("RecordCycleResults() error = %v", err)
	}
	if cycle.ParticipantCount != 3 || cycle.MatchedUserCount != 2 {
		t.Fatalf("cycle counts = %#v, want participant=3 matched=2", cycle)
	}
	if env.cache.Len() != 3 {
		t.Fatalf("cache size = %d, want 3", env.cache.Len())
	}

	progressA := getStagedProgress(t, env, users[0].Token)
	progressB := getStagedProgress(t, env, users[1].Token)
	progressC := getStagedProgress(t, env, users[2].Token)
	if progressA.ResultRevealAt == "" || progressB.ResultRevealAt == "" {
		t.Fatalf("matched reveal times = %q / %q, want both assigned", progressA.ResultRevealAt, progressB.ResultRevealAt)
	}
	if progressA.ResultRevealAt != progressB.ResultRevealAt {
		t.Fatalf("matched reveal times = %q / %q, want synchronized", progressA.ResultRevealAt, progressB.ResultRevealAt)
	}
	if progressA.ResultRevealAt != initialRevealAt[users[0].UserID] && progressA.ResultRevealAt != initialRevealAt[users[1].UserID] {
		t.Fatalf("synchronized revealAt = %q, want one of original values %#v", progressA.ResultRevealAt, initialRevealAt)
	}
	if progressC.ResultRevealAt != "" {
		t.Fatalf("failed user revealAt = %q, want empty", progressC.ResultRevealAt)
	}

	status, blockedCurrent := doStagedJSON[stagedCurrentData](t, env, http.MethodGet, "/api/v1/matches/current", users[0].Token, nil)
	if status != http.StatusForbidden {
		t.Fatalf("matched GET /matches/current before reveal status = %d, want %d", status, http.StatusForbidden)
	}
	if !strings.Contains(blockedCurrent.Message, progressA.ResultRevealAt) {
		t.Fatalf("blocked message = %q, want reveal time %q", blockedCurrent.Message, progressA.ResultRevealAt)
	}

	status, cycleStatus := doStagedJSON[stagedCycleStatusData](t, env, http.MethodGet, "/api/v1/matches/cycle/status", users[0].Token, nil)
	if status != http.StatusOK {
		t.Fatalf("GET /matches/cycle/status status = %d, want %d", status, http.StatusOK)
	}
	if !cycleStatus.Data.HasUserResult || cycleStatus.Data.CanViewUserResult {
		t.Fatalf("cycle status = %#v, want user result but not viewable yet", cycleStatus.Data)
	}
	if cycleStatus.Data.UserRevealAt != progressA.ResultRevealAt {
		t.Fatalf("cycle status userRevealAt = %q, want %q", cycleStatus.Data.UserRevealAt, progressA.ResultRevealAt)
	}

	status, failedCurrent := doStagedJSON[stagedCurrentData](t, env, http.MethodGet, "/api/v1/matches/current", users[2].Token, nil)
	if status != http.StatusOK {
		t.Fatalf("failed GET /matches/current status = %d, want %d", status, http.StatusOK)
	}
	if failedCurrent.Data.ResultStatus != match.ResultStatusFailed {
		t.Fatalf("failed resultStatus = %q, want %q", failedCurrent.Data.ResultStatus, match.ResultStatusFailed)
	}
	if failedCurrent.Data.RevealAt != "" {
		t.Fatalf("failed revealAt = %q, want empty", failedCurrent.Data.RevealAt)
	}

	revealAt, err := time.Parse(time.RFC3339, progressA.ResultRevealAt)
	if err != nil {
		t.Fatalf("parse revealAt %q: %v", progressA.ResultRevealAt, err)
	}
	*env.now = revealAt.Add(time.Minute)

	status, matchedCurrent := doStagedJSON[stagedCurrentData](t, env, http.MethodGet, "/api/v1/matches/current", users[0].Token, nil)
	if status != http.StatusOK {
		t.Fatalf("matched GET /matches/current after reveal status = %d, want %d", status, http.StatusOK)
	}
	if matchedCurrent.Data.ResultStatus != match.ResultStatusMatched {
		t.Fatalf("matched resultStatus = %q, want %q", matchedCurrent.Data.ResultStatus, match.ResultStatusMatched)
	}
	if matchedCurrent.Data.Partner == nil || len(matchedCurrent.Data.Partner.Contacts) == 0 {
		t.Fatalf("matched partner = %#v, want contacts", matchedCurrent.Data.Partner)
	}

	progressAfterReveal := getStagedProgress(t, env, users[0].Token)
	if progressAfterReveal.ResultRevealAt != "" {
		t.Fatalf("progress revealAt after reveal = %q, want empty", progressAfterReveal.ResultRevealAt)
	}

	*env.now = time.Date(2026, 3, 27, 21, 20, 0, 0, time.FixedZone("CST", 8*3600))
	_ = getStagedProgress(t, env, users[2].Token)
	for _, user := range users {
		progress := getStagedProgress(t, env, user.Token)
		if progress.ResultRevealAt != "" {
			t.Fatalf("progress revealAt after latest window for user %s = %q, want empty", user.Nickname, progress.ResultRevealAt)
		}
	}
}

func newStagedRevealEnv(t *testing.T) *stagedRevealEnv {
	t.Helper()

	baseDSN := matchIntegrationBaseDSN(t)
	ctx := context.Background()

	adminDB, err := database.Open(ctx, config.DatabaseConfig{DSN: baseDSN})
	if err != nil {
		t.Fatalf("open integration admin database: %v", err)
	}

	dbName := fmt.Sprintf("soulink_match_reveal_it_%d", time.Now().UnixNano())
	if _, err := adminDB.ExecContext(ctx, fmt.Sprintf(`CREATE DATABASE "%s"`, dbName)); err != nil {
		_ = adminDB.Close()
		t.Fatalf("create integration database %q: %v", dbName, err)
	}

	testDSN := replaceMatchDatabaseName(t, baseDSN, dbName)
	db, err := database.Open(ctx, config.DatabaseConfig{DSN: testDSN})
	if err != nil {
		_ = dropMatchIntegrationDatabase(context.Background(), adminDB, dbName)
		_ = adminDB.Close()
		t.Fatalf("open integration test database: %v", err)
	}
	if err := migrate.Up(ctx, db); err != nil {
		_ = db.Close()
		_ = dropMatchIntegrationDatabase(context.Background(), adminDB, dbName)
		_ = adminDB.Close()
		t.Fatalf("run integration migrations: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	dataStore := store.New(db)
	sender := &stagedRevealSender{}
	cache := &stagedRevealCache{items: make(map[string]match.ResultView)}
	tokenManager := auth.NewTokenManager("staged-reveal-secret", time.Hour)
	authService := auth.NewService(dataStore, sender, tokenManager, logger, config.AuthConfig{
		JWTSecret:                "staged-reveal-secret",
		JWTTTL:                   time.Hour,
		VerificationCodeTTL:      5 * time.Minute,
		VerificationCodeCooldown: time.Second,
		BCryptCost:               bcrypt.MinCost,
	})
	catalog, err := questionnaire.LoadDefaultCatalog()
	if err != nil {
		_ = db.Close()
		_ = dropMatchIntegrationDatabase(context.Background(), adminDB, dbName)
		_ = adminDB.Close()
		t.Fatalf("load default questionnaire catalog: %v", err)
	}

	currentTime := time.Date(2026, 3, 27, 18, 0, 0, 0, time.FixedZone("CST", 8*3600))
	schedule := &reveal.ScheduleSet{
		Timezone: "Asia/Shanghai",
		Schedules: []reveal.Schedule{{
			QuestionnaireKey:     catalog.Key,
			QuestionnaireVersion: catalog.Version,
			StartAt:              "2026-03-27T21:00:00+08:00",
			EndAt:                "2026-03-27T21:10:00+08:00",
			BatchCount:           2,
		}},
	}
	revealManager := reveal.NewManager(dataStore, schedule)
	questionnaireService := questionnaire.NewService(
		dataStore,
		catalog,
		questionnaire.WithNowFunc(func() time.Time { return currentTime }),
		questionnaire.WithResultRevealCoordinator(revealManager),
	)
	matchService := match.NewService(
		dataStore,
		match.WithNowFunc(func() time.Time { return currentTime }),
		match.WithRevealManager(revealManager),
		match.WithResultCache(cache, time.Hour),
	)

	router := app.NewRouter(app.Dependencies{
		Logger:               logger,
		AllowedOrigins:       []string{"http://localhost:5173"},
		AuthMiddleware:       auth.Middleware(dataStore, tokenManager),
		AuthHandler:          auth.NewHandler(authService),
		ProfileHandler:       profile.NewHandler(profile.NewService(dataStore)),
		MetaHandler:          meta.NewHandler(meta.NewService()),
		QuestionnaireHandler: questionnaire.NewHandler(questionnaireService),
		MatchHandler:         match.NewHandler(matchService),
	})

	server := httptest.NewServer(router)
	env := &stagedRevealEnv{
		server:               server,
		client:               server.Client(),
		adminDB:              adminDB,
		db:                   db,
		dbName:               dbName,
		sender:               sender,
		cache:                cache,
		matchService:         matchService,
		questionnaireCatalog: catalog,
		now:                  &currentTime,
	}

	t.Cleanup(func() {
		server.Close()
		_ = db.Close()
		_ = dropMatchIntegrationDatabase(context.Background(), adminDB, dbName)
		_ = adminDB.Close()
	})
	return env
}

func (s *stagedRevealSender) Send(_ context.Context, message mail.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = append(s.messages, message)
	return nil
}

func (s *stagedRevealSender) latestCode(t *testing.T, to string) string {
	t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := len(s.messages) - 1; i >= 0; i-- {
		message := s.messages[i]
		if message.To != to {
			continue
		}
		matches := stagedRevealCodePattern.FindStringSubmatch(message.Text)
		if len(matches) == 2 {
			return matches[1]
		}
	}
	t.Fatalf("no verification code found for %q", to)
	return ""
}

func (c *stagedRevealCache) LoadCurrent(_ context.Context, cycleID, userID int64) (match.ResultView, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	view, ok := c.items[fmt.Sprintf("%d:%d", cycleID, userID)]
	if !ok {
		return match.ResultView{}, match.ErrResultCacheNotFound
	}
	return view, nil
}

func (c *stagedRevealCache) StoreCurrent(_ context.Context, cycleID, userID int64, view match.ResultView, _ time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[fmt.Sprintf("%d:%d", cycleID, userID)] = view
	return nil
}

func (c *stagedRevealCache) DeleteCurrent(_ context.Context, cycleID, userID int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, fmt.Sprintf("%d:%d", cycleID, userID))
	return nil
}

func (c *stagedRevealCache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.items)
}

func registerStagedUser(t *testing.T, env *stagedRevealEnv, email, nickname, department, grade, password string) stagedUser {
	t.Helper()
	status, _ := doStagedJSON[map[string]any](t, env, http.MethodPost, "/api/v1/auth/register/send-code", "", map[string]string{"email": email})
	if status != http.StatusOK {
		t.Fatalf("register/send-code status = %d, want %d", status, http.StatusOK)
	}
	code := env.sender.latestCode(t, email)
	status, registerResp := doStagedJSON[stagedAuthData](t, env, http.MethodPost, "/api/v1/auth/register", "", map[string]string{
		"email":      email,
		"code":       code,
		"password":   password,
		"nickname":   nickname,
		"department": department,
		"grade":      grade,
	})
	if status != http.StatusOK {
		t.Fatalf("register status = %d, want %d", status, http.StatusOK)
	}
	return stagedUser{UserID: registerResp.Data.User.ID, Email: email, Nickname: nickname, Token: registerResp.Data.Token}
}

func upsertStagedContact(t *testing.T, env *stagedRevealEnv, token, contactType, value string) {
	t.Helper()
	status, _ := doStagedJSON[map[string]any](t, env, http.MethodPut, fmt.Sprintf("/api/v1/users/me/contacts/%s", contactType), token, map[string]string{"value": value})
	if status != http.StatusOK {
		t.Fatalf("upsert contact status = %d, want %d", status, http.StatusOK)
	}
}

func submitStagedQuestionnaire(t *testing.T, env *stagedRevealEnv, token string) stagedProgressData {
	t.Helper()
	answers, err := buildStagedRequiredAnswers(env.questionnaireCatalog)
	if err != nil {
		t.Fatalf("build required answers: %v", err)
	}
	status, saveResp := doStagedJSON[questionnaire.SaveAnswersResult](t, env, http.MethodPut, "/api/v1/questionnaires/current/answers", token, map[string]any{"answers": answers, "action": "submit"})
	if status != http.StatusOK {
		t.Fatalf("save questionnaire status = %d, want %d", status, http.StatusOK)
	}
	if !saveResp.Data.HasCompletedQuestionnaire {
		t.Fatal("save questionnaire hasCompletedQuestionnaire = false, want true")
	}
	return getStagedProgress(t, env, token)
}

func getStagedProgress(t *testing.T, env *stagedRevealEnv, token string) stagedProgressData {
	t.Helper()
	status, progressResp := doStagedJSON[stagedProgressData](t, env, http.MethodGet, "/api/v1/questionnaires/current/progress", token, nil)
	if status != http.StatusOK {
		t.Fatalf("progress status = %d, want %d", status, http.StatusOK)
	}
	if !progressResp.Data.HasCompletedQuestionnaire {
		t.Fatal("progress hasCompletedQuestionnaire = false, want true")
	}
	return progressResp.Data
}

func buildStagedRequiredAnswers(catalog *questionnaire.Catalog) ([]map[string]any, error) {
	answers := make([]map[string]any, 0, catalog.RequiredCount)
	for _, question := range catalog.Questions {
		if !question.Required {
			continue
		}
		value, err := defaultQuestionAnswer(question)
		if err != nil {
			return nil, err
		}
		answer := map[string]any{"questionId": question.ID, "value": value}
		if question.SupportsImportance {
			answer["importance"] = false
		}
		answers = append(answers, answer)
	}
	return answers, nil
}

func doStagedJSON[T any](t *testing.T, env *stagedRevealEnv, method, path, token string, body any) (int, matchAPIEnvelope[T]) {
	t.Helper()
	var bodyReader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		bodyReader = bytes.NewReader(payload)
	}
	req, err := http.NewRequest(method, env.server.URL+path, bodyReader)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := env.client.Do(req)
	if err != nil {
		t.Fatalf("perform request %s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	var envelope matchAPIEnvelope[T]
	if err := json.Unmarshal(rawBody, &envelope); err != nil {
		t.Fatalf("decode response body %q: %v", string(rawBody), err)
	}
	return resp.StatusCode, envelope
}
