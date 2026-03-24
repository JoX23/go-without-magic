package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/JoX23/go-without-magic/tools/codegen/conflict"
	"github.com/JoX23/go-without-magic/tools/codegen/render"
	"github.com/JoX23/go-without-magic/tools/codegen/schema"
)

// Generate ejecuta el subcomando generate.
func Generate(args []string) error {
	fs := flag.NewFlagSet("generate", flag.ContinueOnError)
	schemaPath := fs.String("schema", "", "Path al archivo YAML de la entidad (requerido)")
	fs.StringVar(schemaPath, "s", "", "Path al archivo YAML de la entidad (shorthand)")
	outDir := fs.String("out", ".", "Directorio raíz del proyecto")
	fs.StringVar(outDir, "o", ".", "Directorio raíz del proyecto (shorthand)")
	dryRun := fs.Bool("dry-run", false, "Mostrar qué se generaría sin escribir nada")
	force := fs.Bool("force", false, "Sobreescribir archivos existentes")
	fs.BoolVar(force, "f", false, "Sobreescribir archivos existentes (shorthand)")
	backup := fs.Bool("backup", false, "Crear backup antes de sobreescribir (requiere --force)")
	verbose := fs.Bool("verbose", false, "Output detallado")
	fs.BoolVar(verbose, "v", false, "Output detallado (shorthand)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *schemaPath == "" {
		fs.Usage()
		return fmt.Errorf("--schema es requerido")
	}

	// Resolver rootDir a path absoluto
	root, err := filepath.Abs(*outDir)
	if err != nil {
		return fmt.Errorf("resolving output dir: %w", err)
	}

	// Leer el module path desde go.mod
	module, err := readModuleName(root)
	if err != nil {
		return fmt.Errorf("reading go.mod: %w", err)
	}

	fmt.Printf("[codegen] Schema:  %s\n", *schemaPath)
	fmt.Printf("[codegen] Root:    %s\n", root)
	fmt.Printf("[codegen] Module:  %s\n", module)

	// Cargar y validar schema
	s, err := schema.Load(*schemaPath)
	if err != nil {
		return fmt.Errorf("loading schema: %w", err)
	}

	fmt.Printf("[codegen] Entity:  %s (%d fields)\n\n", s.Name, len(s.Fields))

	// Construir template data
	data, err := render.Build(s, module)
	if err != nil {
		return fmt.Errorf("building template data: %w", err)
	}

	// Planificar archivos a generar
	files := render.Plan(data, root)

	// Resolver conflictos
	opts := conflict.Options{
		Force:  *force,
		Backup: *backup,
		DryRun: *dryRun,
	}
	resolved := conflict.Resolve(files, root, opts)

	// Ejecutar y reportar
	written, skipped := 0, 0
	var hints []render.OutputFile

	for _, rf := range resolved {
		if rf.Resolution == conflict.ResolutionHint {
			hints = append(hints, rf.OutputFile)
			continue
		}

		relPath := relativePath(rf.DestPath, root)
		label := rf.Resolution.Label()

		if rf.Resolution == conflict.ResolutionSkip {
			fmt.Printf("  %s  %s  (exists, use --force to overwrite)\n", label, relPath)
			skipped++
			continue
		}

		if rf.Resolution == conflict.ResolutionDryRun {
			fmt.Printf("  %s  %s\n", label, relPath)
			continue
		}

		if rf.Resolution == conflict.ResolutionBackup {
			if err := copyFile(rf.DestPath, rf.BackupPath); err != nil {
				return fmt.Errorf("creating backup: %w", err)
			}
			if *verbose {
				fmt.Printf("         backup → %s\n", rf.BackupPath)
			}
		}

		// Renderizar
		content, err := render.Render(data, rf.TemplatePath)
		if err != nil {
			return fmt.Errorf("rendering %s: %w", rf.TemplatePath, err)
		}

		// Escribir
		if err := render.WriteFile(rf.DestPath, content, rf.IsProto); err != nil {
			return fmt.Errorf("writing %s: %w", rf.DestPath, err)
		}

		fmt.Printf("  %s  %s\n", label, relPath)
		written++
	}

	fmt.Printf("\n[codegen] Generated: %d files", written)
	if skipped > 0 {
		fmt.Printf(", Skipped: %d (use --force to overwrite)", skipped)
	}
	fmt.Println()

	// Imprimir hints
	for _, h := range hints {
		content, err := render.Render(data, h.TemplatePath)
		if err != nil {
			continue
		}
		fmt.Println(content)
	}

	return nil
}

func readModuleName(rootDir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(rootDir, "go.mod"))
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	return "", fmt.Errorf("module line not found in go.mod")
}

func relativePath(abs, root string) string {
	rel := strings.TrimPrefix(abs, root)
	return strings.TrimPrefix(rel, string(os.PathSeparator))
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}
