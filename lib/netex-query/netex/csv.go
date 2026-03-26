package netex

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
)

// CSVWriter wraps a csv.Writer with buffered I/O.
type CSVWriter struct {
	file   *os.File
	buf    *bufio.Writer
	writer *csv.Writer
	count  int
}

// NewCSVWriter creates a new CSV file with the given header row.
func NewCSVWriter(dir, filename string, header []string) (*CSVWriter, error) {
	path := filepath.Join(dir, filename)
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("create %s: %w", path, err)
	}
	buf := bufio.NewWriterSize(f, 64*1024)
	w := csv.NewWriter(buf)
	if err := w.Write(header); err != nil {
		f.Close()
		return nil, fmt.Errorf("write header %s: %w", path, err)
	}
	return &CSVWriter{file: f, buf: buf, writer: w}, nil
}

// WriteRow writes a single data row.
func (w *CSVWriter) WriteRow(row []string) {
	w.writer.Write(row)
	w.count++
}

// Close flushes and closes the file.
func (w *CSVWriter) Close() error {
	w.writer.Flush()
	if err := w.writer.Error(); err != nil {
		w.file.Close()
		return err
	}
	if err := w.buf.Flush(); err != nil {
		w.file.Close()
		return err
	}
	return w.file.Close()
}

// Count returns the number of data rows written.
func (w *CSVWriter) Count() int {
	return w.count
}

// TableDef defines one CSV output table.
// Used by both profiles and the CSV output system.
type TableDef struct {
	EntityType string   // maps to Entity.Type
	FileName   string   // e.g., "lines.csv"
	Columns    []Column // ordered columns
}

// Column defines one CSV column.
type Column struct {
	Header string // CSV header name
	Field  string // key in Entity.Fields
}

// CSVOutput manages CSV writers for all entity types, driven by table definitions.
type CSVOutput struct {
	writers map[string]*CSVWriter
	tables  map[string]*TableDef
}

// NewCSVOutput creates a CSVOutput with writers for each table definition.
func NewCSVOutput(outputDir string, tableDefs []TableDef) (*CSVOutput, error) {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return nil, fmt.Errorf("create output dir: %w", err)
	}

	o := &CSVOutput{
		writers: make(map[string]*CSVWriter, len(tableDefs)),
		tables:  make(map[string]*TableDef, len(tableDefs)),
	}

	for i := range tableDefs {
		td := &tableDefs[i]
		headers := make([]string, len(td.Columns))
		for j, col := range td.Columns {
			headers[j] = col.Header
		}
		w, err := NewCSVWriter(outputDir, td.FileName, headers)
		if err != nil {
			o.CloseAll()
			return nil, err
		}
		o.writers[td.EntityType] = w
		o.tables[td.EntityType] = td
	}

	return o, nil
}

// Handle processes a parsed entity, writing it to the appropriate CSV if the entity type is configured.
func (o *CSVOutput) Handle(e Entity) {
	td, ok := o.tables[e.Type]
	if !ok {
		return // profile doesn't output this entity type
	}
	w := o.writers[e.Type]
	row := make([]string, len(td.Columns))
	for i, col := range td.Columns {
		row[i] = e.Fields[col.Field]
	}
	w.WriteRow(row)
}

// CloseAll flushes and closes all CSV writers.
func (o *CSVOutput) CloseAll() {
	for _, w := range o.writers {
		w.Close()
	}
}

// PrintSummary prints row counts for each CSV file.
func (o *CSVOutput) PrintSummary() {
	for entityType, w := range o.writers {
		td := o.tables[entityType]
		fmt.Printf("  %-30s %s (%d rows)\n", entityType, td.FileName, w.Count())
	}
}
