// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v5/api4"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/spf13/cobra"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"
)

func (s *MmctlE2ETestSuite) TestChannelUsersAddCmdF() {
	s.SetupTestHelper().InitBasic()

	s.th.CreateUser()
	user, appErr := s.th.App.CreateUser(&model.User{Email: s.th.GenerateTestEmail(), Username: model.NewId(), Password: model.NewId()})
	s.Require().Nil(appErr)

	_, appErr = s.th.App.AddUserToTeam(s.th.BasicTeam.Id, user.Id, "")
	s.Require().Nil(appErr)

	channelName := api4.GenerateTestChannelName()
	channel, appErr := s.th.App.CreateChannel(&model.Channel{
		TeamId:      s.th.BasicTeam.Id,
		Name:        channelName,
		DisplayName: "db_" + channelName,
		Type:        model.CHANNEL_OPEN,
	}, false)
	s.Require().Nil(appErr)

	s.RunForSystemAdminAndLocal("Add user to nonexistent channel", func(c client.Client) {
		printer.Clean()

		nonexistentChannelName := "nonexistent"
		err := channelUsersAddCmdF(c, &cobra.Command{}, []string{nonexistentChannelName, user.Id})
		s.Require().NotNil(err)
		s.Require().Equal(fmt.Sprintf("unable to find channel \"%s\"", nonexistentChannelName), err.Error())
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Add user to nonexistent channel/Client", func() {
		printer.Clean()

		_, appErr := s.th.App.AddChannelMember(s.th.BasicUser.Id, channel, "", "")
		s.Require().Nil(appErr)
		defer func() {
			appErr := s.th.App.RemoveUserFromChannel(s.th.BasicUser.Id, s.th.SystemAdminUser.Id, channel)
			s.Require().Nil(appErr)
		}()

		nonexistentChannelName := "nonexistent"
		err := channelUsersAddCmdF(s.th.Client, &cobra.Command{}, []string{nonexistentChannelName, user.Id})
		s.Require().NotNil(err)
		s.Require().Equal(fmt.Sprintf("unable to find channel \"%s\"", nonexistentChannelName), err.Error())
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.RunForSystemAdminAndLocal("Add nonexistent user to channel", func(c client.Client) {
		printer.Clean()

		nonexistentUserName := "nonexistent"
		err := channelUsersAddCmdF(c, &cobra.Command{}, []string{channel.Id, nonexistentUserName})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(fmt.Sprintf("Can't find user '%s'", nonexistentUserName), printer.GetErrorLines()[0])
	})

	s.Run("Add nonexistent user to channel/Client", func() {
		printer.Clean()

		_, appErr := s.th.App.AddChannelMember(s.th.BasicUser.Id, channel, "", "")
		s.Require().Nil(appErr)
		defer func() {
			appErr := s.th.App.RemoveUserFromChannel(s.th.BasicUser.Id, s.th.SystemAdminUser.Id, channel)
			s.Require().Nil(appErr)
		}()

		nonexistentUserName := "nonexistent"
		err := channelUsersAddCmdF(s.th.Client, &cobra.Command{}, []string{channel.Id, nonexistentUserName})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(fmt.Sprintf("Can't find user '%s'", nonexistentUserName), printer.GetErrorLines()[0])
	})

	s.Run("Add user to channel without permission/Client", func() {
		printer.Clean()

		err := channelUsersAddCmdF(s.th.Client, &cobra.Command{}, []string{channel.Id, user.Id})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(fmt.Sprintf("Unable to add '%s' to %s. Error: : You do not have the appropriate permissions., ", user.Id, channelName), printer.GetErrorLines()[0])
	})

	s.Run("Add user to channel/Client", func() {
		printer.Clean()

		_, appErr := s.th.App.AddChannelMember(s.th.BasicUser.Id, channel, "", "")
		s.Require().Nil(appErr)
		defer func() {
			appErr = s.th.App.RemoveUserFromChannel(s.th.BasicUser.Id, s.th.SystemAdminUser.Id, channel)
			s.Require().Nil(appErr)
		}()

		err := channelUsersAddCmdF(s.th.Client, &cobra.Command{}, []string{channel.Id, user.Id})
		s.Require().Nil(err)
		defer func() {
			appErr = s.th.App.RemoveUserFromChannel(user.Id, s.th.SystemAdminUser.Id, channel)
			s.Require().Nil(appErr)
		}()
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)

		members, appErr := s.th.App.GetChannelMembersByIds(channel.Id, []string{user.Id})
		s.Require().Nil(appErr)
		s.Require().Len(*members, 1)
		s.Require().Equal(user.Id, (*members)[0].UserId)
	})

	s.RunForSystemAdminAndLocal("Add user to channel", func(c client.Client) {
		printer.Clean()

		err := channelUsersAddCmdF(c, &cobra.Command{}, []string{channel.Id, user.Id})
		s.Require().Nil(err)
		defer func() {
			appErr := s.th.App.RemoveUserFromChannel(user.Id, s.th.SystemAdminUser.Id, channel)
			s.Require().Nil(appErr)
		}()
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)

		members, appErr := s.th.App.GetChannelMembersByIds(channel.Id, []string{user.Id})
		s.Require().Nil(appErr)
		s.Require().Len(*members, 1)
		s.Require().Equal(user.Id, (*members)[0].UserId)
	})
}
