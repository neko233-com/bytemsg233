package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"
)

// Format is tooling only. It preserves JSON key order and numeric lexemes:
// field order is significant when tags are assigned implicitly.
func Format(filename string, data []byte, indent int) ([]byte, error) {
	if indent < 1 || indent > 8 {
		return nil, fmt.Errorf("indent must be between 1 and 8")
	}
	format := strings.ToLower(filepath.Ext(filename))
	if format == ".bmsg" {
		if bytes.HasPrefix(bytes.TrimSpace(data), []byte("{")) {
			format = ".json"
		} else if _, err := ParseYAML(data); err == nil {
			format = ".yaml"
		}
	}
	switch format {
	case ".json":
		var out bytes.Buffer
		if err := json.Indent(&out, bytes.TrimSpace(data), "", strings.Repeat(" ", indent)); err != nil {
			return nil, err
		}
		out.WriteByte('\n')
		return out.Bytes(), nil
	case ".yaml", ".yml":
		var node yaml.Node
		decoder := yaml.NewDecoder(bytes.NewReader(data))
		if err := decoder.Decode(&node); err != nil {
			return nil, err
		}
		var extra yaml.Node
		if err := decoder.Decode(&extra); err != io.EOF {
			return nil, fmt.Errorf("expected one YAML document")
		}
		var out bytes.Buffer
		encoder := yaml.NewEncoder(&out)
		encoder.SetIndent(indent)
		if err := encoder.Encode(&node); err != nil {
			return nil, err
		}
		if err := encoder.Close(); err != nil {
			return nil, err
		}
		return out.Bytes(), nil
	case ".bmsg":
		return formatLegacy(data, indent)
	default:
		return nil, fmt.Errorf("format supports .bmsg.json, .json, .yaml, .yml and .bmsg")
	}
}

func formatLegacy(data []byte, indent int) ([]byte, error) {
	var out strings.Builder
	depth := 0
	for _, raw := range strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		var tokens []string
		comment := ""
		for i := 0; i < len(line); {
			c := line[i]
			if unicode.IsSpace(rune(c)) {
				i++
				continue
			}
			if strings.HasPrefix(line[i:], "//") {
				comment = line[i:]
				break
			}
			start := i
			if c == '"' || c == '\'' {
				i++
				closed := false
				for i < len(line) {
					if line[i] == '\\' {
						i += 2
						continue
					}
					if line[i] == c {
						i++
						closed = true
						break
					}
					i++
				}
				if !closed {
					return nil, fmt.Errorf("unterminated string")
				}
			} else if strings.ContainsRune("{}<>,=:", rune(c)) {
				i++
			} else {
				for i < len(line) && !unicode.IsSpace(rune(line[i])) && !strings.ContainsRune("{}<>,=:\"'", rune(line[i])) && !strings.HasPrefix(line[i:], "//") {
					i++
				}
			}
			tokens = append(tokens, line[start:i])
		}
		lineDepth := depth
		if len(tokens) > 0 && tokens[0] == "}" {
			lineDepth--
		}
		if lineDepth < 0 {
			return nil, fmt.Errorf("unexpected closing brace")
		}
		out.WriteString(strings.Repeat(" ", lineDepth*indent))
		for i, token := range tokens {
			if i > 0 && token != ">" && token != "," && token != ":" && token != "<" && tokens[i-1] != "<" {
				out.WriteByte(' ')
			}
			out.WriteString(token)
			if token == "{" {
				depth++
			}
			if token == "}" {
				depth--
			}
			if depth < 0 {
				return nil, fmt.Errorf("unexpected closing brace")
			}
		}
		if comment != "" {
			if len(tokens) > 0 {
				out.WriteByte(' ')
			}
			out.WriteString(comment)
		}
		out.WriteByte('\n')
	}
	if depth != 0 {
		return nil, fmt.Errorf("unclosed message or enum block")
	}
	return []byte(out.String()), nil
}
