package pygen

import (
	"strings"
	"testing"

	"github.com/neko233-com/bytemsg233/pkg/codegen"
	"github.com/neko233-com/bytemsg233/pkg/schema"
)

func TestPythonGenerator(t *testing.T) {
	gen := New()
	s := &schema.Schema{
		Version: "bymsg/v1",
		Package: "user",
		Messages: map[string]*schema.Message{
			"User": {
				Description: &schema.Description{En: "User profile"},
				Fields: map[string]*schema.Field{
					"name": {Type: "string", Tag: 1, Description: &schema.Description{En: "Display name"}},
					"age":  {Type: "uint32", Tag: 2, Description: &schema.Description{En: "Age"}},
				},
			},
		},
		Enums: map[string]*schema.Enum{
			"Status": {
				Values: map[string]int{"ACTIVE": 0, "INACTIVE": 1},
			},
		},
	}

	files, err := gen.Generate(s, &codegen.GenerateOptions{})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if files[0].Path != "ByteMsg233_Export.py" {
		t.Fatalf("generated path = %q, want ByteMsg233_Export.py", files[0].Path)
	}

	content := string(files[0].Content)
	if !strings.Contains(content, "from dataclasses import dataclass") {
		t.Error("Expected dataclass import")
	}
	if !strings.Contains(content, "from enum import IntEnum") {
		t.Error("Expected IntEnum import")
	}
	if !strings.Contains(content, "class Status(IntEnum):") {
		t.Error("Expected enum class")
	}
	if !strings.Contains(content, "def from_value(cls, value: int) -> \"Status\":") {
		t.Error("Expected enum from_value helper")
	}
	if !strings.Contains(content, "@dataclass") {
		t.Error("Expected dataclass decorator")
	}
	if !strings.Contains(content, "\"\"\"User profile\"\"\"") {
		t.Error("Expected class docstring")
	}
	if !strings.Contains(content, "class User:") {
		t.Error("Expected User class")
	}
	if !strings.Contains(content, "# Display name") {
		t.Error("Expected field comment")
	}
	if !strings.Contains(content, "name: str = \"\"") {
		t.Error("Expected str field")
	}
	if !strings.Contains(content, "age: int = 0") {
		t.Error("Expected int field")
	}
	if !strings.Contains(content, "def acquire(cls) -> \"User\":") {
		t.Error("Expected pool acquire helper")
	}
	if !strings.Contains(content, "def release(self) -> None:") {
		t.Error("Expected pool release helper")
	}
	if !strings.Contains(content, "def reset(self) -> None:") {
		t.Error("Expected reset helper")
	}
}

func TestPythonNestedTypes(t *testing.T) {
	gen := New()
	s := &schema.Schema{
		Version: "bymsg/v1",
		Package: "test",
		Messages: map[string]*schema.Message{
			"Test": {
				Fields: map[string]*schema.Field{
					"tags":     {Type: "list<string>", Tag: 1},
					"metadata": {Type: "map<string, string>", Tag: 2},
				},
			},
		},
		Enums: map[string]*schema.Enum{},
	}

	files, err := gen.Generate(s, &codegen.GenerateOptions{})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	content := string(files[0].Content)
	if !strings.Contains(content, "List[str]") {
		t.Error("Expected List[str]")
	}
	if !strings.Contains(content, "Dict[str, str]") {
		t.Error("Expected Dict[str, str]")
	}
}
