package operations

import (
	"net/http"
	"strings"
	"time"

	"soulink-backend/internal/auth"
	"soulink-backend/internal/httpapi"
)

type Handler struct {
	collector   *Collector
	adminEmails map[string]struct{}
	nowFunc     func() time.Time
}

func NewHandler(collector *Collector, adminEmails []string) *Handler {
	admins := make(map[string]struct{}, len(adminEmails))
	for _, email := range adminEmails {
		if normalized := normalizeEmail(email); normalized != "" {
			admins[normalized] = struct{}{}
		}
	}
	return &Handler{collector: collector, adminEmails: admins, nowFunc: time.Now}
}

func (h *Handler) Metrics(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}
	if _, ok := h.adminEmails[normalizeEmail(user.Email)]; !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusForbidden, httpapi.CodeForbidden, "仅管理员可以查看运行指标"))
		return
	}
	if h.collector == nil {
		httpapi.Error(w, httpapi.FeatureNotReady("operational metrics"))
		return
	}
	httpapi.OK(w, h.collector.Snapshot(h.nowFunc()))
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
