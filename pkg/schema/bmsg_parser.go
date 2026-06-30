package schema

import (
	"bufio"
	"fmt"
	"strings"
)

// ParseBmsg parses a .bmsg file into a Schema
func ParseBmsg(data []byte) (*Schema, error) {
	s := &Schema{
		Messages: make(map[string]*Message),
		Enums:    make(map[string]*Enum),
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	var currentBlock string
	var currentName string
	var currentMessage *Message
	var currentEnum *Enum

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			continue
		}

		if strings.HasPrefix(trimmed, "schema:") {
			s.Version = strings.TrimSpace(strings.TrimPrefix(trimmed, "schema:"))
			continue
		}
		if strings.HasPrefix(trimmed, "package:") {
			s.Package = strings.TrimSpace(strings.TrimPrefix(trimmed, "package:"))
			continue
		}

		// Enum block
		if strings.HasPrefix(trimmed, "enum ") && strings.HasSuffix(trimmed, "{") {
			namePart := strings.TrimPrefix(trimmed, "enum ")
			namePart = strings.TrimSuffix(namePart, "{")
			currentName = strings.TrimSpace(namePart)
			currentEnum = &Enum{
				Values: make(map[string]int),
			}
			currentBlock = "enum"
			continue
		}

		// Message block
		if strings.HasPrefix(trimmed, "message ") && strings.HasSuffix(trimmed, "{") {
			namePart := strings.TrimPrefix(trimmed, "message ")
			namePart = strings.TrimSuffix(namePart, "{")
			currentName = strings.TrimSpace(namePart)
			currentMessage = &Message{
				Fields: make(map[string]*Field),
			}
			currentBlock = "message"
			continue
		}

		// Closing brace
		if trimmed == "}" {
			if currentBlock == "enum" && currentEnum != nil {
				s.Enums[currentName] = currentEnum
				currentEnum = nil
			} else if currentBlock == "message" && currentMessage != nil {
				s.Messages[currentName] = currentMessage
				currentMessage = nil
			}
			currentBlock = ""
			continue
		}

		// Inside enum block
		if currentBlock == "enum" && currentEnum != nil {
			parts := strings.SplitN(trimmed, "=", 2)
			if len(parts) == 2 {
				name := strings.TrimSpace(parts[0])
				var value int
				_, _ = fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &value)
				currentEnum.Values[name] = value
			}
			continue
		}

		// Inside message block
		if currentBlock == "message" && currentMessage != nil {
			name, field := parseBmsgField(trimmed)
			if field != nil {
				currentMessage.Fields[name] = field
			}
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan bmsg: %w", err)
	}
	if err := validate(s); err != nil {
		return nil, err
	}
	return s, nil
}

func parseBmsgField(line string) (string, *Field) {
	var desc *Description
	mainPart := line
	if idx := strings.Index(line, "//"); idx >= 0 {
		mainPart = line[:idx]
		descPart := strings.TrimSpace(line[idx+2:])
		desc = parseInlineDescription(descPart)
	}

	trimmed := strings.TrimSpace(mainPart)
	if trimmed == "" {
		return "", nil
	}

	fieldType, rest := extractType(trimmed)
	if fieldType == "" {
		return "", nil
	}

	restParts := strings.Fields(rest)
	if len(restParts) < 3 {
		return "", nil
	}

	fieldName := restParts[0]
	var tag int
	_, _ = fmt.Sscanf(restParts[2], "%d", &tag)

	return fieldName, &Field{
		Type:        fieldType,
		Tag:         tag,
		Description: desc,
	}
}

func extractType(s string) (string, string) {
	// Handle generic types like map<string, string> or list<uint32>
	if strings.HasPrefix(s, "map<") || strings.HasPrefix(s, "list<") {
		end := strings.Index(s, ">")
		if end >= 0 {
			return s[:end+1], s[end+1:]
		}
	}
	// Simple type
	parts := strings.SplitN(s, " ", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return s, ""
}

func parseInlineDescription(s string) *Description {
	parts := strings.SplitN(s, "|", 2)
	if len(parts) != 2 {
		return nil
	}

	zh := strings.TrimSpace(removeQuotes(parts[0]))
	en := strings.TrimSpace(removeQuotes(parts[1]))

	if zh == "" && en == "" {
		return nil
	}

	return &Description{
		Zh: zh,
		En: en,
	}
}

func removeQuotes(s string) string {
	s = strings.ReplaceAll(s, "\"", "")
	s = strings.ReplaceAll(s, "\\", "")
	return strings.TrimSpace(s)
}
