package mongoDB

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestCommandWithoutInitialize(t *testing.T) {

	ctx := context.Background()
	mongoDBInstance = nil

	t.Run("Should return error when execute command", func(t *testing.T) {
		err := NewCommand[User](ctx, nil).InsertOne()
		assert.EqualError(t, err, dbNotInitializedError)
	})

	t.Run("Delete", func(t *testing.T) {
		err := NewCommand[User](ctx, nil).DeleteOne()
		assert.EqualError(t, err, dbNotInitializedError)
	})
}

func TestInsert(t *testing.T) {
	ctx := context.Background()
	InitializeMongoDBTest(ctx)

	t.Run("Should return error when execute a command", func(t *testing.T) {
		err := NewCommand[User](ctx, nil).InsertOne()
		assert.EqualError(t, err, modelIsEmptyError)
	})

	t.Run("Should execute insert one", func(t *testing.T) {
		birth, _ := time.Parse("2006-01-02", "2021-11-22")
		user := &User{"123", "Usuário teste insert one", birth, Profile{"100", "ADMIN"}}

		err := NewCommand[User](ctx, user).InsertOne()
		assert.Nil(t, err)

		filter := bson.M{"name": "Usuário teste insert one"}
		result, err := NewQuery[User](ctx, filter).One()
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Usuário teste insert one", result.Name)
	})

	t.Run("Should execute insert many", func(t *testing.T) {
		birth, _ := time.Parse("2006-01-02", "2021-11-22")
		userA := &User{"123a", "Usuário teste insert many", birth, Profile{"100", "ADMIN"}}
		userB := &User{"123b", "Usuário teste insert many", birth, Profile{"100", "ADMIN"}}

		users := []*User{userA, userB}

		err := NewCommand[User](ctx, users).InsertMany()
		assert.Nil(t, err)

		filter := bson.M{"name": "Usuário teste insert many"}
		result, err := NewQuery[User](ctx, filter).Many()
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, len(result))
		assert.Equal(t, "123a", result[0].ID)
		assert.Equal(t, "123b", result[1].ID)
	})
}

func TestUpdate(t *testing.T) {
	ctx := context.Background()
	InitializeMongoDBTest(ctx)

	t.Run("Should return error when execute a command", func(t *testing.T) {
		err := NewCommand[User](ctx, nil).ReplaceOne()
		assert.EqualError(t, err, modelIsEmptyError)
	})

	t.Run("Should execute update one", func(t *testing.T) {
		birth, _ := time.Parse("2006-01-02", "2021-11-22")
		user := &User{"123", "Usuário teste insert one", birth, Profile{"100", "ADMIN"}}

		err := NewCommand[User](ctx, user).InsertOne()
		assert.Nil(t, err)

		filter := bson.M{"name": "Usuário teste insert one"}
		result, err := NewQuery[User](ctx, filter).One()
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Usuário teste insert one", result.Name)

		user.Name = "Usuário teste update one"
		err = NewCommand[User](ctx, user).ReplaceOne()
		assert.Nil(t, err)

		result, err = NewQuery[User](ctx, filter).One()
		assert.NoError(t, err)
		assert.Nil(t, result)

		filter = bson.M{"name": "Usuário teste update one"}
		result, err = NewQuery[User](ctx, filter).One()
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Usuário teste update one", result.Name)
	})
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	InitializeMongoDBTest(ctx)

	t.Run("Should return error when DeleteOne", func(t *testing.T) {
		err := NewCommand[User](ctx, nil).DeleteOne()
		assert.EqualError(t, err, modelIsEmptyError)
	})

	t.Run("Should execute delete one", func(t *testing.T) {
		birth, _ := time.Parse("2006-01-02", "2021-11-22")
		user := &User{"1234", "Usuário teste delete one", birth, Profile{"100", "ADMIN"}}

		err := NewCommand[User](ctx, user).InsertOne()
		assert.Nil(t, err)

		filter := bson.M{"name": "Usuário teste delete one"}
		result, err := NewQuery[User](ctx, filter).One()
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Usuário teste delete one", result.Name)

		err = NewCommand[User](ctx, user).DeleteOne()
		assert.Nil(t, err)

		result, err = NewQuery[User](ctx, filter).One()
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Should execute delete many", func(t *testing.T) {
		birth, _ := time.Parse("2006-01-02", "2021-11-22")
		userA := &User{"123c", "Usuário teste delete many", birth, Profile{"100", "ADMIN"}}
		userB := &User{"123d", "Usuário teste delete many", birth, Profile{"100", "ADMIN"}}

		users := []*User{userA, userB}

		err := NewCommand[User](ctx, users).InsertMany()
		assert.Nil(t, err)

		filter := bson.M{"name": "Usuário teste delete many"}
		result, err := NewQuery[User](ctx, filter).Many()
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, len(result))
		assert.Equal(t, "123c", result[0].ID)
		assert.Equal(t, "123d", result[1].ID)

		err = NewCommand[User](ctx, users).DeleteMany()
		assert.Nil(t, err)

		result, err = NewQuery[User](ctx, filter).Many()
		assert.NoError(t, err)
		assert.Nil(t, result)
	})
}
