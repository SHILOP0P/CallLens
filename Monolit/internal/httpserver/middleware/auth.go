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

const accessTokenCookieName = "access_token"

func Auth(secret string, refreshSessionRepository repository.RefreshSessionRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawToken, ok := accessTokenFromRequest(r)
			if !ok {
				response.WriteError(w, http.StatusUnauthorized, response.CodeInvalidAuthorizationHeader, "invalid authorization header")
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

func accessTokenFromRequest(r *http.Request) (string, bool) {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			return "", false
		}

		rawToken := strings.TrimSpace(strings.TrimPrefix(authHeader, bearerPrefix))
		return rawToken, rawToken != ""
	}

	if cookie, err := r.Cookie(accessTokenCookieName); err == nil {
		rawToken := strings.TrimSpace(cookie.Value)
		return rawToken, rawToken != ""
	}

	return "", false
}
