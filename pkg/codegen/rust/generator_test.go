package rustgen

import (
	"strings"
	"testing"

	"github.com/neko233-com/bytemsg233/pkg/codegen"
	"github.com/neko233-com/bytemsg233/pkg/schema"
)

func TestRustGeneratorPath(t *testing.T) {
	gen := New()
	s := &schema.Schema{
		Version:         "bymsg/v1",
		ProtocolVersion: 7,
		Package:         "test",
		Messages: map[string]*schema.Message{
			"User": {
				Fields: map[string]*schema.Field{
					"id": {Type: "uint32", Tag: 1},
				},
			},
		},
		Enums: map[string]*schema.Enum{},
	}

	files, err := gen.Generate(s, &codegen.GenerateOptions{})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if files[0].Path != "ByteMsg233_Export.rs" {
		t.Fatalf("generated path = %q, want ByteMsg233_Export.rs", files[0].Path)
	}
	if !strings.Contains(string(files[0].Content), "pub struct User") {
		t.Fatal("expected User struct")
	}
	if !strings.Contains(string(files[0].Content), "pub const BYTE_MSG_PROTOCOL_VERSION: u64 = 7;") ||
		!strings.Contains(string(files[0].Content), "pub fn get_bytemsg233_protocol_version() -> u64") ||
		strings.Contains(string(files[0].Content), "BYTE_MSG_PROTOCOL_FINGERPRINT") {
		t.Fatal("expected only protocol version constant")
	}
}
