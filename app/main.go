package app

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"log/slog"
	"net/http"
	"os"
	"standard-http-mongodb-storage/core/dsl"
	"standard-http-mongodb-storage/core/responses"
	"standard-http-mongodb-storage/core/validation"
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

type customResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *customResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func wrapStatus(c *gin.Context) {
	buff := new(bytes.Buffer)
	c.Writer = &customResponseWriter{body: buff, ResponseWriter: c.Writer}

	c.Next()

	// After request handler execution
	if !strings.Contains(c.Writer.Header().Get("Content-Type"), "application/json") {
		switch c.Writer.Status() {
		case http.StatusNotFound:
			c.Abort()
			responses.NotFound(c)
		case http.StatusInternalServerError:
			c.Abort()
			responses.InternalError(c)
		case http.StatusMethodNotAllowed:
			c.Abort()
			responses.MethodNotAllowed(c)
		case http.StatusForbidden:
			c.Abort()
			responses.AuthForbidden(c)
		case http.StatusUnauthorized:
			c.Abort()
			responses.AuthNotFound(c)
		}
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
	router := gin.New()
	router.Use(gin.Logger(), gin.CustomRecovery(func(c *gin.Context, err any) {
		if err != nil {
			log.Printf("panic recovered: %v", err)
		}
		c.Abort()
		responses.InternalError(c)
	}), wrapStatus)

	// Configure the endpoints.
	for resourceKey, resource := range settings.Resources {
		registerEndpoints(client, router, resourceKey, &resource, &settings.Auth, resourcesValidatorMaker, logger)
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
