package middleware

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

const (
	userIDContextKey    contextKey = "user_id"
	sessionIDContextKey contextKey = "session_id"
	userRoleContextKey  contextKey = "user_role"
)

func ContextWithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDContextKey, userID)
}

func ContextWithUserRole(ctx context.Context, userRole string) context.Context {
	return context.WithValue(ctx, userRoleContextKey, userRole)
}

func ContextWithSessionID(ctx context.Context, sessionID uuid.UUID) context.Context {
	return context.WithValue(ctx, sessionIDContextKey, sessionID)
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDContextKey).(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return uuid.Nil, false
	}

	return userID, true
}

func SessionIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	sessionID, ok := ctx.Value(sessionIDContextKey).(uuid.UUID)
	if !ok || sessionID == uuid.Nil {
		return uuid.Nil, false
	}

	return sessionID, true
}

func UserRoleFromContext(ctx context.Context) (string, bool) {
	userRole, ok := ctx.Value(userRoleContextKey).(string)
	if !ok || userRole == "" {
		return "", false
	}

	return userRole, true
}
