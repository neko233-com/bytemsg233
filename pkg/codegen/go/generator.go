package gocodegen

import (
	"fmt"
	"sort"
	"strings"

	"github.com/neko233-com/bytemsg233/pkg/codegen"
	"github.com/neko233-com/bytemsg233/pkg/i18n"
	"github.com/neko233-com/bytemsg233/pkg/schema"
)

type Generator struct{}

type typeKind int

const (
	typeScalar typeKind = iota
	typeEnum
	typeMessage
	typeList
	typeMap
)

type typeSpec struct {
	Raw   string
	Go    string
	Kind  typeKind
	Key   *typeSpec
	Value *typeSpec
}

type generatorFeatures struct {
	Enums     bool
	Messages  bool
	Maps      bool
	Floats    bool
	PacketIDs bool
}

func New() *Generator { return &Generator{} }

func (g *Generator) Name() string          { return "go" }
func (g *Generator) FileExtension() string { return ".go" }

func (g *Generator) Generate(s *schema.Schema, options *codegen.GenerateOptions) ([]*codegen.GeneratedFile, error) {
	var buf strings.Builder

	prevLocale := i18n.GetLocale()
	if options != nil && options.Locale != "" {
		i18n.SetLocale(options.Locale)
		defer i18n.SetLocale(prevLocale)
	}

	packageName := s.Package
	if options != nil && options.Package != "" {
		packageName = options.Package
	}

	features, err := g.detectFeatures(s)
	if err != nil {
		return nil, err
	}

	buf.WriteString(fmt.Sprintf("package %s\n\n", packageName))
	g.generateImports(&buf, features)

	for _, name := range codegen.SortedEnumNames(s) {
		g.generateEnum(&buf, name, s.Enums[name])
		buf.WriteString("\n")
	}

	if features.Messages {
		g.generateWireHelpers(&buf, features)
	}

	for _, name := range codegen.SortedMessageNames(s) {
		if err := g.generateMessage(&buf, s, name, s.Messages[name]); err != nil {
			return nil, err
		}
		buf.WriteString("\n")
	}

	if features.PacketIDs {
		g.generatePacketRegistry(&buf, s)
	}

	return []*codegen.GeneratedFile{
		{Path: "types" + g.FileExtension(), Content: []byte(buf.String())},
	}, nil
}

func (g *Generator) generateImports(buf *strings.Builder, features generatorFeatures) {
	if !features.Messages {
		if features.Enums {
			buf.WriteString("import \"fmt\"\n\n")
		}
		return
	}

	imports := []string{
		"\"bytes\"",
		"\"encoding/binary\"",
		"\"fmt\"",
		"\"io\"",
		"\"strconv\"",
		"bytemsgBinary \"github.com/neko233-com/bytemsg233/pkg/binary\"",
	}
	if features.Floats {
		imports = append(imports, "\"math\"")
	}
	if features.PacketIDs {
		imports = append(imports, "\"reflect\"")
	}
	if features.Maps {
		imports = append(imports, "\"sort\"")
	}
	sort.Strings(imports)

	buf.WriteString("import (\n")
	for _, item := range imports {
		buf.WriteString("\t")
		buf.WriteString(item)
		buf.WriteString("\n")
	}
	buf.WriteString(")\n\n")
}

func (g *Generator) generateEnum(buf *strings.Builder, name string, enum *schema.Enum) {
	if enum.Description != nil {
		desc := i18n.GetDescription(enum.Description.Zh, enum.Description.En)
		if desc != "" {
			buf.WriteString(fmt.Sprintf("// %s\n", desc))
		}
	}

	buf.WriteString(fmt.Sprintf("type %s int32\n\n", name))
	buf.WriteString("const (\n")
	values := codegen.SortedEnumValues(enum)
	for _, value := range values {
		buf.WriteString(fmt.Sprintf("\t%s%s %s = %d\n", name, codegen.ToPascalCase(value.Name), name, value.Value))
	}
	buf.WriteString(")\n\n")

	buf.WriteString(fmt.Sprintf("func (x %s) String() string {\n", name))
	buf.WriteString("\tswitch x {\n")
	for _, value := range values {
		buf.WriteString(fmt.Sprintf("\tcase %s%s:\n", name, codegen.ToPascalCase(value.Name)))
		buf.WriteString(fmt.Sprintf("\t\treturn %q\n", value.Name))
	}
	buf.WriteString("\tdefault:\n")
	buf.WriteString(fmt.Sprintf("\t\treturn fmt.Sprintf(\"%s(%%d)\", int32(x))\n", name))
	buf.WriteString("\t}\n")
	buf.WriteString("}\n\n")

	buf.WriteString(fmt.Sprintf("func Parse%s(value int32) (%s, bool) {\n", name, name))
	buf.WriteString("\tswitch value {\n")
	for _, value := range values {
		buf.WriteString(fmt.Sprintf("\tcase %d:\n", value.Value))
		buf.WriteString(fmt.Sprintf("\t\treturn %s%s, true\n", name, codegen.ToPascalCase(value.Name)))
	}
	buf.WriteString("\tdefault:\n")
	buf.WriteString(fmt.Sprintf("\t\treturn %s(0), false\n", name))
	buf.WriteString("\t}\n")
	buf.WriteString("}\n\n")

	buf.WriteString(fmt.Sprintf("func (x %s) IsValid() bool {\n", name))
	buf.WriteString("\t_, ok := Parse" + name + "(int32(x))\n")
	buf.WriteString("\treturn ok\n")
	buf.WriteString("}\n")
}

