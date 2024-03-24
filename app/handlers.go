package app

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log/slog"
	"standard-http-mongodb-storage/core/dsl"
	"standard-http-mongodb-storage/core/responses"
	"strings"
)

// validate executes the validator and, on errors, dumps a response.
func validate(context echo.Context, value any, validator_ *validator.Validate) (bool, error) {
	if err := validator_.Struct(value); err != nil {
		var errVE validator.ValidationErrors
		if ok := errors.As(err, &errVE); ok {
			return false, responses.InvalidFormat(context, errVE)
		} else {
			return false, responses.UnexpectedFormat(context)
		}
	}
	return true, nil
}

// readJSONBody attempts to read a JSON body from the request and parse the object.
func readJSONBody(context echo.Context, make_ func() any, validator_ *validator.Validate) (any, bool, error) {
	// 1. Check that there's a JSON body
	if context.Request().Body == nil || !strings.Contains(strings.ToLower(context.Request().Header.Get("Content-Type")), "application/json") {
		return nil, false, responses.UnexpectedFormat(context)
	}

	// 2. Convert it to an instance of map[string]any or the given struct type.
	myValue := make_()
	if err := (&echo.DefaultBinder{}).BindBody(context, myValue); err != nil {
		return nil, false, responses.UnexpectedFormat(context)
	}

	// 3. If a validator is specified, validate.
	//    NOTES: This will not be invoked in maps.
	if validator_ != nil {
		if valid, err := validate(context, myValue, validator_); !valid {
			return nil, false, err
		}
	}

	// Return the parsed body.
	return myValue, true, nil
}

// simpleCreate is the full handler of the POST endpoint for simple resources.
func simpleCreate(
	ctx echo.Context, createOne CreateOneFunc, getOne GetOneFunc, make_ func() any,
	validatorMaker func() *validator.Validate, logger *slog.Logger,
) error {
	if _, err := getOne(ctx, primitive.NilObjectID); err == nil {
		return responses.AlreadyExists(ctx)
	} else if parsed, ok, err := readJSONBody(ctx, make_, validatorMaker()); ok {
		if id, err := createOne(ctx, parsed); err == nil {
			return responses.Created(ctx, id)
		} else if isDuplicateKeyError(err) {
			return responses.DuplicateKey(ctx)
		} else {
			logger.Error("An error occurred: " + err.Error())
			return responses.InternalError(ctx)
		}
	} else {
		return err
	}
}

// simpleGet is the full handler of the GET endpoint for simple resources.
func simpleGet(
	ctx echo.Context, getOne GetOneFunc, logger *slog.Logger,
) error {
	if element, err := getOne(ctx, primitive.NilObjectID); err == nil {
		return responses.OkWith(ctx, element)
	} else if errors.Is(err, mongo.ErrNoDocuments) {
		return responses.NotFound(ctx)
	} else {
		logger.Error("An error occurred: " + err.Error())
		return responses.InternalError(ctx)
	}
}

// simpleDelete is the full handler of the DELETE endpoint for simple resources.
func simpleDelete(
	ctx echo.Context, deleteOne DeleteOneFunc, logger *slog.Logger,
) error {
	if deleted, err := deleteOne(ctx, primitive.NilObjectID); err != nil {
		logger.Error("An error occurred: " + err.Error())
		return responses.InternalError(ctx)
	} else if !deleted {
		return responses.NotFound(ctx)
	} else {
		return responses.Ok(ctx)
	}
}

