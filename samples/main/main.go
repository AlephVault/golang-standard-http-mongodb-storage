package main

import (
	"context"
	"github.com/AlephVault/golang-standard-http-mongodb-storage/app"
	"github.com/AlephVault/golang-standard-http-mongodb-storage/core/auth"
	"github.com/AlephVault/golang-standard-http-mongodb-storage/core/dsl"
	"github.com/AlephVault/golang-standard-http-mongodb-storage/samples/payments"
	"github.com/AlephVault/golang-standard-http-mongodb-storage/samples/universe"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func LaunchServer() {
	settings := &dsl.Settings{
		Debug:      true,
		Connection: dsl.Connection{}, // Default connection data or from env. vars.
		Global: dsl.Global{
			ListMaxResults: 20,
		},
		Auth: dsl.Auth{
			TableRef: dsl.TableRef{
				Db:         "",
				Collection: "",
			},
		},
		Resources: map[string]dsl.Resource{
			"universe": universe.UniverseResource,
			"payments": payments.PaymentsResource,
		},
	}

	if application, err := app.MakeServer(settings, nil, func(client *mongo.Client, settings *dsl.Settings) {
		collection := client.Database(settings.Auth.Db).Collection(settings.Auth.Collection)
		ctx := context.Background()
		token := auth.AuthToken{}
		if result := collection.FindOne(ctx, bson.M{"_deleted": bson.M{"$ne": true}}).Decode(&token); result != nil {
			if _, err := collection.InsertOne(ctx, &auth.AuthToken{
				ApiKey:     "sample-abcdef",
				ValidUntil: nil,
				Permissions: bson.M{
					"universe": bson.A{"read", "write", "delete"},
				},
			}); err != nil {
				panic(err)
			}
		}
	}); err != nil {
		// Remember this is an example.
		panic(err)
	} else {
		// It will panic only on error.
		panic(application.Run("0.0.0.0:8888"))
	}
}

func main() {
	LaunchServer()
}
