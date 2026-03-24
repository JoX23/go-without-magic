package schema

// Schema representa la definición de una entidad leída desde el YAML.
type Schema struct {
	Version    string         `yaml:"version"`
	Name       string         `yaml:"name"`
	Profile    string         `yaml:"profile"`
	Generate   GenerateFlags  `yaml:"generate"`
	Fields     []Field        `yaml:"fields"`
	LookupKeys []LookupKey    `yaml:"lookup_keys"`
}

// GenerateFlags controla qué capas se generan.
// Todos los campos son true por defecto si no se especifica.
type GenerateFlags struct {
	Domain             *bool `yaml:"domain"`
	Service            *bool `yaml:"service"`
	HTTPHandler        *bool `yaml:"http_handler"`
	MemoryRepository   *bool `yaml:"memory_repository"`
	PostgresRepository *bool `yaml:"postgres_repository"`
	GRPC               *bool `yaml:"grpc"`
}

// Field describe un campo de la entidad.
type Field struct {
	Name     string   `yaml:"name"`
	Type     string   `yaml:"type"`
	Validate string   `yaml:"validate"`
	Unique   bool     `yaml:"unique"`
	Optional bool     `yaml:"optional"`
	Values   []string `yaml:"values"` // solo para type: enum
}

// LookupKey describe una búsqueda secundaria a generar en el repositorio.
type LookupKey struct {
	Field string `yaml:"field"`
}

// Profiles predefinidos.
const (
	ProfileFull       = "full"
	ProfileAPI        = "api"
	ProfileDomainOnly = "domain-only"
	ProfileNoGRPC     = "no-grpc"
)

// ResolveGenerate resuelve las flags de generación según el profile y
// los overrides explícitos en la sección generate:.
func (s *Schema) ResolveGenerate() ResolvedGenerate {
	profile := s.Profile
	if profile == "" {
		profile = ProfileFull
	}

	// Valores base según profile
	r := baseForProfile(profile)

	// Overrides explícitos del YAML
	g := s.Generate
	if g.Domain != nil {
		r.Domain = *g.Domain
	}
	if g.Service != nil {
		r.Service = *g.Service
	}
	if g.HTTPHandler != nil {
		r.HTTPHandler = *g.HTTPHandler
	}
	if g.MemoryRepository != nil {
		r.MemoryRepository = *g.MemoryRepository
	}
	if g.PostgresRepository != nil {
		r.PostgresRepository = *g.PostgresRepository
	}
	if g.GRPC != nil {
		r.GRPC = *g.GRPC
	}

	return r
}

// ResolvedGenerate contiene las flags ya resueltas (sin punteros).
type ResolvedGenerate struct {
	Domain             bool
	Service            bool
	HTTPHandler        bool
	MemoryRepository   bool
	PostgresRepository bool
	GRPC               bool
}

func baseForProfile(profile string) ResolvedGenerate {
	switch profile {
	case ProfileAPI:
		return ResolvedGenerate{
			Domain: true, Service: true, HTTPHandler: true,
			MemoryRepository: true,
		}
	case ProfileDomainOnly:
		return ResolvedGenerate{Domain: true}
	case ProfileNoGRPC:
		return ResolvedGenerate{
			Domain: true, Service: true, HTTPHandler: true,
			MemoryRepository: true, PostgresRepository: true,
		}
	default: // full
		return ResolvedGenerate{
			Domain: true, Service: true, HTTPHandler: true,
			MemoryRepository: true, PostgresRepository: true, GRPC: true,
		}
	}
}
