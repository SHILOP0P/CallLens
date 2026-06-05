package refresh

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
)

var ErrEmptySecret = errors.New("refresh token secret is empty")

func Generate() (string, error) {
	const tokenSize = 32

	buffer := make([]byte, tokenSize)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func Hash(rawToken, secret string) (string, error) {
	if secret == "" {
		return "", ErrEmptySecret
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(rawToken))

	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}
