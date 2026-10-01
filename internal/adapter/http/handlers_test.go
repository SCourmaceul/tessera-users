package http

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/SCourmaceul/tetra-kit/adapter/lambdahttp"

	"github.com/SCourmaceul/tessera-users/internal/app"
	"github.com/SCourmaceul/tessera-users/internal/domain"
)

// stubRepo est un app.Repository minimal : un profil u-1 membre de core.
type stubRepo struct {
	profile domain.Profile
	member  domain.Member
}

func newStubRepo() *stubRepo {
	at := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	return &stubRepo{
		profile: domain.Profile{ID: "u-1", Pseudo: "ada", Spaces: []string{"core"}, CreatedAt: at},
		member:  domain.Member{UserID: "u-1", Space: "core", Pseudo: "ada", Role: "member", JoinedAt: at},
	}
}

func (r *stubRepo) CreateProfile(context.Context, domain.Profile) error { return nil }
func (r *stubRepo) GetProfile(_ context.Context, id string) (domain.Profile, error) {
	if id != r.profile.ID {
		return domain.Profile{}, domain.ErrUserNotFound
	}
	return r.profile, nil
}
func (r *stubRepo) SaveProfile(_ context.Context, p domain.Profile) error { r.profile = p; return nil }
func (r *stubRepo) CreateMember(context.Context, domain.Member) error     { return domain.ErrAlreadyMember }
func (r *stubRepo) GetMember(_ context.Context, space, id string) (domain.Member, error) {
	if space != r.member.Space || id != r.member.UserID {
		return domain.Member{}, domain.ErrUserNotFound
	}
	return r.member, nil
}
func (r *stubRepo) SaveMember(_ context.Context, m domain.Member) error { r.member = m; return nil }
func (r *stubRepo) ListMembers(_ context.Context, space string) ([]domain.Member, error) {
	if space != r.member.Space {
		return nil, nil
	}
	return []domain.Member{r.member}, nil
}

func call(t *testing.T, h lambdahttp.HandlerFunc, req lambdahttp.Request) (int, map[string]any) {
	t.Helper()
	resp, err := lambdahttp.Wrap(h)(context.Background(), req)
	if err != nil {
		t.Fatalf("Wrap a renvoyé une erreur : %v", err)
	}
	var body map[string]any
	if resp.Body != "" {
		if err := json.Unmarshal([]byte(resp.Body), &body); err != nil {
			t.Fatalf("corps non JSON : %q", resp.Body)
		}
	}
	return resp.StatusCode, body
}

func authed(space string) lambdahttp.Request {
	return lambdahttp.Request{RequestContext: lambdahttp.RequestContext{Space: space, UserID: "u-1", UserRoles: []string{"member"}}}
}

func TestGetMe(t *testing.T) {
	h := GetMe(app.GetMe{Repo: newStubRepo()})
	status, body := call(t, h, authed("core"))
	if status != http.StatusOK || body["pseudo"] != "ada" {
		t.Fatalf("%d %v", status, body)
	}
	if spaces := body["spaces"].([]any); len(spaces) != 1 || spaces[0].(map[string]any)["space"] != "core" {
		t.Errorf("spaces = %v", body["spaces"])
	}
	if status, body := call(t, h, lambdahttp.Request{}); status != http.StatusUnauthorized || body["error"] != "UNAUTHORIZED" {
		t.Errorf("anonyme : %d %v", status, body)
	}
}

func TestUpdateMe(t *testing.T) {
	h := UpdateMe(app.UpdateMe{Repo: newStubRepo()})
	req := authed("core")
	req.Body = `{"pseudo":"ada_l"}`
	if status, body := call(t, h, req); status != http.StatusOK || body["pseudo"] != "ada_l" {
		t.Errorf("%d %v", status, body)
	}
	req.Body = `{"pseudo":"a"}`
	if status, body := call(t, h, req); status != http.StatusBadRequest || body["error"] != "INVALID_PSEUDO" {
		t.Errorf("pseudo invalide : %d %v", status, body)
	}
	req.Body = `{"email":"x@y.z"}`
	if status, body := call(t, h, req); status != http.StatusBadRequest || body["error"] != "INVALID_BODY" {
		t.Errorf("champ inconnu : %d %v", status, body)
	}
}

func TestJoinSpace(t *testing.T) {
	h := JoinSpace(app.JoinSpace{Repo: newStubRepo()})
	if status, body := call(t, h, authed("core")); status != http.StatusConflict || body["error"] != "ALREADY_MEMBER" {
		t.Errorf("déjà membre : %d %v", status, body)
	}
}

func TestGetAndListUsers(t *testing.T) {
	repo := newStubRepo()
	req := authed("core")
	req.PathParameters = map[string]string{"id": "u-1"}
	if status, body := call(t, GetUser(app.GetUser{Repo: repo}), req); status != http.StatusOK || body["id"] != "u-1" || body["role"] != "member" {
		t.Errorf("GetUser : %d %v", status, body)
	}
	req.RequestContext.Space = "flux"
	if status, body := call(t, GetUser(app.GetUser{Repo: repo}), req); status != http.StatusNotFound || body["error"] != "USER_NOT_FOUND" {
		t.Errorf("GetUser hors espace : %d %v", status, body)
	}
	status, body := call(t, ListUsers(app.ListUsers{Repo: repo}), authed("core"))
	if status != http.StatusOK || len(body["users"].([]any)) != 1 {
		t.Errorf("ListUsers : %d %v", status, body)
	}
	// Un espace vide renvoie une liste vide, pas null
	if _, body := call(t, ListUsers(app.ListUsers{Repo: repo}), authed("flow")); body["users"] == nil {
		t.Errorf("ListUsers vide : %v", body)
	}
}