// simpleUpdate is the full handler of the PATCH endpoint for simple resources.
func simpleUpdate(
	ctx echo.Context, getOne GetOneFunc, idGetter IDGetter, replaceOne ReplaceOneFunc, makeMap func() any,
	simulatedUpdate SimulatedUpdateFunc, validatorMaker func() *validator.Validate, logger *slog.Logger,
) error {
	if element, err := getOne(ctx, primitive.NilObjectID); err == nil {
		if updates, success, err := readJSONBody(ctx, makeMap, nil); success {
			id := idGetter(element)
			if result, err := simulatedUpdate(ctx, id, element, updates); err != nil {
				logger.Error("An error occurred: " + err.Error())
				return responses.InternalError(ctx)
			} else if result, err := validate(ctx, result, validatorMaker()); !result {
				return err
			} else if updated, err := replaceOne(ctx, id, result); err != nil {
				logger.Error("An error occurred: " + err.Error())
				return responses.InternalError(ctx)
			} else if updated {
				return responses.OkWith(ctx, result)
			} else {
				return responses.NotFound(ctx)
			}
		} else {
			return err
		}
	} else if errors.Is(err, mongo.ErrNoDocuments) {
		return responses.NotFound(ctx)
	} else {
		logger.Error("An error occurred: " + err.Error())
		return responses.InternalError(ctx)
	}
}

// simpleReplace is the full handler of the PUT endpoint for simple resources.
func simpleReplace(
	ctx echo.Context, replaceOne ReplaceOneFunc, make_ func() any, validatorMaker func() *validator.Validate,
	logger *slog.Logger,
) error {
	if replacement, ok, err := readJSONBody(ctx, make_, validatorMaker()); ok {
		if ok, err := replaceOne(ctx, primitive.NilObjectID, replacement); err != nil {
			logger.Error("An error occurred: " + err.Error())
			return responses.InternalError(ctx)
		} else if !ok {
			return responses.NotFound(ctx)
		} else {
			return responses.Ok(ctx)
		}
	} else {
		return err
	}
}

// listCreate is the full handler of the POST endpoint for list resources.
func listCreate(
	ctx echo.Context, createOne CreateOneFunc, make_ func() any, validatorMaker func() *validator.Validate,
	logger *slog.Logger,
) error {
	if parsed, ok, err := readJSONBody(ctx, make_, validatorMaker()); ok {
		if id, err := createOne(ctx, parsed); err == nil {
			return responses.Created(ctx, id)
		} else if isDuplicateKeyError(err) {
			return responses.DuplicateKey(ctx)
		} else {
			return err
		}
	} else {
		return err
	}
}

// listGet is the full handler of the GET endpoint for list resources.
func listGet(
	ctx echo.Context, getMany GetManyFunc, defaultLimit int64, logger *slog.Logger,
) error {
	var skip, limit int64 = 0, defaultLimit

	// Get "skip" query parameter
	_ = echo.QueryParamsBinder(ctx).Int64("skip", &skip).Int64("limit", &limit)

	if result, err := getMany(ctx, skip, limit); err != nil {
		logger.Error("An error occurred: " + err.Error())
		return responses.InternalError(ctx)
	} else {
		return responses.OkWith(ctx, result)
	}
}

// listItemGet is the full handler of the GET endpoint for list item resources.
func listItemGet(
	ctx echo.Context, getOne GetOneFunc, id primitive.ObjectID, logger *slog.Logger,
) error {
	if element, err := getOne(ctx, id); err == nil {
		return responses.OkWith(ctx, element)
	} else if errors.Is(err, mongo.ErrNoDocuments) {
		return responses.NotFound(ctx)
	} else {
		logger.Error("An error occurred: " + err.Error())
		return responses.InternalError(ctx)
	}
}

// listItemUpdate is the full handler of the PATCH endpoint for the list item resources.
func listItemUpdate(
	ctx echo.Context, getOne GetOneFunc, replaceOne ReplaceOneFunc, makeMap func() any, id primitive.ObjectID,
	simulatedUpdate SimulatedUpdateFunc, validatorMaker func() *validator.Validate, logger *slog.Logger,
) error {
	if element, err := getOne(ctx, id); err == nil {
		if updates, success, err := readJSONBody(ctx, makeMap, nil); success {
			if result, err := simulatedUpdate(ctx, id, element, updates); err != nil {
				logger.Error("An error occurred: " + err.Error())
				return responses.InternalError(ctx)
			} else if result, err := validate(ctx, result, validatorMaker()); !result {
				return err
			} else if updated, err := replaceOne(ctx, id, result); err != nil {
				logger.Error("An error occurred: " + err.Error())
				return responses.InternalError(ctx)
			} else if updated {
				return responses.OkWith(ctx, result)
			} else {
				return responses.NotFound(ctx)
			}
		} else {
			return err
		}
	} else if errors.Is(err, mongo.ErrNoDocuments) {
		return responses.NotFound(ctx)
	} else {
		logger.Error("An error occurred: " + err.Error())
		return responses.InternalError(ctx)
	}
}

