// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/spf13/cobra"
)

func (s *MmctlE2ETestSuite) TestSamlAuthDataResetCmd() {
	s.SetupEnterpriseTestHelper().InitBasic()
	defer s.TearDownTest()

	// sanity check
	s.Require().NotNil(s.th.App.Saml())

	user := s.th.BasicUser
	resetAuthDataToID := func() {
		_, err := s.th.App.Srv().Store.User().UpdateAuthData(
			user.Id, model.USER_AUTH_SERVICE_SAML, model.NewString("some-id"), "", false)
		s.Require().Nil(err)
	}
	clearCache := func() {
		err := s.th.App.Srv().InvalidateAllCaches()
		s.Require().Nil(err)
	}
	checkAuthDataWasNotReset := func() {
		clearCache()
		retrievedUser, appErr := s.th.App.GetUser(user.Id)
		s.Require().Nil(appErr)
		s.Require().Equal("some-id", *retrievedUser.AuthData)
	}
	checkAuthDataWasReset := func() {
		clearCache()
		retrievedUser, appErr := s.th.App.GetUser(user.Id)
		s.Require().Nil(appErr)
		s.Require().Equal(user.Email, *retrievedUser.AuthData)
	}

	resetAuthDataToID()
	s.Run("dry run", func() {
		cmd := &cobra.Command{}
		cmd.Flags().Bool("dry-run", true, "")
		err := samlAuthDataResetCmdF(s.th.SystemAdminClient, cmd, nil)
		s.Require().Nil(err)

		checkAuthDataWasNotReset()
	})

	s.Run("real run", func() {
		cmd := &cobra.Command{}
		err := samlAuthDataResetCmdF(s.th.SystemAdminClient, cmd, nil)
		s.Require().Nil(err)

		checkAuthDataWasReset()
	})

	resetAuthDataToID()
	s.Run("with specific user IDs", func() {
		cmd := &cobra.Command{}
		cmd.Flags().StringSlice("users", []string{s.th.BasicUser2.Id}, "")
		err := samlAuthDataResetCmdF(s.th.SystemAdminClient, cmd, nil)
		s.Require().Nil(err)
		checkAuthDataWasNotReset()

		cmd = &cobra.Command{}
		cmd.Flags().StringSlice("users", []string{user.Id}, "")
		err = samlAuthDataResetCmdF(s.th.SystemAdminClient, cmd, nil)
		s.Require().Nil(err)
		checkAuthDataWasReset()
	})

	resetAuthDataToID()
	// delete user
	s.th.App.UpdateUserActive(user.Id, false)
	s.Run("without deleted users", func() {
		cmd := &cobra.Command{}
		err := samlAuthDataResetCmdF(s.th.SystemAdminClient, cmd, nil)
		s.Require().Nil(err)
		checkAuthDataWasNotReset()
	})
	s.Run("with deleted users", func() {
		cmd := &cobra.Command{}
		cmd.Flags().Bool("include-deleted", true, "")
		err := samlAuthDataResetCmdF(s.th.SystemAdminClient, cmd, nil)
		s.Require().Nil(err)
		checkAuthDataWasReset()
	})
}
