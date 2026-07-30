package auth

import (
	"net/http"
	"strconv"
	"strings"

	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/store"
)

// Middleware 校验 Bearer Token，并将用户信息注入上下文。
func Middleware(dataStore *store.Store, tokens *TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawToken, ok := extractBearerToken(r.Header.Get("Authorization"))
			if !ok {
				httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
				return
			}

			claims, err := tokens.Parse(rawToken)
			if err != nil {
				httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeInvalidToken, "Token 无效或已过期"))
				return
			}

			session, user, err := dataStore.GetActiveSession(r.Context(), claims.ID)
			if err != nil {
				if err == store.ErrNotFound {
					httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeInvalidToken, "Token 无效或已过期"))
					return
				}
				httpapi.Error(w, httpapi.InternalError(err))
				return
			}

			if claims.Subject != strconv.FormatInt(user.ID, 10) {
				httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeInvalidToken, "Token 无效或已过期"))
				return
			}

			ctx := WithContext(r.Context(), Context{User: user, Session: session})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractBearerToken 从 Authorization 头中提取 Bearer Token。
func extractBearerToken(header string) (string, bool) {
	parts := strings.Fields(strings.TrimSpace(header))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}
	return parts[1], true
}
