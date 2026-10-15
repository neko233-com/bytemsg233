package schema

import (
	"bytes"
	"reflect"
	"testing"
)

func TestFormatPreservesSchemaAndIsIdempotent(t *testing.T) {
	for _, tt := range []struct{ name, input string }{
		{"game.bmsg.json", `{"schema":"bymsg/v1","Hero":{"z":"uint32","a":"string","big":{"type":"uint64","comment":"18446744073709551615"}}}`},
		{"game.bmsg.yaml", "schema: bymsg/v1\n# retained comment\nmessages:\n  Hero:\n    fields:\n      z:\n        type: uint32\n        tag: 1\n      a:\n        type: string\n        tag: 2\n"},
		{"game.bmsg", "schema: bymsg/v1\npackage: game\nmessage Hero {\nuint32 z = 1 // \"中文\" | \"English\"\nmap<string, uint32> attrs = 2\n}\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			original, err := ImportFormatForTest(tt.name, []byte(tt.input))
			if err != nil {
				t.Fatal(err)
			}
			formatted, err := Format(tt.name, []byte(tt.input), 2)
			if err != nil {
				t.Fatal(err)
			}
			again, err := Format(tt.name, formatted, 2)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(formatted, again) {
				t.Fatalf("non-idempotent: %s / %s", formatted, again)
			}
			actual, err := ImportFormatForTest(tt.name, formatted)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(original, actual) {
				t.Fatalf("schema changed: %s", formatted)
			}
		})
	}
}

func ImportFormatForTest(name string, data []byte) (*Schema, error) {
	ext := "bmsg"
	if name == "game.bmsg.json" {
		ext = "json"
	}
	if name == "game.bmsg.yaml" {
		ext = "yaml"
	}
	return Import(ext, data, nil)
}

func TestFormatJSONKeepsOrderAndIntegerLexemes(t *testing.T) {
	input := []byte(`{"z":18446744073709551615,"a":9007199254740993}`)
	actual, err := Format("x.bmsg.json", input, 2)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(actual, []byte("18446744073709551615")) || !bytes.Contains(actual, []byte("9007199254740993")) || bytes.Index(actual, []byte(`"z"`)) > bytes.Index(actual, []byte(`"a"`)) {
		t.Fatalf("values/order changed: %s", actual)
	}
}

func TestFormatRejectsInvalidDocuments(t *testing.T) {
	for _, tt := range []struct{ name, input string }{{"x.json", "{bad"}, {"x.yaml", "a: ["}, {"x.yaml", "a: 1\n---\nb: 2"}, {"x.bmsg", "message Hero {"}, {"x.proto", "message Hero {}"}} {
		if _, err := Format(tt.name, []byte(tt.input), 2); err == nil {
			t.Fatalf("accepted invalid %s: %s", tt.name, tt.input)
		}
	}
}

func TestEditorCheckRejectsNullMessagesAndFields(t *testing.T) {
	for _, input := range []string{
		`{"schema":"bymsg/v1","Hero":null}`,
		`{"schema":"bymsg/v1","messages":{"Hero":{"fields":{"id":null}}}}`,
	} {
		if _, err := ParseJSON([]byte(input)); err == nil {
			t.Fatalf("accepted null schema entry: %s", input)
		}
	}
	if _, err := ParseYAML([]byte("schema: bymsg/v1\nmessages:\n  Hero: null\n")); err == nil {
		t.Fatal("accepted null YAML message")
	}
}
