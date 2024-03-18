package app

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log/slog"
	"standard-http-mongodb-storage/core/dsl"
	"standard-http-mongodb-storage/core/responses"
)

func registerEndpoints(
	client *mongo.Client, router *gin.Engine, key string,
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
	client *mongo.Client, router *gin.Engine, key string,
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

	make_ := resource.ModelType
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

	for _, verb := range resource.Verbs {
		switch verb {
		case dsl.CreateVerb:
			router.POST("/"+key, func(context *gin.Context) {
				if !authenticate(context, authCollection, key, "write") {
					return
				}
				simpleCreate(context, createOne, getOne, make_, validatorMaker, logger)
			})
		case dsl.ReadVerb:
			router.GET("/"+key, func(context *gin.Context) {
				if !authenticate(context, authCollection, key, "read") {
					return
				}
				simpleGet(context, getOne, logger)
			})
		case dsl.UpdateVerb:
			router.PATCH("/"+key, func(context *gin.Context) {
				if !authenticate(context, authCollection, key, "write") {
					return
				}
				simpleUpdate(context, getOne, idGetter, replaceOne, makeMap, simulatedUpdate, validatorMaker, logger)
			})
		case dsl.ReplaceVerb:
			router.PUT("/"+key, func(context *gin.Context) {
				if !authenticate(context, authCollection, key, "write") {
					return
				}
				simpleReplace(context, replaceOne, make_, validatorMaker, logger)
			})
		case dsl.DeleteVerb:
			router.DELETE("/"+key, func(context *gin.Context) {
				if !authenticate(context, authCollection, key, "delete") {
					return
				}
				simpleDelete(context, deleteOne, logger)
			})
		default:
			slog.Info("Ignoring an unknown verb", "verb", verb)
		}
	}

	router.GET("/"+key+"/:method", func(context *gin.Context) {
		if !authenticate(context, authCollection, key, "read") {
			return
		}
		resourceMethod(
			context, collection, filter, key, dsl.View, context.Param("method"), methods, client,
			validatorMaker, logger,
		)
	})
	router.POST("/"+key+"/:method", func(context *gin.Context) {
		if !authenticate(context, authCollection, key, "write") {
			return
		}
		resourceMethod(
			context, collection, filter, key, dsl.Operation, context.Param("method"), methods, client,
			validatorMaker, logger,
		)
	})
}

func registerListResourceEndpoints(
	client *mongo.Client, router *gin.Engine, key string,
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

	make_ := resource.ModelType
	makeMap := func() any { return make(bson.M) }

	createOne := makeCreateOne(collection)
	getMany := makeGetMany(collection, make_, softDelete, filter, projection, sort)
	getOne := makeGetOne(collection, make_, softDelete, filter, itemProjection, sort)
	updateOne := makeUpdateOne(collection, filter, softDelete)
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
			router.GET("/"+key, func(context *gin.Context) {
				if !authenticate(context, authCollection, key, "read") {
					return
				}
				listCreate(context, createOne, make_, validatorMaker, logger)
			})
		case dsl.ListVerb:
			router.POST("/"+key, func(context *gin.Context) {
				if !authenticate(context, authCollection, key, "write") {
					return
				}
				listGet(context, getMany, listMaxResults, logger)
			})
		case dsl.ReadVerb:
			itemReadDefined = true
			router.GET("/"+key+"/:id_or_method", func(context *gin.Context) {
				if !authenticate(context, authCollection, key, "read") {
					return
				}
				if id, ok := checkId(context, "id_or_method", true); ok {
					listItemGet(context, getOne, id, validatorMaker, logger)
					return
				} else {
					resourceMethod(
						context, collection, filter, key, dsl.Operation, context.Param("method"), methods, client,
						validatorMaker, logger,
					)
				}
			})
		case dsl.UpdateVerb:
			router.PATCH("/"+key+"/:id", func(context *gin.Context) {
				if !authenticate(context, authCollection, key, "write") {
					return
				}
				if id, ok := checkId(context, "id", true); !ok {
					responses.NotFound(context)
				} else {
					listItemUpdate(context, updateOne, makeMap, id, simulatedUpdate, validatorMaker, logger)
				}
			})
		case dsl.ReplaceVerb:
			router.PUT("/"+key+"/:id", func(context *gin.Context) {
				if !authenticate(context, authCollection, key, "write") {
					return
				}
				if id, ok := checkId(context, "id", true); !ok {
					responses.NotFound(context)
				} else {
					listItemReplace(context, replaceOne, make_, id, validatorMaker, logger)
				}
			})
		case dsl.DeleteVerb:
			router.DELETE("/"+key+"/:id", func(context *gin.Context) {
				if !authenticate(context, authCollection, key, "delete") {
					return
				}
				if id, ok := checkId(context, "id", true); !ok {
					responses.NotFound(context)
				} else {
					listItemDelete(context, deleteOne, id, validatorMaker, logger)
				}
			})
		default:
			slog.Info("Ignoring an unknown verb", "verb", verb)
		}
	}

	if !itemReadDefined {
		router.GET("/"+key+"/:method", func(context *gin.Context) {
			if !authenticate(context, authCollection, key, "read") {
				return
			}
			resourceMethod(
				context, collection, filter, key, dsl.Operation, context.Param("method"), methods, client,
				validatorMaker, logger,
			)
		})
	}

	router.POST("/"+key+"/:method", func(context *gin.Context) {
		if !authenticate(context, authCollection, key, "write") {
			return
		}
		resourceMethod(
			context, collection, filter, key, dsl.Operation, context.Param("method"), methods, client,
			validatorMaker, logger,
		)
	})
	router.GET("/"+key+"/:id/:method", func(context *gin.Context) {
		if !authenticate(context, authCollection, key, "read") {
			return
		}
		if id, ok := checkId(context, "id", true); !ok {
			responses.NotFound(context)
		} else {
			itemMethod(
				context, collection, filter, key, dsl.View, id, context.Param("method"), itemMethods, client,
				validatorMaker, logger,
			)
		}
	})
	router.POST("/"+key+"/:id/:method", func(context *gin.Context) {
		if !authenticate(context, authCollection, key, "write") {
			return
		}
		if id, ok := checkId(context, "id", true); !ok {
			responses.NotFound(context)
		} else {
			itemMethod(
				context, collection, filter, key, dsl.Operation, id, context.Param("method"), itemMethods, client,
				validatorMaker, logger,
			)
		}
	})
}
