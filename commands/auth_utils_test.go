package commands

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"os"
	"os/user"
	"path/filepath"
	"testing"
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

	t.Run("should return the default xdg config location if it exists and nothing else is set", func(t *testing.T) {
		testUser.HomeDir = t.TempDir()
		SetUser(testUser)

		expected := fmt.Sprintf("%s/.config/mmctl/.mmctl", testUser.HomeDir)

		viper.Set("config", getDefaultConfigPath())
		if err := createFile(expected); err != nil {
			panic(err)
		}

		p, err := resolveConfigFilePath()
		require.Nil(t, err)
		require.Equal(t, expected, p)
	})

	t.Run("should return legacy config location if default one does not exists", func(t *testing.T) {
		testUser.HomeDir = t.TempDir()
		SetUser(testUser)

		expected := fmt.Sprintf("%s/.mmctl", testUser.HomeDir)

		viper.Set("config", getDefaultConfigPath())

		p, err := resolveConfigFilePath()
		require.Nil(t, err)
		require.Equal(t, expected, p)
	})

	t.Run("should return config location from xdg environment variable", func(t *testing.T) {
		testUser.HomeDir = t.TempDir()
		SetUser(testUser)

		expected := fmt.Sprintf("%s/test/.mmctl", testUser.HomeDir)

		os.Setenv("XDG_CONFIG_HOME", filepath.Dir(expected))
		viper.Set("config", getDefaultConfigPath())
		if err := createFile(expected); err != nil {
			panic(err)
		}

		p, err := resolveConfigFilePath()
		require.Nil(t, err)
		require.Equal(t, expected, p)
	})

	t.Run("should return the user-defined cofnig path if one is set", func(t *testing.T) {
		tmp := t.TempDir()
		testUser.HomeDir = "path/should/be/ignored"
		SetUser(testUser)

		expected := fmt.Sprintf("%s/.mmctl", tmp)

		os.Setenv("XDG_CONFIG_HOME", "path/should/be/ignored")
		viper.Set("config", tmp)
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
