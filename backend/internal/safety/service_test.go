package safety

import (
	"testing"
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