// listItemReplace is the full handler of the PUT endpoint for the list item resources.
func listItemReplace(
	ctx echo.Context, replaceOne ReplaceOneFunc, make_ func() any, id primitive.ObjectID,
	validatorMaker func() *validator.Validate, logger *slog.Logger,
) error {
	if replacement, ok, err := readJSONBody(ctx, make_, validatorMaker()); ok {
		if ok, err := replaceOne(ctx, id, replacement); err != nil {
			logger.Error("An error occurred: " + err.Error())
			return responses.InternalError(ctx)
		} else if !ok {
			return responses.NotFound(ctx)
		} else {
			return responses.Ok(ctx)
		}
	} else {
		return err
	}
}

// listItemDelete is the full  handler of the DELETE endpoint for the list item resources.
func listItemDelete(
	ctx echo.Context, deleteOne DeleteOneFunc, id primitive.ObjectID, logger *slog.Logger,
) error {
	if deleted, err := deleteOne(ctx, id); err != nil {
		logger.Error("An error occurred: " + err.Error())
		return responses.InternalError(ctx)
	} else if !deleted {
		return responses.NotFound(ctx)
	} else {
		return responses.Ok(ctx)
	}
}

// resourceMethod is the full handler of a resource method.
func resourceMethod(
	ctx echo.Context, collection *mongo.Collection, filter bson.M, resourceKey string, methodType dsl.MethodType,
	method string, methods map[string]dsl.ResourceMethod, client *mongo.Client, validatorMaker func() *validator.Validate,
	logger *slog.Logger,
) (err error) {
	if !strings.HasPrefix(method, "~") {
		return responses.NotFound(ctx)
	}
	method = method[1:]
	if resourceMethod, ok := methods[method]; !ok || resourceMethod.Handler == nil || resourceMethod.Type != methodType {
		return responses.NotFound(ctx)
	} else {
		defer func() {
			if v := recover(); v != nil {
				logger.Error("An error occurred (it was panicked): " + err.Error())
				err = responses.InternalError(ctx)
			}
		}()
		logger.Debug(
			"Invoking custom method: (type=%v name=%s) on resource: %s", methodType, method, resourceKey,
		)
		return resourceMethod.Handler(ctx, client, resourceKey, method, collection, validatorMaker, filter)
	}
}

// itemMethod is the full handler of a resource method.
func itemMethod(
	ctx echo.Context, collection *mongo.Collection, filter bson.M, resourceKey string, methodType dsl.MethodType,
	id primitive.ObjectID, method string, methods map[string]dsl.ItemMethod, client *mongo.Client,
	validatorMaker func() *validator.Validate, logger *slog.Logger,
) (err error) {
	if !strings.HasPrefix(method, "~") {
		return responses.NotFound(ctx)
	}
	method = method[1:]
	if itemMethod, ok := methods[method]; !ok || itemMethod.Handler == nil || itemMethod.Type != methodType {
		return responses.NotFound(ctx)
	} else {
		defer func() {
			if v := recover(); v != nil {
				logger.Error("An error occurred (it was panicked): " + err.Error())
				err = responses.InternalError(ctx)
			}
		}()
		logger.Debug(
			"Invoking custom item method: (type=%v name=%s) on resource: %s", methodType, method, resourceKey,
		)
		return itemMethod.Handler(ctx, client, resourceKey, method, collection, validatorMaker, filter, id)
	}
}
