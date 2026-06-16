package gocodegen

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neko233-com/bytemsg233/pkg/codegen"
	"github.com/neko233-com/bytemsg233/pkg/schema"
)

func TestGoGenerator(t *testing.T) {
	gen := New()

	if gen.Name() != "go" {
		t.Errorf("Expected name 'go', got '%s'", gen.Name())
	}

	if gen.FileExtension() != ".go" {
		t.Errorf("Expected extension '.go', got '%s'", gen.FileExtension())
	}

	s := &schema.Schema{
		Version: "bymsg/v1",
		Package: "user",
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
					"ADMIN": 0,
					"USER":  1,
				},
			},
		},
	}

	files, err := gen.Generate(s, &codegen.GenerateOptions{})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(files))
	}
	if files[0].Path != "ByteMsg233_Export.go" {
		t.Fatalf("generated path = %q, want ByteMsg233_Export.go", files[0].Path)
	}

	content := string(files[0].Content)

	if !strings.Contains(content, "package user") {
		t.Error("Expected package declaration")
	}
	if !strings.Contains(content, "type UserProfile struct") {
		t.Error("Expected UserProfile struct")
	}
	if !strings.Contains(content, "// User profile") {
		t.Error("Expected class comment")
	}
	if !strings.Contains(content, "// User ID") {
		t.Error("Expected field comment")
	}
	if !strings.Contains(content, "Id uint32") {
		t.Error("Expected Id field")
	}
	if !strings.Contains(content, "Name string") {
		t.Error("Expected Name field")
	}
	if !strings.Contains(content, "`json:\"name\" bytemsg:\"2\"`") {
		t.Error("Expected JSON field tag")
	}
	if !strings.Contains(content, "type UserType int32") {
		t.Error("Expected UserType enum")
	}
	if !strings.Contains(content, "UserTypeAdmin UserType = 0") {
		t.Error("Expected ADMIN constant")
	}
	if !strings.Contains(content, "func ParseUserType(value int32) (UserType, bool)") {
		t.Error("Expected enum parse helper")
	}
	if !strings.Contains(content, "func (x UserType) String() string") {
		t.Error("Expected enum String helper")
	}
	if !strings.Contains(content, "func AcquireUserProfile() *UserProfile") {
		t.Error("Expected pool acquire helper")
	}
	if !strings.Contains(content, "func ReleaseUserProfile(value *UserProfile)") {
		t.Error("Expected pool release helper")
	}
	if !strings.Contains(content, "func (x *UserProfile) Reset()") {
		t.Error("Expected reset method")
	}
	if !strings.Contains(content, "func (x *UserProfile) MarshalByteMsgPrettyString() (string, error)") {
		t.Error("Expected pretty string marshal helper")
	}
	if !strings.Contains(content, "func (x *UserProfile) UnmarshalByteMsgPrettyString(value string) error") {
		t.Error("Expected pretty string unmarshal helper")
	}
}

