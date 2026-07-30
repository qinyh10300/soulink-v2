package mail

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	"soulink-backend/internal/config"
)

func TestNewSMTPSenderLoadsAccountsFromFile(t *testing.T) {
	path := writeSMTPAccountsFile(t, `{
  "fromName": "Soulink",
  "accounts": [
    {
      "fromAddress": "sender1@example.com",
      "smtpHost": "smtp.163.com",
      "smtpPort": 465,
      "smtpPassword": "password-1"
    },
    {
      "fromAddress": "sender2@example.com",
      "smtpHost": "smtp.qq.com",
      "smtpPort": 587,
      "smtpPassword": "password-2",
      "fromName": "Soulink 2"
    }
  ]
}`)

	sender, err := NewSMTPSender(config.MailConfig{
		FromName:         "Fallback",
		SMTPPort:         1025,
		SMTPAccountsFile: path,
	}, nil)
	if err != nil {
		t.Fatalf("NewSMTPSender() error: %v", err)
	}

	if len(sender.accounts) != 2 {
		t.Fatalf("len(accounts) = %d, want 2", len(sender.accounts))
	}
	if sender.accounts[0].transport.Host != "smtp.163.com" || sender.accounts[0].transport.Port != 465 {
		t.Fatalf("account[0] transport = %s:%d, want %s:%d", sender.accounts[0].transport.Host, sender.accounts[0].transport.Port, "smtp.163.com", 465)
	}
	if sender.accounts[1].transport.Host != "smtp.qq.com" || sender.accounts[1].transport.Port != 587 {
		t.Fatalf("account[1] transport = %s:%d, want %s:%d", sender.accounts[1].transport.Host, sender.accounts[1].transport.Port, "smtp.qq.com", 587)
	}
	if sender.accounts[0].transport.FromAddress != "sender1@example.com" {
		t.Fatalf("account[0].from = %q, want %q", sender.accounts[0].transport.FromAddress, "sender1@example.com")
	}
	if sender.accounts[1].name != "Soulink 2" {
		t.Fatalf("account[1].name = %q, want %q", sender.accounts[1].name, "Soulink 2")
	}
}

func TestSMTPSenderSendRotatesAccountsRoundRobin(t *testing.T) {
	path := writeSMTPAccountsFile(t, `{
  "fromName": "Soulink",
  "accounts": [
    {
      "fromAddress": "sender1@example.com",
      "smtpHost": "smtp.163.com",
      "smtpPort": 465,
      "smtpPassword": "password-1"
    },
    {
      "fromAddress": "sender2@example.com",
      "smtpHost": "smtp.qq.com",
      "smtpPort": 587,
      "smtpPassword": "password-2"
    }
  ]
}`)

	sender, err := NewSMTPSender(config.MailConfig{
		SMTPAccountsFile: path,
	}, nil)
	if err != nil {
		t.Fatalf("NewSMTPSender() error: %v", err)
	}

	var usedTransports []string
	var senders []string
	sender.sendMessage = func(ctx context.Context, cfg SMTPTransportConfig, to []string, payload []byte) error {
		usedTransports = append(usedTransports, cfg.Host+":"+strconv.Itoa(cfg.Port))
		senders = append(senders, cfg.FromAddress)
		return nil
	}

	for i := 0; i < 3; i++ {
		if err := sender.Send(context.Background(), Message{
			To:      "receiver@example.com",
			Subject: "test",
			Text:    "hello",
		}); err != nil {
			t.Fatalf("Send() error: %v", err)
		}
	}

	wantTransports := []string{"smtp.163.com:465", "smtp.qq.com:587", "smtp.163.com:465"}
	if !reflect.DeepEqual(usedTransports, wantTransports) {
		t.Fatalf("usedTransports = %#v, want %#v", usedTransports, wantTransports)
	}

	wantSenders := []string{"sender1@example.com", "sender2@example.com", "sender1@example.com"}
	if !reflect.DeepEqual(senders, wantSenders) {
		t.Fatalf("senders = %#v, want %#v", senders, wantSenders)
	}
}

