// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func CheckVersionMatch(version, serverVersion string) bool {
	maj, min, _ := model.SplitVersion(version)
	srvMaj, srvMin, _ := model.SplitVersion(serverVersion)

	return maj == srvMaj && min == srvMin
}

func withClient(fn func(c client.Client, cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		c, serverVersion, err := InitClient()
		if err != nil {
			return err
		}

		valid := CheckVersionMatch(Version, serverVersion)
		if !valid {
			if viper.GetBool("strict") {
				return fmt.Errorf("Server version %s doesn't match with mmctl version %s. Please update mmctl or use --skip-version-check to ignore", serverVersion, Version)
			} else {
				printer.PrintError("WARNING: server version " + serverVersion + " doesn't match mmctl version " + Version)
			}
		}

		return fn(c, cmd, args)
	}
}

func InitClientWithUsernameAndPassword(username, password, instanceUrl string) (*model.Client4, string, error) {
	client := model.NewAPIv4Client(instanceUrl)
	_, response := client.Login(username, password)
	if response.Error != nil {
		return nil, "", response.Error
	}
	return client, response.ServerVersion, nil
}

func InitClientWithMFA(username, password, mfaToken, instanceUrl string) (*model.Client4, string, error) {
	client := model.NewAPIv4Client(instanceUrl)
	_, response := client.LoginWithMFA(username, password, mfaToken)
	if response.Error != nil {
		return nil, "", response.Error
	}
	return client, response.ServerVersion, nil
}

func InitClientWithCredentials(credentials *Credentials) (*model.Client4, string, error) {
	client := model.NewAPIv4Client(credentials.InstanceUrl)

	client.AuthType = model.HEADER_BEARER
	client.AuthToken = credentials.AuthToken

	_, response := client.GetMe("")
	if response.Error != nil {
		return nil, "", response.Error
	}

	return client, response.ServerVersion, nil
}

func InitClient() (*model.Client4, string, error) {
	credentials, err := GetCurrentCredentials()
	if err != nil {
		return nil, "", err
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
