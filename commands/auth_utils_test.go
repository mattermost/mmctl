// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestResolveConfigFilePath(t *testing.T) {
	originalUser := *currentUser
	defer func() {
		SetUser(&originalUser)
	}()

	testUser, err := user.Current()
	require.NoError(t, err)

	t.Run("should return the default config file location if nothing else is set", func(t *testing.T) {
		tmp, _ := ioutil.TempDir("", "mmctl-")
		defer os.RemoveAll(tmp)
		testUser.HomeDir = tmp
		SetUser(testUser)

		expected := filepath.Join(getDefaultConfigHomePath(), configParent, configFileName)

		viper.Set("config", filepath.Join(xdgConfigHomeVar, configParent, configFileName))

		p := resolveConfigFilePath()
		require.Equal(t, expected, p)
	})

	t.Run("should return config file location from xdg environment variable", func(t *testing.T) {
		tmp, _ := ioutil.TempDir("", "mmctl-")
		defer os.RemoveAll(tmp)
		testUser.HomeDir = tmp
		SetUser(testUser)

		expected := filepath.Join(testUser.HomeDir, ".config", configParent, configFileName)

		_ = os.Setenv("XDG_CONFIG_HOME", filepath.Join(testUser.HomeDir, ".config"))
		viper.Set("config", filepath.Join(xdgConfigHomeVar, configParent, configFileName))

		p := resolveConfigFilePath()
		require.Equal(t, expected, p)
	})

	t.Run("should return the user-defined config file path if one is set", func(t *testing.T) {
		tmp, _ := ioutil.TempDir("", "mmctl-")
		defer os.RemoveAll(tmp)

		testUser.HomeDir = "path/should/be/ignored"
		SetUser(testUser)

		expected := filepath.Join(tmp, configFileName)

		err := os.Setenv("XDG_CONFIG_HOME", "path/should/be/ignored")
		require.NoError(t, err)
		viper.Set("config", expected)

		p := resolveConfigFilePath()
		require.Equal(t, expected, p)
	})

	t.Run("should resolve config file path if $HOME variable is used", func(t *testing.T) {
		tmp, _ := ioutil.TempDir("", "mmctl-")
		defer os.RemoveAll(tmp)

		testUser.HomeDir = "path/should/be/ignored"
		SetUser(testUser)

		expected := filepath.Join(testUser.HomeDir, "/.config/mmctl/config")

		err := os.Setenv("XDG_CONFIG_HOME", "path/should/be/ignored")
		require.NoError(t, err)
		viper.Set("config", "$HOME/.config/mmctl/config")

		p := resolveConfigFilePath()
		require.Equal(t, expected, p)
	})
}
