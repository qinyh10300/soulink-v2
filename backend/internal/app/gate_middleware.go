package app

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"soulink-backend/internal/gate"
	"soulink-backend/internal/httpapi"
)

type GateChecker interface {
	Check(ctx context.Context, gateKey string) (gate.Status, error)
}

// MatchWriteGuard 在匹配计算时间窗内封闭问卷、资料和联系方式写接口。
func MatchWriteGuard(logger *slog.Logger, checker GateChecker) func(http.Handler) http.Handler {
	return GateWindowMiddleware(logger, checker, gate.GateKeyMatchingWrite, "当前正在进行匹配计算，相关写接口暂时关闭")
}

// GateWindowMiddleware 根据运行时 gate 状态决定是否放行请求。
func GateWindowMiddleware(logger *slog.Logger, checker GateChecker, gateKey string, fallbackMessage string) func(http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	if checker == nil || strings.TrimSpace(gateKey) == "" {
		return passthroughMiddleware
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			status, err := checker.Check(r.Context(), gateKey)
			if err != nil {
				logger.Error("check interface gate", "gate_key", gateKey, "path", r.URL.Path, "error", err)
				httpapi.Error(w, httpapi.InternalError(err))
				return
			}
			if status.Active {
				message := strings.TrimSpace(status.Reason)
				if message == "" {
					message = strings.TrimSpace(fallbackMessage)
				}
				if message == "" {
					message = "当前接口暂时关闭"
				}
				logger.Warn("request blocked by interface gate", "gate_key", gateKey, "path", r.URL.Path, "starts_at", status.StartsAt, "ends_at", status.EndsAt)
				httpapi.Error(w, httpapi.NewError(http.StatusForbidden, httpapi.CodeForbidden, message))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func passthroughMiddleware(next http.Handler) http.Handler {
	return next
}
