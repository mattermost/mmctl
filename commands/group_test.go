// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

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

func (s *MmctlUnitTestSuite) TestTeamGroupDisableCmd() {
	s.Run("Disable existing team", func() {
		printer.Clean()
		teamArg := "example-team-id"
		mockTeam := model.Team{Id: teamArg}
		teamPatch := model.TeamPatch{GroupConstrained: model.NewBool(false)}

		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			PatchTeam(teamArg, &teamPatch).
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := teamGroupDisableCmdF(s.client, &cobra.Command{}, []string{teamArg})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 0)
	})

	s.Run("Disable nonexisting team", func() {
		printer.Clean()
		teamArg := "example-team-id"

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

		err := teamGroupDisableCmdF(s.client, &cobra.Command{}, []string{teamArg})
		s.Require().NotNil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
		s.EqualError(err, "Unable to find team '"+teamArg+"'")
	})

	s.Run("Error response from PatchTeam", func() {
		printer.Clean()
		teamArg := "example-team-id"
		mockTeam := model.Team{Id: teamArg}
		teamPatch := model.TeamPatch{GroupConstrained: model.NewBool(false)}
		errMessage := "PatchTeam Error"
		mockError := &model.AppError{Message: errMessage}

		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			PatchTeam(teamArg, &teamPatch).
			Return(nil, &model.Response{Error: mockError}).
			Times(1)

		err := teamGroupDisableCmdF(s.client, &cobra.Command{}, []string{teamArg})
		s.Require().NotNil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
		s.EqualError(err, mockError.Error())
	})
}

