package app

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"standard-http-mongodb-storage/core/dsl"
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

	router.GET("/"+key, func(context *gin.Context) {
		if !authenticate(context, authCollection, key, "read") {
			return
		}

		// TODO implement GET, 404, or error.
	})
	router.GET("/"+key+"/:method", func(context *gin.Context) {
		if !authenticate(context, authCollection, key, "read") {
			return
		}

		// TODO simple resource GET + run a VIEW method.
	})
	router.POST("/"+key, func(context *gin.Context) {
		if !authenticate(context, authCollection, key, "write") {
			return
		}

		// TODO simple resource CREATE, 409, or some error.
	})
	router.POST("/"+key+"/:method", func(context *gin.Context) {
		if !authenticate(context, authCollection, key, "write") {
			return
		}

		// TODO simple resource GET + run an OPERATION method.
	})
	router.PATCH("/"+key, func(context *gin.Context) {
		if !authenticate(context, authCollection, key, "write") {
			return
		}

		// TODO simple resource PATCH, 404, or some error.
	})
	router.PUT("/"+key, func(context *gin.Context) {
		if !authenticate(context, authCollection, key, "write") {
			return
		}

		// TODO simple resource UPSERT, or some error.
	})
	router.DELETE("/"+key, func(context *gin.Context) {
		if !authenticate(context, authCollection, key, "delete") {
			return
		}

		// TODO simple resource PATH
	})
}

func registerListResourceEndpoints(
	client *mongo.Client, router *gin.Engine, key string,
	resource *dsl.Resource, auth *dsl.Auth,
) {
	// tmpUpdatesCollection := client.Database("~tmp").Collection("updates")
	authCollection := client.Database(auth.Db).Collection(auth.Collection)

	router.GET("/"+key, func(context *gin.Context) {
		if !authenticate(context, authCollection, key, "read") {
			return
		}

		// TODO multiple resource GET, or some error.
	})
	router.POST("/"+key, func(context *gin.Context) {
		if !authenticate(context, authCollection, key, "write") {
			return
		}

		// TODO multiple resource CREATE, or some error.
	})
	router.GET("/"+key+"/:id", func(context *gin.Context) {
		if !authenticate(context, authCollection, key, "read") {
			return
		}

		// TODO GET one element, 404, or some error.
		// TODO alternatively, run a collection's VIEW method.
	})
	router.POST("/"+key+"/:id", func(context *gin.Context) {
		if !authenticate(context, authCollection, key, "write") {
			return
		}

		// TODO run a collection's OPERATION method.
	})
	router.PATCH("/"+key+"/:id", func(context *gin.Context) {
		if !authenticate(context, authCollection, key, "write") {
			return
		}

		// TODO PATCH one element, 404, or some error.
	})
	router.PUT("/"+key+"/:id", func(context *gin.Context) {
		if !authenticate(context, authCollection, key, "write") {
			return
		}

		// TODO UPSERT one element, 404, or some error.
	})
	router.DELETE("/"+key+"/:id", func(context *gin.Context) {
		if !authenticate(context, authCollection, key, "delete") {
			return
		}

		// TODO DELETE one element, 404, or some error.
	})
	router.GET("/"+key+"/:id/:method", func(context *gin.Context) {
		if !authenticate(context, authCollection, key, "read") {
			return
		}

		// TODO GET one element + execute an element's VIEW method.
	})
	router.POST("/"+key+"/:id/:method", func(context *gin.Context) {
		if !authenticate(context, authCollection, key, "write") {
			return
		}

		// TODO GET one element + execute an element's OPERATION method.
	})
}
