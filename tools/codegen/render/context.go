package render

import (
	"fmt"
	"strings"

	"github.com/JoX23/go-without-magic/tools/codegen/schema"
	"github.com/JoX23/go-without-magic/tools/codegen/typemap"
)

// TemplateData es la estructura que se pasa a todos los templates.
type TemplateData struct {
	Module   string       // "github.com/JoX23/go-without-magic"
	Entity   EntityData
	Generate schema.ResolvedGenerate
}

// EntityData contiene toda la información derivada de la entidad.
type EntityData struct {
	Name              string // "Product" (PascalCase)
	NameLower         string // "product"
	NameSnake         string // "product" (para columnas SQL y rutas)
	NamePlural        string // "products"
	Fields            []FieldData
	ConstructorFields []FieldData // solo campos requeridos y no-enum (para firma del constructor)
	LookupKeys        []LookupKeyData
	HasEnum           bool // true si algún campo es tipo enum
	HasUUID           bool // true si algún campo es tipo uuid (necesita import)
	HasTime           bool // true si algún campo es tipo time (necesita import)
}

// FieldData contiene la información de un campo ya procesada para templates.
type FieldData struct {
	Name         string // "SKU" (PascalCase)
	NameSnake    string // "sku"
	NameLower    string // "sku"
	GoType       string // "string", "*string", "uuid.UUID"
	GoBaseType   string // "string" (sin *)
	PostgresType string // "TEXT NOT NULL"
	ProtoType    string // "string"
	ProtoNum     int    // número de campo en proto (ID siempre es 1)
	ValidateTag  string // "required,min=3"
	IsUnique     bool
	IsOptional   bool
	IsEnum       bool
	EnumValues   []string
	EnumTypeName string // "ProductStatus"
}

// LookupKeyData describe una búsqueda secundaria.
type LookupKeyData struct {
	Field      FieldData
	MapName    string // "bySKU"
	MethodName string // "FindBySKU"
	// IsUnique determina si retorna (*Entity, error) o ([]*Entity, error)
	IsUnique bool
}

// Build construye el TemplateData a partir de un Schema y el module path.
func Build(s *schema.Schema, module string) (*TemplateData, error) {
	entity, err := buildEntity(s, module)
	if err != nil {
		return nil, err
	}

	return &TemplateData{
		Module:   module,
		Entity:   entity,
		Generate: s.ResolveGenerate(),
	}, nil
}

func buildEntity(s *schema.Schema, module string) (EntityData, error) {
	e := EntityData{
		Name:       s.Name,
		NameLower:  strings.ToLower(s.Name),
		NameSnake:  toSnake(s.Name),
		NamePlural: pluralize(strings.ToLower(s.Name)),
	}

	// Construir campos
	for i, f := range s.Fields {
		fd, err := buildField(f, s.Name, i+2) // proto num empieza en 2 (1 = id)
		if err != nil {
			return EntityData{}, fmt.Errorf("field %q: %w", f.Name, err)
		}
		e.Fields = append(e.Fields, fd)

		// Campos de constructor: requeridos y no-enum
		if !f.Optional && f.Type != "enum" {
			e.ConstructorFields = append(e.ConstructorFields, fd)
		}

		if f.Type == "enum" {
			e.HasEnum = true
		}
		if f.Type == "uuid" {
			e.HasUUID = true
		}
		if f.Type == "time" {
			e.HasTime = true
		}
	}

	// Construir lookup keys
	fieldMap := make(map[string]FieldData)
	for _, fd := range e.Fields {
		fieldMap[strings.ToLower(fd.Name)] = fd
	}

	for _, lk := range s.LookupKeys {
		fd, ok := fieldMap[strings.ToLower(lk.Field)]
		if !ok {
			return EntityData{}, fmt.Errorf("lookup_key %q: field not found", lk.Field)
		}
		e.LookupKeys = append(e.LookupKeys, LookupKeyData{
			Field:      fd,
			MapName:    "by" + fd.Name,
			MethodName: "FindBy" + fd.Name,
			IsUnique:   fd.IsUnique,
		})
	}

	_ = module
	return e, nil
}

func buildField(f schema.Field, entityName string, protoNum int) (FieldData, error) {
	tm, ok := typemap.Lookup(f.Type)
	if !ok {
		return FieldData{}, fmt.Errorf("unknown type %q", f.Type)
	}

	goType := tm.GoType
	if f.Type == "enum" {
		goType = entityName + f.Name // e.g. "ProductStatus"
	}

	postgresType := tm.PostgresType
	if f.Optional {
		postgresType += " NULL"
	} else {
		postgresType += " NOT NULL"
	}
	if f.Unique {
		postgresType += " UNIQUE"
	}

	if f.Optional && goType != "" {
		goType = "*" + goType
	}

	fd := FieldData{
		Name:         f.Name,
		NameSnake:    toSnake(f.Name),
		NameLower:    strings.ToLower(f.Name),
		GoType:       goType,
		GoBaseType:   tm.GoType,
		PostgresType: postgresType,
		ProtoType:    tm.ProtoType,
		ProtoNum:     protoNum,
		ValidateTag:  f.Validate,
		IsUnique:     f.Unique,
		IsOptional:   f.Optional,
		IsEnum:       f.Type == "enum",
		EnumValues:   f.Values,
		EnumTypeName: entityName + f.Name,
	}

	return fd, nil
}