func (g *Generator) generateWireHelpers(buf *strings.Builder, features generatorFeatures) {
	buf.WriteString("const (\n")
	buf.WriteString("\tByteMsgPacketPoolLimit = 10000\n")
	buf.WriteString("\tbyteMsgWireTypeVarint = 0\n")
	buf.WriteString("\tbyteMsgWireTypeFixed64 = 1\n")
	buf.WriteString("\tbyteMsgWireTypeLengthDelimited = 2\n")
	buf.WriteString("\tbyteMsgWireTypeFixed32 = 5\n")
	buf.WriteString(")\n\n")

	buf.WriteString("type byteMsgFastDecoder struct {\n")
	buf.WriteString("\treader *bytes.Reader\n")
	buf.WriteString("}\n\n")

	buf.WriteString("func (decoder byteMsgFastDecoder) ReadVarint() (uint64, error) {\n")
	buf.WriteString("\treturn binary.ReadUvarint(decoder.reader)\n")
	buf.WriteString("}\n\n")

	buf.WriteString("func (decoder byteMsgFastDecoder) ReadZigzag() (int64, error) {\n")
	buf.WriteString("\tvalue, err := decoder.ReadVarint()\n")
	buf.WriteString("\tif err != nil {\n\t\treturn 0, err\n\t}\n")
	buf.WriteString("\treturn bytemsgBinary.ZigzagDecode(value), nil\n")
	buf.WriteString("}\n\n")

	buf.WriteString("func (decoder byteMsgFastDecoder) ReadString() (string, error) {\n")
	buf.WriteString("\tlength, err := decoder.ReadVarint()\n")
	buf.WriteString("\tif err != nil {\n\t\treturn \"\", err\n\t}\n")
	buf.WriteString("\tbuf := make([]byte, length)\n")
	buf.WriteString("\tif _, err := io.ReadFull(decoder.reader, buf); err != nil {\n\t\treturn \"\", err\n\t}\n")
	buf.WriteString("\treturn string(buf), nil\n")
	buf.WriteString("}\n\n")

	buf.WriteString("func (decoder byteMsgFastDecoder) ReadBytes() ([]byte, error) {\n")
	buf.WriteString("\tlength, err := decoder.ReadVarint()\n")
	buf.WriteString("\tif err != nil {\n\t\treturn nil, err\n\t}\n")
	buf.WriteString("\tbuf := make([]byte, length)\n")
	buf.WriteString("\t_, err = io.ReadFull(decoder.reader, buf)\n")
	buf.WriteString("\treturn buf, err\n")
	buf.WriteString("}\n\n")

	buf.WriteString("func (decoder byteMsgFastDecoder) ReadFixed32() (uint32, error) {\n")
	buf.WriteString("\tvar buf [4]byte\n")
	buf.WriteString("\tif _, err := io.ReadFull(decoder.reader, buf[:]); err != nil {\n\t\treturn 0, err\n\t}\n")
	buf.WriteString("\treturn binary.LittleEndian.Uint32(buf[:]), nil\n")
	buf.WriteString("}\n\n")

	buf.WriteString("func (decoder byteMsgFastDecoder) ReadFixed64() (uint64, error) {\n")
	buf.WriteString("\tvar buf [8]byte\n")
	buf.WriteString("\tif _, err := io.ReadFull(decoder.reader, buf[:]); err != nil {\n\t\treturn 0, err\n\t}\n")
	buf.WriteString("\treturn binary.LittleEndian.Uint64(buf[:]), nil\n")
	buf.WriteString("}\n\n")

	buf.WriteString("func (decoder byteMsgFastDecoder) ReadFieldHeader() (tag int, wireType int, err error) {\n")
	buf.WriteString("\tvalue, err := decoder.ReadVarint()\n")
	buf.WriteString("\tif err != nil {\n\t\treturn 0, 0, err\n\t}\n")
	buf.WriteString("\treturn int(value >> 3), int(value & 0x7), nil\n")
	buf.WriteString("}\n\n")

	buf.WriteString("func byteMsgBoolToUint64(value bool) uint64 {\n")
	buf.WriteString("\tif value {\n\t\treturn 1\n\t}\n")
	buf.WriteString("\treturn 0\n")
	buf.WriteString("}\n\n")

	buf.WriteString("func byteMsgAppendTextInt(dst []byte, value int64) []byte {\n")
	buf.WriteString("\treturn strconv.AppendInt(dst, value, 10)\n")
	buf.WriteString("}\n\n")

	buf.WriteString("func byteMsgUnexpectedWireType(field string, got int, want int) error {\n")
	buf.WriteString("\treturn fmt.Errorf(\"bytemsg233: field %s wire type mismatch: got %d want %d\", field, got, want)\n")
	buf.WriteString("}\n\n")

	buf.WriteString("func byteMsgSkipUnknownField(decoder byteMsgFastDecoder, wireType int) error {\n")
	buf.WriteString("\tswitch wireType {\n")
	buf.WriteString("\tcase byteMsgWireTypeVarint:\n\t\t_, err := decoder.ReadVarint()\n\t\treturn err\n")
	buf.WriteString("\tcase byteMsgWireTypeFixed64:\n\t\t_, err := decoder.ReadFixed64()\n\t\treturn err\n")
	buf.WriteString("\tcase byteMsgWireTypeLengthDelimited:\n\t\t_, err := decoder.ReadBytes()\n\t\treturn err\n")
	buf.WriteString("\tcase byteMsgWireTypeFixed32:\n\t\t_, err := decoder.ReadFixed32()\n\t\treturn err\n")
	buf.WriteString("\tdefault:\n\t\treturn fmt.Errorf(\"bytemsg233: unsupported wire type %d\", wireType)\n")
	buf.WriteString("\t}\n")
	buf.WriteString("}\n\n")
}

func (g *Generator) generateMessage(buf *strings.Builder, s *schema.Schema, name string, msg *schema.Message) error {
	if msg.Description != nil {
		desc := i18n.GetDescription(msg.Description.Zh, msg.Description.En)
		if desc != "" {
			buf.WriteString(fmt.Sprintf("// %s\n", desc))
		}
	}

	buf.WriteString(fmt.Sprintf("type %s struct {\n", name))
	for _, fieldName := range codegen.SortedFieldNames(msg) {
		field := msg.Fields[fieldName]
		spec, err := g.typeSpec(s, field.Type)
		if err != nil {
			return fmt.Errorf("%s.%s: %w", name, fieldName, err)
		}
		if field.Description != nil {
			desc := i18n.GetDescription(field.Description.Zh, field.Description.En)
			if desc != "" {
				buf.WriteString(fmt.Sprintf("\t// %s\n", desc))
			}
		}

		goName := codegen.ToPascalCase(fieldName)
		buf.WriteString(fmt.Sprintf("\t%s %s `bytemsg:\"%d\"`\n", goName, spec.Go, field.Tag))
	}
	buf.WriteString("}\n\n")

	poolName := codegen.ToCamelCase(name) + "Pool"
	buf.WriteString(fmt.Sprintf("var %s = make(chan *%s, ByteMsgPacketPoolLimit)\n\n", poolName, name))

	buf.WriteString(fmt.Sprintf("// Acquire%s gets a reusable %s from the pool.\n", name, name))
	buf.WriteString(fmt.Sprintf("func Acquire%s() *%s {\n", name, name))
	buf.WriteString("\tselect {\n")
	buf.WriteString(fmt.Sprintf("\tcase value := <-%s:\n", poolName))
	buf.WriteString("\t\treturn value\n")
	buf.WriteString("\tdefault:\n")
	buf.WriteString(fmt.Sprintf("\t\treturn &%s{}\n", name))
	buf.WriteString("\t}\n")
	buf.WriteString("}\n\n")

	buf.WriteString(fmt.Sprintf("// Release%s resets a %s and returns it to the pool.\n", name, name))
	buf.WriteString(fmt.Sprintf("func Release%s(value *%s) {\n", name, name))
	buf.WriteString("\tif value == nil {\n\t\treturn\n\t}\n")
	buf.WriteString("\tvalue.Reset()\n")
	buf.WriteString("\tselect {\n")
	buf.WriteString(fmt.Sprintf("\tcase %s <- value:\n", poolName))
	buf.WriteString("\tdefault:\n")
	buf.WriteString("\t}\n")
	buf.WriteString("}\n\n")

	buf.WriteString(fmt.Sprintf("func (x *%s) ReleaseByteMsgPacket() {\n", name))
	buf.WriteString(fmt.Sprintf("\tRelease%s(x)\n", name))
	buf.WriteString("}\n\n")

	buf.WriteString(fmt.Sprintf("// Reset clears %s before it is reused.\n", name))
	buf.WriteString(fmt.Sprintf("func (x *%s) Reset() {\n", name))
	buf.WriteString("\tif x == nil {\n\t\treturn\n\t}\n")
	for _, fieldName := range codegen.SortedFieldNames(msg) {
		field := msg.Fields[fieldName]
		spec, err := g.typeSpec(s, field.Type)
		if err != nil {
			return err
		}
		goName := codegen.ToPascalCase(fieldName)
		buf.WriteString(fmt.Sprintf("\tx.%s = %s\n", goName, g.zeroValueExpr(s, spec)))
	}
	buf.WriteString("}\n\n")

	for _, fieldName := range codegen.SortedFieldNames(msg) {
		field := msg.Fields[fieldName]
		spec, err := g.typeSpec(s, field.Type)
		if err != nil {
			return err
		}
		goName := codegen.ToPascalCase(fieldName)
		buf.WriteString(fmt.Sprintf("func (x *%s) Get%s() %s {\n", name, goName, spec.Go))
		buf.WriteString("\tif x != nil {\n")
		buf.WriteString(fmt.Sprintf("\t\treturn x.%s\n", goName))
		buf.WriteString("\t}\n")
		buf.WriteString(fmt.Sprintf("\treturn %s\n", g.zeroValueExpr(s, spec)))
		buf.WriteString("}\n\n")
	}

	g.generateMarshalMethods(buf, s, name, msg)
	if err := g.generateUnmarshalMethods(buf, s, name, msg); err != nil {
		return err
	}
	g.generateTextMethods(buf, s, name, msg)
	return nil
}

