// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package printer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/template"
)

const (
	FormatPlain = "plain"
	FormatJSON  = "json"
)

type Printer struct { //nolint
	writer  io.Writer
	eWriter io.Writer

	Format     string
	Single     bool
	Lines      []interface{}
	ErrorLines []interface{}
}

var printer Printer

func init() {
	printer.writer = os.Stdout
	printer.eWriter = os.Stderr
}

// SetFormat sets the format for the final output of the printer
func SetFormat(t string) {
	printer.Format = t
}

// SetSingle sets the single flag on the printer. If this flag is set, the
// printer will check the size of stored elements before printing, and
// if there is only one, it will be printed on its own instead of
// inside a list
func SetSingle(single bool) {
	printer.Single = single
}

// PrintT prints an element. Depending on the format, the element can be
// formatted and printed as a structure or used to populate the
// template
func PrintT(templateString string, v interface{}) {
	switch printer.Format {
	case FormatPlain:
		t := template.Must(template.New("").Parse(templateString))
		var tpl bytes.Buffer
		if err := t.Execute(&tpl, v); err != nil {
			PrintError("Can't print the message using the provided template: " + templateString)
		}
		tplString := tpl.String()
		printer.Lines = append(printer.Lines, tplString)
		fmt.Fprintln(printer.writer, tplString)
	case FormatJSON:
		printer.Lines = append(printer.Lines, v)
	}
}

// Print an element. If the format requires a template, the element
// will be printed as a structure with field names using the print
// verb %+v
func Print(v interface{}) {
	PrintT("{{printf \"%+v\" .}}", v)
}

// Flush writes the elements accumulated in the printer
func Flush() {
	if printer.Format == FormatJSON {
		var b []byte
		switch {
		case printer.Single && len(printer.Lines) == 0:
			return
		case printer.Single && len(printer.Lines) == 1:
			b, _ = json.MarshalIndent(printer.Lines[0], "", "  ")
		default:
			b, _ = json.MarshalIndent(printer.Lines, "", "  ")
		}

		fmt.Fprintln(printer.writer, string(b))
		printer.Lines = []interface{}{}
	}
}

// Clean resets the printer's accumulated lines
func Clean() {
	printer.Lines = []interface{}{}
	printer.ErrorLines = []interface{}{}
}

// GetLines returns the printer's accumulated lines
func GetLines() []interface{} {
	return printer.Lines
}

// GetErrorLines returns the printer's accumulated error lines
func GetErrorLines() []interface{} {
	return printer.ErrorLines
}

// PrintError prints to the stderr.
func PrintError(msg string) {
	printer.ErrorLines = append(printer.ErrorLines, msg)
	fmt.Fprintln(printer.eWriter, msg)
}
