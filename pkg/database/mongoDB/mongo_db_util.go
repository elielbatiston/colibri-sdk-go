package mongoDB

import (
	"reflect"

	"github.com/colibriproject-dev/colibri-sdk-go/pkg/base/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoDBModel interface {
	CollectionName() string
	GetID() string
}

// getDataList retrieves a list of items from the given mongo.Client object.
//
// It takes a mongo.Client object as input and returns a list of items of type T and an error.
func getDataList[T any](cursor *mongo.Cursor) ([]T, error) {
	list := make([]T, 0)
	exist := false
	for cursor.Next(mongoDBCtxInstance) {
		exist = true
		var model T
		err := cursor.Decode(&model)
		if err != nil {
			return nil, err
		}

		list = append(list, model)
	}

	if exist {
		return list, nil
	}

	return nil, nil
}

// getMongoCollection returns a instance of mongo.Collection
//
// instance of mongo.Client and a instance of interface
// returns mongo.Collection
func getMongoCollection(instance *mongo.Client, model MongoDBModel) *mongo.Collection {
	collectionName := model.CollectionName()
	return instance.Database(config.MONGODB_NAME).Collection(collectionName)
}

// isStructEmpty return if object is an empty structure
//
// any object
// return bool if object is empty
func isStructEmpty(s interface{}) bool {
	val := reflect.ValueOf(s)

	if isPointer(val) {
		if val.IsNil() {
			return true
		}
		val = val.Elem()
	}

	if isSliceOrArray(val) {
		return val.Len() == 0
	}

	if isStruct(val) {
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			switch field.Kind() {
			case reflect.String:
				if field.String() != "" {
					return false
				}
			case reflect.Map, reflect.Slice:
				if field.Len() != 0 {
					return false
				}
			case reflect.Struct:
				if !isStructEmpty(field.Interface()) {
					return false
				}
			case reflect.Ptr:
				if !field.IsNil() && !isStructEmpty(field.Interface()) {
					return false
				}
			}
		}
	}

	return true
}

// isPointer return if val is a pointer
//
// reflect.Value
// return bool if val is a pointer
func isPointer(val reflect.Value) bool {
	return val.Kind() == reflect.Ptr
}

// isSliceOrArray return if val is a slice or an array
//
// reflect.Value
// return bool if val is a slice or an array
func isSliceOrArray(val reflect.Value) bool {
	return val.Kind() == reflect.Slice || val.Kind() == reflect.Array
}

// isStruct return if val is a struct
//
// reflect.Value
// return bool if val is a struct
func isStruct(val reflect.Value) bool {
	return val.Kind() == reflect.Struct
}

// buildFilterFromModels return a bson.M containig a filter with the Ids of a []interface{}
//
// interface{}
// return a bson.M containig a filter with the Ids of a []interface{} or an error
func buildFilterFromModel(models []MongoDBModel) (bson.M, error) {
	ids := make([]interface{}, len(models))

	for i, model := range models {
		ids[i] = model.GetID()
	}

	// Create a filter to delete
	filter := bson.M{"_id": bson.M{"$in": ids}}

	return filter, nil
}

// getReadPref return a readpref.ReadPref
//
// return a readpref.ReadPref connection
func getReadPref() *readpref.ReadPref {
	if config.MONGODB_READPREF == "" {
		return nil
	}

	switch config.MONGODB_READPREF {
	case config.MONGODB_READPREF:
		return readpref.Secondary()
	case config.MONGODB_READPREF_SECONDARY:
		return readpref.SecondaryPreferred()
	default:
		return readpref.Primary()
	}
}
