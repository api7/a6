package cmdutil

import (
	"encoding/json"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// Exporter formats and writes data to the given writer.
type Exporter struct {
	format string
	writer io.Writer
}

// NewExporter creates a new Exporter for the given format and writer.
// Supported formats: "json", "yaml".
func NewExporter(format string, writer io.Writer) *Exporter {
	return &Exporter{
		format: format,
		writer: writer,
	}
}

// ValidateExportFormat returns an error if format is not one of the formats
// that Exporter.Write accepts. The empty string is treated as valid because
// every export command defaults it to "yaml" later in its run.
//
// Callers should invoke this before any work that could short-circuit the
// run (e.g. an "empty collection" early return), so an invalid -o flag is
// rejected consistently regardless of the result set size.
func ValidateExportFormat(format string) error {
	switch format {
	case "", "json", "yaml":
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

// Write formats and writes the given data.
func (e *Exporter) Write(data interface{}) error {
	switch e.format {
	case "json":
		return e.writeJSON(data)
	case "yaml":
		return e.writeYAML(data)
	default:
		return fmt.Errorf("unsupported output format: %s", e.format)
	}
}

func (e *Exporter) writeJSON(data interface{}) error {
	enc := json.NewEncoder(e.writer)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

func (e *Exporter) writeYAML(data interface{}) error {
	enc := yaml.NewEncoder(e.writer)
	defer enc.Close()
	return enc.Encode(data)
}
