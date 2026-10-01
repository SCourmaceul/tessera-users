// POST /api/v1/users/me/spaces : rejoint l'espace de la requête.
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
	lambda.Start(lambdahttp.Wrap(httpadapter.JoinSpace(app.JoinSpace{Repo: deps.Store})))
}
