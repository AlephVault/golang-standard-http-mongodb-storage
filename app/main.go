package app

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log/slog"
	"standard-http-mongodb-storage/core/dsl"
	"standard-http-mongodb-storage/core/validation"
)

// MakeServer is used to create a server. The connection
// is created and established, but the server is not run
// immediately.
func MakeServer(settings *dsl.Settings, setupValidator func(*validator.Validate)) (app *Application, err error) {
	settings.Prepare()

	// Attempt a connection and get the client. Also, prepare
	// the indices (if any is configured).
	var client *mongo.Client
	if client, err = settings.Connection.Connect(); err != nil {
		return
	} else if err = prepareIndices(client, settings); err != nil {
		return
	}

	// Make the validator to use and validate the settings.
	settingsValidator := validation.Validator()
	if setupValidator != nil {
		setupValidator(settingsValidator)
	}
	if err = settingsValidator.Struct(settings); err != nil {
		return
	}

	// Configure slog logging level.
	if settings.Debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	// Configure the endpoints.
	router := gin.Default()
	registerEndpoints(router)

	// Create the final application object.
	app = &Application{
		router: router,
	}
	return
}

func prepareIndices(client *mongo.Client, settings *dsl.Settings) (err error) {
	bg := context.Background()
	authIndices := client.Database(settings.Auth.Db).Collection(settings.Auth.Collection).Indexes()
	name := "api-key"
	sparse := true
	unique := true
	if _, err = authIndices.CreateOne(
		bg, mongo.IndexModel{
			Keys: bson.M{"api-key": 1}, // 1=Ascending.
			Options: &options.IndexOptions{
				ExpireAfterSeconds:      nil,
				Name:                    &name,
				Sparse:                  &sparse,
				StorageEngine:           nil,
				Unique:                  &unique,
				Version:                 nil,
				DefaultLanguage:         nil,
				LanguageOverride:        nil,
				TextVersion:             nil,
				Weights:                 nil,
				SphereVersion:           nil,
				Bits:                    nil,
				Max:                     nil,
				Min:                     nil,
				BucketSize:              nil,
				PartialFilterExpression: nil,
				Collation:               nil,
				WildcardProjection:      nil,
				Hidden:                  nil,
			},
		},
	); err != nil {
		return
	}

	for _, resource := range settings.Resources {
		for name, index := range resource.Indexes {
			unique := index.Unique
			fields := index.Fields
			fieldsMap := bson.M{}
			for _, field := range fields {
				var type_ interface{}
				switch field[0] {
				case '-':
					type_ = -1
				case '@':
					type_ = "2dsphere"
				case '#':
					type_ = "hashed"
				case '~':
					type_ = "text"
				default:
					type_ = "ascending"
				}
				fieldsMap[field] = type_
			}
			if _, err = client.Database(resource.Db).Collection(resource.Collection).Indexes().CreateOne(
				bg, mongo.IndexModel{
					Keys: bson.M{"api-key": 1}, // 1=Ascending.
					Options: &options.IndexOptions{
						ExpireAfterSeconds:      nil,
						Name:                    &name,
						Sparse:                  &sparse,
						StorageEngine:           nil,
						Unique:                  &unique,
						Version:                 nil,
						DefaultLanguage:         nil,
						LanguageOverride:        nil,
						TextVersion:             nil,
						Weights:                 nil,
						SphereVersion:           nil,
						Bits:                    nil,
						Max:                     nil,
						Min:                     nil,
						BucketSize:              nil,
						PartialFilterExpression: nil,
						Collation:               nil,
						WildcardProjection:      nil,
						Hidden:                  nil,
					},
				},
			); err != nil {
				return
			}
		}
	}

	return
}
