// 集成测试运行说明：
//  1. 请先确保本地 PostgreSQL 已启动，并且 .env 或环境变量中的 DB_DSN 可正常连接。
//  2. 该测试默认需要显式开启，避免在普通 go test 时误连真实数据库。
//  3. Windows cmd 运行方式：
//     set SOULINK_RUN_INTEGRATION=1
//     go test ./internal/match -run TestMatchIntegration_UserViewsOwnResults -count=1 -v
//  4. Windows PowerShell 运行方式：
//     $env:SOULINK_RUN_INTEGRATION='1'
//     go test ./internal/match -run TestMatchIntegration_UserViewsOwnResults -count=1 -v
package match_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"soulink-backend/internal/app"
	"soulink-backend/internal/auth"
	"soulink-backend/internal/config"
	"soulink-backend/internal/match"
	"soulink-backend/internal/meta"
	"soulink-backend/internal/migrate"
	"soulink-backend/internal/platform/database"
	"soulink-backend/internal/profile"
	"soulink-backend/internal/questionnaire"
	"soulink-backend/internal/store"
)

const matchRunIntegrationEnv = "SOULINK_RUN_INTEGRATION"

type matchAPIEnvelope[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

type currentResultData struct {
	MatchID      int64    `json:"matchId"`
	ResultStatus string   `json:"resultStatus"`
	Message      string   `json:"message"`
	Score        *float64 `json:"score"`
	Reason       string   `json:"reason"`
	Partner      *struct {
		Nickname string `json:"nickname"`
		Contacts []struct {
			Type  string `json:"type"`
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"contacts"`
	} `json:"partner"`
}

type historyData struct {
	Items []currentResultData `json:"items"`
}

type cycleStatusData struct {
	HasCurrentCycle    bool   `json:"hasCurrentCycle"`
	HasUserResult      bool   `json:"hasUserResult"`
	UserResultStatus   string `json:"userResultStatus"`
	MatchedUserCount   int    `json:"matchedUserCount"`
	ParticipantCount   int    `json:"participantCount"`
	SuccessRatePercent int    `json:"successRatePercent"`
}

type matchIntegrationEnv struct {
	server       *httptest.Server
	client       *http.Client
	adminDB      *sql.DB
	db           *sql.DB
	dbName       string
	store        *store.Store
	tokenManager *auth.TokenManager
	matchService *match.Service
}

func TestMatchIntegration_UserViewsOwnResults(t *testing.T) {
	if os.Getenv(matchRunIntegrationEnv) != "1" {
		t.Skipf("set %s=1 to run match integration tests", matchRunIntegrationEnv)
	}

	env := newMatchIntegrationEnv(t)
	ctx := context.Background()
	now := time.Date(2026, 3, 27, 20, 0, 0, 0, time.UTC)

	userMatched := createIntegrationUser(t, env, "match-user-a@example.com", "用户甲", "计算机学院", "本科三年级", now)
	partner := createIntegrationUser(t, env, "match-user-b@example.com", "用户乙", "经济学院", "本科二年级", now)
	userFailed := createIntegrationUser(t, env, "match-user-c@example.com", "用户丙", "管理学院", "本科四年级", now)

	upsertContact(t, env, partner.ID, "wechat", "partner_wechat", now)
	upsertContact(t, env, partner.ID, "phone", "13800000000", now)

	matchedToken := issueIntegrationToken(t, env, userMatched.ID, now)
	failedToken := issueIntegrationToken(t, env, userFailed.ID, now)

	matchedScore := 91.5
	partnerID := partner.ID
	userMatchedID := userMatched.ID
	if _, err := env.matchService.RecordCycleResults(ctx, match.RecordCycleInput{
		CycleKey:     "2026-w13",
		Title:        "2026 第 13 周匹配",
		RevealStatus: "revealed",
		RevealAt:     now,
		Results: []match.RecordUserResultInput{
			{
				UserID:        userMatched.ID,
				ResultStatus:  match.ResultStatusMatched,
				PartnerUserID: &partnerID,
				MatchScore:    &matchedScore,
				MatchReason:   "你们在兴趣偏好和生活节奏上高度匹配",
				Message:       "匹配成功",
			},
			{
				UserID:        partner.ID,
				ResultStatus:  match.ResultStatusMatched,
				PartnerUserID: &userMatchedID,
				MatchScore:    &matchedScore,
				MatchReason:   "你们在兴趣偏好和生活节奏上高度匹配",
				Message:       "匹配成功",
			},
			{
				UserID:       userFailed.ID,
				ResultStatus: match.ResultStatusFailed,
				Message:      "本轮未匹配成功",
			},
		},
	}); err != nil {
		t.Fatalf("RecordCycleResults() error = %v", err)
	}

	status, matchedCurrent := doMatchJSON[currentResultData](t, env, http.MethodGet, "/api/v1/matches/current", matchedToken)
	if status != http.StatusOK {
		t.Fatalf("matched GET /matches/current status = %d, want %d", status, http.StatusOK)
	}
	if matchedCurrent.Data.ResultStatus != match.ResultStatusMatched {
		t.Fatalf("matched current resultStatus = %q, want %q", matchedCurrent.Data.ResultStatus, match.ResultStatusMatched)
	}
	if matchedCurrent.Data.Score == nil || *matchedCurrent.Data.Score != matchedScore {
		t.Fatalf("matched current score = %#v, want %v", matchedCurrent.Data.Score, matchedScore)
	}
	if matchedCurrent.Data.Partner == nil || len(matchedCurrent.Data.Partner.Contacts) != 2 {
		t.Fatalf("matched current partner contacts = %#v, want 2 contacts", matchedCurrent.Data.Partner)
	}
	if matchedCurrent.Data.Partner.Contacts[0].Label != "手机号" && matchedCurrent.Data.Partner.Contacts[1].Label != "手机号" {
		t.Fatalf("matched current partner contacts missing phone label: %#v", matchedCurrent.Data.Partner.Contacts)
	}

	status, failedCurrent := doMatchJSON[currentResultData](t, env, http.MethodGet, "/api/v1/matches/current", failedToken)
	if status != http.StatusOK {
		t.Fatalf("failed GET /matches/current status = %d, want %d", status, http.StatusOK)
	}
	if failedCurrent.Data.ResultStatus != match.ResultStatusFailed {
		t.Fatalf("failed current resultStatus = %q, want %q", failedCurrent.Data.ResultStatus, match.ResultStatusFailed)
	}
	if failedCurrent.Data.Partner != nil {
		t.Fatalf("failed current partner = %#v, want nil", failedCurrent.Data.Partner)
	}

	status, historyResp := doMatchJSON[historyData](t, env, http.MethodGet, "/api/v1/matches/history", matchedToken)
	if status != http.StatusOK {
		t.Fatalf("GET /matches/history status = %d, want %d", status, http.StatusOK)
	}
	if len(historyResp.Data.Items) != 1 {
		t.Fatalf("history items = %d, want %d", len(historyResp.Data.Items), 1)
	}
	if historyResp.Data.Items[0].MatchID != matchedCurrent.Data.MatchID {
		t.Fatalf("history first matchId = %d, want %d", historyResp.Data.Items[0].MatchID, matchedCurrent.Data.MatchID)
	}

	status, detailResp := doMatchJSON[currentResultData](t, env, http.MethodGet, fmt.Sprintf("/api/v1/matches/%d", matchedCurrent.Data.MatchID), matchedToken)
	if status != http.StatusOK {
		t.Fatalf("GET /matches/{id} status = %d, want %d", status, http.StatusOK)
	}
	if detailResp.Data.Partner == nil || detailResp.Data.Partner.Nickname != "用户乙" {
		t.Fatalf("detail partner = %#v, want 用户乙", detailResp.Data.Partner)
	}

	status, cycleStatusResp := doMatchJSON[cycleStatusData](t, env, http.MethodGet, "/api/v1/matches/cycle/status", matchedToken)
	if status != http.StatusOK {
		t.Fatalf("GET /matches/cycle/status status = %d, want %d", status, http.StatusOK)
	}
	if !cycleStatusResp.Data.HasCurrentCycle || !cycleStatusResp.Data.HasUserResult {
		t.Fatalf("cycle status = %#v, want current cycle and user result", cycleStatusResp.Data)
	}
	if cycleStatusResp.Data.UserResultStatus != match.ResultStatusMatched {
		t.Fatalf("cycle status userResultStatus = %q, want %q", cycleStatusResp.Data.UserResultStatus, match.ResultStatusMatched)
	}
	if cycleStatusResp.Data.ParticipantCount != 3 || cycleStatusResp.Data.MatchedUserCount != 2 || cycleStatusResp.Data.SuccessRatePercent != 66 {
		t.Fatalf("cycle status counts = %#v, want participant=3 matched=2 rate=66", cycleStatusResp.Data)
	}
}

func newMatchIntegrationEnv(t *testing.T) *matchIntegrationEnv {
	t.Helper()

	baseDSN := matchIntegrationBaseDSN(t)
	ctx := context.Background()

	adminDB, err := database.Open(ctx, config.DatabaseConfig{DSN: baseDSN})
	if err != nil {
		t.Fatalf("open integration admin database: %v", err)
	}

	dbName := fmt.Sprintf("soulink_match_it_%d", time.Now().UnixNano())
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
	tokenManager := auth.NewTokenManager("match-integration-secret", time.Hour)
	matchService := match.NewService(dataStore)

	router := app.NewRouter(app.Dependencies{
		Logger:               logger,
		AllowedOrigins:       []string{"http://localhost:5173"},
		AuthMiddleware:       auth.Middleware(dataStore, tokenManager),
		AuthHandler:          auth.NewHandler(nil),
		ProfileHandler:       profile.NewHandler(profile.NewService(dataStore)),
		MetaHandler:          meta.NewHandler(meta.NewService()),
		QuestionnaireHandler: questionnaire.NewHandler(),
		MatchHandler:         match.NewHandler(matchService),
	})

	server := httptest.NewServer(router)
	env := &matchIntegrationEnv{
		server:       server,
		client:       server.Client(),
		adminDB:      adminDB,
		db:           db,
		dbName:       dbName,
		store:        dataStore,
		tokenManager: tokenManager,
		matchService: matchService,
	}

	t.Cleanup(func() {
		server.Close()
		_ = db.Close()
		_ = dropMatchIntegrationDatabase(context.Background(), adminDB, dbName)
		_ = adminDB.Close()
	})

	return env
}

func createIntegrationUser(t *testing.T, env *matchIntegrationEnv, email, nickname, department, grade string, now time.Time) store.User {
	t.Helper()

	user, err := env.store.CreateUser(context.Background(), nil, store.CreateUserParams{
		Email:        email,
		PasswordHash: "integration-hash",
		Nickname:     nickname,
		Department:   department,
		Grade:        grade,
		Now:          now,
	})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	return user
}

func upsertContact(t *testing.T, env *matchIntegrationEnv, userID int64, contactType, contactValue string, now time.Time) {
	t.Helper()

	if _, err := env.store.UpsertUserContactMethod(context.Background(), nil, store.UpsertContactMethodParams{
		UserID:       userID,
		ContactType:  contactType,
		ContactValue: contactValue,
		CreatedAt:    now,
		UpdatedAt:    now,
	}); err != nil {
		t.Fatalf("UpsertUserContactMethod() error = %v", err)
	}
	if err := env.store.SyncUserPrimaryContact(context.Background(), nil, userID, now); err != nil {
		t.Fatalf("SyncUserPrimaryContact() error = %v", err)
	}
}

func issueIntegrationToken(t *testing.T, env *matchIntegrationEnv, userID int64, now time.Time) string {
	t.Helper()

	sessionID := uuid.NewString()
	token, expiresAt, err := env.tokenManager.Issue(userID, sessionID, now)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	if _, err := env.store.CreateSession(context.Background(), nil, store.CreateSessionParams{
		ID:         sessionID,
		UserID:     userID,
		ExpiresAt:  expiresAt,
		CreatedAt:  now,
		LastSeenAt: now,
	}); err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	return token
}

func doMatchJSON[T any](t *testing.T, env *matchIntegrationEnv, method, path, token string) (int, matchAPIEnvelope[T]) {
	t.Helper()

	req, err := http.NewRequest(method, env.server.URL+path, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
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

func matchIntegrationBaseDSN(t *testing.T) string {
	t.Helper()

	for _, key := range []string{"SOULINK_INTEGRATION_TEST_DSN", "DB_DSN"} {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}

	if value, ok := lookupMatchEnvFileValue(filepath.Join("..", "..", ".env"), "DB_DSN"); ok {
		return value
	}

	return "postgres://solink:soulink@127.0.0.1:15432/soulink-web?sslmode=disable"
}

func lookupMatchEnvFileValue(path string, key string) (string, bool) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}

	prefix := key + "="
	for _, line := range strings.Split(string(contents), "\n") {
		trimmed := strings.TrimSpace(strings.TrimPrefix(line, "\uFEFF"))
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.HasPrefix(trimmed, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, prefix)), true
		}
	}

	return "", false
}

func replaceMatchDatabaseName(t *testing.T, dsn, dbName string) string {
	t.Helper()

	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("parse dsn %q: %v", dsn, err)
	}

	parsed.Path = "/" + dbName
	return parsed.String()
}

func dropMatchIntegrationDatabase(ctx context.Context, adminDB *sql.DB, dbName string) error {
	_, err := adminDB.ExecContext(ctx, fmt.Sprintf(`DROP DATABASE IF EXISTS "%s" WITH (FORCE)`, dbName))
	return err
}
