package refresh

import (
	"encoding/base64"
	"errors"
	"testing"
)

func TestGenerateAndHash(t *testing.T) {
	token, err := Generate()
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	decoded, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(decoded) != 32 {
		t.Fatalf("generated token is invalid: len=%d err=%v", len(decoded), err)
	}

	first, err := Hash(token, "secret")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	second, _ := Hash(token, "secret")
	if first == "" || first != second {
		t.Fatalf("hashes are not deterministic: %q != %q", first, second)
	}
	if _, err := Hash(token, ""); !errors.Is(err, ErrEmptySecret) {
		t.Fatalf("empty secret error = %v", err)
	}
}
