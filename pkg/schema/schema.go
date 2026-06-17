package schema

// Schema represents a complete ByteMsg233 schema file.
type Schema struct {
	Version         string              `yaml:"schema" json:"schema"`
	ProtocolVersion uint64              `yaml:"protocolVersion,omitempty" json:"protocolVersion,omitempty"`
	Package         string              `yaml:"package" json:"package"`
	Messages        map[string]*Message `yaml:"messages" json:"messages"`
	Enums           map[string]*Enum    `yaml:"enums" json:"enums"`
}

// Message represents a message type definition
type Message struct {
	Fields      map[string]*Field `yaml:"fields" json:"fields"`
	Description *Description      `yaml:"description,omitempty" json:"description,omitempty"`
	PacketID    int               `yaml:"packetId,omitempty" json:"packetId,omitempty"`
}

// Field represents a field in a message
type Field struct {
	Type        string       `yaml:"type" json:"type"`
	Description *Description `yaml:"description,omitempty" json:"description,omitempty"`
	Comment     string       `yaml:"comment,omitempty" json:"comment,omitempty"`
	Tag         int          `yaml:"tag" json:"tag"`
}

// Enum represents an enumeration type
type Enum struct {
	Values      map[string]int `yaml:"values" json:"values"`
	Description *Description   `yaml:"description,omitempty" json:"description,omitempty"`
}

// Description holds i18n descriptions
type Description struct {
	Zh string `yaml:"zh,omitempty" json:"zh,omitempty"`
	En string `yaml:"en,omitempty" json:"en,omitempty"`
}
