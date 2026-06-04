package password

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var ErrEmptyPepper = errors.New("password pepper is empty")

func Hash(rawPassword, pepper string) (string, error) {
	if pepper == "" {
		return "", ErrEmptyPepper
	}

	pepperedPassword := pepperPassword(rawPassword, pepper)

	hash, err := bcrypt.GenerateFromPassword([]byte(pepperedPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func Compare(rawPassword string, passwordHash string, pepper string) error {
	if pepper == "" {
		return ErrEmptyPepper
	}

	pepperedPassword := pepperPassword(rawPassword, pepper)

	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(pepperedPassword))
}

func pepperPassword(rawPassword string, pepper string) string {
	mac := hmac.New(sha256.New, []byte(pepper))
	mac.Write([]byte(rawPassword))

	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
