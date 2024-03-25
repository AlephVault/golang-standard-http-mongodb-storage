package payments

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"standard-http-mongodb-storage/core/dsl"
	"standard-http-mongodb-storage/core/formats"
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
		// TODO IMPLEMENT THIS EXAMPLE.
		Methods: map[string]dsl.ResourceMethod{
			"get-from": dsl.ResourceMethod{
				Type:    dsl.View,
				Handler: nil,
			},
			"clear-from": dsl.ResourceMethod{
				Type:    dsl.Operation,
				Handler: nil,
			},
		},
		ItemMethods: map[string]dsl.ItemMethod{
			"put-now": dsl.ItemMethod{
				Type:    dsl.Operation,
				Handler: nil,
			},
			"get-prive": dsl.ItemMethod{
				Type:    dsl.View,
				Handler: nil,
			},
		},
	}
)
