package dsl

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// MethodType is the method operation type (implies the method to use).
type MethodType uint

const (
	View MethodType = iota
	Operation
)

// ListMethodHandler is a method that handles a specific collection
// and some filtering data, related to the whole collection.
type ListMethodHandler func(
	context *gin.Context, client *mongo.Client, resource, method, db, collection string, filter bson.M,
)

// ListMethod stands for a method entry which involves a handler and
// also telling whether it is a view or an operator. This handler is
// related to the whole list.
type ListMethod struct {
	Type    MethodType        `validate:"min=0,max=1"`
	Handler ListMethodHandler `validate:"required"`
}

// ItemMethodHandler is a method that handles a specific collection
// and some filtering data, now related to an item in particular.
type ItemMethodHandler func(
	context *gin.Context, client *mongo.Client, resource, method, db, collection string, filter bson.M,
	id primitive.ObjectID,
)

// ItemMethod stands for a method entry which involves a handler and
// also telling whether it is a view or an operator. This handler is
// related to a particular item.
type ItemMethod struct {
	Type    MethodType        `validate:"min=0,max=1"`
	Handler ListMethodHandler `validate:"required"`
}
