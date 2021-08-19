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

	t.Run("should return the default config location if nothing else is set", func(t *testing.T) {
		tmp, _ := ioutil.TempDir("", "mmctl-")
		defer os.RemoveAll(tmp)
		testUser.HomeDir = tmp
		SetUser(testUser)

		viper.Set("config-path", getDefaultConfigPath())

		expected := filepath.Join(testUser.HomeDir, ".config", configFileName)

		err := createFile(expected)
		require.NoError(t, err)

		p := resolveConfigFilePath()
		require.Equal(t, expected, p)
	})

	t.Run("should return the home directory if config file exists there", func(t *testing.T) {
		tmp, _ := ioutil.TempDir("", "mmctl-")
		defer os.RemoveAll(tmp)
		testUser.HomeDir = tmp
		SetUser(testUser)

		expected := filepath.Join(testUser.HomeDir, "."+configFileName)
		// create $HOME/.mmctl
		err := createFile(expected)
		require.NoError(t, err)

		viper.Set("config-path", getDefaultConfigPath())

		p := resolveConfigFilePath()
		require.Equal(t, expected, p)
	})

	t.Run("should return config location from xdg environment variable", func(t *testing.T) {
		tmp, _ := ioutil.TempDir("", "mmctl-")
		defer os.RemoveAll(tmp)
		testUser.HomeDir = tmp
		SetUser(testUser)

		expected := filepath.Join(testUser.HomeDir, ".config", "mmctl")

		_ = os.Setenv("XDG_CONFIG_HOME", filepath.Join(testUser.HomeDir, ".config"))
		viper.Set("config-path", getDefaultConfigPath())

		err := createFile(expected)
		require.NoError(t, err)

		p := resolveConfigFilePath()
		require.Equal(t, expected, p)
	})

	t.Run("should return the user-defined config path if one is set", func(t *testing.T) {
		tmp, _ := ioutil.TempDir("", "mmctl-")
		defer os.RemoveAll(tmp)

		testUser.HomeDir = "path/should/be/ignored"
		SetUser(testUser)

		expected := filepath.Join(tmp, configFileName)

		err := os.Setenv("XDG_CONFIG_HOME", "path/should/be/ignored")
		require.NoError(t, err)
		viper.Set("config-path", tmp)

		err = createFile(expected)
		require.NoError(t, err)

		p := resolveConfigFilePath()
		require.Equal(t, expected, p)
	})
}

func createFile(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	if _, err := os.Create(path); err != nil {
		return err
	}
	return nil
}

func TestReadSecretFromFile(t *testing.T) {
	f, err := ioutil.TempFile(t.TempDir(), "mmctl")
	require.NoError(t, err)

	_, err = f.WriteString("test-pass")
	require.NoError(t, err)

	t.Run("password from file", func(t *testing.T) {
		var pass string
		err := readSecretFromFile(f.Name(), &pass)
		require.NoError(t, err)
		require.Equal(t, "test-pass", pass)
	})

	t.Run("no file path is provided", func(t *testing.T) {
		pass := "test-pass-2"
		err := readSecretFromFile("", &pass)
		require.NoError(t, err)
		require.Equal(t, "test-pass-2", pass)
	})

	t.Run("nonexistent file provided", func(t *testing.T) {
		var pass string
		err := readSecretFromFile(filepath.Join(t.TempDir(), "bla"), &pass)
		require.Error(t, err)
		require.Empty(t, pass)
	})
}
