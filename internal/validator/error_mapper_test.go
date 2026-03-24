package validator

import (
	"errors"
	"net/http"
	"testing"
)

type customDomainError struct{}

func (customDomainError) Error() string   { return "custom domain issue" }
func (customDomainError) StatusCode() int { return http.StatusConflict }

func TestMapToHTTP_ValidationError(t *testing.T) {
	ve := ValidationErrors{{Field: "Email", Tag: "email", Msg: "invalid"}}
	app := MapToHTTP(ve)

	if app.Status != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", app.Status)
	}
	if app.Code != "validation_error" {
		t.Fatalf("expected validation_error, got %s", app.Code)
	}
}

func TestMapToHTTP_ParseError(t *testing.T) {
	app := MapToHTTP(&ParseError{Cause: errors.New("syntax")})
	if app.Status != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", app.Status)
	}
}

func TestMapToHTTP_DomainError(t *testing.T) {
	app := MapToHTTP(customDomainError{})
	if app.Status != http.StatusConflict {
		t.Fatalf("expected 409, got %d", app.Status)
	}
}

func TestMapToHTTP_UnknownError(t *testing.T) {
	app := MapToHTTP(errors.New("foo"))
	if app.Status != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", app.Status)
	}
}
