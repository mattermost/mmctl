// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mmctl/v6/client"
	"github.com/mattermost/mmctl/v6/mocks"
	"github.com/mattermost/mmctl/v6/printer"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/mattermost/mattermost-server/v6/api4"
)

var EnableEnterpriseTests string

type MmctlUnitTestSuite struct {
	suite.Suite
	mockCtrl *gomock.Controller
	client   *mocks.MockClient
}

func (s *MmctlUnitTestSuite) SetupTest() {
	printer.Clean()
	printer.SetFormat(printer.FormatJSON)

	s.mockCtrl = gomock.NewController(s.T())
	s.client = mocks.NewMockClient(s.mockCtrl)
}

func (s *MmctlUnitTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

// LoginAs login user using the given username
func (s *MmctlUnitTestSuite) LoginAs(username string) func() {
	return loginAs(s.T(), username)
}

type MmctlE2ETestSuite struct {
	suite.Suite
	th *api4.TestHelper
}

func (s *MmctlE2ETestSuite) SetupTest() {
	printer.Clean()
	printer.SetFormat(printer.FormatJSON)
}

func (s *MmctlE2ETestSuite) TearDownTest() {
	// if a test helper was used, we run the teardown and remove it
	// from the structure to avoid reusing the same helper between
	// tests
	if s.th != nil {
		s.th.TearDown()
		s.th = nil
	}
}

func (s *MmctlE2ETestSuite) SetupTestHelper() *api4.TestHelper {
	s.th = api4.Setup(s.T())
	return s.th
}

func (s *MmctlE2ETestSuite) SetupEnterpriseTestHelper() *api4.TestHelper {
	if EnableEnterpriseTests != "true" {
		s.T().SkipNow()
	}
	s.th = api4.SetupEnterprise(s.T())
	return s.th
}

// RunForSystemAdminAndLocal runs a test function for both SystemAdmin
// and Local clients. Several commands work in the same way when used
// by a fully privileged user and through the local mode, so this
// helper facilitates checking both
func (s *MmctlE2ETestSuite) RunForSystemAdminAndLocal(testName string, fn func(client.Client)) {
	s.Run(testName+"/SystemAdminClient", func() {
		fn(s.th.SystemAdminClient)
	})

	s.Run(testName+"/LocalClient", func() {
		fn(s.th.LocalClient)
	})
}

// RunForAllClients runs a test function for all the clients
// registered in the TestHelper
func (s *MmctlE2ETestSuite) RunForAllClients(testName string, fn func(client.Client)) {
	s.Run(testName+"/Client", func() {
		fn(s.th.Client)
	})

	s.Run(testName+"/SystemAdminClient", func() {
		fn(s.th.SystemAdminClient)
	})

	s.Run(testName+"/LocalClient", func() {
		fn(s.th.LocalClient)
	})
}

// LoginAs login user using the given username
func (s *MmctlE2ETestSuite) LoginAs(username string) func() {
	return loginAs(s.T(), username)
}

func loginAs(t *testing.T, username string) func() {
	err := os.Setenv("XDG_CONFIG_HOME", "path/should/be/ignored")
	require.NoError(t, err)

	tmp, _ := os.MkdirTemp("", "mmctl-")
	path := filepath.Join(tmp, configFileName)
	viper.Set("config", path)

	err = SaveCredentials(Credentials{Username: username})
	require.NoError(t, err)

	return func() { os.RemoveAll(tmp) }
}
