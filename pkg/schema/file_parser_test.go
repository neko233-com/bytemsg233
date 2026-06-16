package schema

import (
	"os"
	"testing"
)

func TestParseFileBmsg(t *testing.T) {
	s, err := ParseFile("../../testdata/user.bmsg")
	if err != nil {
		t.Fatalf("ParseFile .bmsg: %v", err)
	}
	if s.Version != "bymsg/v1" {
		t.Errorf("Expected version bymsg/v1, got %s", s.Version)
	}
	if len(s.Messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(s.Messages))
	}
	if len(s.Enums) != 2 {
		t.Errorf("Expected 2 enums, got %d", len(s.Enums))
	}
}

func TestParseFileBmsgYAMLCompatibility(t *testing.T) {
	t.Run(".bmsg can contain the default JSON DSL", func(t *testing.T) {
		dir := t.TempDir()
		path := dir + "/user.bmsg"
		data := []byte(`{
  "schema": "bymsg/v1",
  "package": "com.example.user",
  "User": {
    "id": { "type": "uint32", "tag": 1 },
    "name": { "type": "string", "tag": 2 }
  }
}`)
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatal(err)
		}

		s, err := ParseFile(path)
		if err != nil {
			t.Fatalf("ParseFile JSON .bmsg: %v", err)
		}
		if _, ok := s.Messages["User"]; !ok {
			t.Fatal("Expected User message")
		}
	})
}

func TestParseFileYAML(t *testing.T) {
	s, err := ParseFile("../../testdata/user.bmsg.yaml")
	if err != nil {
		t.Fatalf("ParseFile .yaml: %v", err)
	}
	if s.Version != "bymsg/v1" {
		t.Errorf("Expected version bymsg/v1, got %s", s.Version)
	}
	if len(s.Messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(s.Messages))
	}
}

func TestParseFileJSON(t *testing.T) {
	s, err := ParseFile("../../testdata/user.json")
	if err != nil {
		t.Fatalf("ParseFile .json: %v", err)
	}
	if s.Version != "bymsg/v1" {
		t.Errorf("Expected version bymsg/v1, got %s", s.Version)
	}
	if len(s.Messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(s.Messages))
	}
	if len(s.Enums) != 1 {
		t.Errorf("Expected 1 enum, got %d", len(s.Enums))
	}
	msg := s.Messages["UserProfile"]
	if msg.Fields["id"].Type != "uint32" {
		t.Errorf("Expected uint32, got %s", msg.Fields["id"].Type)
	}
}

func TestParseFileNativeJSON(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/game.bmsg.json"
	data := []byte(`{
  "schema": "bymsg/v1",
  "package": "com.example.game",
  "enums": {
    "HeroState": {
      "values": {
        "IDLE": 0,
        "MOVING": 1
      }
    }
  },
  "Hero": {
    "description": {
      "zh": "英雄",
      "en": "Hero"
    },
    "id": {
      "type": "uint32",
      "tag": 1
    },
    "state": {
      "type": "HeroState",
      "tag": 2
    }
  }
}`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	s, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile native json: %v", err)
	}
	if len(s.Messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(s.Messages))
	}
	if s.Messages["Hero"].Fields["state"].Type != "HeroState" {
		t.Fatalf("Expected Hero.state to use HeroState")
	}
	if len(s.Enums) != 1 {
		t.Fatalf("Expected 1 enum, got %d", len(s.Enums))
	}
}

