package commands

import (
	"github.com/mattermost/mattermost-server/model"
)

func getTeamsFromTeamArgs(c *model.Client4, teamArgs []string) []*model.Team {
	teams := make([]*model.Team, 0, len(teamArgs))
	for _, teamArg := range teamArgs {
		team := getTeamFromTeamArg(c, teamArg)
		teams = append(teams, team)
	}
	return teams
}

func getTeamFromTeamArg(c *model.Client4, teamArg string) *model.Team {
	var team *model.Team
	team, _ = c.GetTeam(teamArg, "")

	if team == nil {
		team, _ = c.GetTeamByName(teamArg, "")
	}

	return team
}
