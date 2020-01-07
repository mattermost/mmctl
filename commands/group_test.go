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
