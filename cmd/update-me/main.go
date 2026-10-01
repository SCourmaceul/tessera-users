// PATCH /api/v1/users/me : modifie le pseudo et l'avatar.
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
	lambda.Start(lambdahttp.Wrap(httpadapter.UpdateMe(app.UpdateMe{Repo: deps.Store})))
}
