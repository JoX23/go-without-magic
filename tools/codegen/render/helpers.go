package render

import (
	"strings"
	"text/template"
	"unicode"
)

// HelperFuncs retorna las funciones disponibles en los templates.
func HelperFuncs() template.FuncMap {
	return template.FuncMap{
		"toLower":    strings.ToLower,
		"toUpper":    strings.ToUpper,
		"toPascal":   toPascal,
		"toSnake":    toSnake,
		"pluralize":  pluralize,
		"quote":      func(s string) string { return `"` + s + `"` },
		"trimPrefix": strings.TrimPrefix,
		"hasPrefix":  strings.HasPrefix,
		"join":       strings.Join,
		"add": func(a, b int) int { return a + b },
	}
}

// toPascal convierte snake_case o cualquier string a PascalCase.
// Si ya es PascalCase lo deja igual.
func toPascal(s string) string {
	if s == "" {
		return s
	}
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})
	var b strings.Builder
	for _, p := range parts {
		if len(p) == 0 {
			continue
		}
		runes := []rune(p)
		b.WriteRune(unicode.ToUpper(runes[0]))
		for _, r := range runes[1:] {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// toSnake convierte PascalCase o camelCase a snake_case.
func toSnake(s string) string {
	var b strings.Builder
	runes := []rune(s)
	for i, r := range runes {
		if unicode.IsUpper(r) && i > 0 {
			b.WriteRune('_')
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

// pluralize retorna la forma plural simple en inglés.
func pluralize(s string) string {
	lower := strings.ToLower(s)
	// Casos irregulares comunes
	irregulars := map[string]string{
		"person": "people",
		"child":  "children",
		"datum":  "data",
	}
	if p, ok := irregulars[lower]; ok {
		// Conservar capitalización del original
		if unicode.IsUpper([]rune(s)[0]) {
			return strings.ToUpper(p[:1]) + p[1:]
		}
		return p
	}

	// Reglas básicas
	switch {
	case strings.HasSuffix(lower, "s") || strings.HasSuffix(lower, "x") ||
		strings.HasSuffix(lower, "z") || strings.HasSuffix(lower, "ch") ||
		strings.HasSuffix(lower, "sh"):
		return s + "es"
	case strings.HasSuffix(lower, "y") && len(s) > 1 &&
		!isVowel(rune(lower[len(lower)-2])):
		return s[:len(s)-1] + "ies"
	default:
		return s + "s"
	}
}

func isVowel(r rune) bool {
	return strings.ContainsRune("aeiou", r)
}