func (g *Generator) generateMarshalMethods(buf *strings.Builder, s *schema.Schema, name string, msg *schema.Message) {
	buf.WriteString(fmt.Sprintf("func (x *%s) MarshalByteMsg() ([]byte, error) {\n", name))
	buf.WriteString("\tbuf := bytemsgBinary.GetBuffer()\n")
	buf.WriteString("\tdefer bytemsgBinary.PutBuffer(buf)\n")
	buf.WriteString("\tif err := x.MarshalByteMsgTo(buf); err != nil {\n\t\treturn nil, err\n\t}\n")
	buf.WriteString("\treturn append([]byte(nil), buf.Bytes()...), nil\n")
	buf.WriteString("}\n\n")

	buf.WriteString(fmt.Sprintf("func (x *%s) MarshalByteMsgTo(buffer *bytes.Buffer) error {\n", name))
	buf.WriteString("\tif x == nil {\n\t\treturn nil\n\t}\n")
	if len(msg.Fields) == 0 {
		buf.WriteString("\treturn nil\n")
		buf.WriteString("}\n\n")
		return
	}
	buf.WriteString("\tencoder := bytemsgBinary.NewEncoderValue(buffer)\n")
	for _, fieldName := range codegen.SortedFieldNames(msg) {
		field := msg.Fields[fieldName]
		spec, _ := g.typeSpec(s, field.Type)
		goName := codegen.ToPascalCase(fieldName)
		g.generateMarshalField(buf, s, goName, field.Tag, spec, "x."+goName, "\t")
	}
	buf.WriteString("\treturn nil\n")
	buf.WriteString("}\n\n")
}

func (g *Generator) generateMarshalField(buf *strings.Builder, s *schema.Schema, goName string, tag int, spec *typeSpec, valueExpr string, indent string) {
	switch spec.Kind {
	case typeList, typeMap:
		buf.WriteString(fmt.Sprintf("%sif len(%s) > 0 {\n", indent, valueExpr))
		buf.WriteString(fmt.Sprintf("%s\tif err := encoder.WriteFieldHeader(%d, byteMsgWireTypeLengthDelimited); err != nil {\n%s\t\treturn err\n%s\t}\n", indent, tag, indent, indent))
		g.generateWriteValue(buf, s, "encoder", valueExpr, spec, indent+"\t", goName)
		buf.WriteString(fmt.Sprintf("%s}\n", indent))
	case typeMessage:
		buf.WriteString(fmt.Sprintf("%s%sBuf := bytemsgBinary.GetBuffer()\n", indent, goName))
		buf.WriteString(fmt.Sprintf("%sif err := %s.MarshalByteMsgTo(%sBuf); err != nil {\n%s\tbytemsgBinary.PutBuffer(%sBuf)\n%s\treturn err\n%s}\n", indent, valueExpr, goName, indent, goName, indent, indent))
		buf.WriteString(fmt.Sprintf("%sif %sBuf.Len() > 0 {\n", indent, goName))
		buf.WriteString(fmt.Sprintf("%s\tif err := encoder.WriteFieldHeader(%d, byteMsgWireTypeLengthDelimited); err != nil {\n%s\t\tbytemsgBinary.PutBuffer(%sBuf)\n%s\t\treturn err\n%s\t}\n", indent, tag, indent, goName, indent, indent))
		buf.WriteString(fmt.Sprintf("%s\tif err := encoder.WriteBytes(%sBuf.Bytes()); err != nil {\n%s\t\tbytemsgBinary.PutBuffer(%sBuf)\n%s\t\treturn err\n%s\t}\n", indent, goName, indent, goName, indent, indent))
		buf.WriteString(fmt.Sprintf("%s}\n", indent))
		buf.WriteString(fmt.Sprintf("%sbytemsgBinary.PutBuffer(%sBuf)\n", indent, goName))
	default:
		zero := g.zeroValueExpr(s, spec)
		check := fmt.Sprintf("%s != %s", valueExpr, zero)
		if spec.Raw == "bytes" {
			check = fmt.Sprintf("len(%s) > 0", valueExpr)
		}
		buf.WriteString(fmt.Sprintf("%sif %s {\n", indent, check))
		buf.WriteString(fmt.Sprintf("%s\tif err := encoder.WriteFieldHeader(%d, %s); err != nil {\n%s\t\treturn err\n%s\t}\n", indent, tag, g.wireTypeConst(spec), indent, indent))
		g.generateWriteValue(buf, s, "encoder", valueExpr, spec, indent+"\t", goName)
		buf.WriteString(fmt.Sprintf("%s}\n", indent))
	}
}

