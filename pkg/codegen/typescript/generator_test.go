package tsgen

import (
	"strings"
	"testing"

	"github.com/neko233-com/bytemsg233/pkg/codegen"
	"github.com/neko233-com/bytemsg233/pkg/schema"
)

func TestTypeScriptGenerator(t *testing.T) {
	gen := New()
	s := &schema.Schema{
		Version:         "bymsg/v1",
		ProtocolVersion: 7,
		Package:         "user",
		Messages: map[string]*schema.Message{
			"UserProfile": {
				Description: &schema.Description{En: "User profile"},
				Fields: map[string]*schema.Field{
					"id":   {Type: "uint32", Tag: 1, Description: &schema.Description{En: "User ID"}},
					"name": {Type: "string", Tag: 2, Description: &schema.Description{En: "Display name"}},
				},
			},
		},
		Enums: map[string]*schema.Enum{
			"UserType": {
				Values: map[string]int{"ADMIN": 0, "USER": 1},
			},
		},
	}

	files, err := gen.Generate(s, &codegen.GenerateOptions{})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if files[0].Path != "ByteMsg233_Export.ts" {
		t.Fatalf("generated path = %q, want ByteMsg233_Export.ts", files[0].Path)
	}

	content := string(files[0].Content)
	if !strings.Contains(content, "class ByteMsgObjectPool") {
		t.Error("Expected shared object pool helper")
	}
	if !strings.Contains(content, "export class UserProfile") {
		t.Error("Expected UserProfile class")
	}
	if !strings.Contains(content, "export const ByteMsgProtocolVersion = 7;") ||
		!strings.Contains(content, "export function getByteMsg233ProtocolVersion(): number") ||
		strings.Contains(content, "ByteMsgProtocolFingerprint") {
		t.Error("Expected only protocol version constant")
	}
	if !strings.Contains(content, "/** User profile */") {
		t.Error("Expected class comment")
	}
	if !strings.Contains(content, "/** User ID */") {
		t.Error("Expected field comment")
	}
	if !strings.Contains(content, "id: number = 0;") {
		t.Error("Expected id: number")
	}
	if !strings.Contains(content, "name: string = \"\";") {
		t.Error("Expected name: string")
	}
	if !strings.Contains(content, "export enum UserType") {
		t.Error("Expected UserType enum")
	}
	if !strings.Contains(content, "export namespace UserType") {
		t.Error("Expected enum namespace helper")
	}
	if !strings.Contains(content, "export function fromValue(value: number): UserType") {
		t.Error("Expected enum fromValue helper")
	}
	if !strings.Contains(content, "static acquire(init?: Partial<UserProfile>): UserProfile") {
		t.Error("Expected acquire helper")
	}
	if !strings.Contains(content, "release(): void") {
		t.Error("Expected release helper")
	}
	if !strings.Contains(content, "reset(): void") {
		t.Error("Expected reset method")
	}
}

func TestTypeScriptNestedTypes(t *testing.T) {
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
	if !strings.Contains(content, "string[]") {
		t.Error("Expected string[]")
	}
	if !strings.Contains(content, "Record<string, string>") {
		t.Error("Expected Record<string, string>")
	}
}
