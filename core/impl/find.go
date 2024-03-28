package impl

import (
	"github.com/AlephVault/golang-standard-http-mongodb-storage/core/responses"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
	"log/slog"
)

// GetDocument gets the only document from a FindOne result.
func GetDocument(
	context echo.Context, findOneResult *mongo.SingleResult, element any,
	logger ...*slog.Logger,
) (bool, error) {
	if err := findOneResult.Decode(element); err != nil {
		return false, responses.FindOneOperationError(context, err, logger...)
	} else {
		return true, nil
	}
}

// GetDocuments gets the documents from a Find result.
func GetDocuments[T any](
	context echo.Context, cursor *mongo.Cursor, elements *[]T,
	logger ...*slog.Logger,
) (bool, error) {
	newElements := []T{}
	ctx := context.Request().Context()
	log := func(err error) { slog.Error("An error occurred: " + err.Error()) }
	if len(logger) > 0 {
		firstLogger := logger[0]
		log = func(err error) {
			firstLogger.Error("An error occurred: " + err.Error())
		}
	}

	defer func(cursor *mongo.Cursor) {
		_ = cursor.Close(ctx)
	}(cursor)
	for cursor.Next(ctx) {
		var t T
		if err := cursor.Decode(&t); err != nil {
			log(err)
			return false, responses.InternalError(context)
		}
		newElements = append(newElements, t)
	}

	*elements = newElements
	return true, nil
}
