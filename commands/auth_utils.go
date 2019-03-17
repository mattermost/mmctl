package commands

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"os/user"
)

type Credentials struct {
	Name        string `json:"name"`
	Username    string `json:"username"`
	AuthToken   string `json:"authToken"`
	InstanceUrl string `json:"instanceUrl"`
	Active      bool   `json:"active"`
}

type CredentialsList map[string]*Credentials

func getConfigFilePath() (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", err
	}
	return user.HomeDir + "/.mmctl", nil
}

func ReadCredentialsList() (*CredentialsList, error) {
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

	var credentialsList CredentialsList
	if err := json.Unmarshal(fileContents, &credentialsList); err != nil {
		return nil, errors.New("There was a problem parsing the credentials file. Error: " + err.Error())
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
	return nil, errors.New("No current context available. Please use the \"auth set\" command.")
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
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	marshaledCredentialsList, _ := json.Marshal(credentialsList)

	if err := ioutil.WriteFile(configFilePath, marshaledCredentialsList, 0600); err != nil {
		return errors.New("Cannot save the credentials. Error: " + err.Error())
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
		return errors.New("Cannot find credentials for server " + name)
	}

	return SaveCredentialsList(credentialsList)
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
