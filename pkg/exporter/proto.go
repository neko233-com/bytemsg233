package exporter

import (
	"fmt"
	"strings"

	"github.com/neko233-com/bytemsg233/pkg/codegen"
	"github.com/neko233-com/bytemsg233/pkg/schema"
)

func Proto(s *schema.Schema) ([]byte, error) {
	if s == nil {
		return nil, fmt.Errorf("schema is nil")
	}

	var buf strings.Builder
	buf.WriteString("syntax = \"proto3\";\n\n")
	if s.Package != "" {
		buf.WriteString(fmt.Sprintf("package %s;\n\n", s.Package))
	}

	if s.Version != "" {
		buf.WriteString(fmt.Sprintf("// ByteMsg233 schema: %s\n", s.Version))
	}
	if s.ProtocolVersion > 0 {
		buf.WriteString(fmt.Sprintf("// ByteMsg233 protocolVersion: %d\n", s.ProtocolVersion))
	}
	if s.Version != "" || s.ProtocolVersion > 0 {
		buf.WriteString("\n")
	}

	for _, enumName := range codegen.SortedEnumNames(s) {
		buf.WriteString(fmt.Sprintf("enum %s {\n", enumName))
		for _, value := range codegen.SortedEnumValues(s.Enums[enumName]) {
			buf.WriteString(fmt.Sprintf("  %s = %d;\n", value.Name, value.Value))
		}
		buf.WriteString("}\n\n")
	}

	for _, msgName := range codegen.SortedMessageNames(s) {
		msg := s.Messages[msgName]
		if msg.PacketID > 0 {
			buf.WriteString(fmt.Sprintf("// ByteMsg233 packetId: %d\n", msg.PacketID))
		}
		buf.WriteString(fmt.Sprintf("message %s {\n", msgName))
		for _, fieldName := range codegen.SortedFieldNames(msg) {
			field := msg.Fields[fieldName]
			fieldType, err := protoFieldType(s, field.Type)
			if err != nil {
				return nil, fmt.Errorf("%s.%s: %w", msgName, fieldName, err)
			}
			buf.WriteString(fmt.Sprintf("  %s %s = %d;\n", fieldType, fieldName, field.Tag))
		}
		buf.WriteString("}\n\n")
	}

	return []byte(buf.String()), nil
}

type protoExporter struct{}

func (protoExporter) Name() string { return "proto" }
func (protoExporter) Extensions() []string {
	return []string{".proto"}
}
func (protoExporter) Export(s *schema.Schema, _ *ExportOptions) ([]byte, error) {
	return Proto(s)
}

type protoTypeKind int

const (
	protoTypeScalar protoTypeKind = iota
	protoTypeEnum
	protoTypeMessage
	protoTypeList
	protoTypeMap
)

type protoTypeSpec struct {
	Raw   string
	Kind  protoTypeKind
	Key   *protoTypeSpec
	Value *protoTypeSpec
}

func protoFieldType(s *schema.Schema, schemaType string) (string, error) {
	spec, err := parseProtoTypeSpec(s, schemaType)
	if err != nil {
		return "", err
	}
	return renderProtoFieldType(spec)
}

