package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"
)

const (
	LOGGER_TYPE_PLAIN = "plain"
	LOGGER_TYPE_JSON  = "json"
)

type Logger struct {
	LogType string
	Lines   []interface{}
}

func (l *Logger) SetType(t string) {
	l.LogType = t
}

func (l *Logger) PrintT(templateString string, v interface{}) {
	switch l.LogType {
	case LOGGER_TYPE_PLAIN:
		t := template.Must(template.New("").Parse(templateString))
		var tpl bytes.Buffer
		t.Execute(&tpl, v)
		fmt.Println(tpl.String())
	case LOGGER_TYPE_JSON:
		l.Lines = append(l.Lines, v)
	}
}

func (l *Logger) Print(v interface{}) {
	l.PrintT("{{printf \"%+v\" .}}", v)
}

func (l *Logger) Flush() {
	if l.LogType == LOGGER_TYPE_JSON {
		b, _ := json.MarshalIndent(l.Lines, "", "  ")
		fmt.Println(string(b))
		l.Lines = []interface{}{}
	}
}

var Log Logger
