package token

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestGenerateAndParseAccessToken(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	raw, err := GenerateAccessTokenWithSession(userID, sessionID, "user", "secret", time.Minute, 7)
	if err != nil {
		t.Fatalf("GenerateAccessTokenWithSession: %v", err)
	}
	claims, err := ParseAccessToken(raw, "secret")
	if err != nil {
		t.Fatalf("ParseAccessToken: %v", err)
	}
	if claims.UserID != userID || claims.SessionID != sessionID || claims.Role != "user" || claims.AccessVersion != 7 {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestGenerateAccessTokenWithoutSessionIsRejected(t *testing.T) {
	raw, err := GenerateAccessToken(uuid.New(), "user", "secret", time.Minute)
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}
	if _, err := ParseAccessToken(raw, "secret"); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("ParseAccessToken error = %v", err)
	}
}

func TestParseAccessTokenRejectsInvalidTokens(t *testing.T) {
	tests := []struct {
		name string
		raw  func() string
	}{
		{name: "malformed", raw: func() string { return "not-a-token" }},
		{name: "wrong secret", raw: func() string {
			raw, _ := GenerateAccessTokenWithSession(uuid.New(), uuid.New(), "user", "one", time.Minute, 1)
			return raw
		}},
		{name: "expired", raw: func() string {
			raw, _ := GenerateAccessTokenWithSession(uuid.New(), uuid.New(), "user", "secret", -time.Minute, 1)
			return raw
		}},
		{name: "empty user", raw: func() string {
			raw, _ := GenerateAccessTokenWithSession(uuid.Nil, uuid.New(), "user", "secret", time.Minute, 1)
			return raw
		}},
		{name: "empty role", raw: func() string {
			raw, _ := GenerateAccessTokenWithSession(uuid.New(), uuid.New(), "", "secret", time.Minute, 1)
			return raw
		}},
		{name: "wrong algorithm", raw: func() string {
			claims := Claims{
				UserID:    uuid.New(),
				SessionID: uuid.New(),
				Role:      "user",
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
				},
			}
			raw, _ := jwt.NewWithClaims(jwt.SigningMethodNone, claims).SignedString(jwt.UnsafeAllowNoneSignatureType)
			return raw
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ParseAccessToken(tt.raw(), "secret"); !errors.Is(err, ErrInvalidToken) {
				t.Fatalf("ParseAccessToken error = %v", err)
			}
		})
	}
}
