package commands

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"os/user"
)

type Credentials struct {
	Username    string `json:"username"`
	AuthToken   string `json:"authToken"`
	InstanceUrl string `json:"instanceUrl"`
}

func getConfigFilePath() (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", err
	}
	return user.HomeDir + "/.mmctl", nil
}

func ReadCredentials() (*Credentials, error) {
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(configFilePath); err != nil {
		return nil, errors.New("Cannot read user credentials, maybe you need to use login first. Error: " + err.Error())
	}

	fileContents, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, errors.New("There was a problem reading the credentials file. Error: " + err.Error())
	}

	var credentials Credentials
	if err := json.Unmarshal(fileContents, &credentials); err != nil {
		return nil, errors.New("There was a problem parsing the credentials file. Error: " + err.Error())
	}

	return &credentials, nil
}

func SaveCredentials(credentials Credentials) error {
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	marshaledCredentials, _ := json.Marshal(credentials)

	if err := ioutil.WriteFile(configFilePath, marshaledCredentials, 0600); err != nil {
		return errors.New("Cannot save the credentials. Error: " + err.Error())
	}

	return nil
}

func CleanCredentials() error {
	configFilePath, err := getConfigFilePath()
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
