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

	configFileName   = ".mmctl"
	xdgConfigHomeKey = "XDG_CONFIG_HOME"
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

var currentUser *user.User

func init() {
	newUser, err := user.Current()
	if err != nil {
		panic(err)
	}

	SetUser(newUser)
}

func getDefaultConfigPath() string {
	configPath := "$HOME"

	if p, ok := os.LookupEnv(xdgConfigHomeKey); ok {
		if err := os.MkdirAll(filepath.Join(p, "mmctl"), 0700); err == nil {
			configPath = filepath.Join(p, "mmctl")
		}
	}
	return configPath
}

func resolveConfigFilePath() (string, error) {
	configPath := strings.Replace(viper.GetString("config-path"), "$HOME", currentUser.HomeDir, 1)

	// check if config path exists
	if _, err := os.Stat(configPath); err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}
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

	marshaledCredentialsList, _ := json.MarshalIndent(credentialsList, "", "    ")

	if err := ioutil.WriteFile(configPath, marshaledCredentialsList, 0600); err != nil {
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

func SetUser(newUser *user.User) {
	currentUser = newUser
}
