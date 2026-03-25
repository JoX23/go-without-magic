package cmd

import (
	"fmt"

	"github.com/JoX23/go-without-magic/tools/codegen/typemap"
)

// List muestra la tabla de tipos soportados y los profiles disponibles.
func List() {
	fmt.Println("Tipos de campo soportados:")
	fmt.Println()
	fmt.Printf("  %-10s  %-15s  %-18s  %-10s\n", "YAML", "Go", "PostgreSQL", "Proto")
	fmt.Println("  " + repeat("─", 60))
	for _, t := range typemap.Table {
		goType := t.GoType
		if goType == "" {
			goType = "<EntityField>"
		}
		fmt.Printf("  %-10s  %-15s  %-18s  %-10s\n",
			t.YAMLType, goType, t.PostgresType, t.ProtoType)
	}

	fmt.Println()
	fmt.Println("Profiles disponibles:")
	fmt.Println()
	fmt.Printf("  %-15s  %s\n", "PROFILE", "CAPAS GENERADAS")
	fmt.Println("  " + repeat("─", 60))
	fmt.Printf("  %-15s  %s\n", "full (default)", "domain, service, http, memory, postgres, grpc")
	fmt.Printf("  %-15s  %s\n", "full-async", "domain, service, http, memory, postgres, grpc, kafka")
	fmt.Printf("  %-15s  %s\n", "api", "domain, service, http, memory")
	fmt.Printf("  %-15s  %s\n", "domain-only", "domain")
	fmt.Printf("  %-15s  %s\n", "no-grpc", "domain, service, http, memory, postgres")
	fmt.Println()
}

func repeat(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
