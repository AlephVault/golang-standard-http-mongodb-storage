package app

import (
	"errors"
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

// Run runs the actual web server.
func (application *Application) Run(addr ...string) error {
	if application.router != nil {
		return errors.New("the router is null")
	}
	return application.router.Run(addr...)
}
