package table

import (
	"errors"
	"io"
	"path/filepath"
	"strings"
)

const (
	// ConsoleOutput represents a special file writes to console.
	ConsoleOutput = "<console>"
)

// NewReader creates new table reader from a file.
func NewReader(file string) (Reader, error) {
	switch strings.ToLower(filepath.Ext(file)) {
	case ".xlsx":
		return newExcelTableReader(file)
	case ".csv":
		return newCsvTableReader(file)
	default:
		return nil, errors.New("unsupported file: " + file)
	}
}

// NewWriter creates new table writer for the file
func NewWriter(file string) (Writer, error) {
	if len(file) == 0 || file == ConsoleOutput {
		return newConsoleTableWriter(), nil
	}
	switch strings.ToLower(filepath.Ext(file)) {
	case ".xlsx":
		return newExcelTableWriter(file)
	case ".csv":
		return newCsvTableWriter(file)
	}
	return nil, errors.New("unsupported file: " + file)
}

// Reader is an interface provides feature to read tables.
type Reader interface {
	io.Closer
	ReadRow(row *[]string) bool
}

// Writer is an interface provides feature to write tables.
type Writer interface {
	io.Closer
	WriteHeader(row *[]string) error
	WriteRow(row *[]string) error
}
