package nosqlDB

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Command struct
type Command struct {
	ctx   context.Context
	model interface{}
}

type DeleteManyCommand struct {
	ctx    context.Context
	filter bson.M
}

// NewCommand create a new pointer to Command struct
func NewCommand(ctx context.Context, model interface{}) *Command {
	return &Command{ctx, model}
}

// InsertOne insert a document into the database
//
// No parameters.
// Returns an Error
func (c *Command) InsertOne() error {
	return c.InsertOneInInstance(nosqlDBInstance)
}

// InsertOneInInstance inserts a single document in the provided database instance.
//
// instance: The *mongo.Client instance to execute.
// Returns an error.
func (c *Command) InsertOneInInstance(instance *mongo.Client) error {
	if err := c.validate(instance); err != nil {
		return err
	}

	_, err := getMongoCollection(nosqlDBInstance, c.model).InsertOne(c.ctx, c.model)
	if err != nil {
		return err
	}

	return nil
}

// InsertMany insert many documents into the database
//
// No parameters.
// Returns an Error
func (c *Command) InsertMany() error {
	return c.InsertManyInInstance(nosqlDBInstance)
}

// InsertManyInInstance inserts multiple documents in the provided database instance.
//
// instance: The *mongo.Client instance to execute.
// Returns an error.
func (c *Command) InsertManyInInstance(instance *mongo.Client) error {
	if err := c.validate(instance); err != nil {
		return err
	}

	models, err := c.modelToSlice()
	if err != nil {
		return err
	}

	_, err = getMongoCollection(nosqlDBInstance, c.model).InsertMany(c.ctx, models)
	if err != nil {
		return err
	}

	return nil
}

// DeleteOne delete a document from a database
//
// No parameters.
// Returns an error.
func (c *Command) DeleteOne() error {
	return c.DeleteOneInInstance(nosqlDBInstance)
}

// DeleteOneInInstance delete a single document in the provided database instance.
//
// instance: The *mongo.Client instance to execute.
// Returns an error.
func (c *Command) DeleteOneInInstance(instance *mongo.Client) error {
	if err := c.validate(instance); err != nil {
		return err
	}

	_, err := getMongoCollection(nosqlDBInstance, c.model).DeleteOne(c.ctx, c.model)
	if err != nil {
		return err
	}

	return nil
}

// DeleteMany delete many documents from a database
//
// filter: bson.M for filter
// Returns an error.
func (c *Command) DeleteMany() error {
	return c.DeleteManyInInstance(nosqlDBInstance)
}

// DeleteManyInInstance delete multiple documents in the provided database instance.
//
// instance: The *mongo.Client instance to execute and bson.M for filter.
// Returns an error.
func (c *Command) DeleteManyInInstance(instance *mongo.Client) error {
	if err := c.validate(instance); err != nil {
		return err
	}

	filter, err := buildFilterFromModel(c.model)
	if err != nil {
		return err
	}

	_, err = getMongoCollection(nosqlDBInstance, c.model).DeleteMany(c.ctx, filter)
	if err != nil {
		return err
	}

	return nil
}

// validate checks if the Command instance is initialized.
//
// filter: bson.M for filter
// Returns an error.
func (c *Command) validate(instance *mongo.Client) error {
	if instance == nil {
		return errors.New(nosql_db_not_initialized_error)
	}

	if isStructEmpty(c.model) {
		return errors.New(nosql_model_is_empty_error)
	}

	return nil
}

func (c *Command) modelToSlice() ([]interface{}, error) {
	v := reflect.ValueOf(c.model)
	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("model is not a slice")
	}

	var docs []interface{}
	for i := 0; i < v.Len(); i++ {
		docs = append(docs, v.Index(i).Interface())
	}

	return docs, nil
}
