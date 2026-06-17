package tsgen

import (
	"fmt"
	"strings"

	"github.com/neko233-com/bytemsg233/pkg/codegen"
	"github.com/neko233-com/bytemsg233/pkg/i18n"
	"github.com/neko233-com/bytemsg233/pkg/schema"
)

type Generator struct{}

func New() *Generator { return &Generator{} }

func (g *Generator) Name() string          { return "typescript" }
func (g *Generator) FileExtension() string { return ".ts" }

func (g *Generator) Generate(s *schema.Schema, options *codegen.GenerateOptions) ([]*codegen.GeneratedFile, error) {
	var buf strings.Builder

	prevLocale := i18n.GetLocale()
	if options != nil && options.Locale != "" {
		i18n.SetLocale(options.Locale)
		defer i18n.SetLocale(prevLocale)
	}

	if len(s.Messages) > 0 {
		buf.WriteString("class ByteMsgObjectPool<T extends { reset(): void }> {\n")
		buf.WriteString("\tprivate readonly items: T[] = [];\n\n")
		buf.WriteString("\tconstructor(private readonly factory: () => T) {}\n\n")
		buf.WriteString("\tacquire(): T {\n")
		buf.WriteString("\t\tconst item = this.items.pop();\n")
		buf.WriteString("\t\treturn item ?? this.factory();\n")
		buf.WriteString("\t}\n\n")
		buf.WriteString("\trelease(item: T): void {\n")
		buf.WriteString("\t\titem.reset();\n")
		buf.WriteString("\t\tthis.items.push(item);\n")
		buf.WriteString("\t}\n")
		buf.WriteString("}\n\n")
	}
	buf.WriteString(fmt.Sprintf("export const ByteMsgProtocolVersion = %d;\n", s.ProtocolVersion))
	buf.WriteString("export function getByteMsg233ProtocolVersion(): number { return ByteMsgProtocolVersion; }\n")
	buf.WriteString("\n")

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
	if enum.Description != nil {
		desc := i18n.GetDescription(enum.Description.Zh, enum.Description.En)
		if desc != "" {
			buf.WriteString(fmt.Sprintf("/** %s */\n", desc))
		}
	}

	buf.WriteString(fmt.Sprintf("export enum %s {\n", name))
	values := codegen.SortedEnumValues(enum)
	for _, value := range values {
		buf.WriteString(fmt.Sprintf("\t%s = %d,\n", codegen.ToPascalCase(value.Name), value.Value))
	}
	buf.WriteString("}\n\n")

	buf.WriteString(fmt.Sprintf("export namespace %s {\n", name))
	buf.WriteString(fmt.Sprintf("\texport function fromValue(value: number): %s {\n", name))
	buf.WriteString("\t\tswitch (value) {\n")
	for _, value := range values {
		buf.WriteString(fmt.Sprintf("\t\t\tcase %d:\n", value.Value))
		buf.WriteString(fmt.Sprintf("\t\t\t\treturn %s.%s;\n", name, codegen.ToPascalCase(value.Name)))
	}
	buf.WriteString("\t\t\tdefault:\n")
	buf.WriteString(fmt.Sprintf("\t\t\t\tthrow new Error(`Unknown %s value: ${value}`);\n", name))
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t}\n")
	buf.WriteString("}\n")
}

func (g *Generator) generateClass(buf *strings.Builder, s *schema.Schema, name string, msg *schema.Message) {
	if msg.Description != nil {
		desc := i18n.GetDescription(msg.Description.Zh, msg.Description.En)
		if desc != "" {
			buf.WriteString(fmt.Sprintf("/** %s */\n", desc))
		}
	}

	buf.WriteString(fmt.Sprintf("export class %s {\n", name))
	buf.WriteString(fmt.Sprintf("\tprivate static readonly pool = new ByteMsgObjectPool<%s>(() => new %s());\n\n", name, name))

	for _, fieldName := range codegen.SortedFieldNames(msg) {
		field := msg.Fields[fieldName]
		if field.Description != nil {
			desc := i18n.GetDescription(field.Description.Zh, field.Description.En)
			if desc != "" {
				buf.WriteString(fmt.Sprintf("\t/** %s */\n", desc))
			}
		}
		buf.WriteString(fmt.Sprintf("\t%s: %s = %s;\n", fieldName, g.mapType(field.Type), g.defaultValueExpr(s, field.Type)))
	}

	buf.WriteString("\n")
	buf.WriteString(fmt.Sprintf("\tconstructor(init?: Partial<%s>) {\n", name))
	buf.WriteString("\t\tif (init) {\n")
	buf.WriteString("\t\t\tObject.assign(this, init);\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t}\n\n")

	buf.WriteString(fmt.Sprintf("\tstatic acquire(init?: Partial<%s>): %s {\n", name, name))
	buf.WriteString(fmt.Sprintf("\t\tconst value = %s.pool.acquire();\n", name))
	buf.WriteString("\t\tif (init) {\n")
	buf.WriteString("\t\t\tObject.assign(value, init);\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t\treturn value;\n")
	buf.WriteString("\t}\n\n")

	buf.WriteString("\trelease(): void {\n")
	buf.WriteString(fmt.Sprintf("\t\t%s.pool.release(this);\n", name))
	buf.WriteString("\t}\n\n")

	buf.WriteString("\treset(): void {\n")
	for _, fieldName := range codegen.SortedFieldNames(msg) {
		field := msg.Fields[fieldName]
		buf.WriteString(fmt.Sprintf("\t\tthis.%s = %s;\n", fieldName, g.defaultValueExpr(s, field.Type)))
	}
	buf.WriteString("\t}\n")
	buf.WriteString("}\n")
}

func (g *Generator) mapType(schemaType string) string {
	switch schemaType {
	case "bool":
		return "boolean"
	case "int32", "int64", "uint32", "uint64", "float32", "float64":
		return "number"
	case "string":
		return "string"
	case "bytes":
		return "Uint8Array"
	default:
		if strings.HasPrefix(schemaType, "list<") {
			inner := strings.TrimPrefix(schemaType, "list<")
			inner = strings.TrimSuffix(inner, ">")
			return fmt.Sprintf("%s[]", g.mapType(inner))
		}
		if strings.HasPrefix(schemaType, "map<") {
			inner := strings.TrimPrefix(schemaType, "map<")
			inner = strings.TrimSuffix(inner, ">")
			parts := strings.SplitN(inner, ",", 2)
			if len(parts) == 2 {
				keyType := g.mapType(strings.TrimSpace(parts[0]))
				valueType := g.mapType(strings.TrimSpace(parts[1]))
				return fmt.Sprintf("Record<%s, %s>", keyType, valueType)
			}
		}
		return schemaType
	}
}

func (g *Generator) defaultValueExpr(s *schema.Schema, schemaType string) string {
	switch schemaType {
	case "bool":
		return "false"
	case "int32", "int64", "uint32", "uint64", "float32", "float64":
		return "0"
	case "string":
		return "\"\""
	case "bytes":
		return "new Uint8Array(0)"
	default:
		if strings.HasPrefix(schemaType, "list<") {
			return "[]"
		}
		if strings.HasPrefix(schemaType, "map<") {
			return "{}"
		}
		if enum, ok := s.Enums[schemaType]; ok {
			if value, exists := codegen.DefaultEnumValue(enum); exists {
				return fmt.Sprintf("%s.%s", schemaType, codegen.ToPascalCase(value.Name))
			}
			return fmt.Sprintf("0 as %s", schemaType)
		}
		return fmt.Sprintf("undefined as unknown as %s", schemaType)
	}
}

func init() {
	codegen.Register(New())
}
