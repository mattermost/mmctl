// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package printer

import (
	"bytes"
	"errors"
	"os"

	"golang.org/x/term"
)

func checkInteractiveTerminal() error {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return err
	}

	if (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		return errors.New("this is not an interactive shell")
	}

	return nil
}

func termHeight(file *os.File) int {
	_, h, err := term.GetSize(int(file.Fd()))
	if err != nil {
		panic(err)
	}

	return h
}

func lineCount(b []byte) int {
	return bytes.Count(b, []byte{'\n'})
}
