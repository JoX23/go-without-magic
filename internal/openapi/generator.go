package openapi

import (
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
)

// OpenAPI representa un spec base compatible con OpenAPI 3.0.
type OpenAPI struct {
    Openapi    string              `json:"openapi"`
    Info       Info                `json:"info"`
    Paths      map[string]PathItem `json:"paths"`
    Components Components          `json:"components,omitempty"`
}

// Info describe la metadata de la API.
type Info struct {
    Title       string `json:"title"`
    Version     string `json:"version"`
    Description string `json:"description,omitempty"`
}

// PathItem es un mapa por método HTTP.
type PathItem map[string]Operation

// Operation llena la operación de un endpoint.
type Operation struct {
    Summary     string              `json:"summary,omitempty"`
    Description string              `json:"description,omitempty"`
    Tags        []string            `json:"tags,omitempty"`
    Parameters  []Parameter         `json:"parameters,omitempty"`
    RequestBody *RequestBody       `json:"requestBody,omitempty"`
    Responses   map[string]Response `json:"responses"`
}

// Parameter describe un parámetro de entrada.
type Parameter struct {
    Name        string `json:"name"`
    In          string `json:"in"`
    Required    bool   `json:"required"`
    Description string `json:"description,omitempty"`
    Schema      Schema `json:"schema"`
}

// RequestBody describe el cuerpo de la petición.
type RequestBody struct {
    Description string               `json:"description,omitempty"`
    Required    bool                 `json:"required"`
    Content     map[string]MediaType `json:"content"`
}

// Response representa la respuesta de un endpoint.
type Response struct {
    Description string               `json:"description"`
    Content     map[string]MediaType `json:"content,omitempty"`
}

// MediaType representa un tipo de contenido.
type MediaType struct {
    Schema Schema `json:"schema"`
}

// Components simplifica la sección components.
type Components struct {
    Schemas map[string]Schema `json:"schemas,omitempty"`
}

// Schema permite una representación flexible de objetos.
type Schema map[string]interface{}

// New crea un spec base OpenAPI 3.0.
func New(title, version, description string) *OpenAPI {
    return &OpenAPI{
        Openapi: "3.0.0",
        Info: Info{
            Title:       title,
            Version:     version,
            Description: description,
        },
        Paths: make(map[string]PathItem),
        Components: Components{
            Schemas: make(map[string]Schema),
        },
    }
}

// AddPath añade un endpoint al spec.
func (o *OpenAPI) AddPath(path, method string, op Operation) {
    method = strings.ToLower(method)
    if o.Paths[path] == nil {
        o.Paths[path] = make(PathItem)
    }
    o.Paths[path][method] = op
}

// AddSchema añade una definición de schema reusable.
func (o *OpenAPI) AddSchema(name string, schema Schema) {
    if o.Components.Schemas == nil {
        o.Components.Schemas = make(map[string]Schema)
    }
    o.Components.Schemas[name] = schema
}

// ToJSON serializa el spec a JSON con indent.
func (o *OpenAPI) ToJSON() ([]byte, error) {
    return json.MarshalIndent(o, "", "  ")
}

// Handler devuelve un endpoint HTTP para servir el spec.
func (o *OpenAPI) Handler() http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        bytes, err := o.ToJSON()
        if err != nil {
            http.Error(w, fmt.Sprintf("failed to serialize OpenAPI spec: %v", err), http.StatusInternalServerError)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write(bytes)
    })
}
