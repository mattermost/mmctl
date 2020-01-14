// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var TeamUsersCmd = &cobra.Command{
	Use:   "users",
	Short: "Management of team users",
}

var TeamUsersRemoveCmd = &cobra.Command{
	Use:        "remove [team] [users]",
	Short:      "Remove users from team",
	Long:       "Remove some users from team",
	Example:    "  team remove myteam user@example.com username",
	Deprecated: "please use \"archive\" instead",
	Args:       cobra.MinimumNArgs(2),
	RunE:       withClient(archiveUsersCmdF),
}

var TeamUsersArchiveCmd = &cobra.Command{
	Use:     "archive [team] [users]",
	Short:   "Archive users from team",
	Long:    "Archive some users from team",
	Example: "  team archive myteam user@example.com username",
	Args:    cobra.MinimumNArgs(2),
	RunE:    withClient(archiveUsersCmdF),
}

var TeamUsersAddCmd = &cobra.Command{
	Use:     "add [team] [users]",
	Short:   "Add users to team",
	Long:    "Add some users to team",
	Example: "  team add myteam user@example.com username",
	Args:    cobra.MinimumNArgs(2),
	RunE:    withClient(addUsersCmdF),
}

func init() {
	TeamUsersCmd.AddCommand(
		RemoveUsersCmd,
		ArchiveUsersCmd,
		AddUsersCmd,
	)

	TeamCmd.AddCommand(TeamUsersCmd)
}

func archiveUsersCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	team := getTeamFromTeamArg(c, args[0])
	if team == nil {
		return errors.New("Unable to find team '" + args[0] + "'")
	}

	users := getUsersFromUserArgs(c, args[1:])
	for i, user := range users {
		removeUserFromTeam(c, team, user, args[i+1])
	}

	return nil
}

func removeUserFromTeam(c client.Client, team *model.Team, user *model.User, userArg string) {
	if user == nil {
		printer.PrintError("Can't find user '" + userArg + "'")
		return
	}
	if _, response := c.RemoveTeamMember(team.Id, user.Id); response.Error != nil {
		printer.PrintError("Unable to remove '" + userArg + "' from " + team.Name + ". Error: " + response.Error.Error())
	}
}

func addUsersCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	team := getTeamFromTeamArg(c, args[0])
	if team == nil {
		return errors.New("Unable to find team '" + args[0] + "'")
	}

	users := getUsersFromUserArgs(c, args[1:])
	for i, user := range users {
		addUserToTeam(c, team, user, args[i+1])
	}

	return nil
}

func addUserToTeam(c client.Client, team *model.Team, user *model.User, userArg string) {
	if user == nil {
		printer.PrintError("Can't find user '" + userArg + "'")
		return
	}

	if _, response := c.AddTeamMember(team.Id, user.Id); response.Error != nil {
		printer.PrintError("Unable to add '" + userArg + "' to " + team.Name + ". Error: " + response.Error.Error())
	}
}
