package universe

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"standard-http-mongodb-storage/core/dsl"
	"standard-http-mongodb-storage/core/responses"
	"standard-http-mongodb-storage/samples/core"
	"strings"
)

// UniverseVersion stands for the version of the game's universe / layout.
type UniverseVersion struct {
	Major    uint `validate:"required" bson:"major" json:"major"`
	Minor    uint `validate:"required" bson:"minor" json:"minor"`
	Revision uint `validate:"required" bson:"revision" json:"revision"`
}

// Universe is a sample singleton for the whole game layout.
type Universe struct {
	ID      primitive.ObjectID `bson:"_id"`
	Caption string             `validate:"required,gt=0" bson:"caption" json:"caption"`
	Motd    string             `validate:"required,gt=0" bson:"motd" json:"motd"`
	Version UniverseVersion    `validate:"required,dive" bson:"version" json:"version"`
}

// SetMotdBody stands for the body for a "set-motd" method.
type SetMotdBody struct {
	Motd string `validate:"required,gt=0" json:"motd"`
}

// SetMotd changes the current motd of the universe.
func SetMotd(
	context echo.Context, client *mongo.Client, resource, method string, collection *mongo.Collection,
	validatorMaker func() *validator.Validate, filter bson.M,
) (err error) {
	defer func() {
		if v := recover(); v != nil {
			err = responses.UnexpectedFormat(context)
		}
	}()

	if strings.Contains(strings.ToLower(context.Request().Header.Get("Content-Type")), "application/json") {
		return responses.UnexpectedFormat(context)
	}

	body := SetMotdBody{}
	if err := (&echo.DefaultBinder{}).BindBody(context, &body); err != nil {
		return responses.UnexpectedFormat(context)
	} else if err := core.SampleValidator.Struct(&body); err != nil {
		return responses.InvalidFormat(context, err.(validator.ValidationErrors))
	}

	if result, err := collection.UpdateOne(
		context.Request().Context(), filter, bson.M{"$set": bson.M{"motd": body.Motd}},
	); err != nil {
		return responses.InternalError(context)
	} else if result.ModifiedCount == 1 {
		return responses.Ok(context)
	} else {
		return responses.NotFound(context)
	}
}

// GetVersion is a handler that returns the current version.
func GetVersion(
	context echo.Context, client *mongo.Client, resource, method string, collection *mongo.Collection,
	validatorMaker func() *validator.Validate, filter bson.M,
) error {
	if result := collection.FindOne(context.Request().Context(), filter); result != nil && result.Err() == nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			return responses.NotFound(context)
		} else {
			return responses.InternalError(context)
		}
	} else {
		universe := Universe{}
		if err := result.Decode(&universe); err != nil {
			return responses.InternalError(context)
		} else {
			return responses.OkWith(context, universe.Version)
		}
	}
}

var (
	UniverseResource = dsl.Resource{
		Type: dsl.SimpleResource,
		TableRef: dsl.TableRef{
			Db:         "mydb",
			Collection: "universe",
		},
		SoftDelete: true,
		ModelType:  dsl.ModelType[Universe],
		// Projection: bson.M{"foo": "bar"},
		Projection: bson.M{"caption": 1, "motd": 1},
		Methods: map[string]dsl.ResourceMethod{
			"set-motd": {
				Type:    dsl.Operation,
				Handler: SetMotd,
			},
			"version": {
				Type:    dsl.View,
				Handler: GetVersion,
			},
		},
	}
)
