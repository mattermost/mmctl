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
		//mockTeam := model.Team{Id: teamArg}

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
		mockGroupList := []*model.Group{}
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
			Return(mockGroupList, 0, &model.Response{Error: nil}).
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
		mockGroupList := []*model.Group{&model.Group{}}
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
			Return(mockGroupList, 1, &model.Response{Error: nil}).
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
		mockGroupList := []*model.Group{&model.Group{}}
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
			Return(mockGroupList, 1, &model.Response{Error: nil}).
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
