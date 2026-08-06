package safety

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
	"time"

	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/store"
)

func TestCapabilitiesNormalizesAdminEmail(t *testing.T) {
	service := NewService(nil, []string{"Admin@Mails.Tsinghua.Edu.Cn"})

	if !service.Capabilities(" admin@mails.tsinghua.edu.cn ").IsAdmin {
		t.Fatal("Capabilities().IsAdmin = false, want true")
	}
	if service.Capabilities("student@mails.tsinghua.edu.cn").IsAdmin {
		t.Fatal("Capabilities().IsAdmin = true for a non-admin")
	}
}

func TestUpdateReportValidatesReviewNotesBeforeStoreAccess(t *testing.T) {
	service := NewService(nil, []string{"admin@mails.tsinghua.edu.cn"})

	_, err := service.UpdateReport(t.Context(), 7, "admin@mails.tsinghua.edu.cn", 3, ReportUpdateInput{
		Status:      "reviewing",
		ReviewNotes: strings.Repeat("备", 2001),
		AssignToMe:  true,
	})
	appErr := httpapi.AsAppError(err)
	if appErr.Code != httpapi.CodeInvalidParams {
		t.Fatalf("UpdateReport() code = %d, want %d", appErr.Code, httpapi.CodeInvalidParams)
	}
}

func TestSubmitReportValidatesInputBeforeStoreAccess(t *testing.T) {
	service := NewService(nil, nil)

	_, err := service.SubmitReport(t.Context(), 1, ReportInput{
		ContextType: ContextIntent,
		ContextID:   2,
		Category:    "unknown",
		Details:     "这是一段足够长的举报说明",
	})
	if err == nil {
		t.Fatal("SubmitReport() error = nil, want invalid category error")
	}
}

func TestPublicTrustProfileMasksPrivateReviewState(t *testing.T) {
	requestID := int64(18)
	profile := store.UserTrustProfile{
		EmailVerified: true, ProfileComplete: true, PhotoStatus: "rejected",
		PhotoRequestID:   sql.NullInt64{Int64: requestID, Valid: true},
		PhotoReviewNotes: "包含私密审核说明", PhotoAppealReason: "包含申诉理由",
	}

	public := trustProfileView(profile, false)
	if public.PhotoStatus != "not_verified" || public.PhotoRequestID != nil ||
		public.PhotoReviewNotes != "" || public.PhotoAppealReason != "" {
		t.Fatalf("public trust profile leaked private review state: %#v", public)
	}
	private := trustProfileView(profile, true)
	if private.PhotoRequestID == nil || *private.PhotoRequestID != requestID ||
		private.PhotoReviewNotes == "" || private.PhotoAppealReason == "" {
		t.Fatalf("private trust profile omitted owner state: %#v", private)
	}
}

func TestNormalizePhotoEvidenceDecodesAndReencodes(t *testing.T) {
	inputImage := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	inputImage.Set(0, 0, color.NRGBA{R: 55, G: 109, B: 88, A: 255})
	var raw bytes.Buffer
	if err := png.Encode(&raw, inputImage); err != nil {
		t.Fatalf("encode fixture: %v", err)
	}

	mimeType, normalized, err := normalizePhotoEvidence(raw.Bytes())
	if err != nil {
		t.Fatalf("normalizePhotoEvidence() error = %v", err)
	}
	if mimeType != "image/png" || len(normalized) == 0 {
		t.Fatalf("normalized evidence = %q, %d bytes", mimeType, len(normalized))
	}
	if _, _, err := image.Decode(bytes.NewReader(normalized)); err != nil {
		t.Fatalf("normalized evidence is not decodable: %v", err)
	}
	if _, _, err := normalizePhotoEvidence([]byte("not an image")); err == nil {
		t.Fatal("invalid evidence was accepted")
	}
}

func TestReportRiskSignalsAndSnapshotRemainStructured(t *testing.T) {
	signals := deriveReportRiskSignals("加我微信做推广", "不来就曝光你的宿舍号")
	for _, expected := range []string{"spam", "contact_inducement", "privacy", "coercion"} {
		found := false
		for _, signal := range signals {
			if signal == expected {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("signals = %#v, missing %q", signals, expected)
		}
	}
	coercionSignals := deriveReportRiskSignals("你必须单独见我，而且不许告诉别人")
	if len(coercionSignals) != 1 || coercionSignals[0] != "coercion" {
		t.Fatalf("coercion synonyms were not detected: %#v", coercionSignals)
	}

	record := store.SafetyReport{
		ID: 8, ConversationID: sql.NullInt64{Int64: 12, Valid: true},
		ContextSnapshot: json.RawMessage(`{"version":1,"messages":[{"senderRole":"reported_user","content":"固定证据"}]}`),
		RiskSignals:     json.RawMessage(`["privacy"]`), CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	view := toReportView(record)
	if view.ContextType != ContextDirect || view.ContextID != 12 || len(view.RiskSignals) != 1 {
		t.Fatalf("report view did not retain structured context: %#v", view)
	}
}

func TestSafetyCaseViewOmitsReporterDetailsAndSnapshot(t *testing.T) {
	encoded, err := json.Marshal(SafetyCaseView{
		ID: 1, TicketID: "SR-000001", Category: "other", ContextType: ContextDirect,
		ContextTitle: "一对一会话", Status: "resolved", CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("marshal safety case: %v", err)
	}
	for _, forbiddenField := range []string{"reporterId", "reporterEmail", "details", "contextSnapshot", "riskSignals"} {
		if bytes.Contains(encoded, []byte(forbiddenField)) {
			t.Fatalf("safety case leaked %q: %s", forbiddenField, encoded)
		}
	}
}

func TestReviewAppealChecksAdminBeforeStoreAccess(t *testing.T) {
	service := NewService(nil, []string{"admin@mails.tsinghua.edu.cn"})
	_, err := service.ReviewReportAppeal(t.Context(), 4, "student@mails.tsinghua.edu.cn", 9, ReportAppealReviewInput{
		Status: "accepted",
	})
	if appErr := httpapi.AsAppError(err); appErr.Code != httpapi.CodeForbidden {
		t.Fatalf("ReviewReportAppeal() code = %d, want %d", appErr.Code, httpapi.CodeForbidden)
	}
}
