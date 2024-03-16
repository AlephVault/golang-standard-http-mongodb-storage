package app

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log/slog"
	"os"
	"standard-http-mongodb-storage/core/dsl"
	"standard-http-mongodb-storage/core/validation"
)

type Panicked struct {
	With any
}

func (panicked *Panicked) Error() string {
	return fmt.Sprintf("Panicked with: %v", panicked.With)
}

// MakeServer is used to create a server. The connection
// is created and established, but the server is not run
// immediately.
func MakeServer(
	settings *dsl.Settings, setupValidator func(*validator.Validate),
	setup func(*mongo.Client, *dsl.Settings),
) (app *Application, err error) {
	defer func() {
		if v := recover(); v != nil {
			app, err = nil, &Panicked{v}
		}
	}()
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
	level := slog.LevelInfo
	if settings.Debug {
		level = slog.LevelDebug
	}
	logger := slog.NewLogLogger(slog.NewTextHandler(os.Stdout, nil), level)

	// Configure the endpoints.
	router := gin.Default()
	for resourceKey, resource := range settings.Resources {
		registerEndpoints(client, router, resourceKey, &resource, &settings.Auth, logger)
	}

	// Create the final application object.
	app = &Application{
		router: router,
	}

	// Make a setup, using the client and the settings.
	if setup != nil {
		setup(client, settings)
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
				Name:   &name,
				Sparse: &sparse,
				Unique: &unique,
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
				var type_ any
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
					type_ = 1
				}
				fieldsMap[field] = type_
			}
			if _, err = client.Database(resource.Db).Collection(resource.Collection).Indexes().CreateOne(
				bg, mongo.IndexModel{
					Keys: fieldsMap,
					Options: &options.IndexOptions{
						Name:   &name,
						Sparse: &sparse,
						Unique: &unique,
					},
				},
			); err != nil {
				return
			}
		}
	}

	return
}
