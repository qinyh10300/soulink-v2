package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type Envelope struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// OK 返回默认成功响应。
func OK(w http.ResponseWriter, data any) {
	Respond(w, http.StatusOK, "ok", data)
}

// Respond 写出带自定义 message 的成功响应。
func Respond(w http.ResponseWriter, status int, message string, data any) {
	writeJSON(w, status, Envelope{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	})
}

// Error 写出统一错误响应。
func Error(w http.ResponseWriter, err error) {
	appErr := AsAppError(err)
	if appErr.Code == CodeInternalError && appErr.Err != nil {
		slog.Error("internal api error", "status", appErr.HTTPStatus, "code", appErr.Code, "error", appErr.Err)
	}
	writeJSON(w, appErr.HTTPStatus, Envelope{
		Code:    appErr.Code,
		Message: appErr.Message,
		Data:    nil,
	})
}

// writeJSON 将响应体编码为 JSON 并写回客户端。
func writeJSON(w http.ResponseWriter, status int, payload Envelope) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
