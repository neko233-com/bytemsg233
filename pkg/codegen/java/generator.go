package javagen

import (
	"fmt"
	"strings"

	"github.com/neko233-com/bytemsg233/pkg/codegen"
	"github.com/neko233-com/bytemsg233/pkg/i18n"
	"github.com/neko233-com/bytemsg233/pkg/schema"
)

type Generator struct{}

func New() *Generator { return &Generator{} }

func (g *Generator) Name() string          { return "java" }
func (g *Generator) FileExtension() string { return ".java" }

func (g *Generator) Generate(s *schema.Schema, options *codegen.GenerateOptions) ([]*codegen.GeneratedFile, error) {
	prevLocale := i18n.GetLocale()
	if options != nil && options.Locale != "" {
		i18n.SetLocale(options.Locale)
		defer i18n.SetLocale(prevLocale)
	}

	files := make([]*codegen.GeneratedFile, 0, len(s.Enums)+len(s.Messages)+1)
	{
		var buf strings.Builder
		g.writePackage(&buf, s.Package)
		g.generateProtocolInfo(&buf, s)
		files = append(files, &codegen.GeneratedFile{
			Path:    "ByteMsgProtocolInfo" + g.FileExtension(),
			Content: []byte(buf.String()),
		})
	}

	for _, name := range codegen.SortedEnumNames(s) {
		var buf strings.Builder
		g.writePackage(&buf, s.Package)
		g.generateEnum(&buf, name, s.Enums[name])
		files = append(files, &codegen.GeneratedFile{
			Path:    name + g.FileExtension(),
			Content: []byte(buf.String()),
		})
	}

	for _, name := range codegen.SortedMessageNames(s) {
		var buf strings.Builder
		g.writePackage(&buf, s.Package)
		g.writeMessageImports(&buf)
		g.generateClass(&buf, s, name, s.Messages[name])
		files = append(files, &codegen.GeneratedFile{
			Path:    name + g.FileExtension(),
			Content: []byte(buf.String()),
		})
	}

	return files, nil
}

func (g *Generator) writePackage(buf *strings.Builder, packageName string) {
	if packageName != "" {
		buf.WriteString(fmt.Sprintf("package %s;\n\n", packageName))
	}
}

func (g *Generator) writeMessageImports(buf *strings.Builder) {
	buf.WriteString("import java.util.ArrayDeque;\n")
	buf.WriteString("import java.util.ArrayList;\n")
	buf.WriteString("import java.util.Deque;\n")
	buf.WriteString("import java.util.HashMap;\n")
	buf.WriteString("import java.util.List;\n")
	buf.WriteString("import java.util.Map;\n\n")
}

func (g *Generator) generateProtocolInfo(buf *strings.Builder, s *schema.Schema) {
	buf.WriteString("public final class ByteMsgProtocolInfo {\n")
	buf.WriteString("\tprivate ByteMsgProtocolInfo() {}\n\n")
	buf.WriteString(fmt.Sprintf("\tpublic static final long VERSION = %dL;\n", s.ProtocolVersion))
	buf.WriteString("\tpublic static long getByteMsg233ProtocolVersion() { return VERSION; }\n")
	buf.WriteString("}\n")
}

