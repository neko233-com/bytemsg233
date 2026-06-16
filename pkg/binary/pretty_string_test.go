package binary

import (
	"strings"
	"testing"
)

type prettyInner struct {
	Label string `json:"label"`
}

type prettyMessage struct {
	ID      uint64              `json:"id"`
	Active  bool                `json:"active"`
	Score   float32             `json:"score"`
	Payload []byte              `json:"payload"`
	Flags   map[bool]string     `json:"flags"`
	Ratios  map[float32]uint32  `json:"ratios"`
	Names   map[string][]uint32 `json:"names"`
	Inner   prettyInner         `json:"inner"`
}

func TestPrettyStringDebugOutput(t *testing.T) {
	source := prettyMessage{
		ID:      42,
		Active:  true,
		Score:   9.5,
		Payload: []byte{1, 2, 3},
		Flags:   map[bool]string{false: "off", true: "on"},
		Ratios:  map[float32]uint32{1.5: 15, 2.5: 25},
		Names:   map[string][]uint32{"alpha": {1, 2}},
		Inner:   prettyInner{Label: "core"},
	}

	pretty, err := MarshalPrettyString(&source)
	if err != nil {
		t.Fatalf("MarshalPrettyString: %v", err)
	}
	if !strings.Contains(pretty, "\n  \"id\": 42") {
		t.Fatalf("debug text missing json field name: %s", pretty)
	}
	if !strings.Contains(pretty, "\"key\": false") || !strings.Contains(pretty, "\"payload\": \"AQID\"") {
		t.Fatalf("debug text missing map entries or bytes: %s", pretty)
	}

	if strings.Contains(pretty, "Unmarshal") {
		t.Fatalf("debug output must not advertise string deserialization: %s", pretty)
	}
}
