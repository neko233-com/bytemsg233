package schema

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

func ParseProto(data []byte) (*Schema, error) {
	parser := newProtoParser(data)
	s, err := parser.parse()
	if err != nil {
		return nil, err
	}
	if s.Version == "" {
		s.Version = "bymsg/v1"
	}
	if err := validateProtoSchemaTypes(s); err != nil {
		return nil, err
	}
	if err := validate(s); err != nil {
		return nil, err
	}
	return s, nil
}

type protoImporter struct{}

func (protoImporter) Name() string { return "proto" }
func (protoImporter) Extensions() []string {
	return []string{".proto"}
}
func (protoImporter) Import(data []byte, _ *ImportOptions) (*Schema, error) {
	return ParseProto(data)
}

type protoTokenKind int

const (
	protoTokenEOF protoTokenKind = iota
	protoTokenIdent
	protoTokenNumber
	protoTokenString
	protoTokenSymbol
	protoTokenComment
)

type protoToken struct {
	Kind  protoTokenKind
	Value string
	Line  int
}

type protoLexer struct {
	input []rune
	pos   int
	line  int
}

func newProtoLexer(data []byte) *protoLexer {
	return &protoLexer{input: []rune(string(data)), line: 1}
}

func (l *protoLexer) next() protoToken {
	l.skipWhitespace()
	if l.pos >= len(l.input) {
		return protoToken{Kind: protoTokenEOF, Line: l.line}
	}

	line := l.line
	ch := l.input[l.pos]
	if ch == '/' && l.peek(1) == '/' {
		l.pos += 2
		start := l.pos
		for l.pos < len(l.input) && l.input[l.pos] != '\n' {
			l.pos++
		}
		return protoToken{Kind: protoTokenComment, Value: string(l.input[start:l.pos]), Line: line}
	}
	if ch == '/' && l.peek(1) == '*' {
		l.pos += 2
		start := l.pos
		for l.pos < len(l.input) {
			if l.input[l.pos] == '*' && l.peek(1) == '/' {
				value := string(l.input[start:l.pos])
				l.pos += 2
				return protoToken{Kind: protoTokenComment, Value: value, Line: line}
			}
			if l.input[l.pos] == '\n' {
				l.line++
			}
			l.pos++
		}
		return protoToken{Kind: protoTokenEOF, Line: line}
	}
	if isProtoIdentStart(ch) {
		start := l.pos
		l.pos++
		for l.pos < len(l.input) && isProtoIdentPart(l.input[l.pos]) {
			l.pos++
		}
		return protoToken{Kind: protoTokenIdent, Value: string(l.input[start:l.pos]), Line: line}
	}
	if unicode.IsDigit(ch) || ch == '-' {
		start := l.pos
		l.pos++
		for l.pos < len(l.input) && unicode.IsDigit(l.input[l.pos]) {
			l.pos++
		}
		return protoToken{Kind: protoTokenNumber, Value: string(l.input[start:l.pos]), Line: line}
	}
	if ch == '"' || ch == '\'' {
		quote := ch
		l.pos++
		var buf strings.Builder
		for l.pos < len(l.input) {
			current := l.input[l.pos]
			if current == quote {
				l.pos++
				return protoToken{Kind: protoTokenString, Value: buf.String(), Line: line}
			}
			if current == '\\' && l.pos+1 < len(l.input) {
				l.pos++
				buf.WriteRune(l.input[l.pos])
				l.pos++
				continue
			}
			if current == '\n' {
				l.line++
			}
			buf.WriteRune(current)
			l.pos++
		}
		return protoToken{Kind: protoTokenString, Value: buf.String(), Line: line}
	}

	l.pos++
	return protoToken{Kind: protoTokenSymbol, Value: string(ch), Line: line}
}

func (l *protoLexer) skipWhitespace() {
	for l.pos < len(l.input) {
		switch l.input[l.pos] {
		case ' ', '\t', '\r':
			l.pos++
		case '\n':
			l.line++
			l.pos++
		default:
			return
		}
	}
}

func (l *protoLexer) peek(offset int) rune {
	if l.pos+offset >= len(l.input) {
		return 0
	}
	return l.input[l.pos+offset]
}

func isProtoIdentStart(ch rune) bool {
	return ch == '_' || unicode.IsLetter(ch)
}

func isProtoIdentPart(ch rune) bool {
	return ch == '_' || unicode.IsLetter(ch) || unicode.IsDigit(ch)
}

type protoParser struct {
	lexer           *protoLexer
	current         protoToken
	pendingComments []string
	sawSyntax       bool
	schema          *Schema
}

