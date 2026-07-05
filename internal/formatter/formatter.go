// Package formatter provides formatting and pretty-printing for Protocol Buffer schemas.
package formatter

import (
	"fmt"
	"sort"
	"strings"

	"github.com/EdgarOrtegaRamirez/protoforge/internal/models"
)

// Options controls formatting behavior.
type Options struct {
	Indent     string // Indentation string (default: "  ")
	LineWidth  int    // Maximum line width (default: 120)
	SortFields bool   // Sort fields by number
}

// DefaultOptions returns default formatting options.
func DefaultOptions() *Options {
	return &Options{
		Indent:     "  ",
		LineWidth:  120,
		SortFields: false,
	}
}

// Format formats a parsed proto file back to .proto syntax.
func Format(pf *models.ProtoFile, opts *Options) string {
	if opts == nil {
		opts = DefaultOptions()
	}

	var sb strings.Builder

	// Syntax declaration
	if pf.Syntax != "" {
		sb.WriteString(fmt.Sprintf("syntax = \"%s\";\n\n", pf.Syntax))
	}

	// Package
	if pf.Package != "" {
		sb.WriteString(fmt.Sprintf("package %s;\n\n", pf.Package))
	}

	// Options
	for _, opt := range pf.Options {
		sb.WriteString(fmt.Sprintf("option %s = %s;\n", opt.Name, formatValue(opt.Value)))
	}
	if len(pf.Options) > 0 {
		sb.WriteString("\n")
	}

	// Imports
	for _, imp := range pf.Imports {
		prefix := ""
		if imp.Weak {
			prefix = "weak "
		} else if imp.Public {
			prefix = "public "
		}
		sb.WriteString(fmt.Sprintf("import %s\"%s\";\n", prefix, imp.Path))
	}
	if len(pf.Imports) > 0 {
		sb.WriteString("\n")
	}

	// Messages
	for _, msg := range pf.Messages {
		sb.WriteString(formatMessage(msg, "", opts))
		sb.WriteString("\n")
	}

	// Enums
	for _, enum := range pf.Enums {
		sb.WriteString(formatEnum(enum, "", opts))
		sb.WriteString("\n")
	}

	// Services
	for _, svc := range pf.Services {
		sb.WriteString(formatService(svc, opts))
		sb.WriteString("\n")
	}

	return sb.String()
}

func formatMessage(msg *models.Message, indent string, opts *Options) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%smessage %s {\n", indent, msg.Name))
	inner := indent + opts.Indent

	// Nested enums first
	for _, enum := range msg.Enums {
		sb.WriteString(formatEnum(enum, inner, opts))
		sb.WriteString("\n")
	}

	// Nested messages
	for _, nested := range msg.Messages {
		sb.WriteString(formatMessage(nested, inner, opts))
		sb.WriteString("\n")
	}

	// Fields
	fields := msg.Fields
	if opts.SortFields {
		sorted := make([]*models.Field, len(fields))
		copy(sorted, fields)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Number < sorted[j].Number
		})
		fields = sorted
	}

	for _, field := range fields {
		if field.Label != "" {
			sb.WriteString(fmt.Sprintf("%s%s %s %s = %d", inner, field.Label, field.Type, field.Name, field.Number))
		} else {
			sb.WriteString(fmt.Sprintf("%s%s %s = %d", inner, field.Type, field.Name, field.Number))
		}
		if len(field.Options) > 0 {
			sb.WriteString(" [")
			for i, opt := range field.Options {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(fmt.Sprintf("%s = %s", opt.Name, opt.Value))
			}
			sb.WriteString("]")
		}
		sb.WriteString(";\n")
	}

	// Maps
	for _, mf := range msg.Maps {
		sb.WriteString(fmt.Sprintf("%smap<%s, %s> %s = %d;\n", inner, mf.KeyType, mf.ValueType, mf.Name, mf.Number))
	}

	// Oneofs
	for _, oo := range msg.OneOfs {
		sb.WriteString(fmt.Sprintf("%soneof %s {\n", inner, oo.Name))
		ooInner := inner + opts.Indent
		for _, field := range oo.Fields {
			sb.WriteString(fmt.Sprintf("%s%s %s = %d;\n", ooInner, field.Type, field.Name, field.Number))
		}
		sb.WriteString(fmt.Sprintf("%s}\n", inner))
	}

	// Extensions
	for _, ext := range msg.Extensions {
		if ext.RangeEnd == -1 {
			sb.WriteString(fmt.Sprintf("%sextensions %d to max;\n", inner, ext.RangeStart))
		} else if ext.RangeStart == ext.RangeEnd {
			sb.WriteString(fmt.Sprintf("%sextensions %d;\n", inner, ext.RangeStart))
		} else {
			sb.WriteString(fmt.Sprintf("%sextensions %d to %d;\n", inner, ext.RangeStart, ext.RangeEnd))
		}
	}

	// Options
	for _, opt := range msg.Options {
		sb.WriteString(fmt.Sprintf("%soption %s = %s;\n", inner, opt.Name, formatValue(opt.Value)))
	}

	sb.WriteString(fmt.Sprintf("%s}\n", indent))
	return sb.String()
}

func formatEnum(enum *models.Enum, indent string, opts *Options) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%senum %s {\n", indent, enum.Name))
	inner := indent + opts.Indent

	for _, val := range enum.Values {
		sb.WriteString(fmt.Sprintf("%s%s = %d;\n", inner, val.Name, val.Number))
	}

	// Options
	for _, opt := range enum.Options {
		sb.WriteString(fmt.Sprintf("%soption %s = %s;\n", inner, opt.Name, formatValue(opt.Value)))
	}

	sb.WriteString(fmt.Sprintf("%s}\n", indent))
	return sb.String()
}

func formatService(svc *models.Service, opts *Options) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("service %s {\n", svc.Name))
	inner := opts.Indent

	// Options
	for _, opt := range svc.Options {
		sb.WriteString(fmt.Sprintf("%soption %s = %s;\n", inner, opt.Name, formatValue(opt.Value)))
	}

	// Methods
	for _, method := range svc.Methods {
		clientStream := ""
		if method.ClientStreaming {
			clientStream = "stream "
		}
		serverStream := ""
		if method.ServerStreaming {
			serverStream = "stream "
		}

		sb.WriteString(fmt.Sprintf("%srpc %s (%s%s) returns (%s%s)", inner, method.Name, clientStream, method.InputType, serverStream, method.OutputType))

		if len(method.Options) > 0 {
			sb.WriteString(" {\n")
			methodInner := inner + opts.Indent
			for _, opt := range method.Options {
				sb.WriteString(fmt.Sprintf("%soption %s = %s;\n", methodInner, opt.Name, formatValue(opt.Value)))
			}
			sb.WriteString(fmt.Sprintf("%s}\n", inner))
		} else {
			sb.WriteString(";\n")
		}
	}

	sb.WriteString("}\n")
	return sb.String()
}

func formatValue(val string) string {
	// If it's already quoted, return as is
	if strings.HasPrefix(val, "\"") || strings.HasPrefix(val, "'") {
		return val
	}
	// If it looks like a string (contains non-numeric chars), quote it
	if strings.ContainsAny(val, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ./") && !strings.HasPrefix(val, "true") && !strings.HasPrefix(val, "false") {
		return fmt.Sprintf("\"%s\"", val)
	}
	return val
}

// Minify produces a minified version of the proto file (no comments, minimal whitespace).
func Minify(pf *models.ProtoFile) string {
	opts := &Options{
		Indent:     "",
		LineWidth:  0,
		SortFields: false,
	}
	return Format(pf, opts)
}
