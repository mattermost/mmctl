// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package importer

import (
	"fmt"
	"strings"
)

type ImportFileInfo struct {
	ArchiveName string `json:"archive_name"`
	FileName    string `json:"file_name,omitempty"`
	CurrentLine uint64 `json:"current_line,omitempty"`
	TotalLines  uint64 `json:"total_lines,omitempty"`
}

type ImportValidationError struct { //nolint:govet
	ImportFileInfo
	FieldName       string          `json:"field_name,omitempty"`
	Err             error           `json:"error"`
	Suggestion      string          `json:"suggestion,omitempty"`
	SuggestedValues []any           `json:"suggested_values,omitempty"`
	ApplySuggestion func(any) error `json:"-"`
}

func (e *ImportValidationError) Error() string {
	msg := &strings.Builder{}
	msg.WriteString("import validation error")

	if e.FileName != "" || e.ArchiveName != "" {
		fmt.Fprintf(msg, " in %q->%q:%d", e.ArchiveName, e.FileName, e.CurrentLine)
	}

	if e.FieldName != "" {
		fmt.Fprintf(msg, " field %q", e.FieldName)
	}

	if e.Err != nil {
		fmt.Fprintf(msg, ": %s", e.Err)
	}

	return msg.String()
}
