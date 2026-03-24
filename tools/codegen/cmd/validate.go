package cmd

import (
	"flag"
	"fmt"

	"github.com/JoX23/go-without-magic/tools/codegen/schema"
)

// Validate ejecuta el subcomando validate.
func Validate(args []string) error {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	schemaPath := fs.String("schema", "", "Path al archivo YAML (requerido)")
	fs.StringVar(schemaPath, "s", "", "Path al archivo YAML (shorthand)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *schemaPath == "" {
		return fmt.Errorf("--schema es requerido")
	}

	s, err := schema.Load(*schemaPath)
	if err != nil {
		return fmt.Errorf("❌  %w", err)
	}

	gen := s.ResolveGenerate()

	fmt.Printf("✅  Schema válido\n")
	fmt.Printf("    Entidad:  %s\n", s.Name)
	fmt.Printf("    Campos:   %d\n", len(s.Fields))
	fmt.Printf("    Profile:  %s\n", profileName(s.Profile))
	fmt.Printf("    Capas:    ")

	layers := []string{}
	if gen.Domain {
		layers = append(layers, "domain")
	}
	if gen.Service {
		layers = append(layers, "service")
	}
	if gen.HTTPHandler {
		layers = append(layers, "http")
	}
	if gen.MemoryRepository {
		layers = append(layers, "memory")
	}
	if gen.PostgresRepository {
		layers = append(layers, "postgres")
	}
	if gen.GRPC {
		layers = append(layers, "grpc")
	}

	for i, l := range layers {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(l)
	}
	fmt.Println()

	return nil
}

func profileName(p string) string {
	if p == "" {
		return "full (default)"
	}
	return p
}
