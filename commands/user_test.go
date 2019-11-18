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
		s.Require().Equal("Unable to find user '"+arg+"'", printer.GetErrorLines()[0])
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
				DeleteUser(argUsers[i].Id).
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
				DeleteUser(user.Id).
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
			DeleteUser(mockUser1.Id).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := userDeactivateCmdF(s.client, &cobra.Command{}, []string{mockUser1.Email, nonexistentEmail})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal("Unable to find user '"+nonexistentEmail+"'", printer.GetErrorLines()[0])
	})

}
