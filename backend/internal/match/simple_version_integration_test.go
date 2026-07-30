// 集成测试运行说明：
//  1. 请先确保本地 PostgreSQL 已启动，并且 .env 或环境变量中的 DB_DSN 可正常连接。
//  2. 该测试默认需要显式开启，避免在普通 go test 时误连真实数据库。
//  3. Windows cmd 运行方式：
//     set SOULINK_RUN_INTEGRATION=1
//     go test ./internal/match -run TestMatchIntegration_RegisterThreeUsersSubmitSameVersionAndRunSimpleVersionMatch -count=1 -v
//  4. Windows PowerShell 运行方式：
//     $env:SOULINK_RUN_INTEGRATION='1'
//     go test ./internal/match -run TestMatchIntegration_RegisterThreeUsersSubmitSameVersionAndRunSimpleVersionMatch -count=1 -v
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
	"soulink-backend/internal/store"
)

var simpleVersionCodePattern = regexp.MustCompile(`\b(\d{6})\b`)

type simpleVersionAuthResponseData struct {
	Token string `json:"token"`
	User  struct {
		ID       int64  `json:"id"`
		Email    string `json:"email"`
		Nickname string `json:"nickname"`
	} `json:"user"`
}

type simpleVersionExpireData struct {
	ExpireSeconds int `json:"expireSeconds"`
}

type simpleVersionProgressData struct {
	Version                   string `json:"version"`
	AnsweredCount             int    `json:"answeredCount"`
	HasCompletedQuestionnaire bool   `json:"hasCompletedQuestionnaire"`
}

type simpleVersionSender struct {
	mu       sync.Mutex
	messages []mail.Message
}

type simpleVersionMatchEnv struct {
	server               *httptest.Server
	client               *http.Client
	adminDB              *sql.DB
	db                   *sql.DB
	dbName               string
	sender               *simpleVersionSender
	matchService         *match.Service
	questionnaireCatalog *questionnaire.Catalog
}

type simpleRegisteredUser struct {
	UserID   int64
	Email    string
	Nickname string
	Token    string
}

