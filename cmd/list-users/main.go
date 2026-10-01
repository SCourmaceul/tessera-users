// GET /api/v1/users : membres de l'espace.
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
	lambda.Start(lambdahttp.Wrap(httpadapter.ListUsers(app.ListUsers{Repo: deps.Store})))
}
