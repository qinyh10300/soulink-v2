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

func TestPublicProfileResponseIncludesAuthorizedPreferences(t *testing.T) {
	now := time.Date(2026, 8, 7, 9, 0, 0, 0, time.UTC)
	response := toPublicProfileResponse(PublicDetails{
		User: store.User{
			ID: 4, Email: "student@mails.tsinghua.edu.cn", Nickname: "清华同学",
			Department: "计算机系", Grade: "本科三年级", CreatedAt: now,
		},
		Settings: PrivacySettings{
			ProfilePublic: true, SchedulePublic: true, PreferencesPublic: true,
			AvatarURL: "/media/asset-key",
		},
		Preferences: &PublicPreferences{
			Bio: "想找长期项目伙伴", Skills: []string{"产品"}, Interests: []string{"AI"},
		},
	})

	if response.AvatarURL != "/media/asset-key" || !response.SchedulePublic || !response.PreferencesPublic {
		t.Fatalf("unexpected public profile settings: %#v", response)
	}
	if response.Preferences == nil || response.Preferences.Bio == "" || len(response.Preferences.Skills) != 1 {
		t.Fatalf("authorized preferences missing: %#v", response.Preferences)
	}
}
