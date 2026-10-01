// GET /api/v1/users/:id : profil d'un membre de l'espace.
package main

import (
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/SCourmaceul/tetra-kit/adapter/lambdahttp"

	httpadapter "github.com/SCourmaceul/tessera-users/internal/adapter/http"
	"github.com/SCourmaceul/tessera-users/internal/app"
	"github.com/SCourmaceul/tessera-users/internal/platform"
)

func main() {
	deps := platform.MustLoad()
	lambda.Start(lambdahttp.Wrap(httpadapter.GetUser(app.GetUser{Repo: deps.Store})))
}
