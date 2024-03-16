package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"log/slog"
	"standard-http-mongodb-storage/core/dsl"
	"standard-http-mongodb-storage/core/responses"
	"strings"
)

func registerEndpoints(
	client *mongo.Client, router *gin.Engine, key string,
	resource *dsl.Resource, auth *dsl.Auth,
) {
	if resource.Type == dsl.SimpleResource {
		registerSimpleResourceEndpoints(client, router, key, resource, auth)
	} else {
		registerListResourceEndpoints(client, router, key, resource, auth)
	}
}

func registerSimpleResourceEndpoints(
	client *mongo.Client, router *gin.Engine, key string,
	resource *dsl.Resource, auth *dsl.Auth,
) {
	// tmpUpdatesCollection := client.Database("~tmp").Collection("updates")
	authCollection := client.Database(auth.Db).Collection(auth.Collection)

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

				// TODO simple resource CREATE, 409, or some error.
			})
			break
		case dsl.ReadVerb:
			router.GET("/"+key, func(context *gin.Context) {
				if !authenticate(context, authCollection, key, "read") {
					return
				}

				// TODO implement GET, 404, or error.
			})
			break
		case dsl.UpdateVerb:
			router.PATCH("/"+key, func(context *gin.Context) {
				if !authenticate(context, authCollection, key, "write") {
					return
				}

				// TODO simple resource PATCH, 404, or some error.
			})
			break
		case dsl.ReplaceVerb:
			router.PUT("/"+key, func(context *gin.Context) {
				if !authenticate(context, authCollection, key, "write") {
					return
				}

				// TODO simple resource UPSERT, or some error.
			})
			break
		case dsl.DeleteVerb:
			router.DELETE("/"+key, func(context *gin.Context) {
				if !authenticate(context, authCollection, key, "delete") {
					return
				}

				// TODO simple resource PATH
			})
			break
		default:
			slog.Info("Ignoring an unknown verb", "verb", verb)
		}
	}

	router.GET("/"+key+"/:method", func(context *gin.Context) {
		if !authenticate(context, authCollection, key, "read") {
			return
		}

		// TODO simple resource GET + run a VIEW method.
	})
	router.POST("/"+key+"/:method", func(context *gin.Context) {
		if !authenticate(context, authCollection, key, "write") {
			return
		}

		// TODO simple resource GET + run an OPERATION method.
	})
}

func registerListResourceEndpoints(
	client *mongo.Client, router *gin.Engine, key string,
	resource *dsl.Resource, auth *dsl.Auth,
) {
	// tmpUpdatesCollection := client.Database("~tmp").Collection("updates")
	authCollection := client.Database(auth.Db).Collection(auth.Collection)
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
		case dsl.ListVerb:
			router.GET("/"+key, func(context *gin.Context) {
				if !authenticate(context, authCollection, key, "read") {
					return
				}

				// TODO multiple resource GET, or some error.
			})
		case dsl.CreateVerb:
			router.POST("/"+key, func(context *gin.Context) {
				if !authenticate(context, authCollection, key, "write") {
					return
				}

				// TODO multiple resource CREATE, or some error.
			})
		case dsl.ReadVerb:
			itemReadDefined = true
			router.GET("/"+key+"/:id_or_method", func(context *gin.Context) {
				if !authenticate(context, authCollection, key, "read") {
					return
				}

				if id, ok := checkId(context, "id_or_method", true); ok {
					// TODO implement here.
					return
				}

				method := context.Param("id_or_method")
				if !strings.HasPrefix(method, "~") {
					responses.NotFound(context)
					return
				}
				method = method[1:]
				if listMethod, ok := resource.ListMethods[method]; !ok || listMethod.Handler == nil || listMethod.Type == dsl.Operation {
					responses.NotFound(context)
					return
				} else {
					defer func() {
						if v := recover(); v != nil {
							slog.Error(fmt.Sprintf("Panic! %v", v))
							responses.InternalError(context)
						}
					}()
					listMethod.Handler(context, client, key, method, resource.Db, resource.Collection, resource.Filter)
				}
			})
		case dsl.UpdateVerb:
			router.PATCH("/"+key+"/:id", func(context *gin.Context) {
				if !authenticate(context, authCollection, key, "write") {
					return
				}

				if id, ok := checkId(context, "id", true); !ok {
					return
				} else {
					// TODO PATCH one element, 404, or some error.
				}
			})
		case dsl.ReplaceVerb:
			router.PUT("/"+key+"/:id", func(context *gin.Context) {
				if !authenticate(context, authCollection, key, "write") {
					return
				}

				if id, ok := checkId(context, "id", true); !ok {
					return
				} else {
					// TODO UPSERT one element, 404, or some error.
				}
			})
		case dsl.DeleteVerb:
			router.DELETE("/"+key+"/:id", func(context *gin.Context) {
				if !authenticate(context, authCollection, key, "delete") {
					return
				}

				if id, ok := checkId(context, "id", true); !ok {
					return
				} else {
					// TODO DELETE one element, 404, or some error.
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

			method := context.Param("method")
			if !strings.HasPrefix(method, "~") {
				responses.NotFound(context)
				return
			}
			method = method[1:]
			if listMethod, ok := resource.ListMethods[method]; !ok || listMethod.Handler == nil || listMethod.Type == dsl.Operation {
				responses.NotFound(context)
				return
			} else {
				defer func() {
					if v := recover(); v != nil {
						slog.Error(fmt.Sprintf("Panic! %v", v))
						responses.InternalError(context)
					}
				}()
				listMethod.Handler(context, client, key, method, resource.Db, resource.Collection, resource.Filter)
			}
		})
	}

	router.POST("/"+key+"/:method", func(context *gin.Context) {
		if !authenticate(context, authCollection, key, "write") {
			return
		}

		// TODO run a collection's OPERATION method.
	})
	router.GET("/"+key+"/:id/:method", func(context *gin.Context) {
		if !authenticate(context, authCollection, key, "read") {
			return
		}

		if id, ok := checkId(context, "id", true); !ok {
			return
		} else {
			// TODO GET one element + execute an element's VIEW method.
		}
	})
	router.POST("/"+key+"/:id/:method", func(context *gin.Context) {
		if !authenticate(context, authCollection, key, "write") {
			return
		}

		if id, ok := checkId(context, "id", true); !ok {
			return
		} else {
			// TODO GET one element + execute an element's OPERATION method.
		}
	})
}
