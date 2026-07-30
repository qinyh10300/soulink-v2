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