func (s *MmctlUnitTestSuite) TestChannelGroupListCmd() {
	s.Run("List groups for existing channel and team, when a single group exists", func() {
		printer.Clean()

		teamId := "team-id"
		channelId := "channel-id"
		groupName := "group-name"

		mockTeam := model.Team{Id: teamId}
		mockChannel := model.Channel{Id: channelId}
		mockGroup := &model.Group{Name: groupName}
		mockGroups := []*model.Group{mockGroup}

		groupOpts := &model.GroupSearchOpts{
			PageOpts: &model.PageOpts{
				Page:    0,
				PerPage: 9999,
			},
		}

		cmdArg := teamId + ":" + channelId

		s.client.
			EXPECT().
			GetTeam(teamId, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannelByNameIncludeDeleted(channelId, teamId, "").
			Return(&mockChannel, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetGroupsByChannel(channelId, *groupOpts).
			Return(mockGroups, 0, &model.Response{Error: nil}).
			Times(1)

		err := channelGroupListCmdF(s.client, &cobra.Command{}, []string{cmdArg})
		s.Require().Nil(err)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], mockGroup)
	})

	s.Run("List groups for existing channel and team, when multiple groups exist", func() {
		printer.Clean()

		teamId := "team-id"
		channelId := "channel-id"

		mockTeam := model.Team{Id: teamId}
		mockChannel := model.Channel{Id: channelId}
		mockGroups := []*model.Group{
			&model.Group{Name: "group1"},
			&model.Group{Name: "group2"},
		}

		groupOpts := &model.GroupSearchOpts{
			PageOpts: &model.PageOpts{
				Page:    0,
				PerPage: 9999,
			},
		}

		cmdArg := teamId + ":" + channelId

		s.client.
			EXPECT().
			GetTeam(teamId, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannelByNameIncludeDeleted(channelId, teamId, "").
			Return(&mockChannel, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetGroupsByChannel(channelId, *groupOpts).
			Return(mockGroups, 0, &model.Response{Error: nil}).
			Times(1)

		err := channelGroupListCmdF(s.client, &cobra.Command{}, []string{cmdArg})
		s.Require().Nil(err)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 2)
		s.Require().Equal(printer.GetLines()[0], mockGroups[0])
		s.Require().Equal(printer.GetLines()[1], mockGroups[1])
	})

	s.Run("List groups for existing channel and team, when no groups exist", func() {
		printer.Clean()

		teamId := "team-id"
		channelId := "channel-id"

		mockTeam := model.Team{Id: teamId}
		mockChannel := model.Channel{Id: channelId}
		mockGroups := []*model.Group{}

		groupOpts := &model.GroupSearchOpts{
			PageOpts: &model.PageOpts{
				Page:    0,
				PerPage: 9999,
			},
		}

		cmdArg := teamId + ":" + channelId

		s.client.
			EXPECT().
			GetTeam(teamId, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannelByNameIncludeDeleted(channelId, teamId, "").
			Return(&mockChannel, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetGroupsByChannel(channelId, *groupOpts).
			Return(mockGroups, 0, &model.Response{Error: nil}).
			Times(1)

		err := channelGroupListCmdF(s.client, &cobra.Command{}, []string{cmdArg})
		s.Require().Nil(err)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 0)
	})

	s.Run("List groups for a nonexistent channel", func() {
		printer.Clean()

		teamId := "team-id"
		channelId := "channel-id"

		mockTeam := model.Team{Id: teamId}

		cmdArg := teamId + ":" + channelId

		s.client.
			EXPECT().
			GetTeam(teamId, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannelByNameIncludeDeleted(channelId, teamId, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannel(channelId, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := channelGroupListCmdF(s.client, &cobra.Command{}, []string{cmdArg})
		s.Require().NotNil(err)
		s.EqualError(err, "Unable to find channel '"+cmdArg+"'")
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 0)
	})

	s.Run("List groups for a nonexistent team", func() {
		printer.Clean()

		teamId := "team-id"
		channelId := "channel-id"

		cmdArg := teamId + ":" + channelId

		s.client.
			EXPECT().
			GetTeam(teamId, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeamByName(teamId, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := channelGroupListCmdF(s.client, &cobra.Command{}, []string{cmdArg})
		s.Require().NotNil(err)
		s.EqualError(err, "Unable to find channel '"+cmdArg+"'")
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 0)
	})

	s.Run("Return error when GetGroupsByChannel returns error", func() {
		printer.Clean()

		teamId := "team-id"
		channelId := "channel-id"

		mockTeam := model.Team{Id: teamId}
		mockChannel := model.Channel{Id: channelId}
		mockError := model.AppError{Message: "Mock error"}

		groupOpts := &model.GroupSearchOpts{
			PageOpts: &model.PageOpts{
				Page:    0,
				PerPage: 9999,
			},
		}

		cmdArg := teamId + ":" + channelId

		s.client.
			EXPECT().
			GetTeam(teamId, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannelByNameIncludeDeleted(channelId, teamId, "").
			Return(&mockChannel, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetGroupsByChannel(channelId, *groupOpts).
			Return(nil, 0, &model.Response{Error: &mockError}).
			Times(1)

		err := channelGroupListCmdF(s.client, &cobra.Command{}, []string{cmdArg})
		s.Require().Equal(err, &mockError)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 0)
	})

	s.Run("Return error when GetChannelByNameIncludeDeleted returns error", func() {
		printer.Clean()

		teamId := "team-id"
		channelId := "channel-id"

		mockTeam := model.Team{Id: teamId}
		mockError := model.AppError{Message: "Mock error"}

		cmdArg := teamId + ":" + channelId

		s.client.
			EXPECT().
			GetTeam(teamId, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannelByNameIncludeDeleted(channelId, teamId, "").
			Return(nil, &model.Response{Error: &mockError}).
			Times(1)

		s.client.
			EXPECT().
			GetChannel(channelId, "").
			Return(nil, &model.Response{Error: &mockError}).
			Times(1)

		err := channelGroupListCmdF(s.client, &cobra.Command{}, []string{cmdArg})
		s.EqualError(err, "Unable to find channel '"+cmdArg+"'")
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 0)
	})

	s.Run("Return error when GetTeam returns error", func() {
		printer.Clean()

		teamId := "team-id"
		channelId := "channel-id"

		mockError := model.AppError{Message: "Mock error"}

		cmdArg := teamId + ":" + channelId

		s.client.
			EXPECT().
			GetTeam(teamId, "").
			Return(nil, &model.Response{Error: &mockError}).
			Times(1)

		s.client.
			EXPECT().
			GetTeamByName(teamId, "").
			Return(nil, &model.Response{Error: &mockError}).
			Times(1)

		err := channelGroupListCmdF(s.client, &cobra.Command{}, []string{cmdArg})
		s.EqualError(err, "Unable to find channel '"+cmdArg+"'")
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 0)
	})
}

func (s *MmctlUnitTestSuite) TestTeamGroupListCmd() {
	s.Run("Team group list returns error when passing a nonexistent team", func() {
		printer.Clean()

		s.client.
			EXPECT().
			GetTeam("team1", "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeamByName("team1", "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		cmd := &cobra.Command{}
		err := teamGroupListCmdF(s.client, cmd, []string{"team1"})

		s.Require().NotNil(err)
		s.Require().Equal(err.Error(), "Unable to find team 'team1'")
	})

	s.Run("Team group list return error when GetGroupsByTeam returns error", func() {
		printer.Clean()
		groupID := "group1"
		groupID2 := "group2"
		mockError := &model.AppError{Message: "Get groups by team error"}
		group1 := model.Group{Id: groupID, DisplayName: "DisplayName1"}
		group2 := model.Group{Id: groupID2, DisplayName: "DisplayName2"}

		groups := []*model.Group{
			&group1,
			&group2,
		}

		mockTeam := model.Team{Id: "team1"}
		groupOpts := model.GroupSearchOpts{
			PageOpts: &model.PageOpts{
				Page:    0,
				PerPage: 9999,
			},
		}

		s.client.
			EXPECT().
			GetTeam("team1", "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetGroupsByTeam("team1", groupOpts).
			Return(groups, 2, &model.Response{Error: mockError}).
			Times(1)

		cmd := &cobra.Command{}
		err := teamGroupListCmdF(s.client, cmd, []string{"team1"})

		s.Require().NotNil(err)
		s.Require().Equal(err, mockError)
	})

	s.Run("Team group list should print group in console on success", func() {
		printer.Clean()
		groupID := "group1"
		groupID2 := "group2"
		group1 := model.Group{Id: groupID, DisplayName: "DisplayName1"}
		group2 := model.Group{Id: groupID2, DisplayName: "DisplayName2"}

		groups := []*model.Group{
			&group1,
			&group2,
		}

		mockTeam := model.Team{Id: "team1"}
		groupOpts := model.GroupSearchOpts{
			PageOpts: &model.PageOpts{
				Page:    0,
				PerPage: 9999,
			},
		}

		s.client.
			EXPECT().
			GetTeam("team1", "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetGroupsByTeam("team1", groupOpts).
			Return(groups, 2, &model.Response{Error: nil}).
			Times(1)

		cmd := &cobra.Command{}
		err := teamGroupListCmdF(s.client, cmd, []string{"team1"})

		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 2)
		s.Require().Equal(printer.GetLines()[0], &group1)
		s.Require().Equal(printer.GetLines()[1], &group2)
	})
}

func (s *MmctlUnitTestSuite) TestTeamGroupStatusCmd() {
	s.Run("Should fail when team is not found", func() {
		printer.Clean()

		teamID := "teamId"
		arg := teamID
		args := []string{arg}
		cmd := &cobra.Command{}

		s.client.
			EXPECT().
			GetTeam(teamID, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeamByName(teamID, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := teamGroupStatusCmdF(s.client, cmd, args)

		s.Require().EqualError(err, "Unable to find team '"+args[0]+"'")
	})

	s.Run("Should show valid response when group constraints status for a team is not present", func() {
		printer.Clean()

		teamID := "teamId"
		arg := teamID
		args := []string{arg}
		cmd := &cobra.Command{}
		team := &model.Team{Id: teamID}

		s.client.
			EXPECT().
			GetTeam(teamID, "").
			Return(team, &model.Response{Error: nil}).
			Times(1)

		err := teamGroupStatusCmdF(s.client, cmd, args)

		s.Require().Nil(err)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], "Disabled")
	})

	s.Run("Should show valid response when group constraints status for a team is enabled", func() {
		printer.Clean()

		teamID := "teamId"
		arg := teamID
		args := []string{arg}
		cmd := &cobra.Command{}
		team := &model.Team{Id: teamID, GroupConstrained: model.NewBool(true)}

		s.client.
			EXPECT().
			GetTeam(teamID, "").
			Return(team, &model.Response{Error: nil}).
			Times(1)

		err := teamGroupStatusCmdF(s.client, cmd, args)

		s.Require().Nil(err)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], "Enabled")
	})

	s.Run("Should show valid response when group constraints status for a team is disabled", func() {
		printer.Clean()

		teamID := "teamId"
		arg := teamID
		args := []string{arg}
		cmd := &cobra.Command{}
		team := &model.Team{Id: teamID, GroupConstrained: model.NewBool(false)}

		s.client.
			EXPECT().
			GetTeam(teamID, "").
			Return(team, &model.Response{Error: nil}).
			Times(1)

		err := teamGroupStatusCmdF(s.client, cmd, args)

		s.Require().Nil(err)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], "Disabled")
	})
}
