package nosqlDB

import (
	"context"
	"testing"
	"time"

	"github.com/colibri-project-io/colibri-sdk-go/pkg/base/test"
	"github.com/colibri-project-io/colibri-sdk-go/pkg/database/cacheDB"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestQueryWithoutInitialize(t *testing.T) {
	ctx := context.Background()
	nosqlDBInstance = nil

	t.Run("Should return error when execute query one without filter with db not initialized error", func(t *testing.T) {
		result, err := NewQuery[User](ctx, nil).One()
		assert.EqualError(t, err, nosql_db_not_initialized_error)
		assert.Nil(t, result)
	})

	t.Run("Should return error when execute query one with filter with db not initialized error", func(t *testing.T) {
		filter := bson.M{"_id": 1}
		result, err := NewQuery[User](ctx, filter).One()
		assert.EqualError(t, err, nosql_db_not_initialized_error)
		assert.Nil(t, result)
	})

	t.Run("Should return error when execute query many without filter with db not initialized error", func(t *testing.T) {
		result, err := NewQuery[User](ctx, nil).Many()
		assert.EqualError(t, err, nosql_db_not_initialized_error)
		assert.Nil(t, result)
	})

	t.Run("Should return error when execute query many with filter with db not initialized error", func(t *testing.T) {
		filter := bson.M{"_id": 1}
		result, err := NewQuery[User](ctx, filter).Many()
		assert.EqualError(t, err, nosql_db_not_initialized_error)
		assert.Nil(t, result)
	})
}

func TestQuery(t *testing.T) {
	ctx := context.Background()
	InitializeMongoDBTest(ctx)

	t.Run("Should execute one without filter", func(t *testing.T) {
		result, err := NewQuery[User](ctx, nil).One()
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "ADMIN USER", result.Name)
	})

	t.Run("Should execute one with filter", func(t *testing.T) {
		filter := bson.M{"name": "ADMIN USER"}
		result, err := NewQuery[User](ctx, filter).One()
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "ADMIN USER", result.Name)
	})

	t.Run("Should execute many without filter", func(t *testing.T) {
		result, err := NewQuery[User](ctx, nil).Many()
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 2)
		assert.Equal(t, "ADMIN USER", result[0].Name)
		assert.Equal(t, "OTHER USER", result[1].Name)
	})

	t.Run("Should execute many with filter", func(t *testing.T) {
		filter := bson.M{"name": "ADMIN USER"}
		result, err := NewQuery[User](ctx, filter).Many()
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 1)
		assert.Equal(t, "ADMIN USER", result[0].Name)
	})
}

func TestQueryWithoutCacheDBInitialize(t *testing.T) {
	cache := cacheDB.NewCache[User]("TestQueryWithoutCacheDBInitialize", time.Hour)
	ctx := context.Background()
	InitializeMongoDBTest(ctx)

	t.Run("Should return error when one without filter with cache", func(t *testing.T) {
		dbResult, dbErr := NewCachedQuery(ctx, cache, nil).One()
		cacheResult, cacheErr := cache.One(ctx)

		assert.NoError(t, dbErr)
		assert.NotNil(t, dbResult)
		assert.Equal(t, "ADMIN USER", dbResult.Name)
		assert.Error(t, cacheErr, "Cache not initialized")
		assert.Nil(t, cacheResult)
	})

	t.Run("Should return error when one with filter with cache", func(t *testing.T) {
		filter := bson.M{"name": "ADMIN USER"}
		dbResult, dbErr := NewCachedQuery(ctx, cache, filter).One()
		cacheResult, cacheErr := cache.One(ctx)

		assert.NoError(t, dbErr)
		assert.NotNil(t, dbResult)
		assert.Equal(t, "ADMIN USER", dbResult.Name)
		assert.Error(t, cacheErr, "Cache not initialized")
		assert.Nil(t, cacheResult)
	})

	t.Run("Should return error when many without filter with cache", func(t *testing.T) {
		dbResult, dbErr := NewCachedQuery(ctx, cache, nil).Many()
		cacheResult, cacheErr := cache.Many(ctx)

		assert.NoError(t, dbErr)
		assert.NotNil(t, dbResult)
		assert.Len(t, dbResult, 2)
		assert.Equal(t, "ADMIN USER", dbResult[0].Name)
		assert.Error(t, cacheErr, "Cache not initialized")
		assert.Nil(t, cacheResult)
	})

	t.Run("Should return error when many with params with cache", func(t *testing.T) {
		filter := bson.M{"name": "ADMIN USER"}
		dbResult, dbErr := NewCachedQuery(ctx, cache, filter).Many()
		cacheResult, cacheErr := cache.Many(ctx)

		assert.NoError(t, dbErr)
		assert.NotNil(t, dbResult)
		assert.Len(t, dbResult, 1)
		assert.Equal(t, "ADMIN USER", dbResult[0].Name)
		assert.Error(t, cacheErr, "Cache not initialized")
		assert.Nil(t, cacheResult)
	})
}

