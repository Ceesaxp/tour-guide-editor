// internal/validators/validators.go
package validators

import (
	"github.com/go-playground/validator/v10"
)

// RegisterCustomValidations adds custom validation rules to the validator
func RegisterCustomValidations(v *validator.Validate) error {
	// Register required_if validation
	if err := v.RegisterValidation("required_if", requiredIf); err != nil {
		return err
	}
	return nil
}

// requiredIf implements conditional required validation
func requiredIf(fl validator.FieldLevel) bool {
	// Get the field name and expected value from the tag
	params := fl.Param()
	if params == "" {
		return true
	}

	// For our specific case with quiz type
	field := fl.Parent().FieldByName(params)
	if !field.IsValid() {
		return true
	}

	// Check if the field matches the expected value
	if field.String() == "quiz" {
		// For quiz type, options should not be empty
		return len(fl.Field().Interface().([]string)) >= 2
	}

	return true
}
