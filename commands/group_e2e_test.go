// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"github.com/mattermost/mattermost-server/v5/api4"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/spf13/cobra"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"
)

func (s *MmctlE2ETestSuite) TestChannelGroupEnableCmd() {
	s.SetupEnterpriseTestHelper().InitBasic()

	channelName := api4.GenerateTestChannelName()
	channel, appErr := s.th.App.CreateChannel(s.th.Context, &model.Channel{
		TeamId:      s.th.BasicTeam.Id,
		Name:        channelName,
		DisplayName: "dn_" + channelName,
		Type:        model.CHANNEL_OPEN,
	}, false)
	s.Require().Nil(appErr)
	defer func() {
		err := s.th.App.DeleteChannel(s.th.Context, channel, "")
		s.Require().Nil(err)
	}()

	id := model.NewId()
	group, appErr := s.th.App.CreateGroup(&model.Group{
		DisplayName: "dn_" + id,
		Name:        model.NewString("name" + id),
		Source:      model.GroupSourceLdap,
		Description: "description_" + id,
		RemoteId:    model.NewId(),
	})
	s.Require().Nil(appErr)
	defer func() {
		_, err := s.th.App.DeleteGroup(group.Id)
		s.Require().Nil(err)
	}()

	_, appErr = s.th.App.UpsertGroupSyncable(&model.GroupSyncable{
		GroupId:    group.Id,
		SyncableId: channel.Id,
		Type:       model.GroupSyncableTypeChannel,
	})
	s.Require().Nil(appErr)
	defer func() {
		_, err := s.th.App.DeleteGroupSyncable(group.Id, channel.Id, model.GroupSyncableTypeChannel)
		s.Require().Nil(err)
	}()

	s.Run("Should not allow regular user to enable group for channel", func() {
		printer.Clean()

		err := channelGroupEnableCmdF(s.th.Client, &cobra.Command{}, []string{s.th.BasicTeam.Name + ":" + channelName})
		s.Require().Error(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.RunForSystemAdminAndLocal("Should enable group sync for the channel", func(c client.Client) {
		printer.Clean()

		err := channelGroupEnableCmdF(c, &cobra.Command{}, []string{s.th.BasicTeam.Name + ":" + channelName})
		s.Require().NoError(err)

		channel.GroupConstrained = model.NewBool(false)
		defer func() {
			_, err := s.th.App.UpdateChannel(channel)
			s.Require().Nil(err)
		}()

		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)

		ch, appErr := s.th.App.GetChannel(channel.Id)
		s.Require().Nil(appErr)
		s.Require().True(ch.IsGroupConstrained())
	})
}

func (s *MmctlE2ETestSuite) TestChannelGroupDisableCmd() {
	s.SetupEnterpriseTestHelper().InitBasic()

	channelName := api4.GenerateTestChannelName()
	channel, appErr := s.th.App.CreateChannel(s.th.Context, &model.Channel{
		TeamId:      s.th.BasicTeam.Id,
		Name:        channelName,
		DisplayName: "dn_" + channelName,
		Type:        model.CHANNEL_OPEN,
	}, false)
	s.Require().Nil(appErr)
	defer func() {
		err := s.th.App.DeleteChannel(s.th.Context, channel, "")
		s.Require().Nil(err)
	}()

	id := model.NewId()
	group, appErr := s.th.App.CreateGroup(&model.Group{
		DisplayName: "dn_" + id,
		Name:        model.NewString("name" + id),
		Source:      model.GroupSourceLdap,
		Description: "description_" + id,
		RemoteId:    model.NewId(),
	})
	s.Require().Nil(appErr)
	defer func() {
		_, err := s.th.App.DeleteGroup(group.Id)
		s.Require().Nil(err)
	}()

	_, appErr = s.th.App.UpsertGroupSyncable(&model.GroupSyncable{
		GroupId:    group.Id,
		SyncableId: channel.Id,
		Type:       model.GroupSyncableTypeChannel,
	})
	s.Require().Nil(appErr)
	defer func() {
		_, err := s.th.App.DeleteGroupSyncable(group.Id, channel.Id, model.GroupSyncableTypeChannel)
		s.Require().Nil(err)
	}()

	channel.GroupConstrained = model.NewBool(true)
	defer func() {
		_, err := s.th.App.UpdateChannel(channel)
		s.Require().Nil(err)
	}()

	s.Run("Should not allow regular user to disable group for channel", func() {
		printer.Clean()

		err := channelGroupEnableCmdF(s.th.Client, &cobra.Command{}, []string{s.th.BasicTeam.Name + ":" + channelName})
		s.Require().Error(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.RunForSystemAdminAndLocal("Should disable group sync for the channel", func(c client.Client) {
		printer.Clean()

		err := channelGroupDisableCmdF(c, &cobra.Command{}, []string{s.th.BasicTeam.Name + ":" + channelName})
		s.Require().NoError(err)

		channel.GroupConstrained = model.NewBool(true)
		defer func() {
			_, err := s.th.App.UpdateChannel(channel)
			s.Require().Nil(err)
		}()

		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)

		ch, appErr := s.th.App.GetChannel(channel.Id)
		s.Require().Nil(appErr)
		s.Require().False(ch.IsGroupConstrained())
	})
}
