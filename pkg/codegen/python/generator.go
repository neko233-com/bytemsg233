package pygen

import (
	"fmt"
	"strings"

	"github.com/neko233-com/bytemsg233/pkg/codegen"
	"github.com/neko233-com/bytemsg233/pkg/i18n"
	"github.com/neko233-com/bytemsg233/pkg/schema"
)

type Generator struct{}

func New() *Generator { return &Generator{} }

func (g *Generator) Name() string          { return "python" }
func (g *Generator) FileExtension() string { return ".py" }

func (g *Generator) Generate(s *schema.Schema, options *codegen.GenerateOptions) ([]*codegen.GeneratedFile, error) {
	var buf strings.Builder

	prevLocale := i18n.GetLocale()
	if options != nil && options.Locale != "" {
		i18n.SetLocale(options.Locale)
		defer i18n.SetLocale(prevLocale)
	}

	buf.WriteString("from dataclasses import dataclass, field\n")
	buf.WriteString("from enum import IntEnum\n")
	buf.WriteString("from typing import ClassVar, Dict, List\n\n")

	for _, name := range codegen.SortedEnumNames(s) {
		g.generateEnum(&buf, name, s.Enums[name])
		buf.WriteString("\n")
	}

	for _, name := range codegen.SortedMessageNames(s) {
		g.generateClass(&buf, s, name, s.Messages[name])
		buf.WriteString("\n")
	}

	return []*codegen.GeneratedFile{
		{Path: "ByteMsg233_Export" + g.FileExtension(), Content: []byte(buf.String())},
	}, nil
}

func (g *Generator) generateEnum(buf *strings.Builder, name string, enum *schema.Enum) {
	buf.WriteString(fmt.Sprintf("class %s(IntEnum):\n", name))
	if enum.Description != nil {
		desc := i18n.GetDescription(enum.Description.Zh, enum.Description.En)
		if desc != "" {
			buf.WriteString(fmt.Sprintf("\t\"\"\"%s\"\"\"\n", desc))
		}
	}
	for _, value := range codegen.SortedEnumValues(enum) {
		buf.WriteString(fmt.Sprintf("\t%s = %d\n", value.Name, value.Value))
	}
	buf.WriteString("\n")
	buf.WriteString("\t@classmethod\n")
	buf.WriteString(fmt.Sprintf("\tdef from_value(cls, value: int) -> \"%s\":\n", name))
	buf.WriteString("\t\treturn cls(value)\n")
}

func (g *Generator) generateClass(buf *strings.Builder, s *schema.Schema, name string, msg *schema.Message) {
	buf.WriteString("@dataclass\n")
	buf.WriteString(fmt.Sprintf("class %s:\n", name))
	if msg.Description != nil {
		desc := i18n.GetDescription(msg.Description.Zh, msg.Description.En)
		if desc != "" {
			buf.WriteString(fmt.Sprintf("\t\"\"\"%s\"\"\"\n", desc))
		}
	}

	for _, fieldName := range codegen.SortedFieldNames(msg) {
		fieldDef := msg.Fields[fieldName]
		if fieldDef.Description != nil {
			desc := i18n.GetDescription(fieldDef.Description.Zh, fieldDef.Description.En)
			if desc != "" {
				buf.WriteString(fmt.Sprintf("\t# %s\n", desc))
			}
		}
		buf.WriteString(fmt.Sprintf("\t%s: %s = %s\n", fieldName, g.mapType(fieldDef.Type), g.defaultValueExpr(s, fieldDef.Type)))
	}
	buf.WriteString(fmt.Sprintf("\t_pool: ClassVar[List[\"%s\"]] = []\n\n", name))

	buf.WriteString("\t@classmethod\n")
	buf.WriteString(fmt.Sprintf("\tdef acquire(cls) -> \"%s\":\n", name))
	buf.WriteString("\t\tif cls._pool:\n")
	buf.WriteString("\t\t\treturn cls._pool.pop()\n")
	buf.WriteString("\t\treturn cls()\n\n")

	buf.WriteString("\tdef release(self) -> None:\n")
	buf.WriteString("\t\tself.reset()\n")
	buf.WriteString("\t\tself.__class__._pool.append(self)\n\n")

	buf.WriteString("\tdef reset(self) -> None:\n")
	for _, fieldName := range codegen.SortedFieldNames(msg) {
		fieldDef := msg.Fields[fieldName]
		buf.WriteString(fmt.Sprintf("\t\tself.%s = %s\n", fieldName, g.resetValueExpr(s, fieldDef.Type)))
	}
}

func (g *Generator) mapType(schemaType string) string {
	switch schemaType {
	case "bool":
		return "bool"
	case "int32", "int64", "uint32", "uint64":
		return "int"
	case "float32", "float64":
		return "float"
	case "string":
		return "str"
	case "bytes":
		return "bytes"
	default:
		if strings.HasPrefix(schemaType, "list<") {
			inner := strings.TrimPrefix(schemaType, "list<")
			inner = strings.TrimSuffix(inner, ">")
			return fmt.Sprintf("List[%s]", g.mapType(inner))
		}
		if strings.HasPrefix(schemaType, "map<") {
			inner := strings.TrimPrefix(schemaType, "map<")
			inner = strings.TrimSuffix(inner, ">")
			parts := strings.SplitN(inner, ",", 2)
			if len(parts) == 2 {
				keyType := g.mapType(strings.TrimSpace(parts[0]))
				valueType := g.mapType(strings.TrimSpace(parts[1]))
				return fmt.Sprintf("Dict[%s, %s]", keyType, valueType)
			}
		}
		return schemaType
	}
}

func (g *Generator) defaultValueExpr(s *schema.Schema, schemaType string) string {
	switch schemaType {
	case "bool":
		return "False"
	case "int32", "int64", "uint32", "uint64":
		return "0"
	case "float32", "float64":
		return "0.0"
	case "string":
		return "\"\""
	case "bytes":
		return "b\"\""
	default:
		if strings.HasPrefix(schemaType, "list<") {
			return "field(default_factory=list)"
		}
		if strings.HasPrefix(schemaType, "map<") {
			return "field(default_factory=dict)"
		}
		if enum, ok := s.Enums[schemaType]; ok {
			if value, exists := codegen.DefaultEnumValue(enum); exists {
				return fmt.Sprintf("%s.%s", schemaType, value.Name)
			}
		}
		return "None"
	}
}

func (g *Generator) resetValueExpr(s *schema.Schema, schemaType string) string {
	switch schemaType {
	case "bool":
		return "False"
	case "int32", "int64", "uint32", "uint64":
		return "0"
	case "float32", "float64":
		return "0.0"
	case "string":
		return "\"\""
	case "bytes":
		return "b\"\""
	default:
		if strings.HasPrefix(schemaType, "list<") {
			return "[]"
		}
		if strings.HasPrefix(schemaType, "map<") {
			return "{}"
		}
		if enum, ok := s.Enums[schemaType]; ok {
			if value, exists := codegen.DefaultEnumValue(enum); exists {
				return fmt.Sprintf("%s.%s", schemaType, value.Name)
			}
		}
		return "None"
	}
}

func init() {
	codegen.Register(New())
}
