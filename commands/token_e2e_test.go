// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/spf13/cobra"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"
)

func (s *MmctlE2ETestSuite) TestTokenGenerateForUserCmd() {
	s.SetupTestHelper().InitBasic()

	tokenDescription := model.NewRandomString(10)

	previousVal := s.th.App.Config().ServiceSettings.EnableUserAccessTokens
	s.th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableUserAccessTokens = true })
	defer s.th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableUserAccessTokens = *previousVal })

	s.Run("Generate token via Local Admin", func() {
		printer.Clean()

		err := generateTokenForAUserCmdF(s.th.LocalClient, &cobra.Command{}, []string{s.th.BasicUser.Id, tokenDescription})
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Contains(
			err.Error(),
			fmt.Sprintf(`There doesn't appear to be an api call for the url='/api/v4/users/%v/tokens`, s.th.BasicUser.Id))
	})

	s.Run("Generate token for admin with admin", func() {
		printer.Clean()

		err := generateTokenForAUserCmdF(s.th.SystemAdminClient, &cobra.Command{}, []string{s.th.SystemAdminUser.Email, tokenDescription})
		s.Require().Nil(err)
		s.Require().Len(printer.GetErrorLines(), 0)

		userTokens, appErr := s.th.App.GetUserAccessTokensForUser(s.th.SystemAdminUser.Id, 0, 1)
		s.Require().Nil(appErr)
		s.Require().Equal(1, len(userTokens))

		userToken, appErr := s.th.App.GetUserAccessToken(userTokens[0].Id, false)
		s.Require().Nil(appErr)

		expectedUserToken := printer.GetLines()[0].(*model.UserAccessToken)

		s.Require().Equal(expectedUserToken, userToken)
	})

	s.Run("Generate token for user with admin", func() {
		printer.Clean()

		err := generateTokenForAUserCmdF(s.th.SystemAdminClient, &cobra.Command{}, []string{s.th.BasicUser.Email, tokenDescription})
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Contains(
			err.Error(),
			`You cannot create an access token for another user.`)
	})

	s.Run("Generate token for user with user", func() {
		printer.Clean()

		_, response := s.th.SystemAdminClient.UpdateUserRoles(s.th.BasicUser.Id, model.SYSTEM_USER_ROLE_ID+" "+model.SYSTEM_USER_ACCESS_TOKEN_ROLE_ID)
		s.Require().NotNil(response)
		s.Require().Nil(response.Error)
		defer s.th.SystemAdminClient.UpdateUserRoles(s.th.BasicUser.Id, model.SYSTEM_USER_ROLE_ID)

		err := generateTokenForAUserCmdF(s.th.Client, &cobra.Command{}, []string{s.th.BasicUser.Email, tokenDescription})
		s.Require().Nil(err)
		s.Require().Len(printer.GetErrorLines(), 0)

		userTokens, appErr := s.th.App.GetUserAccessTokensForUser(s.th.BasicUser.Id, 0, 1)
		s.Require().Nil(appErr)
		s.Require().Equal(1, len(userTokens))

		userToken, appErr := s.th.App.GetUserAccessToken(userTokens[0].Id, false)
		s.Require().Nil(appErr)

		expectedUserToken := printer.GetLines()[0].(*model.UserAccessToken)

		s.Require().Equal(expectedUserToken, userToken)
	})

	s.RunForSystemAdminAndLocal("Generate token for nonexistent user", func(c client.Client) {
		printer.Clean()

		nonExistentUserEmail := s.th.GenerateTestEmail()

		err := generateTokenForAUserCmdF(c, &cobra.Command{}, []string{nonExistentUserEmail, tokenDescription})
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Equal(
			fmt.Sprintf(`could not retrieve user information of %q`, nonExistentUserEmail),
			err.Error())
	})

	s.Run("Generate token without permission", func() {
		printer.Clean()

		user, appErr := s.th.App.CreateUser(&model.User{Email: s.th.GenerateTestEmail(), Username: model.NewId(), Password: model.NewId()})
		s.Require().Nil(appErr)

		err := generateTokenForAUserCmdF(s.th.Client, &cobra.Command{}, []string{user.Email, tokenDescription})
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Equal(
			fmt.Sprintf(`could not create token for %q: : You do not have the appropriate permissions., `, user.Email),
			err.Error())

		userTokens, appErr := s.th.App.GetUserAccessTokensForUser(user.Id, 0, 1)
		s.Require().Nil(appErr)
		s.Require().Equal(0, len(userTokens))
	})
}