func TestParseFileMinimalGameJSON(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/game.bmsg.json"
	data := []byte(`{
  "schema": "bymsg/v1",
  "package": "com.example.game",
  "Hero": {
    "packetId": 1001,
    "comment": "Hero packet",
    "id": { "type": "uint32", "comment": "Hero ID" },
    "name": "string",
    "profile": "HeroProfile"
  },
  "HeroProfile": {
    "level": "uint32"
  }
}`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	s, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile minimal game json: %v", err)
	}

	hero := s.Messages["Hero"]
	if hero.PacketID != 1001 {
		t.Fatalf("Expected packetId 1001, got %d", hero.PacketID)
	}
	if hero.Fields["id"].Tag != 1 || hero.Fields["name"].Tag != 2 || hero.Fields["profile"].Tag != 3 {
		t.Fatalf("Expected declaration-order tags, got id=%d name=%d profile=%d", hero.Fields["id"].Tag, hero.Fields["name"].Tag, hero.Fields["profile"].Tag)
	}
	if hero.Fields["profile"].Type != "HeroProfile" {
		t.Fatalf("Expected message class reference HeroProfile")
	}
	if hero.Fields["id"].Description == nil || hero.Fields["id"].Description.En != "Hero ID" {
		t.Fatalf("Expected comment shorthand to become description")
	}
}

func TestParseFileFirstClassEnumMapListComments(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/battle.bmsg.json"
	data := []byte(`{
  "schema": "bymsg/v1",
  "package": "com.example.battle",
  "enums": {
    "MoveState": ["IDLE", "RUN", "DASH"],
    "ItemQuality": {
      "comment": "Item quality",
      "values": ["WHITE", "GREEN", "BLUE"]
    },
    "ErrorCode": {
      "OK": 0,
      "TIMEOUT": 10
    }
  },
  "BattlePacket": {
    "packetId": 2001,
    "comment": "Battle packet",
    "state": "MoveState",
    "skill_ids": { "list": "uint32", "comment": "Skill IDs" },
    "attrs": { "map": ["string", "uint32"], "comment": "Attributes" },
    "inventory": { "map": { "key": "uint32", "value": "ItemQuality" } }
  }
}`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	s, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile first-class json: %v", err)
	}

	if s.Enums["MoveState"].Values["DASH"] != 2 {
		t.Fatalf("Expected enum array values to auto-number")
	}
	if s.Enums["ItemQuality"].Description == nil || s.Enums["ItemQuality"].Description.En != "Item quality" {
		t.Fatalf("Expected enum comment shorthand")
	}
	if s.Enums["ErrorCode"].Values["TIMEOUT"] != 10 {
		t.Fatalf("Expected enum object shorthand")
	}

	msg := s.Messages["BattlePacket"]
	if msg.Fields["skill_ids"].Type != "list<uint32>" {
		t.Fatalf("Expected structured list type, got %s", msg.Fields["skill_ids"].Type)
	}
	if msg.Fields["attrs"].Type != "map<string, uint32>" {
		t.Fatalf("Expected structured map array type, got %s", msg.Fields["attrs"].Type)
	}
	if msg.Fields["inventory"].Type != "map<uint32, ItemQuality>" {
		t.Fatalf("Expected structured map object type, got %s", msg.Fields["inventory"].Type)
	}
	if msg.Fields["attrs"].Description == nil || msg.Fields["attrs"].Description.En != "Attributes" {
		t.Fatalf("Expected field comment shorthand")
	}
}

func TestParseFileMessagesWrapperWithShorthand(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/wrapped.bmsg.json"
	data := []byte(`{
  "schema": "bymsg/v1",
  "package": "com.example.wrapped",
  "messages": {
    "Player": {
      "comment": "Player packet",
      "id": "uint64",
      "tags": { "list": "string" },
      "attrs": { "map": ["string", "string"] }
    }
  }
}`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	s, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile wrapped shorthand: %v", err)
	}
	msg := s.Messages["Player"]
	if msg.Fields["id"].Type != "uint64" {
		t.Fatalf("Expected string field shorthand")
	}
	if msg.Fields["tags"].Type != "list<string>" {
		t.Fatalf("Expected wrapped structured list")
	}
	if msg.Fields["attrs"].Type != "map<string, string>" {
		t.Fatalf("Expected wrapped structured map")
	}
}
