// Package models defines the AST nodes for Protocol Buffer schema files.
package models

import (
	"fmt"
	"strings"
)

// ProtoFile represents a parsed .proto file.
type ProtoFile struct {
	Syntax    string         // e.g. "proto3", "proto2"
	Package   string         // Package declaration
	Options   []*Option      // File-level options
	Imports   []*Import      // Import statements
	Messages  []*Message     // Message definitions
	Enums     []*Enum        // Enum definitions
	Services  []*Service     // Service definitions
	Extensions []*Extension  // Extension ranges (proto2)
	Comments  map[string]string // Trailing comments keyed by path
}

// Import represents an import statement.
type Import struct {
	Weak    bool   // weak import
	Public  bool   // public import
	Path    string // File path
	Comment string
}

// Option represents a key-value option (e.g., option java_package = "com.example";).
type Option struct {
	Name     string
	Value    string
	IsCustom bool // true if it's a custom option (e.g., my.option)
	Comment  string
}

// Message represents a message definition.
type Message struct {
	Name       string
	Fields     []*Field
	OneOfs     []*OneOf
	Messages   []*Message   // Nested messages
	Enums      []*Enum      // Nested enums
	Maps       []*MapField  // Map fields
	Options    []*Option
	Extensions []*Extension
	Comment    string
}

// Field represents a message field.
type Field struct {
	Number     int
	Name       string
	Type       string // The proto type name
	Label      string // optional, required, repeated, or "" (proto3 default)
	DefaultValue string
	Options    []*Option
	IsMap      bool
	Comment    string
}

// OneOf represents a oneof declaration.
type OneOf struct {
	Name   string
	Fields []*Field
	Options []*Option
	Comment string
}

// MapField represents a map<K,V> field.
type MapField struct {
	KeyType   string
	ValueType string
	Name      string
	Number    int
	Options   []*Option
	Comment   string
}

// Enum represents an enum definition.
type Enum struct {
	Name    string
	Values  []*EnumValue
	Options []*Option
	Comment string
}

// EnumValue represents a value within an enum.
type EnumValue struct {
	Name    string
	Number  int
	Options []*Option
	Comment string
}

// Service represents a service definition.
type Service struct {
	Name    string
	Methods []*Method
	Options []*Option
	Comment string
}

// Method represents an RPC method within a service.
type Method struct {
	Name            string
	InputType       string
	OutputType      string
	ClientStreaming  bool
	ServerStreaming  bool
	Options         []*Option
	Comment         string
}

// Extension represents an extension range.
type Extension struct {
	RangeStart int
	RangeEnd   int
	Comment    string
}

// ProtoPackage represents a package analysis result.
type ProtoPackage struct {
	Name      string
	Messages  int
	Enums     int
	Services  int
	Methods   int
	Fields    int
	Imports   int
}

// FieldLabel represents the label of a field.
type FieldLabel int

const (
	LabelNone FieldLabel = iota
	LabelOptional
	LabelRequired
	LabelRepeated
)

func (f *Field) LabelEnum() FieldLabel {
	switch strings.ToLower(f.Label) {
	case "optional":
		return LabelOptional
	case "required":
		return LabelRequired
	case "repeated":
		return LabelRepeated
	default:
		return LabelNone
	}
}

// FullyQualifiedPath returns the full path for a message/enum.
func (m *Message) FullyQualifiedPath(packageName string) string {
	return fmt.Sprintf("%s.%s", packageName, m.Name)
}

// FullyQualifiedPath returns the full path for an enum.
func (e *Enum) FullyQualifiedPath(packageName string) string {
	return fmt.Sprintf("%s.%s", packageName, e.Name)
}

// ScalarTypes is the set of all scalar protobuf types.
var ScalarTypes = map[string]bool{
	"double": true, "float": true,
	"int32": true, "int64": true, "uint32": true, "uint64": true,
	"sint32": true, "sint64": true,
	"fixed32": true, "fixed64": true,
	"sfixed32": true, "sfixed64": true,
	"bool": true,
	"string": true,
	"bytes": true,
}

// IsScalarType checks if a type name is a built-in scalar type.
func IsScalarType(t string) bool {
	return ScalarTypes[t]
}

// IsMessageType checks if a type name looks like a message reference (starts with uppercase or contains a dot).
func IsMessageType(t string) bool {
	if len(t) == 0 {
		return false
	}
	// Map types
	if strings.HasPrefix(t, "map<") {
		return false
	}
	// Scalar types
	if IsScalarType(t) {
		return false
	}
	// Well-known types
	knownTypes := map[string]bool{
		"google.protobuf.Timestamp": true,
		"google.protobuf.Duration":  true,
		"google.protobuf.Empty":     true,
		"google.protobuf.Any":       true,
		"google.protobuf.Struct":    true,
		"google.protobuf.Value":     true,
		"google.protobuf.ListValue": true,
		"google.protobuf.FieldMask": true,
	}
	if knownTypes[t] {
		return true
	}
	// Simple name starting with uppercase
	return t[0] >= 'A' && t[0] <= 'Z'
}

// ParseMapType extracts key and value types from a map<key,value> type string.
func ParseMapType(t string) (key, val string, ok bool) {
	if !strings.HasPrefix(t, "map<") || !strings.HasSuffix(t, ">") {
		return "", "", false
	}
	inner := t[4 : len(t)-1]
	// Find the comma, handling nested angle brackets
	depth := 0
	for i, c := range inner {
		switch c {
		case '<':
			depth++
		case '>':
			depth--
		case ',':
			if depth == 0 {
				return strings.TrimSpace(inner[:i]), strings.TrimSpace(inner[i+1:]), true
			}
		}
	}
	return "", "", false
}
