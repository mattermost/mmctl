// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"fmt"
	"os"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mmctl/client"
)

const (
	InstanceURL      = "http://localhost:8065"
	SysadminUsername = "sysadmin"
	SysadminPass     = "Sys@dmin-sample1"
	UserUsername     = "user-1"
	UserPass         = "SampleUs@r-1"
)

type TestHelper struct {
	Client            client.Client
	SystemAdminClient client.Client
	BasicUser         *model.User
	SystemAdminUser   *model.User
}

func setupTestHelper() (*TestHelper, error) {
	instanceURL := InstanceURL
	if os.Getenv("MMCTL_INSTANCE_URL") != "" {
		instanceURL = os.Getenv("MMCTL_INSTANCE_URL")
	}

	sysadminClient, err := InitClientWithUsernameAndPassword(SysadminUsername, SysadminPass, instanceURL)
	if err != nil {
		return nil, fmt.Errorf("system admin client failed to connect: %s", err)
	}
	sysadminUser, response := sysadminClient.GetUserByUsername(SysadminUsername, "")
	if response.Error != nil {
		return nil, fmt.Errorf("couldn't retrieve system admin user with username %s: %s", SysadminUsername, response.Error)
	}

	client, err := InitClientWithUsernameAndPassword(UserUsername, UserPass, instanceURL)
	if err != nil {
		return nil, fmt.Errorf("basic client failed to connect: %s", err)
	}
	basicUser, response := client.GetUserByUsername(UserUsername, "")
	if response.Error != nil {
		return nil, fmt.Errorf("couldn't retrieve basic user with username %s: %s", UserUsername, response.Error)
	}

	th := &TestHelper{
		Client:            client,
		SystemAdminClient: sysadminClient,
		BasicUser:         basicUser,
		SystemAdminUser:   sysadminUser,
	}

	return th, nil
}
