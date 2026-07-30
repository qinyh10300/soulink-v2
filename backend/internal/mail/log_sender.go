package mail

import (
	"context"
	"log/slog"

	"soulink-backend/internal/config"
)

type LogSender struct {
	logger *slog.Logger
	cfg    config.MailConfig
}

// Send 将邮件内容写入日志，便于本地开发调试。
func (s *LogSender) Send(_ context.Context, message Message) error {
	logger := s.logger
	if logger == nil {
		logger = slog.Default()
	}

	logger.Info("mail message",
		"driver", "log",
		"from", s.cfg.FromAddress,
		"to", message.To,
		"subject", message.Subject,
		"text", message.Text,
	)
	return nil
}
