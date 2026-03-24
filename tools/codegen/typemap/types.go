package typemap

// TypeMapping define la correspondencia entre el tipo declarado en YAML
// y los tipos concretos en cada capa del stack.
type TypeMapping struct {
	YAMLType     string
	GoType       string // tipo base sin *
	GoImport     string // import package path, "" si no necesita
	PostgresType string
	ProtoType    string
}

// Table es la tabla completa de tipos soportados.
var Table = []TypeMapping{
	{
		YAMLType:     "string",
		GoType:       "string",
		PostgresType: "TEXT",
		ProtoType:    "string",
	},
	{
		YAMLType:     "int",
		GoType:       "int",
		PostgresType: "INTEGER",
		ProtoType:    "int32",
	},
	{
		YAMLType:     "int64",
		GoType:       "int64",
		PostgresType: "BIGINT",
		ProtoType:    "int64",
	},
	{
		YAMLType:     "float64",
		GoType:       "float64",
		PostgresType: "NUMERIC(12,4)",
		ProtoType:    "double",
	},
	{
		YAMLType:     "bool",
		GoType:       "bool",
		PostgresType: "BOOLEAN",
		ProtoType:    "bool",
	},
	{
		YAMLType:     "uuid",
		GoType:       "uuid.UUID",
		GoImport:     "github.com/google/uuid",
		PostgresType: "UUID",
		ProtoType:    "string",
	},
	{
		YAMLType:     "time",
		GoType:       "time.Time",
		GoImport:     "time",
		PostgresType: "TIMESTAMPTZ",
		ProtoType:    "string",
	},
	{
		YAMLType:     "enum",
		GoType:       "", // se genera en el template como <Entity><Field>
		PostgresType: "TEXT",
		ProtoType:    "string",
	},
}

// index para búsqueda rápida
var index = func() map[string]TypeMapping {
	m := make(map[string]TypeMapping, len(Table))
	for _, t := range Table {
		m[t.YAMLType] = t
	}
	return m
}()

// Lookup retorna el TypeMapping para un tipo YAML dado.
func Lookup(yamlType string) (TypeMapping, bool) {
	t, ok := index[yamlType]
	return t, ok
}
