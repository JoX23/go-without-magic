package render

import (
	"bytes"
	"fmt"
	"go/format"
	"io/fs"
	"os"
	"path/filepath"
	"text/template"
)

// templatesFS es el FS de templates inyectado desde main via SetFS.
var templatesFS fs.FS

// SetFS inyecta el FS de templates embebidos (llamado desde main).
func SetFS(f fs.FS) {
	templatesFS = f
}

// OutputFile describe un archivo a generar.
type OutputFile struct {
	TemplatePath string // path dentro del FS, e.g. "templates/domain/entity.go.tmpl"
	DestPath     string // path absoluto de destino
	IsHint       bool   // si es true, se imprime en stdout en vez de escribir
	IsProto      bool   // si es true, no pasar por go/format
}

// Plan retorna la lista de archivos a generar según las flags resueltas.
func Plan(data *TemplateData, rootDir string) []OutputFile {
	e := data.Entity
	g := data.Generate
	var files []OutputFile

	add := func(tmpl, dest string) {
		files = append(files, OutputFile{
			TemplatePath: "templates/" + tmpl,
			DestPath:     filepath.Join(rootDir, dest),
		})
	}
	addProto := func(tmpl, dest string) {
		files = append(files, OutputFile{
			TemplatePath: "templates/" + tmpl,
			DestPath:     filepath.Join(rootDir, dest),
			IsProto:      true,
		})
	}
	addHint := func(tmpl string) {
		files = append(files, OutputFile{
			TemplatePath: "templates/" + tmpl,
			IsHint:       true,
		})
	}

	if g.Domain {
		add("domain/entity.go.tmpl", fmt.Sprintf("internal/domain/%s_entity.go", e.NameLower))
		add("domain/errors.go.tmpl", fmt.Sprintf("internal/domain/%s_errors.go", e.NameLower))
		add("domain/repository.go.tmpl", fmt.Sprintf("internal/domain/%s_repository.go", e.NameLower))
	}
	if g.Service {
		add("service/service.go.tmpl", fmt.Sprintf("internal/service/%s_service.go", e.NameLower))
	}
	if g.HTTPHandler {
		add("handler/http_handler.go.tmpl", fmt.Sprintf("internal/handler/http/%s_handler.go", e.NameLower))
	}
	if g.MemoryRepository {
		add("repository/memory_repository.go.tmpl", fmt.Sprintf("internal/repository/memory/%s_repository.go", e.NameLower))
	}
	if g.PostgresRepository {
		add("repository/postgres_repository.go.tmpl", fmt.Sprintf("internal/repository/postgres/%s_repository.go", e.NameLower))
	}
	if g.GRPC {
		addProto("grpc/proto.proto.tmpl", fmt.Sprintf("internal/grpc/proto/%s.proto", e.NameLower))
		add("grpc/grpc_service.go.tmpl", fmt.Sprintf("internal/grpc/%s_service.go", e.NameLower))
	}
	if g.KafkaHandler {
		add("kafka/kafka_handler.go.tmpl", fmt.Sprintf("internal/kafka/handler/%s_handler.go", e.NameLower))
		addHint("hints/kafka_wiring.txt.tmpl")
	}

	addHint("hints/main_wiring.txt.tmpl")

	return files
}

// Render ejecuta un template y retorna el resultado como string.
func Render(data *TemplateData, tmplPath string) (string, error) {
	if templatesFS == nil {
		return "", fmt.Errorf("templates FS not initialized: call render.SetFS() first")
	}

	content, err := fs.ReadFile(templatesFS, tmplPath)
	if err != nil {
		return "", fmt.Errorf("reading template %q: %w", tmplPath, err)
	}

	tmpl, err := template.New(tmplPath).Funcs(HelperFuncs()).Parse(string(content))
	if err != nil {
		return "", fmt.Errorf("parsing template %q: %w", tmplPath, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("executing template %q: %w", tmplPath, err)
	}

	return buf.String(), nil
}

// WriteFile escribe el contenido en destPath, aplicando gofmt si corresponde.
func WriteFile(destPath, content string, isProto bool) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("creating directories: %w", err)
	}

	var out []byte
	if !isProto {
		formatted, err := format.Source([]byte(content))
		if err != nil {
			// Si falla gofmt, escribir sin formatear para facilitar depuración
			out = []byte(content)
		} else {
			out = formatted
		}
	} else {
		out = []byte(content)
	}

	if err := os.WriteFile(destPath, out, 0o644); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}
	return nil
}
