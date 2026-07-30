package mail

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
)

// SMTPTransportConfig 描述单个 SMTP 账号的实际连接参数。
type SMTPTransportConfig struct {
	Host        string
	Port        int
	Username    string
	Password    string
	FromAddress string
}

// SendRawSMTP 根据端口自动选择 SMTPS(465) 或 SMTP/STARTTLS(其他端口) 发送邮件。
func SendRawSMTP(ctx context.Context, cfg SMTPTransportConfig, to []string, payload []byte) error {
	if strings.TrimSpace(cfg.Host) == "" {
		return fmt.Errorf("smtp host is required")
	}
	if cfg.Port <= 0 {
		return fmt.Errorf("smtp port must be greater than zero")
	}
	if strings.TrimSpace(cfg.FromAddress) == "" {
		return fmt.Errorf("smtp from address is required")
	}
	if strings.TrimSpace(cfg.Username) != "" && strings.TrimSpace(cfg.Password) == "" {
		return fmt.Errorf("smtp password is required for %s", cfg.FromAddress)
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	dialer := &net.Dialer{}
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		ServerName: cfg.Host,
	}

	var (
		conn        net.Conn
		implicitTLS bool
		err         error
	)
	if usesImplicitTLS(cfg.Port) {
		implicitTLS = true
		conn, err = (&tls.Dialer{NetDialer: dialer, Config: tlsConfig}).DialContext(ctx, "tcp", addr)
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("dial smtp server: %w", err)
	}
	defer conn.Close()

	if deadline, ok := ctx.Deadline(); ok {
		if err := conn.SetDeadline(deadline); err != nil {
			return fmt.Errorf("set smtp deadline: %w", err)
		}
	}

	stopCancel := make(chan struct{})
	defer close(stopCancel)
	go func() {
		select {
		case <-ctx.Done():
			_ = conn.Close()
		case <-stopCancel:
		}
	}()

	client, err := smtp.NewClient(conn, cfg.Host)
	if err != nil {
		return fmt.Errorf("create smtp client: %w", err)
	}
	defer client.Close()

	if !implicitTLS {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(tlsConfig); err != nil {
				return fmt.Errorf("starttls: %w", err)
			}
		}
	}

	if strings.TrimSpace(cfg.Username) != "" {
		if ok, _ := client.Extension("AUTH"); !ok {
			return fmt.Errorf("smtp auth is not supported by %s", addr)
		}
		auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	if err := client.Mail(cfg.FromAddress); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("smtp rcpt %s: %w", recipient, err)
		}
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := writer.Write(payload); err != nil {
		_ = writer.Close()
		return fmt.Errorf("write smtp payload: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("close smtp payload: %w", err)
	}

	if err := client.Quit(); err != nil {
		return fmt.Errorf("smtp quit: %w", err)
	}
	return nil
}

func usesImplicitTLS(port int) bool {
	return port == 465
}
