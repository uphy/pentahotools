package table

import (
	"github.com/pkg/errors"
	"github.com/tealeg/xlsx"
)

type excelTable struct {
	sheet *xlsx.Sheet
	row   int
}

func newExcelTableReader(file string) (Reader, error) {
	xlsxFile, err := xlsx.OpenFile(file)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read excel file")
	}
	if len(xlsxFile.Sheets) != 1 {
		return nil, errors.New("excel file has multiple sheets:" + file)
	}
	excelTable := excelTable{
		sheet: xlsxFile.Sheets[0],
		row:   0,
	}
	return &excelTable, nil
}

func (t *excelTable) ReadRow(row *[]string) bool {
	if t.row >= len(t.sheet.Rows) {
		return false
	}
	excelRow := t.sheet.Rows[t.row]
	for i := 0; i < len(*row); i++ {
		if i < len(excelRow.Cells) {
			(*row)[i], _ = excelRow.Cells[i].String()
		} else {
			(*row)[i] = ""
		}
	}
	t.row++
	return true
}
func (t *excelTable) Close() error {
	return nil
}

type excelTableWriter struct {
	file  *xlsx.File
	sheet *xlsx.Sheet
	path  string
}

func newExcelTableWriter(file string) (Writer, error) {
	xlsxFile := xlsx.NewFile()
	sheet, err := xlsxFile.AddSheet("UserList")
	if err != nil {
		return nil, err
	}
	return &excelTableWriter{xlsxFile, sheet, file}, nil
}

func (c *excelTableWriter) WriteHeader(row *[]string) error {
	c.writeRow(row, true)
	return nil
}

func (c *excelTableWriter) WriteRow(row *[]string) error {
	c.writeRow(row, false)
	return nil
}

func (c *excelTableWriter) writeRow(row *[]string, isHeader bool) {
	r := c.sheet.AddRow()
	for _, s := range *row {
		cell := r.AddCell()
		cell.SetString(s)
		style := cell.GetStyle()
		style.Border = *xlsx.NewBorder("thin", "thin", "thin", "thin")
		if isHeader {
			style.Font.Bold = true
			style.Alignment.Horizontal = "center"
		}
	}
}

func setHeaderString(cell *xlsx.Cell, header string) {
	setString(cell, header)
}

func setString(cell *xlsx.Cell, value string) {
	cell.SetString(value)
	cell.GetStyle().Border = *xlsx.NewBorder("thin", "thin", "thin", "thin")
}

func (c *excelTableWriter) Close() error {
	for _, col := range c.sheet.Cols {
		col.Width = 30
	}
	return c.file.Save(c.path)
}
