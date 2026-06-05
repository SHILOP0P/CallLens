package middleware

import (
	"calllens/monolit/internal/auth/token"
	"calllens/monolit/internal/logger"
	"calllens/monolit/internal/repository"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

func Auth(secret string, refreshSessionRepository repository.RefreshSessionRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "invalid authorization header", http.StatusUnauthorized)
				return
			}

			const bearerPrefix = "Bearer "
			if !strings.HasPrefix(authHeader, bearerPrefix) {
				http.Error(w, "invalid authorization header", http.StatusUnauthorized)
				return
			}

			rawToken := strings.TrimSpace(strings.TrimPrefix(authHeader, bearerPrefix))
			if rawToken == "" {
				http.Error(w, "empty access token", http.StatusUnauthorized)
				return
			}

			claims, err := token.ParseAccessToken(rawToken, secret)
			if err != nil {
				http.Error(w, "invalid access token", http.StatusUnauthorized)
				return
			}

			if claims.SessionID == uuid.Nil {
				http.Error(w, "invalid access token", http.StatusUnauthorized)
				return
			}

			session, err := refreshSessionRepository.GetRefreshSessionByUUID(r.Context(), claims.SessionID)
			if err != nil {
				http.Error(w, "invalid access token", http.StatusUnauthorized)
				return
			}

			if session.UserID != claims.UserID || session.RevokedAt != nil || !session.ExpiresAt.After(time.Now().UTC()) {
				http.Error(w, "invalid access token", http.StatusUnauthorized)
				return
			}

			ctx := ContextWithUserID(r.Context(), claims.UserID)
			ctx = ContextWithSessionID(ctx, claims.SessionID)
			ctx = ContextWithUserRole(ctx, claims.Role)
			ctx = logger.ContextWithUserID(ctx, claims.UserID.String())

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
