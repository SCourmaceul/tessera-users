package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestValidatePseudo(t *testing.T) {
	valid := []string{"ada", "  Ada_Lovelace ", "zoé.b-42", strings.Repeat("a", 30)}
	for _, p := range valid {
		if _, err := ValidatePseudo(p); err != nil {
			t.Errorf("ValidatePseudo(%q) = %v", p, err)
		}
	}
	invalid := []string{"", "ab", strings.Repeat("a", 31), "ada lovelace", "ada@home", "<script>"}
	for _, p := range invalid {
		if _, err := ValidatePseudo(p); !errors.Is(err, ErrInvalidPseudo) {
			t.Errorf("ValidatePseudo(%q) = %v, want ErrInvalidPseudo", p, err)
		}
	}
	if got, _ := ValidatePseudo("  ada  "); got != "ada" {
		t.Errorf("pseudo non nettoyé : %q", got)
	}
}

func TestValidateAvatarURL(t *testing.T) {
	for _, u := range []string{"", "https://cdn.example.com/a.png"} {
		if _, err := ValidateAvatarURL(u); err != nil {
			t.Errorf("ValidateAvatarURL(%q) = %v", u, err)
		}
	}
	for _, u := range []string{"http://cdn.example.com/a.png", "javascript:alert(1)", "https://", "a.png", "https://x.io/" + strings.Repeat("a", 2048)} {
		if _, err := ValidateAvatarURL(u); !errors.Is(err, ErrInvalidAvatarURL) {
			t.Errorf("ValidateAvatarURL(%q) = %v, want ErrInvalidAvatarURL", u, err)
		}
	}
}

func TestDefaultPseudo(t *testing.T) {
	tests := []struct{ preferred, email, want string }{
		{"Ada", "ada@tetra.local", "Ada"},
		{"", "grace.hopper@tetra.local", "grace.hopper"},
		{"x", "j+o@tetra.local", "user-0a1b2c3d"},
		{"", "", "user-0a1b2c3d"},
	}
	for _, tt := range tests {
		if got := DefaultPseudo("0a1b2c3d-4e5f", tt.preferred, tt.email); got != tt.want {
			t.Errorf("DefaultPseudo(%q, %q) = %q, want %q", tt.preferred, tt.email, got, tt.want)
		}
	}
	// Le pseudo proposé est toujours valide
	long := strings.Repeat("é", 40) + "@tetra.local"
	if _, err := ValidatePseudo(DefaultPseudo("id", "", long)); err != nil {
		t.Errorf("pseudo par défaut invalide : %v", err)
	}
}
