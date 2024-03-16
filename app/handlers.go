package app

import (
	"github.com/gin-gonic/gin"
	"standard-http-mongodb-storage/core/dsl"
)

// simpleCreate is the full handler of the POST endpoint for simple resources.
func simpleCreate(ctx *gin.Context, createOne CreateOneFunc) {
	// TODO.
}

// simpleGet is the full handler of the GET endpoint for simple resources.
func simpleGet(ctx *gin.Context, getOne GetOneFunc) {
	// TODO.
}

// simpleDelete is the full handler of the DELETE endpoint for simple resources.
func simpleDelete(ctx *gin.Context, deleteOne DeleteOneFunc) {
	// TODO.
}

// simpleUpdate is the full handler of the PATCH endpoint for simple resources.
func simpleUpdate(ctx *gin.Context, updateOne UpdateOneFunc, simulatedUpdate SimulatedUpdateFunc) {
	// TODO.
}

// simpleReplace is the full handler of the PUT endpoint for simple resources.
func simpleReplace(ctx *gin.Context, replaceOne ReplaceOneFunc) {
	// TODO.
}

// simpleMethod is the full handler of a method.
func simpleMethod(ctx *gin.Context, methodType dsl.MethodType, methods map[string]dsl.ResourceMethod) {
	// TODO.
}
