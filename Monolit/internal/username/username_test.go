package username

import (
	"regexp"
	"strings"
	"testing"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
		ok    bool
	}{
		{name: "empty", input: "  ", want: "", ok: false},
		{name: "handle", input: " @John.Doe ", want: "@john_doe", ok: true},
		{name: "cyrillic", input: "Дмитрий Мухачев", want: "@dmitrii_mukhachev", ok: true},
		{name: "starts with digit", input: "12345", want: "@12345", ok: false},
		{name: "too short", input: "abc", want: "@abc", ok: false},
		{name: "trim underscores", input: "__valid_name__", want: "@valid_name", ok: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := Normalize(tt.input)
			if got != tt.want || ok != tt.ok {
				t.Fatalf("Normalize(%q) = %q, %v; want %q, %v", tt.input, got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestNormalizeTruncatesToMaxLength(t *testing.T) {
	got, ok := Normalize("a" + strings.Repeat("b", MaxLength+10))
	if !ok || len(strings.TrimPrefix(got, "@")) != MaxLength {
		t.Fatalf("Normalize long value = %q, %v", got, ok)
	}
}

func TestGenerate(t *testing.T) {
	pattern := regexp.MustCompile(`^@[a-z][a-z0-9_]{3,16}_[a-z0-9]{6}$`)

	for _, parts := range [][]string{{"Дмитрий"}, {"12", "Valid Name"}, {"", "x"}} {
		got, err := Generate(parts...)
		if err != nil {
			t.Fatalf("Generate(%q): %v", parts, err)
		}
		if !pattern.MatchString(got) {
			t.Fatalf("generated username %q does not match expected format", got)
		}
	}
}
