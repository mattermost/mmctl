package commands

import (
	"strings"

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
		s.Require().Nil(err)
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
		s.Require().Nil(err)
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
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 2)
		s.Require().Equal("Unable send reset password email to email "+emailArg[1]+". Error: "+mockError.Error(), printer.GetErrorLines()[0])
		s.Require().Equal("Unable send reset password email to email "+emailArg[2]+". Error: "+mockError.Error(), printer.GetErrorLines()[1])
	})
}
