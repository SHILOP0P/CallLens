package username

import (
	"crypto/rand"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/mozillazg/go-unidecode"
)

const (
	MinLength = 4
	MaxLength = 24
)

var usernamePattern = regexp.MustCompile(`^@[a-z][a-z0-9_]{3,23}$`)

func Normalize(value string) (string, bool) {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return "", false
	}

	value = strings.TrimPrefix(value, "@")
	value = transliterate(value)
	value = sanitize(value)
	value = strings.Trim(value, "_")
	if len(value) > MaxLength {
		value = strings.TrimRight(value[:MaxLength], "_")
	}

	normalized := "@" + value
	return normalized, usernamePattern.MatchString(normalized)
}

func Generate(parts ...string) (string, error) {
	base := ""
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		base = transliterate(part)
		base = sanitize(base)
		base = strings.Trim(base, "_")
		if len(base) >= MinLength && startsWithLetter(base) {
			break
		}
	}

	if len(base) < MinLength || !startsWithLetter(base) {
		base = "user"
	}
	if len(base) > 16 {
		base = strings.TrimRight(base[:16], "_")
	}

	suffix, err := randomSuffix(6)
	if err != nil {
		return "", err
	}

	return "@" + strings.TrimRight(base, "_") + "_" + suffix, nil
}

func transliterate(value string) string {
	return strings.ToLower(unidecode.Unidecode(value))
}

func sanitize(value string) string {
	var builder strings.Builder
	lastUnderscore := false

	for _, r := range value {
		valid := r <= unicode.MaxASCII && (unicode.IsLetter(r) || unicode.IsDigit(r))
		if valid {
			builder.WriteRune(r)
			lastUnderscore = false
			continue
		}

		if !lastUnderscore {
			builder.WriteByte('_')
			lastUnderscore = true
		}
	}

	return builder.String()
}

func startsWithLetter(value string) bool {
	if value == "" {
		return false
	}
	first := rune(value[0])
	return first >= 'a' && first <= 'z'
}

func randomSuffix(length int) (string, error) {
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("read random username suffix: %w", err)
	}

	for i, b := range bytes {
		bytes[i] = alphabet[int(b)%len(alphabet)]
	}

	return string(bytes), nil
}
