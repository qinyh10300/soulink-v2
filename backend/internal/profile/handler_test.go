package profile

import (
	"testing"
	"time"

	"soulink-backend/internal/store"
)

func TestContactsForResponseFallsBackToReadonlyEmail(t *testing.T) {
	now := time.Date(2026, 3, 26, 10, 0, 0, 0, time.UTC)
	details := Details{
		User: store.User{
			Email:     "fallback@example.com",
			CreatedAt: now,
		},
	}

	contacts := contactsForResponse(details)
	if len(contacts) != 1 {
		t.Fatalf("contacts len = %d, want 1", len(contacts))
	}
	if contacts[0].Type != "email" || contacts[0].Value != "fallback@example.com" {
		t.Fatalf("unexpected fallback contact: %#v", contacts[0])
	}
	if !contacts[0].Readonly {
		t.Fatalf("fallback contact should be readonly: %#v", contacts[0])
	}
	if contacts[0].UpdatedAt != now.Format(time.RFC3339) {
		t.Fatalf("fallback contact updatedAt = %q, want %q", contacts[0].UpdatedAt, now.Format(time.RFC3339))
	}
}

func TestContactsForResponsePrefersStoredContacts(t *testing.T) {
	now := time.Date(2026, 3, 26, 10, 30, 0, 0, time.UTC)
	details := Details{
		User: store.User{
			Email:     "fallback@example.com",
			CreatedAt: now.Add(-time.Hour),
		},
		Contacts: []store.ContactMethod{{
			ContactType:  "wechat",
			ContactValue: "wx_123",
			UpdatedAt:    now,
		}},
	}

	contacts := contactsForResponse(details)
	if len(contacts) != 1 {
		t.Fatalf("contacts len = %d, want 1", len(contacts))
	}
	if contacts[0].Type != "wechat" || contacts[0].Value != "wx_123" {
		t.Fatalf("unexpected stored contact: %#v", contacts[0])
	}
	if contacts[0].Readonly {
		t.Fatalf("stored contact should not be readonly: %#v", contacts[0])
	}
}
