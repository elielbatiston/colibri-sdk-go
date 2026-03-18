package mongoDB

import (
	"context"
	"testing"

	"github.com/colibriproject-dev/colibri-sdk-go/pkg/base/types"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestPageQueryWithoutInitialize(t *testing.T) {

	mongoDBInstance = nil

	t.Run("Should return error when execute page query with db not initialized error", func(t *testing.T) {
		orders := []types.Sort{
			{Direction: types.DESC, Field: "name"},
			{Direction: types.ASC, Field: "birthday"},
		}
		page := types.NewPageRequest(1, 1, orders)
		_, err := NewPageQuery[User](context.Background(), page, bson.M{}).Execute()
		assert.EqualError(t, err, dbNotInitializedError)
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
		assert.EqualError(t, err, pageIsEmptyError)
	})

	t.Run("Should execute page query", func(t *testing.T) {
		result, err := NewPageQuery[User](ctx, page, bson.M{}).Execute()
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, len(result.Items))
		assert.Equal(t, "ADMIN USER", result.Items[0].Name)
		assert.Equal(t, uint64(2), result.TotalItems)
	})
}
