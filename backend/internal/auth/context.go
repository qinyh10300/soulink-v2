package auth

import (
	"context"
	"net/http"

	"soulink-backend/internal/store"
)

type contextKey string

const authContextKey contextKey = "auth_context"

type Context struct {
	User    store.User
	Session store.Session
}

// WithContext 将当前登录用户和会话写入请求上下文。
func WithContext(ctx context.Context, value Context) context.Context {
	return context.WithValue(ctx, authContextKey, value)
}

// CurrentUser 从请求上下文中读取当前登录用户。
func CurrentUser(r *http.Request) (store.User, bool) {
	value, ok := r.Context().Value(authContextKey).(Context)
	if !ok {
		return store.User{}, false
	}
	return value.User, true
}

// CurrentSession 从请求上下文中读取当前登录会话。
func CurrentSession(r *http.Request) (store.Session, bool) {
	value, ok := r.Context().Value(authContextKey).(Context)
	if !ok {
		return store.Session{}, false
	}
	return value.Session, true
}
