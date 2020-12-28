// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mmctl/client"
)

func getTeamsFromTeamArgs(c client.Client, teamArgs []string) []*model.Team {
	teams := make([]*model.Team, 0, len(teamArgs))
	for _, teamArg := range teamArgs {
		team := getTeamFromTeamArg(c, teamArg)
		teams = append(teams, team)
	}
	return teams
}

func getTeamFromTeamArg(c client.Client, teamArg string) *model.Team {
	if checkDots(teamArg) || checkSlash(teamArg) {
		return nil
	}

	var team *model.Team
	team, _ = c.GetTeam(teamArg, "")

	if team == nil {
		team, _ = c.GetTeamByName(teamArg, "")
	}
	return team
}

func getTeamsFromArgs(c client.Client, teamArgs []string) ([]*model.Team, *FindEntitySummary) {
	var (
		teams  []*model.Team
		errors []error
	)
	for _, arg := range teamArgs {
		team, err := getTeamFromArg(c, arg)
		if err != nil {
			errors = append(errors, err)
		} else {
			teams = append(teams, team)
		}
	}
	if len(errors) > 0 {
		summary := &FindEntitySummary{
			Errors: errors,
		}
		return teams, summary
	}
	return teams, nil
}

func getTeamFromArg(c client.Client, teamArg string) (*model.Team, error) {
	if checkDots(teamArg) || checkSlash(teamArg) {
		return nil, ErrEntityNotFound{Type: "team", ID: teamArg}
	}
	var (
		team     *model.Team
		response *model.Response
	)
	team, response = c.GetTeam(teamArg, "")
	if isErrorSevere(response) {
		return nil, response.Error
	}
	if team != nil {
		return team, nil
	}
	team, response = c.GetTeamByName(teamArg, "")
	if isErrorSevere(response) {
		return nil, response.Error
	}
	if team == nil {
		return nil, ErrEntityNotFound{Type: "team", ID: teamArg}
	}
	return team, nil
}
