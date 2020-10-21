// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mmctl/printer"

	"github.com/spf13/cobra"
)

func (s *MmctlUnitTestSuite) TestListLdapGroupsCmd() {
	s.Run("Failure getting Ldap Groups", func() {
		printer.Clean()
		mockError := model.AppError{Id: "Mock Error"}

		s.client.
			EXPECT().
			GetLdapGroups().
			Return(nil, &model.Response{Error: &mockError}).
			Times(1)

		err := listLdapGroupsCmdF(s.client, &cobra.Command{}, []string{})
		s.Require().Equal(&mockError, err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("List several groups", func() {
		printer.Clean()
		mockList := []*model.Group{
			{DisplayName: "Group1"},
			{DisplayName: "Group2"},
			{DisplayName: "Group3"},
		}

		s.client.
			EXPECT().
			GetLdapGroups().
			Return(mockList, &model.Response{Error: nil}).
			Times(1)

		err := listLdapGroupsCmdF(s.client, &cobra.Command{}, []string{})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 3)
		for i, v := range mockList {
			s.Require().Equal(v, printer.GetLines()[i])
		}
		s.Require().Len(printer.GetErrorLines(), 0)
	})
}

func (s *MmctlUnitTestSuite) TestTeamGroupEnableCmd() {
	s.Run("Enable unexisting team", func() {
		printer.Clean()

		arg := "teamID"

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

		arg := "teamID"
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

		arg := "teamID"
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
			Return([]*model.GroupWithSchemeAdmin{}, 0, &model.Response{Error: nil}).
			Times(1)

		err := teamGroupEnableCmdF(s.client, &cobra.Command{}, []string{arg})
		s.Require().EqualError(err, "Team '"+arg+"' has no groups associated. It cannot be group-constrained")
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
	})

	s.Run("Error patching the team", func() {
		printer.Clean()

		arg := "teamID"
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
			Return([]*model.GroupWithSchemeAdmin{{}}, 1, &model.Response{Error: nil}).
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

		arg := "teamID"
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
			Return([]*model.GroupWithSchemeAdmin{{}}, 1, &model.Response{Error: nil}).
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

		teamID := "team-id"
		channelID := "channel-id"
		groupName := "group-name"

		mockTeam := model.Team{Id: teamID}
		mockChannel := model.Channel{Id: channelID}
		mockGroup := &model.GroupWithSchemeAdmin{Group: model.Group{Name: model.NewString(groupName)}}
		mockGroups := []*model.GroupWithSchemeAdmin{mockGroup}

		groupOpts := &model.GroupSearchOpts{
			PageOpts: &model.PageOpts{
				Page:    0,
				PerPage: 9999,
			},
		}

		cmdArg := teamID + ":" + channelID

		s.client.
			EXPECT().
			GetTeam(teamID, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannelByNameIncludeDeleted(channelID, teamID, "").
			Return(&mockChannel, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetGroupsByChannel(channelID, *groupOpts).
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

		teamID := "team-id"
		channelID := "channel-id"

		mockTeam := model.Team{Id: teamID}
		mockChannel := model.Channel{Id: channelID}
		mockGroups := []*model.GroupWithSchemeAdmin{
			{Group: model.Group{Name: model.NewString("group1")}},
			{Group: model.Group{Name: model.NewString("group2")}},
		}

		groupOpts := &model.GroupSearchOpts{
			PageOpts: &model.PageOpts{
				Page:    0,
				PerPage: 9999,
			},
		}

		cmdArg := teamID + ":" + channelID

		s.client.
			EXPECT().
			GetTeam(teamID, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannelByNameIncludeDeleted(channelID, teamID, "").
			Return(&mockChannel, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetGroupsByChannel(channelID, *groupOpts).
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

		teamID := "team-id"
		channelID := "channel-id"

		mockTeam := model.Team{Id: teamID}
		mockChannel := model.Channel{Id: channelID}
		mockGroups := []*model.GroupWithSchemeAdmin{}

		groupOpts := &model.GroupSearchOpts{
			PageOpts: &model.PageOpts{
				Page:    0,
				PerPage: 9999,
			},
		}

		cmdArg := teamID + ":" + channelID

		s.client.
			EXPECT().
			GetTeam(teamID, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannelByNameIncludeDeleted(channelID, teamID, "").
			Return(&mockChannel, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetGroupsByChannel(channelID, *groupOpts).
			Return(mockGroups, 0, &model.Response{Error: nil}).
			Times(1)

		err := channelGroupListCmdF(s.client, &cobra.Command{}, []string{cmdArg})
		s.Require().Nil(err)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 0)
	})

	s.Run("List groups for a nonexistent channel", func() {
		printer.Clean()

		teamID := "team-id"
		channelID := "channel-id"

		mockTeam := model.Team{Id: teamID}

		cmdArg := teamID + ":" + channelID

		s.client.
			EXPECT().
			GetTeam(teamID, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannelByNameIncludeDeleted(channelID, teamID, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannel(channelID, "").
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

		teamID := "team-id"
		channelID := "channel-id"

		cmdArg := teamID + ":" + channelID

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

		err := channelGroupListCmdF(s.client, &cobra.Command{}, []string{cmdArg})
		s.Require().NotNil(err)
		s.EqualError(err, "Unable to find channel '"+cmdArg+"'")
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 0)
	})

	s.Run("Return error when GetGroupsByChannel returns error", func() {
		printer.Clean()

		teamID := "team-id"
		channelID := "channel-id"

		mockTeam := model.Team{Id: teamID}
		mockChannel := model.Channel{Id: channelID}
		mockError := model.AppError{Message: "Mock error"}

		groupOpts := &model.GroupSearchOpts{
			PageOpts: &model.PageOpts{
				Page:    0,
				PerPage: 9999,
			},
		}

		cmdArg := teamID + ":" + channelID

		s.client.
			EXPECT().
			GetTeam(teamID, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannelByNameIncludeDeleted(channelID, teamID, "").
			Return(&mockChannel, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetGroupsByChannel(channelID, *groupOpts).
			Return(nil, 0, &model.Response{Error: &mockError}).
			Times(1)

		err := channelGroupListCmdF(s.client, &cobra.Command{}, []string{cmdArg})
		s.Require().Equal(err, &mockError)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 0)
	})

	s.Run("Return error when GetChannelByNameIncludeDeleted returns error", func() {
		printer.Clean()

		teamID := "team-id"
		channelID := "channel-id"

		mockTeam := model.Team{Id: teamID}
		mockError := model.AppError{Message: "Mock error"}

		cmdArg := teamID + ":" + channelID

		s.client.
			EXPECT().
			GetTeam(teamID, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannelByNameIncludeDeleted(channelID, teamID, "").
			Return(nil, &model.Response{Error: &mockError}).
			Times(1)

		s.client.
			EXPECT().
			GetChannel(channelID, "").
			Return(nil, &model.Response{Error: &mockError}).
			Times(1)

		err := channelGroupListCmdF(s.client, &cobra.Command{}, []string{cmdArg})
		s.EqualError(err, "Unable to find channel '"+cmdArg+"'")
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 0)
	})

	s.Run("Return error when GetTeam returns error", func() {
		printer.Clean()

		teamID := "team-id"
		channelID := "channel-id"

		mockError := model.AppError{Message: "Mock error"}

		cmdArg := teamID + ":" + channelID

		s.client.
			EXPECT().
			GetTeam(teamID, "").
			Return(nil, &model.Response{Error: &mockError}).
			Times(1)

		s.client.
			EXPECT().
			GetTeamByName(teamID, "").
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

		group1 := model.GroupWithSchemeAdmin{Group: model.Group{Id: groupID, DisplayName: "DisplayName1"}}
		group2 := model.GroupWithSchemeAdmin{Group: model.Group{Id: groupID2, DisplayName: "DisplayName2"}}

		groups := []*model.GroupWithSchemeAdmin{
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
		group1 := model.GroupWithSchemeAdmin{Group: model.Group{Id: groupID, DisplayName: "DisplayName1"}}
		group2 := model.GroupWithSchemeAdmin{Group: model.Group{Id: groupID2, DisplayName: "DisplayName2"}}

		groups := []*model.GroupWithSchemeAdmin{
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

		teamID := "teamID"
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

		teamID := "teamID"
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

		teamID := "teamID"
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

		teamID := "teamID"
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

func (s *MmctlUnitTestSuite) TestChannelGroupStatusCmd() {
	s.Run("Should fail to get group constrain status of a channel when team is not found", func() {
		printer.Clean()

		teamID := "teamID"
		channelID := "channelID"
		arg := strings.Join([]string{teamID, channelID}, ":")
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

		err := channelGroupStatusCmdF(s.client, cmd, args)

		s.Require().EqualError(err, "Unable to find channel '"+args[0]+"'")
	})

	s.Run("Should fail to get group constrain status of a channel when channel is not found", func() {
		printer.Clean()

		teamID := "teamID"
		channelID := "channelID"
		arg := strings.Join([]string{teamID, channelID}, ":")
		args := []string{arg}
		cmd := &cobra.Command{}

		team := &model.Team{Id: teamID}

		s.client.
			EXPECT().
			GetTeam(teamID, "").
			Return(team, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannelByNameIncludeDeleted(channelID, teamID, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannel(channelID, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := channelGroupStatusCmdF(s.client, cmd, args)

		s.Require().EqualError(err, "Unable to find channel '"+args[0]+"'")
	})

	s.Run("Should get valid response when channel's group constrain status is enabled", func() {
		printer.Clean()

		teamID := "teamID"
		channelID := "channelID"
		arg := strings.Join([]string{teamID, channelID}, ":")
		args := []string{arg}
		cmd := &cobra.Command{}

		team := &model.Team{Id: teamID}
		channel := &model.Channel{Id: channelID, GroupConstrained: model.NewBool(true)}

		s.client.
			EXPECT().
			GetTeam(teamID, "").
			Return(team, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannelByNameIncludeDeleted(channelID, teamID, "").
			Return(channel, &model.Response{Error: nil}).
			Times(1)

		err := channelGroupStatusCmdF(s.client, cmd, args)

		s.Require().Nil(err)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], "Enabled")
	})

	s.Run("Should get valid response when channel's group constrain status is disabled", func() {
		printer.Clean()

		teamID := "teamID"
		channelID := "channelID"
		arg := strings.Join([]string{teamID, channelID}, ":")
		args := []string{arg}
		cmd := &cobra.Command{}

		team := &model.Team{Id: teamID}
		channel := &model.Channel{Id: channelID, GroupConstrained: model.NewBool(false)}

		s.client.
			EXPECT().
			GetTeam(teamID, "").
			Return(team, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannelByNameIncludeDeleted(channelID, teamID, "").
			Return(channel, &model.Response{Error: nil}).
			Times(1)

		err := channelGroupStatusCmdF(s.client, cmd, args)

		s.Require().Nil(err)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], "Disabled")
	})

	s.Run("Should get valid response when channel's group constrain status is not present", func() {
		printer.Clean()

		teamID := "teamID"
		channelID := "channelID"
		arg := strings.Join([]string{teamID, channelID}, ":")
		args := []string{arg}
		cmd := &cobra.Command{}

		team := &model.Team{Id: teamID}
		channel := &model.Channel{Id: channelID}

		s.client.
			EXPECT().
			GetTeam(teamID, "").
			Return(team, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannelByNameIncludeDeleted(channelID, teamID, "").
			Return(channel, &model.Response{Error: nil}).
			Times(1)

		err := channelGroupStatusCmdF(s.client, cmd, args)

		s.Require().Nil(err)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Equal(printer.GetLines()[0], "Disabled")
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
		group := &model.GroupWithSchemeAdmin{Group: model.Group{Name: model.NewString("group-name")}}
		mockGroups := []*model.GroupWithSchemeAdmin{group}
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
		group := &model.GroupWithSchemeAdmin{Group: model.Group{Name: model.NewString("group-name")}}
		mockGroups := []*model.GroupWithSchemeAdmin{group}
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
		mockGroups := []*model.GroupWithSchemeAdmin{}
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

	s.Run("Enable group constraints with GetChannelByNameIncludeDeleted error", func() {
		printer.Clean()

		teamArg := "team-id"
		mockTeam := model.Team{Id: teamArg}
		channelPart := "channel-id"
		mockChannel := model.Channel{Id: channelPart}
		channelArg := teamArg + ":" + channelPart
		group := &model.GroupWithSchemeAdmin{Group: model.Group{Name: model.NewString("group-name")}}
		mockGroups := []*model.GroupWithSchemeAdmin{group}
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
			Return(nil, &model.Response{Error: &mockError}).
			Times(1)

		s.client.
			EXPECT().
			GetChannel(channelPart, "").
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
}

func (s *MmctlUnitTestSuite) TestChannelGroupDisableCmdF() {
	s.Run("Disable group constraints with existing team and channel", func() {
		printer.Clean()

		teamArg := "team-id"
		mockTeam := model.Team{Id: teamArg}
		channelPart := "channel-id"
		mockChannel := model.Channel{Id: channelPart}
		channelArg := strings.Join([]string{teamArg, channelPart}, ":")

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
			PatchChannel(channelPart, &model.ChannelPatch{GroupConstrained: model.NewBool(false)}).
			Return(&mockChannel, &model.Response{Error: nil}).
			Times(1)

		err := channelGroupDisableCmdF(s.client, &cobra.Command{}, []string{channelArg})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Disable group constraints with nonexistent team", func() {
		printer.Clean()

		teamArg := "team-id"
		channelPart := "channel-id"
		channelArg := strings.Join([]string{teamArg, channelPart}, ":")

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

		err := channelGroupDisableCmdF(s.client, &cobra.Command{}, []string{channelArg})
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.EqualError(err, "Unable to find channel '"+channelArg+"'")
	})

	s.Run("Disable group constraints with nonexistent channel", func() {
		printer.Clean()

		teamArg := "team-id"
		mockTeam := model.Team{Id: teamArg}
		channelPart := "channel-id"
		channelArg := strings.Join([]string{teamArg, channelPart}, ":")

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

		err := channelGroupDisableCmdF(s.client, &cobra.Command{}, []string{channelArg})
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.EqualError(err, "Unable to find channel '"+channelArg+"'")
	})

	s.Run("Disable group constraints with GetTeam error", func() {
		printer.Clean()

		teamArg := "team-id"
		mockTeam := model.Team{Id: teamArg}
		channelPart := "channel-id"
		mockChannel := model.Channel{Id: channelPart}
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
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetChannelByNameIncludeDeleted(channelPart, teamArg, "").
			Return(&mockChannel, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			PatchChannel(channelPart, &model.ChannelPatch{GroupConstrained: model.NewBool(false)}).
			Return(&mockChannel, &model.Response{Error: nil}).
			Times(1)

		err := channelGroupDisableCmdF(s.client, &cobra.Command{}, []string{channelArg})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Disable group constraints with GetTeamByName error", func() {
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

		err := channelGroupDisableCmdF(s.client, &cobra.Command{}, []string{channelArg})
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.EqualError(err, "Unable to find channel '"+channelArg+"'")
	})

	s.Run("Disable group constraints with GetChannelByNameIncludeDeleted error", func() {
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

		err := channelGroupDisableCmdF(s.client, &cobra.Command{}, []string{channelArg})
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.EqualError(err, "Unable to find channel '"+channelArg+"'")
	})

	s.Run("Disable group constraints with PatchChannel error", func() {
		printer.Clean()

		teamArg := "team-id"
		mockTeam := model.Team{Id: teamArg}
		channelPart := "channel-id"
		mockChannel := model.Channel{Id: channelPart}
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
			Return(&mockChannel, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			PatchChannel(channelPart, &model.ChannelPatch{GroupConstrained: model.NewBool(false)}).
			Return(nil, &model.Response{Error: &mockError}).
			Times(1)

		err := channelGroupDisableCmdF(s.client, &cobra.Command{}, []string{channelArg})
		s.Require().NotNil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.EqualError(err, mockError.Error())
	})
}
