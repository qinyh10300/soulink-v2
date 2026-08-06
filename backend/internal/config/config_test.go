package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadEnvFileOverridesExistingEnvironment(t *testing.T) {
	t.Setenv("APP_ENV", "from-env")
	t.Setenv("HTTP_ADDR", ":9999")
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("APP_ENV=from-dotenv\nHTTP_ADDR=:8088\n"), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.AppEnv != "from-dotenv" {
		t.Fatalf("AppEnv = %q, want %q", cfg.AppEnv, "from-dotenv")
	}
	if cfg.HTTPAddr != ":8088" {
		t.Fatalf("HTTPAddr = %q, want %q", cfg.HTTPAddr, ":8088")
	}
}

func TestLoadEnvFileIgnoresMissingFile(t *testing.T) {
	dir := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("HTTPAddr = %q, want default %q", cfg.HTTPAddr, ":8080")
	}
}

func TestLoadReadsSMTPAccountsFile(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("SMTP_ACCOUNTS_FILE=configs/mail/smtp_accounts.local.json\nSMTP_MAX_ATTEMPTS=5\n"), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Mail.SMTPAccountsFile != "configs/mail/smtp_accounts.local.json" {
		t.Fatalf("SMTPAccountsFile = %q, want %q", cfg.Mail.SMTPAccountsFile, "configs/mail/smtp_accounts.local.json")
	}
	if cfg.Mail.SMTPMaxAttempts != 5 {
		t.Fatalf("SMTPMaxAttempts = %d, want %d", cfg.Mail.SMTPMaxAttempts, 5)
	}
}

func TestLoadReadsRedisConfig(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	contents := "REDIS_ENABLED=true\nREDIS_ADDR=127.0.0.1:16379\nREDIS_PASSWORD=secret\nREDIS_DB=2\nREDIS_PREFIX=soulink:test:\n"
	if err := os.WriteFile(envPath, []byte(contents), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if !cfg.Redis.Enabled {
		t.Fatal("Redis.Enabled = false, want true")
	}
	if cfg.Redis.Addr != "127.0.0.1:16379" {
		t.Fatalf("Redis.Addr = %q, want %q", cfg.Redis.Addr, "127.0.0.1:16379")
	}
	if cfg.Redis.Password != "secret" {
		t.Fatalf("Redis.Password = %q, want %q", cfg.Redis.Password, "secret")
	}
	if cfg.Redis.DB != 2 {
		t.Fatalf("Redis.DB = %d, want %d", cfg.Redis.DB, 2)
	}
	if cfg.Redis.Prefix != "soulink:test:" {
		t.Fatalf("Redis.Prefix = %q, want %q", cfg.Redis.Prefix, "soulink:test:")
	}
}

func TestLoadReadsMatchConfig(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	contents := "MATCH_RESULT_CACHE_TTL=36h\n"
	if err := os.WriteFile(envPath, []byte(contents), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Match.ResultCacheTTL != 36*time.Hour {
		t.Fatalf("Match.ResultCacheTTL = %s, want %s", cfg.Match.ResultCacheTTL, 36*time.Hour)
	}
}

func TestLoadReadsAIConfig(t *testing.T) {
	preserveEnvironment(t, "AI_PROVIDER", "AI_BASE_URL", "AI_API_KEY", "AI_MODEL", "AI_TIMEOUT", "AI_MAX_RETRIES")
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	contents := "AI_PROVIDER=openai-compatible\nAI_BASE_URL=https://model.example/v1\nAI_API_KEY=test-key\nAI_MODEL=test-model\nAI_TIMEOUT=4s\nAI_MAX_RETRIES=1\n"
	if err := os.WriteFile(envPath, []byte(contents), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.AI.Provider != "openai-compatible" || cfg.AI.BaseURL != "https://model.example/v1" {
		t.Fatalf("AI provider config = %#v", cfg.AI)
	}
	if cfg.AI.APIKey != "test-key" || cfg.AI.Model != "test-model" {
		t.Fatalf("AI credentials/model config = %#v", cfg.AI)
	}
	if cfg.AI.Timeout != 4*time.Second || cfg.AI.MaxRetries != 1 {
		t.Fatalf("AI reliability config = %#v", cfg.AI)
	}
}

func TestLoadRejectsInvalidAIRetries(t *testing.T) {
	preserveEnvironment(t, "AI_MAX_RETRIES")
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("AI_MAX_RETRIES=9\n"), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted AI_MAX_RETRIES outside the supported range")
	}
}

func preserveEnvironment(t *testing.T, keys ...string) {
	t.Helper()
	type value struct {
		text string
		set  bool
	}
	previous := make(map[string]value, len(keys))
	for _, key := range keys {
		text, set := os.LookupEnv(key)
		previous[key] = value{text: text, set: set}
	}
	t.Cleanup(func() {
		for _, key := range keys {
			item := previous[key]
			if item.set {
				_ = os.Setenv(key, item.text)
			} else {
				_ = os.Unsetenv(key)
			}
		}
	})
}

func TestLoadReadsQuestionnaireConfig(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	contents := "QUESTIONNAIRE_CATALOG_FILE=configs/questionnaire/custom_template.md\nSOULAGENT_ENABLED=false\n"
	if err := os.WriteFile(envPath, []byte(contents), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Questionnaire.CatalogFile != "configs/questionnaire/custom_template.md" {
		t.Fatalf("Questionnaire.CatalogFile = %q, want %q", cfg.Questionnaire.CatalogFile, "configs/questionnaire/custom_template.md")
	}
	if cfg.Questionnaire.SoulAgentEnabled {
		t.Fatal("Questionnaire.SoulAgentEnabled = true, want false")
	}
}

func TestLoadReadsSafetyAdminEmails(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	contents := "ADMIN_EMAILS=Admin@Mails.Tsinghua.Edu.Cn, ops@tsinghua.edu.cn\n"
	if err := os.WriteFile(envPath, []byte(contents), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if len(cfg.Safety.AdminEmails) != 2 {
		t.Fatalf("len(Safety.AdminEmails) = %d, want 2", len(cfg.Safety.AdminEmails))
	}
	if cfg.Safety.AdminEmails[0] != "admin@mails.tsinghua.edu.cn" {
		t.Fatalf("Safety.AdminEmails[0] = %q", cfg.Safety.AdminEmails[0])
	}
}

func TestLoadReadsDemoConfig(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	contents := "DEMO_SEED_ENABLED=true\nDEMO_SEED_PASSWORD=local-password\nDEMO_ADMIN_EMAIL=Hack@Mails.Tsinghua.Edu.Cn\nDEMO_ADMIN_PASSWORD=local-admin-password\n"
	if err := os.WriteFile(envPath, []byte(contents), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if !cfg.Demo.Enabled {
		t.Fatal("Demo.Enabled = false, want true")
	}
	if cfg.Demo.Password != "local-password" {
		t.Fatalf("Demo.Password = %q", cfg.Demo.Password)
	}
	if cfg.Demo.AdminEmail != "hack@mails.tsinghua.edu.cn" {
		t.Fatalf("Demo.AdminEmail = %q", cfg.Demo.AdminEmail)
	}
	if cfg.Demo.AdminPassword != "local-admin-password" {
		t.Fatalf("Demo.AdminPassword = %q", cfg.Demo.AdminPassword)
	}
}
