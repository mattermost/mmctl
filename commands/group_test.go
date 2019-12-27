package commands

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mmctl/printer"
	"github.com/spf13/cobra"
)

func (s *MmctlUnitTestSuite) TestTeamGroupEnableCmd() {
	s.Run("Enable unexisting team", func() {
		printer.Clean()

		arg := "teamId"

		s.client.
			EXPECT().
			GetTeam(arg, "").
			Return(nil, &model.Response{Error: &model.AppError{}}).
			Times(1)

		s.client.
			EXPECT().
			GetTeamByName(arg, "").
			Return(nil, &model.Response{Error: &model.AppError{}}).
			Times(1)

		err := teamGroupEnableCmdF(s.client, &cobra.Command{}, []string{arg})
		s.Require().EqualError(err, "Unable to find team '"+arg+"'")
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
	})

	s.Run("Error while getting the team groups", func() {
		printer.Clean()

		arg := "teamId"
		mockTeam := model.Team{Id: arg}
		mockError := model.AppError{Message: "Mock error"}
		groupOpts := model.GroupSearchOpts{
			PageOpts: &model.PageOpts{
				Page:    0,
				PerPage: 10,
			},
		}

		s.client.
			EXPECT().
			GetTeam(arg, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetGroupsByTeam(mockTeam.Id, groupOpts).
			Return(nil, 0, &model.Response{Error: &mockError}).
			Times(1)

		err := teamGroupEnableCmdF(s.client, &cobra.Command{}, []string{arg})
		s.Require().Equal(&mockError, err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
	})

	s.Run("No groups on team", func() {
		printer.Clean()

		arg := "teamId"
		mockTeam := model.Team{Id: arg}
		groupOpts := model.GroupSearchOpts{
			PageOpts: &model.PageOpts{
				Page:    0,
				PerPage: 10,
			},
		}

		s.client.
			EXPECT().
			GetTeam(arg, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetGroupsByTeam(mockTeam.Id, groupOpts).
			Return([]*model.Group{}, 0, &model.Response{Error: nil}).
			Times(1)

		err := teamGroupEnableCmdF(s.client, &cobra.Command{}, []string{arg})
		s.Require().EqualError(err, "Team '"+arg+"' has no groups associated. It cannot be group-constrained")
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
	})

	s.Run("Error patching the team", func() {
		printer.Clean()

		arg := "teamId"
		mockTeam := model.Team{Id: arg}
		mockError := model.AppError{Message: "Mock error"}
		groupOpts := model.GroupSearchOpts{
			PageOpts: &model.PageOpts{
				Page:    0,
				PerPage: 10,
			},
		}
		teamPatch := model.TeamPatch{GroupConstrained: model.NewBool(true)}

		s.client.
			EXPECT().
			GetTeam(arg, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetGroupsByTeam(mockTeam.Id, groupOpts).
			Return([]*model.Group{&model.Group{}}, 1, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			PatchTeam(mockTeam.Id, &teamPatch).
			Return(nil, &model.Response{Error: &mockError}).
			Times(1)

		err := teamGroupEnableCmdF(s.client, &cobra.Command{}, []string{arg})
		s.Require().Equal(&mockError, err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
	})

	s.Run("Successfully enable group", func() {
		printer.Clean()

		arg := "teamId"
		mockTeam := model.Team{Id: arg}
		groupOpts := model.GroupSearchOpts{
			PageOpts: &model.PageOpts{
				Page:    0,
				PerPage: 10,
			},
		}
		teamPatch := model.TeamPatch{GroupConstrained: model.NewBool(true)}

		s.client.
			EXPECT().
			GetTeam(arg, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetGroupsByTeam(mockTeam.Id, groupOpts).
			Return([]*model.Group{&model.Group{}}, 1, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			PatchTeam(mockTeam.Id, &teamPatch).
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		err := teamGroupEnableCmdF(s.client, &cobra.Command{}, []string{arg})
		s.Require().NoError(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
	})
}

func (s *MmctlUnitTestSuite) TestChannelGroupEnableCmdF() {
	s.Run("Enable group constraints with existing team and channel", func() {
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

	s.Run("Enable group constraints with GetTeam error", func() {
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

	s.Run("Enable group constraints with GetChannelByNameIncludeDeleted error", func() {
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

	s.Run("Enable group constraints with GetGroupsByChannel error", func() {
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

	s.Run("Enable group constraints with PatchChannel error", func() {
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

	s.Run("Enable group constraints with no associated groups", func() {
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

	s.Run("Enable group constraints with nonexistent team", func() {
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

	s.Run("Enable group constraints with nonexistent channel", func() {
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