func TestGoGeneratorNestedTypes(t *testing.T) {
	gen := New()
	s := &schema.Schema{
		Version: "bymsg/v1",
		Package: "test",
		Messages: map[string]*schema.Message{
			"Test": {
				Fields: map[string]*schema.Field{
					"tags":     {Type: "list<string>", Tag: 1},
					"metadata": {Type: "map<string, string>", Tag: 2},
					"nested":   {Type: "map<string, list<uint32>>", Tag: 3},
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
	if !strings.Contains(content, "Tags []string") {
		t.Error("Expected Tags []string")
	}
	if !strings.Contains(content, "Metadata map[string]string") {
		t.Error("Expected Metadata map[string]string")
	}
}

func TestGoGeneratorEnumOnlyImportsFmt(t *testing.T) {
	gen := New()
	s := &schema.Schema{
		Version:  "bymsg/v1",
		Package:  "enumonly",
		Messages: map[string]*schema.Message{},
		Enums: map[string]*schema.Enum{
			"Status": {
				Values: map[string]int{
					"UNKNOWN": 0,
					"OK":      1,
				},
			},
		},
	}

	files, err := gen.Generate(s, &codegen.GenerateOptions{})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	content := string(files[0].Content)
	if !strings.Contains(content, "import \"fmt\"") {
		t.Fatalf("enum-only generated code must import fmt:\n%s", content)
	}
}

func TestGeneratedGoCodeCompilesAndRoundTrips(t *testing.T) {
	gen := New()
	s := &schema.Schema{
		Version: "bymsg/v1",
		Package: "protocol",
		Messages: map[string]*schema.Message{
			"Empty": {
				PacketID: 1001,
				Fields:   map[string]*schema.Field{},
			},
			"Inner": {
				Fields: map[string]*schema.Field{
					"score": {Type: "int32", Tag: 1},
					"label": {Type: "string", Tag: 2},
				},
			},
			"Player": {
				PacketID: 1002,
				Fields: map[string]*schema.Field{
					"id":      {Type: "uint64", Tag: 1},
					"active":  {Type: "bool", Tag: 2},
					"power":   {Type: "float64", Tag: 3},
					"mood":    {Type: "PlayerMood", Tag: 4},
					"tags":    {Type: "list<string>", Tag: 5},
					"attrs":   {Type: "map<string, uint32>", Tag: 6},
					"nested":  {Type: "map<string, list<uint32>>", Tag: 7},
					"inner":   {Type: "Inner", Tag: 8},
					"inners":  {Type: "list<Inner>", Tag: 9},
					"payload": {Type: "bytes", Tag: 10},
				},
			},
		},
		Enums: map[string]*schema.Enum{
			"PlayerMood": {
				Values: map[string]int{
					"UNKNOWN": 0,
					"HAPPY":   1,
					"ANGRY":   2,
				},
			},
		},
	}

	files, err := gen.Generate(s, &codegen.GenerateOptions{})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	tmpDir := t.TempDir()
	repoRoot, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	writeFile(t, filepath.Join(tmpDir, "go.mod"), "module generatedcheck\n\ngo 1.26\n\nrequire github.com/neko233-com/bytemsg233 v0.0.0\n\nreplace github.com/neko233-com/bytemsg233 => "+filepath.ToSlash(repoRoot)+"\n")
	writeFile(t, filepath.Join(tmpDir, "types.go"), string(files[0].Content))
	writeFile(t, filepath.Join(tmpDir, "types_test.go"), generatedGoRoundTripTest)

	cmd := exec.Command("go", "test", "./...")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("generated code test failed: %v\n%s", err, output)
	}
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

const generatedGoRoundTripTest = `package protocol

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestGeneratedRoundTripAndRegistry(t *testing.T) {
	source := &Player{
		Id: 42,
		Active: true,
		Power: 12.5,
		Mood: PlayerMoodHappy,
		Tags: []string{"alpha", "beta"},
		Attrs: map[string]uint32{"level": 7, "vip": 2},
		Nested: map[string][]uint32{"rewards": {1, 2, 3}},
		Inner: Inner{Score: -9, Label: "core"},
		Inners: []Inner{{Score: 1, Label: "a"}, {Score: -2, Label: "b"}},
		Payload: []byte{1, 2, 3},
	}

	var buf bytes.Buffer
	if err := source.MarshalByteMsgTo(&buf); err != nil {
		t.Fatalf("marshal: %v", err)
	}

	target := AcquirePlayer()
	defer ReleasePlayer(target)
	if err := target.UnmarshalByteMsg(buf.Bytes()); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !reflect.DeepEqual(source, target) {
		t.Fatalf("roundtrip mismatch:\nsource=%#v\ntarget=%#v", source, target)
	}

	packet, ok := AcquireByteMsgPacketById(1002)
	if !ok {
		t.Fatalf("packet id 1002 not registered")
	}
	if _, ok := packet.(*Player); !ok {
		t.Fatalf("packet id 1002 type = %T, want *Player", packet)
	}
	if !ReleaseByteMsgPacket(1002, packet) {
		t.Fatalf("release packet id 1002 failed")
	}

	text := string(source.AppendByteMsgText(make([]byte, 0, 512)))
	if !strings.Contains(text, "Player{") || !strings.Contains(text, "Id:42") || !strings.Contains(text, "Inner{") {
		t.Fatalf("debug text missing fields: %s", text)
	}

	textBuf := make([]byte, 0, 1024)
	allocs := testing.AllocsPerRun(1000, func() {
		dst := textBuf[:0]
		dst = source.AppendByteMsgText(dst)
		if len(dst) == 0 {
			panic("empty text")
		}
	})
	if allocs != 0 {
		t.Fatalf("AppendByteMsgText allocs = %v, want 0", allocs)
	}

	pretty, err := source.MarshalByteMsgPrettyString()
	if err != nil {
		t.Fatalf("pretty marshal: %v", err)
	}
	if !strings.Contains(pretty, "\n  \"id\": 42") || !strings.Contains(pretty, "\"inner\": {") {
		t.Fatalf("pretty string missing expected JSON shape: %s", pretty)
	}
	var fromPretty Player
	if err := fromPretty.UnmarshalByteMsgPrettyString(pretty); err != nil {
		t.Fatalf("pretty unmarshal: %v", err)
	}
	if !reflect.DeepEqual(source, &fromPretty) {
		t.Fatalf("pretty roundtrip mismatch:\nsource=%#v\ntarget=%#v", source, fromPretty)
	}
}

func TestGeneratedEmptyPacketZeroAlloc(t *testing.T) {
	packet := AcquireEmpty()
	defer ReleaseEmpty(packet)

	allocs := testing.AllocsPerRun(1000, func() {
		var buf bytes.Buffer
		if err := packet.MarshalByteMsgTo(&buf); err != nil {
			panic(err)
		}
		if buf.Len() != 0 {
			panic("empty packet encoded bytes")
		}
		if err := packet.UnmarshalByteMsg(nil); err != nil {
			panic(err)
		}
	})
	if allocs != 0 {
		t.Fatalf("empty packet MarshalByteMsgTo/UnmarshalByteMsg allocs = %v, want 0", allocs)
	}
}

func TestGeneratedPacketPoolLimit(t *testing.T) {
	for {
		select {
		case <-playerPool:
		default:
			goto drained
		}
	}

drained:
	for i := 0; i < ByteMsgPacketPoolLimit+1; i++ {
		ReleasePlayer(&Player{Id: uint64(i)})
	}
	if len(playerPool) != ByteMsgPacketPoolLimit {
		t.Fatalf("player pool len = %d, want %d", len(playerPool), ByteMsgPacketPoolLimit)
	}
	for {
		select {
		case <-playerPool:
		default:
			return
		}
	}
}
`
