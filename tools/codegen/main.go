// codegen es el generador de código de go-without-magic.
//
// Genera boilerplate de Clean Architecture a partir de un schema YAML.
//
// Uso:
//
//	go run ./tools/codegen/ generate --schema tools/codegen/examples/product.yaml
//	go run ./tools/codegen/ validate --schema tools/codegen/examples/product.yaml
//	go run ./tools/codegen/ list
package main

import (
	"fmt"
	"os"

	"github.com/JoX23/go-without-magic/tools/codegen/cmd"
	"github.com/JoX23/go-without-magic/tools/codegen/render"
)

func main() {
	// Inyectar el FS de templates embebidos al renderer
	render.SetFS(templatesFS)

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	subcommand := os.Args[1]
	args := os.Args[2:]

	var err error
	switch subcommand {
	case "generate", "gen":
		err = cmd.Generate(args)
	case "validate", "val":
		err = cmd.Validate(args)
	case "list":
		cmd.List()
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "subcomando desconocido: %q\n\n", subcommand)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "\n[codegen] error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(`
go-without-magic code generator

Uso:
  go run ./tools/codegen/ <subcomando> [flags]

Subcomandos:
  generate  --schema <file> [--out <dir>] [--dry-run] [--force] [--backup]
            Genera código a partir de un schema YAML

  validate  --schema <file>
            Valida el schema YAML sin generar nada

  list      Muestra tipos soportados y profiles disponibles

Ejemplos:
  go run ./tools/codegen/ generate --schema tools/codegen/examples/product.yaml
  go run ./tools/codegen/ generate --schema product.yaml --dry-run
  go run ./tools/codegen/ generate --schema product.yaml --force --backup
  go run ./tools/codegen/ validate --schema product.yaml
  go run ./tools/codegen/ list

`)
}
