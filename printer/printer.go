package printer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"
)

const (
	FORMAT_PLAIN = "plain"
	FORMAT_JSON  = "json"
)

type Printer struct {
	Format string
	Single bool
	Lines  []interface{}
}

var printer Printer

func SetFormat(t string) {
	printer.Format = t
}

func SetSingle(single bool) {
	printer.Single = single
}

func PrintT(templateString string, v interface{}) {
	switch printer.Format {
	case FORMAT_PLAIN:
		t := template.Must(template.New("").Parse(templateString))
		var tpl bytes.Buffer
		t.Execute(&tpl, v)
		fmt.Println(tpl.String())
	case FORMAT_JSON:
		printer.Lines = append(printer.Lines, v)
	}
}

func Print(v interface{}) {
	PrintT("{{printf \"%+v\" .}}", v)
}

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