func TestUsesImplicitTLS(t *testing.T) {
	if !usesImplicitTLS(465) {
		t.Fatalf("usesImplicitTLS(465) = false, want true")
	}
	if usesImplicitTLS(587) {
		t.Fatalf("usesImplicitTLS(587) = true, want false")
	}
}

func TestNewSMTPSenderFallsBackToSingleAccountConfig(t *testing.T) {
	sender, err := NewSMTPSender(config.MailConfig{
		FromName:     "Soulink",
		FromAddress:  "single@example.com",
		SMTPHost:     "smtp.example.com",
		SMTPPort:     587,
		SMTPUsername: "single@example.com",
		SMTPPassword: "password",
	}, nil)
	if err != nil {
		t.Fatalf("NewSMTPSender() error: %v", err)
	}

	if len(sender.accounts) != 1 {
		t.Fatalf("len(accounts) = %d, want 1", len(sender.accounts))
	}
	if sender.accounts[0].transport.FromAddress != "single@example.com" {
		t.Fatalf("account[0].from = %q, want %q", sender.accounts[0].transport.FromAddress, "single@example.com")
	}
}

func writeSMTPAccountsFile(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "smtp_accounts.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write smtp accounts file: %v", err)
	}
	return path
}

func TestSMTPSenderRetriesWithNextAccountOnNonRecipientError(t *testing.T) {
	path := writeSMTPAccountsFile(t, `{
  "accounts": [
    {
      "fromAddress": "sender1@example.com",
      "smtpHost": "smtp1.example.com",
      "smtpPort": 465,
      "smtpPassword": "password-1"
    },
    {
      "fromAddress": "sender2@example.com",
      "smtpHost": "smtp2.example.com",
      "smtpPort": 587,
      "smtpPassword": "password-2"
    }
  ]
}`)

	sender, err := NewSMTPSender(config.MailConfig{
		SMTPAccountsFile: path,
		SMTPMaxAttempts:  3,
	}, nil)
	if err != nil {
		t.Fatalf("NewSMTPSender() error: %v", err)
	}

	var attempts []string
	sender.sendMessage = func(ctx context.Context, cfg SMTPTransportConfig, to []string, payload []byte) error {
		attempts = append(attempts, cfg.FromAddress)
		if len(attempts) == 1 {
			return strconv.ErrSyntax
		}
		return nil
	}

	if err := sender.Send(context.Background(), Message{
		To:      "receiver@example.com",
		Subject: "retry",
		Text:    "hello",
	}); err != nil {
		t.Fatalf("Send() error: %v", err)
	}

	wantAttempts := []string{"sender1@example.com", "sender2@example.com"}
	if !reflect.DeepEqual(attempts, wantAttempts) {
		t.Fatalf("attempts = %#v, want %#v", attempts, wantAttempts)
	}
}

func TestSMTPSenderDoesNotRetryWhenRecipientNotFound(t *testing.T) {
	path := writeSMTPAccountsFile(t, `{
  "accounts": [
    {
      "fromAddress": "sender1@example.com",
      "smtpHost": "smtp1.example.com",
      "smtpPort": 465,
      "smtpPassword": "password-1"
    },
    {
      "fromAddress": "sender2@example.com",
      "smtpHost": "smtp2.example.com",
      "smtpPort": 587,
      "smtpPassword": "password-2"
    }
  ]
}`)

	sender, err := NewSMTPSender(config.MailConfig{
		SMTPAccountsFile: path,
		SMTPMaxAttempts:  3,
	}, nil)
	if err != nil {
		t.Fatalf("NewSMTPSender() error: %v", err)
	}

	attemptCount := 0
	sender.sendMessage = func(ctx context.Context, cfg SMTPTransportConfig, to []string, payload []byte) error {
		attemptCount++
		return fmt.Errorf("smtp rcpt receiver@example.com: 550 5.1.1 user unknown")
	}

	if err := sender.Send(context.Background(), Message{
		To:      "receiver@example.com",
		Subject: "no-retry",
		Text:    "hello",
	}); err == nil {
		t.Fatal("Send() error = nil, want non-nil")
	}

	if attemptCount != 1 {
		t.Fatalf("attemptCount = %d, want 1", attemptCount)
	}
}
