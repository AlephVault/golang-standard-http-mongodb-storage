package dsl

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
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

// ResourceMethodHandler is a method that handles a specific
// collection and some filtering data, related to the whole
// collection. For simple resources, this will imply the only
// record existing in it.
type ResourceMethodHandler func(
	context echo.Context, client *mongo.Client, resource, method string, collection *mongo.Collection,
	validatorMaker func() *validator.Validate, filter bson.M,
) error

// ResourceMethod stands for a method entry which involves a handler and
// also telling whether it is a view or an operator. This handler is
// related to the whole list.
type ResourceMethod struct {
	Type    MethodType            `validate:"min=0,max=1"`
	Handler ResourceMethodHandler `validate:"required"`
}

// ItemMethodHandler is a method that handles a specific collection
// and some filtering data, now related to an item in particular.
type ItemMethodHandler func(
	context echo.Context, client *mongo.Client, resource, method string, collection *mongo.Collection,
	validatorMaker func() *validator.Validate, filter bson.M, id primitive.ObjectID,
) error

// ItemMethod stands for a method entry which involves a handler and
// also telling whether it is a view or an operator. This handler is
// related to a particular item.
type ItemMethod struct {
	Type    MethodType        `validate:"min=0,max=1"`
	Handler ItemMethodHandler `validate:"required"`
}
