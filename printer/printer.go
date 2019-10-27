package printer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"text/template"
)

const (
	FORMAT_PLAIN = "plain"
	FORMAT_JSON  = "json"
)

type Printer struct {
	Format     string
	Single     bool
	Lines      []interface{}
	ErrorLines []interface{}
}

var printer Printer

// Sets the format for the final output of the printer
func SetFormat(t string) {
	printer.Format = t
}

// Sets the single flag on the printer. If this flag is set, the
// printer will check the size of stored elements before printing, and
// if there is only one, it will be printed on its own instead of
// inside a list
func SetSingle(single bool) {
	printer.Single = single
}

// Prints an element. Depending on the format, the element can be
// formatted and printed as a structure or used to populate the
// template
func PrintT(templateString string, v interface{}) {
	switch printer.Format {
	case FORMAT_PLAIN:
		t := template.Must(template.New("").Parse(templateString))
		var tpl bytes.Buffer
		t.Execute(&tpl, v)
		tplString := tpl.String()
		printer.Lines = append(printer.Lines, tplString)
		fmt.Println(tplString)
	case FORMAT_JSON:
		printer.Lines = append(printer.Lines, v)
	}
}

// Prints an element. If the format requires a template, the element
// will be printed as a structure with field names using the print
// verb %+v
func Print(v interface{}) {
	PrintT("{{printf \"%+v\" .}}", v)
}

// Prints the elements accumulated in the printer
func Flush() {
	if printer.Format == FORMAT_JSON {
		var b []byte
		if printer.Single && len(printer.Lines) == 1 {
			b, _ = json.MarshalIndent(printer.Lines[0], "", "  ")
		} else {
			b, _ = json.MarshalIndent(printer.Lines, "", "  ")
		}

		fmt.Println(string(b))
		printer.Lines = []interface{}{}
	}
}

// Resets the printer's cummulated lines
func Clean() {
	printer.Lines = []interface{}{}
	printer.ErrorLines = []interface{}{}
}

// Returns the printer's cummulated lines
func GetLines() []interface{} {
	return printer.Lines
}

// Returns the printer's cummulated error lines
func GetErrorLines() []interface{} {
	return printer.ErrorLines
}

// Prints an error string to the stderr.
func PrintError(msg string) {
	printer.ErrorLines = append(printer.ErrorLines, msg)
	fmt.Fprintln(os.Stderr, msg)
}
