package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseFile parses a schema file by extension.
//
// JSON is the default DSL. YAML remains supported, and legacy .bmsg syntax is
// accepted as a compatibility/export target for future tooling.
func ParseFile(path string) (*Schema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".bmsg":
		s, err := ParseJSON(data)
		if err == nil {
			return s, nil
		}
		s, err = ParseYAML(data)
		if err == nil {
			return s, nil
		}
		return ParseBmsg(data)
	case ".yaml", ".yml":
		return ParseYAML(data)
	case ".json":
		return ParseJSON(data)
	case ".toml":
		return ParseTOML(data)
	default:
		s, err := ParseJSON(data)
		if err == nil {
			return s, nil
		}
		s, err = ParseYAML(data)
		if err == nil {
			return s, nil
		}
		return ParseBmsg(data)
	}
}

// ParseYAML parses YAML content
func ParseYAML(data []byte) (*Schema, error) {
	var schema Schema
	if err := yaml.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}
	if err := validate(&schema); err != nil {
		return nil, err
	}
	return &schema, nil
}

// ParseJSON parses JSON content
func ParseJSON(data []byte) (*Schema, error) {
	native, nativeErr := parseNativeJSON(data)
	if nativeErr == nil {
		if err := validate(native); err != nil {
			return nil, err
		}
		return native, nil
	}

	var schema Schema
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}
	if len(schema.Messages) == 0 {
		return nil, nativeErr
	}
	if err := validate(&schema); err != nil {
		return nil, err
	}
	return &schema, nil
}

func parseNativeJSON(data []byte) (*Schema, error) {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}

	s := &Schema{
		Messages: make(map[string]*Message),
		Enums:    make(map[string]*Enum),
	}

	if raw, ok := root["schema"]; ok {
		_ = json.Unmarshal(raw, &s.Version)
	}
	if raw, ok := root["package"]; ok {
		_ = json.Unmarshal(raw, &s.Package)
	}
	if raw, ok := root["enums"]; ok {
		enums, err := parseNativeJSONEnums(raw)
		if err != nil {
			return nil, fmt.Errorf("parse json enums: %w", err)
		}
		s.Enums = enums
	}
	if raw, ok := root["messages"]; ok {
		messages, err := parseNativeJSONMessages(raw)
		if err != nil {
			return nil, fmt.Errorf("parse json messages: %w", err)
		}
		s.Messages = messages
	}

	for name, raw := range root {
		if isReservedNativeJSONKey(name) {
			continue
		}

		msg, err := parseNativeJSONMessage(raw)
		if err != nil {
			return nil, fmt.Errorf("parse json message %s: %w", name, err)
		}
		s.Messages[name] = msg
	}

	return s, nil
}

func parseNativeJSONMessages(raw json.RawMessage) (map[string]*Message, error) {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(raw, &root); err != nil {
		return nil, err
	}

	messages := make(map[string]*Message, len(root))
	for name, msgRaw := range root {
		msg, err := parseNativeJSONMessage(msgRaw)
		if err != nil {
			return nil, fmt.Errorf("message %s: %w", name, err)
		}
		messages[name] = msg
	}
	return messages, nil
}

