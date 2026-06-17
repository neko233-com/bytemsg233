package javagen

import (
	"strings"
	"testing"

	"github.com/neko233-com/bytemsg233/pkg/codegen"
	"github.com/neko233-com/bytemsg233/pkg/schema"
)

func TestJavaGenerator(t *testing.T) {
	gen := New()
	s := &schema.Schema{
		Version:         "bymsg/v1",
		ProtocolVersion: 7,
		Package:         "com.example",
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

	if len(files) != 3 {
		t.Fatalf("Expected 3 files, got %d", len(files))
	}

	var userContent string
	var enumContent string
	var protocolContent string
	for _, file := range files {
		switch file.Path {
		case "User.java":
			userContent = string(file.Content)
		case "Status.java":
			enumContent = string(file.Content)
		case "ByteMsgProtocolInfo.java":
			protocolContent = string(file.Content)
		}
	}

	if !strings.Contains(userContent, "package com.example;") {
		t.Error("Expected package")
	}
	if !strings.Contains(userContent, "public class User") {
		t.Error("Expected class")
	}
	if !strings.Contains(protocolContent, "public static final long VERSION = 7L;") ||
		!strings.Contains(protocolContent, "public static long getByteMsg233ProtocolVersion() { return VERSION; }") ||
		strings.Contains(protocolContent, "FINGERPRINT") {
		t.Error("Expected protocol info file")
	}
	if !strings.Contains(userContent, "/** User profile */") {
		t.Error("Expected class comment")
	}
	if !strings.Contains(userContent, "/** Display name */") {
		t.Error("Expected field comment")
	}
	if !strings.Contains(userContent, "private String name = \"\";") {
		t.Error("Expected String field")
	}
	if !strings.Contains(enumContent, "public enum Status") {
		t.Error("Expected enum")
	}
	if !strings.Contains(enumContent, "public static boolean isDefined(int value)") {
		t.Error("Expected enum isDefined helper")
	}
	if !strings.Contains(enumContent, "public static Status fromValue(int value)") {
		t.Error("Expected enum fromValue helper")
	}
	if !strings.Contains(userContent, "import java.util.List;") {
		t.Error("Expected List import")
	}
	if !strings.Contains(userContent, "import java.util.Map;") {
		t.Error("Expected Map import")
	}
	if strings.Contains(userContent, "Concurrent") || strings.Contains(userContent, "java.util.concurrent") {
		t.Error("Java generated pool must not use concurrent collections")
	}
	if !strings.Contains(userContent, "private static final Deque<User> POOL = new ArrayDeque<>()") {
		t.Error("Expected single-threaded ArrayDeque pool")
	}
	if !strings.Contains(userContent, "public static User acquire()") {
		t.Error("Expected pool acquire helper")
	}
	if !strings.Contains(userContent, "public static void release(User value)") {
		t.Error("Expected pool release helper")
	}
	if !strings.Contains(userContent, "public void reset()") {
		t.Error("Expected reset helper")
	}
}

func TestJavaNestedTypes(t *testing.T) {
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

	var content string
	for _, file := range files {
		if file.Path == "Test.java" {
			content = string(file.Content)
			break
		}
	}
	if !strings.Contains(content, "List<String>") {
		t.Error("Expected List<String>")
	}
	if !strings.Contains(content, "Map<String, String>") {
		t.Error("Expected Map<String, String>")
	}
}
