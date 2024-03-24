package app

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log/slog"
	"reflect"
	"standard-http-mongodb-storage/core/dsl"
	"standard-http-mongodb-storage/core/responses"
)

func registerEndpoints(
	client *mongo.Client, router *echo.Echo, key string,
	resource *dsl.Resource, auth *dsl.Auth, resourcesValidatorMaker func() *validator.Validate,
	listMaxResults int64, logger *slog.Logger,
) {
	if resource.Type == dsl.SimpleResource {
		registerSimpleResourceEndpoints(client, router, key, resource, auth, resourcesValidatorMaker, logger)
	} else {
		registerListResourceEndpoints(
			client, router, key, resource, auth, resourcesValidatorMaker, listMaxResults, logger,
		)
	}
}

func registerSimpleResourceEndpoints(
	client *mongo.Client, router *echo.Echo, key string,
	resource *dsl.Resource, auth *dsl.Auth, validatorMaker func() *validator.Validate,
	logger *slog.Logger,
) {
	tmpUpdatesCollection := client.Database("~tmp").Collection("updates")
	authCollection := client.Database(auth.Db).Collection(auth.Collection)
	collection := client.Database(resource.Db).Collection(resource.Collection)
	filter := resource.Filter
	softDelete := resource.SoftDelete
	sort := resource.Sort
	projection := resource.Projection
	methods := resource.Methods
	idGetter := makeIDGetter(resource.ModelType())

	modelType_ := reflect.TypeOf(resource.ModelType())
	make_ := func() any { return reflect.New(modelType_).Interface() }
	makeMap := func() any { return make(bson.M) }

	createOne := makeCreateOne(collection)
	getOne := makeGetOne(collection, make_, softDelete, filter, projection, sort)
	replaceOne := makeReplaceOne(collection, filter, softDelete)
	deleteOne := makeDeleteOne(collection, filter, softDelete)
	simulatedUpdate := makeSimulatedUpdate(tmpUpdatesCollection, make_)

	verbs := resource.Verbs
	if len(verbs) == 0 {
		verbs = []dsl.ResourceVerb{
			dsl.CreateVerb, dsl.ReadVerb, dsl.UpdateVerb, dsl.ReplaceVerb, dsl.DeleteVerb,
		}
	}

	for _, verb := range verbs {
		switch verb {
		case dsl.CreateVerb:
			router.POST("/"+key, func(context echo.Context) error {
				if success, err := authenticate(context, authCollection, key, "write"); !success {
					return err
				}
				return simpleCreate(context, createOne, getOne, make_, validatorMaker, logger)
			})
		case dsl.ReadVerb:
			router.GET("/"+key, func(context echo.Context) error {
				if success, err := authenticate(context, authCollection, key, "read"); !success {
					return err
				}
				return simpleGet(context, getOne, logger)
			})
		case dsl.UpdateVerb:
			router.PATCH("/"+key, func(context echo.Context) error {
				if success, err := authenticate(context, authCollection, key, "write"); !success {
					return err
				}
				return simpleUpdate(context, getOne, idGetter, replaceOne, makeMap, simulatedUpdate, validatorMaker, logger)
			})
		case dsl.ReplaceVerb:
			router.PUT("/"+key, func(context echo.Context) error {
				if success, err := authenticate(context, authCollection, key, "write"); !success {
					return err
				}
				return simpleReplace(context, replaceOne, make_, validatorMaker, logger)
			})
		case dsl.DeleteVerb:
			router.DELETE("/"+key, func(context echo.Context) error {
				if success, err := authenticate(context, authCollection, key, "delete"); !success {
					return err
				}
				return simpleDelete(context, deleteOne, logger)
			})
		default:
			slog.Info("Ignoring an unknown verb", "verb", verb)
		}
	}

	router.GET("/"+key+"/:method", func(context echo.Context) error {
		if success, err := authenticate(context, authCollection, key, "read"); !success {
			return err
		}
		return resourceMethod(
			context, collection, filter, key, dsl.View, context.Param("method"), methods, client,
			validatorMaker, logger,
		)
	})
	router.POST("/"+key+"/:method", func(context echo.Context) error {
		if success, err := authenticate(context, authCollection, key, "write"); !success {
			return err
		}
		return resourceMethod(
			context, collection, filter, key, dsl.Operation, context.Param("method"), methods, client,
			validatorMaker, logger,
		)
	})
}

