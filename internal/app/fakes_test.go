package app

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/SCourmaceul/tetra-kit/event"

	"github.com/SCourmaceul/tessera-users/internal/domain"
)

var testNow = time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)

func fixedClock() time.Time { return testNow }

type memberKey struct{ space, userID string }

// fakeRepo est un Repository en mémoire.
type fakeRepo struct {
	profiles map[string]domain.Profile
	members  map[memberKey]domain.Member
	failSave error
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{profiles: map[string]domain.Profile{}, members: map[memberKey]domain.Member{}}
}

func (r *fakeRepo) CreateProfile(_ context.Context, p domain.Profile) error {
	if _, ok := r.profiles[p.ID]; ok {
		return ErrProfileExists
	}
	r.profiles[p.ID] = p
	return nil
}

func (r *fakeRepo) GetProfile(_ context.Context, id string) (domain.Profile, error) {
	p, ok := r.profiles[id]
	if !ok {
		return domain.Profile{}, domain.ErrUserNotFound
	}
	p.Spaces = slices.Clone(p.Spaces)
	return p, nil
}

func (r *fakeRepo) SaveProfile(_ context.Context, p domain.Profile) error {
	if r.failSave != nil {
		return r.failSave
	}
	r.profiles[p.ID] = p
	return nil
}

func (r *fakeRepo) CreateMember(_ context.Context, m domain.Member) error {
	k := memberKey{m.Space, m.UserID}
	if _, ok := r.members[k]; ok {
		return domain.ErrAlreadyMember
	}
	r.members[k] = m
	return nil
}

func (r *fakeRepo) GetMember(_ context.Context, space, id string) (domain.Member, error) {
	m, ok := r.members[memberKey{space, id}]
	if !ok {
		return domain.Member{}, domain.ErrUserNotFound
	}
	return m, nil
}

func (r *fakeRepo) SaveMember(_ context.Context, m domain.Member) error {
	r.members[memberKey{m.Space, m.UserID}] = m
	return nil
}

func (r *fakeRepo) ListMembers(_ context.Context, space string) ([]domain.Member, error) {
	var out []domain.Member
	for k, m := range r.members {
		if k.space == space {
			out = append(out, m)
		}
	}
	return out, nil
}

// fakePublisher enregistre les événements publiés.
type fakePublisher struct {
	published []string
	err       error
}

func (p *fakePublisher) Publish(_ context.Context, eventType, space string, data any) (event.Event, error) {
	if p.err != nil {
		return event.Event{}, p.err
	}
	p.published = append(p.published, eventType+"@"+space)
	return event.New(eventType, Source, space, data)
}

var errBoom = errors.New("boom")
