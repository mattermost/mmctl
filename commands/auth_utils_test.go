// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"fmt"
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
	if err != nil {
		panic(err)
	}

	t.Run("should return the default config location if nothing else is set", func(t *testing.T) {
		tmp, _ := ioutil.TempDir("", "mmctl-")
		defer os.RemoveAll(tmp)
		testUser.HomeDir = tmp
		SetUser(testUser)

		expected := fmt.Sprintf("%s/.mmctl", testUser.HomeDir)

		viper.Set("config-path", getDefaultConfigPath())
		if err := createFile(expected); err != nil {
			panic(err)
		}

		p, err := resolveConfigFilePath()
		require.Nil(t, err)
		require.Equal(t, expected, p)
	})

	t.Run("should return config location from xdg environment variable", func(t *testing.T) {
		tmp, _ := ioutil.TempDir("", "mmctl-")
		defer os.RemoveAll(tmp)
		testUser.HomeDir = tmp
		SetUser(testUser)

		expected := filepath.Join(testUser.HomeDir, ".config", "mmctl", ".mmctl")

		_ = os.Setenv("XDG_CONFIG_HOME", filepath.Join(testUser.HomeDir, ".config"))
		viper.Set("config-path", getDefaultConfigPath())
		if err := createFile(expected); err != nil {
			panic(err)
		}

		p, err := resolveConfigFilePath()
		require.Nil(t, err)
		require.Equal(t, expected, p)
	})

	t.Run("should return the user-defined cofnig path if one is set", func(t *testing.T) {
		tmp, _ := ioutil.TempDir("", "mmctl-")
		defer os.RemoveAll(tmp)

		testUser.HomeDir = "path/should/be/ignored"
		SetUser(testUser)

		expected := fmt.Sprintf("%s/.mmctl", tmp)

		_ = os.Setenv("XDG_CONFIG_HOME", "path/should/be/ignored")
		viper.Set("config-path", tmp)
		if err := createFile(expected); err != nil {
			panic(err)
		}

		p, err := resolveConfigFilePath()
		require.Nil(t, err)
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
