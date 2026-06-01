// Package validatorx wires go-playground/validator into Gin so that binding
// failures surface as the structured, translated field errors defined in the
// response package. Call Setup once during startup.
package validatorx

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	"github.com/dirhamt/billing-hotspot/backend/internal/response"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTrans "github.com/go-playground/validator/v10/translations/en"
)

var translator ut.Translator

// Setup configures Gin's underlying validator engine: it makes error field
// names match JSON tags and installs English human-readable messages. Safe to
// call once at boot.
func Setup() error {
	engine, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return errors.New("validatorx: gin validator engine is not *validator.Validate")
	}

	// Report the JSON field name (not the Go struct field) in error details.
	engine.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" || name == "" {
			return fld.Name
		}
		return name
	})

	english := en.New()
	uni := ut.New(english, english)
	trans, _ := uni.GetTranslator("en")
	translator = trans
	return enTrans.RegisterDefaultTranslations(engine, trans)
}

// FieldErrors converts a binding/validation error into structured field errors.
// The boolean reports whether the error was a recognised validation error (as
// opposed to, say, malformed JSON which should be handled as a 400 instead).
func FieldErrors(err error) ([]response.FieldError, bool) {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return nil, false
	}

	details := make([]response.FieldError, 0, len(ve))
	for _, fe := range ve {
		msg := fe.Error()
		if translator != nil {
			msg = fe.Translate(translator)
		}
		details = append(details, response.FieldError{
			Field:   fe.Field(),
			Message: msg,
			Tag:     fe.Tag(),
		})
	}
	return details, true
}

// IsSyntaxError reports whether the error is a JSON decoding problem rather than
// a validation failure, so callers can return a 400 with a clear message.
func IsSyntaxError(err error) bool {
	var (
		syntax    *json.SyntaxError
		unmarshal *json.UnmarshalTypeError
	)
	return errors.As(err, &syntax) || errors.As(err, &unmarshal)
}