func parseProtoTypeSpec(s *schema.Schema, schemaType string) (*protoTypeSpec, error) {
	schemaType = strings.TrimSpace(schemaType)
	if strings.HasPrefix(schemaType, "list<") {
		inner, ok := unwrapProtoGeneric(schemaType, "list")
		if !ok {
			return nil, fmt.Errorf("invalid list type %q", schemaType)
		}
		value, err := parseProtoTypeSpec(s, inner)
		if err != nil {
			return nil, err
		}
		return &protoTypeSpec{Raw: schemaType, Kind: protoTypeList, Value: value}, nil
	}
	if strings.HasPrefix(schemaType, "map<") {
		inner, ok := unwrapProtoGeneric(schemaType, "map")
		if !ok {
			return nil, fmt.Errorf("invalid map type %q", schemaType)
		}
		parts := splitProtoGenericArgs(inner)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid map type %q", schemaType)
		}
		key, err := parseProtoTypeSpec(s, parts[0])
		if err != nil {
			return nil, err
		}
		value, err := parseProtoTypeSpec(s, parts[1])
		if err != nil {
			return nil, err
		}
		return &protoTypeSpec{Raw: schemaType, Kind: protoTypeMap, Key: key, Value: value}, nil
	}
	if s != nil {
		if _, ok := s.Enums[schemaType]; ok {
			return &protoTypeSpec{Raw: schemaType, Kind: protoTypeEnum}, nil
		}
		if _, ok := s.Messages[schemaType]; ok {
			return &protoTypeSpec{Raw: schemaType, Kind: protoTypeMessage}, nil
		}
	}
	switch schemaType {
	case "bool", "int32", "int64", "uint32", "uint64", "float32", "float64", "string", "bytes":
		return &protoTypeSpec{Raw: schemaType, Kind: protoTypeScalar}, nil
	default:
		return nil, fmt.Errorf("unknown type %q", schemaType)
	}
}

func renderProtoFieldType(spec *protoTypeSpec) (string, error) {
	if spec == nil {
		return "", fmt.Errorf("type is nil")
	}
	switch spec.Kind {
	case protoTypeList:
		if spec.Value.Kind == protoTypeList || spec.Value.Kind == protoTypeMap {
			return "", fmt.Errorf("proto exporter does not support nested repeated/map field type %q", spec.Raw)
		}
		value, err := renderProtoFieldType(spec.Value)
		if err != nil {
			return "", err
		}
		return "repeated " + value, nil
	case protoTypeMap:
		key, err := renderProtoMapKeyType(spec.Key)
		if err != nil {
			return "", err
		}
		if spec.Value.Kind == protoTypeList || spec.Value.Kind == protoTypeMap {
			return "", fmt.Errorf("proto exporter does not support repeated/map map value type %q", spec.Raw)
		}
		value, err := renderProtoFieldType(spec.Value)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("map<%s, %s>", key, value), nil
	case protoTypeEnum, protoTypeMessage:
		return spec.Raw, nil
	default:
		return renderProtoScalarType(spec.Raw)
	}
}

func renderProtoMapKeyType(spec *protoTypeSpec) (string, error) {
	if spec == nil {
		return "", fmt.Errorf("map key type is nil")
	}
	if spec.Kind != protoTypeScalar {
		return "", fmt.Errorf("proto map key type %q is not supported", spec.Raw)
	}
	switch spec.Raw {
	case "bool", "int32", "int64", "uint32", "uint64", "string":
		return renderProtoScalarType(spec.Raw)
	default:
		return "", fmt.Errorf("proto map key type %q is not supported", spec.Raw)
	}
}

func renderProtoScalarType(schemaType string) (string, error) {
	switch schemaType {
	case "bool", "uint32", "uint64", "string", "bytes":
		return schemaType, nil
	case "int32":
		return "sint32", nil
	case "int64":
		return "sint64", nil
	case "float32":
		return "float", nil
	case "float64":
		return "double", nil
	default:
		return "", fmt.Errorf("unknown scalar type %q", schemaType)
	}
}

func unwrapProtoGeneric(value string, name string) (string, bool) {
	prefix := name + "<"
	if !strings.HasPrefix(value, prefix) || !strings.HasSuffix(value, ">") {
		return "", false
	}
	return strings.TrimSpace(value[len(prefix) : len(value)-1]), true
}

func splitProtoGenericArgs(value string) []string {
	var parts []string
	depth := 0
	start := 0
	for i, r := range value {
		switch r {
		case '<':
			depth++
		case '>':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, strings.TrimSpace(value[start:i]))
				start = i + 1
			}
		}
	}
	parts = append(parts, strings.TrimSpace(value[start:]))
	return parts
}

func init() {
	RegisterExporter(protoExporter{})
}
