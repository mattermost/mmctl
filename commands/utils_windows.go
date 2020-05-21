// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"fmt"
	"os"
)

func checkValidSocket(socketPath string) error {
	// check file mode and permissions
	fi, err := os.Stat(socketPath)
	if err != nil && os.IsNotExist(err) {
		return fmt.Errorf("socket file %q doesn't exists, please check the server configuration for local mode", socketPath)
	} else if err != nil {
		return err
	}
	if fi.Mode() != expectedSocketMode {
		return fmt.Errorf("invalid file mode for file %q, it must be a socket with 0600 permissions", socketPath)
	}

	return nil
}
