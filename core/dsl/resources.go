package dsl

import (
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
)

// ResourceType is an enumeration to tell whether it is a list
// resource (standard) or one with just one record.
type ResourceType uint

const (
	ListResource ResourceType = iota
	SimpleResource
)

// ResourceVerb is an enumeration to tell the allowed verbs into
// these resources: list, create, read, replace, update, delete.
type ResourceVerb uint

const (
	ListVerb ResourceVerb = iota
	CreateVerb
	ReadVerb
	ReplaceVerb
	UpdateVerb
	DeleteVerb
	LastVerb = DeleteVerb
)

// ModelTypeFunction is a function that returns the model to use.
type ModelTypeFunction func() interface{}

// ModelType can be used to instantiate any struct. It should be
// used in the DSL of a resource.
func ModelType[T any]() interface{} {
	var t T
	return t
}

// Resource stands for the rules regarding a particular
// resource (in the end, a collection).
type Resource struct {
	TableRef       `validate:"dive"`
	Type           ResourceType `validate:"min=0,max=1"`
	Sort           bson.D
	Filter         bson.M
	ItemProjection bson.D `validate:"excluded_if=Type 1"`
	Projection     bson.D
	ItemMethods    map[string]ItemMethod     `validate:"excluded_if=Type 1,dive,keys,method-name,endkeys,dive"`
	Methods        map[string]ResourceMethod `validate:"dive,keys,method-name,endkeys,dive"`
	ModelType      ModelTypeFunction         `validate:"required"`
	Verbs          []ResourceVerb            `validate:"dive,verbs"`
	SoftDelete     bool
	ListMaxResults uint
	Indexes        map[string]Index `validate:"dive,keys,mdb-name,endkeys"`
}

// Resources belong to a mapping.
type Resources map[string]Resource

// ValidateVerbs does a custom validation function on the verbs:
// If the resource is of list type ListResource then allow all
// the verbs. Otherwise, allow all but the List verb.
func ValidateVerbs(fl validator.FieldLevel) bool {
	resource := fl.Parent().Interface().(Resource)

	for _, verb := range resource.Verbs {
		if resource.Type == ListResource && (verb > 5) {
			return false
		}

		if resource.Type == SimpleResource && (verb == ListVerb || verb > LastVerb) {
			return false
		}
	}

	return true
}
