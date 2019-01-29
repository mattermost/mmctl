package commands

import (
	"github.com/mattermost/mattermost-server/model"
)

func InitClientWithCredentials(credentials *Credentials) (*model.Client4, error) {
	client := model.NewAPIv4Client(credentials.InstanceUrl)
	_, response := client.Login(credentials.Username, credentials.Password)
	if response.Error != nil {
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
