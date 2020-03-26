// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mmctl/printer"

	"github.com/spf13/cobra"
)

func (s *MmctlUnitTestSuite) TestGenerateTokenForAUserCmd() {
	s.Run("Should generate a token for a user", func() {
		printer.Clean()

		userArg := "userId1"
		mockUser := model.User{Id: "userId1", Email: "user1@example.com", Username: "user1"}
		mockToken := model.UserAccessToken{Token: "token-id", Description: "token-desc"}

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
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			CreateUserAccessToken(mockUser.Id, mockToken.Description).
			Return(&mockToken, &model.Response{Error: nil}).
			Times(1)

		err := generateTokenForAUserCmdF(s.client, &cobra.Command{}, []string{mockUser.Id, mockToken.Description})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(&mockToken, printer.GetLines()[0])
	})

	s.Run("Should fail on an invalid username", func() {
		printer.Clean()

		userArg := "some-text"
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
			Return(nil, &model.Response{Error: &model.AppError{Message: "No user found with the given ID"}}).
			Times(1)

		err := generateTokenForAUserCmdF(s.client, &cobra.Command{}, []string{userArg, "description"})
		s.Require().NotNil(err)
		s.Require().Contains(err.Error(), fmt.Sprintf("could not retrieve user information of %q", userArg))
	})

	s.Run("Should fail if can't create tokens for a valid user", func() {
		printer.Clean()

		userArg := "user1"
		mockUser := model.User{Id: "userId1", Email: "user1@example.com", Username: "user1"}

		s.client.
			EXPECT().
			GetUserByEmail(userArg, "").
			Return(nil, &model.Response{Error: &model.AppError{Message: "No user found with the given email"}}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(userArg, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			CreateUserAccessToken(mockUser.Id, "description").
			Return(nil, &model.Response{Error: &model.AppError{Message: "error-message"}}).
			Times(1)

		err := generateTokenForAUserCmdF(s.client, &cobra.Command{}, []string{"user1", "description"})
		s.Require().NotNil(err)
		s.Require().Contains(err.Error(), fmt.Sprintf("could not create token for %q:", "user1"))
	})
}

func (s *MmctlUnitTestSuite) TestListTokensOfAUserCmdF() {
	s.Run("Should list access tokens of a user", func() {
		printer.Clean()

		command := cobra.Command{}
		command.Flags().Int("page", 0, "")
		command.Flags().Int("per-page", 2, "")
		command.Flags().Bool("all", true, "")
		command.Flags().Bool("active", false, "")
		command.Flags().Bool("inactive", false, "")

		mockUser := model.User{Id: "userId1", Email: "user1@example.com", Username: "user1"}
		mockToken1 := model.UserAccessToken{IsActive: true, Id: "token-1-id", Description: "token-1-desc"}
		mockToken2 := model.UserAccessToken{IsActive: false, Id: "token-2-id", Description: "token-2-desc"}

		s.client.
			EXPECT().
			GetUserByEmail(mockUser.Id, "").
			Return(nil, &model.Response{Error: &model.AppError{Message: "No user found with the given email"}}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(mockUser.Id, "").
			Return(nil, &model.Response{Error: &model.AppError{Message: "No user found with the given username"}}).
			Times(1)

		s.client.
			EXPECT().
			GetUser(mockUser.Id, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserAccessTokensForUser(mockUser.Id, 0, 9999).
			Return(
				[]*model.UserAccessToken{&mockToken1, &mockToken2},
				&model.Response{Error: nil},
			).Times(1)

		err := listTokensOfAUserCmdF(s.client, &command, []string{mockUser.Id})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 2)
		s.Require().Equal(&mockToken1, printer.GetLines()[0])
		s.Require().Equal(&mockToken2, printer.GetLines()[1])
	})

	s.Run("Should list only active user access tokens of a user", func() {
		printer.Clean()

		command := cobra.Command{}
		command.Flags().Int("page", 0, "")
		command.Flags().Int("per-page", 2, "")
		command.Flags().Bool("all", false, "")
		command.Flags().Bool("active", true, "")
		command.Flags().Bool("inactive", false, "")

		mockUser := model.User{Id: "userId1", Email: "user1@example.com", Username: "user1"}
		mockToken1 := model.UserAccessToken{IsActive: true, Id: "token-1-id", Description: "token-1-desc"}
		mockToken2 := model.UserAccessToken{IsActive: false, Id: "token-2-id", Description: "token-2-desc"}

		s.client.
			EXPECT().
			GetUserByEmail(mockUser.Email, "").
			Return(&mockUser, &model.Response{Error: &model.AppError{Message: "No user found with the given email"}}).
			Times(1)

		s.client.
			EXPECT().
			GetUserAccessTokensForUser(mockUser.Id, 0, 2).
			Return(
				[]*model.UserAccessToken{&mockToken1, &mockToken2},
				&model.Response{Error: nil},
			).Times(1)

		err := listTokensOfAUserCmdF(s.client, &command, []string{mockUser.Email})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(&mockToken1, printer.GetLines()[0])
	})

	s.Run("Should err on a absent user", func() {
		printer.Clean()

		userArg := "test-user"
		command := cobra.Command{}
		command.Flags().Int("page", 0, "")
		command.Flags().Int("per-page", 2, "")
		command.Flags().Bool("all", false, "")
		command.Flags().Bool("active", false, "")
		command.Flags().Bool("inactive", false, "")

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
			Return(nil, &model.Response{Error: &model.AppError{Message: "No user found with the given user ID"}}).
			Times(1)

		err := listTokensOfAUserCmdF(s.client, &command, []string{userArg})
		s.Require().NotNil(err)
		s.Require().Contains(err.Error(), fmt.Sprintf("could not retrieve user information of %q", userArg))
	})

	s.Run("Should error if there are no user access tokens for a valid user", func() {
		printer.Clean()

		command := cobra.Command{}
		command.Flags().Int("page", 0, "")
		command.Flags().Int("per-page", 2, "")
		command.Flags().Bool("all", false, "")
		command.Flags().Bool("active", true, "")
		command.Flags().Bool("inactive", false, "")

		mockUser := model.User{Id: "userId1", Email: "user1@example.com", Username: "user1"}

		s.client.
			EXPECT().
			GetUserByEmail(mockUser.Email, "").
			Return(&mockUser, &model.Response{Error: &model.AppError{Message: "No user found with the given email"}}).
			Times(1)

		s.client.
			EXPECT().
			GetUserAccessTokensForUser(mockUser.Id, 0, 2).
			Return(
				[]*model.UserAccessToken{},
				&model.Response{Error: nil},
			).Times(1)

		err := listTokensOfAUserCmdF(s.client, &command, []string{mockUser.Email})
		s.Require().NotNil(err)
		s.Require().Equal(err.Error(), fmt.Sprintf("there are no tokens for the %q", mockUser.Email))
	})
}

func (s *MmctlUnitTestSuite) TestRevokeTokenForAUserCmdF() {
	s.Run("Should revoke user access tokens", func() {
		printer.Clean()

		mockToken1 := model.UserAccessToken{Token: "token-id1", Description: "token-desc1", Id: "123456"}
		mockToken2 := model.UserAccessToken{Token: "token-id2", Description: "token-desc2", Id: "234567"}

		s.client.
			EXPECT().
			RevokeUserAccessToken(mockToken1.Id).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			RevokeUserAccessToken(mockToken2.Id).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := revokeTokenForAUserCmdF(s.client, &cobra.Command{}, []string{mockToken1.Id, mockToken2.Id})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
	})

	s.Run("Should fail if can't revoke user access token", func() {
		s.client.
			EXPECT().
			RevokeUserAccessToken("token-id").
			Return(false, &model.Response{Error: &model.AppError{Message: "some-error"}}).
			Times(1)

		err := revokeTokenForAUserCmdF(s.client, &cobra.Command{}, []string{"token-id"})
		s.Require().NotNil(err)
		s.Require().Contains(err.Error(), fmt.Sprintf("could not revoke token %q", "token-id"))
	})
}
