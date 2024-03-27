package impl

import (
	"github.com/AlephVault/golang-standard-http-mongodb-storage/core/responses"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
	"log/slog"
)

// GetDocument gets the only document from a FindOne result.
func GetDocument[T any](
	context echo.Context, findOneResult *mongo.SingleResult, element *T,
	logger ...*slog.Logger,
) error {
	if err := findOneResult.Decode(element); err != nil {
		return responses.FindOneOperationError(context, err, logger...)
	} else {
		return nil
	}
}

// GetDocuments gets the documents from a Find result.
func GetDocuments[T any](
	context echo.Context, cursor *mongo.Cursor, elements *[]T,
	logger ...*slog.Logger,
) error {
	newElements := []T{}
	ctx := context.Request().Context()
	log := func(error) {}
	if len(logger) > 0 {
		firstLogger := logger[0]
		log = func(err error) {
			firstLogger.Error("An error occurred: " + err.Error())
		}
	}

	for cursor.Next(ctx) {
		var t T
		if err := cursor.Decode(&t); err != nil {
			log(err)
			return responses.InternalError(context)
		}
		newElements = append(newElements, t)
	}

	*elements = newElements
	return nil
}