func parseNativeJSONMessage(raw json.RawMessage) (*Message, error) {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, err
	}

	msg := &Message{Fields: make(map[string]*Field)}
	if rawDesc, ok := obj["description"]; ok {
		var desc Description
		if err := json.Unmarshal(rawDesc, &desc); err != nil {
			return nil, fmt.Errorf("description: %w", err)
		}
		msg.Description = &desc
	}
	if rawComment, ok := obj["comment"]; ok && msg.Description == nil {
		var comment string
		if err := json.Unmarshal(rawComment, &comment); err != nil {
			return nil, fmt.Errorf("comment: %w", err)
		}
		msg.Description = &Description{Zh: comment, En: comment}
	}
	if rawPacketID, ok := obj["packetId"]; ok {
		if err := json.Unmarshal(rawPacketID, &msg.PacketID); err != nil {
			return nil, fmt.Errorf("packetId: %w", err)
		}
	}

	if rawFields, ok := obj["fields"]; ok {
		if err := json.Unmarshal(rawFields, &msg.Fields); err != nil {
			return nil, fmt.Errorf("fields: %w", err)
		}
		assignMissingTags(msg, orderedObjectKeys(rawFields))
		return msg, nil
	}

	for _, fieldName := range orderedObjectKeys(raw) {
		if isReservedNativeJSONMessageKey(fieldName) {
			continue
		}
		rawField := obj[fieldName]
		field, err := parseNativeJSONField(rawField)
		if err != nil {
			return nil, fmt.Errorf("field %s: %w", fieldName, err)
		}
		msg.Fields[fieldName] = field
	}
	assignMissingTags(msg, orderedObjectKeys(raw))

	return msg, nil
}

func parseNativeJSONField(raw json.RawMessage) (*Field, error) {
	var field Field
	if err := json.Unmarshal(raw, &field); err == nil && field.Type != "" {
		normalizeComment(&field)
		return &field, nil
	}

	if err := json.Unmarshal(raw, &field); err == nil {
		if fieldType, ok, err := parseStructuredFieldType(raw); ok || err != nil {
			if err != nil {
				return nil, err
			}
			field.Type = fieldType
			normalizeComment(&field)
			return &field, nil
		}
	}

	var fieldType string
	if err := json.Unmarshal(raw, &fieldType); err == nil && fieldType != "" {
		return &Field{Type: fieldType}, nil
	}

	return nil, fmt.Errorf("expected field object with type/tag")
}

func parseNativeJSONEnums(raw json.RawMessage) (map[string]*Enum, error) {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(raw, &root); err != nil {
		return nil, err
	}

	enums := make(map[string]*Enum, len(root))
	for enumName, enumRaw := range root {
		enum, err := parseNativeJSONEnum(enumRaw)
		if err != nil {
			return nil, fmt.Errorf("enum %s: %w", enumName, err)
		}
		enums[enumName] = enum
	}
	return enums, nil
}

func parseNativeJSONEnum(raw json.RawMessage) (*Enum, error) {
	var names []string
	if err := json.Unmarshal(raw, &names); err == nil {
		return enumFromNames(names), nil
	}

	var values map[string]int
	if err := json.Unmarshal(raw, &values); err == nil && len(values) > 0 {
		return &Enum{Values: values}, nil
	}

	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, err
	}

	enum := &Enum{Values: make(map[string]int)}
	if rawDesc, ok := obj["description"]; ok {
		var desc Description
		if err := json.Unmarshal(rawDesc, &desc); err != nil {
			return nil, fmt.Errorf("description: %w", err)
		}
		enum.Description = &desc
	}
	if rawComment, ok := obj["comment"]; ok && enum.Description == nil {
		var comment string
		if err := json.Unmarshal(rawComment, &comment); err != nil {
			return nil, fmt.Errorf("comment: %w", err)
		}
		enum.Description = &Description{Zh: comment, En: comment}
	}
	if rawValues, ok := obj["values"]; ok {
		if err := json.Unmarshal(rawValues, &enum.Values); err == nil {
			return enum, nil
		}
		var valueNames []string
		if err := json.Unmarshal(rawValues, &valueNames); err == nil {
			enum.Values = enumFromNames(valueNames).Values
			return enum, nil
		}
		return nil, fmt.Errorf("values must be an object or string array")
	}

	for _, key := range orderedObjectKeys(raw) {
		if key == "description" || key == "comment" {
			continue
		}
		var value int
		if err := json.Unmarshal(obj[key], &value); err != nil {
			return nil, fmt.Errorf("value %s must be an integer", key)
		}
		enum.Values[key] = value
	}
	return enum, nil
}

func enumFromNames(names []string) *Enum {
	values := make(map[string]int, len(names))
	for i, name := range names {
		values[name] = i
	}
	return &Enum{Values: values}
}

