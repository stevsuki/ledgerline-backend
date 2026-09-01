// Package validator: a go-playground/validator wrapper + error message translation.
package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"github.com/stevensuki/ledgerline-backend/pkg/response"
)

var validate *validator.Validate

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterTagNameFunc(tagNameFunc)
}

// tagNameFunc: use the field name from the json/form tag, not the Go field name.
func tagNameFunc(fld reflect.StructField) string {
	for _, tag := range []string{"json", "form", "uri"} {
		name := strings.SplitN(fld.Tag.Get(tag), ",", 2)[0]
		if name != "" && name != "-" {
			return name
		}
	}
	return fld.Name
}

// RegisterGinValidator: apply the same configuration to the internal Gin validator.
func RegisterGinValidator() {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}
	v.RegisterTagNameFunc(tagNameFunc)
}

// Translate: validator errors -> []response.FieldError.
func Translate(err error) []response.FieldError {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return nil
	}

	out := make([]response.FieldError, 0, len(ve))
	for _, fe := range ve {
		out = append(out, response.FieldError{Field: fe.Field(), Message: message(fe)})
	}
	return out
}

func message(fe validator.FieldError) string {
	field := fe.Field()
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters/value", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters/value", field, fe.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, fe.Param())
	case "uuid", "uuid4":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "eqfield":
		return fmt.Sprintf("%s must equal %s", field, fe.Param())
	case "gte":
		return fmt.Sprintf("%s must be at least %s", field, fe.Param())
	case "lte":
		return fmt.Sprintf("%s must be at most %s", field, fe.Param())
	default:
		return fmt.Sprintf("%s is invalid (rule: %s)", field, fe.Tag())
	}
}
