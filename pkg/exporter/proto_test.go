package exporter

import (
	"strings"
	"testing"

	"github.com/neko233-com/bytemsg233/pkg/schema"
)

func TestProtoExporter(t *testing.T) {
	s := &schema.Schema{
		Version:         "bymsg/v1",
		ProtocolVersion: 7,
		Package:         "example.game",
		Enums: map[string]*schema.Enum{
			"PlayerState": {
				Values: map[string]int{
					"PLAYER_STATE_UNKNOWN": 0,
					"PLAYER_STATE_ACTIVE":  1,
				},
			},
		},
		Messages: map[string]*schema.Message{
			"Player": {
				PacketID: 1001,
				Fields: map[string]*schema.Field{
					"id":      {Type: "uint64", Tag: 1},
					"score":   {Type: "int32", Tag: 2},
					"name":    {Type: "string", Tag: 3},
					"tags":    {Type: "list<string>", Tag: 4},
					"attrs":   {Type: "map<string, uint32>", Tag: 5},
					"state":   {Type: "PlayerState", Tag: 6},
					"profile": {Type: "PlayerProfile", Tag: 7},
					"heat":    {Type: "float32", Tag: 8},
					"power":   {Type: "float64", Tag: 9},
					"payload": {Type: "bytes", Tag: 10},
				},
			},
			"PlayerProfile": {
				Fields: map[string]*schema.Field{
					"level": {Type: "uint32", Tag: 1},
				},
			},
		},
	}

	data, err := Proto(s)
	if err != nil {
		t.Fatalf("Proto failed: %v", err)
	}
	content := string(data)
	expected := []string{
		"syntax = \"proto3\";",
		"package example.game;",
		"// ByteMsg233 schema: bymsg/v1",
		"// ByteMsg233 protocolVersion: 7",
		"enum PlayerState {",
		"  PLAYER_STATE_UNKNOWN = 0;",
		"// ByteMsg233 packetId: 1001",
		"message Player {",
		"  uint64 id = 1;",
		"  sint32 score = 2;",
		"  repeated string tags = 4;",
		"  map<string, uint32> attrs = 5;",
		"  PlayerState state = 6;",
		"  PlayerProfile profile = 7;",
		"  float heat = 8;",
		"  double power = 9;",
		"  bytes payload = 10;",
	}
	for _, item := range expected {
		if !strings.Contains(content, item) {
			t.Fatalf("expected proto output to contain %q:\n%s", item, content)
		}
	}
}

func TestProtoExporterRejectsUnsupportedMapKey(t *testing.T) {
	s := &schema.Schema{
		Version: "bymsg/v1",
		Package: "example",
		Messages: map[string]*schema.Message{
			"Bad": {
				Fields: map[string]*schema.Field{
					"weights": {Type: "map<float32, string>", Tag: 1},
				},
			},
		},
		Enums: map[string]*schema.Enum{},
	}

	if _, err := Proto(s); err == nil || !strings.Contains(err.Error(), "proto map key type") {
		t.Fatalf("expected proto map key error, got %v", err)
	}
}

func TestExporterRegistry(t *testing.T) {
	s := &schema.Schema{
		Version:  "bymsg/v1",
		Package:  "example",
		Messages: map[string]*schema.Message{},
		Enums:    map[string]*schema.Enum{},
	}

	if ext, err := Extension("markdown"); err != nil || ext != ".md" {
		t.Fatalf("Extension(markdown) = %q, %v", ext, err)
	}
	if data, err := Export("proto", s, nil); err != nil || !strings.Contains(string(data), "syntax = \"proto3\";") {
		t.Fatalf("Export(proto) failed: %v\n%s", err, data)
	}
}

func TestProtoRoundTrip(t *testing.T) {
	source := &schema.Schema{
		Version:         "bymsg/v1",
		ProtocolVersion: 9,
		Package:         "roundtrip",
		Enums: map[string]*schema.Enum{
			"Status": {Values: map[string]int{"STATUS_UNKNOWN": 0, "STATUS_OK": 1}},
		},
		Messages: map[string]*schema.Message{
			"Packet": {
				PacketID: 7001,
				Fields: map[string]*schema.Field{
					"id":     {Type: "uint64", Tag: 1},
					"status": {Type: "Status", Tag: 2},
					"tags":   {Type: "list<string>", Tag: 3},
					"attrs":  {Type: "map<string, uint32>", Tag: 4},
				},
			},
		},
	}

	data, err := Proto(source)
	if err != nil {
		t.Fatalf("Proto failed: %v", err)
	}
	target, err := schema.ParseProto(data)
	if err != nil {
		t.Fatalf("ParseProto failed: %v\n%s", err, data)
	}
	if target.Version != source.Version || target.ProtocolVersion != source.ProtocolVersion || target.Package != source.Package {
		t.Fatalf("metadata mismatch: %#v", target)
	}
	if target.Messages["Packet"].PacketID != 7001 {
		t.Fatalf("packet id mismatch: %d", target.Messages["Packet"].PacketID)
	}
	if target.Messages["Packet"].Fields["attrs"].Type != "map<string, uint32>" {
		t.Fatalf("attrs type mismatch: %s", target.Messages["Packet"].Fields["attrs"].Type)
	}
}

func TestProtoExporterKeepsNativeJSONProtocolVersion(t *testing.T) {
	s, err := schema.ParseJSON([]byte(`{
  "schema": "bymsg/v1",
  "protocolVersion": 11,
  "package": "jsonproto",
  "Ping": {
    "id": "uint64"
  }
}`))
	if err != nil {
		t.Fatalf("ParseJSON failed: %v", err)
	}
	data, err := Proto(s)
	if err != nil {
		t.Fatalf("Proto failed: %v", err)
	}
	if !strings.Contains(string(data), "// ByteMsg233 protocolVersion: 11") {
		t.Fatalf("proto output missing protocolVersion:\n%s", data)
	}
}
