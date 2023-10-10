package universe

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"standard-http-mongodb-storage/core/dsl"
	"standard-http-mongodb-storage/core/responses"
	"standard-http-mongodb-storage/samples/core"
	"strings"
)

// UniverseVersion stands for the version of the game's universe / layout.
type UniverseVersion struct {
	Major    uint `bson:"major" json:"major"`
	Minor    uint `bson:"minor" json:"minor"`
	Revision uint `bson:"revision" json:"revision"`
}

// Universe is a sample singleton for the whole game layout.
type Universe struct {
	Caption string          `validate:"required,gt=0" bson:"caption" json:"caption"`
	Motd    string          `validate:"required,gt=0" bson:"motd" json:"motd"`
	Version UniverseVersion `validate:"dive" bson:"version" json:"version"`
}

// SetMotdBody stands for the body for a "set-motd" method.
type SetMotdBody struct {
	Motd string `validate:"required,gt=0" json:"motd"`
}

// SetMotd changes the current motd of the universe.
func SetMotd(context *gin.Context, client *mongo.Client, resource, method, db, collection string, filter bson.M) {
	defer func() {
		if v := recover(); v != nil {
			responses.UnexpectedFormat(context)
		}
	}()

	if strings.Contains(strings.ToLower(context.GetHeader("Content-Type")), "application/json") {
		responses.UnexpectedFormat(context)
		return
	}

	body := SetMotdBody{}
	if err := context.ShouldBindJSON(&body); err != nil {
		responses.UnexpectedFormat(context)
		return
	} else if err := core.SampleValidator.Struct(&body); err != nil {
		responses.InvalidFormat(context, err.(validator.ValidationErrors))
		return
	}

	if result, err := client.Database(db).Collection(collection).UpdateOne(
		context, filter, bson.M{"$set": bson.M{"motd": body.Motd}},
	); err != nil {
		responses.InternalError(context)
		return
	} else if result.ModifiedCount == 1 {
		responses.Ok(context)
	} else {
		responses.NotFound(context)
	}
}

// GetVersion is a handler that returns the current version.
func GetVersion(context *gin.Context, client *mongo.Client, resource, method, db, collection string, filter bson.M) {
	if result := client.Database(db).Collection(collection).FindOne(context, filter); result != nil && result.Err() == nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			responses.NotFound(context)
		} else {
			responses.InternalError(context)
		}
		return
	} else {
		universe := Universe{}
		if err := result.Decode(&universe); err != nil {
			responses.InternalError(context)
			return
		} else {
			responses.OkWith(context, universe.Version)
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
		ModelType:  Universe{},
		// ListProjection: bson.D{{"foo", "bar"}},
		ItemProjection: bson.D{{"caption", 1}, {"motd", 1}},
		ItemMethods: map[string]dsl.ItemMethod{
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
