package client

import (
	"fmt"
	"io"
	"strings"
)

// IndentWriter indents print.
type IndentWriter struct {
	writer io.Writer
	indent string
}

func NewIndentWriter(writer io.Writer) *IndentWriter {
	return &IndentWriter{writer: writer}
}

func (w *IndentWriter) IncrementLevel() {
	w.indent = "  " + w.indent
}

func (w *IndentWriter) DecrementLevel() {
	if len(w.indent) < 2 {
		w.indent = ""
	} else {
		w.indent = w.indent[0 : len(w.indent)-2]
	}
}

func (w *IndentWriter) printIndent() {
	fmt.Fprint(w.writer, w.indent)
}

func (w *IndentWriter) Println(s interface{}) {
	w.printIndent()
	fmt.Fprintln(w.writer, s)
}

func (w *IndentWriter) PrintMultiline(s string) {
	for _, line := range strings.Split(s, "\n") {
		w.Println(line)
	}
}

func (w *IndentWriter) Printf(format string, s ...interface{}) {
	w.printIndent()
	fmt.Fprintf(w.writer, format, s...)
}