func TestCachedQuery(t *testing.T) {
	ctx := context.Background()

	InitializeMongoDBTest(ctx)
	test.InitializeCacheDBTest()
	cacheDB.Initialize()

	cache := cacheDB.NewCache[User]("TestCachedQuery", time.Hour)

	t.Run("Should execute one without filter with cache", func(t *testing.T) {
		cacheInitialData, cacheInitialErr := cache.One(ctx)
		result, err := NewCachedQuery(ctx, cache, nil).One()
		cacheFinalData, cacheFinalErr := cache.One(ctx)
		cacheDelErr := cache.Del(ctx)

		assert.NoError(t, cacheInitialErr)
		assert.Nil(t, cacheInitialData)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "ADMIN USER", result.Name)
		assert.NoError(t, cacheFinalErr)
		assert.NotNil(t, cacheFinalData)
		assert.Equal(t, "ADMIN USER", cacheFinalData.Name)
		assert.NoError(t, cacheDelErr)
	})

	t.Run("Should execute one with filter with cache", func(t *testing.T) {
		cacheInitialData, cacheInitialErr := cache.One(ctx)
		filter := bson.M{"name": "ADMIN USER"}
		result, err := NewCachedQuery(ctx, cache, filter).One()
		cacheFinalData, cacheFinalErr := cache.One(ctx)
		cacheDelErr := cache.Del(ctx)

		assert.NoError(t, cacheInitialErr)
		assert.Nil(t, cacheInitialData)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "ADMIN USER", result.Name)
		assert.NoError(t, cacheFinalErr)
		assert.NotNil(t, cacheFinalData)
		assert.Equal(t, "ADMIN USER", cacheFinalData.Name)
		assert.NoError(t, cacheDelErr)
	})

	t.Run("Should execute many without filter with cache", func(t *testing.T) {
		cacheInitialData, cacheInitialErr := cache.One(ctx)
		result, err := NewCachedQuery(ctx, cache, nil).Many()
		cacheFinalData, cacheFinalErr := cache.Many(ctx)
		cacheDelErr := cache.Del(ctx)

		assert.NoError(t, cacheInitialErr)
		assert.Nil(t, cacheInitialData)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 2)
		assert.Equal(t, "ADMIN USER", result[0].Name)
		assert.NoError(t, cacheFinalErr)
		assert.NotNil(t, cacheFinalData)
		assert.Len(t, result, 2)
		assert.Equal(t, "ADMIN USER", result[0].Name)
		assert.NoError(t, cacheDelErr)
	})

	t.Run("Should execute many with params with cache", func(t *testing.T) {
		cacheInitialData, cacheInitialErr := cache.One(ctx)
		filter := bson.M{"name": "ADMIN USER"}
		result, err := NewCachedQuery(ctx, cache, filter).Many()
		cacheFinalData, cacheFinalErr := cache.Many(ctx)
		cacheDelErr := cache.Del(ctx)

		assert.NoError(t, cacheInitialErr)
		assert.Nil(t, cacheInitialData)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 1)
		assert.Equal(t, "ADMIN USER", result[0].Name)
		assert.NoError(t, cacheFinalErr)
		assert.NotNil(t, cacheFinalData)
		assert.Len(t, result, 1)
		assert.Equal(t, "ADMIN USER", result[0].Name)
		assert.NoError(t, cacheDelErr)
	})
}
