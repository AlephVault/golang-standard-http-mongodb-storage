package payments

import (
	"github.com/AlephVault/golang-standard-http-mongodb-storage/core/dsl"
	"github.com/AlephVault/golang-standard-http-mongodb-storage/core/formats"
	"github.com/AlephVault/golang-standard-http-mongodb-storage/core/impl"
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
						if success, err := impl.GetDocuments[Payment](context, cursor, &elements); success {
							return err
						} else {
							return responses.OkWith(context, elements)
						}
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
					v := Payment{}
					if success, err := impl.GetDocument(context, collection.FindOne(ctx, bson.M{"_id": id}), &v); success {
						return responses.OkWith(context, v)
					} else {
						return err
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

					v := Payment{}
					if success, err := impl.GetDocument(context, collection.FindOne(ctx, bson.M{"_id": id}), &v); success {
						return responses.OkWith(context, v.Amount)
					} else {
						return err
					}
				},
			},
		},
	}
)
