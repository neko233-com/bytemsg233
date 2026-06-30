package compiler

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/neko233-com/bytemsg233/pkg/codegen"
	_ "github.com/neko233-com/bytemsg233/pkg/codegen/csharp"
	_ "github.com/neko233-com/bytemsg233/pkg/codegen/go"
	_ "github.com/neko233-com/bytemsg233/pkg/codegen/java"
	_ "github.com/neko233-com/bytemsg233/pkg/codegen/python"
	_ "github.com/neko233-com/bytemsg233/pkg/codegen/rust"
	_ "github.com/neko233-com/bytemsg233/pkg/codegen/typescript"
	"github.com/neko233-com/bytemsg233/pkg/i18n"
	"github.com/neko233-com/bytemsg233/pkg/schema"
)

// Compiler compiles schema files supported by the schema import registry.
type Compiler struct{}

// New creates a new compiler
func New() *Compiler {
	return &Compiler{}
}

// CompileOptions contains options for compilation
type CompileOptions struct {
	InputFile string
	OutputDir string
	Languages []string
	Locale    string
}

// Compile compiles a schema file
func (c *Compiler) Compile(options *CompileOptions) error {
	s, err := schema.ParseFile(options.InputFile)
	if err != nil {
		return fmt.Errorf("failed to parse schema: %w", err)
	}

	if err := os.MkdirAll(options.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", options.OutputDir, err)
	}

	if options.Locale != "" {
		i18n.SetLocale(options.Locale)
	}

	for _, lang := range options.Languages {
		generator, err := codegen.Get(lang)
		if err != nil {
			return fmt.Errorf("generator not found for language %s: %w", lang, err)
		}

		genOptions := &codegen.GenerateOptions{
			OutputDir: options.OutputDir,
			Locale:    options.Locale,
		}

		files, err := generator.Generate(s, genOptions)
		if err != nil {
			return fmt.Errorf("generation failed for %s: %w", lang, err)
		}

		for _, file := range files {
			path := filepath.Join(options.OutputDir, file.Path)
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return fmt.Errorf("failed to create output directory for %s: %w", path, err)
			}
			if err := os.WriteFile(path, file.Content, 0644); err != nil {
				return fmt.Errorf("failed to write file %s: %w", path, err)
			}
		}
	}

	return nil
}