func parseStructuredFieldType(raw json.RawMessage) (string, bool, error) {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		return "", false, nil
	}

	if rawList, ok := obj["list"]; ok {
		var inner string
		if err := json.Unmarshal(rawList, &inner); err != nil || inner == "" {
			return "", true, fmt.Errorf("list must be a type string")
		}
		return fmt.Sprintf("list<%s>", inner), true, nil
	}

	if rawMap, ok := obj["map"]; ok {
		keyType, valueType, err := parseMapType(rawMap)
		if err != nil {
			return "", true, err
		}
		return fmt.Sprintf("map<%s, %s>", keyType, valueType), true, nil
	}

	return "", false, nil
}

func parseMapType(raw json.RawMessage) (string, string, error) {
	var pair []string
	if err := json.Unmarshal(raw, &pair); err == nil {
		if len(pair) != 2 || pair[0] == "" || pair[1] == "" {
			return "", "", fmt.Errorf("map array must be [keyType, valueType]")
		}
		return pair[0], pair[1], nil
	}

	var obj struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(raw, &obj); err != nil {
		return "", "", fmt.Errorf("map must be [keyType, valueType] or {key,value}")
	}
	if obj.Key == "" || obj.Value == "" {
		return "", "", fmt.Errorf("map key and value are required")
	}
	return obj.Key, obj.Value, nil
}

func normalizeComment(field *Field) {
	if field == nil || field.Comment == "" || field.Description != nil {
		return
	}
	field.Description = &Description{Zh: field.Comment, En: field.Comment}
}

func assignMissingTags(msg *Message, orderedKeys []string) {
	if msg == nil {
		return
	}

	used := make(map[int]bool)
	for _, field := range msg.Fields {
		if field.Tag > 0 {
			used[field.Tag] = true
		}
		normalizeComment(field)
	}

	next := 1
	assign := func(name string) {
		field := msg.Fields[name]
		if field == nil || field.Tag > 0 {
			return
		}
		for used[next] {
			next++
		}
		field.Tag = next
		used[next] = true
		next++
	}

	for _, key := range orderedKeys {
		if isReservedNativeJSONMessageKey(key) {
			continue
		}
		assign(key)
	}
	for name := range msg.Fields {
		assign(name)
	}
}

func isReservedNativeJSONKey(key string) bool {
	switch key {
	case "schema", "$schema", "package", "namespace", "messages", "enums":
		return true
	default:
		return false
	}
}

func isReservedNativeJSONMessageKey(key string) bool {
	switch key {
	case "description", "comment", "packetId", "fields":
		return true
	default:
		return false
	}
}

func orderedObjectKeys(raw []byte) []string {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	token, err := decoder.Token()
	if err != nil {
		return nil
	}
	if delim, ok := token.(json.Delim); !ok || delim != '{' {
		return nil
	}

	var keys []string
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return keys
		}
		key, ok := token.(string)
		if !ok {
			return keys
		}
		keys = append(keys, key)
		if err := skipJSONValue(decoder); err != nil {
			return keys
		}
	}
	return keys
}

func skipJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}

	if delim, ok := token.(json.Delim); ok {
		switch delim {
		case '{':
			for decoder.More() {
				if _, err := decoder.Token(); err != nil {
					return err
				}
				if err := skipJSONValue(decoder); err != nil {
					return err
				}
			}
			_, err := decoder.Token()
			return err
		case '[':
			for decoder.More() {
				if err := skipJSONValue(decoder); err != nil {
					return err
				}
			}
			_, err := decoder.Token()
			return err
		}
	}

	if err == io.EOF {
		return nil
	}
	return nil
}

// ParseTOML parses TOML content
func ParseTOML(data []byte) (*Schema, error) {
	// TOML support: convert to YAML internally
	// Simple approach: parse TOML line by line and build Schema
	// For now, require explicit format flag
	return nil, fmt.Errorf("toml support coming soon, use .bmsg.json, .json, or .yaml")
}
