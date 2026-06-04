package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var ErrInvalidToken = errors.New("invalid token")

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(userID uuid.UUID, role string, secret string, ttl time.Duration) (string, error) {
	now := time.Now().UTC()

	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return jwtToken.SignedString([]byte(secret))
}

func ParseAccessToken(rawToken, secret string) (Claims, error) {
	claims := &Claims{}

	parsedToken, err := jwt.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})
	if err != nil {
		return Claims{}, ErrInvalidToken
	}

	if !parsedToken.Valid {
		return Claims{}, ErrInvalidToken
	}

	if claims.UserID == uuid.Nil {
		return Claims{}, ErrInvalidToken
	}

	if claims.Role == "" {
		return Claims{}, ErrInvalidToken
	}

	return *claims, nil
}
