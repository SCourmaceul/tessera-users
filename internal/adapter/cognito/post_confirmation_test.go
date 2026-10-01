package cognito

import (
	"context"
	"testing"

	"github.com/aws/aws-lambda-go/events"

	"github.com/SCourmaceul/tetra-kit/event"

	"github.com/SCourmaceul/tessera-users/internal/app"
	"github.com/SCourmaceul/tessera-users/internal/domain"
)

type recordingRepo struct {
	app.Repository
	created []domain.Profile
}

func (r *recordingRepo) CreateProfile(_ context.Context, p domain.Profile) error {
	r.created = append(r.created, p)
	return nil
}

type nopPublisher struct{}

func (nopPublisher) Publish(_ context.Context, t, s string, d any) (event.Event, error) {
	return event.New(t, app.Source, s, d)
}

func TestPostConfirmation(t *testing.T) {
	repo := &recordingRepo{}
	h := PostConfirmation(app.RegisterUser{Repo: repo, Events: nopPublisher{}})

	evt := events.CognitoEventUserPoolsPostConfirmation{}
	evt.TriggerSource = "PostConfirmation_ConfirmSignUp"
	evt.Request.UserAttributes = map[string]string{"sub": "u-1", "email": "ada@tetra.local"}
	out, err := h(context.Background(), evt)
	if err != nil {
		t.Fatal(err)
	}
	if out.TriggerSource != evt.TriggerSource {
		t.Error("l'événement n'est pas renvoyé à Cognito")
	}
	if len(repo.created) != 1 || repo.created[0].ID != "u-1" || repo.created[0].Pseudo != "ada" {
		t.Errorf("profils créés = %+v", repo.created)
	}

	// Une réinitialisation de mot de passe ne crée rien
	evt.TriggerSource = "PostConfirmation_ConfirmForgotPassword"
	if _, err := h(context.Background(), evt); err != nil || len(repo.created) != 1 {
		t.Errorf("mot de passe oublié : err = %v, profils = %d", err, len(repo.created))
	}
}
