package middleware

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
)

// CustomValidator ทำให้ router เรียก c.Validate(&request) ได้
// โดยตรวจตาม tag `validate:"..."` ใน models
type CustomValidator struct {
	Validator *validator.Validate
}

func NewCustomValidator() *CustomValidator {
	return &CustomValidator{Validator: validator.New()}
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.Validator.Struct(i); err != nil {
		var errs validator.ValidationErrors
		if errors.As(err, &errs) {
			var messages []string
			for _, e := range errs {
				switch e.Tag() {
				case "required":
					messages = append(messages, e.Field()+" is required field")
				case "gt":
					messages = append(messages, e.Field()+" must be greater than "+e.Param())
				default:
					messages = append(messages, e.Field()+" is invalid")
				}
			}
			return errors.New(strings.ToLower(strings.Join(messages, ", ")))
		}
		return err
	}
	return nil
}
