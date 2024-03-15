package app

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"standard-http-mongodb-storage/core/dsl"
)

func registerEndpoints(client *mongo.Client, router *gin.Engine, key string, resource *dsl.Resource) {
	if resource.Type == dsl.SimpleResource {
		registerSimpleResourceEndpoints(client, router, key, resource)
	} else {
		registerListResourceEndpoints(client, router, key, resource)
	}
}

func registerSimpleResourceEndpoints(client *mongo.Client, router *gin.Engine, key string, resource *dsl.Resource) {
	// tmpUpdatesCollection := client.Database("~tmp").Collection("updates")

	router.GET("/"+key, func(context *gin.Context) {
		// TODO simple resource GET, 404, or some error.
	})
	router.GET("/"+key+"/:method", func(context *gin.Context) {
		// TODO simple resource GET + run a VIEW method.
	})
	router.POST("/"+key, func(context *gin.Context) {
		// TODO simple resource CREATE, 409, or some error.
	})
	router.POST("/"+key+"/:method", func(context *gin.Context) {
		// TODO simple resource GET + run an OPERATION method.
	})
	router.PATCH("/"+key, func(context *gin.Context) {
		// TODO simple resource PATCH, 404, or some error.
	})
	router.PUT("/"+key, func(context *gin.Context) {
		// TODO simple resource UPSERT, or some error.
	})
	router.DELETE("/"+key, func(context *gin.Context) {
		// TODO simple resource PATH
	})
}

func registerListResourceEndpoints(client *mongo.Client, router *gin.Engine, key string, resource *dsl.Resource) {
	// tmpUpdatesCollection := client.Database("~tmp").Collection("updates")

	router.GET("/"+key, func(context *gin.Context) {
		// TODO multiple resource GET, or some error.
	})
	router.POST("/"+key, func(context *gin.Context) {
		// TODO multiple resource CREATE, or some error.
	})
	router.GET("/"+key+"/:id", func(context *gin.Context) {
		// TODO GET one element, 404, or some error.
		// TODO alternatively, run a collection's VIEW method.
	})
	router.POST("/"+key+"/:id", func(context *gin.Context) {
		// TODO run a collection's OPERATION method.
	})
	router.PATCH("/"+key+"/:id", func(context *gin.Context) {
		// TODO PATCH one element, 404, or some error.
	})
	router.PUT("/"+key+"/:id", func(context *gin.Context) {
		// TODO UPSERT one element, 404, or some error.
	})
	router.DELETE("/"+key+"/:id", func(context *gin.Context) {
		// TODO DELETE one element, 404, or some error.
	})
	router.GET("/"+key+"/:id/:method", func(context *gin.Context) {
		// TODO GET one element + execute an element's VIEW method.
	})
	router.POST("/"+key+"/:id/:method", func(context *gin.Context) {
		// TODO GET one element + execute an element's OPERATION method.
	})
}
