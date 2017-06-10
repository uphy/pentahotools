package table

import (
	"encoding/csv"
	"os"

	"github.com/pkg/errors"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

type csvTable struct {
	file   *os.File
	reader *csv.Reader
}

func newCsvTable(file string) (Table, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, errors.Wrap(err, "csv file open failed")
	}
	reader := csv.NewReader(transform.NewReader(f, japanese.ShiftJIS.NewDecoder()))
	reader.LazyQuotes = true
	// skip header row
	_, err = reader.Read()
	if err != nil {
		return nil, err
	}

	return &csvTable{
		file:   f,
		reader: reader,
	}, nil
}

func (t *csvTable) ReadRow(row *[]string) bool {
	r, err := t.reader.Read()
	if err != nil {
		return false
	}
	for i := 0; i < len(*row); i++ {
		if i < len(r) {
			(*row)[i] = r[i]
		} else {
			(*row)[i] = ""
		}
	}
	return true
}

func (t *csvTable) Close() error {
	return t.file.Close()
}
