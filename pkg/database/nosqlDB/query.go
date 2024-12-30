package nosqlDB

import (
	"context"
	"errors"

	"github.com/colibri-project-io/colibri-sdk-go/pkg/database/cacheDB"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Query is a struct for nosql query
type Query[T any] struct {
	ctx    context.Context
	cache  *cacheDB.Cache[T]
	filter bson.M
}

// NewQuery create a new pointer to Query struct
//
// ctx: the context.Context for the query
// filter: bson.M for filter
// Returns a pointer to Query struct
func NewQuery[T any](ctx context.Context, filter bson.M) *Query[T] {
	return &Query[T]{ctx, nil, filter}
}

// NewCachedQuery create a new pointer to Query struct with cache
//
// ctx: the context.Context for the query
// cache: the cacheDB.Cache to store the query result
// filter: bson.M for filter
// Returns a pointer to Query struct
func NewCachedQuery[T any](ctx context.Context, cache *cacheDB.Cache[T], filter bson.M) (q *Query[T]) {
	return &Query[T]{ctx, cache, filter}
}

// Many returns a slice of T value
//
// No parameters are required. Returns a slice of T value and an error.
func (q *Query[T]) Many() ([]T, error) {
	return q.ManyInInstance(nosqlDBInstance)
}

// ManyInInstance retrieves multiple items of type T for the given NoSQL instance.
//
// instance: The *mongo.Client instance to execute the query.
// Returns a slice of retrieved items of type T and an error.
func (q *Query[T]) ManyInInstance(instance *mongo.Client) ([]T, error) {
	if err := q.validate(instance); err != nil {
		return nil, err
	}

	if q.cache == nil {
		return q.fetchMany(instance)
	}

	result, err := q.cache.Many(q.ctx)
	if result == nil || err != nil {
		return q.fetchMany(instance)
	}
	return result, nil
}

// fetchMany retrieves multiple items of type T for the given NoSQL instance.
//
// instance: The *mongo.Client instance to execute the query.
// Returns a slice of retrieved items of type T and an error.
func (q *Query[T]) fetchMany(instance *mongo.Client) ([]T, error) {
	model := new(T)
	cursor, err := getMongoCollection(instance, model).Find(q.ctx, q.filter, nil)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(q.ctx)

	list, err := getDataList[T](cursor)
	if err != nil {
		return nil, err
	}

	if q.cache != nil {
		q.cache.Set(q.ctx, list)
	}

	return list, nil
}

// One return a pointer of T value
//
// No parameters.
// Returns a pointer of T and an error.
func (q *Query[T]) One() (*T, error) {
	return q.OneInInstance(nosqlDBInstance)
}

// OneInInstance retrieves a single item of type T for the given NoSQL instance.
//
// instance: The *mongo.Client instance to execute the query.
// Returns a pointer of T and an error.
func (q *Query[T]) OneInInstance(instance *mongo.Client) (*T, error) {
	if err := q.validate(instance); err != nil {
		return nil, err
	}

	if q.cache == nil {
		return q.fetchOne(instance)
	}

	result, err := q.cache.One(q.ctx)
	if result == nil || err != nil {
		return q.fetchOne(instance)
	}
	return result, nil
}

// fetchOne retrieves a single item of type T for the given NoSQL instance.
//
// instance: The *mongo.Client instance to execute the query.
// Returns a pointer of T and an error.
func (q *Query[T]) fetchOne(instance *mongo.Client) (*T, error) {
	model := new(T)
	err := getMongoCollection(instance, model).FindOne(q.ctx, q.filter).Decode(model)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, err
	} else if err == mongo.ErrNoDocuments {
		return nil, nil
	}

	if q.cache != nil {
		q.cache.Set(q.ctx, model)
	}

	return model, nil
}

// validate checks if the Query instance is initialized and if the query is empty.
//
// instance: The *mongo.Client instance to execute the query.
// Returns an error.
func (q *Query[T]) validate(instance *mongo.Client) error {
	if instance == nil {
		return errors.New(nosql_db_not_initialized_error)
	}

	return nil
}
