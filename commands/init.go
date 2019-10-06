package commands

import (
	"strings"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mmctl/client"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func withClient(fn func(c client.Client, cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		c, err := InitClient()
		if err != nil {
			return err
		}
		return fn(c, cmd, args)
	}
}

func InitClientWithUsernameAndPassword(username, password, instanceUrl string) (*model.Client4, error) {
	client := model.NewAPIv4Client(instanceUrl)
	_, response := client.Login(username, password)
	if response.Error != nil {
		return nil, response.Error
	}
	return client, nil
}

func InitClientWithMFA(username, password, mfaToken, instanceUrl string) (*model.Client4, error) {
	client := model.NewAPIv4Client(instanceUrl)
	_, response := client.LoginWithMFA(username, password, mfaToken)
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
	credentials, err := GetCurrentCredentials()
	if err != nil {
		return nil, err
	}
	return InitClientWithCredentials(credentials)
}

func InitWebSocketClient() (*model.WebSocketClient, error) {
	credentials, err := GetCurrentCredentials()
	if err != nil {
		return nil, err
	}
	client, appErr := model.NewWebSocketClient4(strings.Replace(credentials.InstanceUrl, "http", "ws", 1), credentials.AuthToken)
	if appErr != nil {
		return nil, errors.Wrap(appErr, "unable to create the websockets connection")
	}
	return client, nil
}
