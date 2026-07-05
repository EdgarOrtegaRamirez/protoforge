// Package output provides different output format renderers.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Format represents an output format.
type Format int

const (
	FormatText Format = iota
	FormatJSON
	FormatCompact
	FormatMarkdown
)

// ParseFormat parses a format string.
func ParseFormat(s string) Format {
	switch strings.ToLower(s) {
	case "json":
		return FormatJSON
	case "compact":
		return FormatCompact
	case "markdown", "md":
		return FormatMarkdown
	default:
		return FormatText
	}
}

// Writer writes output in the specified format.
type Writer struct {
	format Format
	w      io.Writer
}

// New creates a new Writer.
func New(w io.Writer, format Format) *Writer {
	return &Writer{format: format, w: w}
}

// WriteString writes a string to the output.
func (w *Writer) WriteString(s string) error {
	_, err := fmt.Fprint(w.w, s)
	return err
}

// WriteJSON writes JSON data.
func (w *Writer) WriteJSON(data interface{}) error {
	encoder := json.NewEncoder(w.w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// WriteCompactJSON writes compact JSON.
func (w *Writer) WriteCompactJSON(data interface{}) error {
	return json.NewEncoder(w.w).Encode(data)
}

// WriteError writes an error message.
func (w *Writer) WriteError(msg string) error {
	_, err := fmt.Fprintf(w.w, "Error: %s\n", msg)
	return err
}

// WriteSuccess writes a success message.
func (w *Writer) WriteSuccess(msg string) error {
	_, err := fmt.Fprintf(w.w, "✓ %s\n", msg)
	return err
}

// WriteWarning writes a warning message.
func (w *Writer) WriteWarning(msg string) error {
	_, err := fmt.Fprintf(w.w, "⚠ %s\n", msg)
	return err
}

// Format returns the current output format.
func (w *Writer) Format() Format {
	return w.format
}
