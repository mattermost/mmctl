// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"os"
	"fmt"
	"io/ioutil"

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


func (s *MmctlE2ETestSuite) TestDeleteUsersCmd() {
	s.SetupTestHelper().InitBasic()

	s.RunForSystemAdminAndLocal("Delete user", func(c client.Client) {
		printer.Clean()

		previousVal := s.th.App.Config().ServiceSettings.EnableAPIUserDeletion
		s.th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableAPIUserDeletion = true })
		defer s.th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableAPIUserDeletion = *previousVal })

		cmd := &cobra.Command{}
		confirm := true
		cmd.Flags().BoolVar(&confirm, "confirm",confirm,"confirm")

		newUser, appErr := s.th.CreateUser()
		s.Require().Nil(appErr)
		err := deleteUsersCmdF(c ,cmd, []string{newUser.Email})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		s.Len(printer.GetErrorLines(), 0)
		
		deletedUser := printer.GetLines()[0].(*model.User)
		s.Require().Equal(newUser.Username, deletedUser.Username)

		// expect user deleted
		_, err = s.th.App.GetUser(newUser.Id)
		s.Require().NotNil(err)
		s.Require().Equal(err.Error(), "SqlUserStore.Get: Unable to find the user., user_id=store.sql_user.missing_account.const, sql: no rows in result set")
	})

	s.RunForSystemAdminAndLocal("Delete user confirm using prompt", func(c client.Client) {
		printer.Clean()

		previousVal := s.th.App.Config().ServiceSettings.EnableAPIUserDeletion
		s.th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableAPIUserDeletion = true })
		defer func() {
			s.th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableAPIUserDeletion = *previousVal })
		}()

		cmd := &cobra.Command{}

		// create temp file to replace stdin
		content := []byte("YES\nYES\n")
		tmpfile, err := ioutil.TempFile("", "inputfile")
		s.Require().Nil(err)
		defer os.Remove(tmpfile.Name()) // remove temp file
	
		_,err = tmpfile.Write(content)
		s.Require().Nil(err)
		_,err = tmpfile.Seek(0, 0)
		s.Require().Nil(err)
		
		// replace stdin to do input in testing
		oldStdin := os.Stdin
		defer func() { os.Stdin = oldStdin }() // restore
		os.Stdin = tmpfile

		newUser := s.th.CreateUser()
		err = deleteUsersCmdF(c ,cmd, []string{newUser.Email})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		s.Len(printer.GetErrorLines(), 0)
		
		deletedUser := printer.GetLines()[0].(*model.User)
		s.Require().Equal(newUser.Username, deletedUser.Username)

		// expect user deleted
		_, err = s.th.App.GetUser(newUser.Id)
		s.Require().NotNil(err)
		s.Require().Equal(err.Error(), "SqlUserStore.Get: Unable to find the user., user_id=store.sql_user.missing_account.const, sql: no rows in result set")
	})

	s.RunForSystemAdminAndLocal("Delete nonexistent user", func(c client.Client) {
		printer.Clean()
		emailArg := "nonexistentUser@example.com"

		previousVal := s.th.App.Config().ServiceSettings.EnableAPIUserDeletion
		s.th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableAPIUserDeletion = true })
		defer func() {
			s.th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableAPIUserDeletion = *previousVal })
		}()

		cmd := &cobra.Command{}
		confirm := true
		cmd.Flags().BoolVar(&confirm, "confirm",confirm,"confirm")

		err := deleteUsersCmdF(s.th.Client ,cmd, []string{emailArg})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 1)
		s.Equal("Unable to find user '"+emailArg+"'", printer.GetErrorLines()[0])
	})

	s.Run("Delete user without permission", func() {
		printer.Clean()

		previousVal := s.th.App.Config().ServiceSettings.EnableAPIUserDeletion
		s.th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableAPIUserDeletion = true })
		defer func() {
			s.th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableAPIUserDeletion = *previousVal })
		}()

		cmd := &cobra.Command{}
		confirm := true
		cmd.Flags().BoolVar(&confirm, "confirm",confirm,"confirm")

		newUser := s.th.CreateUser()
		err := deleteUsersCmdF(s.th.Client ,cmd, []string{newUser.Email})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 1)
		s.Require().Equal(printer.GetErrorLines()[0], fmt.Sprintf("Unable to delete user '%s' error: : You do not have the appropriate permissions., ", newUser.Username))

		// expect user not deleted
		user, err := s.th.App.GetUser(newUser.Id)
		s.Require().Nil(err)
		s.Require().Equal(newUser.Username, user.Username)
	})

	s.Run("Delete user with disabled config", func() {
		printer.Clean()

		previousVal := s.th.App.Config().ServiceSettings.EnableAPIUserDeletion
		s.th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableAPIUserDeletion = false })
		defer func() {
			s.th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableAPIUserDeletion = *previousVal })
		}()

		cmd := &cobra.Command{}
		confirm := true
		cmd.Flags().BoolVar(&confirm, "confirm",confirm,"confirm")

		newUser := s.th.CreateUser()
		err := deleteUsersCmdF(s.th.SystemAdminClient ,cmd, []string{newUser.Email})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 1)
		s.Require().Equal(printer.GetErrorLines()[0], fmt.Sprintf("Unable to delete user '%s' error: : Permanent user deletion feature is not enabled. Please contact your System Administrator., ", newUser.Username))

		// expect user not deleted
		user, err := s.th.App.GetUser(newUser.Id)
		s.Require().Nil(err)
		s.Require().Equal(newUser.Username, user.Username)
	})
}
