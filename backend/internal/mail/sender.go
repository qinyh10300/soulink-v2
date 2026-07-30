package mail

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"soulink-backend/internal/config"
)

type Message struct {
	To      string
	Subject string
	Text    string
}

type Sender interface {
	Send(ctx context.Context, message Message) error
}

// NewSender 按配置创建邮件发送器实现。
func NewSender(cfg config.MailConfig, logger *slog.Logger) (Sender, error) {
	switch strings.ToLower(cfg.Driver) {
	case "", "log", "stdout", "noop":
		return &LogSender{
			logger: logger,
			cfg:    cfg,
		}, nil
	case "smtp":
		return NewSMTPSender(cfg, logger)
	default:
		return nil, fmt.Errorf("unsupported mail driver: %s", cfg.Driver)
	}
}