func (g *Generator) generateWriteValue(buf *strings.Builder, s *schema.Schema, encoderName string, valueExpr string, spec *typeSpec, indent string, prefix string) {
	switch spec.Kind {
	case typeList:
		listBuf := prefix + "ListBuf"
		listEncoder := prefix + "ListEncoder"
		itemName := prefix + "Item"
		buf.WriteString(fmt.Sprintf("%s%s := bytemsgBinary.GetBuffer()\n", indent, listBuf))
		buf.WriteString(fmt.Sprintf("%s%s := bytemsgBinary.NewEncoderValue(%s)\n", indent, listEncoder, listBuf))
		buf.WriteString(fmt.Sprintf("%sif err := %s.WriteVarint(uint64(len(%s))); err != nil {\n%s\tbytemsgBinary.PutBuffer(%s)\n%s\treturn err\n%s}\n", indent, listEncoder, valueExpr, indent, listBuf, indent, indent))
		buf.WriteString(fmt.Sprintf("%sfor _, %s := range %s {\n", indent, itemName, valueExpr))
		g.generateWriteValue(buf, s, listEncoder, itemName, spec.Value, indent+"\t", prefix+"Item")
		buf.WriteString(fmt.Sprintf("%s}\n", indent))
		buf.WriteString(fmt.Sprintf("%sif err := %s.WriteBytes(%s.Bytes()); err != nil {\n%s\tbytemsgBinary.PutBuffer(%s)\n%s\treturn err\n%s}\n", indent, encoderName, listBuf, indent, listBuf, indent, indent))
		buf.WriteString(fmt.Sprintf("%sbytemsgBinary.PutBuffer(%s)\n", indent, listBuf))
	case typeMap:
		mapBuf := prefix + "MapBuf"
		mapEncoder := prefix + "MapEncoder"
		keysName := prefix + "Keys"
		keyName := prefix + "Key"
		buf.WriteString(fmt.Sprintf("%s%s := bytemsgBinary.GetBuffer()\n", indent, mapBuf))
		buf.WriteString(fmt.Sprintf("%s%s := bytemsgBinary.NewEncoderValue(%s)\n", indent, mapEncoder, mapBuf))
		buf.WriteString(fmt.Sprintf("%sif err := %s.WriteVarint(uint64(len(%s))); err != nil {\n%s\tbytemsgBinary.PutBuffer(%s)\n%s\treturn err\n%s}\n", indent, mapEncoder, valueExpr, indent, mapBuf, indent, indent))
		buf.WriteString(fmt.Sprintf("%s%s := make([]%s, 0, len(%s))\n", indent, keysName, spec.Key.Go, valueExpr))
		buf.WriteString(fmt.Sprintf("%sfor %s := range %s {\n%s\t%s = append(%s, %s)\n%s}\n", indent, keyName, valueExpr, indent, keysName, keysName, keyName, indent))
		if spec.Key.Go == "bool" {
			buf.WriteString(fmt.Sprintf("%ssort.Slice(%s, func(i, j int) bool { return !%s[i] && %s[j] })\n", indent, keysName, keysName, keysName))
		} else {
			buf.WriteString(fmt.Sprintf("%ssort.Slice(%s, func(i, j int) bool { return %s[i] < %s[j] })\n", indent, keysName, keysName, keysName))
		}
		buf.WriteString(fmt.Sprintf("%sfor _, %s := range %s {\n", indent, keyName, keysName))
		g.generateWriteValue(buf, s, mapEncoder, keyName, spec.Key, indent+"\t", prefix+"Key")
		g.generateWriteValue(buf, s, mapEncoder, valueExpr+"["+keyName+"]", spec.Value, indent+"\t", prefix+"Value")
		buf.WriteString(fmt.Sprintf("%s}\n", indent))
		buf.WriteString(fmt.Sprintf("%sif err := %s.WriteBytes(%s.Bytes()); err != nil {\n%s\tbytemsgBinary.PutBuffer(%s)\n%s\treturn err\n%s}\n", indent, encoderName, mapBuf, indent, mapBuf, indent, indent))
		buf.WriteString(fmt.Sprintf("%sbytemsgBinary.PutBuffer(%s)\n", indent, mapBuf))
	case typeMessage:
		nestedBuf := prefix + "MsgBuf"
		buf.WriteString(fmt.Sprintf("%s%s := bytemsgBinary.GetBuffer()\n", indent, nestedBuf))
		buf.WriteString(fmt.Sprintf("%sif err := %s.MarshalByteMsgTo(%s); err != nil {\n%s\tbytemsgBinary.PutBuffer(%s)\n%s\treturn err\n%s}\n", indent, valueExpr, nestedBuf, indent, nestedBuf, indent, indent))
		buf.WriteString(fmt.Sprintf("%sif err := %s.WriteBytes(%s.Bytes()); err != nil {\n%s\tbytemsgBinary.PutBuffer(%s)\n%s\treturn err\n%s}\n", indent, encoderName, nestedBuf, indent, nestedBuf, indent, indent))
		buf.WriteString(fmt.Sprintf("%sbytemsgBinary.PutBuffer(%s)\n", indent, nestedBuf))
	case typeEnum:
		buf.WriteString(fmt.Sprintf("%sif err := %s.WriteZigzag(int64(%s)); err != nil {\n%s\treturn err\n%s}\n", indent, encoderName, valueExpr, indent, indent))
	default:
		switch spec.Raw {
		case "bool":
			buf.WriteString(fmt.Sprintf("%sif err := %s.WriteVarint(byteMsgBoolToUint64(%s)); err != nil {\n%s\treturn err\n%s}\n", indent, encoderName, valueExpr, indent, indent))
		case "int32", "int64":
			buf.WriteString(fmt.Sprintf("%sif err := %s.WriteZigzag(int64(%s)); err != nil {\n%s\treturn err\n%s}\n", indent, encoderName, valueExpr, indent, indent))
		case "uint32", "uint64":
			buf.WriteString(fmt.Sprintf("%sif err := %s.WriteVarint(uint64(%s)); err != nil {\n%s\treturn err\n%s}\n", indent, encoderName, valueExpr, indent, indent))
		case "float32":
			buf.WriteString(fmt.Sprintf("%sif err := %s.WriteFixed32(math.Float32bits(%s)); err != nil {\n%s\treturn err\n%s}\n", indent, encoderName, valueExpr, indent, indent))
		case "float64":
			buf.WriteString(fmt.Sprintf("%sif err := %s.WriteFixed64(math.Float64bits(%s)); err != nil {\n%s\treturn err\n%s}\n", indent, encoderName, valueExpr, indent, indent))
		case "string":
			buf.WriteString(fmt.Sprintf("%sif err := %s.WriteString(%s); err != nil {\n%s\treturn err\n%s}\n", indent, encoderName, valueExpr, indent, indent))
		case "bytes":
			buf.WriteString(fmt.Sprintf("%sif err := %s.WriteBytes(%s); err != nil {\n%s\treturn err\n%s}\n", indent, encoderName, valueExpr, indent, indent))
		}
	}
}