func newProtoParser(data []byte) *protoParser {
	parser := &protoParser{
		lexer: newProtoLexer(data),
		schema: &Schema{
			Messages: make(map[string]*Message),
			Enums:    make(map[string]*Enum),
		},
	}
	parser.next()
	return parser
}

func (p *protoParser) parse() (*Schema, error) {
	for p.current.Kind != protoTokenEOF {
		if p.current.Kind == protoTokenSymbol && p.current.Value == ";" {
			p.next()
			continue
		}
		if p.current.Kind != protoTokenIdent {
			return nil, p.errorf("expected top-level declaration, got %q", p.current.Value)
		}
		switch p.current.Value {
		case "syntax":
			if err := p.parseSyntax(); err != nil {
				return nil, err
			}
		case "package":
			if err := p.parsePackage(); err != nil {
				return nil, err
			}
		case "enum":
			if err := p.parseEnum(); err != nil {
				return nil, err
			}
		case "message":
			if err := p.parseMessage(); err != nil {
				return nil, err
			}
		default:
			if isUnsupportedProtoKeyword(p.current.Value) {
				return nil, p.unsupported(p.current.Value)
			}
			return nil, p.errorf("unexpected top-level declaration %q", p.current.Value)
		}
	}
	if !p.sawSyntax {
		return nil, fmt.Errorf("proto import requires syntax = \"proto3\"")
	}
	return p.schema, nil
}

func (p *protoParser) parseSyntax() error {
	p.pendingComments = nil
	p.next()
	if err := p.expectSymbol("="); err != nil {
		return err
	}
	p.next()
	value, err := p.expectString()
	if err != nil {
		return err
	}
	if value != "proto3" {
		return p.errorf("unsupported proto syntax %q; only proto3 is supported", value)
	}
	p.sawSyntax = true
	p.next()
	if err := p.expectSymbol(";"); err != nil {
		return err
	}
	p.next()
	return nil
}

func (p *protoParser) parsePackage() error {
	p.pendingComments = nil
	p.next()
	name, err := p.parseDottedName()
	if err != nil {
		return err
	}
	p.schema.Package = name
	if err := p.expectSymbol(";"); err != nil {
		return err
	}
	p.next()
	return nil
}

func (p *protoParser) parseEnum() error {
	p.pendingComments = nil
	p.next()
	name, err := p.expectIdentifier()
	if err != nil {
		return err
	}
	enum := &Enum{Values: make(map[string]int)}
	p.next()
	if err := p.expectSymbol("{"); err != nil {
		return err
	}
	p.next()
	for !(p.current.Kind == protoTokenSymbol && p.current.Value == "}") {
		if p.current.Kind == protoTokenEOF {
			return p.errorf("unterminated enum %s", name)
		}
		if p.current.Kind == protoTokenSymbol && p.current.Value == ";" {
			p.next()
			continue
		}
		if p.current.Kind == protoTokenIdent && isUnsupportedProtoKeyword(p.current.Value) {
			return p.unsupported(p.current.Value)
		}
		p.pendingComments = nil
		valueName, err := p.expectIdentifier()
		if err != nil {
			return err
		}
		p.next()
		if err := p.expectSymbol("="); err != nil {
			return err
		}
		p.next()
		value, err := p.expectNumber()
		if err != nil {
			return err
		}
		enum.Values[valueName] = value
		p.next()
		if p.current.Kind == protoTokenSymbol && p.current.Value == "[" {
			return p.errorf("enum value options are not supported")
		}
		if err := p.expectSymbol(";"); err != nil {
			return err
		}
		p.next()
	}
	p.schema.Enums[name] = enum
	p.pendingComments = nil
	p.next()
	return nil
}

func (p *protoParser) parseMessage() error {
	packetID := packetIDFromComments(p.pendingComments)
	p.pendingComments = nil
	p.next()
	name, err := p.expectIdentifier()
	if err != nil {
		return err
	}
	msg := &Message{Fields: make(map[string]*Field), PacketID: packetID}
	p.next()
	if err := p.expectSymbol("{"); err != nil {
		return err
	}
	p.next()
	for !(p.current.Kind == protoTokenSymbol && p.current.Value == "}") {
		if p.current.Kind == protoTokenEOF {
			return p.errorf("unterminated message %s", name)
		}
		if p.current.Kind == protoTokenSymbol && p.current.Value == ";" {
			p.next()
			continue
		}
		if p.current.Kind == protoTokenIdent {
			switch p.current.Value {
			case "message", "enum":
				return p.errorf("nested %s declarations are not supported", p.current.Value)
			case "oneof", "reserved", "extensions", "extend", "option", "import", "service", "rpc", "optional", "required":
				return p.unsupported(p.current.Value)
			}
		}
		p.pendingComments = nil
		fieldName, field, err := p.parseField()
		if err != nil {
			return err
		}
		msg.Fields[fieldName] = field
	}
	p.schema.Messages[name] = msg
	p.pendingComments = nil
	p.next()
	return nil
}

