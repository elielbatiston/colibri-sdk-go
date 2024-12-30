package nosqlDB

import (
	"context"
	"testing"

	"github.com/colibri-project-io/colibri-sdk-go/pkg/base/types"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestPageQueryWithoutInitialize(t *testing.T) {

	nosqlDBInstance = nil

	t.Run("Should return error when execute page query with db not initialized error", func(t *testing.T) {
		orders := []types.Sort{
			{Direction: types.DESC, Field: "name"},
			{Direction: types.ASC, Field: "birthday"},
		}
		page := types.NewPageRequest(1, 1, orders)
		_, err := NewPageQuery[User](context.Background(), page, bson.M{}).Execute()
		assert.EqualError(t, err, nosql_db_not_initialized_error)
	})
}

func TestPageQuery(t *testing.T) {
	ctx := context.Background()
	InitializeMongoDBTest(ctx)

	orders := []types.Sort{
		{Direction: types.DESC, Field: "name"},
		{Direction: types.ASC, Field: "birthday"},
	}
	page := types.NewPageRequest(1, 1, orders)

	t.Run("Should return error when execute page query without page info", func(t *testing.T) {
		_, err := NewPageQuery[User](ctx, nil, bson.M{}).Execute()
		assert.EqualError(t, err, page_is_empty_error)
	})

	t.Run("Should execute page query", func(t *testing.T) {
		result, err := NewPageQuery[User](ctx, page, bson.M{}).Execute()
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, len(result.Content))
		assert.Equal(t, "ADMIN USER", result.Content[0].Name)
		assert.Equal(t, uint64(2), result.TotalElements)
	})
}
