package commands

import (
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
			&model.Group{DisplayName: "Group1"},
			&model.Group{DisplayName: "Group2"},
			&model.Group{DisplayName: "Group3"},
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
