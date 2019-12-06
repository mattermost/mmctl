package commands

import (
	"fmt"
	"os"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mmctl/client"
)

const (
	INSTANCE_URL      = "http://localhost:8065"
	SYSADMIN_USERNAME = "sysadmin"
	SYSADMIN_PASS     = "Sys@dmin-sample1"
	USER_USERNAME     = "user-1"
	USER_PASS         = "SampleUs@r-1"
)

type TestHelper struct {
	Client            client.Client
	SystemAdminClient client.Client
	BasicUser         *model.User
	SystemAdminUser   *model.User
}

func setupTestHelper() (*TestHelper, error) {
	instanceUrl := INSTANCE_URL
	if os.Getenv("MMCTL_INSTANCE_URL") != "" {
		instanceUrl = os.Getenv("MMCTL_INSTANCE_URL")
	}

	sysadminClient, err := InitClientWithUsernameAndPassword(SYSADMIN_USERNAME, SYSADMIN_PASS, instanceUrl)
	if err != nil {
		return nil, fmt.Errorf("SystemAdminClient client failed to connect: %s", err)
	}
	sysadminUser, response := sysadminClient.GetUserByUsername(SYSADMIN_USERNAME, "")
	if response.Error != nil {
		return nil, fmt.Errorf("Couldn't retrieve system admin user with username %s: %s", SYSADMIN_USERNAME, response.Error)
	}

	client, err := InitClientWithUsernameAndPassword(USER_USERNAME, USER_PASS, instanceUrl)
	if err != nil {
		return nil, fmt.Errorf("Basic client failed to connect: %s", err)
	}
	basicUser, response := client.GetUserByUsername(USER_USERNAME, "")
	if response.Error != nil {
		return nil, fmt.Errorf("Couldn't retrieve basic user with username %s: %s", USER_USERNAME, response.Error)
	}

	th := &TestHelper{
		Client:            client,
		SystemAdminClient: sysadminClient,
		BasicUser:         basicUser,
		SystemAdminUser:   sysadminUser,
	}

	return th, nil
}
