package auth

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AuthToken defines an authentication token.
type AuthToken struct {
	ID          primitive.ObjectID  `bson:"_id,omitempty"`
	ApiKey      string              `bson:"api-key"`
	ValidUntil  *primitive.DateTime `bson:"timestamp,omitempty"`
	Permissions bson.M              `bson:"permissions,omitempty"`
}
