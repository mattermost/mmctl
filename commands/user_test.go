// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mmctl/printer"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (s *MmctlUnitTestSuite) TestUserActivateCmd() {
	s.Run("Activate user", func() {
		printer.Clean()
		emailArg := "example@example.com"
		mockUser := model.User{Id: "example", Username: "ExampleUser", Email: emailArg}

		s.client.
			EXPECT().
			GetUserByEmail(emailArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserActive(mockUser.Id, true).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := userActivateCmdF(s.client, &cobra.Command{}, []string{emailArg})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Try to activate unexistent user", func() {
		printer.Clean()
		emailArg := "example@example.com"

		s.client.
			EXPECT().
			GetUserByEmail(emailArg, "").
			Return(nil, &model.Response{Error: &model.AppError{Id: "Mock Error"}}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(emailArg, "").
			Return(nil, &model.Response{Error: &model.AppError{Id: "Mock Error"}}).
			Times(1)

		s.client.
			EXPECT().
			GetUser(emailArg, "").
			Return(nil, &model.Response{Error: &model.AppError{Id: "Mock Error"}}).
			Times(1)

		err := userActivateCmdF(s.client, &cobra.Command{}, []string{emailArg})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(fmt.Errorf("can't find user '%v'", emailArg).Error(), printer.GetErrorLines()[0])
	})

	s.Run("Fail to activate user", func() {
		printer.Clean()
		emailArg := "example@example.com"
		mockUser := model.User{Id: "example", Username: "ExampleUser", Email: emailArg}

		s.client.
			EXPECT().
			GetUserByEmail(emailArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserActive(mockUser.Id, true).
			Return(false, &model.Response{Error: &model.AppError{Id: "Mock Error"}}).
			Times(1)

		err := userActivateCmdF(s.client, &cobra.Command{}, []string{emailArg})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(fmt.Errorf("unable to change activation status of user: %v", emailArg).Error(), printer.GetErrorLines()[0])
	})

	s.Run("Activate several users with unexistent ones and failed ones", func() {
		printer.Clean()
		emailArgs := []string{"example0@example0.com", "null", "example2@example2.com", "failure@failure.com", "example4@example4.com"}
		mockUser0 := model.User{Id: "example0", Username: "ExampleUser0", Email: emailArgs[0]}
		mockUser2 := model.User{Id: "example2", AuthService: "other", Username: "ExampleUser2", Email: emailArgs[2]}
		mockUser3 := model.User{Id: "failure", Username: "FailureUser", Email: emailArgs[3]}
		mockUser4 := model.User{Id: "example4", Username: "ExampleUser4", Email: emailArgs[4]}

		s.client.
			EXPECT().
			GetUserByEmail(emailArgs[0], "").
			Return(&mockUser0, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(emailArgs[1], "").
			Return(nil, &model.Response{Error: &model.AppError{Id: "Mock Error"}}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(emailArgs[1], "").
			Return(nil, &model.Response{Error: &model.AppError{Id: "Mock Error"}}).
			Times(1)

		s.client.
			EXPECT().
			GetUser(emailArgs[1], "").
			Return(nil, &model.Response{Error: &model.AppError{Id: "Mock Error"}}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(emailArgs[2], "").
			Return(&mockUser2, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(emailArgs[3], "").
			Return(&mockUser3, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(emailArgs[4], "").
			Return(&mockUser4, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserActive(mockUser0.Id, true).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserActive(mockUser2.Id, true).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserActive(mockUser3.Id, true).
			Return(false, &model.Response{Error: &model.AppError{Id: "Mock Error"}}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserActive(mockUser4.Id, true).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := userActivateCmdF(s.client, &cobra.Command{}, emailArgs)
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 2)
		s.Require().Equal(fmt.Errorf("can't find user '%v'", emailArgs[1]).Error(), printer.GetErrorLines()[0])
		s.Require().Equal(fmt.Errorf("unable to change activation status of user: %v", emailArgs[3]).Error(), printer.GetErrorLines()[1])
	})
}

func (s *MmctlUnitTestSuite) TestDeactivateUserCmd() {
	s.Run("Deactivate user", func() {
		printer.Clean()
		emailArg := "example@example.com"
		mockUser := model.User{Id: "example", Username: "ExampleUser", Email: emailArg}

		s.client.
			EXPECT().
			GetUserByEmail(emailArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserActive(mockUser.Id, false).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := userDeactivateCmdF(s.client, &cobra.Command{}, []string{emailArg})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Try to deactivate unexistent user", func() {
		printer.Clean()
		emailArg := "example@example.com"

		s.client.
			EXPECT().
			GetUserByEmail(emailArg, "").
			Return(nil, &model.Response{Error: &model.AppError{Id: "Mock Error"}}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(emailArg, "").
			Return(nil, &model.Response{Error: &model.AppError{Id: "Mock Error"}}).
			Times(1)

		s.client.
			EXPECT().
			GetUser(emailArg, "").
			Return(nil, &model.Response{Error: &model.AppError{Id: "Mock Error"}}).
			Times(1)

		err := userDeactivateCmdF(s.client, &cobra.Command{}, []string{emailArg})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(fmt.Errorf("can't find user '%v'", emailArg).Error(), printer.GetErrorLines()[0])
	})

	s.Run("Fail to deactivate user", func() {
		printer.Clean()
		emailArg := "example@example.com"
		mockUser := model.User{Id: "example", Username: "ExampleUser", Email: emailArg}

		s.client.
			EXPECT().
			GetUserByEmail(emailArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserActive(mockUser.Id, false).
			Return(false, &model.Response{Error: &model.AppError{Id: "Mock Error"}}).
			Times(1)

		err := userDeactivateCmdF(s.client, &cobra.Command{}, []string{emailArg})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(fmt.Errorf("unable to change activation status of user: %v", emailArg).Error(), printer.GetErrorLines()[0])
	})

	s.Run("Deactivate SSO user", func() {
		printer.Clean()
		emailArg := "example@example.com"
		mockUser := model.User{Id: "example", AuthService: "other", Username: "ExampleUser", Email: emailArg}

		s.client.
			EXPECT().
			GetUserByEmail(emailArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserActive(mockUser.Id, false).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := userDeactivateCmdF(s.client, &cobra.Command{}, []string{emailArg})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal("You must also deactivate user "+emailArg+" in the SSO provider or they will be reactivated on next login or sync.", printer.GetLines()[0])
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Deactivate several users with unexistent ones, SSO ones and failed ones", func() {
		printer.Clean()
		emailArgs := []string{"example0@example0.com", "null", "example2@example2.com", "failure@failure.com", "example4@example4.com"}
		mockUser0 := model.User{Id: "example0", Username: "ExampleUser0", Email: emailArgs[0]}
		mockUser2 := model.User{Id: "example2", AuthService: "other", Username: "ExampleUser2", Email: emailArgs[2]}
		mockUser3 := model.User{Id: "failure", Username: "FailureUser", Email: emailArgs[3]}
		mockUser4 := model.User{Id: "example4", Username: "ExampleUser4", Email: emailArgs[4]}

		s.client.
			EXPECT().
			GetUserByEmail(emailArgs[0], "").
			Return(&mockUser0, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(emailArgs[1], "").
			Return(nil, &model.Response{Error: &model.AppError{Id: "Mock Error"}}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(emailArgs[1], "").
			Return(nil, &model.Response{Error: &model.AppError{Id: "Mock Error"}}).
			Times(1)

		s.client.
			EXPECT().
			GetUser(emailArgs[1], "").
			Return(nil, &model.Response{Error: &model.AppError{Id: "Mock Error"}}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(emailArgs[2], "").
			Return(&mockUser2, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(emailArgs[3], "").
			Return(&mockUser3, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(emailArgs[4], "").
			Return(&mockUser4, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserActive(mockUser0.Id, false).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserActive(mockUser2.Id, false).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserActive(mockUser3.Id, false).
			Return(false, &model.Response{Error: &model.AppError{Id: "Mock Error"}}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserActive(mockUser4.Id, false).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := userDeactivateCmdF(s.client, &cobra.Command{}, emailArgs)
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal("You must also deactivate user "+emailArgs[2]+" in the SSO provider or they will be reactivated on next login or sync.", printer.GetLines()[0])
		s.Require().Len(printer.GetErrorLines(), 2)
		s.Require().Equal(fmt.Errorf("can't find user '%v'", emailArgs[1]).Error(), printer.GetErrorLines()[0])
		s.Require().Equal(fmt.Errorf("unable to change activation status of user: %v", emailArgs[3]).Error(), printer.GetErrorLines()[1])
	})
}

func (s *MmctlUnitTestSuite) TestDeleteAllUsersCmd() {
	s.Run("Delete all users", func() {
		printer.Clean()
		cmd := &cobra.Command{}
		cmd.Flags().Bool("confirm", true, "")

		s.client.
			EXPECT().
			PermanentDeleteAllUsers().
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := deleteAllUsersCmdF(s.client, cmd, []string{})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Equal(printer.GetLines()[0], "All users successfully deleted")
	})

	s.Run("Delete all users call fails", func() {
		printer.Clean()
		cmd := &cobra.Command{}
		cmd.Flags().Bool("confirm", true, "")

		s.client.
			EXPECT().
			PermanentDeleteAllUsers().
			Return(false, &model.Response{Error: &model.AppError{Id: "Mock Error"}}).
			Times(1)

		err := deleteAllUsersCmdF(s.client, cmd, []string{})
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})
}

func (s *MmctlUnitTestSuite) TestSearchUserCmd() {
	s.Run("Search for an existing user", func() {
		emailArg := "example@example.com"
		mockUser := model.User{Username: "ExampleUser", Email: emailArg}

		s.client.
			EXPECT().
			GetUserByEmail(emailArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		err := searchUserCmdF(s.client, &cobra.Command{}, []string{emailArg})
		s.Require().Nil(err)
		s.Require().Equal(&mockUser, printer.GetLines()[0])
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Search for a nonexistent user", func() {
		printer.Clean()
		arg := "example@example.com"

		s.client.
			EXPECT().
			GetUserByEmail(arg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(arg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUser(arg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := searchUserCmdF(s.client, &cobra.Command{}, []string{arg})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Equal("Unable to find user 'example@example.com'", printer.GetErrorLines()[0])
	})

	s.Run("Avoid path traversal", func() {
		printer.Clean()
		arg := "test/../hello?@mattermost.com"

		err := searchUserCmdF(s.client, &cobra.Command{}, []string{arg})
		s.Require().Nil(err)
		s.Require().Equal("Unable to find user 'test/../hello?@mattermost.com'", printer.GetErrorLines()[0])
	})
}

func (s *MmctlUnitTestSuite) TestSendPasswordResetEmailCmd() {
	s.Run("Send one reset email", func() {
		printer.Clean()
		emailArg := "example@example.com"

		s.client.
			EXPECT().
			SendPasswordResetEmail(emailArg).
			Return(false, &model.Response{Error: nil}).
			Times(1)

		err := sendPasswordResetEmailCmdF(s.client, &cobra.Command{}, []string{emailArg})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Send one reset email and receive error", func() {
		printer.Clean()
		emailArg := "example@example.com"
		mockError := model.AppError{Id: "Mock Error"}

		s.client.
			EXPECT().
			SendPasswordResetEmail(emailArg).
			Return(false, &model.Response{Error: &mockError}).
			Times(1)

		err := sendPasswordResetEmailCmdF(s.client, &cobra.Command{}, []string{emailArg})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal("Unable send reset password email to email "+emailArg+". Error: "+mockError.Error(), printer.GetErrorLines()[0])
	})

	s.Run("Send several reset emails and receive some errors", func() {
		printer.Clean()
		emailArg := []string{
			"example1@example.com",
			"error1@example.com",
			"error2@example.com",
			"example2@example.com",
			"example3@example.com"}
		mockError := model.AppError{Id: "Mock Error"}

		for _, email := range emailArg {
			if strings.HasPrefix(email, "error") {
				s.client.
					EXPECT().
					SendPasswordResetEmail(email).
					Return(false, &model.Response{Error: &mockError}).
					Times(1)
			} else {
				s.client.
					EXPECT().
					SendPasswordResetEmail(email).
					Return(false, &model.Response{Error: nil}).
					Times(1)
			}
		}

		err := sendPasswordResetEmailCmdF(s.client, &cobra.Command{}, emailArg)
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 2)
		s.Require().Equal("Unable send reset password email to email "+emailArg[1]+". Error: "+mockError.Error(), printer.GetErrorLines()[0])
		s.Require().Equal("Unable send reset password email to email "+emailArg[2]+". Error: "+mockError.Error(), printer.GetErrorLines()[1])
	})
}

func (s *MmctlUnitTestSuite) TestUserInviteCmd() {
	s.Run("Invite user to an existing team by Id", func() {
		printer.Clean()
		argUser := "example@example.com"
		argTeam := "teamId"

		s.client.
			EXPECT().
			GetTeam(argTeam, "").
			Return(&model.Team{Id: argTeam}, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			InviteUsersToTeam(argTeam, []string{argUser}).
			Return(false, &model.Response{Error: nil}).
			Times(1)

		err := userInviteCmdF(s.client, &cobra.Command{}, []string{argUser, argTeam})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal("Invites may or may not have been sent.", printer.GetLines()[0])
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Invite user to an existing team by name", func() {
		printer.Clean()
		argUser := "example@example.com"
		argTeam := "teamName"
		resultID := "teamId"

		s.client.
			EXPECT().
			GetTeam(argTeam, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeamByName(argTeam, "").
			Return(&model.Team{Id: resultID}, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			InviteUsersToTeam(resultID, []string{argUser}).
			Return(false, &model.Response{Error: nil}).
			Times(1)

		err := userInviteCmdF(s.client, &cobra.Command{}, []string{argUser, argTeam})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal("Invites may or may not have been sent.", printer.GetLines()[0])
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Invite user to several existing teams by name and id", func() {
		printer.Clean()
		argUser := "example@example.com"
		argTeam := []string{"teamName1", "teamId2", "teamId3", "teamName4"}
		resultTeamModels := [4]*model.Team{
			{Id: "teamId1"},
			{Id: "teamId2"},
			{Id: "teamId3"},
			{Id: "teamId4"},
		}

		// Setup GetTeam
		s.client.
			EXPECT().
			GetTeam(argTeam[0], "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeam(argTeam[1], "").
			Return(resultTeamModels[1], &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeam(argTeam[2], "").
			Return(resultTeamModels[2], &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeam(argTeam[3], "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		// Setup GetTeamByName
		s.client.
			EXPECT().
			GetTeamByName(argTeam[0], "").
			Return(resultTeamModels[0], &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeamByName(argTeam[3], "").
			Return(resultTeamModels[3], &model.Response{Error: nil}).
			Times(1)

		// Setup InviteUsersToTeam
		for _, resultTeamModel := range resultTeamModels {
			s.client.
				EXPECT().
				InviteUsersToTeam(resultTeamModel.Id, []string{argUser}).
				Return(false, &model.Response{Error: nil}).
				Times(1)
		}

		err := userInviteCmdF(s.client, &cobra.Command{}, append([]string{argUser}, argTeam...))
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), len(argTeam))
		for i := 0; i < len(argTeam); i++ {
			s.Require().Equal("Invites may or may not have been sent.", printer.GetLines()[i])
		}
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Invite user to an un-existing team", func() {
		printer.Clean()
		argUser := "example@example.com"
		argTeam := "unexistent"

		s.client.
			EXPECT().
			GetTeam(argTeam, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeamByName(argTeam, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := userInviteCmdF(s.client, &cobra.Command{}, []string{argUser, argTeam})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal("can't find team '"+argTeam+"'", printer.GetErrorLines()[0])
	})

	s.Run("Invite user to an existing team and fail invite", func() {
		printer.Clean()
		argUser := "example@example.com"
		argTeam := "teamId"
		resultName := "teamName"
		mockError := model.NewAppError("", "Mock Error", nil, "", 0)

		s.client.
			EXPECT().
			GetTeam(argTeam, "").
			Return(&model.Team{Id: argTeam, Name: resultName}, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			InviteUsersToTeam(argTeam, []string{argUser}).
			Return(false, &model.Response{Error: mockError}).
			Times(1)

		err := userInviteCmdF(s.client, &cobra.Command{}, []string{argUser, argTeam})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal("Unable to invite user with email "+argUser+" to team "+resultName+". Error: "+mockError.Error(), printer.GetErrorLines()[0])
	})

	s.Run("Invite user to several existing and non-existing teams by name and id and reject one invite", func() {
		printer.Clean()
		argUser := "example@example.com"
		argTeam := []string{"teamName1", "unexistent", "teamId3", "teamName4", "reject", "teamId6"}
		resultTeamModels := [6]*model.Team{
			{Id: "teamId1", Name: "teamName1"},
			nil,
			{Id: "teamId3", Name: "teamName3"},
			{Id: "teamId4", Name: "teamName4"},
			{Id: "reject", Name: "rejectName"},
			{Id: "teamId6", Name: "teamName6"},
		}
		mockError := model.NewAppError("", "Mock Error", nil, "", 0)

		// Setup GetTeam
		s.client.
			EXPECT().
			GetTeam(argTeam[0], "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeam(argTeam[1], "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeam(argTeam[2], "").
			Return(resultTeamModels[2], &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeam(argTeam[3], "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeam(argTeam[4], "").
			Return(resultTeamModels[4], &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeam(argTeam[5], "").
			Return(resultTeamModels[5], &model.Response{Error: nil}).
			Times(1)

		// Setup GetTeamByName
		s.client.
			EXPECT().
			GetTeamByName(argTeam[0], "").
			Return(resultTeamModels[0], &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeamByName(argTeam[1], "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeamByName(argTeam[3], "").
			Return(resultTeamModels[3], &model.Response{Error: nil}).
			Times(1)

		// Setup InviteUsersToTeam
		s.client.
			EXPECT().
			InviteUsersToTeam(resultTeamModels[0].Id, []string{argUser}).
			Return(false, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			InviteUsersToTeam(resultTeamModels[2].Id, []string{argUser}).
			Return(false, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			InviteUsersToTeam(resultTeamModels[3].Id, []string{argUser}).
			Return(false, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			InviteUsersToTeam(resultTeamModels[4].Id, []string{argUser}).
			Return(false, &model.Response{Error: mockError}).
			Times(1)

		s.client.
			EXPECT().
			InviteUsersToTeam(resultTeamModels[5].Id, []string{argUser}).
			Return(false, &model.Response{Error: nil}).
			Times(1)

		err := userInviteCmdF(s.client, &cobra.Command{}, append([]string{argUser}, argTeam...))
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 4)
		for i := 0; i < 4; i++ {
			s.Require().Equal("Invites may or may not have been sent.", printer.GetLines()[i])
		}
		s.Require().Len(printer.GetErrorLines(), 2)
		s.Require().Equal("can't find team '"+argTeam[1]+"'", printer.GetErrorLines()[0])
		s.Require().Equal("Unable to invite user with email "+argUser+" to team "+resultTeamModels[4].Name+". Error: "+mockError.Error(), printer.GetErrorLines()[1])
	})
}

func (s *MmctlUnitTestSuite) TestUserCreateCmd() {
	mockUser := model.User{
		Username: "username",
		Password: "password",
		Email:    "email",
	}

	s.Run("Create user with email missing", func() {
		printer.Clean()

		command := cobra.Command{}
		command.Flags().String("username", mockUser.Username, "")
		command.Flags().String("password", mockUser.Password, "")

		error := userCreateCmdF(s.client, &command, []string{})

		s.Require().Equal("Email is required: flag accessed but not defined: email", error.Error())
	})

	s.Run("Create user with username missing", func() {
		printer.Clean()

		command := cobra.Command{}
		command.Flags().String("email", mockUser.Email, "")
		command.Flags().String("password", mockUser.Password, "")

		error := userCreateCmdF(s.client, &command, []string{})

		s.Require().Equal("Username is required: flag accessed but not defined: username", error.Error())
	})

	s.Run("Create user with password missing", func() {
		printer.Clean()

		command := cobra.Command{}
		command.Flags().String("username", mockUser.Username, "")
		command.Flags().String("email", mockUser.Email, "")

		error := userCreateCmdF(s.client, &command, []string{})

		s.Require().Equal("Password is required: flag accessed but not defined: password", error.Error())
	})

	s.Run("Create a regular user", func() {
		printer.Clean()

		s.client.
			EXPECT().
			CreateUser(&mockUser).
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		command := cobra.Command{}
		command.Flags().String("username", mockUser.Username, "")
		command.Flags().String("email", mockUser.Email, "")
		command.Flags().String("password", mockUser.Password, "")

		error := userCreateCmdF(s.client, &command, []string{})

		s.Require().Nil(error)
		s.Require().Equal(&mockUser, printer.GetLines()[0])
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Create a regular user with client returning error", func() {
		printer.Clean()

		s.client.
			EXPECT().
			CreateUser(&mockUser).
			Return(&mockUser, &model.Response{Error: &model.AppError{Message: "Remote error"}}).
			Times(1)

		command := cobra.Command{}
		command.Flags().String("username", mockUser.Username, "")
		command.Flags().String("email", mockUser.Email, "")
		command.Flags().String("password", mockUser.Password, "")

		error := userCreateCmdF(s.client, &command, []string{})

		s.Require().Equal("Unable to create user. Error: : Remote error, ", error.Error())
	})

	s.Run("Create a sysAdmin user", func() {
		printer.Clean()

		s.client.
			EXPECT().
			CreateUser(&mockUser).
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserRoles(mockUser.Id, "system_user system_admin").
			Return(true, &model.Response{Error: nil}).
			Times(1)

		command := cobra.Command{}
		command.Flags().String("username", mockUser.Username, "")
		command.Flags().String("email", mockUser.Email, "")
		command.Flags().String("password", mockUser.Password, "")
		command.Flags().Bool("system_admin", true, "")

		error := userCreateCmdF(s.client, &command, []string{})

		s.Require().Nil(error)
		s.Require().Equal(&mockUser, printer.GetLines()[0])
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Create a sysAdmin user with client returning error", func() {
		printer.Clean()

		s.client.
			EXPECT().
			CreateUser(&mockUser).
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserRoles(mockUser.Id, "system_user system_admin").
			Return(false, &model.Response{Error: &model.AppError{Message: "Remote error"}}).
			Times(1)

		command := cobra.Command{}
		command.Flags().String("username", mockUser.Username, "")
		command.Flags().String("email", mockUser.Email, "")
		command.Flags().String("password", mockUser.Password, "")
		command.Flags().Bool("system_admin", true, "")

		error := userCreateCmdF(s.client, &command, []string{})

		s.Require().Equal("Unable to update user roles. Error: : Remote error, ", error.Error())
	})
}

func (s *MmctlUnitTestSuite) TestUpdateUserEmailCmd() {
	s.Run("Two arguments are not provided", func() {
		printer.Clean()

		command := cobra.Command{}

		error := updateUserEmailCmdF(s.client, &command, []string{})

		s.Require().EqualError(error, "expected two arguments. See help text for details")
	})

	s.Run("Invalid email provided", func() {
		printer.Clean()

		userArg := "testUser"
		emailArg := "invalidEmail"
		command := cobra.Command{}

		error := updateUserEmailCmdF(s.client, &command, []string{userArg, emailArg})

		s.Require().EqualError(error, "invalid email: 'invalidEmail'")
	})

	s.Run("User not found using email, username or id as identifier", func() {
		printer.Clean()

		command := cobra.Command{}
		userArg := "testUser"
		emailArg := "example@example.com"

		s.client.
			EXPECT().
			GetUserByEmail(userArg, "").
			Return(nil, &model.Response{Error: &model.AppError{Message: "No user found with the given email"}}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(userArg, "").
			Return(nil, &model.Response{Error: &model.AppError{Message: "No user found with the given username"}}).
			Times(1)

		s.client.
			EXPECT().
			GetUser(userArg, "").
			Return(nil, &model.Response{Error: &model.AppError{Message: "No user found with the given id"}}).
			Times(1)

		error := updateUserEmailCmdF(s.client, &command, []string{userArg, emailArg})

		s.Require().EqualError(error, "unable to find user 'testUser'")
	})

	s.Run("Client returning error while updating user", func() {
		printer.Clean()

		command := cobra.Command{}
		userArg := "testUser"
		emailArg := "example@example.com"

		currentUser := model.User{Username: "testUser", Password: "password", Email: "email"}

		s.client.
			EXPECT().
			GetUserByEmail(userArg, "").
			Return(nil, &model.Response{Error: &model.AppError{Message: "No user found with the given email"}}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(userArg, "").
			Return(&currentUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUser(&currentUser).
			Return(nil, &model.Response{Error: &model.AppError{Message: "Remote error"}}).
			Times(1)

		error := updateUserEmailCmdF(s.client, &command, []string{userArg, emailArg})

		s.Require().EqualError(error, ": Remote error, ")
	})

	s.Run("User email is updated successfully using username as identifier", func() {
		printer.Clean()

		command := cobra.Command{}
		userArg := "testUser"
		emailArg := "example@example.com"

		currentUser := model.User{Username: "testUser", Password: "password", Email: "email"}
		updatedUser := model.User{Username: "testUser", Password: "password", Email: emailArg}

		s.client.
			EXPECT().
			GetUserByEmail(userArg, "").
			Return(nil, &model.Response{Error: &model.AppError{Message: "No user found with the given email"}}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(userArg, "").
			Return(&currentUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUser(&currentUser).
			Return(&updatedUser, &model.Response{Error: nil}).
			Times(1)

		error := updateUserEmailCmdF(s.client, &command, []string{userArg, emailArg})

		s.Require().Nil(error)
		s.Require().Equal(&updatedUser, printer.GetLines()[0])
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("User email is updated successfully using email as identifier", func() {
		printer.Clean()

		command := cobra.Command{}
		userArg := "user@email.com"
		emailArg := "example@example.com"

		currentUser := model.User{Username: "testUser", Password: "password", Email: "email"}
		updatedUser := model.User{Username: "testUser", Password: "password", Email: emailArg}

		s.client.
			EXPECT().
			GetUserByEmail(userArg, "").
			Return(&currentUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUser(&currentUser).
			Return(&updatedUser, &model.Response{Error: nil}).
			Times(1)

		error := updateUserEmailCmdF(s.client, &command, []string{userArg, emailArg})

		s.Require().Nil(error)
		s.Require().Equal(&updatedUser, printer.GetLines()[0])
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("User email is updated successfully using id as identifier", func() {
		printer.Clean()

		command := cobra.Command{}
		userArg := "userId"
		emailArg := "example@example.com"

		currentUser := model.User{Username: "testUser", Password: "password", Email: "email"}
		updatedUser := model.User{Username: "testUser", Password: "password", Email: emailArg}

		s.client.
			EXPECT().
			GetUserByEmail(userArg, "").
			Return(nil, &model.Response{Error: &model.AppError{Message: "No user found with the given email"}}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(userArg, "").
			Return(nil, &model.Response{Error: &model.AppError{Message: "No user found with the given username"}}).
			Times(1)

		s.client.
			EXPECT().
			GetUser(userArg, "").
			Return(&currentUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUser(&currentUser).
			Return(&updatedUser, &model.Response{Error: nil}).
			Times(1)

		error := updateUserEmailCmdF(s.client, &command, []string{userArg, emailArg})

		s.Require().Nil(error)
		s.Require().Equal(&updatedUser, printer.GetLines()[0])
		s.Require().Len(printer.GetErrorLines(), 0)
	})
}

func (s *MmctlUnitTestSuite) TestResetUserMfaCmd() {
	s.Run("One user without problems", func() {
		printer.Clean()

		s.client.
			EXPECT().
			GetUserByEmail("userId", "").
			Return(&model.User{Id: "userId"}, nil).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserMfa("userId", "", false).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := resetUserMfaCmdF(s.client, &cobra.Command{}, []string{"userId"})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Cannot find one user", func() {
		printer.Clean()

		s.client.
			EXPECT().
			GetUserByEmail("userId", "").
			Return(nil, nil).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername("userId", "").
			Return(nil, nil).
			Times(1)

		s.client.
			EXPECT().
			GetUser("userId", "").
			Return(nil, nil).
			Times(1)

		err := resetUserMfaCmdF(s.client, &cobra.Command{}, []string{"userId"})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(printer.GetErrorLines()[0], "Unable to find user 'userId'")
	})

	s.Run("One user, unable to reset", func() {
		printer.Clean()
		mockError := model.AppError{Message: "Mock error"}

		s.client.
			EXPECT().
			GetUserByEmail("userId", "").
			Return(&model.User{Id: "userId"}, nil).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserMfa("userId", "", false).
			Return(false, &model.Response{Error: &mockError}).
			Times(1)

		err := resetUserMfaCmdF(s.client, &cobra.Command{}, []string{"userId"})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(printer.GetErrorLines()[0], "Unable to reset user 'userId' MFA. Error: "+mockError.Error())
	})

	s.Run("Several users, with unknown users and users unable to be reset", func() {
		printer.Clean()
		users := []string{"user0", "error1", "user2", "unknown3", "user4"}
		mockError := model.AppError{Message: "Mock error"}

		for _, user := range users {
			if user != "unknown3" {
				s.client.
					EXPECT().
					GetUserByEmail(user, "").
					Return(&model.User{Id: user}, nil).
					Times(1)
			} else {
				s.client.
					EXPECT().
					GetUserByEmail(user, "").
					Return(nil, nil).
					Times(1)

				s.client.
					EXPECT().
					GetUserByUsername(user, "").
					Return(nil, nil).
					Times(1)

				s.client.
					EXPECT().
					GetUser(user, "").
					Return(nil, nil).
					Times(1)
			}
		}

		for _, user := range users {
			if user == "error1" {
				s.client.
					EXPECT().
					UpdateUserMfa(user, "", false).
					Return(false, &model.Response{Error: &mockError}).
					Times(1)
			} else if user != "unknown3" {
				s.client.
					EXPECT().
					UpdateUserMfa(user, "", false).
					Return(true, &model.Response{Error: nil}).
					Times(1)
			}
		}

		err := resetUserMfaCmdF(s.client, &cobra.Command{}, users)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 2)
		s.Require().Equal(printer.GetErrorLines()[0], "Unable to reset user '"+users[1]+"' MFA. Error: "+mockError.Error())
		s.Require().Equal(printer.GetErrorLines()[1], "Unable to find user '"+users[3]+"'")
	})
}

func (s *MmctlUnitTestSuite) TestListUserCmdF() {
	cmd := &cobra.Command{}
	cmd.Flags().Int("page", 0, "")
	cmd.Flags().Int("per-page", 200, "")
	cmd.Flags().Bool("all", false, "")

	s.Run("Listing users with paging", func() {
		printer.Clean()

		email := "example@example.com"
		mockUser := model.User{Username: "ExampleUser", Email: email}

		page := 0
		perPage := 1
		showAll := false
		_ = cmd.Flags().Set("page", strconv.Itoa(page))
		_ = cmd.Flags().Set("per-page", strconv.Itoa(perPage))
		_ = cmd.Flags().Set("all", strconv.FormatBool(showAll))

		s.client.
			EXPECT().
			GetUsers(page, perPage, "").
			Return([]*model.User{&mockUser}, &model.Response{Error: nil}).
			Times(1)

		err := listUsersCmdF(s.client, cmd, []string{})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(&mockUser, printer.GetLines()[0])
	})

	s.Run("Listing all the users", func() {
		printer.Clean()

		email1 := "example1@example.com"
		mockUser1 := model.User{Username: "ExampleUser1", Email: email1}
		email2 := "example2@example.com"
		mockUser2 := model.User{Username: "ExampleUser2", Email: email2}

		page := 0
		perPage := 1
		showAll := true
		_ = cmd.Flags().Set("page", strconv.Itoa(page))
		_ = cmd.Flags().Set("per-page", strconv.Itoa(perPage))
		_ = cmd.Flags().Set("all", strconv.FormatBool(showAll))

		s.client.
			EXPECT().
			GetUsers(0, perPage, "").
			Return([]*model.User{&mockUser1}, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUsers(1, perPage, "").
			Return([]*model.User{&mockUser2}, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUsers(2, perPage, "").
			Return([]*model.User{}, &model.Response{Error: nil}).
			Times(1)

		err := listUsersCmdF(s.client, cmd, []string{})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 2)
		s.Require().Equal(&mockUser1, printer.GetLines()[0])
		s.Require().Equal(&mockUser2, printer.GetLines()[1])
	})

	s.Run("Try to list all the users when there are no uses in store", func() {
		printer.Clean()

		page := 0
		perPage := 1
		showAll := false
		_ = cmd.Flags().Set("page", strconv.Itoa(page))
		_ = cmd.Flags().Set("per-page", strconv.Itoa(perPage))
		_ = cmd.Flags().Set("all", strconv.FormatBool(showAll))

		s.client.
			EXPECT().
			GetUsers(page, perPage, "").
			Return([]*model.User{}, &model.Response{Error: nil}).
			Times(1)

		err := listUsersCmdF(s.client, cmd, []string{})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
	})

	s.Run("Return an error from GetUsers call and verify that error is properly returned", func() {
		printer.Clean()

		page := 0
		perPage := 1
		showAll := false
		_ = cmd.Flags().Set("page", strconv.Itoa(page))
		_ = cmd.Flags().Set("per-page", strconv.Itoa(perPage))
		_ = cmd.Flags().Set("all", strconv.FormatBool(showAll))

		mockError := model.AppError{Id: "Mock Error"}
		mockErrorW := errors.Wrap(&mockError, "Failed to fetch users")

		s.client.
			EXPECT().
			GetUsers(page, perPage, "").
			Return(nil, &model.Response{Error: &mockError}).
			Times(1)

		err := listUsersCmdF(s.client, cmd, []string{})
		s.Require().NotNil(err)
		s.Require().EqualError(err, mockErrorW.Error())
	})

	s.Run("Start with page 2 where a server has total 3 pages", func() {
		printer.Clean()

		email := "example@example.com"
		mockUser := model.User{Username: "ExampleUser", Email: email}

		page := 2
		perPage := 1
		showAll := false
		_ = cmd.Flags().Set("page", strconv.Itoa(page))
		_ = cmd.Flags().Set("per-page", strconv.Itoa(perPage))
		_ = cmd.Flags().Set("all", strconv.FormatBool(showAll))

		s.client.
			EXPECT().
			GetUsers(page, perPage, "").
			Return([]*model.User{&mockUser}, &model.Response{Error: nil}).
			Times(1)

		err := listUsersCmdF(s.client, cmd, []string{})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(&mockUser, printer.GetLines()[0])
	})
}

func (s *MmctlUnitTestSuite) TestUserDeactivateCmd() {
	s.Run("Deactivate an existing user using email", func() {
		printer.Clean()
		emailArg := "example@example.com"
		mockUser := model.User{Username: "ExampleUser", Email: emailArg}

		s.client.
			EXPECT().
			GetUserByEmail(emailArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserActive(mockUser.Id, false).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := userDeactivateCmdF(s.client, &cobra.Command{}, []string{mockUser.Email})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})
	s.Run("Deactivate an existing user by username", func() {
		printer.Clean()
		emailArg := "example@exam.com"
		usernameArg := "ExampleUser"
		mockUser := model.User{Username: usernameArg, Email: emailArg}

		s.client.
			EXPECT().
			GetUserByEmail(usernameArg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(usernameArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserActive(mockUser.Id, false).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := userDeactivateCmdF(s.client, &cobra.Command{}, []string{mockUser.Username})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Deactivate an existing user by id", func() {
		printer.Clean()
		mockUser := model.User{Id: "userId1", Username: "ExampleUser", Email: "example@exam.com"}

		s.client.
			EXPECT().
			GetUserByEmail(mockUser.Id, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(mockUser.Id, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUser(mockUser.Id, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserActive(mockUser.Id, false).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := userDeactivateCmdF(s.client, &cobra.Command{}, []string{mockUser.Id})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Deactivate SSO user", func() {
		printer.Clean()
		arg := "example@example.com"
		mockUser := model.User{Username: "ExampleUser", Email: arg, AuthService: "SSO"}

		s.client.
			EXPECT().
			GetUserByEmail(arg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserActive(mockUser.Id, false).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := userDeactivateCmdF(s.client, &cobra.Command{}, []string{arg})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal("You must also deactivate user "+arg+" in the SSO provider or they will be reactivated on next login or sync.", printer.GetLines()[0])
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Deactivate nonexistent user", func() {
		printer.Clean()
		arg := "example@example.com"

		s.client.
			EXPECT().
			GetUserByEmail(arg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(arg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUser(arg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := userDeactivateCmdF(s.client, &cobra.Command{}, []string{arg})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(fmt.Errorf("can't find user '%v'", arg).Error(), printer.GetErrorLines()[0])
	})

	s.Run("Delete multiple users", func() {
		printer.Clean()
		mockUser1 := model.User{Id: "userId1", Email: "user1@example.com", Username: "user1"}
		mockUser2 := model.User{Id: "userId2", Email: "user2@example.com", Username: "user2"}
		mockUser3 := model.User{Id: "userId3", Email: "user3@example.com", Username: "user3"}

		argEmails := []string{mockUser1.Email, mockUser2.Email, mockUser3.Email}
		argUsers := []model.User{mockUser1, mockUser2, mockUser3}

		for i := 0; i < len(argEmails); i++ {
			s.client.
				EXPECT().
				GetUserByEmail(argEmails[i], "").
				Return(&argUsers[i], &model.Response{Error: nil}).
				Times(1)
		}

		for i := 0; i < len(argEmails); i++ {
			s.client.
				EXPECT().
				UpdateUserActive(argUsers[i].Id, false).
				Return(true, &model.Response{Error: nil}).
				Times(1)
		}

		err := userDeactivateCmdF(s.client, &cobra.Command{}, argEmails)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Delete multiple users with argument mixture of emails usernames and userIds", func() {
		printer.Clean()
		mockUser1 := model.User{Id: "userId1", Email: "user1@example.com", Username: "user1"}
		mockUser2 := model.User{Id: "userId2", Email: "user2@example.com", Username: "user2"}
		mockUser3 := model.User{Id: "userId3", Email: "user3@example.com", Username: "user3"}

		argsDelete := []string{mockUser1.Id, mockUser2.Email, mockUser3.Username}
		argUsers := []model.User{mockUser1, mockUser2, mockUser3}

		// mockUser1
		s.client.
			EXPECT().
			GetUserByEmail(argsDelete[0], "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(argsDelete[0], "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUser(argsDelete[0], "").
			Return(&argUsers[0], &model.Response{Error: nil}).
			Times(1)

		// mockUser2
		s.client.
			EXPECT().
			GetUserByEmail(argsDelete[1], "").
			Return(&argUsers[1], &model.Response{Error: nil}).
			Times(1)

		// mockUser3
		s.client.
			EXPECT().
			GetUserByEmail(argsDelete[2], "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			GetUserByUsername(argsDelete[2], "").
			Return(&argUsers[2], &model.Response{Error: nil}).
			Times(1)

		for _, user := range argUsers {
			s.client.
				EXPECT().
				UpdateUserActive(user.Id, false).
				Return(true, &model.Response{Error: nil}).
				Times(1)
		}

		err := userDeactivateCmdF(s.client, &cobra.Command{}, argsDelete)
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Delete multiple users with an non existent user", func() {
		printer.Clean()
		mockUser1 := model.User{Id: "userId1", Email: "user1@example.com", Username: "user1"}
		nonexistentEmail := "example@example.com"

		// mockUser1
		s.client.
			EXPECT().
			GetUserByEmail(mockUser1.Email, "").
			Return(&mockUser1, &model.Response{Error: nil}).
			Times(1)

		// nonexistent email
		s.client.
			EXPECT().
			GetUserByEmail(nonexistentEmail, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(nonexistentEmail, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUser(nonexistentEmail, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserActive(mockUser1.Id, false).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := userDeactivateCmdF(s.client, &cobra.Command{}, []string{mockUser1.Email, nonexistentEmail})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(fmt.Errorf("can't find user '%v'", nonexistentEmail).Error(), printer.GetErrorLines()[0])
	})
}
