package schema

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Load lee un archivo YAML y retorna el Schema validado.
func Load(path string) (*Schema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading schema file: %w", err)
	}

	var s Schema
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	if err := Validate(&s); err != nil {
		return nil, fmt.Errorf("invalid schema: %w", err)
	}

	return &s, nil
}
