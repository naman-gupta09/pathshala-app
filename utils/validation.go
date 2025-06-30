package utils

import (
	"github.com/go-playground/validator/v10"
)

func FormatValidationError(err error) map[string]string {
	res := map[string]string{}

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldErr := range validationErrors {
			fieldName := fieldErr.Field()
			tag := fieldErr.Tag()

			switch tag {
			case "required":
				res[fieldName] = "This field is required"
			case "email":
				res[fieldName] = "Invalid email format"
			case "oneof":
				res[fieldName] = "Invalid value"
			case "gte":
				res[fieldName] = "Must be at least " + fieldErr.Param()
			default:
				res[fieldName] = "Invalid value"
			}
		}
	}

	return res
}
