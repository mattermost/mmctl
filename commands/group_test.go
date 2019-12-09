package commands

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mmctl/printer"

	"github.com/spf13/cobra"
)

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
