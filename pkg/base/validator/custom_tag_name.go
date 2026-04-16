package validator

import (
	"reflect"
	"strings"
)

func registerCustomTagName() {
	instance.validator.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.Split(fld.Tag.Get("json"), ",")[0]
		if name != "" && name != "-" {
			return name
		}

		// Fallback para a tag form
		name = strings.Split(fld.Tag.Get("form"), ",")[0]
		if name != "" && name != "-" {
			return name
		}

		return fld.Name
	})
}
