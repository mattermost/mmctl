package commands

import (
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mmctl/printer"

	"github.com/spf13/cobra"
)

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
			DeleteUser(mockUser.Id).
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
			DeleteUser(mockUser.Id).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := userDeactivateCmdF(s.client, &cobra.Command{}, []string{mockUser.Username})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Deactivate an existing user by id", func() {
		printer.Clean()
		mockUser := model.User{Username: "ExampleUser", Email: "example@exam.com"}

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
			DeleteUser(mockUser.Id).
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
			DeleteUser(mockUser.Id).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := userDeactivateCmdF(s.client, &cobra.Command{}, []string{arg})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal("You must also deactivate user " + arg + " in the SSO provider or they will be reactivated on next login or sync.", printer.GetLines()[0])
		s.Require().Len(printer.GetErrorLines(), 0)
	})

}
