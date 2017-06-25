package table

import (
	"encoding/csv"
	"os"

	"github.com/pkg/errors"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

type csvTableReader struct {
	file   *os.File
	reader *csv.Reader
}

func newCsvTableReader(file string) (Reader, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, errors.Wrap(err, "csv file open failed")
	}
	reader := csv.NewReader(transform.NewReader(f, japanese.ShiftJIS.NewDecoder()))
	reader.LazyQuotes = true
	return &csvTableReader{
		file:   f,
		reader: reader,
	}, nil
}

func (t *csvTableReader) ReadRow(row *[]string) bool {
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

func (t *csvTableReader) Close() error {
	return t.file.Close()
}

type csvTableWriter struct {
	file   *os.File
	writer *csv.Writer
}

func newCsvTableWriter(file string) (Writer, error) {
	f, err := os.Create(file)
	if err != nil {
		return nil, errors.Wrap(err, "csv file open failed")
	}
	writer := csv.NewWriter(transform.NewWriter(f, japanese.ShiftJIS.NewDecoder()))
	return &csvTableWriter{f, writer}, nil
}

func (c *csvTableWriter) WriteHeader(row *[]string) error {
	return c.WriteRow(row)
}

func (c *csvTableWriter) WriteRow(row *[]string) error {
	return c.writer.Write(*row)
}

func (c *csvTableWriter) Close() error {
	c.writer.Flush()
	return c.file.Close()
}