func (g *Generator) generateUnmarshalMethods(buf *strings.Builder, s *schema.Schema, name string, msg *schema.Message) error {
	buf.WriteString(fmt.Sprintf("func (x *%s) UnmarshalByteMsg(data []byte) error {\n", name))
	buf.WriteString(fmt.Sprintf("\tif x == nil {\n\t\treturn fmt.Errorf(\"bytemsg233: nil target %s\")\n\t}\n", name))
	buf.WriteString("\tx.Reset()\n")
	if len(msg.Fields) == 0 {
		buf.WriteString("\treturn nil\n")
		buf.WriteString("}\n\n")
		return nil
	}
	buf.WriteString("\tvar reader bytes.Reader\n")
	buf.WriteString("\treader.Reset(data)\n")
	buf.WriteString("\treturn x.unmarshalByteMsgFrom(&reader)\n")
	buf.WriteString("}\n\n")

	buf.WriteString(fmt.Sprintf("func (x *%s) unmarshalByteMsgFrom(reader *bytes.Reader) error {\n", name))
	buf.WriteString("\tdecoder := byteMsgFastDecoder{reader: reader}\n")
	buf.WriteString("\tfor {\n")
	buf.WriteString("\t\ttag, wireType, err := decoder.ReadFieldHeader()\n")
	buf.WriteString("\t\tif err != nil {\n")
	buf.WriteString("\t\t\tif err == io.EOF {\n\t\t\t\treturn nil\n\t\t\t}\n")
	buf.WriteString("\t\t\treturn err\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t\tswitch tag {\n")
	for _, fieldName := range codegen.SortedFieldNames(msg) {
		field := msg.Fields[fieldName]
		spec, err := g.typeSpec(s, field.Type)
		if err != nil {
			return err
		}
		goName := codegen.ToPascalCase(fieldName)
		buf.WriteString(fmt.Sprintf("\t\tcase %d:\n", field.Tag))
		g.generateReadValue(buf, s, "decoder", "x."+goName, spec, "\t\t\t", name+goName)
	}
	buf.WriteString("\t\tdefault:\n")
	buf.WriteString("\t\t\tif err := byteMsgSkipUnknownField(decoder, wireType); err != nil {\n\t\t\t\treturn err\n\t\t\t}\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t}\n")
	buf.WriteString("}\n\n")
	return nil
}

func (g *Generator) generateReadValue(buf *strings.Builder, s *schema.Schema, decoderName string, targetExpr string, spec *typeSpec, indent string, prefix string) {
	switch spec.Kind {
	case typeList:
		bytesName := prefix + "Bytes"
		readerName := prefix + "Reader"
		decoderVar := prefix + "Decoder"
		lenName := prefix + "Len"
		itemName := prefix + "Item"
		buf.WriteString(fmt.Sprintf("%sif wireType != byteMsgWireTypeLengthDelimited {\n%s\treturn byteMsgUnexpectedWireType(%q, wireType, byteMsgWireTypeLengthDelimited)\n%s}\n", indent, indent, targetExpr, indent))
		buf.WriteString(fmt.Sprintf("%s%s, err := %s.ReadBytes()\n%sif err != nil {\n%s\treturn err\n%s}\n", indent, bytesName, decoderName, indent, indent, indent))
		buf.WriteString(fmt.Sprintf("%svar %s bytes.Reader\n%s%s.Reset(%s)\n", indent, readerName, indent, readerName, bytesName))
		buf.WriteString(fmt.Sprintf("%s%s := byteMsgFastDecoder{reader: &%s}\n", indent, decoderVar, readerName))
		buf.WriteString(fmt.Sprintf("%s%s, err := %s.ReadVarint()\n%sif err != nil {\n%s\treturn err\n%s}\n", indent, lenName, decoderVar, indent, indent, indent))
		buf.WriteString(fmt.Sprintf("%s%s = make(%s, 0, int(%s))\n", indent, targetExpr, spec.Go, lenName))
		buf.WriteString(fmt.Sprintf("%sfor i := uint64(0); i < %s; i++ {\n", indent, lenName))
		buf.WriteString(fmt.Sprintf("%s\tvar %s %s\n", indent, itemName, spec.Value.Go))
		g.generateReadNestedValue(buf, s, decoderVar, itemName, spec.Value, indent+"\t", prefix+"Item")
		buf.WriteString(fmt.Sprintf("%s\t%s = append(%s, %s)\n", indent, targetExpr, targetExpr, itemName))
		buf.WriteString(fmt.Sprintf("%s}\n", indent))
	case typeMap:
		bytesName := prefix + "Bytes"
		readerName := prefix + "Reader"
		decoderVar := prefix + "Decoder"
		lenName := prefix + "Len"
		keyName := prefix + "Key"
		valueName := prefix + "Value"
		buf.WriteString(fmt.Sprintf("%sif wireType != byteMsgWireTypeLengthDelimited {\n%s\treturn byteMsgUnexpectedWireType(%q, wireType, byteMsgWireTypeLengthDelimited)\n%s}\n", indent, indent, targetExpr, indent))
		buf.WriteString(fmt.Sprintf("%s%s, err := %s.ReadBytes()\n%sif err != nil {\n%s\treturn err\n%s}\n", indent, bytesName, decoderName, indent, indent, indent))
		buf.WriteString(fmt.Sprintf("%svar %s bytes.Reader\n%s%s.Reset(%s)\n", indent, readerName, indent, readerName, bytesName))
		buf.WriteString(fmt.Sprintf("%s%s := byteMsgFastDecoder{reader: &%s}\n", indent, decoderVar, readerName))
		buf.WriteString(fmt.Sprintf("%s%s, err := %s.ReadVarint()\n%sif err != nil {\n%s\treturn err\n%s}\n", indent, lenName, decoderVar, indent, indent, indent))
		buf.WriteString(fmt.Sprintf("%s%s = make(%s, int(%s))\n", indent, targetExpr, spec.Go, lenName))
		buf.WriteString(fmt.Sprintf("%sfor i := uint64(0); i < %s; i++ {\n", indent, lenName))
		buf.WriteString(fmt.Sprintf("%s\tvar %s %s\n", indent, keyName, spec.Key.Go))
		g.generateReadNestedValue(buf, s, decoderVar, keyName, spec.Key, indent+"\t", prefix+"Key")
		buf.WriteString(fmt.Sprintf("%s\tvar %s %s\n", indent, valueName, spec.Value.Go))
		g.generateReadNestedValue(buf, s, decoderVar, valueName, spec.Value, indent+"\t", prefix+"Value")
		buf.WriteString(fmt.Sprintf("%s\t%s[%s] = %s\n", indent, targetExpr, keyName, valueName))
		buf.WriteString(fmt.Sprintf("%s}\n", indent))
	case typeMessage:
		bytesName := prefix + "Bytes"
		readerName := prefix + "Reader"
		buf.WriteString(fmt.Sprintf("%sif wireType != byteMsgWireTypeLengthDelimited {\n%s\treturn byteMsgUnexpectedWireType(%q, wireType, byteMsgWireTypeLengthDelimited)\n%s}\n", indent, indent, targetExpr, indent))
		buf.WriteString(fmt.Sprintf("%s%s, err := %s.ReadBytes()\n%sif err != nil {\n%s\treturn err\n%s}\n", indent, bytesName, decoderName, indent, indent, indent))
		buf.WriteString(fmt.Sprintf("%svar %s bytes.Reader\n%s%s.Reset(%s)\n", indent, readerName, indent, readerName, bytesName))
		buf.WriteString(fmt.Sprintf("%sif err := %s.unmarshalByteMsgFrom(&%s); err != nil {\n%s\treturn err\n%s}\n", indent, targetExpr, readerName, indent, indent))
	default:
		buf.WriteString(fmt.Sprintf("%sif wireType != %s {\n%s\treturn byteMsgUnexpectedWireType(%q, wireType, %s)\n%s}\n", indent, g.wireTypeConst(spec), indent, targetExpr, g.wireTypeConst(spec), indent))
		g.generateReadNestedValue(buf, s, decoderName, targetExpr, spec, indent, prefix)
	}
}

