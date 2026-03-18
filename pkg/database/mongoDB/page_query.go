package mongoDB

import (
	"context"
	"errors"

	"github.com/colibriproject-dev/colibri-sdk-go/pkg/base/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// PageQuery struct
type PageQuery[T MongoDBModel] struct {
	ctx    context.Context
	page   *types.PageRequest
	filter bson.M
}

// NewPageQuery create a new pointer to PageQuery struct
//
// ctx: the context.Context for the query
// page: the types.PageRequest for the query
// filter: bson.M for filter
// Returns a pointer to PageQuery struct
func NewPageQuery[T MongoDBModel](ctx context.Context, page *types.PageRequest, filter bson.M) *PageQuery[T] {
	return &PageQuery[T]{ctx, page, filter}
}

// Execute returns a pointer of page type with slice of T data
//
// No parameters.
// Returns a pointer to PageQuery struct and an error.
func (q *PageQuery[T]) Execute() (*types.Page[T], error) {
	return q.ExecuteInInstance(mongoDBInstance)
}

// ExecuteInInstance executes the page query in the given database instance.
//
// Parameters:
// - instance: the mongo.Client instance to execute the query in.
// Returns a Page of type T and an error.
func (q *PageQuery[T]) ExecuteInInstance(instance *mongo.Client) (*types.Page[T], error) {
	if err := q.validate(instance); err != nil {
		return nil, err
	}

	var result types.Page[T]
	var err error
	result.TotalItems, err = q.pageTotal(instance)
	if err != nil {
		return nil, err
	}

	result.Items, err = q.pageData(instance)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// pageTotal calculates the total number of records in the query result.
//
// Parameters:
// - instance: the mongodb client instance to execute the query in.
// Returns a uint64 representing the total number of records and an error.
func (q *PageQuery[T]) pageTotal(instance *mongo.Client) (uint64, error) {
	var model T
	totalItems, err := getMongoCollection(instance, model).CountDocuments(q.ctx, q.filter)
	if err != nil {
		return 0, err
	}

	return uint64(totalItems), nil
}

// pageData retrieves data for the page query from the given database instance.
//
// Parameters:
// - instance: the mongodb client instance to retrieve data from.
// Returns a slice of type T and an error.
func (q *PageQuery[T]) pageData(instance *mongo.Client) ([]T, error) {
	var model T
	cursor, err := getMongoCollection(instance, model).Find(q.ctx, q.filter, q.GetFindOptions())
	if err != nil {
		return nil, err
	}
	defer cursor.Close(q.ctx)

	return getDataList[T](cursor)
}

// validate checks if the PageQuery instance is initialized, if the page is empty, and if the query is empty.
//
// instance: the mongodb client instance to validate against
// Returns an error.
func (q *PageQuery[T]) validate(instance *mongo.Client) error {
	if instance == nil {
		return errors.New(dbNotInitializedError)
	}

	if q.page == nil {
		return errors.New(pageIsEmptyError)
	}

	return nil
}

func (q *PageQuery[T]) GetFindOptions() *options.FindOptions {
	limit := int64(q.page.Size)
	skip := int64((q.page.Page - 1) * q.page.Size)

	return &options.FindOptions{
		Limit: &limit,
		Skip:  &skip,
		Sort:  q.page.GetMongoOrder(),
	}
}
