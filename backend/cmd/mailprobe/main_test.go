package main

import (
	"mime"
	"strings"
	"testing"

	"soulink-backend/internal/mail"
)

func TestBuildPayloadIncludesFromName(t *testing.T) {
	account := probeAccount{
		transport: mail.SMTPTransportConfig{
			FromAddress: "sender@example.com",
		},
		fromName: "测试发件人",
	}

	payload := string(buildPayload(account, "receiver@example.com", "测试主题", "hello"))
	fromHeader := "From: " + mime.QEncoding.Encode("utf-8", "测试发件人") + " <sender@example.com>"
	if !strings.Contains(payload, fromHeader) {
		t.Fatalf("payload does not contain expected From header: %q", fromHeader)
	}
}
