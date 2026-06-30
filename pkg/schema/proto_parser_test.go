package schema

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseProto(t *testing.T) {
	data := []byte(`syntax = "proto3";

package protocol;

// ByteMsg233 schema: bymsg/v1
// ByteMsg233 protocolVersion: 7

enum PlayerState {
  PLAYER_STATE_UNKNOWN = 0;
  PLAYER_STATE_ACTIVE = 1;
}

message PlayerProfile {
  uint32 level = 1;
}

// ByteMsg233 packetId: 1001
message Player {
  uint64 id = 1;
  sint32 score = 2;
  string name = 3;
  repeated string tags = 4;
  map<string, uint32> attrs = 5;
  PlayerState state = 6;
  PlayerProfile profile = 7;
  float heat = 8;
  double power = 9;
  bytes payload = 10;
}
`)

	s, err := ParseProto(data)
	if err != nil {
		t.Fatalf("ParseProto failed: %v", err)
	}
	if s.Version != "bymsg/v1" || s.ProtocolVersion != 7 || s.Package != "protocol" {
		t.Fatalf("metadata mismatch: %#v", s)
	}
	if s.Enums["PlayerState"].Values["PLAYER_STATE_ACTIVE"] != 1 {
		t.Fatalf("enum value mismatch")
	}
	player := s.Messages["Player"]
	if player.PacketID != 1001 {
		t.Fatalf("packet id = %d, want 1001", player.PacketID)
	}
	expectedTypes := map[string]string{
		"id":      "uint64",
		"score":   "int32",
		"name":    "string",
		"tags":    "list<string>",
		"attrs":   "map<string, uint32>",
		"state":   "PlayerState",
		"profile": "PlayerProfile",
		"heat":    "float32",
		"power":   "float64",
		"payload": "bytes",
	}
	for fieldName, expectedType := range expectedTypes {
		if player.Fields[fieldName].Type != expectedType {
			t.Fatalf("%s type = %s, want %s", fieldName, player.Fields[fieldName].Type, expectedType)
		}
	}
}

func TestImportFileProto(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "protocol.proto")
	if err := os.WriteFile(path, []byte(`syntax = "proto3";
package protocol;
message Ping {
  uint64 id = 1;
}
`), 0644); err != nil {
		t.Fatal(err)
	}

	s, err := ImportFile(path, nil)
	if err != nil {
		t.Fatalf("ImportFile proto failed: %v", err)
	}
	if s.Version != "bymsg/v1" {
		t.Fatalf("expected default schema version, got %s", s.Version)
	}
	if s.Messages["Ping"].Fields["id"].Type != "uint64" {
		t.Fatalf("expected uint64 field")
	}
}

func TestParseProtoRejectsUnsupportedSyntax(t *testing.T) {
	_, err := ParseProto([]byte(`syntax = "proto3";
package protocol;
message Bad {
  oneof value {
    string name = 1;
  }
}
`))
	if err == nil || !strings.Contains(err.Error(), "oneof is not supported") {
		t.Fatalf("expected unsupported oneof error, got %v", err)
	}
}

func TestParseProtoRejectsUnsupportedMapKey(t *testing.T) {
	_, err := ParseProto([]byte(`syntax = "proto3";
package protocol;
message Bad {
  map<float, string> weights = 1;
}
`))
	if err == nil || !strings.Contains(err.Error(), "map key type") {
		t.Fatalf("expected unsupported map key error, got %v", err)
	}
}

func TestParseProtoRejectsUnsupportedScalar(t *testing.T) {
	_, err := ParseProto([]byte(`syntax = "proto3";
package protocol;
message Bad {
  fixed32 id = 1;
}
`))
	if err == nil || !strings.Contains(err.Error(), `unknown type "fixed32"`) {
		t.Fatalf("expected unsupported scalar error, got %v", err)
	}
}

func TestParseProtoRejectsRepeatedMap(t *testing.T) {
	_, err := ParseProto([]byte(`syntax = "proto3";
package protocol;
message Bad {
  repeated map<string, uint32> attrs = 1;
}
`))
	if err == nil || !strings.Contains(err.Error(), "repeated map fields are not supported") {
		t.Fatalf("expected repeated map error, got %v", err)
	}
}

func TestParseProtoPacketCommentDoesNotLeakFromField(t *testing.T) {
	s, err := ParseProto([]byte(`syntax = "proto3";
package protocol;
message First {
  uint64 id = 1;
  // ByteMsg233 packetId: 999
}
message Second {
  uint64 id = 1;
}
`))
	if err != nil {
		t.Fatalf("ParseProto failed: %v", err)
	}
	if s.Messages["Second"].PacketID != 0 {
		t.Fatalf("field comment leaked into next message packet id: %d", s.Messages["Second"].PacketID)
	}
}
