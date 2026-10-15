package schema

import (
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"
)

// Parse parses YAML content into a Schema
func Parse(data []byte) (*Schema, error) {
	var schema Schema
	if err := yaml.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("failed to parse schema: %w", err)
	}

	if err := validate(&schema); err != nil {
		return nil, err
	}

	return &schema, nil
}

// validate performs basic validation on the schema
func validate(s *Schema) error {
	if s.Version == "" {
		return fmt.Errorf("schema version is required")
	}

	if s.Messages == nil {
		s.Messages = make(map[string]*Message)
	}

	if s.Enums == nil {
		s.Enums = make(map[string]*Enum)
	}

	for name, msg := range s.Messages {
		if msg == nil {
			return fmt.Errorf("message %s must be an object", name)
		}
		if msg.Fields == nil {
			msg.Fields = make(map[string]*Field)
		}
		for fieldName, field := range msg.Fields {
			if field == nil {
				return fmt.Errorf("message %s, field %s must be a type or object", name, fieldName)
			}
		}
		assignMissingTagsStable(msg)
		for fieldName, field := range msg.Fields {
			if field.Tag <= 0 {
				return fmt.Errorf("message %s, field %s: tag must be positive", name, fieldName)
			}
		}
	}

	return nil
}

func assignMissingTagsStable(msg *Message) {
	if msg == nil {
		return
	}

	used := make(map[int]bool)
	names := make([]string, 0, len(msg.Fields))
	for name, field := range msg.Fields {
		names = append(names, name)
		if field.Tag > 0 {
			used[field.Tag] = true
		}
		normalizeComment(field)
	}
	sort.Strings(names)

	next := 1
	for _, name := range names {
		field := msg.Fields[name]
		if field.Tag > 0 {
			continue
		}
		for used[next] {
			next++
		}
		field.Tag = next
		used[next] = true
		next++
	}
}
