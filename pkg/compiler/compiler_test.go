package compiler

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCompilerBmsg(t *testing.T) {
	tmpDir := t.TempDir()

	schemaContent := `schema: bymsg/v1
package: test

message User {
    uint32 id = 1
    string name = 2
}
`
	schemaPath := filepath.Join(tmpDir, "user.bmsg")
	if err := os.WriteFile(schemaPath, []byte(schemaContent), 0644); err != nil {
		t.Fatal(err)
	}

	comp := New()
	err := comp.Compile(&CompileOptions{
		InputFile: schemaPath,
		OutputDir: tmpDir,
		Languages: []string{"go"},
	})
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	outputPath := filepath.Join(tmpDir, "ByteMsg233_Export.go")
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Expected ByteMsg233_Export.go to be created")
	}
}

func TestCompilerYaml(t *testing.T) {
	tmpDir := t.TempDir()

	schemaContent := `
schema: bymsg/v1
package: test

messages:
  User:
    fields:
      id:
        type: uint32
        tag: 1
`
	schemaPath := filepath.Join(tmpDir, "user.bmsg.yaml")
	if err := os.WriteFile(schemaPath, []byte(schemaContent), 0644); err != nil {
		t.Fatal(err)
	}

	comp := New()
	err := comp.Compile(&CompileOptions{
		InputFile: schemaPath,
		OutputDir: tmpDir,
		Languages: []string{"go"},
	})
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	outputPath := filepath.Join(tmpDir, "ByteMsg233_Export.go")
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Expected ByteMsg233_Export.go to be created")
	}
}

func TestCompilerMultipleLanguages(t *testing.T) {
	tmpDir := t.TempDir()

	schemaContent := `schema: bymsg/v1
package: test

message User {
    uint32 id = 1
    string name = 2
    list<string> tags = 3
    map<string, string> meta = 4
}
`
	schemaPath := filepath.Join(tmpDir, "user.bmsg")
	if err := os.WriteFile(schemaPath, []byte(schemaContent), 0644); err != nil {
		t.Fatal(err)
	}

	comp := New()
	err := comp.Compile(&CompileOptions{
		InputFile: schemaPath,
		OutputDir: tmpDir,
		Languages: []string{"go", "csharp", "java", "typescript", "rust", "python"},
	})
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	expectedFiles := map[string]string{
		"ByteMsg233_Export.go": "func AcquireUser() *User",
		"ByteMsg233_Export.cs": "public static User Rent()",
		"User.java":            "public static User acquire()",
		"ByteMsg233_Export.ts": "export class User",
		"ByteMsg233_Export.rs": "pub struct User",
		"ByteMsg233_Export.py": "def acquire(cls) -> \"User\":",
	}

	for file, expected := range expectedFiles {
		path := filepath.Join(tmpDir, file)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("Expected %s to be created: %v", file, err)
			continue
		}
		content := string(data)
		if len(content) == 0 {
			t.Errorf("Expected %s to have content", file)
			continue
		}
		t.Logf("=== %s ===\n%s", file, content[:min(len(content), 200)])
		if expected != "" && !contains(content, expected) {
			t.Errorf("Expected %s to contain %q", file, expected)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
