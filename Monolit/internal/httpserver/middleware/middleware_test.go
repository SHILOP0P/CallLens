package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"calllens/monolit/internal/auth/token"
	"calllens/monolit/internal/logger"
	"calllens/monolit/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func TestContextValues(t *testing.T) {
	ctx := context.Background()
	if _, ok := UserIDFromContext(ctx); ok {
		t.Fatal("unexpected user ID")
	}
	if _, ok := SessionIDFromContext(ctx); ok {
		t.Fatal("unexpected session ID")
	}
	if _, ok := UserRoleFromContext(ctx); ok {
		t.Fatal("unexpected user role")
	}

	userID := uuid.New()
	sessionID := uuid.New()
	ctx = ContextWithUserID(ctx, userID)
	ctx = ContextWithSessionID(ctx, sessionID)
	ctx = ContextWithUserRole(ctx, "user")
	if got, ok := UserIDFromContext(ctx); !ok || got != userID {
		t.Fatalf("user ID = %v, %v", got, ok)
	}
	if got, ok := SessionIDFromContext(ctx); !ok || got != sessionID {
		t.Fatalf("session ID = %v, %v", got, ok)
	}
	if got, ok := UserRoleFromContext(ctx); !ok || got != "user" {
		t.Fatalf("role = %q, %v", got, ok)
	}
}

func TestAccessTokenFromRequest(t *testing.T) {
	tests := []struct {
		name   string
		header string
		cookie string
		want   string
		ok     bool
	}{
		{name: "bearer", header: "Bearer token", want: "token", ok: true},
		{name: "empty bearer", header: "Bearer  ", ok: false},
		{name: "invalid scheme", header: "Basic token", ok: false},
		{name: "cookie", cookie: "cookie-token", want: "cookie-token", ok: true},
		{name: "empty cookie", cookie: " ", ok: false},
		{name: "missing", ok: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			if tt.cookie != "" {
				req.AddCookie(&http.Cookie{Name: accessTokenCookieName, Value: tt.cookie})
			}
			got, ok := accessTokenFromRequest(req)
			if got != tt.want || ok != tt.ok {
				t.Fatalf("accessTokenFromRequest = %q, %v; want %q, %v", got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestAuthRejectsInvalidRequests(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	secret := "secret"
	validToken, _ := token.GenerateAccessTokenWithSession(userID, sessionID, "user", secret, time.Minute)

	tests := []struct {
		name  string
		token string
		repo  *authRefreshSessionRepository
	}{
		{name: "missing token", repo: &authRefreshSessionRepository{}},
		{name: "invalid token", token: "invalid", repo: &authRefreshSessionRepository{}},
		{name: "repository error", token: validToken, repo: &authRefreshSessionRepository{err: errors.New("db error")}},
		{name: "different user", token: validToken, repo: &authRefreshSessionRepository{session: models.RefreshSession{ID: sessionID, UserID: uuid.New(), ExpiresAt: time.Now().Add(time.Hour)}}},
		{name: "revoked", token: validToken, repo: &authRefreshSessionRepository{session: models.RefreshSession{ID: sessionID, UserID: userID, RevokedAt: timePtr(time.Now()), ExpiresAt: time.Now().Add(time.Hour)}}},
		{name: "expired", token: validToken, repo: &authRefreshSessionRepository{session: models.RefreshSession{ID: sessionID, UserID: userID, ExpiresAt: time.Now().Add(-time.Hour)}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			handler := Auth(secret, tt.repo)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				called = true
			}))
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if called || rec.Code != http.StatusUnauthorized {
				t.Fatalf("called=%v status=%d", called, rec.Code)
			}
		})
	}
}

func TestRecoverer(t *testing.T) {
	tests := []struct {
		name       string
		handler    http.Handler
		wantStatus int
	}{
		{name: "normal", handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }), wantStatus: http.StatusNoContent},
		{name: "panic before write", handler: http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("boom") }), wantStatus: http.StatusInternalServerError},
		{name: "panic after write", handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusAccepted); panic("boom") }), wantStatus: http.StatusAccepted},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			Recoverer(nil)(tt.handler).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}

func TestRequestLoggerStatusesAndRoute(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusBadRequest, http.StatusInternalServerError} {
		router := chi.NewRouter()
		router.Use(RequestLogger(logger.NewNop()))
		router.Get("/items/{id}", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(status)
			_, _ = w.Write([]byte("body"))
		})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/items/1?q=test", nil))
		if rec.Code != status {
			t.Fatalf("status = %d, want %d", rec.Code, status)
		}
	}
}

func timePtr(value time.Time) *time.Time { return &value }
