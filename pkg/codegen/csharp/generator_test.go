package csharpgen

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neko233-com/bytemsg233/pkg/codegen"
	"github.com/neko233-com/bytemsg233/pkg/schema"
)

func TestCSharpGenerator(t *testing.T) {
	gen := New()

	if gen.Name() != "csharp" {
		t.Errorf("Expected name 'csharp', got '%s'", gen.Name())
	}

	s := &schema.Schema{
		Version:         "bymsg/v1",
		ProtocolVersion: 7,
		Package:         "Example.User",
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
				Values: map[string]int{
					"Admin": 0,
					"User":  1,
				},
			},
		},
	}

	files, err := gen.Generate(s, &codegen.GenerateOptions{})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	content := string(files[0].Content)
	if files[0].Path != "ByteMsg233_Export.cs" {
		t.Fatalf("generated path = %q, want ByteMsg233_Export.cs", files[0].Path)
	}

	if !strings.Contains(content, "namespace Example.User") {
		t.Error("Expected namespace declaration")
	}
	if !strings.Contains(content, "public partial class UserProfile") {
		t.Error("Expected partial UserProfile class")
	}
	if !strings.Contains(content, "public const ulong Version = 7UL;") ||
		!strings.Contains(content, "public static ulong GetByteMsg233ProtocolVersion() => Version;") ||
		strings.Contains(content, "Fingerprint") {
		t.Error("Expected only protocol version constant")
	}
	if strings.Contains(content, "public sealed class UserProfile") {
		t.Error("C# generated classes must not be sealed")
	}
	if !strings.Contains(content, "/// User profile") {
		t.Error("Expected class comment")
	}
	if !strings.Contains(content, "/// User ID") {
		t.Error("Expected field comment")
	}
	if !strings.Contains(content, "public uint Id") {
		t.Error("Expected Id property")
	}
	if !strings.Contains(content, "public string Name") {
		t.Error("Expected Name property")
	}
	if !strings.Contains(content, "public enum UserType") {
		t.Error("Expected UserType enum")
	}
	if !strings.Contains(content, "public static class UserTypeExtensions") {
		t.Error("Expected enum extensions helper")
	}
	if !strings.Contains(content, "public static UserType FromValue(int raw)") {
		t.Error("Expected enum FromValue helper")
	}
	if !strings.Contains(content, "public static UserProfile Rent()") {
		t.Error("Expected pool rent helper")
	}
	if !strings.Contains(content, "public static void Prewarm(int count)") {
		t.Error("Expected pool prewarm helper")
	}
	if !strings.Contains(content, "private static readonly Stack<UserProfile> Pool = new Stack<UserProfile>()") {
		t.Error("Expected Stack-backed Unity-friendly pool")
	}
	if strings.Contains(content, "Concurrent") {
		t.Error("C# generated pool must not use concurrent collections")
	}
	if strings.Contains(content, "PoolLock") || strings.Contains(content, "lock (") {
		t.Error("C# generated pool must stay single-threaded without locks")
	}
	if !strings.Contains(content, "public static void Return(UserProfile value)") {
		t.Error("Expected pool return helper")
	}
	if !strings.Contains(content, "public void Reset()") {
		t.Error("Expected reset helper")
	}
}

func TestCSharpNestedTypes(t *testing.T) {
	gen := New()
	s := &schema.Schema{
		Version: "bymsg/v1",
		Package: "Test",
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
	if !strings.Contains(content, "List<string>") {
		t.Error("Expected List<string>")
	}
	if !strings.Contains(content, "Dictionary<string, string>") {
		t.Error("Expected Dictionary<string, string>")
	}
	if !strings.Contains(content, "Tags.Clear();") {
		t.Error("Expected list reset to clear without allocation")
	}
	if !strings.Contains(content, "Metadata.Clear();") {
		t.Error("Expected dictionary reset to clear without allocation")
	}
}

func TestCSharpPartialClassAndNestedReset(t *testing.T) {
	gen := New()
	s := &schema.Schema{
		Version: "bymsg/v1",
		Package: "Unity.Game",
		Messages: map[string]*schema.Message{
			"Inner": {
				Fields: map[string]*schema.Field{
					"score": {Type: "uint32", Tag: 1},
				},
			},
			"Outer": {
				Fields: map[string]*schema.Field{
					"inner": {Type: "Inner", Tag: 1},
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
	if !strings.Contains(content, "public partial class Inner") || !strings.Contains(content, "public partial class Outer") {
		t.Error("Expected all generated C# messages to be partial")
	}
	if !strings.Contains(content, "public Inner Inner { get; set; } = new Inner();") {
		t.Error("Expected nested message default new instance")
	}
	if !strings.Contains(content, "Inner = new Inner();") || !strings.Contains(content, "Inner.Reset();") {
		t.Error("Expected nested reset to reuse existing instance or create when null")
	}
}

func TestGeneratedCSharpPartialExtensionCompiles(t *testing.T) {
	if _, err := exec.LookPath("dotnet"); err != nil {
		t.Skip("dotnet not available")
	}

	gen := New()
	s := &schema.Schema{
		Version: "bymsg/v1",
		Package: "Unity.Game",
		Messages: map[string]*schema.Message{
			"Hero": {
				Fields: map[string]*schema.Field{
					"id":   {Type: "uint32", Tag: 1},
					"name": {Type: "string", Tag: 2},
					"tags": {Type: "list<string>", Tag: 3},
				},
			},
		},
		Enums: map[string]*schema.Enum{},
	}

	files, err := gen.Generate(s, &codegen.GenerateOptions{})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	tmpDir := t.TempDir()
	writeCSharpFile(t, filepath.Join(tmpDir, "Generated.cs"), string(files[0].Content))
	writeCSharpFile(t, filepath.Join(tmpDir, "Hero.Extensions.cs"), `namespace Unity.Game
{
	public partial class Hero
	{
		public bool IsNamed()
		{
			return Name.Length > 0;
		}
	}
}
`)
	writeCSharpFile(t, filepath.Join(tmpDir, "GeneratedCheck.csproj"), `<Project Sdk="Microsoft.NET.Sdk">
	<PropertyGroup>
		<TargetFramework>netstandard2.0</TargetFramework>
		<LangVersion>latest</LangVersion>
	</PropertyGroup>
</Project>
`)

	cmd := exec.Command("dotnet", "build", "--nologo", "-v:q")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("dotnet build failed: %v\n%s", err, output)
	}
}

func writeCSharpFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
