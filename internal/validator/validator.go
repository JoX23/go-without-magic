package validator

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
)

// Validator encapsula el validador de structs.
type Validator struct {
	v *validator.Validate
}

// New crea un validador con reglas y traducciones por defecto.
func New() *Validator {
	return &Validator{v: validator.New()}
}

// Validate valida un struct con tags validate.
func (v *Validator) Validate(i interface{}) error {
	if err := v.v.Struct(i); err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			return ValidationErrorsFrom(ve)
		}
		return err
	}
	return nil
}

// ParseJSON y valida body JSON en struct target.
func (v *Validator) ParseJSON(r *http.Request, target interface{}) error {
	if r.Body == nil {
		return fmt.Errorf("empty request body")
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return &ParseError{Cause: err}
	}

	if err := v.Validate(target); err != nil {
		return err
	}

	return nil
}

// ValidationError representa un field validation error.
type ValidationError struct {
	Field string `json:"field"`
	Tag   string `json:"tag"`
	Param string `json:"param,omitempty"`
	Msg   string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s failed on %s: %s", e.Field, e.Tag, e.Msg)
}

// ValidationErrors agrupa validation errors.
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	return fmt.Sprintf("validation failed: %d error(s)", len(ve))
}

// ValidationErrorsFrom convierte validator.ValidationErrors a nuestro tipo.
func ValidationErrorsFrom(errors validator.ValidationErrors) ValidationErrors {
	result := make(ValidationErrors, 0, len(errors))
	for _, err := range errors {
		result = append(result, ValidationError{
			Field: err.Field(),
			Tag:   err.Tag(),
			Param: err.Param(),
			Msg:   err.Error(),
		})
	}
	return result
}

// ParseError marca un fallo de parseo JSON.
type ParseError struct {
	Cause error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("invalid JSON: %v", e.Cause)
}

func (e *ParseError) Unwrap() error {
	return e.Cause
}
