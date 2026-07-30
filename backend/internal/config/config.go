package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv         string
	HTTPAddr       string
	AllowedOrigins []string
	Database       DatabaseConfig
	Redis          RedisConfig
	Auth           AuthConfig
	Mail           MailConfig
	Match          MatchConfig
	Questionnaire  QuestionnaireConfig
}

type DatabaseConfig struct {
	DSN         string
	AutoMigrate bool
}

type RedisConfig struct {
	Enabled  bool
	Addr     string
	Password string
	DB       int
	Prefix   string
}

type AuthConfig struct {
	JWTSecret                string
	JWTTTL                   time.Duration
	VerificationCodeTTL      time.Duration
	VerificationCodeCooldown time.Duration
	BCryptCost               int
}

type MailConfig struct {
	Driver           string
	FromName         string
	FromAddress      string
	SMTPHost         string
	SMTPPort         int
	SMTPUsername     string
	SMTPPassword     string
	SMTPAccountsFile string
	SMTPMaxAttempts  int
}

type QuestionnaireConfig struct {
	CatalogFile      string
	SoulAgentEnabled bool
}

type MatchConfig struct {
	ResultCacheTTL time.Duration
}

func Load() (Config, error) {
	if err := loadDotEnvFile(".env"); err != nil {
		return Config{}, err
	}

	jwtTTL, err := envDuration("JWT_TTL", 72*time.Hour)
	if err != nil {
		return Config{}, err
	}
	codeTTL, err := envDuration("VERIFICATION_CODE_TTL", 5*time.Minute)
	if err != nil {
		return Config{}, err
	}
	codeCooldown, err := envDuration("VERIFICATION_CODE_COOLDOWN", 60*time.Second)
	if err != nil {
		return Config{}, err
	}
	bcryptCost, err := envInt("BCRYPT_COST", 12)
	if err != nil {
		return Config{}, err
	}
	smtpPort, err := envInt("SMTP_PORT", 1025)
	if err != nil {
		return Config{}, err
	}
	smtpMaxAttempts, err := envInt("SMTP_MAX_ATTEMPTS", 3)
	if err != nil {
		return Config{}, err
	}
	redisDB, err := envInt("REDIS_DB", 0)
	if err != nil {
		return Config{}, err
	}
	autoMigrate, err := envBool("DB_AUTO_MIGRATE", true)
	if err != nil {
		return Config{}, err
	}
	redisEnabled, err := envBool("REDIS_ENABLED", false)
	if err != nil {
		return Config{}, err
	}
	resultCacheTTL, err := envDuration("MATCH_RESULT_CACHE_TTL", 24*time.Hour)
	if err != nil {
		return Config{}, err
	}
	soulAgentEnabled, err := envBool("SOULAGENT_ENABLED", true)
	if err != nil {
		return Config{}, err
	}

	return Config{
		AppEnv:         envString("APP_ENV", "development"),
		HTTPAddr:       envString("HTTP_ADDR", ":8080"),
		AllowedOrigins: splitCSV(envString("ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:3000")),
		Database: DatabaseConfig{
			DSN:         envString("DB_DSN", "postgres://postgres:postgres@localhost:5432/soulink?sslmode=disable"),
			AutoMigrate: autoMigrate,
		},
		Redis: RedisConfig{
			Enabled:  redisEnabled,
			Addr:     envString("REDIS_ADDR", "127.0.0.1:16379"),
			Password: envString("REDIS_PASSWORD", ""),
			DB:       redisDB,
			Prefix:   envString("REDIS_PREFIX", "soulink:"),
		},
		Auth: AuthConfig{
			JWTSecret:                envString("JWT_SECRET", "change-me-in-production"),
			JWTTTL:                   jwtTTL,
			VerificationCodeTTL:      codeTTL,
			VerificationCodeCooldown: codeCooldown,
			BCryptCost:               bcryptCost,
		},
		Mail: MailConfig{
			Driver:           strings.ToLower(envString("MAIL_DRIVER", "log")),
			FromName:         envString("MAIL_FROM_NAME", "Soulink"),
			FromAddress:      envString("MAIL_FROM_ADDRESS", "noreply@soulink.local"),
			SMTPHost:         envString("SMTP_HOST", "localhost"),
			SMTPPort:         smtpPort,
			SMTPUsername:     envString("SMTP_USERNAME", ""),
			SMTPPassword:     envString("SMTP_PASSWORD", ""),
			SMTPAccountsFile: envString("SMTP_ACCOUNTS_FILE", ""),
			SMTPMaxAttempts:  smtpMaxAttempts,
		},
		Questionnaire: QuestionnaireConfig{
			CatalogFile:      envString("QUESTIONNAIRE_CATALOG_FILE", "thu_questionnaire_template.md"),
			SoulAgentEnabled: soulAgentEnabled,
		},
		Match: MatchConfig{
			ResultCacheTTL: resultCacheTTL,
		},
	}, nil
}

func loadDotEnvFile(path string) error {
	contents, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read %s: %w", path, err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(contents)))
	for lineNo := 1; scanner.Scan(); lineNo++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return fmt.Errorf("parse %s line %d: missing '='", path, lineNo)
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" {
			return fmt.Errorf("parse %s line %d: empty key", path, lineNo)
		}
		parsedValue, err := parseDotEnvValue(value)
		if err != nil {
			return fmt.Errorf("parse %s line %d: %w", path, lineNo, err)
		}
		if err := os.Setenv(key, parsedValue); err != nil {
			return fmt.Errorf("set env %s from %s: %w", key, path, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan %s: %w", path, err)
	}
	return nil
}

func parseDotEnvValue(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	if strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
		parsed, err := strconv.Unquote(value)
		if err != nil {
			return "", err
		}
		return parsed, nil
	}
	if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
		return strings.TrimSuffix(strings.TrimPrefix(value, "'"), "'"), nil
	}
	return value, nil
}

func envString(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return parsed, nil
}

func envBool(key string, fallback bool) (bool, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("parse %s: %w", key, err)
	}
	return parsed, nil
}

func envDuration(key string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return parsed, nil
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}
