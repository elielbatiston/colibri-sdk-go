package validator

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestCustomTagName(t *testing.T) {
	Initialize()

	t.Run("Should resolve field names using json and form tags", func(t *testing.T) {
		req := testRequest{}

		err := Struct(req)

		assert.Error(t, err)

		validationErrors, ok := err.(validator.ValidationErrors)
		assert.True(t, ok)

		fields := make(map[string]bool)
		for _, e := range validationErrors {
			fields[e.Field()] = true
		}

		assert.True(t, fields["sort"], "esperado campo 'sort'")
		assert.True(t, fields["name"], "esperado campo 'name'")
		assert.True(t, fields["Age"], "esperado fallback para 'Age'")
	})
}

type testRequest struct {
	Sort_ string `json:"sort" validate:"required"`
	Name  string `form:"name" validate:"required"`
	Age   int    `validate:"required"`
}