func (p *protoParser) parseField() (string, *Field, error) {
	repeated := false
	if p.current.Kind == protoTokenIdent && p.current.Value == "repeated" {
		repeated = true
		p.next()
	}

	fieldType, err := p.parseFieldType()
	if err != nil {
		return "", nil, err
	}
	if repeated {
		if strings.HasPrefix(fieldType, "map<") {
			return "", nil, p.errorf("repeated map fields are not supported")
		}
		fieldType = "list<" + fieldType + ">"
	}

	fieldName, err := p.expectIdentifier()
	if err != nil {
		return "", nil, err
	}
	p.next()
	if err := p.expectSymbol("="); err != nil {
		return "", nil, err
	}
	p.next()
	tag, err := p.expectNumber()
	if err != nil {
		return "", nil, err
	}
	if tag <= 0 {
		return "", nil, p.errorf("field %s tag must be positive", fieldName)
	}
	p.next()
	if p.current.Kind == protoTokenSymbol && p.current.Value == "[" {
		return "", nil, p.errorf("field options are not supported")
	}
	if err := p.expectSymbol(";"); err != nil {
		return "", nil, err
	}
	p.next()
	return fieldName, &Field{Type: fieldType, Tag: tag}, nil
}

func (p *protoParser) parseFieldType() (string, error) {
	if p.current.Kind != protoTokenIdent {
		return "", p.errorf("expected field type")
	}
	if p.current.Value == "map" {
		return p.parseMapType()
	}
	typeName, err := p.parseProtoTypeName()
	if err != nil {
		return "", err
	}
	return protoTypeToSchemaType(typeName), nil
}

func (p *protoParser) parseMapType() (string, error) {
	p.next()
	if err := p.expectSymbol("<"); err != nil {
		return "", err
	}
	p.next()
	keyType, err := p.parseProtoTypeName()
	if err != nil {
		return "", err
	}
	schemaKeyType := protoTypeToSchemaType(keyType)
	if !isSupportedProtoMapKey(schemaKeyType) {
		return "", p.errorf("proto map key type %q is not supported", keyType)
	}
	if err := p.expectSymbol(","); err != nil {
		return "", err
	}
	p.next()
	valueType, err := p.parseProtoTypeName()
	if err != nil {
		return "", err
	}
	schemaValueType := protoTypeToSchemaType(valueType)
	if err := p.expectSymbol(">"); err != nil {
		return "", err
	}
	p.next()
	return fmt.Sprintf("map<%s, %s>", schemaKeyType, schemaValueType), nil
}

func (p *protoParser) parseProtoTypeName() (string, error) {
	name, err := p.expectIdentifier()
	if err != nil {
		return "", err
	}
	p.next()
	if p.current.Kind == protoTokenSymbol && p.current.Value == "." {
		return "", p.errorf("qualified proto type names are not supported")
	}
	return name, nil
}

func (p *protoParser) parseDottedName() (string, error) {
	name, err := p.expectIdentifier()
	if err != nil {
		return "", err
	}
	var parts []string
	parts = append(parts, name)
	p.next()
	for p.current.Kind == protoTokenSymbol && p.current.Value == "." {
		p.next()
		part, err := p.expectIdentifier()
		if err != nil {
			return "", err
		}
		parts = append(parts, part)
		p.next()
	}
	return strings.Join(parts, "."), nil
}

func (p *protoParser) next() {
	for {
		token := p.lexer.next()
		if token.Kind != protoTokenComment {
			p.current = token
			return
		}
		p.consumeComment(token.Value)
	}
}

func (p *protoParser) consumeComment(value string) {
	text := strings.TrimSpace(value)
	if text == "" {
		return
	}
	if commentValue, ok := byteMsgProtoCommentValue(text, "schema"); ok {
		p.schema.Version = commentValue
	}
	if commentValue, ok := byteMsgProtoCommentValue(text, "protocolVersion"); ok {
		version, err := strconv.ParseUint(commentValue, 10, 64)
		if err == nil {
			p.schema.ProtocolVersion = version
		}
	}
	p.pendingComments = append(p.pendingComments, text)
}

func (p *protoParser) expectIdentifier() (string, error) {
	if p.current.Kind != protoTokenIdent {
		return "", p.errorf("expected identifier, got %q", p.current.Value)
	}
	return p.current.Value, nil
}

