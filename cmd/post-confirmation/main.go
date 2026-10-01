// Trigger Cognito post-confirmation : crée le profil et publie user.created.
package main

import (
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/SCourmaceul/tessera-users/internal/adapter/cognito"
	"github.com/SCourmaceul/tessera-users/internal/app"
	"github.com/SCourmaceul/tessera-users/internal/platform"
)

func main() {
	deps := platform.MustLoad()
	lambda.Start(cognito.PostConfirmation(app.RegisterUser{Repo: deps.Store, Events: deps.Publisher()}))
}
