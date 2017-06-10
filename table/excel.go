package table

import (
	"github.com/pkg/errors"
	"github.com/tealeg/xlsx"
)

type excelTable struct {
	sheet *xlsx.Sheet
	row   int
}

func newExcelTable(file string) (Table, error) {
	xlsxFile, err := xlsx.OpenFile(file)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read excel file")
	}
	if len(xlsxFile.Sheets) != 1 {
		return nil, errors.New("excel file has multiple sheets:" + file)
	}
	excelTable := excelTable{
		sheet: xlsxFile.Sheets[0],
		row:   1, // Header consumes 1 row
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
