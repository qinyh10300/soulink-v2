package safety

import (
	"strings"
	"testing"

	"soulink-backend/internal/httpapi"
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
