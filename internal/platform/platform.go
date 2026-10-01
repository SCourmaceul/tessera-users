// Package platform construit les dépendances communes aux Lambdas du service à partir de
// l'environnement : configuration AWS, repository DynamoDB, publisher SNS.
//
// Variables : TABLE_NAME (table du service), EVENTS_TOPIC_ARN (topic tetra-events, pour les
// Lambdas qui publient). En local, AWS_ENDPOINT_URL=http://localhost:4566 vise LocalStack.
package platform

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sns"

	"github.com/SCourmaceul/tetra-kit/adapter/dynamo"
	"github.com/SCourmaceul/tetra-kit/adapter/eventbus"

	userstore "github.com/SCourmaceul/tessera-users/internal/adapter/dynamo"
	"github.com/SCourmaceul/tessera-users/internal/app"
)

// Deps regroupe les dépendances d'une Lambda.
type Deps struct {
	AWS   aws.Config
	Store userstore.Store
}

// MustLoad journalise en JSON et charge la configuration AWS et le repository. Une
// configuration invalide fait échouer le démarrage de la Lambda (init error).
func MustLoad() Deps {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(fmt.Errorf("configuration AWS : %w", err))
	}
	table := dynamo.NewTable(dynamodb.NewFromConfig(cfg), mustEnv("TABLE_NAME"))
	return Deps{AWS: cfg, Store: userstore.Store{Table: table}}
}

// Publisher renvoie le publisher du topic tetra-events.
func (d Deps) Publisher() *eventbus.SNSPublisher {
	return eventbus.NewSNSPublisher(sns.NewFromConfig(d.AWS), mustEnv("EVENTS_TOPIC_ARN"), app.Source)
}

func mustEnv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		panic(fmt.Errorf("variable d'environnement %s manquante", name))
	}
	return v
}
