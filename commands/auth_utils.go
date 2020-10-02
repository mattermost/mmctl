// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const (
	MethodPassword = "P"
	MethodToken    = "T"
	MethodMFA      = "M"
)

const (
	configFileName = ".mmctl"
	xdgConfigHome  = ".config"
)

type Credentials struct {
	Name        string `json:"name"`
	Username    string `json:"username"`
	AuthToken   string `json:"authToken"`
	AuthMethod  string `json:"authMethod"`
	InstanceURL string `json:"instanceUrl"`
	Active      bool   `json:"active"`
}

type CredentialsList map[string]*Credentials

func getDefaultConfigPath() string {
	return filepath.Join("$HOME", xdgConfigHome, "mmctl")
}

func resolveConfigFilePath() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}

	configPath := strings.Replace(viper.GetString("config"), "$HOME", u.HomeDir, 1)

	// probe user-defined or xdg-compliant default config path
	if _, err := os.Stat(filepath.Join(configPath, configFileName)); err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}
		// conditionally fall back to legacy config path
		configPath = u.HomeDir
	}

	return filepath.Join(configPath, configFileName), nil
}

func ReadCredentialsList() (*CredentialsList, error) {
	configPath, err := resolveConfigFilePath()
	if err != nil {
		return nil, err
	}

	if _, err = os.Stat(configPath); err != nil {
		return nil, errors.WithMessage(err, "cannot read user credentials, maybe you need to use login first")
	}

	fileContents, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, errors.WithMessage(err, "there was a problem reading the credentials file")
	}

	var credentialsList CredentialsList
	if err := json.Unmarshal(fileContents, &credentialsList); err != nil {
		return nil, errors.WithMessage(err, "there was a problem parsing the credentials file")
	}

	return &credentialsList, nil
}

func GetCurrentCredentials() (*Credentials, error) {
	credentialsList, err := ReadCredentialsList()
	if err != nil {
		return nil, err
	}

	for _, c := range *credentialsList {
		if c.Active {
			return c, nil
		}
	}
	return nil, errors.Errorf("no current context available. please use the %q command", "auth set")
}

func GetCredentials(name string) (*Credentials, error) {
	credentialsList, err := ReadCredentialsList()
	if err != nil {
		return nil, err
	}

	for _, c := range *credentialsList {
		if c.Name == name {
			return c, nil
		}
	}
	return nil, errors.Errorf("couldn't find credentials for connection %q", name)
}

func SaveCredentials(credentials Credentials) error {
	credentialsList, err := ReadCredentialsList()
	if err != nil {
		credentialsList = &CredentialsList{}
		credentials.Active = true
	}

	(*credentialsList)[credentials.Name] = &credentials
	return SaveCredentialsList(credentialsList)
}

func SaveCredentialsList(credentialsList *CredentialsList) error {
	configPath, err := resolveConfigFilePath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(configPath); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if err := os.MkdirAll(configPath, 0700); err != nil {
			return errors.WithMessage(err, "cannot create config directory")
		}
	}

	marshaledCredentialsList, _ := json.MarshalIndent(credentialsList, "", "    ")

	if err := ioutil.WriteFile(filepath.Join(configPath, configFileName), marshaledCredentialsList, 0600); err != nil {
		return errors.WithMessage(err, "cannot save the credentials")
	}

	return nil
}

func SetCurrent(name string) error {
	credentialsList, err := ReadCredentialsList()
	if err != nil {
		return err
	}

	found := false
	for _, c := range *credentialsList {
		if c.Name == name {
			found = true
			c.Active = true
		} else {
			c.Active = false
		}
	}

	if !found {
		return errors.Errorf("cannot find credentials for server %q", name)
	}

	return SaveCredentialsList(credentialsList)
}

func CleanCredentials() error {
	configFilePath, err := resolveConfigFilePath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(configFilePath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if err := os.Remove(configFilePath); err != nil {
		return err
	}
	return nil
}
