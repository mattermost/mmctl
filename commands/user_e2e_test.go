// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/spf13/cobra"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"
)

func (s *MmctlE2ETestSuite) TestUserActivateCmd() {
	s.SetupTestHelper().InitBasic()

	user, appErr := s.th.App.CreateUser(&model.User{Email: s.th.GenerateTestEmail(), Username: model.NewId(), Password: model.NewId()})
	s.Require().Nil(appErr)

	s.RunForSystemAdminAndLocal("Activate user", func(c client.Client) {
		printer.Clean()

		_, appErr := s.th.App.UpdateActive(user, false)
		s.Require().Nil(appErr)

		err := userActivateCmdF(c, &cobra.Command{}, []string{user.Email})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)

		ruser, err := s.th.App.GetUser(user.Id)
		s.Require().Nil(err)
		s.Require().Zero(ruser.DeleteAt)
	})

	s.Run("Activate user without permissions", func() {
		printer.Clean()

		_, appErr := s.th.App.UpdateActive(user, false)
		s.Require().Nil(appErr)

		err := userActivateCmdF(s.th.Client, &cobra.Command{}, []string{user.Email})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(printer.GetErrorLines()[0], "unable to change activation status of user: "+user.Email)

		ruser, err := s.th.App.GetUser(user.Id)
		s.Require().Nil(err)
		s.Require().NotZero(ruser.DeleteAt)
	})

	s.RunForAllClients("Activate nonexistent user", func(c client.Client) {
		printer.Clean()

		err := userActivateCmdF(c, &cobra.Command{}, []string{"nonexistent@email"})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(printer.GetErrorLines()[0], "can't find user 'nonexistent@email'")
	})
}

func (s *MmctlE2ETestSuite) TestUserDeactivateCmd() {
	s.SetupTestHelper().InitBasic()

	user, appErr := s.th.App.CreateUser(&model.User{Email: s.th.GenerateTestEmail(), Username: model.NewId(), Password: model.NewId()})
	s.Require().Nil(appErr)

	s.RunForSystemAdminAndLocal("Deactivate user", func(c client.Client) {
		printer.Clean()

		_, appErr := s.th.App.UpdateActive(user, true)
		s.Require().Nil(appErr)

		err := userDeactivateCmdF(c, &cobra.Command{}, []string{user.Email})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)

		ruser, err := s.th.App.GetUser(user.Id)
		s.Require().Nil(err)
		s.Require().NotZero(ruser.DeleteAt)
	})

	s.Run("Deactivate user without permissions", func() {
		printer.Clean()

		_, appErr := s.th.App.UpdateActive(user, true)
		s.Require().Nil(appErr)

		err := userDeactivateCmdF(s.th.Client, &cobra.Command{}, []string{user.Email})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(printer.GetErrorLines()[0], "unable to change activation status of user: "+user.Email)

		ruser, err := s.th.App.GetUser(user.Id)
		s.Require().Nil(err)
		s.Require().Zero(ruser.DeleteAt)
	})

	s.RunForAllClients("Deactivate nonexistent user", func(c client.Client) {
		printer.Clean()

		err := userDeactivateCmdF(c, &cobra.Command{}, []string{"nonexistent@email"})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(printer.GetErrorLines()[0], "can't find user 'nonexistent@email'")
	})
}

func (s *MmctlE2ETestSuite) TestSearchUserCmd() {
	s.SetupTestHelper().InitBasic()

	s.RunForAllClients("Search for an existing user", func(c client.Client) {
		printer.Clean()

		err := searchUserCmdF(c, &cobra.Command{}, []string{s.th.BasicUser.Email})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		user := printer.GetLines()[0].(*model.User)
		s.Equal(s.th.BasicUser.Username, user.Username)
		s.Len(printer.GetErrorLines(), 0)
	})

	s.RunForAllClients("Search for a nonexistent user", func(c client.Client) {
		printer.Clean()
		emailArg := "nonexistentUser@example.com"

		err := searchUserCmdF(c, &cobra.Command{}, []string{emailArg})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 1)
		s.Equal("Unable to find user '"+emailArg+"'", printer.GetErrorLines()[0])
	})
}

