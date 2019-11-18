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

func (s *MmctlUnitTestSuite) TestListUserCmdF() {
	s.Run("Listing users with paging", func() {
		printer.Clean()

		emailArg := "example@example.com"
		mockUser := model.User{Username: "ExampleUser", Email: emailArg}
		page := 0
		perPage := 1

		cmd := &cobra.Command{}
		cmd.Flags().Int("per-page", perPage, "")
		cmd.Flags().Int("page", page, "")

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

		emailArg := "example2@example.com"
		mockUser := model.User{Username: "ExampleUser2", Email: emailArg}
		perPage := 200

		cmd := &cobra.Command{}
		cmd.Flags().Bool("all", true, "")
		cmd.Flags().Int("per-page", perPage, "")

		s.client.
			EXPECT().
			GetUsers(0, perPage, "").
			Return([]*model.User{&mockUser}, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUsers(1, perPage, "").
			Return([]*model.User{}, &model.Response{Error: nil}).
			Times(1)

		err := listUsersCmdF(s.client, cmd, []string{})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(&mockUser, printer.GetLines()[0])
	})

	s.Run("Try to listing all the users when there are no uses in store", func() {
		printer.Clean()

		page := 0
		perPage := 1

		cmd := &cobra.Command{}
		cmd.Flags().Int("per-page", perPage, "")
		cmd.Flags().Int("page", page, "")

		s.client.
			EXPECT().
			GetUsers(page, perPage, "").
			Return([]*model.User{}, &model.Response{Error: nil}).
			Times(1)

		err := listUsersCmdF(s.client, cmd, []string{})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
	})
}
