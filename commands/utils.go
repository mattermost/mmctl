// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
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

func getConfirmation(question string, dbConfirmation bool) error {
	if err := checkInteractiveTerminal(); err != nil {
		return fmt.Errorf("could not proceed, either enable --confirm flag or use an interactive shell to complete operation: %w", err)
	}

	var confirm string
	if dbConfirmation {
		fmt.Println("Have you performed a database backup? (YES/NO): ")
		fmt.Scanln(&confirm)

		if confirm != "YES" {
			return errors.New("aborted: You did not answer YES exactly, in all capitals")
		}
	}

	fmt.Println(question + " (YES/NO): ")
	fmt.Scanln(&confirm)
	if confirm != "YES" {
		return errors.New("aborted: You did not answer YES exactly, in all capitals")
	}

	return nil
}
