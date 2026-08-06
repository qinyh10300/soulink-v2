package connection

import (
	"testing"

	"soulink-backend/internal/store"
)

func TestGroupMatchScoringRewardsRoleComplement(t *testing.T) {
	source := store.GroupMatchIntent{
		Category: CategoryProject, RoleOffer: "产品",
		RolesNeeded: []string{"React 前端", "后端"},
		Constraints: []string{"每周同步", "四周交付"},
		TimeWindow:  "周二、周四晚间", Location: "FIT 楼",
		MinMembers: 3, MaxMembers: 5,
	}
	profile := store.MatchingProfile{Skills: []string{"产品", "用户研究"}}
	complement := store.GroupMatchCandidate{
		Category: CategoryProject, RoleOffer: "前端工程",
		Skills:      []string{"React 前端", "TypeScript"},
		Constraints: []string{"每周同步", "四周交付"},
		TimeWindow:  "周四晚间", Location: "FIT 楼",
		MinMembers: 3, MaxMembers: 6,
	}
	generic := store.GroupMatchCandidate{
		Category: CategoryProject, RoleOffer: "产品",
		Skills: []string{"用户研究"}, Constraints: []string{"异步沟通"},
		TimeWindow: "周末上午", Location: "美院",
		MinMembers: 8, MaxMembers: 10,
	}

	complementScore, _, _, _, gaps, role := scoreGroupMatchCandidate(
		source, profile, complement, true,
	)
	genericScore, _, _, _, _, _ := scoreGroupMatchCandidate(
		source, profile, generic, false,
	)
	if complementScore <= genericScore {
		t.Fatalf("expected complement score %d to exceed generic score %d", complementScore, genericScore)
	}
	if role != "React 前端" {
		t.Fatalf("expected matched role, got %q", role)
	}
	if len(gaps) != 1 || gaps[0] != "后端" {
		t.Fatalf("expected only backend gap, got %#v", gaps)
	}
}

func TestGroupMatchCandidateIdentityRequiresConsent(t *testing.T) {
	invitation := store.GroupMatchInvitation{
		CandidateUserID: 22, CandidateNickname: "真实昵称",
		CandidateDepartment: "计算机系", CandidateGrade: "研一",
		GroupOrganizerID: 11, Status: "invited",
	}
	masked := groupMatchCandidateIdentity(invitation, 11, 0, false)
	if masked.Revealed || masked.UserID != 0 || masked.Nickname != "候选同学 A" {
		t.Fatalf("expected masked candidate before consent, got %#v", masked)
	}

	invitation.Status = "accepted"
	revealed := groupMatchCandidateIdentity(invitation, 11, 0, false)
	if !revealed.Revealed || revealed.UserID != 22 || revealed.Nickname != "真实昵称" {
		t.Fatalf("expected accepted candidate identity, got %#v", revealed)
	}

	otherMember := groupMatchCandidateIdentity(invitation, 33, 0, false)
	if otherMember.Revealed {
		t.Fatalf("unformed members must not see one another, got %#v", otherMember)
	}
	formedMember := groupMatchCandidateIdentity(invitation, 33, 0, true)
	if !formedMember.Revealed {
		t.Fatalf("formed accepted members should be revealed, got %#v", formedMember)
	}
}

func TestGroupMatchResponseAutoFormsOnlyAfterConsentBoundary(t *testing.T) {
	first := decideGroupMatchResponse("accepted", 3, 4, 0, 2)
	if first.InvitationStatus != "accepted" || first.ShouldForm {
		t.Fatalf("first acceptance must not form group, got %#v", first)
	}
	last := decideGroupMatchResponse("accepted", 3, 4, 1, 1)
	if last.InvitationStatus != "accepted" || !last.ShouldForm {
		t.Fatalf("last required consent should auto form, got %#v", last)
	}
}

func TestGroupMatchResponseUsesWaitlistAtCapacity(t *testing.T) {
	decision := decideGroupMatchResponse("accepted", 3, 3, 2, 1)
	if decision.InvitationStatus != "waitlisted" || decision.ShouldForm {
		t.Fatalf("capacity race must preserve consent as waitlist, got %#v", decision)
	}
}

func TestGroupMatchDeclineCanReleaseReadyGroup(t *testing.T) {
	decision := decideGroupMatchResponse("declined", 3, 5, 2, 1)
	if decision.InvitationStatus != "declined" || !decision.ShouldForm {
		t.Fatalf("final response should release an already viable group, got %#v", decision)
	}
}
