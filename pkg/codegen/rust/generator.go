package rustgen

import (
	"fmt"
	"strings"

	"github.com/neko233-com/bytemsg233/pkg/codegen"
	"github.com/neko233-com/bytemsg233/pkg/i18n"
	"github.com/neko233-com/bytemsg233/pkg/schema"
)

type Generator struct{}

func New() *Generator { return &Generator{} }

func (g *Generator) Name() string          { return "rust" }
func (g *Generator) FileExtension() string { return ".rs" }

func (g *Generator) Generate(s *schema.Schema, options *codegen.GenerateOptions) ([]*codegen.GeneratedFile, error) {
	var buf strings.Builder

	prevLocale := i18n.GetLocale()
	if options != nil && options.Locale != "" {
		i18n.SetLocale(options.Locale)
		defer i18n.SetLocale(prevLocale)
	}

	if len(s.Messages) > 0 {
		buf.WriteString("use std::collections::HashMap;\n\n")
		buf.WriteString("pub trait ByteMsgResettable {\n")
		buf.WriteString("    fn reset(&mut self);\n")
		buf.WriteString("}\n\n")
		buf.WriteString("pub struct ByteMsgPool<T: ByteMsgResettable + Default> {\n")
		buf.WriteString("    items: Vec<T>,\n")
		buf.WriteString("}\n\n")
		buf.WriteString("impl<T: ByteMsgResettable + Default> ByteMsgPool<T> {\n")
		buf.WriteString("    pub fn new() -> Self { Self { items: Vec::new() } }\n")
		buf.WriteString("    pub fn acquire(&mut self) -> T { self.items.pop().unwrap_or_default() }\n")
		buf.WriteString("    pub fn release(&mut self, mut value: T) { value.reset(); self.items.push(value); }\n")
		buf.WriteString("}\n\n")
	}
	buf.WriteString(fmt.Sprintf("pub const BYTE_MSG_PROTOCOL_VERSION: u64 = %d;\n", s.ProtocolVersion))
	buf.WriteString("pub fn get_bytemsg233_protocol_version() -> u64 { BYTE_MSG_PROTOCOL_VERSION }\n")
	buf.WriteString("\n")

	for _, name := range codegen.SortedEnumNames(s) {
		g.generateEnum(&buf, name, s.Enums[name])
		buf.WriteString("\n")
	}

	for _, name := range codegen.SortedMessageNames(s) {
		g.generateStruct(&buf, s, name, s.Messages[name])
		buf.WriteString("\n")
	}

	return []*codegen.GeneratedFile{
		{Path: "ByteMsg233_Export" + g.FileExtension(), Content: []byte(buf.String())},
	}, nil
}

func (g *Generator) generateEnum(buf *strings.Builder, name string, enum *schema.Enum) {
	if enum.Description != nil {
		desc := i18n.GetDescription(enum.Description.Zh, enum.Description.En)
		if desc != "" {
			buf.WriteString(fmt.Sprintf("/// %s\n", desc))
		}
	}

	buf.WriteString("#[derive(Debug, Clone, Copy, PartialEq, Eq)]\n")
	buf.WriteString("#[repr(i32)]\n")
	buf.WriteString(fmt.Sprintf("pub enum %s {\n", name))
	values := codegen.SortedEnumValues(enum)
	for _, value := range values {
		buf.WriteString(fmt.Sprintf("    %s = %d,\n", codegen.ToPascalCase(value.Name), value.Value))
	}
	buf.WriteString("}\n\n")

	defaultVariant := "None"
	if len(values) > 0 {
		defaultVariant = codegen.ToPascalCase(values[0].Name)
	}

	buf.WriteString(fmt.Sprintf("impl Default for %s {\n", name))
	buf.WriteString(fmt.Sprintf("    fn default() -> Self { %s::%s }\n", name, defaultVariant))
	buf.WriteString("}\n\n")

	buf.WriteString(fmt.Sprintf("impl %s {\n", name))
	buf.WriteString("    pub fn from_value(value: i32) -> Option<Self> {\n")
	buf.WriteString("        match value {\n")
	for _, value := range values {
		buf.WriteString(fmt.Sprintf("            %d => Some(%s::%s),\n", value.Value, name, codegen.ToPascalCase(value.Name)))
	}
	buf.WriteString("            _ => None,\n")
	buf.WriteString("        }\n")
	buf.WriteString("    }\n\n")
	buf.WriteString("    pub fn is_defined(value: i32) -> bool { Self::from_value(value).is_some() }\n")
	buf.WriteString("}\n")
}

func (g *Generator) generateStruct(buf *strings.Builder, s *schema.Schema, name string, msg *schema.Message) {
	if msg.Description != nil {
		desc := i18n.GetDescription(msg.Description.Zh, msg.Description.En)
		if desc != "" {
			buf.WriteString(fmt.Sprintf("/// %s\n", desc))
		}
	}

	buf.WriteString("#[derive(Debug, Clone, Default)]\n")
	buf.WriteString(fmt.Sprintf("pub struct %s {\n", name))
	for _, fieldName := range codegen.SortedFieldNames(msg) {
		field := msg.Fields[fieldName]
		if field.Description != nil {
			desc := i18n.GetDescription(field.Description.Zh, field.Description.En)
			if desc != "" {
				buf.WriteString(fmt.Sprintf("    /// %s\n", desc))
			}
		}
		buf.WriteString(fmt.Sprintf("    pub %s: %s,\n", g.fieldName(fieldName), g.mapType(field.Type)))
	}
	buf.WriteString("}\n\n")

	buf.WriteString(fmt.Sprintf("impl ByteMsgResettable for %s {\n", name))
	buf.WriteString("    fn reset(&mut self) {\n")
	buf.WriteString("        *self = Self::default();\n")
	buf.WriteString("    }\n")
	buf.WriteString("}\n")
}

func (g *Generator) fieldName(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, "-", "_"))
}

func (g *Generator) mapType(schemaType string) string {
	switch schemaType {
	case "bool":
		return "bool"
	case "int32":
		return "i32"
	case "int64":
		return "i64"
	case "uint32":
		return "u32"
	case "uint64":
		return "u64"
	case "float32":
		return "f32"
	case "float64":
		return "f64"
	case "string":
		return "String"
	case "bytes":
		return "Vec<u8>"
	default:
		if strings.HasPrefix(schemaType, "list<") {
			inner := strings.TrimPrefix(schemaType, "list<")
			inner = strings.TrimSuffix(inner, ">")
			return fmt.Sprintf("Vec<%s>", g.mapType(inner))
		}
		if strings.HasPrefix(schemaType, "map<") {
			inner := strings.TrimPrefix(schemaType, "map<")
			inner = strings.TrimSuffix(inner, ">")
			parts := strings.SplitN(inner, ",", 2)
			if len(parts) == 2 {
				keyType := g.mapType(strings.TrimSpace(parts[0]))
				valueType := g.mapType(strings.TrimSpace(parts[1]))
				return fmt.Sprintf("HashMap<%s, %s>", keyType, valueType)
			}
		}
		return schemaType
	}
}

func init() {
	codegen.Register(New())
}
