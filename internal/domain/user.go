// Package domain contient les entités du service users et leurs règles : le profil public
// (commun à tous les espaces) et l'appartenance à un espace (rôle, réputation). Il n'importe
// que le cœur de tetra-kit.
package domain

import (
	"net/url"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/SCourmaceul/tetra-kit/errs"
)

// Rôle attribué dans un espace à son arrivée.
const RoleMember = "member"

const (
	pseudoMinLen    = 3
	pseudoMaxLen    = 30
	avatarURLMaxLen = 2048
)

var (
	ErrUserNotFound     = errs.NotFound("USER_NOT_FOUND", "Utilisateur introuvable")
	ErrAlreadyMember    = errs.Conflict("ALREADY_MEMBER", "Vous avez déjà rejoint cet espace")
	ErrInvalidSpace     = errs.Invalid("INVALID_SPACE", "Espace inconnu")
	ErrInvalidPseudo    = errs.Invalid("INVALID_PSEUDO", "Le pseudo doit contenir de 3 à 30 caractères : lettres, chiffres, « _ », « - » ou « . »")
	ErrInvalidAvatarURL = errs.Invalid("INVALID_AVATAR_URL", "L'avatar doit être une URL https valide")
)

// Profile est le profil public d'un utilisateur, partagé par tous les espaces. Son ID est le
// sub Cognito.
type Profile struct {
	ID        string
	Pseudo    string
	AvatarURL string
	// Spaces liste les espaces rejoints, dans l'ordre d'arrivée.
	Spaces    []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// HasJoined indique si l'utilisateur a rejoint l'espace.
func (p Profile) HasJoined(space string) bool {
	for _, s := range p.Spaces {
		if s == space {
			return true
		}
	}
	return false
}

// Member est un utilisateur vu depuis un espace : son appartenance (rôle, réputation) et une
// copie de son profil public, tenue à jour par la modification du profil.
type Member struct {
	UserID     string
	Space      string
	Pseudo     string
	AvatarURL  string
	Role       string
	Reputation int
	JoinedAt   time.Time
}

// NewMember construit l'appartenance d'un utilisateur qui rejoint un espace.
func NewMember(p Profile, space string, now time.Time) Member {
	return Member{
		UserID:    p.ID,
		Space:     space,
		Pseudo:    p.Pseudo,
		AvatarURL: p.AvatarURL,
		Role:      RoleMember,
		JoinedAt:  now,
	}
}

// ValidatePseudo vérifie le pseudo, débarrassé des espaces autour, et le renvoie.
func ValidatePseudo(pseudo string) (string, error) {
	pseudo = strings.TrimSpace(pseudo)
	if n := utf8.RuneCountInString(pseudo); n < pseudoMinLen || n > pseudoMaxLen {
		return "", ErrInvalidPseudo
	}
	for _, r := range pseudo {
		if !isPseudoRune(r) {
			return "", ErrInvalidPseudo
		}
	}
	return pseudo, nil
}

// ValidateAvatarURL vérifie l'URL de l'avatar et la renvoie. Une chaîne vide retire l'avatar.
func ValidateAvatarURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	if len(raw) > avatarURLMaxLen {
		return "", ErrInvalidAvatarURL
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return "", ErrInvalidAvatarURL
	}
	return raw, nil
}

// DefaultPseudo propose un pseudo valide à l'inscription : le pseudo choisi s'il est valide,
// sinon la partie locale de l'e-mail nettoyée, sinon un pseudo dérivé de l'identifiant.
func DefaultPseudo(userID, preferred, email string) string {
	if p, err := ValidatePseudo(preferred); err == nil {
		return p
	}
	local, _, _ := strings.Cut(email, "@")
	var b strings.Builder
	for _, r := range local {
		if isPseudoRune(r) && b.Len() < pseudoMaxLen {
			b.WriteRune(r)
		}
	}
	if p, err := ValidatePseudo(b.String()); err == nil {
		return p
	}
	id := strings.ReplaceAll(userID, "-", "")
	if len(id) > 8 {
		id = id[:8]
	}
	return "user-" + id
}

func isPseudoRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' || r == '.'
}
