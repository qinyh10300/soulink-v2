package mail

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"mime"
	"os"
	"strings"
	"sync/atomic"

	"soulink-backend/internal/config"
)

type smtpAccount struct {
	transport SMTPTransportConfig
	name      string
}

type smtpAccountFile struct {
	FromName string              `json:"fromName"`
	SMTPHost string              `json:"smtpHost"`
	SMTPPort int                 `json:"smtpPort"`
	Accounts []smtpAccountConfig `json:"accounts"`
}

type smtpAccountConfig struct {
	FromName     string `json:"fromName"`
	FromAddress  string `json:"fromAddress"`
	SMTPHost     string `json:"smtpHost"`
	SMTPPort     int    `json:"smtpPort"`
	SMTPUsername string `json:"smtpUsername"`
	SMTPPassword string `json:"smtpPassword"`
}

type SMTPSender struct {
	next        uint32
	accounts    []smtpAccount
	maxAttempts int
	logger      *slog.Logger
	sendMessage func(context.Context, SMTPTransportConfig, []string, []byte) error
}

// NewSMTPSender creates an SMTP sender that can load multiple accounts from a config file.
func NewSMTPSender(cfg config.MailConfig, logger *slog.Logger) (*SMTPSender, error) {
	accounts, err := buildSMTPAccounts(cfg)
	if err != nil {
		return nil, err
	}

	maxAttempts := cfg.SMTPMaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 1
	}
	if logger == nil {
		logger = slog.Default()
	}

	return &SMTPSender{
		accounts:    accounts,
		maxAttempts: maxAttempts,
		logger:      logger,
		sendMessage: SendRawSMTP,
	}, nil
}

func buildSMTPAccounts(cfg config.MailConfig) ([]smtpAccount, error) {
	if path := strings.TrimSpace(cfg.SMTPAccountsFile); path != "" {
		return loadSMTPAccountsFromFile(path, cfg)
	}

	account, err := buildSMTPAccount(smtpAccountConfig{
		FromName:     cfg.FromName,
		FromAddress:  cfg.FromAddress,
		SMTPHost:     cfg.SMTPHost,
		SMTPPort:     cfg.SMTPPort,
		SMTPUsername: cfg.SMTPUsername,
		SMTPPassword: cfg.SMTPPassword,
	}, cfg, smtpAccountFile{}, false)
	if err != nil {
		return nil, err
	}
	return []smtpAccount{account}, nil
}

func loadSMTPAccountsFromFile(path string, cfg config.MailConfig) ([]smtpAccount, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read smtp accounts file %s: %w", path, err)
	}

	var file smtpAccountFile
	if err := json.Unmarshal(contents, &file); err != nil {
		return nil, fmt.Errorf("parse smtp accounts file %s: %w", path, err)
	}
	if len(file.Accounts) == 0 {
		return nil, fmt.Errorf("smtp accounts file %s: accounts is empty", path)
	}

	accounts := make([]smtpAccount, 0, len(file.Accounts))
	for i, accountCfg := range file.Accounts {
		account, err := buildSMTPAccount(accountCfg, cfg, file, true)
		if err != nil {
			return nil, fmt.Errorf("smtp accounts file %s account %d: %w", path, i+1, err)
		}
		accounts = append(accounts, account)
	}
	return accounts, nil
}

func buildSMTPAccount(accountCfg smtpAccountConfig, cfg config.MailConfig, file smtpAccountFile, fromFile bool) (smtpAccount, error) {
	host := firstNonEmpty(accountCfg.SMTPHost, file.SMTPHost, cfg.SMTPHost)
	if host == "" {
		return smtpAccount{}, fmt.Errorf("smtp host is required")
	}

	port := firstNonZero(accountCfg.SMTPPort, file.SMTPPort, cfg.SMTPPort)
	if port <= 0 {
		return smtpAccount{}, fmt.Errorf("smtp port must be greater than zero")
	}

	from := firstNonEmpty(accountCfg.FromAddress, cfg.FromAddress)
	if from == "" {
		return smtpAccount{}, fmt.Errorf("mail from address is required")
	}

	name := firstNonEmpty(accountCfg.FromName, file.FromName, cfg.FromName)

	username := strings.TrimSpace(accountCfg.SMTPUsername)
	if username == "" && fromFile {
		username = from
	}
	if username == "" {
		username = strings.TrimSpace(cfg.SMTPUsername)
	}

	password := firstNonEmpty(accountCfg.SMTPPassword, cfg.SMTPPassword)
	if username != "" && strings.TrimSpace(password) == "" {
		return smtpAccount{}, fmt.Errorf("smtp password is required for %s", from)
	}

	return smtpAccount{
		transport: SMTPTransportConfig{
			Host:        host,
			Port:        port,
			Username:    username,
			Password:    password,
			FromAddress: from,
		},
		name: name,
	}, nil
}

// Send sends a plain text email and rotates across configured SMTP accounts.
func (s *SMTPSender) Send(ctx context.Context, message Message) error {
	var lastErr error

	for attempt := 1; attempt <= s.maxAttempts; attempt++ {
		account := s.pickAccount()
		payload := strings.Join([]string{
			buildFromHeader(account),
			fmt.Sprintf("To: %s", message.To),
			fmt.Sprintf("Subject: %s", mime.QEncoding.Encode("utf-8", message.Subject)),
			"MIME-Version: 1.0",
			"Content-Type: text/plain; charset=UTF-8",
			"",
			message.Text,
		}, "\r\n")

		if err := s.sendMessage(ctx, account.transport, []string{message.To}, []byte(payload)); err != nil {
			lastErr = err
			recipientNotFound := isRecipientNotFoundError(err)
			s.logger.Error(
				"smtp send failed",
				"attempt", attempt,
				"max_attempts", s.maxAttempts,
				"from", account.transport.FromAddress,
				"smtp_host", account.transport.Host,
				"smtp_port", account.transport.Port,
				"to", message.To,
				"recipient_not_found", recipientNotFound,
				"error", err,
			)
			if recipientNotFound {
				break
			}
			continue
		}
		return nil
	}

	return fmt.Errorf("send smtp mail: %w", lastErr)
}

func (s *SMTPSender) pickAccount() smtpAccount {
	if len(s.accounts) == 1 {
		return s.accounts[0]
	}

	idx := atomic.AddUint32(&s.next, 1) - 1
	return s.accounts[int(idx)%len(s.accounts)]
}

func buildFromHeader(account smtpAccount) string {
	if strings.TrimSpace(account.name) == "" {
		return fmt.Sprintf("From: %s", account.transport.FromAddress)
	}
	return fmt.Sprintf("From: %s <%s>", mime.QEncoding.Encode("utf-8", account.name), account.transport.FromAddress)
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

func isRecipientNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	message := strings.ToLower(err.Error())
	patterns := []string{
		"550 5.1.1",
		"550 user unknown",
		"user unknown",
		"unknown user",
		"no such user",
		"mailbox unavailable",
		"recipient address rejected",
		"unknown local part",
		"invalid recipient",
		"recipient rejected",
		"recipient not found",
	}
	for _, pattern := range patterns {
		if strings.Contains(message, pattern) {
			return true
		}
	}
	return false
}
