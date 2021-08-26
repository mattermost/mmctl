// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

//go:build linux || darwin
// +build linux darwin

package commands

import (
	"fmt"
	"os"
	"os/user"
	"syscall"
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

	// check matching user
	cUser, err := user.Current()
	if err != nil {
		return err
	}
	s, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("cannot get owner of the file %q", socketPath)
	}
	if fmt.Sprint(s.Uid) != cUser.Uid {
		return fmt.Errorf("owner of the file %q must be the same user running mmctl", socketPath)
	}

	return nil
}
