// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package printer

import (
	"context"
	"strings"
	"time"
)

var (
	progressIndicators = []string{"-", "\\", "|", "/"}
	// HideCursor writes the sequence for hiding cursor
	hideCursor = "\x1b[?25l"
	// ShowCursor writes the sequence for resotring show cursor
	showCursor = "\x1b[?25h"
)

// StartProgress returns a cancel function to stop progress indicator
func StartSimpleProgress(c context.Context, message string) func() {
	w := printer.eWriter
	ctx, cancel := context.WithCancel(c)

	if w == nil {
		return cancel
	}

	go func() {
		ticker := time.Tick(100 * time.Millisecond)
		_, _ = w.Write([]byte(message + " "))
		defer func() {
			_, _ = w.Write([]byte(strings.Repeat(string(rune(KeyCtrlH)), len(message+" "))))
		}()

		_, _ = w.Write([]byte(hideCursor))
		defer w.Write([]byte(showCursor)) // nolint:errcheck

		s := progressIndicators[0]
		_, _ = w.Write([]byte(s))

		for i := 1; ; i++ {
			_, _ = w.Write([]byte(strings.Repeat(string(rune(KeyCtrlH)), len(s))))
			select {
			case <-ctx.Done():
				return
			case <-ticker:
				s = progressIndicators[i%len(progressIndicators)]
				_, _ = w.Write([]byte(s))
			}
		}
	}()

	return cancel
}
