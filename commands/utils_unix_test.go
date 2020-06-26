// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"io/ioutil"
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckValidSocket(t *testing.T) {
	t.Run("should return error if the file is not a socket", func(t *testing.T) {
		f, err := ioutil.TempFile(os.TempDir(), "mmctl_socket_")
		require.Nil(t, err)
		defer os.Remove(f.Name())
		require.Nil(t, os.Chmod(f.Name(), 0600))

		require.Error(t, checkValidSocket(f.Name()))
	})

	t.Run("should return error if the file has not the right permissions", func(t *testing.T) {
		f, err := ioutil.TempFile(os.TempDir(), "mmctl_socket_")
		require.Nil(t, err)
		require.Nil(t, os.Remove(f.Name()))

		s, err := net.Listen("unix", f.Name())
		require.Nil(t, err)
		defer s.Close()
		require.Nil(t, os.Chmod(f.Name(), 0777))

		require.Error(t, checkValidSocket(f.Name()))
	})

	t.Run("should return nil if the file is a socket and has the right permissions", func(t *testing.T) {
		f, err := ioutil.TempFile(os.TempDir(), "mmctl_socket_")
		require.Nil(t, err)
		require.Nil(t, os.Remove(f.Name()))

		s, err := net.Listen("unix", f.Name())
		require.Nil(t, err)
		defer s.Close()
		require.Nil(t, os.Chmod(f.Name(), 0600))

		require.Nil(t, checkValidSocket(f.Name()))
	})
}
