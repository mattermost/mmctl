// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package printer

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockWriter struct {
	buffer []byte
}

func (w *mockWriter) Write(b []byte) (n int, err error) {
	w.buffer = append(w.buffer, b...)
	return len(w.buffer) - len(b), nil
}

func TestPrintT(t *testing.T) {
	w := bufio.NewWriter(&bytes.Buffer{})
	printer.writer = w
	printer.Format = FormatPlain

	ts := struct {
		ID int
	}{
		ID: 123,
	}

	t.Run("should execute template", func(t *testing.T) {
		tpl := `testing template {{.ID}}`
		PrintT(tpl, ts)
		assert.Equal(t, "testing template 123", printer.Lines[0], "should be equal")
	})

	t.Run("should fail to execute, no method or field", func(t *testing.T) {
		Clean()
		tpl := `testing template {{.Name}}`
		PrintT(tpl, ts)
		assert.Equal(t, "Can't print the message using the provided template: "+tpl, printer.ErrorLines[0])
	})
}

func TestFlush(t *testing.T) {
	mw := &mockWriter{}
	printer.writer = mw
	printer.Format = FormatJSON

	t.Run("should print a line in JSON format", func(t *testing.T) {
		mw.buffer = []byte{}
		Clean()

		Print("test string")
		Flush()

		assert.Equal(t, "[\n  \"test string\"\n]\n", string(mw.buffer))
	})

	t.Run("should print multi line in JSON format", func(t *testing.T) {
		mw.buffer = []byte{}
		Clean()

		Print("test string-1")
		Print("test string-2")
		Flush()
		assert.Equal(t, "[\n  \"test string-1\",\n  \"test string-2\"\n]\n", string(mw.buffer))
	})
}
