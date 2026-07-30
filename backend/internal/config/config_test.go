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
