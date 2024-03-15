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

// GetOneFunc stands for a function that gets one element.
type GetOneFunc func(context.Context, string) (interface{}, error)

// DeleteOneFunc stands for a function that deletes an element.
type DeleteOneFunc func(context.Context, string) (bool, error)

// UpdateOneFunc stands for a function that updates a document.
type UpdateOneFunc func(context.Context, string, bson.M) (bool, error)

// ReplaceOneFunc stands for a function that replaces a document.
type ReplaceOneFunc func(context.Context, string, interface{}) (bool, error)

// GetManyFunc stands for a function that gets many documents.
type GetManyFunc func(context.Context, int64, int64) ([]interface{}, error)

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
) CreateOneFunc {
	return func(ctx context.Context, content interface{}) (primitive.ObjectID, error) {
		if result, err := client.Database(resDB).Collection(resCollection).InsertOne(ctx, content); err != nil {
			return primitive.ObjectID{}, err
		} else {
			return result.InsertedID.(primitive.ObjectID), nil
		}
	}
}

// makeGetMany
func makeGetMany(
	collection *mongo.Collection, template interface{}, softDelete bool,
	filter bson.M, projection interface{}, sort interface{},
) GetManyFunc {
	// The template WILL be of a struct type.
	// NOT a null value. NOT a pointer to a struct.
	return func(ctx context.Context, page int64, pageSize int64) ([]interface{}, error) {
		var err error
		var filter_ bson.M

		// Set the ID.
		if filter_, err = setId(filter, "", softDelete); err != nil {
			return nil, err
		}

		// Try getting many elements.
		options_ := options.Find().SetProjection(projection).SetReturnKey(true).SetShowRecordID(true).SetSort(sort)
		if pageSize > 0 {
			options_ = options_.SetLimit(pageSize)
			if page > 0 {
				options_ = options_.SetSkip(page * pageSize)
			}
		}
		if cursor, err := collection.Find(
			ctx, filter_, options_,
		); err != nil {
			return nil, err
		} else {
			defer cursor.Close(ctx)
			type_ := reflect.TypeOf(template)
			var elements []interface{}
			for cursor.Next(ctx) {
				element := reflect.New(type_)
				if err := cursor.Decode(&element); err != nil {
					return nil, err
				}
				elements = append(elements, cursor)
			}
			return elements, nil
		}
	}
}

// makeGetOne makes a function that returns a single element. Returns a new element.
func makeGetOne(
	collection *mongo.Collection, template interface{}, softDelete bool,
	filter bson.M, projection interface{}, sort interface{},
) GetOneFunc {
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
		result := collection.FindOne(
			ctx, filter_,
			options.FindOne().SetProjection(projection).SetReturnKey(true).SetShowRecordID(true).SetSort(sort),
		)
		err = result.Err()
		if err != nil {
			return nil, err
		}

		// Decode the result.
		obj := reflect.New(reflect.TypeOf(template))
		if err := result.Decode(&obj); err != nil {
			return nil, err
		} else {
			return obj, nil
		}
	}
}

// makeDeleteOne makes a function that deletes a single element.
func makeDeleteOne(
	collection *mongo.Collection, filter bson.M, softDelete bool,
) DeleteOneFunc {
	return func(ctx context.Context, id string) (bool, error) {
		var err error
		var filter_ bson.M

		// Set the ID.
		if filter_, err = setId(filter, id, softDelete); err != nil {
			return false, err
		}

		// Try deleting an element.
		if result, err := collection.DeleteOne(
			ctx, filter_,
		); err != nil {
			return false, err
		} else {
			return result.DeletedCount > 0, nil
		}
	}
}

// makeUpdateOne makes a function that patches a document.
func makeUpdateOne(
	collection *mongo.Collection, filter bson.M, softDelete bool,
) UpdateOneFunc {
	return func(ctx context.Context, id string, updates bson.M) (bool, error) {
		var err error
		var filter_ bson.M

		// Set the ID.
		if filter_, err = setId(filter, id, softDelete); err != nil {
			return false, err
		}

		// Try updating an element.
		if result, err := collection.UpdateOne(
			ctx, filter_, bson.M{"$set": updates},
		); err != nil {
			return false, err
		} else {
			return result.ModifiedCount > 0, nil
		}
	}
}

// makeReplaceOne makes a function that replaces a document.
func makeReplaceOne(
	collection *mongo.Collection, filter bson.M, softDelete bool,
) ReplaceOneFunc {
	return func(ctx context.Context, id string, replacement interface{}) (bool, error) {
		var err error
		var filter_ bson.M

		// Set the ID.
		if filter_, err = setId(filter, id, softDelete); err != nil {
			return false, err
		}

		// Try replacing an element.
		if result, err := collection.ReplaceOne(
			ctx, filter_, replacement,
		); err != nil {
			return false, err
		} else {
			return result.ModifiedCount > 0, nil
		}
	}
}
