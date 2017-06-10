package table

import (
	"errors"
	"io"
	"path/filepath"
	"strings"
)

// New creates new Table from a file.
func New(file string) (Table, error) {
	switch strings.ToLower(filepath.Ext(file)) {
	case ".xlsx":
		return newExcelTable(file)
	case ".csv":
		return newCsvTable(file)
	default:
		return nil, errors.New("unsupported file: " + file)
	}
}

// Table is an interface represents table structure data.
type Table interface {
	io.Closer
	ReadRow(row *[]string) bool
}
