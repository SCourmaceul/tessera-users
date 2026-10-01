// Package app contient les cas d'usage du service users. Ils déclarent leurs ports
// (Repository, event.Publisher) et ne savent pas qu'ils tournent dans une Lambda.
package app

import (
	"context"
	"time"

	"github.com/SCourmaceul/tessera-users/internal/domain"
)

// Source identifie le service dans les événements publiés.
const Source = "tessera-users"

// EventUserCreated est publié à la création du profil, sans espace.
const EventUserCreated = "user.created"

// Repository est le port de stockage des profils et des appartenances aux espaces.
type Repository interface {
	// CreateProfile renvoie ErrProfileExists si le profil existe déjà.
	CreateProfile(ctx context.Context, p domain.Profile) error
	// GetProfile renvoie domain.ErrUserNotFound si le profil n'existe pas.
	GetProfile(ctx context.Context, userID string) (domain.Profile, error)
	SaveProfile(ctx context.Context, p domain.Profile) error
	// CreateMember renvoie domain.ErrAlreadyMember si l'utilisateur a déjà rejoint l'espace.
	CreateMember(ctx context.Context, m domain.Member) error
	// GetMember renvoie domain.ErrUserNotFound si l'utilisateur n'a pas rejoint l'espace.
	GetMember(ctx context.Context, space, userID string) (domain.Member, error)
	SaveMember(ctx context.Context, m domain.Member) error
	ListMembers(ctx context.Context, space string) ([]domain.Member, error)
}

// Clock donne l'heure courante ; nil vaut time.Now. Remplacée dans les tests.
type Clock func() time.Time

func (c Clock) now() time.Time {
	if c == nil {
		return time.Now().UTC()
	}
	return c().UTC()
}

// Me est le profil de l'utilisateur connecté et ses appartenances aux espaces rejoints.
type Me struct {
	Profile domain.Profile
	Members []domain.Member
}

func loadMe(ctx context.Context, repo Repository, p domain.Profile) (Me, error) {
	me := Me{Profile: p, Members: make([]domain.Member, 0, len(p.Spaces))}
	for _, space := range p.Spaces {
		m, err := repo.GetMember(ctx, space, p.ID)
		if err != nil {
			return Me{}, err
		}
		me.Members = append(me.Members, m)
	}
	return me, nil
}