func (s *MmctlE2ETestSuite) TestCreateUserCmd() {
	s.SetupTestHelper().InitBasic()

	s.RunForAllClients("Should not create a user w/o username", func(c client.Client) {
		printer.Clean()
		email := s.th.GenerateTestEmail()
		cmd := &cobra.Command{}
		cmd.Flags().String("password", "somepass", "")
		cmd.Flags().String("email", email, "")

		err := userCreateCmdF(c, cmd, []string{})
		s.EqualError(err, "Username is required: flag accessed but not defined: username")
		s.Require().Empty(printer.GetLines())
		_, err = s.th.App.GetUserByEmail(email)
		s.Require().NotNil(err)
		s.Require().Contains(err.Error(), "SqlUserStore.GetByEmail: Unable to find the user., email="+email+", sql: no rows in result set")
	})

	s.RunForAllClients("Should not create a user w/o email", func(c client.Client) {
		printer.Clean()
		username := model.NewId()
		cmd := &cobra.Command{}
		cmd.Flags().String("username", username, "")
		cmd.Flags().String("password", "somepass", "")

		err := userCreateCmdF(c, cmd, []string{})
		s.EqualError(err, "Email is required: flag accessed but not defined: email")
		s.Require().Empty(printer.GetLines())
		_, err = s.th.App.GetUserByUsername(username)
		s.Require().NotNil(err)
		s.Require().Contains(err.Error(), "SqlUserStore.GetByUsername: Unable to find an existing account matching your username for this team. This team may require an invite from the team owner to join., sql: no rows in result set")
	})

	s.RunForAllClients("Should not create a user w/o password", func(c client.Client) {
		printer.Clean()
		email := s.th.GenerateTestEmail()
		cmd := &cobra.Command{}
		cmd.Flags().String("username", model.NewId(), "")
		cmd.Flags().String("email", email, "")

		err := userCreateCmdF(c, cmd, []string{})
		s.EqualError(err, "Password is required: flag accessed but not defined: password")
		s.Require().Empty(printer.GetLines())
		_, err = s.th.App.GetUserByEmail(email)
		s.Require().NotNil(err)
		s.Require().Contains(err.Error(), "SqlUserStore.GetByEmail: Unable to find the user., email="+email+", sql: no rows in result set")
	})

	s.Run("Should create a user but w/o system_admin privileges", func() {
		printer.Clean()
		email := s.th.GenerateTestEmail()
		username := model.NewId()
		cmd := &cobra.Command{}
		cmd.Flags().String("username", username, "")
		cmd.Flags().String("email", email, "")
		cmd.Flags().String("password", "password", "")
		cmd.Flags().Bool("system_admin", true, "")

		err := userCreateCmdF(s.th.Client, cmd, []string{})
		s.EqualError(err, "Unable to update user roles. Error: : You do not have the appropriate permissions., ")
		s.Require().Empty(printer.GetLines())
		user, err := s.th.App.GetUserByEmail(email)
		s.Require().Nil(err)
		s.Equal(username, user.Username)
		s.Equal(false, user.IsSystemAdmin())
	})

	s.RunForSystemAdminAndLocal("Should create new system_admin user given required params", func(c client.Client) {
		printer.Clean()
		email := s.th.GenerateTestEmail()
		username := model.NewId()
		cmd := &cobra.Command{}
		cmd.Flags().String("username", username, "")
		cmd.Flags().String("email", email, "")
		cmd.Flags().String("password", "somepass", "")
		cmd.Flags().Bool("system_admin", true, "")

		err := userCreateCmdF(s.th.SystemAdminClient, cmd, []string{})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		user, err := s.th.App.GetUserByEmail(email)
		s.Require().Nil(err)
		s.Equal(username, user.Username)
		s.Equal(true, user.IsSystemAdmin())
	})

	s.RunForAllClients("Should create new user given required params", func(c client.Client) {
		printer.Clean()
		email := s.th.GenerateTestEmail()
		username := model.NewId()
		cmd := &cobra.Command{}
		cmd.Flags().String("username", username, "")
		cmd.Flags().String("email", email, "")
		cmd.Flags().String("password", "somepass", "")

		err := userCreateCmdF(c, cmd, []string{})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		user, err := s.th.App.GetUserByEmail(email)
		s.Require().Nil(err)
		s.Equal(username, user.Username)
		s.Equal(false, user.IsSystemAdmin())
	})
}
