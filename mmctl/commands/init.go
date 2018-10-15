package commands

import (
	"github.com/mattermost/mattermost-server/model"
)

func InitClient() (*model.Client4, error) {
	client := model.NewAPIv4Client("http://localhost:8065")
	_, response := client.Login("sysadmin", "sysadmin")
	if response.Error != nil {
		return nil, response.Error
	}
	return client, nil
}
