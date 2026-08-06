package meetup

import (
	"strings"
	"testing"
	"time"
)

func TestSafetyShareTextContainsOnlyMinimalMeetupFields(t *testing.T) {
	start := time.Date(2026, time.August, 8, 11, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	summary := MeetupSafetySummary{
		MeetupID: 7, StartsAt: start.Format(time.RFC3339),
		EndsAt: start.Add(time.Hour).Format(time.RFC3339), Location: "紫荆操场",
		ParticipantCount: 2, Status: "confirmed",
	}
	text := safetyShareText(summary)
	for _, expected := range []string{"2026-08-08 11:00", "12:00", "紫荆操场", "参与人数：2", "状态：已确认"} {
		if !strings.Contains(text, expected) {
			t.Fatalf("safetyShareText() = %q, missing %q", text, expected)
		}
	}
	for _, forbidden := range []string{"聊天", "标题", "邮箱", "手机号", "参与者姓名"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("safetyShareText() leaked %q: %s", forbidden, text)
		}
	}
}

func TestTrustedContactValidation(t *testing.T) {
	valid := []TrustedContactInput{
		{Label: "室友", ContactType: "email", ContactValue: "roommate@example.com"},
		{Label: "家人", ContactType: "phone", ContactValue: "+86 138 0000 0000"},
		{Label: "朋友", ContactType: "wechat", ContactValue: "trusted_friend"},
	}
	for _, input := range valid {
		if err := validateTrustedContact(input); err != nil {
			t.Fatalf("validateTrustedContact(%#v) error = %v", input, err)
		}
	}
	if err := validateTrustedContact(TrustedContactInput{
		Label: "陌生人", ContactType: "email", ContactValue: "not-an-email",
	}); err == nil {
		t.Fatal("invalid email contact was accepted")
	}
}
