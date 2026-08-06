package demoseed

import (
	"strings"
	"testing"
)

func TestEnsureRejectsInvalidPasswordBeforeHashing(t *testing.T) {
	_, err := Ensure(t.Context(), nil, Config{Password: "short", BCryptCost: 10})
	if err == nil || !strings.Contains(err.Error(), "demo password") {
		t.Fatalf("Ensure() error = %v, want demo password validation error", err)
	}
}

func TestEnsureRejectsNilStore(t *testing.T) {
	_, err := Ensure(t.Context(), nil, Config{
		Password:   "valid-password",
		BCryptCost: 10,
	})
	if err == nil || !strings.Contains(err.Error(), "store is nil") {
		t.Fatalf("Ensure() error = %v, want nil store error", err)
	}
}

func TestEnsureRejectsPartialAdminCredentials(t *testing.T) {
	_, err := Ensure(t.Context(), nil, Config{
		Password:   "valid-password",
		AdminEmail: "hack@mails.tsinghua.edu.cn",
		BCryptCost: 10,
	})
	if err == nil || !strings.Contains(err.Error(), "configured together") {
		t.Fatalf("Ensure() error = %v, want paired admin credentials error", err)
	}
}

func TestEnsureRejectsUnsupportedAdminEmail(t *testing.T) {
	_, err := Ensure(t.Context(), nil, Config{
		Password:      "valid-password",
		AdminEmail:    "hack@example.com",
		AdminPassword: "valid-admin-password",
		BCryptCost:    10,
	})
	if err == nil || !strings.Contains(err.Error(), "Tsinghua domain") {
		t.Fatalf("Ensure() error = %v, want Tsinghua email error", err)
	}
}

func TestEnsureRejectsInvalidAdminPassword(t *testing.T) {
	_, err := Ensure(t.Context(), nil, Config{
		Password:      "valid-password",
		AdminEmail:    "hack@mails.tsinghua.edu.cn",
		AdminPassword: "short",
		BCryptCost:    10,
	})
	if err == nil || !strings.Contains(err.Error(), "admin password") {
		t.Fatalf("Ensure() error = %v, want admin password error", err)
	}
}
