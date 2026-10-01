package app

import (
	"context"
	"errors"
	"testing"

	"github.com/SCourmaceul/tessera-users/internal/domain"
)

func ptr(s string) *string { return &s }

func TestRegisterUser(t *testing.T) {
	ctx := context.Background()
	repo, events := newFakeRepo(), &fakePublisher{}
	uc := RegisterUser{Repo: repo, Events: events, Clock: fixedClock}
	in := RegisterInput{UserID: "u-1", Email: "ada@tetra.local"}

	if err := uc.Execute(ctx, in); err != nil {
		t.Fatal(err)
	}
	p := repo.profiles["u-1"]
	if p.Pseudo != "ada" || !p.CreatedAt.Equal(testNow) {
		t.Errorf("profil = %+v", p)
	}
	if len(events.published) != 1 || events.published[0] != "user.created@" {
		t.Errorf("événements = %v", events.published)
	}

	// Rejouer le trigger ne recrée rien et ne republie pas
	if err := uc.Execute(ctx, RegisterInput{UserID: "u-1", Email: "autre@tetra.local"}); err != nil {
		t.Fatal(err)
	}
	if repo.profiles["u-1"].Pseudo != "ada" || len(events.published) != 1 {
		t.Errorf("trigger rejoué : profil %+v, événements %v", repo.profiles["u-1"], events.published)
	}

	// Un échec de publication ne fait pas échouer l'inscription
	uc.Events = &fakePublisher{err: errBoom}
	if err := uc.Execute(ctx, RegisterInput{UserID: "u-2"}); err != nil {
		t.Errorf("publication en échec : err = %v", err)
	}
	if err := uc.Execute(ctx, RegisterInput{}); err == nil {
		t.Error("inscription sans identifiant acceptée")
	}
}

func seed(t *testing.T) *fakeRepo {
	t.Helper()
	repo := newFakeRepo()
	repo.profiles["u-1"] = domain.Profile{ID: "u-1", Pseudo: "ada"}
	repo.profiles["u-2"] = domain.Profile{ID: "u-2", Pseudo: "Bob"}
	return repo
}

func TestJoinSpaceAndGetMe(t *testing.T) {
	ctx := context.Background()
	repo := seed(t)
	join := JoinSpace{Repo: repo, Clock: fixedClock}

	m, err := join.Execute(ctx, "core", "u-1")
	if err != nil {
		t.Fatal(err)
	}
	if m.Role != domain.RoleMember || m.Pseudo != "ada" || !m.JoinedAt.Equal(testNow) {
		t.Errorf("membre = %+v", m)
	}
	if _, err := join.Execute(ctx, "core", "u-1"); !errors.Is(err, domain.ErrAlreadyMember) {
		t.Errorf("deuxième adhésion : err = %v", err)
	}
	if _, err := join.Execute(ctx, "nowhere", "u-1"); !errors.Is(err, domain.ErrInvalidSpace) {
		t.Errorf("espace inconnu : err = %v", err)
	}
	if _, err := join.Execute(ctx, "flux", "ghost"); !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("profil absent : err = %v", err)
	}

	me, err := GetMe{Repo: repo}.Execute(ctx, "u-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(me.Members) != 1 || me.Members[0].Space != "core" {
		t.Errorf("me = %+v", me)
	}
}

func TestJoinSpaceResumesInterruptedJoin(t *testing.T) {
	ctx := context.Background()
	repo := seed(t)
	repo.failSave = errBoom
	join := JoinSpace{Repo: repo, Clock: fixedClock}
	if _, err := join.Execute(ctx, "core", "u-1"); !errors.Is(err, errBoom) {
		t.Fatalf("err = %v, want errBoom", err)
	}
	// L'appartenance a été créée mais pas reportée dans le profil : un nouvel appel répare
	repo.failSave = nil
	if _, err := join.Execute(ctx, "core", "u-1"); err != nil {
		t.Fatalf("reprise : err = %v", err)
	}
	if !repo.profiles["u-1"].HasJoined("core") {
		t.Error("le profil ne liste pas l'espace rejoint")
	}
}

func TestUpdateMe(t *testing.T) {
	ctx := context.Background()
	repo := seed(t)
	if _, err := (JoinSpace{Repo: repo}).Execute(ctx, "flow", "u-1"); err != nil {
		t.Fatal(err)
	}
	uc := UpdateMe{Repo: repo, Clock: fixedClock}

	me, err := uc.Execute(ctx, "u-1", UpdateInput{Pseudo: ptr(" ada_l "), AvatarURL: ptr("https://cdn.tetra.local/a.png")})
	if err != nil {
		t.Fatal(err)
	}
	if me.Profile.Pseudo != "ada_l" || !me.Profile.UpdatedAt.Equal(testNow) {
		t.Errorf("profil = %+v", me.Profile)
	}
	// La copie du profil dans l'espace suit
	if m := repo.members[memberKey{"flow", "u-1"}]; m.Pseudo != "ada_l" || m.AvatarURL == "" {
		t.Errorf("membre non mis à jour : %+v", m)
	}
	// Un champ absent est inchangé
	me, err = uc.Execute(ctx, "u-1", UpdateInput{AvatarURL: ptr("")})
	if err != nil || me.Profile.Pseudo != "ada_l" || me.Profile.AvatarURL != "" {
		t.Errorf("mise à jour partielle : %+v, %v", me.Profile, err)
	}
	if _, err := uc.Execute(ctx, "u-1", UpdateInput{Pseudo: ptr("a b")}); !errors.Is(err, domain.ErrInvalidPseudo) {
		t.Errorf("pseudo invalide : err = %v", err)
	}
	if repo.profiles["u-1"].Pseudo != "ada_l" {
		t.Error("un pseudo invalide a été enregistré")
	}
}

func TestGetAndListUsers(t *testing.T) {
	ctx := context.Background()
	repo := seed(t)
	for _, id := range []string{"u-1", "u-2"} {
		if _, err := (JoinSpace{Repo: repo}).Execute(ctx, "ignite", id); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := (GetUser{Repo: repo}).Execute(ctx, "ignite", "u-2"); err != nil {
		t.Errorf("GetUser = %v", err)
	}
	// Un utilisateur est introuvable dans un espace qu'il n'a pas rejoint
	if _, err := (GetUser{Repo: repo}).Execute(ctx, "core", "u-2"); !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("GetUser hors espace : err = %v", err)
	}
	members, err := ListUsers{Repo: repo}.Execute(ctx, "ignite")
	if err != nil || len(members) != 2 || members[0].Pseudo != "ada" || members[1].Pseudo != "Bob" {
		t.Errorf("ListUsers = %+v, %v", members, err)
	}
}