func (g *Generator) generateReadNestedValue(buf *strings.Builder, s *schema.Schema, decoderName string, targetExpr string, spec *typeSpec, indent string, prefix string) {
	switch spec.Kind {
	case typeList, typeMap:
		buf.WriteString(fmt.Sprintf("%s%sWireType := byteMsgWireTypeLengthDelimited\n", indent, prefix))
		buf.WriteString(fmt.Sprintf("%swireType = %sWireType\n", indent, prefix))
		g.generateReadValue(buf, s, decoderName, targetExpr, spec, indent, prefix)
	case typeMessage:
		bytesName := prefix + "Bytes"
		readerName := prefix + "Reader"
		buf.WriteString(fmt.Sprintf("%s%s, err := %s.ReadBytes()\n%sif err != nil {\n%s\treturn err\n%s}\n", indent, bytesName, decoderName, indent, indent, indent))
		buf.WriteString(fmt.Sprintf("%svar %s bytes.Reader\n%s%s.Reset(%s)\n", indent, readerName, indent, readerName, bytesName))
		buf.WriteString(fmt.Sprintf("%sif err := %s.unmarshalByteMsgFrom(&%s); err != nil {\n%s\treturn err\n%s}\n", indent, targetExpr, readerName, indent, indent))
	case typeEnum:
		valueName := prefix + "Value"
		buf.WriteString(fmt.Sprintf("%s%s, err := %s.ReadZigzag()\n%sif err != nil {\n%s\treturn err\n%s}\n", indent, valueName, decoderName, indent, indent, indent))
		buf.WriteString(fmt.Sprintf("%s%s = %s(%s)\n", indent, targetExpr, spec.Go, valueName))
	default:
		valueName := prefix + "Value"
		switch spec.Raw {
		case "bool":
			buf.WriteString(fmt.Sprintf("%s%s, err := %s.ReadVarint()\n%sif err != nil {\n%s\treturn err\n%s}\n", indent, valueName, decoderName, indent, indent, indent))
			buf.WriteString(fmt.Sprintf("%s%s = %s != 0\n", indent, targetExpr, valueName))
		case "int32", "int64":
			buf.WriteString(fmt.Sprintf("%s%s, err := %s.ReadZigzag()\n%sif err != nil {\n%s\treturn err\n%s}\n", indent, valueName, decoderName, indent, indent, indent))
			buf.WriteString(fmt.Sprintf("%s%s = %s(%s)\n", indent, targetExpr, spec.Go, valueName))
		case "uint32", "uint64":
			buf.WriteString(fmt.Sprintf("%s%s, err := %s.ReadVarint()\n%sif err != nil {\n%s\treturn err\n%s}\n", indent, valueName, decoderName, indent, indent, indent))
			buf.WriteString(fmt.Sprintf("%s%s = %s(%s)\n", indent, targetExpr, spec.Go, valueName))
		case "float32":
			buf.WriteString(fmt.Sprintf("%s%s, err := %s.ReadFixed32()\n%sif err != nil {\n%s\treturn err\n%s}\n", indent, valueName, decoderName, indent, indent, indent))
			buf.WriteString(fmt.Sprintf("%s%s = math.Float32frombits(%s)\n", indent, targetExpr, valueName))
		case "float64":
			buf.WriteString(fmt.Sprintf("%s%s, err := %s.ReadFixed64()\n%sif err != nil {\n%s\treturn err\n%s}\n", indent, valueName, decoderName, indent, indent, indent))
			buf.WriteString(fmt.Sprintf("%s%s = math.Float64frombits(%s)\n", indent, targetExpr, valueName))
		case "string":
			buf.WriteString(fmt.Sprintf("%s%s, err := %s.ReadString()\n%sif err != nil {\n%s\treturn err\n%s}\n", indent, valueName, decoderName, indent, indent, indent))
			buf.WriteString(fmt.Sprintf("%s%s = %s\n", indent, targetExpr, valueName))
		case "bytes":
			buf.WriteString(fmt.Sprintf("%s%s, err := %s.ReadBytes()\n%sif err != nil {\n%s\treturn err\n%s}\n", indent, valueName, decoderName, indent, indent, indent))
			buf.WriteString(fmt.Sprintf("%s%s = %s\n", indent, targetExpr, valueName))
		}
	}
}

func (g *Generator) generateTextMethods(buf *strings.Builder, s *schema.Schema, name string, msg *schema.Message) {
	buf.WriteString(fmt.Sprintf("func (x *%s) AppendByteMsgText(dst []byte) []byte {\n", name))
	buf.WriteString("\tif x == nil {\n")
	buf.WriteString(fmt.Sprintf("\t\treturn append(dst, %q...)\n", name+"<nil>"))
	buf.WriteString("\t}\n")
	buf.WriteString(fmt.Sprintf("\tdst = append(dst, %q...)\n", name+"{"))
	for index, fieldName := range codegen.SortedFieldNames(msg) {
		field := msg.Fields[fieldName]
		spec, _ := g.typeSpec(s, field.Type)
		goName := codegen.ToPascalCase(fieldName)
		if index > 0 {
			buf.WriteString("\tdst = append(dst, ',')\n")
		}
		buf.WriteString(fmt.Sprintf("\tdst = append(dst, %q...)\n", goName+":"))
		g.generateAppendTextValue(buf, spec, "x."+goName, "\t", name+goName)
	}
	buf.WriteString("\tdst = append(dst, '}')\n")
	buf.WriteString("\treturn dst\n")
	buf.WriteString("}\n\n")

	buf.WriteString(fmt.Sprintf("func (x *%s) ByteMsgText() string {\n", name))
	buf.WriteString("\treturn string(x.AppendByteMsgText(nil))\n")
	buf.WriteString("}\n\n")
}

