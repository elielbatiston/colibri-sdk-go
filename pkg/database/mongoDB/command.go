package mongoDB

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Command struct
type Command[T MongoDBModel] struct {
	ctx   context.Context
	model interface{}
}

type DeleteManyCommand struct {
	ctx    context.Context
	filter bson.M
}

// NewCommand create a new pointer to Command struct
func NewCommand[T MongoDBModel](ctx context.Context, model interface{}) *Command[T] {
	return &Command[T]{ctx, model}
}

// InsertOne insert a document into the database
//
// No parameters.
// Returns an Error
func (c *Command[T]) InsertOne() error {
	return c.InsertOneInInstance(mongoDBInstance)
}

// InsertOneInInstance inserts a single document in the provided database instance.
//
// instance: The *mongo.Client instance to execute.
// Returns an error.
func (c *Command[T]) InsertOneInInstance(instance *mongo.Client) error {
	if err := c.validate(instance, c.model); err != nil {
		return err
	}

	var model MongoDBModel
	var ok bool
	if model, ok = c.model.(MongoDBModel); !ok {
		return errors.New("the model is not a MongoDBModel")
	}

	opts := &options.InsertOneOptions{Comment: model}
	_, err := getMongoCollection(mongoDBInstance, model).InsertOne(c.ctx, model, opts)
	if err != nil {
		return err
	}

	return nil
}

// InsertMany insert many documents into the database
//
// No parameters.
// Returns an Error
func (c *Command[T]) InsertMany() error {
	return c.InsertManyInInstance(mongoDBInstance)
}

// InsertManyInInstance inserts multiple documents in the provided database instance.
//
// instance: The *mongo.Client instance to execute.
// Returns an error.
func (c *Command[T]) InsertManyInInstance(instance *mongo.Client) error {
	if err := c.validate(instance, c.model); err != nil {
		return err
	}

	modelsSlice, err := c.modelToSlice(c.model)
	if err != nil {
		return err
	}

	var model T
	opts := &options.InsertManyOptions{Comment: model}
	_, err = getMongoCollection(mongoDBInstance, model).InsertMany(c.ctx, modelsSlice, opts)
	if err != nil {
		return err
	}

	return nil
}

// ReplaceOne update an existing document in the database
//
// No parameters.
// Returns an Error
func (c *Command[T]) ReplaceOne() error {
	return c.ReplaceOneInInstance(mongoDBInstance)
}

// ReplaceOneInInstance updates a single document in the provided database instance.
//
// instance: The *mongo.Client instance to execute.
// Returns an error.
func (c *Command[T]) ReplaceOneInInstance(instance *mongo.Client) error {
	if err := c.validate(instance, c.model); err != nil {
		return err
	}

	var model MongoDBModel
	var ok bool
	if model, ok = c.model.(MongoDBModel); !ok {
		return errors.New("the model is not a MongoDBModel")
	}

	filter := bson.M{"_id": model.GetID()}
	opts := &options.ReplaceOptions{Comment: model}
	_, err := getMongoCollection(mongoDBInstance, model).ReplaceOne(c.ctx, filter, model, opts)
	if err != nil {
		return err
	}

	return nil
}

// DeleteOne delete a document from a database
//
// No parameters.
// Returns an error.
func (c *Command[T]) DeleteOne() error {
	return c.DeleteOneInInstance(mongoDBInstance)
}

// DeleteOneInInstance delete a single document in the provided database instance.
//
// instance: The *mongo.Client instance to execute.
// Returns an error.
func (c *Command[T]) DeleteOneInInstance(instance *mongo.Client) error {
	if err := c.validate(instance, c.model); err != nil {
		return err
	}

	var model MongoDBModel
	var ok bool
	if model, ok = c.model.(MongoDBModel); !ok {
		return errors.New("the model is not a MongoDBModel")
	}

	filter := bson.M{"_id": model.GetID()}
	opts := &options.DeleteOptions{Comment: model}
	_, err := getMongoCollection(mongoDBInstance, model).DeleteOne(c.ctx, filter, opts)
	if err != nil {
		return err
	}

	return nil
}

// DeleteMany delete many documents from a database
//
// filter: bson.M for filter
// Returns an error.
func (c *Command[T]) DeleteMany() error {
	return c.DeleteManyInInstance(mongoDBInstance)
}

// DeleteManyInInstance delete multiple documents in the provided database instance.
//
// instance: The *mongo.Client instance to execute and bson.M for filter.
// Returns an error.
func (c *Command[T]) DeleteManyInInstance(instance *mongo.Client) error {
	if err := c.validate(instance, c.model); err != nil {
		return err
	}

	models, err := c.toMongoDBModel()
	if err != nil {
		return err
	}

	var model T
	filter := buildFilterFromModels(models)
	opts := &options.DeleteOptions{Comment: models}
	_, err = getMongoCollection(mongoDBInstance, model).DeleteMany(c.ctx, filter, opts)
	if err != nil {
		return err
	}

	return nil
}

// validate checks if the Command instance is initialized.
//
// filter: bson.M for filter
// Returns an error.
func (c *Command[T]) validate(instance *mongo.Client, model interface{}) error {
	if instance == nil {
		return errors.New(dbNotInitializedError)
	}

	if isStructEmpty(model) {
		return errors.New(modelIsEmptyError)
	}

	return nil
}

func (c *Command[T]) modelToSlice(models interface{}) ([]interface{}, error) {
	v := reflect.ValueOf(models)
	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("model is not a slice")
	}

	var newModels []interface{}
	for i := 0; i < v.Len(); i++ {
		model := v.Index(i).Interface()
		value, ok := model.(MongoDBModel)
		if !ok {
			return nil, errors.New("the models are not a MongoDBModel")
		}

		newModels = append(newModels, value)
	}

	return newModels, nil
}

func (c *Command[T]) toMongoDBModel() ([]MongoDBModel, error) {
	var models []MongoDBModel

	modelsSlice, err := c.modelToSlice(c.model)
	if err != nil {
		return nil, err
	}

	for _, model := range modelsSlice {
		value, ok := model.(MongoDBModel)
		if value, ok = model.(MongoDBModel); !ok {
			return nil, errors.New("the models are not a MongoDBModel")
		}

		models = append(models, value)
	}

	return models, nil
}
