package auth

import (
	"testing"
	"time"
)

func TestTokenManagerIssueAndParse(t *testing.T) {
	t.Parallel()

	manager := NewTokenManager("secret", time.Hour)
	now := time.Now().UTC()

	token, expiresAt, err := manager.Issue(42, "session-1", now)
	if err != nil {
		t.Fatalf("Issue() unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("Issue() returned empty token")
	}
	if !expiresAt.After(now) {
		t.Fatal("Issue() returned invalid expiration time")
	}

	claims, err := manager.Parse(token)
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}
	if claims.Subject != "42" {
		t.Fatalf("Parse() subject = %q, want %q", claims.Subject, "42")
	}
	if claims.ID != "session-1" {
		t.Fatalf("Parse() id = %q, want %q", claims.ID, "session-1")
	}
}

func TestTokenManagerRejectsInvalidToken(t *testing.T) {
	t.Parallel()

	manager := NewTokenManager("secret", time.Hour)
	if _, err := manager.Parse("invalid-token"); err == nil {
		t.Fatal("Parse() expected error for invalid token")
	}
}
