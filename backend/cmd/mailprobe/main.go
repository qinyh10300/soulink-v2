package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"mime"
	"os"
	"strings"
	"time"

	"soulink-backend/internal/config"
	"soulink-backend/internal/mail"
)

type accountFile struct {
	FromName string        `json:"fromName"`
	SMTPHost string        `json:"smtpHost"`
	SMTPPort int           `json:"smtpPort"`
	Accounts []smtpAccount `json:"accounts"`
}

type smtpAccount struct {
	FromName     string `json:"fromName"`
	FromAddress  string `json:"fromAddress"`
	SMTPHost     string `json:"smtpHost"`
	SMTPPort     int    `json:"smtpPort"`
	SMTPUsername string `json:"smtpUsername"`
	SMTPPassword string `json:"smtpPassword"`
}

type probeAccount struct {
	transport mail.SMTPTransportConfig
	fromName  string
}

func main() {
	var (
		to      = flag.String("to", "", "测试收件邮箱")
		subject = flag.String("subject", "Soulink SMTP 逐账号测试", "测试邮件主题前缀")
		timeout = flag.Duration("timeout", 20*time.Second, "单封邮件发送超时")
	)
	flag.Parse()

	if strings.TrimSpace(*to) == "" {
		fmt.Fprintln(os.Stderr, "missing required flag: -to")
		os.Exit(2)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}
	if strings.ToLower(cfg.Mail.Driver) != "smtp" {
		fmt.Fprintln(os.Stderr, "MAIL_DRIVER is not smtp")
		os.Exit(1)
	}

	accounts, err := loadAccounts(cfg.Mail)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load smtp accounts: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("开始逐个账号发送测试邮件，共 %d 个账号，收件人=%s\n", len(accounts), *to)

	var failed bool
	for i, account := range accounts {
		ctx, cancel := context.WithTimeout(context.Background(), *timeout)
		subj := fmt.Sprintf("%s [%d/%d] %s", *subject, i+1, len(accounts), time.Now().Format(time.RFC3339))
		body := strings.Join([]string{
			"这是一封 Soulink SMTP 配置测试邮件。",
			fmt.Sprintf("发件账号：%s", account.transport.FromAddress),
			fmt.Sprintf("发送时间：%s", time.Now().Format(time.RFC3339)),
		}, "\n")
		payload := buildPayload(account, *to, subj, body)
		err := mail.SendRawSMTP(ctx, account.transport, []string{*to}, payload)
		cancel()
		if err != nil {
			failed = true
			fmt.Printf("[%d/%d] 失败 from=%s host=%s:%d error=%v\n", i+1, len(accounts), account.transport.FromAddress, account.transport.Host, account.transport.Port, err)
			continue
		}
		fmt.Printf("[%d/%d] 成功 from=%s host=%s:%d\n", i+1, len(accounts), account.transport.FromAddress, account.transport.Host, account.transport.Port)
	}

	if failed {
		os.Exit(1)
	}
}

func loadAccounts(cfg config.MailConfig) ([]probeAccount, error) {
	if strings.TrimSpace(cfg.SMTPAccountsFile) != "" {
		contents, err := os.ReadFile(cfg.SMTPAccountsFile)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", cfg.SMTPAccountsFile, err)
		}
		var file accountFile
		if err := json.Unmarshal(contents, &file); err != nil {
			return nil, fmt.Errorf("parse %s: %w", cfg.SMTPAccountsFile, err)
		}
		if len(file.Accounts) == 0 {
			return nil, fmt.Errorf("smtp accounts file %s has no accounts", cfg.SMTPAccountsFile)
		}
		accounts := make([]probeAccount, 0, len(file.Accounts))
		for i, account := range file.Accounts {
			resolved := probeAccount{
				transport: mail.SMTPTransportConfig{
					FromAddress: firstNonEmpty(account.FromAddress, cfg.FromAddress),
					Host:        firstNonEmpty(account.SMTPHost, file.SMTPHost, cfg.SMTPHost),
					Port:        firstNonZero(account.SMTPPort, file.SMTPPort, cfg.SMTPPort),
					Username:    firstNonEmpty(account.SMTPUsername, cfg.SMTPUsername, account.FromAddress),
					Password:    firstNonEmpty(account.SMTPPassword, cfg.SMTPPassword),
				},
				fromName: firstNonEmpty(account.FromName, file.FromName, cfg.FromName),
			}
			if err := validateAccount(resolved.transport); err != nil {
				return nil, fmt.Errorf("account %d: %w", i+1, err)
			}
			accounts = append(accounts, resolved)
		}
		return accounts, nil
	}

	account := probeAccount{
		transport: mail.SMTPTransportConfig{
			FromAddress: cfg.FromAddress,
			Host:        cfg.SMTPHost,
			Port:        cfg.SMTPPort,
			Username:    cfg.SMTPUsername,
			Password:    cfg.SMTPPassword,
		},
		fromName: cfg.FromName,
	}
	if err := validateAccount(account.transport); err != nil {
		return nil, err
	}
	return []probeAccount{account}, nil
}

func validateAccount(account mail.SMTPTransportConfig) error {
	if strings.TrimSpace(account.FromAddress) == "" {
		return fmt.Errorf("fromAddress is required")
	}
	if strings.TrimSpace(account.Host) == "" {
		return fmt.Errorf("smtpHost is required")
	}
	if account.Port <= 0 {
		return fmt.Errorf("smtpPort must be greater than zero")
	}
	if strings.TrimSpace(account.Username) != "" && strings.TrimSpace(account.Password) == "" {
		return fmt.Errorf("smtpPassword is required for %s", account.FromAddress)
	}
	return nil
}

func buildPayload(account probeAccount, to, subject, body string) []byte {
	payload := strings.Join([]string{
		buildFromHeader(account),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", mime.QEncoding.Encode("utf-8", subject)),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")
	return []byte(payload)
}

func buildFromHeader(account probeAccount) string {
	if strings.TrimSpace(account.fromName) == "" {
		return fmt.Sprintf("From: %s", account.transport.FromAddress)
	}
	return fmt.Sprintf("From: %s <%s>", mime.QEncoding.Encode("utf-8", account.fromName), account.transport.FromAddress)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func firstNonZero(values ...int) int {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}
