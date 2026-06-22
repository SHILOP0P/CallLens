package password

import (
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashAndCompare(t *testing.T) {
	hash, err := Hash("correct horse battery staple", "pepper")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	if err := Compare("correct horse battery staple", hash, "pepper"); err != nil {
		t.Fatalf("Compare valid password: %v", err)
	}
	if err := Compare("wrong", hash, "pepper"); !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		t.Fatalf("Compare wrong password = %v", err)
	}
}

func TestEmptyPepper(t *testing.T) {
	if _, err := Hash("password", ""); !errors.Is(err, ErrEmptyPepper) {
		t.Fatalf("Hash error = %v", err)
	}
	if err := Compare("password", "hash", ""); !errors.Is(err, ErrEmptyPepper) {
		t.Fatalf("Compare error = %v", err)
	}
}
