package app

import (
	"context"
	"errors"
	"slices"
	"strings"

	"github.com/SCourmaceul/tetra-kit/space"

	"github.com/SCourmaceul/tessera-users/internal/domain"
)

// GetMe renvoie le profil de l'utilisateur connecté.
type GetMe struct{ Repo Repository }

func (uc GetMe) Execute(ctx context.Context, userID string) (Me, error) {
	p, err := uc.Repo.GetProfile(ctx, userID)
	if err != nil {
		return Me{}, err
	}
	return loadMe(ctx, uc.Repo, p)
}

// UpdateInput liste les champs modifiables du profil ; un champ nil est inchangé.
type UpdateInput struct {
	Pseudo    *string
	AvatarURL *string
}

// UpdateMe modifie le profil public de l'utilisateur connecté et en reporte la copie dans
// chaque espace rejoint.
type UpdateMe struct {
	Repo  Repository
	Clock Clock
}

func (uc UpdateMe) Execute(ctx context.Context, userID string, in UpdateInput) (Me, error) {
	p, err := uc.Repo.GetProfile(ctx, userID)
	if err != nil {
		return Me{}, err
	}
	if in.Pseudo != nil {
		if p.Pseudo, err = domain.ValidatePseudo(*in.Pseudo); err != nil {
			return Me{}, err
		}
	}
	if in.AvatarURL != nil {
		if p.AvatarURL, err = domain.ValidateAvatarURL(*in.AvatarURL); err != nil {
			return Me{}, err
		}
	}
	p.UpdatedAt = uc.Clock.now()
	if err := uc.Repo.SaveProfile(ctx, p); err != nil {
		return Me{}, err
	}
	me := Me{Profile: p, Members: make([]domain.Member, 0, len(p.Spaces))}
	for _, sp := range p.Spaces {
		m, err := uc.Repo.GetMember(ctx, sp, p.ID)
		if err != nil {
			return Me{}, err
		}
		m.Pseudo, m.AvatarURL = p.Pseudo, p.AvatarURL
		if err := uc.Repo.SaveMember(ctx, m); err != nil {
			return Me{}, err
		}
		me.Members = append(me.Members, m)
	}
	return me, nil
}

// JoinSpace fait rejoindre l'espace de la requête à l'utilisateur connecté.
type JoinSpace struct {
	Repo  Repository
	Clock Clock
}

func (uc JoinSpace) Execute(ctx context.Context, sp, userID string) (domain.Member, error) {
	if !space.Valid(sp) {
		return domain.Member{}, domain.ErrInvalidSpace
	}
	p, err := uc.Repo.GetProfile(ctx, userID)
	if err != nil {
		return domain.Member{}, err
	}
	if p.HasJoined(sp) {
		return domain.Member{}, domain.ErrAlreadyMember
	}
	now := uc.Clock.now()
	m := domain.NewMember(p, sp, now)
	if err := uc.Repo.CreateMember(ctx, m); errors.Is(err, domain.ErrAlreadyMember) {
		// Reprise d'un appel interrompu : l'appartenance existe mais le profil ne la liste pas
		if m, err = uc.Repo.GetMember(ctx, sp, userID); err != nil {
			return domain.Member{}, err
		}
	} else if err != nil {
		return domain.Member{}, err
	}
	p.Spaces = append(p.Spaces, sp)
	p.UpdatedAt = now
	if err := uc.Repo.SaveProfile(ctx, p); err != nil {
		return domain.Member{}, err
	}
	return m, nil
}

// GetUser renvoie le profil d'un utilisateur dans l'espace de la requête. Un utilisateur qui
// n'a pas rejoint l'espace y est introuvable.
type GetUser struct{ Repo Repository }

func (uc GetUser) Execute(ctx context.Context, sp, userID string) (domain.Member, error) {
	if userID == "" {
		return domain.Member{}, domain.ErrUserNotFound
	}
	return uc.Repo.GetMember(ctx, sp, userID)
}

// ListUsers liste les membres de l'espace de la requête, triés par pseudo.
type ListUsers struct{ Repo Repository }

func (uc ListUsers) Execute(ctx context.Context, sp string) ([]domain.Member, error) {
	members, err := uc.Repo.ListMembers(ctx, sp)
	if err != nil {
		return nil, err
	}
	slices.SortFunc(members, func(a, b domain.Member) int {
		return strings.Compare(strings.ToLower(a.Pseudo), strings.ToLower(b.Pseudo))
	})
	return members, nil
}
