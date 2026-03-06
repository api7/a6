package tableprinter

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// TablePrinter renders data as an aligned table.
type TablePrinter struct {
	writer  *tabwriter.Writer
	out     io.Writer
	headers []string
	rows    [][]string
}

// New creates a new TablePrinter that writes to the given writer.
func New(out io.Writer) *TablePrinter {
	return &TablePrinter{
		out: out,
	}
}

// SetHeaders sets the column headers.
func (p *TablePrinter) SetHeaders(headers ...string) {
	p.headers = headers
}

// AddRow adds a row of values.
func (p *TablePrinter) AddRow(values ...string) {
	p.rows = append(p.rows, values)
}

// Render writes the table to the output. Returns an error if the table
// cannot be flushed.
func (p *TablePrinter) Render() error {
	if len(p.rows) == 0 && len(p.headers) == 0 {
		return nil
	}

	p.writer = tabwriter.NewWriter(p.out, 0, 0, 3, ' ', 0)

	if len(p.headers) > 0 {
		fmt.Fprintln(p.writer, strings.Join(p.headers, "\t"))
	}

	for _, row := range p.rows {
		fmt.Fprintln(p.writer, strings.Join(row, "\t"))
	}

	return p.writer.Flush()
}

// RowCount returns the number of data rows (excluding headers).
func (p *TablePrinter) RowCount() int {
	return len(p.rows)
}
