package table

import (
	"errors"
	"io"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	// ConsoleOutput represents a special file writes to console.
	ConsoleOutput = "<console>"
)

const (
	CommonHeaderSize = iota
	CsvSeparator
)

// NewReader creates new table reader from a file.
func NewReader(file string, options map[int]string) (Reader, error) {
	var reader Reader
	var err error
	switch strings.ToLower(filepath.Ext(file)) {
	case ".xlsx":
		reader, err = newExcelTableReader(file)
	case ".csv":
		separator := options[CsvSeparator]
		reader, err = newCsvTableReader(file, separator)
	default:
		return nil, errors.New("unsupported file: " + file)
	}
	if err != nil {
		headerSize, err := strconv.Atoi(options[CommonHeaderSize])
		if err != nil {
			dummy := []string{}
			for i := 0; i < headerSize; i++ {
				reader.ReadRow(&dummy)
			}
		}
	}
	return reader, err
}

// NewWriter creates new table writer for the file
func NewWriter(file string, options map[int]string) (Writer, error) {
	if len(file) == 0 || file == ConsoleOutput {
		return newConsoleTableWriter(), nil
	}
	switch strings.ToLower(filepath.Ext(file)) {
	case ".xlsx":
		return newExcelTableWriter(file)
	case ".csv":
		separator := options[CsvSeparator]
		return newCsvTableWriter(file, separator)
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
