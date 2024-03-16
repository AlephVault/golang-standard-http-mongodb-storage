package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log/slog"
	"standard-http-mongodb-storage/core/dsl"
	"standard-http-mongodb-storage/core/responses"
	"strings"
)

// validate executes the validator and, on errors, dumps a response.
func validate(context *gin.Context, value any, validator_ *validator.Validate) bool {
	if err := validator_.Struct(value); err != nil {
		if errVE, ok := err.(validator.ValidationErrors); ok {
			responses.InvalidFormat(context, errVE)
		} else {
			responses.UnexpectedFormat(context)
		}
		return false
	}
	return true
}

// readJSONBody attempts to read a JSON body from the request and parse the object.
func readJSONBody(context *gin.Context, make_ func() any, validator_ *validator.Validate) (any, bool) {
	// 1. Check that there's a JSON body
	if context.Request.Body == nil || !strings.Contains(strings.ToLower(context.GetHeader("Content-Type")), "application/json") {
		responses.UnexpectedFormat(context)
		return nil, false
	}

	// 2. Convert it to an instance of map[string]any
	myValue := make_()
	if err := context.ShouldBindJSON(&myValue); err != nil {
		responses.UnexpectedFormat(context)
		return nil, false
	}

	// 3. If a validator is specified, validate.
	//    NOTES: This will not be invoked in maps.
	if !validate(context, myValue, validator_) {
		return nil, false
	}

	// Return the parsed body.
	return myValue, true
}

// simpleCreate is the full handler of the POST endpoint for simple resources.
func simpleCreate(
	ctx *gin.Context, createOne CreateOneFunc, make_ func() any, validatorMaker func() *validator.Validate,
	logger *slog.Logger,
) {
	// TODO.
}

// simpleGet is the full handler of the GET endpoint for simple resources.
func simpleGet(
	ctx *gin.Context, getOne GetOneFunc, validatorMaker func() *validator.Validate,
	logger *slog.Logger,
) {
	// TODO.
}

// simpleDelete is the full handler of the DELETE endpoint for simple resources.
func simpleDelete(
	ctx *gin.Context, deleteOne DeleteOneFunc, validatorMaker func() *validator.Validate,
	logger *slog.Logger,
) {
	// TODO.
}

// simpleUpdate is the full handler of the PATCH endpoint for simple resources.
func simpleUpdate(
	ctx *gin.Context, updateOne UpdateOneFunc, makeMap func() any, simulatedUpdate SimulatedUpdateFunc,
	validatorMaker func() *validator.Validate, logger *slog.Logger,
) {
	// TODO.
}

// simpleReplace is the full handler of the PUT endpoint for simple resources.
func simpleReplace(
	ctx *gin.Context, replaceOne ReplaceOneFunc, make_ func() any, validatorMaker func() *validator.Validate,
	logger *slog.Logger,
) {
	// TODO.
}

// listCreate is the full handler of the POST endpoint for list resources.
func listCreate(
	ctx *gin.Context, createOne CreateOneFunc, make_ func() any, validatorMaker func() *validator.Validate,
	logger *slog.Logger,
) {
	// TODO.
}

// listGet is the full handler of the GET endpoint for list resources.
func listGet(
	ctx *gin.Context, getMany GetManyFunc, validatorMaker func() *validator.Validate,
	logger *slog.Logger,
) {
	// TODO.
}

// listItemGet is the full handler of the GET endpoint for list item resources.
func listItemGet(
	ctx *gin.Context, getOne GetOneFunc, id primitive.ObjectID,
	validatorMaker func() *validator.Validate, logger *slog.Logger,
) {
	// TODO.
}

// listItemUpdate is the full handler of the PATCH endpoint for the list item resources.
func listItemUpdate(
	ctx *gin.Context, updateOne UpdateOneFunc, makeMap func() any, id primitive.ObjectID,
	simulatedUpdate SimulatedUpdateFunc, validatorMaker func() *validator.Validate, logger *slog.Logger,
) {
	// TODO.
}

// listItemReplace is the full handler of the PUT endpoint for the list item resources.
func listItemReplace(
	ctx *gin.Context, replaceOne ReplaceOneFunc, make_ func() any, id primitive.ObjectID,
	validatorMaker func() *validator.Validate, logger *slog.Logger,
) {
	// TODO.
}

// listItemDelete is the full  handler of the DELETE endpoint for the list item resources.
func listItemDelete(
	ctx *gin.Context, deleteOne DeleteOneFunc, id primitive.ObjectID,
	validatorMaker func() *validator.Validate, logger *slog.Logger,
) {
	// TODO.
}

// resourceMethod is the full handler of a resource method.
func resourceMethod(
	ctx *gin.Context, collection *mongo.Collection, filter bson.M, resourceKey string, methodType dsl.MethodType,
	method string, methods map[string]dsl.ResourceMethod, client *mongo.Client, validatorMaker func() *validator.Validate,
	logger *slog.Logger,
) {
	if !strings.HasPrefix(method, "~") {
		responses.NotFound(ctx)
		return
	}
	method = method[1:]
	if resourceMethod, ok := methods[method]; !ok || resourceMethod.Handler == nil || resourceMethod.Type != methodType {
		responses.NotFound(ctx)
		return
	} else {
		defer func() {
			if v := recover(); v != nil {
				slog.Error(fmt.Sprintf("Panic! %v", v))
				responses.InternalError(ctx)
			}
		}()
		logger.Debug(
			"Invoking custom method: (type=%v name=%s) on resource: %s", methodType, method, resourceKey,
		)
		resourceMethod.Handler(ctx, client, resourceKey, method, collection, validatorMaker, filter)
	}

}

// itemMethod is the full handler of a resource method.
func itemMethod(
	ctx *gin.Context, collection *mongo.Collection, filter bson.M, resourceKey string, methodType dsl.MethodType,
	id primitive.ObjectID, method string, methods map[string]dsl.ItemMethod, client *mongo.Client,
	validatorMaker func() *validator.Validate, logger *slog.Logger,
) {
	if !strings.HasPrefix(method, "~") {
		responses.NotFound(ctx)
		return
	}
	method = method[1:]
	if itemMethod, ok := methods[method]; !ok || itemMethod.Handler == nil || itemMethod.Type != methodType {
		responses.NotFound(ctx)
		return
	} else {
		defer func() {
			if v := recover(); v != nil {
				slog.Error(fmt.Sprintf("Panic! %v", v))
				responses.InternalError(ctx)
			}
		}()
		logger.Debug(
			"Invoking custom item method: (type=%v name=%s) on resource: %s", methodType, method, resourceKey,
		)
		itemMethod.Handler(ctx, client, resourceKey, method, collection, validatorMaker, filter, id)
	}
}
