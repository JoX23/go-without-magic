package validator

import (
	"encoding/json"
	"net/http"
)

// AppError es la forma estandarizada para errores HTTP.
type AppError struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func (e AppError) Error() string {
	if e.Message == "" {
		return e.Code
	}
	return e.Message
}

// MapToHTTP traduce errores de dominio / validación a AppError.
func MapToHTTP(err error) *AppError {
	if err == nil {
		return nil
	}
	switch e := err.(type) {
	case *ParseError:
		return &AppError{Status: http.StatusBadRequest, Code: "invalid_payload", Message: e.Error()}
	case ValidationErrors:
		return &AppError{Status: http.StatusBadRequest, Code: "validation_error", Message: "validation failed", Details: e}
	case AppError:
		return &e
	case *AppError:
		return e
	case DomainError:
		return &AppError{Status: e.StatusCode(), Code: "domain_error", Message: e.Error()}
	default:
		return &AppError{Status: http.StatusInternalServerError, Code: "internal_error", Message: "internal server error"}
	}
}

// WriteHTTP writes la respuesta HTTP del AppError.
func WriteHTTP(w http.ResponseWriter, err *AppError) {
	if err == nil {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Status)
	json.NewEncoder(w).Encode(err)
}

// DomainError define errores de dominio con código HTTP propio.
type DomainError interface {
	error
	StatusCode() int
}
