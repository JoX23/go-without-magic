package validator

import (
	"net/http"
	"strings"
	"testing"
)

type userRequest struct {
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"required,gte=18"`
}

func TestValidate_Success(t *testing.T) {
	v := New()
	req := userRequest{Email: "test@example.com", Age: 25}
	if err := v.Validate(req); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidate_Fail(t *testing.T) {
	v := New()
	req := userRequest{Email: "bad-email", Age: 16}
	err := v.Validate(req)
	if err == nil {
		t.Fatal("expected validation error")
	}

	ve, ok := err.(ValidationErrors)
	if !ok {
		t.Fatalf("expected ValidationErrors type, got %T", err)
	}
	if len(ve) != 2 {
		t.Fatalf("expected 2 validation errors, got %d", len(ve))
	}
}

func TestParseJSON_Success(t *testing.T) {
	v := New()
	body := `{"email":"test@example.com","age":30}`
	req, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	var p userRequest

	if err := v.ParseJSON(req, &p); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p.Email != "test@example.com" || p.Age != 30 {
		t.Fatalf("unexpected parsed object: %+v", p)
	}
}

func TestParseJSON_InvalidJson(t *testing.T) {
	v := New()
	req, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(`{"email":"bad@`))
	var p userRequest

	err := v.ParseJSON(req, &p)
	if err == nil {
		t.Fatal("expected parse error")
	}
}
