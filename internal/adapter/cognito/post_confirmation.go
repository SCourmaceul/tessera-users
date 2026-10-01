// Package cognito est l'adaptateur du trigger Cognito post-confirmation : il crée le profil
// d'un utilisateur dès que son inscription est confirmée.
package cognito

import (
	"context"

	"github.com/aws/aws-lambda-go/events"

	"github.com/SCourmaceul/tessera-users/internal/app"
)

// triggerConfirmSignUp est la source du trigger à la confirmation d'une inscription. Le même
// trigger est appelé après une réinitialisation de mot de passe (PostConfirmation_ConfirmForgotPassword),
// qui ne crée rien.
const triggerConfirmSignUp = "PostConfirmation_ConfirmSignUp"

// PostConfirmation renvoie le handler du trigger. Cognito attend l'événement reçu en retour ;
// une erreur est remontée à l'utilisateur qui confirme son inscription.
func PostConfirmation(uc app.RegisterUser) func(context.Context, events.CognitoEventUserPoolsPostConfirmation) (events.CognitoEventUserPoolsPostConfirmation, error) {
	return func(ctx context.Context, evt events.CognitoEventUserPoolsPostConfirmation) (events.CognitoEventUserPoolsPostConfirmation, error) {
		if evt.TriggerSource != triggerConfirmSignUp {
			return evt, nil
		}
		attrs := evt.Request.UserAttributes
		err := uc.Execute(ctx, app.RegisterInput{
			UserID:            attrs["sub"],
			Email:             attrs["email"],
			PreferredUsername: attrs["preferred_username"],
		})
		return evt, err
	}
}
