package conflict

import (
	"fmt"
	"os"
	"time"

	"github.com/JoX23/go-without-magic/tools/codegen/render"
)

// Resolution indica qué acción tomar con un archivo.
type Resolution int

const (
	ResolutionWrite     Resolution = iota // Archivo nuevo
	ResolutionSkip                        // Existe y --force=false
	ResolutionOverwrite                   // Existe y --force=true
	ResolutionBackup                      // Existe y --force --backup
	ResolutionDryRun                      // Solo reportar
	ResolutionHint                        // Solo stdout
)

// ResolvedFile extiende OutputFile con la resolución tomada.
type ResolvedFile struct {
	render.OutputFile
	Resolution Resolution
	BackupPath string
}

// Options configura el comportamiento del resolver.
type Options struct {
	Force   bool
	Backup  bool
	DryRun  bool
}

// protectedPaths son archivos que nunca se sobreescriben.
var protectedPaths = map[string]bool{
	"cmd/server/main.go":              true,
	"internal/grpc/error_codes.go":    true,
}

// Resolve toma la lista de archivos planeados y decide qué hacer con cada uno.
func Resolve(files []render.OutputFile, rootDir string, opts Options) []ResolvedFile {
	resolved := make([]ResolvedFile, 0, len(files))

	for _, f := range files {
		rf := ResolvedFile{OutputFile: f}

		if f.IsHint {
			rf.Resolution = ResolutionHint
			resolved = append(resolved, rf)
			continue
		}

		if opts.DryRun {
			rf.Resolution = ResolutionDryRun
			resolved = append(resolved, rf)
			continue
		}

		// Verificar si es un archivo protegido
		relPath := relativeToRoot(f.DestPath, rootDir)
		if protectedPaths[relPath] {
			rf.Resolution = ResolutionSkip
			resolved = append(resolved, rf)
			continue
		}

		// Verificar si existe
		_, err := os.Stat(f.DestPath)
		exists := err == nil

		switch {
		case !exists:
			rf.Resolution = ResolutionWrite
		case opts.Force && opts.Backup:
			rf.Resolution = ResolutionBackup
			rf.BackupPath = f.DestPath + ".bak." + time.Now().Format("20060102T150405")
		case opts.Force:
			rf.Resolution = ResolutionOverwrite
		default:
			rf.Resolution = ResolutionSkip
		}

		resolved = append(resolved, rf)
	}

	return resolved
}

// Apply ejecuta las resoluciones: hace backups, escribe archivos.
func Apply(resolved []ResolvedFile, data interface{ GetContent(string) (string, bool) }) error {
	for _, rf := range resolved {
		if rf.Resolution == ResolutionSkip ||
			rf.Resolution == ResolutionDryRun ||
			rf.Resolution == ResolutionHint {
			continue
		}

		if rf.Resolution == ResolutionBackup {
			if err := copyFile(rf.DestPath, rf.BackupPath); err != nil {
				return fmt.Errorf("creating backup %s: %w", rf.BackupPath, err)
			}
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, content, 0o644)
}

func relativeToRoot(absPath, rootDir string) string {
	if len(absPath) > len(rootDir)+1 {
		return absPath[len(rootDir)+1:]
	}
	return absPath
}

// Label retorna un string legible de la resolución para el output de consola.
func (r Resolution) Label() string {
	switch r {
	case ResolutionWrite:
		return "WRITE"
	case ResolutionSkip:
		return "SKIP "
	case ResolutionOverwrite:
		return "OVER "
	case ResolutionBackup:
		return "BACK "
	case ResolutionDryRun:
		return "DRY  "
	case ResolutionHint:
		return "HINT "
	default:
		return "???  "
	}
}
