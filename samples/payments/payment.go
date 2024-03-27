package payments

import (
	"github.com/AlephVault/golang-standard-http-mongodb-storage/core/dsl"
	"github.com/AlephVault/golang-standard-http-mongodb-storage/core/formats"
	"github.com/AlephVault/golang-standard-http-mongodb-storage/core/responses"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"maps"
	"strings"
	"time"
)

// Payment is a payment record.
type Payment struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	FromAddr string             `validate:"required" bson:"from" json:"from"`
	Amount   int                `validate:"required,gt=0" bson:"amount" json:"amount"`
	When     formats.DateTime   `validate:"required" bson:"when" json:"when"`
}

var (
	PaymentsResource = dsl.Resource{
		Type: dsl.ListResource,
		TableRef: dsl.TableRef{
			Db:         "mydb",
			Collection: "payments",
		},
		SoftDelete: true,
		ModelType:  dsl.ModelType[Payment],
		// Projection: bson.M{"foo": "bar"},
		ItemProjection: bson.M{"from": 1, "amount": 1, "when": 1},
		Methods: map[string]dsl.ResourceMethod{
			"get-from": {
				Type: dsl.View,
				Handler: func(context echo.Context, client *mongo.Client, resource, method string, collection *mongo.Collection, validatorMaker func() *validator.Validate, filter bson.M) error {
					ctx := context.Request().Context()
					var from = ""
					echo.QueryParamsBinder(context).String("from", &from)
					from = strings.TrimSpace(from)

					filter_ := bson.M{}
					maps.Copy(filter_, filter)
					if from != "" {
						filter_["from"] = from
					} else {
						return responses.OkWith(context, []Payment{})
					}

					if cursor, err := collection.Find(ctx, filter_); err != nil {
						return responses.InternalError(context)
					} else {
						elements := []Payment{}
						for cursor.Next(ctx) {
							element := Payment{}
							if err := cursor.Decode(&element); err != nil {
								return responses.InternalError(context)
							}
							elements = append(elements, element)
						}
						return responses.OkWith(context, elements)
					}
				},
			},
			"clear-from": {
				Type: dsl.Operation,
				Handler: func(context echo.Context, client *mongo.Client, resource, method string, collection *mongo.Collection, validatorMaker func() *validator.Validate, filter bson.M) error {
					ctx := context.Request().Context()
					var from = ""
					echo.QueryParamsBinder(context).String("from", &from)
					from = strings.TrimSpace(from)

					if from != "" {
						filter_ := bson.M{}
						maps.Copy(filter_, filter)
						filter_["from"] = from
						if _, err := collection.DeleteMany(ctx, filter_); err != nil {
							return responses.InternalError(context)
						}
					}

					return responses.Ok(context)
				},
			},
		},
		ItemMethods: map[string]dsl.ItemMethod{
			"put-now": {
				Type: dsl.Operation,
				Handler: func(context echo.Context, client *mongo.Client, resource, method string, collection *mongo.Collection, validatorMaker func() *validator.Validate, filter bson.M, id primitive.ObjectID) error {
					ctx := context.Request().Context()
					filter_ := bson.M{}
					maps.Copy(filter_, filter)
					filter_["_id"] = id

					// First, update.
					if result, err := collection.UpdateOne(ctx, filter_, bson.M{"$set": bson.M{"when": primitive.NewDateTimeFromTime(time.Now())}}); err != nil {
						return responses.InternalError(context)
					} else if result.ModifiedCount == 0 {
						return responses.NotFound(context)
					}

					// Then, retrieve.
					if result := collection.FindOne(ctx, bson.M{"_id": id}); result.Err() != nil {
						return responses.FindOneOperationError(context, result.Err())
					} else {
						v := Payment{}
						if err := result.Decode(&v); err != nil {
							return responses.InternalError(context)
						}
						return responses.OkWith(context, v)
					}
				},
			},
			"get-amount": {
				Type: dsl.View,
				Handler: func(context echo.Context, client *mongo.Client, resource, method string, collection *mongo.Collection, validatorMaker func() *validator.Validate, filter bson.M, id primitive.ObjectID) error {
					ctx := context.Request().Context()
					filter_ := bson.M{}
					maps.Copy(filter_, filter)
					filter_["_id"] = id

					if result := collection.FindOne(ctx, filter_); result.Err() != nil {
						return responses.FindOneOperationError(context, result.Err())
					} else {
						v := Payment{}
						if err := result.Decode(&v); err != nil {
							return responses.InternalError(context)
						}
						return responses.OkWith(context, v.Amount)
					}
				},
			},
		},
	}
)
