package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"calllens/monolit/internal/auth/token"
	"calllens/monolit/internal/models"

	"github.com/google/uuid"
)

func TestAuthAcceptsAccessTokenCookie(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	secret := "test-secret"

	rawToken, err := token.GenerateAccessTokenWithSession(userID, sessionID, string(models.UserRoleUser), secret, time.Minute)
	if err != nil {
		t.Fatalf("generate access token: %v", err)
	}

	repo := &authRefreshSessionRepository{
		session: models.RefreshSession{
			ID:        sessionID,
			UserID:    userID,
			ExpiresAt: time.Now().UTC().Add(time.Hour),
		},
	}

	handler := Auth(secret, repo)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID, ok := UserIDFromContext(r.Context())
		if !ok || gotUserID != userID {
			t.Fatalf("user id in context = %v, %v", gotUserID, ok)
		}

		gotSessionID, ok := SessionIDFromContext(r.Context())
		if !ok || gotSessionID != sessionID {
			t.Fatalf("session id in context = %v, %v", gotSessionID, ok)
		}

		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.AddCookie(&http.Cookie{Name: accessTokenCookieName, Value: rawToken, Path: "/"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

type authRefreshSessionRepository struct {
	session models.RefreshSession
	err     error
}

func (r *authRefreshSessionRepository) CreateRefreshSession(ctx context.Context, session models.RefreshSession) (models.RefreshSession, error) {
	return models.RefreshSession{}, nil
}

func (r *authRefreshSessionRepository) GetRefreshSessionByHash(ctx context.Context, refreshTokenHash string) (models.RefreshSession, error) {
	return models.RefreshSession{}, nil
}

func (r *authRefreshSessionRepository) GetRefreshSessionByUUID(ctx context.Context, sessionID uuid.UUID) (models.RefreshSession, error) {
	if r.err != nil {
		return models.RefreshSession{}, r.err
	}
	if r.session.ID != sessionID {
		return models.RefreshSession{}, models.ErrRefreshSessionNotFound
	}

	return r.session, nil
}

func (r *authRefreshSessionRepository) RotateRefreshSession(ctx context.Context, oldRefreshTokenHash string, newRefreshTokenHash string, expiresAt time.Time) (models.RefreshSession, error) {
	return models.RefreshSession{}, nil
}

func (r *authRefreshSessionRepository) RevokeRefreshSession(ctx context.Context, sessionID uuid.UUID, reason string) error {
	return nil
}

func (r *authRefreshSessionRepository) RevokeAllUserRefreshSessions(ctx context.Context, userID uuid.UUID, reason string) error {
	return nil
}
