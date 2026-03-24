package schema

import (
	"fmt"
	"regexp"
	"strings"
)

var pascalCaseRe = regexp.MustCompile(`^[A-Z][a-zA-Z0-9]*$`)

// SupportedTypes lista los tipos de campo válidos.
var SupportedTypes = []string{
	"string", "int", "int64", "float64", "bool", "uuid", "time", "enum",
}

// Validate verifica las reglas semánticas del schema.
func Validate(s *Schema) error {
	if s.Name == "" {
		return fmt.Errorf("name is required")
	}
	if !pascalCaseRe.MatchString(s.Name) {
		return fmt.Errorf("name %q must be PascalCase (e.g. Product, OrderItem)", s.Name)
	}

	if s.Profile != "" {
		validProfiles := []string{ProfileFull, ProfileAPI, ProfileDomainOnly, ProfileNoGRPC}
		valid := false
		for _, p := range validProfiles {
			if s.Profile == p {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("unknown profile %q, valid: %s", s.Profile, strings.Join(validProfiles, ", "))
		}
	}

	if len(s.Fields) == 0 {
		return fmt.Errorf("at least one field is required")
	}

	seen := map[string]bool{}
	supportedSet := map[string]bool{}
	for _, t := range SupportedTypes {
		supportedSet[t] = true
	}

	for i, f := range s.Fields {
		if f.Name == "" {
			return fmt.Errorf("field[%d]: name is required", i)
		}
		if !pascalCaseRe.MatchString(f.Name) {
			return fmt.Errorf("field %q must be PascalCase", f.Name)
		}
		lower := strings.ToLower(f.Name)
		if seen[lower] {
			return fmt.Errorf("duplicate field name: %q", f.Name)
		}
		seen[lower] = true

		if !supportedSet[f.Type] {
			return fmt.Errorf("field %q: unsupported type %q, valid: %s",
				f.Name, f.Type, strings.Join(SupportedTypes, ", "))
		}
		if f.Type == "enum" && len(f.Values) == 0 {
			return fmt.Errorf("field %q is type enum but has no values", f.Name)
		}
	}

	for _, lk := range s.LookupKeys {
		if !seen[strings.ToLower(lk.Field)] {
			return fmt.Errorf("lookup_key %q does not reference a declared field", lk.Field)
		}
	}

	return nil
}