func (g *Generator) generateAppendTextValue(buf *strings.Builder, spec *typeSpec, valueExpr string, indent string, prefix string) {
	switch spec.Kind {
	case typeList:
		itemName := prefix + "TextItem"
		indexName := prefix + "TextIndex"
		buf.WriteString(fmt.Sprintf("%sdst = append(dst, '[')\n", indent))
		buf.WriteString(fmt.Sprintf("%sfor %s, %s := range %s {\n", indent, indexName, itemName, valueExpr))
		buf.WriteString(fmt.Sprintf("%s\tif %s > 0 {\n%s\t\tdst = append(dst, ',')\n%s\t}\n", indent, indexName, indent, indent))
		g.generateAppendTextValue(buf, spec.Value, itemName, indent+"\t", prefix+"Item")
		buf.WriteString(fmt.Sprintf("%s}\n", indent))
		buf.WriteString(fmt.Sprintf("%sdst = append(dst, ']')\n", indent))
	case typeMap:
		keyName := prefix + "TextKey"
		valueName := prefix + "TextValue"
		firstName := prefix + "TextFirst"
		buf.WriteString(fmt.Sprintf("%sdst = append(dst, '{')\n", indent))
		buf.WriteString(fmt.Sprintf("%s%s := true\n", indent, firstName))
		buf.WriteString(fmt.Sprintf("%sfor %s, %s := range %s {\n", indent, keyName, valueName, valueExpr))
		buf.WriteString(fmt.Sprintf("%s\tif !%s {\n%s\t\tdst = append(dst, ',')\n%s\t}\n", indent, firstName, indent, indent))
		buf.WriteString(fmt.Sprintf("%s\t%s = false\n", indent, firstName))
		g.generateAppendTextValue(buf, spec.Key, keyName, indent+"\t", prefix+"Key")
		buf.WriteString(fmt.Sprintf("%s\tdst = append(dst, ':')\n", indent))
		g.generateAppendTextValue(buf, spec.Value, valueName, indent+"\t", prefix+"Value")
		buf.WriteString(fmt.Sprintf("%s}\n", indent))
		buf.WriteString(fmt.Sprintf("%sdst = append(dst, '}')\n", indent))
	case typeMessage:
		buf.WriteString(fmt.Sprintf("%sdst = %s.AppendByteMsgText(dst)\n", indent, valueExpr))
	case typeEnum:
		buf.WriteString(fmt.Sprintf("%sdst = strconv.AppendInt(dst, int64(%s), 10)\n", indent, valueExpr))
	default:
		switch spec.Raw {
		case "bool":
			buf.WriteString(fmt.Sprintf("%sdst = strconv.AppendBool(dst, %s)\n", indent, valueExpr))
		case "int32", "int64":
			buf.WriteString(fmt.Sprintf("%sdst = strconv.AppendInt(dst, int64(%s), 10)\n", indent, valueExpr))
		case "uint32", "uint64":
			buf.WriteString(fmt.Sprintf("%sdst = strconv.AppendUint(dst, uint64(%s), 10)\n", indent, valueExpr))
		case "float32":
			buf.WriteString(fmt.Sprintf("%sdst = strconv.AppendFloat(dst, float64(%s), 'g', -1, 32)\n", indent, valueExpr))
		case "float64":
			buf.WriteString(fmt.Sprintf("%sdst = strconv.AppendFloat(dst, %s, 'g', -1, 64)\n", indent, valueExpr))
		case "string":
			buf.WriteString(fmt.Sprintf("%sdst = strconv.AppendQuote(dst, %s)\n", indent, valueExpr))
		case "bytes":
			buf.WriteString(fmt.Sprintf("%sdst = append(dst, \"bytes(len=\"...)\n", indent))
			buf.WriteString(fmt.Sprintf("%sdst = strconv.AppendInt(dst, int64(len(%s)), 10)\n", indent, valueExpr))
			buf.WriteString(fmt.Sprintf("%sdst = append(dst, ')')\n", indent))
		}
	}
}

func (g *Generator) generatePacketRegistry(buf *strings.Builder, s *schema.Schema) {
	packetNames := make([]string, 0)
	for _, name := range codegen.SortedMessageNames(s) {
		if s.Messages[name].PacketID > 0 {
			packetNames = append(packetNames, name)
		}
	}

	buf.WriteString("type PacketRegistration struct {\n")
	buf.WriteString("\tPacketId int\n")
	buf.WriteString("\tName string\n")
	buf.WriteString("\tType reflect.Type\n")
	buf.WriteString("}\n\n")

	buf.WriteString("var byteMsgPacketAcquirers = map[int]func() any{\n")
	for _, name := range packetNames {
		buf.WriteString(fmt.Sprintf("\t%d: func() any { return Acquire%s() },\n", s.Messages[name].PacketID, name))
	}
	buf.WriteString("}\n\n")

	buf.WriteString("var byteMsgPacketReleasers = map[int]func(any){\n")
	for _, name := range packetNames {
		buf.WriteString(fmt.Sprintf("\t%d: func(value any) { if packet, ok := value.(*%s); ok { Release%s(packet) } },\n", s.Messages[name].PacketID, name, name))
	}
	buf.WriteString("}\n\n")

	buf.WriteString("var byteMsgPacketTypes = map[int]reflect.Type{\n")
	for _, name := range packetNames {
		buf.WriteString(fmt.Sprintf("\t%d: reflect.TypeOf((*%s)(nil)),\n", s.Messages[name].PacketID, name))
	}
	buf.WriteString("}\n\n")

	buf.WriteString("var byteMsgPacketNames = map[int]string{\n")
	for _, name := range packetNames {
		buf.WriteString(fmt.Sprintf("\t%d: %q,\n", s.Messages[name].PacketID, name))
	}
	buf.WriteString("}\n\n")

	buf.WriteString("var byteMsgPacketIDsByType = map[reflect.Type]int{\n")
	for _, name := range packetNames {
		packetID := s.Messages[name].PacketID
		buf.WriteString(fmt.Sprintf("\treflect.TypeOf((*%s)(nil)): %d,\n", name, packetID))
		buf.WriteString(fmt.Sprintf("\treflect.TypeOf(%s{}): %d,\n", name, packetID))
	}
	buf.WriteString("}\n\n")

	buf.WriteString("func AcquireByteMsgPacketById(packetId int) (any, bool) {\n")
	buf.WriteString("\tacquire, ok := byteMsgPacketAcquirers[packetId]\n")
	buf.WriteString("\tif !ok {\n\t\treturn nil, false\n\t}\n")
	buf.WriteString("\treturn acquire(), true\n")
	buf.WriteString("}\n\n")

	buf.WriteString("func ReleaseByteMsgPacket(packetId int, value any) bool {\n")
	buf.WriteString("\trelease, ok := byteMsgPacketReleasers[packetId]\n")
	buf.WriteString("\tif !ok {\n\t\treturn false\n\t}\n")
	buf.WriteString("\trelease(value)\n")
	buf.WriteString("\treturn true\n")
	buf.WriteString("}\n\n")

	buf.WriteString("func GetPacketIdByType(msgType reflect.Type) (int, bool) {\n")
	buf.WriteString("\tpacketId, ok := byteMsgPacketIDsByType[msgType]\n")
	buf.WriteString("\treturn packetId, ok\n")
	buf.WriteString("}\n\n")

	buf.WriteString("func GetTypeByPacketId(packetId int) (reflect.Type, bool) {\n")
	buf.WriteString("\tmsgType, ok := byteMsgPacketTypes[packetId]\n")
	buf.WriteString("\treturn msgType, ok\n")
	buf.WriteString("}\n\n")

	buf.WriteString("func GetAllPacketMappings() map[int]string {\n")
	buf.WriteString("\tresult := make(map[int]string, len(byteMsgPacketNames))\n")
	buf.WriteString("\tfor packetId, name := range byteMsgPacketNames {\n\t\tresult[packetId] = name\n\t}\n")
	buf.WriteString("\treturn result\n")
	buf.WriteString("}\n\n")

	buf.WriteString("func GetAllPacketRegistrations() []PacketRegistration {\n")
	buf.WriteString("\titems := make([]PacketRegistration, 0, len(byteMsgPacketTypes))\n")
	buf.WriteString("\tfor packetId, msgType := range byteMsgPacketTypes {\n")
	buf.WriteString("\t\titems = append(items, PacketRegistration{PacketId: packetId, Name: byteMsgPacketNames[packetId], Type: msgType})\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tsort.Slice(items, func(i, j int) bool { return items[i].PacketId < items[j].PacketId })\n")
	buf.WriteString("\treturn items\n")
	buf.WriteString("}\n\n")

	buf.WriteString("func RegisterAllPackets(register func(packetId int, msgType reflect.Type)) {\n")
	buf.WriteString("\tfor _, item := range GetAllPacketRegistrations() {\n")
	buf.WriteString("\t\tregister(item.PacketId, item.Type)\n")
	buf.WriteString("\t}\n")
	buf.WriteString("}\n\n")

	buf.WriteString("func VisitAllPackets(visit func(packetId int, name string, msgType reflect.Type)) {\n")
	buf.WriteString("\tfor _, item := range GetAllPacketRegistrations() {\n")
	buf.WriteString("\t\tvisit(item.PacketId, item.Name, item.Type)\n")
	buf.WriteString("\t}\n")
	buf.WriteString("}\n")
}

