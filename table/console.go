package table

import (
	"os"

	"github.com/olekukonko/tablewriter"
)

// ConsoleTableWriter writes the table to the console.
type consoleTableWriter struct {
	table *tablewriter.Table
}

func newConsoleTableWriter() Writer {
	return &consoleTableWriter{
		table: tablewriter.NewWriter(os.Stdout),
	}
}

func (c *consoleTableWriter) WriteHeader(row *[]string) error {
	c.table.SetHeader(*row)
	return nil
}

func (c *consoleTableWriter) WriteRow(row *[]string) error {
	c.table.Append(*row)
	return nil
}

func (c *consoleTableWriter) Close() error {
	c.table.Render()
	return nil
}