func registerListResourceEndpoints(
	client *mongo.Client, router *echo.Echo, key string,
	resource *dsl.Resource, auth *dsl.Auth, validatorMaker func() *validator.Validate,
	listMaxResults int64, logger *slog.Logger,
) {
	tmpUpdatesCollection := client.Database("~tmp").Collection("updates")
	authCollection := client.Database(auth.Db).Collection(auth.Collection)
	collection := client.Database(resource.Db).Collection(resource.Collection)
	filter := resource.Filter
	softDelete := resource.SoftDelete
	sort := resource.Sort
	projection := resource.Projection
	itemProjection := resource.ItemProjection
	methods := resource.Methods
	itemMethods := resource.ItemMethods

	modelType_ := reflect.TypeOf(resource.ModelType())
	make_ := func() any { return reflect.New(modelType_).Interface() }
	makeMap := func() any { return make(bson.M) }

	createOne := makeCreateOne(collection)
	getMany := makeGetMany(collection, make_, softDelete, filter, projection, sort)
	getOne := makeGetOne(collection, make_, softDelete, filter, itemProjection, sort)
	replaceOne := makeReplaceOne(collection, filter, softDelete)
	deleteOne := makeDeleteOne(collection, filter, softDelete)
	simulatedUpdate := makeSimulatedUpdate(tmpUpdatesCollection, make_)
	itemReadDefined := false

	verbs := resource.Verbs
	if len(verbs) == 0 {
		verbs = []dsl.ResourceVerb{
			dsl.ListVerb, dsl.CreateVerb, dsl.ReadVerb,
			dsl.UpdateVerb, dsl.ReplaceVerb, dsl.DeleteVerb,
		}
	}

	for _, verb := range verbs {
		switch verb {
		case dsl.CreateVerb:
			router.GET("/"+key, func(context echo.Context) error {
				if success, err := authenticate(context, authCollection, key, "read"); !success {
					return err
				}
				return listCreate(context, createOne, make_, validatorMaker, logger)
			})
		case dsl.ListVerb:
			router.POST("/"+key, func(context echo.Context) error {
				if success, err := authenticate(context, authCollection, key, "write"); !success {
					return err
				}
				return listGet(context, getMany, listMaxResults, logger)
			})
		case dsl.ReadVerb:
			itemReadDefined = true
			router.GET("/"+key+"/:id_or_method", func(context echo.Context) error {
				if success, err := authenticate(context, authCollection, key, "read"); !success {
					return err
				}
				if id, ok := checkId(context, "id_or_method", true); ok {
					return listItemGet(context, getOne, id, logger)
				} else {
					return resourceMethod(
						context, collection, filter, key, dsl.Operation, context.Param("method"), methods, client,
						validatorMaker, logger,
					)
				}
			})
		case dsl.UpdateVerb:
			router.PATCH("/"+key+"/:id", func(context echo.Context) error {
				if success, err := authenticate(context, authCollection, key, "write"); !success {
					return err
				}
				if id, ok := checkId(context, "id", true); !ok {
					return responses.NotFound(context)
				} else {
					return listItemUpdate(context, getOne, replaceOne, makeMap, id, simulatedUpdate, validatorMaker, logger)
				}
			})
		case dsl.ReplaceVerb:
			router.PUT("/"+key+"/:id", func(context echo.Context) error {
				if success, err := authenticate(context, authCollection, key, "write"); !success {
					return err
				}
				if id, ok := checkId(context, "id", true); !ok {
					return responses.NotFound(context)
				} else {
					return listItemReplace(context, replaceOne, make_, id, validatorMaker, logger)
				}
			})
		case dsl.DeleteVerb:
			router.DELETE("/"+key+"/:id", func(context echo.Context) error {
				if success, err := authenticate(context, authCollection, key, "delete"); !success {
					return err
				}
				if id, ok := checkId(context, "id", true); !ok {
					return responses.NotFound(context)
				} else {
					return listItemDelete(context, deleteOne, id, logger)
				}
			})
		default:
			slog.Info("Ignoring an unknown verb", "verb", verb)
		}
	}

	if !itemReadDefined {
		router.GET("/"+key+"/:method", func(context echo.Context) error {
			if success, err := authenticate(context, authCollection, key, "read"); !success {
				return err
			}
			return resourceMethod(
				context, collection, filter, key, dsl.Operation, context.Param("method"), methods, client,
				validatorMaker, logger,
			)
		})
	}

	router.POST("/"+key+"/:method", func(context echo.Context) error {
		if success, err := authenticate(context, authCollection, key, "write"); !success {
			return err
		}
		return resourceMethod(
			context, collection, filter, key, dsl.Operation, context.Param("method"), methods, client,
			validatorMaker, logger,
		)
	})
	router.GET("/"+key+"/:id/:method", func(context echo.Context) error {
		if success, err := authenticate(context, authCollection, key, "read"); !success {
			return err
		}
		if id, ok := checkId(context, "id", true); !ok {
			return responses.NotFound(context)
		} else {
			return itemMethod(
				context, collection, filter, key, dsl.View, id, context.Param("method"), itemMethods, client,
				validatorMaker, logger,
			)
		}
	})
	router.POST("/"+key+"/:id/:method", func(context echo.Context) error {
		if success, err := authenticate(context, authCollection, key, "write"); !success {
			return err
		}
		if id, ok := checkId(context, "id", true); !ok {
			return responses.NotFound(context)
		} else {
			return itemMethod(
				context, collection, filter, key, dsl.Operation, id, context.Param("method"), itemMethods, client,
				validatorMaker, logger,
			)
		}
	})
}
