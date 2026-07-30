package chat

import (
	"log/slog"
	"net/http"

	"nhooyr.io/websocket"

	"soulink-backend/internal/auth"
	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/store"
)

// ServeWS upgrades the HTTP connection to WebSocket.
// Auth is done via ?token=xxx query param since browsers cannot set headers on WebSocket.
func ServeWS(hub *Hub, service *Service, dataStore *store.Store, tokenManager *auth.TokenManager, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rawToken := r.URL.Query().Get("token")
		if rawToken == "" {
			httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
			return
		}

		claims, err := tokenManager.Parse(rawToken)
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
		_ = session

		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true,
		})
		if err != nil {
			logger.Error("ws accept failed", "error", err)
			return
		}

		client := NewClient(user.ID, conn, hub, service, logger)
		client.Run(r.Context())
	}
}
