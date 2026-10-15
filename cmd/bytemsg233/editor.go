package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/neko233-com/bytemsg233/pkg/codegen"
	"github.com/neko233-com/bytemsg233/pkg/schema"
	"github.com/spf13/cobra"
)

// Editor commands use stdin/stdout so IDEs can work on unsaved buffers without
// temporary files or shell interpolation. Generation remains the compile command.
func editorCommands() []*cobra.Command {
	read := func(cmd *cobra.Command, args []string) ([]byte, string, error) {
		stdin, _ := cmd.Flags().GetBool("stdin")
		name := args[0]
		if stdin {
			data, err := io.ReadAll(cmd.InOrStdin())
			return data, name, err
		}
		data, err := os.ReadFile(name)
		return data, name, err
	}
	check := &cobra.Command{Use: "check [file]", Short: "Validate a schema without generating files", Args: cobra.ExactArgs(1)}
	check.Flags().Bool("stdin", false, "Read the editor buffer from stdin; file determines its format")
	check.RunE = func(cmd *cobra.Command, args []string) error {
		data, name, err := read(cmd, args)
		if err != nil {
			return err
		}
		_, err = schema.Import(filepath.Ext(name), data, nil)
		return err
	}
	format := &cobra.Command{Use: "format [file]", Short: "Format JSON, YAML or legacy .bmsg without reordering fields", Args: cobra.ExactArgs(1)}
	format.Flags().Bool("stdin", false, "Read the editor buffer from stdin; file determines its format")
	format.Flags().BoolP("write", "w", false, "Write the formatted file in place (requires a disk file)")
	format.Flags().Int("indent", 2, "Indent width (1-8 spaces)")
	format.RunE = func(cmd *cobra.Command, args []string) error {
		write, _ := cmd.Flags().GetBool("write")
		stdin, _ := cmd.Flags().GetBool("stdin")
		if write && stdin {
			return fmt.Errorf("--write cannot be combined with --stdin")
		}
		data, name, err := read(cmd, args)
		if err != nil {
			return err
		}
		indent, _ := cmd.Flags().GetInt("indent")
		formatted, err := schema.Format(name, data, indent)
		if err != nil {
			return err
		}
		if write {
			info, err := os.Stat(name)
			if err != nil {
				return err
			}
			return os.WriteFile(name, formatted, info.Mode().Perm())
		}
		_, err = cmd.OutOrStdout().Write(formatted)
		return err
	}
	languages := &cobra.Command{Use: "languages", Short: "List implemented code generators", Args: cobra.NoArgs}
	languages.RunE = func(cmd *cobra.Command, _ []string) error {
		names := codegen.List()
		sort.Strings(names)
		return json.NewEncoder(cmd.OutOrStdout()).Encode(names)
	}
	return []*cobra.Command{check, format, languages}
}
