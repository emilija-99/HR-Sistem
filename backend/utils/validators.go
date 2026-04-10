package utils

import "github.com/go-playground/validator/v10"

type Validator struct {
	V *validator.Validate
}

func NewValidator() *Validator {
	v := validator.New()

	v.RegisterValidation("strongpwd", strongPassword)

	return &Validator{V: v}
}

func strongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	return len(password) >= 8 // extend later
}