func (g *Generator) wireTypeConst(spec *typeSpec) string {
	switch spec.Kind {
	case typeList, typeMap, typeMessage:
		return "byteMsgWireTypeLengthDelimited"
	case typeEnum:
		return "byteMsgWireTypeVarint"
	}
	switch spec.Raw {
	case "float32":
		return "byteMsgWireTypeFixed32"
	case "float64":
		return "byteMsgWireTypeFixed64"
	case "string", "bytes":
		return "byteMsgWireTypeLengthDelimited"
	default:
		return "byteMsgWireTypeVarint"
	}
}

func (g *Generator) detectFeatures(s *schema.Schema) (generatorFeatures, error) {
	features := generatorFeatures{Enums: len(s.Enums) > 0, Messages: len(s.Messages) > 0}
	for _, msg := range s.Messages {
		if msg.PacketID > 0 {
			features.PacketIDs = true
		}
		for fieldName, field := range msg.Fields {
			spec, err := g.typeSpec(s, field.Type)
			if err != nil {
				return features, fmt.Errorf("%s: %w", fieldName, err)
			}
			if g.typeHasMap(spec) {
				features.Maps = true
			}
			if g.typeHasFloat(spec) {
				features.Floats = true
			}
		}
	}
	if features.PacketIDs {
		features.Maps = true
	}
	return features, nil
}

func (g *Generator) typeHasMap(spec *typeSpec) bool {
	if spec == nil {
		return false
	}
	if spec.Kind == typeMap {
		return true
	}
	return g.typeHasMap(spec.Key) || g.typeHasMap(spec.Value)
}

func (g *Generator) typeHasFloat(spec *typeSpec) bool {
	if spec == nil {
		return false
	}
	if spec.Raw == "float32" || spec.Raw == "float64" {
		return true
	}
	return g.typeHasFloat(spec.Key) || g.typeHasFloat(spec.Value)
}

func (g *Generator) typeSpec(s *schema.Schema, schemaType string) (*typeSpec, error) {
	schemaType = strings.TrimSpace(schemaType)
	if strings.HasPrefix(schemaType, "list<") {
		inner, ok := unwrapGeneric(schemaType, "list")
		if !ok {
			return nil, fmt.Errorf("invalid list type %q", schemaType)
		}
		value, err := g.typeSpec(s, inner)
		if err != nil {
			return nil, err
		}
		return &typeSpec{Raw: schemaType, Go: "[]" + value.Go, Kind: typeList, Value: value}, nil
	}
	if strings.HasPrefix(schemaType, "map<") {
		inner, ok := unwrapGeneric(schemaType, "map")
		if !ok {
			return nil, fmt.Errorf("invalid map type %q", schemaType)
		}
		parts := splitGenericArgs(inner)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid map type %q", schemaType)
		}
		key, err := g.typeSpec(s, parts[0])
		if err != nil {
			return nil, err
		}
		if !g.isValidMapKey(key) {
			return nil, fmt.Errorf("invalid map key type %q", parts[0])
		}
		value, err := g.typeSpec(s, parts[1])
		if err != nil {
			return nil, err
		}
		return &typeSpec{Raw: schemaType, Go: fmt.Sprintf("map[%s]%s", key.Go, value.Go), Kind: typeMap, Key: key, Value: value}, nil
	}
	if _, ok := s.Enums[schemaType]; ok {
		return &typeSpec{Raw: schemaType, Go: schemaType, Kind: typeEnum}, nil
	}
	if _, ok := s.Messages[schemaType]; ok {
		return &typeSpec{Raw: schemaType, Go: schemaType, Kind: typeMessage}, nil
	}
	switch schemaType {
	case "bool", "int32", "int64", "uint32", "uint64", "float32", "float64", "string":
		return &typeSpec{Raw: schemaType, Go: schemaType, Kind: typeScalar}, nil
	case "bytes":
		return &typeSpec{Raw: schemaType, Go: "[]byte", Kind: typeScalar}, nil
	default:
		return nil, fmt.Errorf("unknown type %q", schemaType)
	}
}

func (g *Generator) isValidMapKey(spec *typeSpec) bool {
	if spec == nil {
		return false
	}
	if spec.Kind == typeEnum {
		return true
	}
	if spec.Kind != typeScalar {
		return false
	}
	switch spec.Raw {
	case "bool", "int32", "int64", "uint32", "uint64", "float32", "float64", "string":
		return true
	default:
		return false
	}
}

func (g *Generator) zeroValueExpr(s *schema.Schema, spec *typeSpec) string {
	switch spec.Kind {
	case typeList, typeMap:
		return "nil"
	case typeEnum:
		if enum, ok := s.Enums[spec.Raw]; ok {
			if value, exists := codegen.DefaultEnumValue(enum); exists {
				return fmt.Sprintf("%s%s", spec.Raw, codegen.ToPascalCase(value.Name))
			}
		}
		return fmt.Sprintf("%s(0)", spec.Go)
	case typeMessage:
		return fmt.Sprintf("*new(%s)", spec.Go)
	}
	switch spec.Raw {
	case "bool":
		return "false"
	case "int32", "int64", "uint32", "uint64", "float32", "float64":
		return "0"
	case "string":
		return "\"\""
	case "bytes":
		return "nil"
	default:
		return "nil"
	}
}

func unwrapGeneric(value string, name string) (string, bool) {
	prefix := name + "<"
	if !strings.HasPrefix(value, prefix) || !strings.HasSuffix(value, ">") {
		return "", false
	}
	return strings.TrimSpace(value[len(prefix) : len(value)-1]), true
}

func splitGenericArgs(value string) []string {
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
	codegen.Register(New())
}
