package commands

import (
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-server/v5/model"
)

func (s *MmctlUnitTestSuite) TestGetTeamArgs() {
	s.Run("team not found", func() {
		notFoundTeam := "notfoundteam"
		notFoundErr := &model.AppError{Message: "team not found", StatusCode: http.StatusNotFound}

		s.client.
			EXPECT().
			GetTeam(notFoundTeam, "").
			Return(nil, &model.Response{Error: notFoundErr}).
			Times(1)
		s.client.
			EXPECT().
			GetTeamByName(notFoundTeam, "").
			Return(nil, &model.Response{Error: notFoundErr}).
			Times(1)

		teams, err := getTeamsFromArgs(s.client, []string{notFoundTeam})
		s.Require().Len(teams, 0)
		s.Require().NotNil(err)
		s.Require().Equal(fmt.Sprintf("1 error occurred:\n\t* team %s not found\n\n", notFoundTeam), err.Error())
	})
	s.Run("bad request", func() {
		badRequestTeam := "badrequest"
		badRequestErr := &model.AppError{Message: "team bad request", StatusCode: http.StatusBadRequest}

		s.client.
			EXPECT().
			GetTeam(badRequestTeam, "").
			Return(nil, &model.Response{Error: badRequestErr}).
			Times(1)
		s.client.
			EXPECT().
			GetTeamByName(badRequestTeam, "").
			Return(nil, &model.Response{Error: badRequestErr}).
			Times(1)

		teams, err := getTeamsFromArgs(s.client, []string{badRequestTeam})
		s.Require().Len(teams, 0)
		s.Require().NotNil(err)
		s.Require().Equal(fmt.Sprintf("1 error occurred:\n\t* team %s not found\n\n", badRequestTeam), err.Error())
	})
	s.Run("forbidden", func() {
		forbidden := "forbidden"
		forbiddenErr := &model.AppError{Message: "team forbidden", StatusCode: http.StatusForbidden}

		s.client.
			EXPECT().
			GetTeam(forbidden, "").
			Return(nil, &model.Response{Error: forbiddenErr}).
			Times(1)

		teams, err := getTeamsFromArgs(s.client, []string{forbidden})
		s.Require().Len(teams, 0)
		s.Require().NotNil(err)
		s.Require().Equal("1 error occurred:\n\t* : team forbidden, \n\n", err.Error())
	})
	s.Run("internal server error", func() {
		errTeam := "internalServerError"
		internalServerErrorErr := &model.AppError{Message: "team internalServerError", StatusCode: http.StatusInternalServerError}

		s.client.
			EXPECT().
			GetTeam(errTeam, "").
			Return(nil, &model.Response{Error: internalServerErrorErr}).
			Times(1)

		teams, err := getTeamsFromArgs(s.client, []string{errTeam})
		s.Require().Len(teams, 0)
		s.Require().NotNil(err)
		s.Require().Equal("1 error occurred:\n\t* : team internalServerError, \n\n", err.Error())
	})
	s.Run("success", func() {
		successID := "success@success.com"
		successTeam := &model.Team{Id: successID}

		s.client.
			EXPECT().
			GetTeam(successID, "").
			Return(successTeam, nil).
			Times(1)

		teams, summary := getTeamsFromArgs(s.client, []string{successID})
		s.Require().Nil(summary)
		s.Require().Len(teams, 1)
		s.Require().Equal(successTeam, teams[0])
	})
}
