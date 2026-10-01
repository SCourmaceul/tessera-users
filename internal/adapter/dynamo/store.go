// Package dynamo implémente app.Repository sur la table DynamoDB du service (tessera-users).
//
// Deux sortes d'items :
//   - le profil, commun à tous les espaces : pk = USER#<id>, sk = PROFILE. C'est la seule clé
//     non préfixée par un espace, le profil n'appartenant à aucun ;
//   - l'appartenance à un espace : pk = SPACE#<espace>#MEMBERS, sk = USER#<id>. Une partition
//     par espace permet de lister ses membres d'une seule requête.
package dynamo

import (
	"context"
	"errors"
	"time"

	"github.com/SCourmaceul/tetra-kit/adapter/dynamo"

	"github.com/SCourmaceul/tessera-users/internal/app"
	"github.com/SCourmaceul/tessera-users/internal/domain"
)

const (
	profileSK    = "PROFILE"
	memberPrefix = "USER#"
)

type profileItem struct {
	dynamo.Key
	ID        string    `dynamodbav:"id"`
	Pseudo    string    `dynamodbav:"pseudo"`
	AvatarURL string    `dynamodbav:"avatar_url,omitempty"`
	Spaces    []string  `dynamodbav:"spaces,omitempty"`
	CreatedAt time.Time `dynamodbav:"created_at"`
	UpdatedAt time.Time `dynamodbav:"updated_at"`
}

type memberItem struct {
	dynamo.Key
	UserID     string    `dynamodbav:"user_id"`
	Space      string    `dynamodbav:"space"`
	Pseudo     string    `dynamodbav:"pseudo"`
	AvatarURL  string    `dynamodbav:"avatar_url,omitempty"`
	Role       string    `dynamodbav:"role"`
	Reputation int       `dynamodbav:"reputation"`
	JoinedAt   time.Time `dynamodbav:"joined_at"`
}

func profileKey(userID string) dynamo.Key {
	return dynamo.Key{PK: "USER#" + userID, SK: profileSK}
}

func membersPK(space string) string {
	return dynamo.PartitionKey(space, "MEMBERS")
}

func memberKey(space, userID string) dynamo.Key {
	return dynamo.Key{PK: membersPK(space), SK: memberPrefix + userID}
}

// Store est le repository DynamoDB du service.
type Store struct{ Table *dynamo.Table }

var _ app.Repository = Store{}

func (s Store) CreateProfile(ctx context.Context, p domain.Profile) error {
	err := s.Table.Create(ctx, toProfileItem(p))
	if errors.Is(err, dynamo.ErrAlreadyExists) {
		return app.ErrProfileExists
	}
	return err
}

func (s Store) GetProfile(ctx context.Context, userID string) (domain.Profile, error) {
	var it profileItem
	if err := s.Table.Get(ctx, profileKey(userID), &it); errors.Is(err, dynamo.ErrNotFound) {
		return domain.Profile{}, domain.ErrUserNotFound
	} else if err != nil {
		return domain.Profile{}, err
	}
	return domain.Profile{
		ID:        it.ID,
		Pseudo:    it.Pseudo,
		AvatarURL: it.AvatarURL,
		Spaces:    it.Spaces,
		CreatedAt: it.CreatedAt,
		UpdatedAt: it.UpdatedAt,
	}, nil
}

func (s Store) SaveProfile(ctx context.Context, p domain.Profile) error {
	return s.Table.Put(ctx, toProfileItem(p))
}

func (s Store) CreateMember(ctx context.Context, m domain.Member) error {
	err := s.Table.Create(ctx, toMemberItem(m))
	if errors.Is(err, dynamo.ErrAlreadyExists) {
		return domain.ErrAlreadyMember
	}
	return err
}

func (s Store) GetMember(ctx context.Context, space, userID string) (domain.Member, error) {
	var it memberItem
	if err := s.Table.Get(ctx, memberKey(space, userID), &it); errors.Is(err, dynamo.ErrNotFound) {
		return domain.Member{}, domain.ErrUserNotFound
	} else if err != nil {
		return domain.Member{}, err
	}
	return it.toDomain(), nil
}

func (s Store) SaveMember(ctx context.Context, m domain.Member) error {
	return s.Table.Put(ctx, toMemberItem(m))
}

func (s Store) ListMembers(ctx context.Context, space string) ([]domain.Member, error) {
	var items []memberItem
	if err := s.Table.Query(ctx, membersPK(space), memberPrefix, &items); err != nil {
		return nil, err
	}
	members := make([]domain.Member, len(items))
	for i, it := range items {
		members[i] = it.toDomain()
	}
	return members, nil
}

func toProfileItem(p domain.Profile) profileItem {
	return profileItem{
		Key:       profileKey(p.ID),
		ID:        p.ID,
		Pseudo:    p.Pseudo,
		AvatarURL: p.AvatarURL,
		Spaces:    p.Spaces,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

func toMemberItem(m domain.Member) memberItem {
	return memberItem{
		Key:        memberKey(m.Space, m.UserID),
		UserID:     m.UserID,
		Space:      m.Space,
		Pseudo:     m.Pseudo,
		AvatarURL:  m.AvatarURL,
		Role:       m.Role,
		Reputation: m.Reputation,
		JoinedAt:   m.JoinedAt,
	}
}

func (it memberItem) toDomain() domain.Member {
	return domain.Member{
		UserID:     it.UserID,
		Space:      it.Space,
		Pseudo:     it.Pseudo,
		AvatarURL:  it.AvatarURL,
		Role:       it.Role,
		Reputation: it.Reputation,
		JoinedAt:   it.JoinedAt,
	}
}
