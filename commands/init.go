package commands

import (
	"github.com/mattermost/mattermost-server/model"
)

func InitClientWithUsernameAndPassword(username, password, instanceUrl string) (*model.Client4, error) {
	client := model.NewAPIv4Client(instanceUrl)
	_, response := client.Login(username, password)
	if response.Error != nil {
		return nil, response.Error
	}
	return client, nil
}

func InitClientWithCredentials(credentials *Credentials) (*model.Client4, error) {
	client := model.NewAPIv4Client(credentials.InstanceUrl)

	client.AuthType = model.HEADER_BEARER
	client.AuthToken = credentials.AuthToken

	if _, response := client.GetMe(""); response.Error != nil {
		return nil, response.Error
	}
	return client, nil
}

func InitClient() (*model.Client4, error) {
	credentials, err := ReadCredentials()
	if err != nil {
		return nil, err
	}
	return InitClientWithCredentials(credentials)
}
