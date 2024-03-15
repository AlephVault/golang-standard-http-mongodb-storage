package app

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"maps"
	"reflect"
)

// CreateOneFunc stands for a function that creates one element.
type CreateOneFunc func(context.Context, interface{}) (primitive.ObjectID, error)

// setId sets the id in a filter, if any. It also sets a filter
// on the _deleted field if softDelete is true.
func setId(filter bson.M, id string, softDelete bool) (bson.M, error) {
	// Set the ID.
	filter_ := bson.M{}
	maps.Copy(filter_, filter)
	if len(id) == 0 {
		return filter_, nil
	} else if id_, err := primitive.ObjectIDFromHex(id); err != nil {
		return nil, err
	} else {
		filter_["_id"] = id_
		if softDelete {
			filter_["_deleted"] = bson.M{"$ne": true}
		}
		return filter_, nil
	}
}

// makeCreateOne creates a document.
func makeCreateOne(
	client *mongo.Client, resDB, resCollection string,
) func(context.Context, interface{}) (primitive.ObjectID, error) {
	return func(ctx context.Context, content interface{}) (primitive.ObjectID, error) {
		if result, err := client.Database(resDB).Collection(resCollection).InsertOne(ctx, content); err != nil {
			return primitive.ObjectID{}, err
		} else {
			return result.InsertedID.(primitive.ObjectID), nil
		}
	}
}

// makeGetOne returns a single element. Returns a POINTER to a new element.
func makeGetOne(
	client *mongo.Client, template interface{}, resDB, resCollection string, softDelete bool,
	filter bson.M, projection interface{}, sort interface{},
) func(context.Context, string) (interface{}, error) {
	// The template WILL be of a struct type.
	// NOT a null value. NOT a pointer to a struct.

	return func(ctx context.Context, id string) (interface{}, error) {
		var err error
		var filter_ bson.M

		// Set the ID.
		if filter_, err = setId(filter, id, softDelete); err != nil {
			return nil, err
		}

		// Try getting an element.
		result := client.Database(resDB).Collection(resCollection).FindOne(
			ctx, filter_,
			options.FindOne().SetProjection(projection).SetReturnKey(true).SetShowRecordID(true).SetSort(sort),
		)

		// Decode the result.
		obj := reflect.New(reflect.TypeOf(template))
		if err := result.Decode(&obj); err != nil {
			return nil, err
		} else {
			return &obj, nil
		}
	}
}

// makeDeleteOne deletes a single element.
func makeDeleteOne(
	client *mongo.Client, resDB, resCollection string, filter bson.M, softDelete bool,
) func(context.Context, string) (bool, error) {
	return func(ctx context.Context, id string) (bool, error) {
		var err error
		var filter_ bson.M

		// Set the ID.
		if filter_, err = setId(filter, id, softDelete); err != nil {
			return false, err
		}

		// Try deleting an element.
		if result, err := client.Database(resDB).Collection(resCollection).DeleteOne(
			ctx, filter_,
		); err != nil {
			return false, err
		} else {
			return result.DeletedCount > 0, nil
		}
	}
}

// makeUpdateOne patches a document.
func makeUpdateOne(
	client *mongo.Client, resDB, resCollection string, filter bson.M, softDelete bool,
) func(context.Context, string, bson.M) (bool, error) {
	return func(ctx context.Context, id string, updates bson.M) (bool, error) {
		var err error
		var filter_ bson.M

		// Set the ID.
		if filter_, err = setId(filter_, id, softDelete); err != nil {
			return false, err
		}

		// Try updating an element.
		if result, err := client.Database(resDB).Collection(resCollection).UpdateOne(
			ctx, filter_, bson.M{"$set": updates},
		); err != nil {
			return false, err
		} else {
			return result.ModifiedCount > 0, nil
		}
	}
}

// makeReplaceOne replaces a document.
func makeReplaceOne(
	client *mongo.Client, resDB, resCollection string, filter bson.M, softDelete bool,
) func(context.Context, string, interface{}) (bool, error) {
	return func(ctx context.Context, id string, replacement interface{}) (bool, error) {
		var err error
		var filter_ bson.M

		// Set the ID.
		if filter_, err = setId(filter, id, softDelete); err != nil {
			return false, err
		}

		// Try replacing an element.
		if result, err := client.Database(resDB).Collection(resCollection).ReplaceOne(
			ctx, filter_, replacement,
		); err != nil {
			return false, err
		} else {
			return result.ModifiedCount > 0, nil
		}
	}
}
