package openapi

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenAPI_Generate(t *testing.T) {
	api := New("Go Without Magic", "v1.0.0", "Test API")

	api.AddPath("/users", http.MethodGet, Operation{
		Summary: "List users",
		Responses: map[string]Response{
			"200": {Description: "OK", Content: map[string]MediaType{"application/json": {Schema: Schema{"type": "array", "items": Schema{"$ref": "#/components/schemas/User"}}}}},
		},
	})

	api.AddSchema("User", Schema{
		"type": "object",
		"properties": map[string]Schema{
			"id":    {"type": "string"},
			"email": {"type": "string", "format": "email"},
		},
	})

	data, err := api.ToJSON()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed OpenAPI
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if parsed.Info.Title != "Go Without Magic" {
		t.Fatalf("unexpected title: %s", parsed.Info.Title)
	}
	if parsed.Paths["/users"]["get"].Summary != "List users" {
		t.Fatalf("unexpected summary")
	}
}

func TestOpenAPI_Handler(t *testing.T) {
	api := New("Go Without Magic", "v1.0.0", "Test API")
	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	rec := httptest.NewRecorder()

	api.Handler().ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", res.StatusCode)
	}
	if !strings.Contains(res.Header.Get("Content-Type"), "application/json") {
		t.Fatalf("unexpected content type: %s", res.Header.Get("Content-Type"))
	}

	_, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("read body failed: %v", err)
	}
}
