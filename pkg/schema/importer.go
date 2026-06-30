package schema

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ImportOptions struct {
	Format string
}

type Importer interface {
	Name() string
	Extensions() []string
	Import(data []byte, options *ImportOptions) (*Schema, error)
}

var (
	importersByName      = make(map[string]Importer)
	importersByExtension = make(map[string]Importer)
)

func RegisterImporter(importer Importer) {
	if importer == nil {
		return
	}
	importersByName[normalizeImportFormat(importer.Name())] = importer
	for _, ext := range importer.Extensions() {
		importersByExtension[normalizeImportFormat(ext)] = importer
	}
}

func Import(format string, data []byte, options *ImportOptions) (*Schema, error) {
	if options != nil && options.Format != "" {
		format = options.Format
	}
	if format != "" {
		importer, ok := importerForFormat(format)
		if !ok {
			return nil, fmt.Errorf("unsupported schema import format %q", format)
		}
		return importer.Import(data, options)
	}
	return importWithFallback(data, options)
}

func ImportFile(path string, options *ImportOptions) (*Schema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	if options != nil && options.Format != "" {
		return Import(options.Format, data, options)
	}

	ext := filepath.Ext(path)
	if importer, ok := importerForFormat(ext); ok {
		return importer.Import(data, options)
	}
	return importWithFallback(data, options)
}

func importerForFormat(format string) (Importer, bool) {
	normalized := normalizeImportFormat(format)
	if normalized == "" {
		return nil, false
	}
	if importer, ok := importersByName[normalized]; ok {
		return importer, true
	}
	importer, ok := importersByExtension[normalized]
	return importer, ok
}

func importWithFallback(data []byte, options *ImportOptions) (*Schema, error) {
	var lastErr error
	for _, format := range []string{"json", "yaml", "bmsg"} {
		importer, ok := importersByName[format]
		if !ok {
			continue
		}
		s, err := importer.Import(data, options)
		if err == nil {
			return s, nil
		}
		lastErr = err
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("unsupported schema import data: expected json, yaml, or bmsg")
}

func normalizeImportFormat(format string) string {
	format = strings.TrimSpace(strings.ToLower(format))
	format = strings.TrimPrefix(format, ".")
	return format
}

type jsonImporter struct{}

func (jsonImporter) Name() string { return "json" }
func (jsonImporter) Extensions() []string {
	return []string{".json"}
}
func (jsonImporter) Import(data []byte, _ *ImportOptions) (*Schema, error) {
	return ParseJSON(data)
}

type yamlImporter struct{}

func (yamlImporter) Name() string { return "yaml" }
func (yamlImporter) Extensions() []string {
	return []string{".yaml", ".yml"}
}
func (yamlImporter) Import(data []byte, _ *ImportOptions) (*Schema, error) {
	return ParseYAML(data)
}

type bmsgImporter struct{}

func (bmsgImporter) Name() string { return "bmsg" }
func (bmsgImporter) Extensions() []string {
	return []string{".bmsg"}
}
func (bmsgImporter) Import(data []byte, options *ImportOptions) (*Schema, error) {
	for _, format := range []string{"json", "yaml"} {
		importer, ok := importersByName[format]
		if !ok {
			continue
		}
		s, err := importer.Import(data, options)
		if err == nil {
			return s, nil
		}
	}
	return ParseBmsg(data)
}

type tomlImporter struct{}

func (tomlImporter) Name() string { return "toml" }
func (tomlImporter) Extensions() []string {
	return []string{".toml"}
}
func (tomlImporter) Import(data []byte, _ *ImportOptions) (*Schema, error) {
	return ParseTOML(data)
}

func init() {
	RegisterImporter(jsonImporter{})
	RegisterImporter(yamlImporter{})
	RegisterImporter(bmsgImporter{})
	RegisterImporter(tomlImporter{})
}