func TestMatchIntegration_RegisterThreeUsersSubmitSameVersionAndRunSimpleVersionMatch(t *testing.T) {
	if os.Getenv(matchRunIntegrationEnv) != "1" {
		t.Skipf("set %s=1 to run match integration tests", matchRunIntegrationEnv)
	}

	env := newSimpleVersionMatchEnv(t)
	registeredUsers := []simpleRegisteredUser{
		registerSimpleVersionUser(t, env, fmt.Sprintf("simple-match-a-%d@mails.tsinghua.edu.cn", time.Now().UnixNano()), "用户甲", "计算机学院", "本科三年级", "Start1234"),
		registerSimpleVersionUser(t, env, fmt.Sprintf("simple-match-b-%d@pku.edu.cn", time.Now().UnixNano()+1), "用户乙", "经济学院", "本科二年级", "Start1234"),
		registerSimpleVersionUser(t, env, fmt.Sprintf("simple-match-c-%d@stu.pku.edu.cn", time.Now().UnixNano()+2), "用户丙", "管理学院", "本科四年级", "Start1234"),
	}

	for index, user := range registeredUsers {
		upsertSimpleVersionContact(t, env, user.Token, "wechat", fmt.Sprintf("it_wechat_%d", index+1))
		submitSimpleVersionQuestionnaire(t, env, user.Token)
	}

	seed := int64(20260325)
	cycle, err := env.matchService.RunSimpleVersionMatch(context.Background(), match.SimpleVersionMatchInput{
		QuestionnaireVersion: env.questionnaireCatalog.Version,
		CycleKey:             fmt.Sprintf("simple-%s", strings.ReplaceAll(env.questionnaireCatalog.Version, ".", "-")),
		Title:                fmt.Sprintf("%s 极简随机匹配", env.questionnaireCatalog.Version),
		RevealStatus:         "revealed",
		RevealAt:             time.Date(2026, 3, 27, 20, 0, 0, 0, time.UTC),
		RandomSeed:           &seed,
	})
	if err != nil {
		t.Fatalf("RunSimpleVersionMatch() error = %v", err)
	}
	if cycle.ParticipantCount != 3 {
		t.Fatalf("cycle participantCount = %d, want 3", cycle.ParticipantCount)
	}
	if cycle.MatchedUserCount != 2 {
		t.Fatalf("cycle matchedUserCount = %d, want 2", cycle.MatchedUserCount)
	}

	statusCounts := map[string]int{}
	for _, user := range registeredUsers {
		status, currentResp := doSimpleVersionJSON[currentResultData](t, env, http.MethodGet, "/api/v1/matches/current", user.Token, nil)
		if status != http.StatusOK {
			t.Fatalf("GET /matches/current for %s status = %d, want %d", user.Nickname, status, http.StatusOK)
		}

		statusCounts[currentResp.Data.ResultStatus]++
		switch currentResp.Data.ResultStatus {
		case match.ResultStatusMatched:
			if currentResp.Data.Score == nil {
				t.Fatalf("matched user %s score is nil", user.Nickname)
			}
			if currentResp.Data.Partner == nil {
				t.Fatalf("matched user %s partner is nil", user.Nickname)
			}
			if len(currentResp.Data.Partner.Contacts) == 0 {
				t.Fatalf("matched user %s partner contacts are empty", user.Nickname)
			}
			if !strings.Contains(currentResp.Data.Reason, env.questionnaireCatalog.Version) {
				t.Fatalf("matched user %s reason = %q, want version %q", user.Nickname, currentResp.Data.Reason, env.questionnaireCatalog.Version)
			}
		case match.ResultStatusFailed:
			if currentResp.Data.Partner != nil {
				t.Fatalf("failed user %s partner = %#v, want nil", user.Nickname, currentResp.Data.Partner)
			}
			if currentResp.Data.Score != nil {
				t.Fatalf("failed user %s score = %#v, want nil", user.Nickname, currentResp.Data.Score)
			}
		default:
			t.Fatalf("unexpected result status %q for user %s", currentResp.Data.ResultStatus, user.Nickname)
		}

		t.Logf("用户=%s 状态=%s 分数=%s 对方=%s 联系方式=%s 提示=%s", user.Nickname, currentResp.Data.ResultStatus, formatScore(currentResp.Data.Score), partnerNickname(currentResp.Data.Partner), formatPartnerContacts(currentResp.Data.Partner), currentResp.Data.Message)
	}

	if statusCounts[match.ResultStatusMatched] != 2 {
		t.Fatalf("matched result count = %d, want 2", statusCounts[match.ResultStatusMatched])
	}
	if statusCounts[match.ResultStatusFailed] != 1 {
		t.Fatalf("failed result count = %d, want 1", statusCounts[match.ResultStatusFailed])
	}
}

