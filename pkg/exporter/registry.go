package exporter

import (
	"fmt"
	"strings"

	"github.com/neko233-com/bytemsg233/pkg/schema"
)

type ExportOptions struct {
	Format string
	Name   string
}

type Exporter interface {
	Name() string
	Extensions() []string
	Export(s *schema.Schema, options *ExportOptions) ([]byte, error)
}

var (
	exportersByName      = make(map[string]Exporter)
	exportersByExtension = make(map[string]Exporter)
)

func RegisterExporter(exporter Exporter) {
	if exporter == nil {
		return
	}
	exportersByName[normalizeExportFormat(exporter.Name())] = exporter
	for _, ext := range exporter.Extensions() {
		exportersByExtension[normalizeExportFormat(ext)] = exporter
	}
}

func Export(format string, s *schema.Schema, options *ExportOptions) ([]byte, error) {
	if options != nil && options.Format != "" {
		format = options.Format
	}
	exporter, ok := exporterForFormat(format)
	if !ok {
		return nil, fmt.Errorf("unsupported export format %q", format)
	}
	return exporter.Export(s, options)
}

func Extension(format string) (string, error) {
	exporter, ok := exporterForFormat(format)
	if !ok {
		return "", fmt.Errorf("unsupported export format %q", format)
	}
	extensions := exporter.Extensions()
	if len(extensions) == 0 {
		return "." + normalizeExportFormat(exporter.Name()), nil
	}
	return extensions[0], nil
}

func exporterForFormat(format string) (Exporter, bool) {
	normalized := normalizeExportFormat(format)
	if normalized == "" {
		return nil, false
	}
	if exporter, ok := exportersByName[normalized]; ok {
		return exporter, true
	}
	exporter, ok := exportersByExtension[normalized]
	return exporter, ok
}

func normalizeExportFormat(format string) string {
	format = strings.TrimSpace(strings.ToLower(format))
	format = strings.TrimPrefix(format, ".")
	return format
}

type markdownExporter struct{}

func (markdownExporter) Name() string { return "md" }
func (markdownExporter) Extensions() []string {
	return []string{".md", ".markdown"}
}
func (markdownExporter) Export(s *schema.Schema, _ *ExportOptions) ([]byte, error) {
	return Markdown(s), nil
}

type htmlExporter struct{}

func (htmlExporter) Name() string { return "html" }
func (htmlExporter) Extensions() []string {
	return []string{".html"}
}
func (htmlExporter) Export(s *schema.Schema, _ *ExportOptions) ([]byte, error) {
	return HTML(s), nil
}

type bmsgExporter struct{}

func (bmsgExporter) Name() string { return "bmsg" }
func (bmsgExporter) Extensions() []string {
	return []string{".bmsg"}
}
func (bmsgExporter) Export(s *schema.Schema, _ *ExportOptions) ([]byte, error) {
	return Bmsg(s), nil
}

func init() {
	RegisterExporter(markdownExporter{})
	RegisterExporter(htmlExporter{})
	RegisterExporter(bmsgExporter{})
}
