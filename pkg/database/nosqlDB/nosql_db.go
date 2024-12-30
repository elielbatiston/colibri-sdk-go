package nosqlDB

import (
	"context"
	"fmt"
	"reflect"

	"github.com/colibri-project-io/colibri-sdk-go/pkg/base/config"
	"github.com/colibri-project-io/colibri-sdk-go/pkg/base/logging"
	"github.com/colibri-project-io/colibri-sdk-go/pkg/base/observer"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	nosql_db_connection_success    string = "NoSQL database connected"
	nosql_db_connection_error      string = "An error occurred while trying to connect to the nosql database. Error: %s"
	nosql_db_not_initialized_error string = "database not initialized"
	nosql_model_is_empty_error     string = "model is empty"
	filter_is_empty_error          string = "filter is empty"
	collection_is_empty_error      string = "collection name is empty"
	page_is_empty_error            string = "page is empty"
)

// nosqlDBObserver is a struct for nosql database observer.
type nosqlDBObserver struct{}

// nosqlDBInstance is a pointer to mongo.Client
var nosqlDBInstance *mongo.Client
var nosqlDBCtxInstance context.Context

// Initialize start connection with nosql database and execute migration.
//
// No parameters.
// No return values.
func Initialize(ctx context.Context) {
	nosqlDBCtxInstance = ctx
	nosqlDB := NewNoSQLDatabaseInstance("NoSQL", config.NOSQL_DB_CONNECTION_URI)
	nosqlDBInstance = nosqlDB
}

// NewNoSQLDatabaseInstance creates a new NoSQL database instance.
//
// Parameters:
// - name: a string representing the name of the database.
// - databaseURL: a string representing the URL of the database.
// Returns a pointer to mongo.Client.
func NewNoSQLDatabaseInstance(name string, dsn string) *mongo.Client {
	clientOptions := options.Client().
		ApplyURI(dsn).
		SetMaxPoolSize(uint64(config.NOSQL_DB_MAX_POOL_SIZE)).
		SetMinPoolSize(uint64(config.NOSQL_DB_MIN_POOL_SIZE))

	nosqlDB, err := mongo.Connect(nosqlDBCtxInstance, clientOptions)
	if err != nil {
		logging.Fatal(nosql_db_connection_error, err)
	}

	if err = nosqlDB.Ping(nosqlDBCtxInstance, nil); err != nil {
		logging.Fatal(nosql_db_connection_error, err)
	}

	logging.Info(nosql_db_connection_success)
	observer.Attach(nosqlDBObserver{})

	return nosqlDB
}

// Close finalize sql database connection
//
// No parameters.
// No return values.
func (o nosqlDBObserver) Close() {
	logging.Info("closing nosql database connection")
	if err := nosqlDBInstance.Disconnect(nosqlDBCtxInstance); err != nil {
		logging.Error("error when closing nosql database connection: %v", err)
		return
	}
}

// getDataList retrieves a list of items from the given mongo.Client object.
//
// It takes a mongo.Client object as input and returns a list of items of type T and an error.
func getDataList[T any](cursor *mongo.Cursor) ([]T, error) {
	list := make([]T, 0)
	exist := false
	for cursor.Next(nosqlDBCtxInstance) {
		exist = true
		model := new(T)
		err := cursor.Decode(&model)
		if err != nil {
			return nil, err
		}

		list = append(list, *model)
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
func getMongoCollection(instance *mongo.Client, model interface{}) *mongo.Collection {
	collectionName := collectionName(model)
	return instance.Database(config.NOSQL_DB_NAME).Collection(collectionName)
}

// collectionName returns the name of the struct
//
// any object
// return a string containing the name of struct
func collectionName(model interface{}) string {
	typeOf := reflect.TypeOf(model)

	if typeOf.Kind() == reflect.Ptr {
		typeOf = typeOf.Elem()
	}

	if typeOf.Kind() == reflect.Slice {
		typeOf = typeOf.Elem()
	}

	return typeOf.Name()
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

// buildFilterFromModel return a bson.M containig a filter with the Ids of a []interface{}
//
// interface{}
// return a bson.M containig a filter with the Ids of a []interface{} or an error
func buildFilterFromModel(model interface{}) (bson.M, error) {
	v := reflect.ValueOf(model)
	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("expected a slice, got %v", v.Kind())
	}

	var ids []interface{}
	for i := 0; i < v.Len(); i++ {
		item := v.Index(i).Interface()

		field := reflect.ValueOf(item).FieldByName("Id")
		if field.IsValid() {
			ids = append(ids, field.Interface())
		} else {
			field = reflect.ValueOf(item).FieldByName("Uuid")
			if field.IsValid() {
				ids = append(ids, field.Interface())
			}
		}
	}

	if len(ids) == 0 {
		return nil, fmt.Errorf("no valid IDs found")
	}

	// Criar o filtro para o delete
	filter := bson.M{"_id": bson.M{"$in": ids}}
	return filter, nil
}