func newSimpleVersionMatchEnv(t *testing.T) *simpleVersionMatchEnv {
	t.Helper()

	baseDSN := matchIntegrationBaseDSN(t)
	ctx := context.Background()

	adminDB, err := database.Open(ctx, config.DatabaseConfig{DSN: baseDSN})
	if err != nil {
		t.Fatalf("open integration admin database: %v", err)
	}

	dbName := fmt.Sprintf("soulink_match_simple_it_%d", time.Now().UnixNano())
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
	sender := &simpleVersionSender{}
	tokenManager := auth.NewTokenManager("simple-version-match-secret", time.Hour)
	authService := auth.NewService(dataStore, sender, tokenManager, logger, config.AuthConfig{
		JWTSecret:                "simple-version-match-secret",
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
	questionnaireService := questionnaire.NewService(dataStore, catalog)
	matchService := match.NewService(dataStore)

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
	env := &simpleVersionMatchEnv{
		server:               server,
		client:               server.Client(),
		adminDB:              adminDB,
		db:                   db,
		dbName:               dbName,
		sender:               sender,
		matchService:         matchService,
		questionnaireCatalog: catalog,
	}

	t.Cleanup(func() {
		server.Close()
		_ = db.Close()
		_ = dropMatchIntegrationDatabase(context.Background(), adminDB, dbName)
		_ = adminDB.Close()
	})

	return env
}

func (s *simpleVersionSender) Send(_ context.Context, message mail.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.messages = append(s.messages, message)
	return nil
}

func (s *simpleVersionSender) LatestCode(t *testing.T, to string, subjectContains string) string {
	t.Helper()

	s.mu.Lock()
	defer s.mu.Unlock()

	for i := len(s.messages) - 1; i >= 0; i-- {
		message := s.messages[i]
		if message.To != to {
			continue
		}
		if subjectContains != "" && !strings.Contains(message.Subject, subjectContains) {
			continue
		}

		matches := simpleVersionCodePattern.FindStringSubmatch(message.Text)
		if len(matches) == 2 {
			return matches[1]
		}
	}

	t.Fatalf("no verification code found for %q with subject containing %q", to, subjectContains)
	return ""
}

func registerSimpleVersionUser(t *testing.T, env *simpleVersionMatchEnv, email, nickname, department, grade, password string) simpleRegisteredUser {
	t.Helper()

	status, sendCodeResp := doSimpleVersionJSON[simpleVersionExpireData](t, env, http.MethodPost, "/api/v1/auth/register/send-code", "", map[string]string{
		"email": email,
	})
	if status != http.StatusOK {
		t.Fatalf("register/send-code for %s status = %d, want %d", email, status, http.StatusOK)
	}
	if sendCodeResp.Data.ExpireSeconds <= 0 {
		t.Fatalf("register/send-code expireSeconds = %d, want > 0", sendCodeResp.Data.ExpireSeconds)
	}

	code := env.sender.LatestCode(t, email, "注册验证码")
	status, registerResp := doSimpleVersionJSON[simpleVersionAuthResponseData](t, env, http.MethodPost, "/api/v1/auth/register", "", map[string]string{
		"email":      email,
		"code":       code,
		"password":   password,
		"nickname":   nickname,
		"department": department,
		"grade":      grade,
	})
	if status != http.StatusOK {
		t.Fatalf("register for %s status = %d, want %d", email, status, http.StatusOK)
	}
	if registerResp.Data.Token == "" {
		t.Fatalf("register token for %s is empty", email)
	}

	return simpleRegisteredUser{
		UserID:   registerResp.Data.User.ID,
		Email:    email,
		Nickname: nickname,
		Token:    registerResp.Data.Token,
	}
}

func upsertSimpleVersionContact(t *testing.T, env *simpleVersionMatchEnv, token, contactType, value string) {
	t.Helper()

	status, _ := doSimpleVersionJSON[map[string]any](t, env, http.MethodPut, fmt.Sprintf("/api/v1/users/me/contacts/%s", contactType), token, map[string]string{
		"value": value,
	})
	if status != http.StatusOK {
		t.Fatalf("PUT /users/me/contacts/%s status = %d, want %d", contactType, status, http.StatusOK)
	}
}

func submitSimpleVersionQuestionnaire(t *testing.T, env *simpleVersionMatchEnv, token string) {
	t.Helper()

	answers, err := buildRequiredQuestionnaireAnswers(env.questionnaireCatalog)
	if err != nil {
		t.Fatalf("buildRequiredQuestionnaireAnswers() error = %v", err)
	}

	status, saveResp := doSimpleVersionJSON[questionnaire.SaveAnswersResult](t, env, http.MethodPut, "/api/v1/questionnaires/current/answers", token, map[string]any{
		"answers": answers,
		"action":  "submit",
	})
	if status != http.StatusOK {
		t.Fatalf("PUT /questionnaires/current/answers status = %d, want %d", status, http.StatusOK)
	}
	if !saveResp.Data.HasCompletedQuestionnaire {
		t.Fatal("questionnaire save result hasCompletedQuestionnaire = false, want true")
	}

	status, progressResp := doSimpleVersionJSON[simpleVersionProgressData](t, env, http.MethodGet, "/api/v1/questionnaires/current/progress", token, nil)
	if status != http.StatusOK {
		t.Fatalf("GET /questionnaires/current/progress status = %d, want %d", status, http.StatusOK)
	}
	if progressResp.Data.Version != env.questionnaireCatalog.Version {
		t.Fatalf("progress version = %q, want %q", progressResp.Data.Version, env.questionnaireCatalog.Version)
	}
	if !progressResp.Data.HasCompletedQuestionnaire {
		t.Fatal("questionnaire progress hasCompletedQuestionnaire = false, want true")
	}
}

func buildRequiredQuestionnaireAnswers(catalog *questionnaire.Catalog) ([]map[string]any, error) {
	answers := make([]map[string]any, 0, catalog.RequiredCount)
	for _, question := range catalog.Questions {
		if !question.Required {
			continue
		}

		value, err := defaultQuestionAnswer(question)
		if err != nil {
			return nil, fmt.Errorf("question %s: %w", question.ID, err)
		}
		answer := map[string]any{
			"questionId": question.ID,
			"value":      value,
		}
		if question.SupportsImportance {
			answer["importance"] = false
		}
		answers = append(answers, answer)
	}
	return answers, nil
}

func defaultQuestionAnswer(question questionnaire.Question) (any, error) {
	switch question.Type {
	case questionnaire.QuestionTypeSingleChoice:
		if len(question.Options) == 0 {
			return nil, fmt.Errorf("single choice has no options")
		}
		return question.Options[0].Value, nil
	case questionnaire.QuestionTypeMultipleChoice:
		if len(question.Options) == 0 {
			return nil, fmt.Errorf("multiple choice has no options")
		}
		count := 1
		if question.MinSelections != nil && *question.MinSelections > count {
			count = *question.MinSelections
		}
		if question.MaxSelections != nil && count > *question.MaxSelections {
			count = *question.MaxSelections
		}
		if count > len(question.Options) {
			count = len(question.Options)
		}
		values := make([]string, 0, count)
		for i := 0; i < count; i++ {
			values = append(values, question.Options[i].Value)
		}
		return values, nil
	case questionnaire.QuestionTypeNumber:
		value := 1
		if question.Min != nil {
			value = *question.Min
		}
		if question.Max != nil && value > *question.Max {
			value = *question.Max
		}
		return value, nil
	case questionnaire.QuestionTypeScale:
		value := 4
		if question.Min != nil && value < *question.Min {
			value = *question.Min
		}
		if question.Max != nil && value > *question.Max {
			value = *question.Max
		}
		return value, nil
	case questionnaire.QuestionTypeRange:
		if question.AllowUnlimited {
			return map[string]any{"unlimited": true}, nil
		}
		min := 0
		max := 0
		if question.Min != nil {
			min = *question.Min
		}
		if question.Max != nil {
			max = *question.Max
		} else {
			max = min
		}
		if max < min {
			max = min
		}
		return map[string]any{"min": min, "max": max}, nil
	case questionnaire.QuestionTypeText:
		value := trimToMaxLength(fmt.Sprintf("集成测试答案-%s", question.ID), question.MaxLength)
		if value == "" {
			value = "测试"
		}
		return value, nil
	default:
		return nil, fmt.Errorf("unsupported question type %q", question.Type)
	}
}

func trimToMaxLength(value string, maxLength *int) string {
	if maxLength == nil || *maxLength <= 0 {
		return value
	}
	runes := []rune(value)
	if len(runes) <= *maxLength {
		return value
	}
	return string(runes[:*maxLength])
}

func doSimpleVersionJSON[T any](t *testing.T, env *simpleVersionMatchEnv, method, path, token string, body any) (int, matchAPIEnvelope[T]) {
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

func formatScore(score *float64) string {
	if score == nil {
		return "-"
	}
	return fmt.Sprintf("%.1f", *score)
}

func partnerNickname(partner *struct {
	Nickname string `json:"nickname"`
	Contacts []struct {
		Type  string `json:"type"`
		Label string `json:"label"`
		Value string `json:"value"`
	} `json:"contacts"`
}) string {
	if partner == nil {
		return "-"
	}
	return partner.Nickname
}

func formatPartnerContacts(partner *struct {
	Nickname string `json:"nickname"`
	Contacts []struct {
		Type  string `json:"type"`
		Label string `json:"label"`
		Value string `json:"value"`
	} `json:"contacts"`
}) string {
	if partner == nil || len(partner.Contacts) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(partner.Contacts))
	for _, contact := range partner.Contacts {
		parts = append(parts, fmt.Sprintf("%s:%s", contact.Label, contact.Value))
	}
	return strings.Join(parts, ", ")
}
