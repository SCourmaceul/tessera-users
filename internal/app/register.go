package app

import (
	"context"
	"errors"
	"log/slog"

	"github.com/SCourmaceul/tetra-kit/event"

	"github.com/SCourmaceul/tessera-users/internal/domain"
)

// ErrProfileExists est renvoyée par Repository.CreateProfile quand le profil existe déjà.
var ErrProfileExists = errors.New("app: le profil existe déjà")

// RegisterInput décrit l'utilisateur dont Cognito vient de confirmer l'inscription.
type RegisterInput struct {
	UserID            string
	Email             string
	PreferredUsername string
}

// UserCreated est la charge utile de l'événement user.created.
type UserCreated struct {
	UserID string `json:"user_id"`
	Pseudo string `json:"pseudo"`
}

// RegisterUser crée le profil public d'un nouvel utilisateur et publie user.created.
// Il est idempotent : un profil déjà créé est laissé tel quel, sans nouvel événement.
type RegisterUser struct {
	Repo   Repository
	Events event.Publisher
	Clock  Clock
}

func (uc RegisterUser) Execute(ctx context.Context, in RegisterInput) error {
	if in.UserID == "" {
		return errors.New("app: inscription sans identifiant utilisateur")
	}
	now := uc.Clock.now()
	p := domain.Profile{
		ID:        in.UserID,
		Pseudo:    domain.DefaultPseudo(in.UserID, in.PreferredUsername, in.Email),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := uc.Repo.CreateProfile(ctx, p); errors.Is(err, ErrProfileExists) {
		return nil
	} else if err != nil {
		return err
	}
	// Le profil est créé : un échec de publication ne doit pas faire échouer la confirmation
	// de l'inscription côté Cognito. Il est journalisé (voir BACKLOG : outbox).
	if _, err := uc.Events.Publish(ctx, EventUserCreated, "", UserCreated{UserID: p.ID, Pseudo: p.Pseudo}); err != nil {
		slog.Default().ErrorContext(ctx, "Échec de publication de user.created",
			slog.String("user_id", p.ID), slog.Any("error", err))
	}
	return nil
}
