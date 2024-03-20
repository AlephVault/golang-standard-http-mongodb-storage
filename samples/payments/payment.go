package payments

import (
	"go.mongodb.org/mongo-driver/bson"
	"standard-http-mongodb-storage/core/dsl"
	"time"
)

// Payment is a payment record.
type Payment struct {
	FromAddr string    `validate:"required" bson:"from" json:"from"`
	Amount   int       `validate:"required,gt=0" bson:"amount" json:"amount"`
	When     time.Time `validate:"required" bson:"when" json:"when"`
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
	}
)
