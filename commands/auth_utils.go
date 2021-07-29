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

	userHomeVar      = "$HOME"
	configFileName   = "mmctl"
	xdgConfigHomeVar = "$XDG_CONFIG_HOME"
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
	configPath := currentUser.HomeDir
	// We use the existing $HOME/.mmctl file if it exists.
	// If not, we try to read XDG_CONFIG_HOME and if we fail,
	// we fallback to $HOME/.config/mmctl.
	if _, err := os.Stat(filepath.Join(currentUser.HomeDir, "."+configFileName)); os.IsNotExist(err) {
		if p, ok := os.LookupEnv(strings.TrimPrefix(xdgConfigHomeVar, "$")); ok {
			configPath = p
		} else {
			configPath = filepath.Join(currentUser.HomeDir, ".config")
		}
	}
	return configPath
}

func resolveConfigFilePath() string {
	configPath := getDefaultConfigPath()
	if p := viper.GetString("config-path"); p != xdgConfigHomeVar {
		configPath = strings.Replace(viper.GetString("config-path"), userHomeVar, currentUser.HomeDir, 1)
	}

	f := configFileName
	if configPath == currentUser.HomeDir {
		f = "." + configFileName
	}

	return filepath.Join(configPath, f)
}

func ReadCredentialsList() (*CredentialsList, error) {
	configPath := resolveConfigFilePath()

	if _, err := os.Stat(configPath); err != nil {
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
		if err := os.MkdirAll(strings.TrimSuffix(resolveConfigFilePath(), configFileName), 0700); err != nil {
			return err
		}
		credentialsList = &CredentialsList{}
		credentials.Active = true
	}

	(*credentialsList)[credentials.Name] = &credentials
	return SaveCredentialsList(credentialsList)
}

func SaveCredentialsList(credentialsList *CredentialsList) error {
	configPath := resolveConfigFilePath()

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
	configFilePath := resolveConfigFilePath()

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

// will read the scret from file, if there is one
func readSecretFromFile(file string, secret *string) error {
	if file != "" {
		b, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}
		*secret = strings.TrimSpace(string(b))
	}

	return nil
}
