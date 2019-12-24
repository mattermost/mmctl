package commands

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mmctl/printer"
	"github.com/spf13/cobra"
)

func (s *MmctlUnitTestSuite) TestChannelGroupEnableCmdF() {
	s.Run("Enable group constrains with existing team and channel", func() {
		printer.Clean()
		teamArg := "team-id"
		mockTeam := model.Team{Id: teamArg}
		channelPart := "channel-id"
		mockChannel := model.Channel{Id: channelPart}
		channelArg := teamArg + ":" + channelPart
		group := &model.Group{Name: "group-name"}
		mockGroups := []*model.Group{group}
		groupOpts := &model.GroupSearchOpts{
			PageOpts: &model.PageOpts{
				Page:    0,
				PerPage: 10,
			},
		}

		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannelByNameIncludeDeleted(channelPart, teamArg, "").
			Return(&mockChannel, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetGroupsByChannel(channelPart, *groupOpts).
			Return(mockGroups, 0, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			PatchChannel(channelPart, &model.ChannelPatch{GroupConstrained: model.NewBool(true)}).
			Return(&mockChannel, &model.Response{Error: nil}).
			Times(1)

		err := channelGroupEnableCmdF(s.client, &cobra.Command{}, []string{channelArg})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Enable group constrains with GetTeam error", func() {
		printer.Clean()
		teamArg := "team-id"
		channelPart := "channel-id"
		channelArg := teamArg + ":" + channelPart
		mockError := model.AppError{Id: "Mock Error"}

		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(nil, &model.Response{Error: &mockError}).
			Times(1)

		s.client.
			EXPECT().
			GetTeamByName(teamArg, "").
			Return(nil, &model.Response{Error: &mockError}).
			Times(1)

		err := channelGroupEnableCmdF(s.client, &cobra.Command{}, []string{channelArg})
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.EqualError(err, "Unable to find channel '"+channelArg+"'")
	})

	s.Run("Enable group constrains with GetChannelByNameIncludeDeleted error", func() {
		printer.Clean()
		teamArg := "team-id"
		mockTeam := model.Team{Id: teamArg}
		channelPart := "channel-id"
		channelArg := teamArg + ":" + channelPart
		mockError := model.AppError{Id: "Mock Error"}

		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannelByNameIncludeDeleted(channelPart, teamArg, "").
			Return(nil, &model.Response{Error: &mockError}).
			Times(1)

		s.client.
			EXPECT().
			GetChannel(channelPart, "").
			Return(nil, &model.Response{Error: &mockError}).
			Times(1)

		err := channelGroupEnableCmdF(s.client, &cobra.Command{}, []string{channelArg})
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.EqualError(err, "Unable to find channel '"+channelArg+"'")
	})

	s.Run("Enable group constrains with GetGroupsByChannel error", func() {
		printer.Clean()
		teamArg := "team-id"
		mockTeam := model.Team{Id: teamArg}
		channelPart := "channel-id"
		mockChannel := model.Channel{Id: channelPart}
		channelArg := teamArg + ":" + channelPart
		mockError := model.AppError{Id: "Mock Error"}
		groupOpts := &model.GroupSearchOpts{
			PageOpts: &model.PageOpts{
				Page:    0,
				PerPage: 10,
			},
		}

		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannelByNameIncludeDeleted(channelPart, teamArg, "").
			Return(&mockChannel, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetGroupsByChannel(channelPart, *groupOpts).
			Return(nil, 0, &model.Response{Error: &mockError}).
			Times(1)

		err := channelGroupEnableCmdF(s.client, &cobra.Command{}, []string{channelArg})
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.EqualError(err, mockError.Error())
	})

	s.Run("Enable group constrains with PatchChannel error", func() {
		printer.Clean()
		teamArg := "team-id"
		mockTeam := model.Team{Id: teamArg}
		channelPart := "channel-id"
		mockChannel := model.Channel{Id: channelPart}
		channelArg := teamArg + ":" + channelPart
		group := &model.Group{Name: "group-name"}
		mockGroups := []*model.Group{group}
		mockError := model.AppError{Id: "Mock Error"}
		groupOpts := &model.GroupSearchOpts{
			PageOpts: &model.PageOpts{
				Page:    0,
				PerPage: 10,
			},
		}

		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannelByNameIncludeDeleted(channelPart, teamArg, "").
			Return(&mockChannel, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetGroupsByChannel(channelPart, *groupOpts).
			Return(mockGroups, 0, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			PatchChannel(channelPart, &model.ChannelPatch{GroupConstrained: model.NewBool(true)}).
			Return(nil, &model.Response{Error: &mockError}).
			Times(1)

		err := channelGroupEnableCmdF(s.client, &cobra.Command{}, []string{channelArg})
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.EqualError(err, mockError.Error())
	})

	s.Run("Enable group constrains with no associated groups", func() {
		printer.Clean()
		teamArg := "team-id"
		mockTeam := model.Team{Id: teamArg}
		channelPart := "channel-id"
		mockChannel := model.Channel{Id: channelPart}
		channelArg := teamArg + ":" + channelPart
		mockGroups := []*model.Group{}
		groupOpts := &model.GroupSearchOpts{
			PageOpts: &model.PageOpts{
				Page:    0,
				PerPage: 10,
			},
		}

		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannelByNameIncludeDeleted(channelPart, teamArg, "").
			Return(&mockChannel, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetGroupsByChannel(channelPart, *groupOpts).
			Return(mockGroups, 0, &model.Response{Error: nil}).
			Times(1)

		err := channelGroupEnableCmdF(s.client, &cobra.Command{}, []string{channelArg})
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.EqualError(err, "Channel '"+channelArg+"' has no groups associated. It cannot be group-constrained")
	})

	s.Run("Enable group constrains with nonexistant team", func() {
		printer.Clean()
		teamArg := "team-id"
		channelPart := "channel-id"
		channelArg := teamArg + ":" + channelPart

		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeamByName(teamArg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := channelGroupEnableCmdF(s.client, &cobra.Command{}, []string{channelArg})
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.EqualError(err, "Unable to find channel '"+channelArg+"'")
	})

	s.Run("Enable group constrains with nonexistant channel", func() {
		printer.Clean()
		teamArg := "team-id"
		mockTeam := model.Team{Id: teamArg}
		channelPart := "channel-id"
		channelArg := teamArg + ":" + channelPart

		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannelByNameIncludeDeleted(channelPart, teamArg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannel(channelPart, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := channelGroupEnableCmdF(s.client, &cobra.Command{}, []string{channelArg})
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.EqualError(err, "Unable to find channel '"+channelArg+"'")
	})
}
