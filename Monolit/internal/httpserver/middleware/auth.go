package middleware

import (
	"calllens/monolit/internal/API/response"
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
				response.WriteError(w, http.StatusUnauthorized, response.CodeInvalidAuthorizationHeader, "invalid authorization header")
				return
			}

			const bearerPrefix = "Bearer "
			if !strings.HasPrefix(authHeader, bearerPrefix) {
				response.WriteError(w, http.StatusUnauthorized, response.CodeInvalidAuthorizationHeader, "invalid authorization header")
				return
			}

			rawToken := strings.TrimSpace(strings.TrimPrefix(authHeader, bearerPrefix))
			if rawToken == "" {
				response.WriteError(w, http.StatusUnauthorized, response.CodeEmptyAccessToken, "empty access token")
				return
			}

			claims, err := token.ParseAccessToken(rawToken, secret)
			if err != nil {
				response.WriteError(w, http.StatusUnauthorized, response.CodeInvalidAccessToken, "invalid access token")
				return
			}

			if claims.SessionID == uuid.Nil {
				response.WriteError(w, http.StatusUnauthorized, response.CodeInvalidAccessToken, "invalid access token")
				return
			}

			session, err := refreshSessionRepository.GetRefreshSessionByUUID(r.Context(), claims.SessionID)
			if err != nil {
				response.WriteError(w, http.StatusUnauthorized, response.CodeInvalidAccessToken, "invalid access token")
				return
			}

			if session.UserID != claims.UserID || session.RevokedAt != nil || !session.ExpiresAt.After(time.Now().UTC()) {
				response.WriteError(w, http.StatusUnauthorized, response.CodeInvalidAccessToken, "invalid access token")
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