func (g *Generator) generateEnum(buf *strings.Builder, name string, enum *schema.Enum) {
	if enum.Description != nil {
		desc := i18n.GetDescription(enum.Description.Zh, enum.Description.En)
		if desc != "" {
			buf.WriteString(fmt.Sprintf("/** %s */\n", desc))
		}
	}

	values := codegen.SortedEnumValues(enum)
	buf.WriteString(fmt.Sprintf("public enum %s {\n", name))
	for i, value := range values {
		suffix := ","
		if i == len(values)-1 {
			suffix = ";"
		}
		buf.WriteString(fmt.Sprintf("\t%s(%d)%s\n", value.Name, value.Value, suffix))
	}
	buf.WriteString("\n")
	buf.WriteString("\tprivate final int value;\n\n")
	buf.WriteString(fmt.Sprintf("\t%s(int value) {\n", name))
	buf.WriteString("\t\tthis.value = value;\n")
	buf.WriteString("\t}\n\n")
	buf.WriteString("\tpublic int getValue() {\n")
	buf.WriteString("\t\treturn value;\n")
	buf.WriteString("\t}\n\n")
	buf.WriteString("\tpublic static boolean isDefined(int value) {\n")
	buf.WriteString("\t\tswitch (value) {\n")
	for _, value := range values {
		buf.WriteString(fmt.Sprintf("\t\t\tcase %d:\n", value.Value))
	}
	buf.WriteString("\t\t\t\treturn true;\n")
	buf.WriteString("\t\t\tdefault:\n")
	buf.WriteString("\t\t\t\treturn false;\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t}\n\n")
	buf.WriteString(fmt.Sprintf("\tpublic static %s fromValue(int value) {\n", name))
	buf.WriteString("\t\tswitch (value) {\n")
	for _, value := range values {
		buf.WriteString(fmt.Sprintf("\t\t\tcase %d:\n", value.Value))
		buf.WriteString(fmt.Sprintf("\t\t\t\treturn %s;\n", value.Name))
	}
	buf.WriteString("\t\t\tdefault:\n")
	buf.WriteString(fmt.Sprintf("\t\t\t\tthrow new IllegalArgumentException(\"Unknown %s value: \" + value);\n", name))
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

	buf.WriteString(fmt.Sprintf("public class %s {\n", name))
	buf.WriteString(fmt.Sprintf("\tprivate static final Deque<%s> POOL = new ArrayDeque<>();\n\n", name))

	for _, fieldName := range codegen.SortedFieldNames(msg) {
		field := msg.Fields[fieldName]
		if field.Description != nil {
			desc := i18n.GetDescription(field.Description.Zh, field.Description.En)
			if desc != "" {
				buf.WriteString(fmt.Sprintf("\t/** %s */\n", desc))
			}
		}
		javaType := g.mapType(s, field.Type, false)
		javaName := codegen.ToCamelCase(fieldName)
		buf.WriteString(fmt.Sprintf("\tprivate %s %s = %s;\n", javaType, javaName, g.defaultValueExpr(s, field.Type)))
	}

	buf.WriteString("\n")
	buf.WriteString(fmt.Sprintf("\tpublic static %s acquire() {\n", name))
	buf.WriteString(fmt.Sprintf("\t\treturn POOL.isEmpty() ? new %s() : POOL.pop();\n", name))
	buf.WriteString("\t}\n\n")

	buf.WriteString(fmt.Sprintf("\tpublic static void release(%s value) {\n", name))
	buf.WriteString("\t\tif (value == null) {\n")
	buf.WriteString("\t\t\treturn;\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t\tvalue.reset();\n")
	buf.WriteString("\t\tPOOL.push(value);\n")
	buf.WriteString("\t}\n\n")

	buf.WriteString("\tpublic void release() {\n")
	buf.WriteString("\t\trelease(this);\n")
	buf.WriteString("\t}\n\n")

	buf.WriteString("\tpublic void reset() {\n")
	for _, fieldName := range codegen.SortedFieldNames(msg) {
		field := msg.Fields[fieldName]
		javaName := codegen.ToCamelCase(fieldName)
		buf.WriteString(fmt.Sprintf("\t\tthis.%s = %s;\n", javaName, g.defaultValueExpr(s, field.Type)))
	}
	buf.WriteString("\t}\n\n")

	for _, fieldName := range codegen.SortedFieldNames(msg) {
		field := msg.Fields[fieldName]
		javaType := g.mapType(s, field.Type, false)
		fieldCamel := codegen.ToCamelCase(fieldName)
		fieldPascal := codegen.ToPascalCase(fieldName)

		buf.WriteString(fmt.Sprintf("\tpublic %s get%s() {\n", javaType, fieldPascal))
		buf.WriteString(fmt.Sprintf("\t\treturn %s;\n", fieldCamel))
		buf.WriteString("\t}\n\n")

		buf.WriteString(fmt.Sprintf("\tpublic void set%s(%s %s) {\n", fieldPascal, javaType, fieldCamel))
		buf.WriteString(fmt.Sprintf("\t\tthis.%s = %s;\n", fieldCamel, fieldCamel))
		buf.WriteString("\t}\n\n")
	}

	buf.WriteString("}\n")
}

func (g *Generator) mapType(s *schema.Schema, schemaType string, boxed bool) string {
	switch schemaType {
	case "bool":
		if boxed {
			return "Boolean"
		}
		return "boolean"
	case "int32", "uint32":
		if boxed {
			return "Integer"
		}
		return "int"
	case "int64", "uint64":
		if boxed {
			return "Long"
		}
		return "long"
	case "float32":
		if boxed {
			return "Float"
		}
		return "float"
	case "float64":
		if boxed {
			return "Double"
		}
		return "double"
	case "string":
		return "String"
	case "bytes":
		return "byte[]"
	default:
		if strings.HasPrefix(schemaType, "list<") {
			inner := strings.TrimPrefix(schemaType, "list<")
			inner = strings.TrimSuffix(inner, ">")
			return fmt.Sprintf("List<%s>", g.mapType(s, inner, true))
		}
		if strings.HasPrefix(schemaType, "map<") {
			inner := strings.TrimPrefix(schemaType, "map<")
			inner = strings.TrimSuffix(inner, ">")
			parts := strings.SplitN(inner, ",", 2)
			if len(parts) == 2 {
				keyType := g.mapType(s, strings.TrimSpace(parts[0]), true)
				valueType := g.mapType(s, strings.TrimSpace(parts[1]), true)
				return fmt.Sprintf("Map<%s, %s>", keyType, valueType)
			}
		}
		return schemaType
	}
}

func (g *Generator) defaultValueExpr(s *schema.Schema, schemaType string) string {
	switch schemaType {
	case "bool":
		return "false"
	case "int32", "uint32", "int64", "uint64":
		return "0"
	case "float32":
		return "0.0f"
	case "float64":
		return "0.0d"
	case "string":
		return "\"\""
	case "bytes":
		return "new byte[0]"
	default:
		if strings.HasPrefix(schemaType, "list<") {
			return "new ArrayList<>()"
		}
		if strings.HasPrefix(schemaType, "map<") {
			return "new HashMap<>()"
		}
		if enum, ok := s.Enums[schemaType]; ok {
			if value, exists := codegen.DefaultEnumValue(enum); exists {
				return fmt.Sprintf("%s.%s", schemaType, value.Name)
			}
		}
		return "null"
	}
}

func init() {
	codegen.Register(New())
}
