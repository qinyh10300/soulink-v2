package playground

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/store"
)

func TestPlazaGroupIdentityRespectsRevealBoundary(t *testing.T) {
	invite := store.PlazaGroupInvite{
		OwnerID: 11, Status: "open", RevealStrategy: "on_form",
	}
	application := store.PlazaGroupApplication{
		ID: 7, UserID: 22, Status: "accepted", AnonymousAlias: "同学 A",
		ApplicantNickname: "真实昵称", ApplicantDepartment: "计算机系", ApplicantGrade: "研一",
		CreatedAt: time.Now(),
	}

	masked := plazaGroupApplicationView(invite, application, invite.OwnerID, true)
	if masked.Identity.Revealed || masked.Identity.UserID != 0 || masked.Identity.Alias != "同学 A" {
		t.Fatalf("expected applicant to stay anonymous before form, got %#v", masked.Identity)
	}

	invite.RevealStrategy = "on_accept"
	revealedToOwner := plazaGroupApplicationView(invite, application, invite.OwnerID, true)
	if !revealedToOwner.Identity.Revealed || revealedToOwner.Identity.UserID != application.UserID {
		t.Fatalf("expected on-accept identity for owner, got %#v", revealedToOwner.Identity)
	}

	invite.Status = "formed"
	invite.RevealStrategy = "on_form"
	revealedAfterForm := plazaGroupApplicationView(invite, application, 33, true)
	if !revealedAfterForm.Identity.Revealed || revealedAfterForm.Identity.Nickname != "真实昵称" {
		t.Fatalf("expected identity after form for accepted member, got %#v", revealedAfterForm.Identity)
	}
}

func TestPlazaGroupApplicationEligibilityRejectsUnauthorizedAndExpiredUsers(t *testing.T) {
	now := time.Date(2026, time.August, 7, 12, 0, 0, 0, time.UTC)
	base := store.PlazaGroupInvite{OwnerID: 11, Status: "open", Deadline: now.Add(time.Hour)}
	tests := []struct {
		name       string
		invite     store.PlazaGroupInvite
		userID     int64
		hasComment bool
		status     int
	}{
		{name: "owner", invite: base, userID: 11, hasComment: true, status: http.StatusForbidden},
		{name: "never replied", invite: base, userID: 22, hasComment: false, status: http.StatusForbidden},
		{name: "deadline passed", invite: store.PlazaGroupInvite{OwnerID: 11, Status: "open", Deadline: now}, userID: 22, hasComment: true, status: http.StatusConflict},
		{name: "already formed", invite: store.PlazaGroupInvite{OwnerID: 11, Status: "formed", Deadline: now.Add(time.Hour)}, userID: 22, hasComment: true, status: http.StatusConflict},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validatePlazaGroupApplicationEligibility(test.invite, test.userID, test.hasComment, now)
			var appErr *httpapi.AppError
			if !errors.As(err, &appErr) || appErr.HTTPStatus != test.status {
				t.Fatalf("expected HTTP %d, got %#v", test.status, err)
			}
		})
	}
	if err := validatePlazaGroupApplicationEligibility(base, 22, true, now); err != nil {
		t.Fatalf("eligible replier was rejected: %v", err)
	}
}

func TestPlazaContentSafetyRejectsContactAndDangerousContent(t *testing.T) {
	unsafe := []string{
		"加我微信 soulink_2026",
		"手机号 13800138000",
		"发邮件到 demo@example.com",
		"我要开盒曝光隐私",
	}
	for _, value := range unsafe {
		if err := validatePlazaContent(value); err == nil {
			t.Fatalf("expected unsafe content to be rejected: %q", value)
		}
	}
	if err := validatePlazaContent("明天晚上在紫操一起跑步，先在群里确认节奏"); err != nil {
		t.Fatalf("expected normal activity content to pass: %v", err)
	}
}

func TestPlazaGroupDraftExtractsTimeAndLocation(t *testing.T) {
	post := store.PlaygroundPost{
		ID: 42, Content: "明天晚上想在紫操约一场轻松跑",
	}
	draft := derivePlazaGroupDraft(post, []store.PlaygroundComment{{Content: "我可以参加"}}, "event")
	if draft.ExtractedTime == "" || draft.ExtractedLocation != "紫操" {
		t.Fatalf("expected time and location extraction, got %#v", draft)
	}
	if draft.Confidence <= 0.58 || draft.Title == "" || draft.Summary == "" {
		t.Fatalf("expected attributable fallback draft, got %#v", draft)
	}
}

func TestPlazaGroupAutoFormsAtConsentThreshold(t *testing.T) {
	invite := store.PlazaGroupInvite{MinMembers: 3, MaxMembers: 5}
	if shouldFormPlazaGroup(invite, store.PlazaGroupApplicationCounts{Accepted: 2, Applied: 1}) {
		t.Fatal("pending application must preserve the owner's decision window")
	}
	if !shouldFormPlazaGroup(invite, store.PlazaGroupApplicationCounts{Accepted: 2}) {
		t.Fatal("group should form after minimum consent and all applications are handled")
	}
	if !shouldFormPlazaGroup(invite, store.PlazaGroupApplicationCounts{Accepted: 4, Applied: 2}) {
		t.Fatal("group should form immediately at capacity")
	}
}
