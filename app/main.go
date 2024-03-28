package app

import (
	"context"
	"fmt"
	"github.com/AlephVault/golang-standard-http-mongodb-storage/core/dsl"
	"github.com/AlephVault/golang-standard-http-mongodb-storage/core/responses"
	"github.com/AlephVault/golang-standard-http-mongodb-storage/core/validation"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
)

// Panicked is a class that wraps a panicked value into an error.
type Panicked struct {
	With any
}

// Error stands for the implementation of the Error interface,
// always rendering the panicked value.
func (panicked *Panicked) Error() string {
	return fmt.Sprintf("Panicked with: %v", panicked.With)
}

func wrapStatus(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := next(c)

		// After request handler execution
		if !strings.Contains(c.Response().Header().Get("Content-Type"), "application/json") {
			switch c.Response().Status {
			case http.StatusNotFound:
				return responses.NotFound(c)
			case http.StatusInternalServerError:
				return responses.InternalError(c)
			case http.StatusMethodNotAllowed:
				return responses.MethodNotAllowed(c)
			case http.StatusForbidden:
				return responses.AuthForbidden(c)
			case http.StatusUnauthorized:
				return responses.AuthNotFound(c)
			}
		}

		// Otherwise, return whatever was returned.
		return err
	}
}

func capturePanic(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		defer func() {
			if v := recover(); v != nil {
				slog.Error(fmt.Sprintf("Panicked! %s\n\nHere:\n", v) + string(debug.Stack()))
				err = responses.InternalError(c)
			}
		}()

		return next(c)
	}
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

	slog.Info("Init::Preparing settings")
	settings.Prepare()

	// Attempt a connection and get the client. Also, prepare
	// the indices (if any is configured).
	slog.Info("Init::Making MongoDB connection")
	var client *mongo.Client
	if client, err = settings.Connection.Connect(); err != nil {
		return
	}
	slog.Info("Init::Preparing database indices")
	if err = prepareIndices(client, settings); err != nil {
		return
	}

	// Make the validator to use and validate the settings.
	slog.Info("Init::Validating the settings")
	if settings.Global.ListMaxResults <= 0 {
		settings.Global.ListMaxResults = dsl.DefaultListMaxSize
	}
	settingsValidator := validation.Validator()
	if err = settingsValidator.Struct(settings); err != nil {
		return
	}

	// Make the validator to use and validate the resources.
	resourcesValidatorMaker := func() *validator.Validate {
		resourcesValidator := validation.Validator()
		if setupValidator != nil {
			setupValidator(resourcesValidator)
		}
		return resourcesValidator
	}

	// Configure slog logging level.
	if settings.Debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create the router.
	slog.Info("Init::Starting the router")
	router := echo.New()
	router.Debug = settings.Debug
	router.Use(capturePanic, wrapStatus)

	// Configure the endpoints.
	slog.Info("Init::Defining the resources")
	for resourceKey, resource := range settings.Resources {
		registerEndpoints(
			client, router, resourceKey, &resource, &settings.Auth, resourcesValidatorMaker,
			settings.Global.ListMaxResults, logger,
		)
	}
	router.Any("/*", func(c echo.Context) error {
		return responses.NotFound(c)
	})

	// Create the final application object.
	slog.Info("Init::Defining the application and applying initial setup")
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
	slog.Info("Init/Indices::Getting the indices collection")
	authIndices := client.Database(settings.Auth.Db).Collection(settings.Auth.Collection).Indexes()
	name := "api-key"
	sparse := true
	unique := true
	slog.Info(fmt.Sprintf("Init/Indices::Creating indices for auth db=%s table=%s", settings.Auth.Db, settings.Auth.Collection))
	if _, err = authIndices.CreateOne(
		bg, mongo.IndexModel{
			Keys: bson.D{{Key: "api-key", Value: 1}}, // 1=Ascending.
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
			fieldsMap := bson.D{}
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
				fieldsMap = append(fieldsMap, bson.E{Key: field, Value: type_})
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
