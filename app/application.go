package app

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// Application is a wrapper defining the router, the
// client layer, and other elements needed to interact
// with the storage and serve it.
type Application struct {
	client *mongo.Client
	router *gin.Engine
}

// TODO implement this later.
