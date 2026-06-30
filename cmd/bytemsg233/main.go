package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/neko233-com/bytemsg233/pkg/compiler"
	"github.com/neko233-com/bytemsg233/pkg/exporter"
	"github.com/neko233-com/bytemsg233/pkg/libinstall"
	"github.com/neko233-com/bytemsg233/pkg/schema"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "bytemsg233",
		Short: "bytemsg233 - A modern serialization framework",
		Long:  "bytemsg233 - A modern serialization framework that replaces Protocol Buffers",
	}

	var compileCmd = &cobra.Command{
		Use:   "compile [file]",
		Short: "Compile a schema to target languages",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			languages, _ := cmd.Flags().GetStringSlice("lang")
			outputDir, _ := cmd.Flags().GetString("output")
			locale, _ := cmd.Flags().GetString("locale")

			comp := compiler.New()
			return comp.Compile(&compiler.CompileOptions{
				InputFile: args[0],
				OutputDir: outputDir,
				Languages: languages,
				Locale:    locale,
			})
		},
	}

	compileCmd.Flags().StringSliceP("lang", "l", []string{"go"}, "Target languages (go, csharp, typescript, rust, java, python)")
	compileCmd.Flags().StringP("output", "o", ".", "Output directory")
	compileCmd.Flags().String("locale", "en", "Locale for comments (en, zh)")

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("bytemsg233 %s (commit: %s, built: %s)\n", version, commit, date)
		},
	}

	var initCmd = &cobra.Command{
		Use:   "init [name]",
		Short: "Initialize a new .bmsg.json file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			template := fmt.Sprintf(`{
  "schema": "bymsg/v1",
  "package": "%s",
  "enums": {
    "Status": ["ACTIVE", "INACTIVE"]
  },
  "Example": {
    "packetId": 1001,
    "comment": "Example message",
    "id": {
      "type": "uint32",
      "comment": "ID"
    },
    "name": {
      "type": "string",
      "comment": "Name"
    },
    "status": "Status",
    "tags": { "list": "string", "comment": "Tags" },
    "attrs": { "map": ["string", "string"], "comment": "Attributes" }
  }
}
`, name)

			filename := fmt.Sprintf("%s.bmsg.json", name)
			return os.WriteFile(filename, []byte(template), 0644)
		},
	}

	var exportCmd = &cobra.Command{
		Use:   "export [file]",
		Short: "Export protocol documentation and compatibility schema files",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			formats, _ := cmd.Flags().GetStringSlice("format")
			outputDir, _ := cmd.Flags().GetString("output")
			baseName, _ := cmd.Flags().GetString("name")
			if baseName == "" {
				baseName = strings.TrimSuffix(filepath.Base(args[0]), filepath.Ext(args[0]))
				baseName = strings.TrimSuffix(baseName, ".bmsg")
			}

			s, err := schema.ImportFile(args[0], nil)
			if err != nil {
				return err
			}
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				return err
			}

			for _, format := range formats {
				content, err := exporter.Export(format, s, &exporter.ExportOptions{Format: format, Name: baseName})
				if err != nil {
					return err
				}
				ext, err := exporter.Extension(format)
				if err != nil {
					return err
				}
				path := filepath.Join(outputDir, baseName+ext)
				if err := os.WriteFile(path, content, 0644); err != nil {
					return err
				}
			}
			return nil
		},
	}
	exportCmd.Flags().StringSliceP("format", "f", []string{"md", "html", "bmsg"}, "Export formats (md, html, bmsg, proto)")
	exportCmd.Flags().StringP("output", "o", ".", "Output directory")
	exportCmd.Flags().String("name", "", "Output base name")

	var installLibCmd = &cobra.Command{
		Use:   "install-lib [language]",
		Short: "Copy a language runtime library into your project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetDir, _ := cmd.Flags().GetString("to")
			repoRoot, _ := cmd.Flags().GetString("repo")
			if targetDir == "" {
				return fmt.Errorf("--to is required")
			}
			if repoRoot == "" {
				var err error
				repoRoot, err = os.Getwd()
				if err != nil {
					return err
				}
			}
			return libinstall.CopyLibrary(repoRoot, args[0], targetDir)
		},
	}
	installLibCmd.Flags().String("to", "", "Target directory in your project")
	installLibCmd.Flags().String("repo", "", "bytemsg233 repository root; defaults to current directory")

	rootCmd.AddCommand(compileCmd, versionCmd, initCmd, exportCmd, installLibCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
