package utils

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validator *validator.Validate
}

func (v *Validator) Validate(i interface{}) error {
	if err := v.validator.Struct(i); err != nil {
		var validationErr validator.ValidationErrors
		if errors.As(err, &validationErr) {
			return convertToUserError(validationErr)
		}
	}
	return nil
}

func NewValidator(ctx context.Context) *Validator {
	v := validator.New()
	v.RegisterTagNameFunc(nameTagFunction)
	return &Validator{
		validator: v,
	}
}

func nameTagFunction(fld reflect.StructField) string {
	name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
	if name == "-" {
		return ""
	}
	return name
}

func convertToUserError(err error) error {
	switch t := err.(type) {
	case validator.ValidationErrors:
		var validationErr string
		for _, v := range t {
			validationErr += fmt.Sprintf("%s - %s ", v.Field(), convertTag(v.Tag()))
		}
		validationErr = "Validation failed for the following field(s): " + validationErr[:len(validationErr)-2]
		return errors.New(validationErr)
	case *validator.InvalidValidationError:
		return fmt.Errorf("cannot validate request with body: %s", t.Type)
	default:
		return t

	}
}

func convertTag(tag string) string {
	switch tag {
	case "required":
		return "is required"
	default:
		return tag
	}
}
