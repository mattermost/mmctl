// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package importer

import (
	"fmt"
	"strings"
)

type ImportFileInfo struct {
	ArchiveName string
	FileName    string
	LineNumber  uint64
	TotalLines  uint64
}

type ImportValidationError struct {
	ImportFileInfo
	FieldName       string
	Err             error
	Suggestion      string
	SuggestedValues []any
	ApplySuggestion func(any) error
}

func (e *ImportValidationError) Error() string {
	msg := &strings.Builder{}
	msg.WriteString("import validation error")

	if e.FileName != "" || e.ArchiveName != "" {
		fmt.Fprintf(msg, " in %q->%q:%d", e.ArchiveName, e.FileName, e.LineNumber)
	}

	if e.FieldName != "" {
		fmt.Fprintf(msg, " field %q", e.FieldName)
	}

	if e.Err != nil {
		fmt.Fprintf(msg, ": %s", e.Err)
	}

	return msg.String()
}
