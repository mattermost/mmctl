package commands

import (
	"strings"

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

func (s *MmctlUnitTestSuite) TestChannelGroupStatusCmd() {
	s.Run("Should fail to get group constrain status of a channel when team is not found", func() {
		printer.Clean()

		teamID := "teamId"
		channelID := "channelId"
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

		teamID := "teamId"
		channelID := "channelId"
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

		teamID := "teamId"
		channelID := "channelId"
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

		teamID := "teamId"
		channelID := "channelId"
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

		teamID := "teamId"
		channelID := "channelId"
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