func (p *protoParser) expectNumber() (int, error) {
	if p.current.Kind != protoTokenNumber {
		return 0, p.errorf("expected number, got %q", p.current.Value)
	}
	value, err := strconv.Atoi(p.current.Value)
	if err != nil {
		return 0, p.errorf("invalid number %q", p.current.Value)
	}
	return value, nil
}

func (p *protoParser) expectString() (string, error) {
	if p.current.Kind != protoTokenString {
		return "", p.errorf("expected string, got %q", p.current.Value)
	}
	return p.current.Value, nil
}

func (p *protoParser) expectSymbol(symbol string) error {
	if p.current.Kind != protoTokenSymbol || p.current.Value != symbol {
		return p.errorf("expected %q, got %q", symbol, p.current.Value)
	}
	return nil
}

func (p *protoParser) errorf(format string, args ...any) error {
	return fmt.Errorf("proto line %d: %s", p.current.Line, fmt.Sprintf(format, args...))
}

func (p *protoParser) unsupported(keyword string) error {
	return p.errorf("proto %s is not supported", keyword)
}

func protoTypeToSchemaType(protoType string) string {
	switch protoType {
	case "float":
		return "float32"
	case "double":
		return "float64"
	case "sint32", "int32":
		return "int32"
	case "sint64", "int64":
		return "int64"
	default:
		return protoType
	}
}

func isSupportedProtoMapKey(schemaType string) bool {
	switch schemaType {
	case "bool", "int32", "int64", "uint32", "uint64", "string":
		return true
	default:
		return false
	}
}

func byteMsgProtoCommentValue(text string, key string) (string, bool) {
	prefix := "ByteMsg233 " + key + ":"
	if !strings.HasPrefix(strings.ToLower(text), strings.ToLower(prefix)) {
		return "", false
	}
	return strings.TrimSpace(text[len(prefix):]), true
}

func packetIDFromComments(comments []string) int {
	for i := len(comments) - 1; i >= 0; i-- {
		value, ok := byteMsgProtoCommentValue(comments[i], "packetId")
		if !ok {
			continue
		}
		packetID, err := strconv.Atoi(value)
		if err == nil {
			return packetID
		}
	}
	return 0
}

func isUnsupportedProtoKeyword(value string) bool {
	switch value {
	case "import", "option", "service", "rpc", "oneof", "reserved", "extensions", "extend":
		return true
	default:
		return false
	}
}

func validateProtoSchemaTypes(s *Schema) error {
	for msgName, msg := range s.Messages {
		for fieldName, field := range msg.Fields {
			if err := validateProtoSchemaType(s, field.Type); err != nil {
				return fmt.Errorf("proto %s.%s: %w", msgName, fieldName, err)
			}
		}
	}
	return nil
}

func validateProtoSchemaType(s *Schema, fieldType string) error {
	fieldType = strings.TrimSpace(fieldType)
	if strings.HasPrefix(fieldType, "list<") {
		inner, ok := unwrapProtoSchemaGeneric(fieldType, "list")
		if !ok {
			return fmt.Errorf("invalid list type %q", fieldType)
		}
		return validateProtoSchemaType(s, inner)
	}
	if strings.HasPrefix(fieldType, "map<") {
		inner, ok := unwrapProtoSchemaGeneric(fieldType, "map")
		if !ok {
			return fmt.Errorf("invalid map type %q", fieldType)
		}
		parts := splitProtoSchemaGenericArgs(inner)
		if len(parts) != 2 {
			return fmt.Errorf("invalid map type %q", fieldType)
		}
		if !isSupportedProtoMapKey(strings.TrimSpace(parts[0])) {
			return fmt.Errorf("proto map key type %q is not supported", strings.TrimSpace(parts[0]))
		}
		return validateProtoSchemaType(s, parts[1])
	}
	if _, ok := s.Enums[fieldType]; ok {
		return nil
	}
	if _, ok := s.Messages[fieldType]; ok {
		return nil
	}
	switch fieldType {
	case "bool", "int32", "int64", "uint32", "uint64", "float32", "float64", "string", "bytes":
		return nil
	default:
		return fmt.Errorf("unknown type %q", fieldType)
	}
}

func unwrapProtoSchemaGeneric(value string, name string) (string, bool) {
	prefix := name + "<"
	if !strings.HasPrefix(value, prefix) || !strings.HasSuffix(value, ">") {
		return "", false
	}
	return strings.TrimSpace(value[len(prefix) : len(value)-1]), true
}

func splitProtoSchemaGenericArgs(value string) []string {
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
	RegisterImporter(protoImporter{})
}
