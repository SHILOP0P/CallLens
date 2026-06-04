package middleware

import (
	"calllens/monolit/internal/auth/token"
	"net/http"
	"strings"
)

func Auth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "invalid authorization header", http.StatusUnauthorized)
				return
			}

			const bearerPrefix = "Bearer "
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

			ctx := ContextWithUserID(r.Context(), claims.UserID)
			ctx = ContextWithUserRole(ctx, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
