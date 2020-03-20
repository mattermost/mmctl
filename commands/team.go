// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"errors"
	"fmt"
	"sort"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/web"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"

	"github.com/spf13/cobra"
)

var TeamCmd = &cobra.Command{
	Use:   "team",
	Short: "Management of teams",
}

var TeamCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a team",
	Long:  `Create a team.`,
	Example: `  team create --name mynewteam --display_name "My New Team"
  team create --name private --display_name "My New Private Team" --private`,
	RunE: withClient(createTeamCmdF),
}

var DeleteTeamsCmd = &cobra.Command{
	Use:   "delete [teams]",
	Short: "Delete teams",
	Long: `Permanently delete some teams.
Permanently deletes a team along with all related information including posts from the database.`,
	Example: "  team delete myteam",
	Args:    cobra.MinimumNArgs(1),
	RunE:    withClient(deleteTeamsCmdF),
}

var ArchiveTeamsCmd = &cobra.Command{
	Use:   "archive [teams]",
	Short: "Archive teams",
	Long: `Archive some teams.
Archives a team along with all related information including posts from the database.`,
	Example: "  team archive myteam",
	Args:    cobra.MinimumNArgs(1),
	RunE:    withClient(archiveTeamsCmdF),
}

var ListTeamsCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all teams",
	Long:    `List all teams on the server.`,
	Example: "  team list",
	RunE:    withClient(listTeamsCmdF),
}

var SearchTeamCmd = &cobra.Command{
	Use:     "search [teams]",
	Short:   "Search for teams",
	Long:    "Search for teams based on name",
	Example: "  team search team1",
	Args:    cobra.MinimumNArgs(1),
	RunE:    withClient(searchTeamCmdF),
}

// RenameTeamCmd is the command to rename team along with its display name
var RenameTeamCmd = &cobra.Command{
	Use:     "rename [team]",
	Short:   "Rename team",
	Long:    "Rename an existing team",
	Example: "  team rename old-team --display_name 'New Display Name'",
	Args:    cobra.ExactArgs(1),
	RunE:    withClient(renameTeamCmdF),
}

func init() {
	TeamCreateCmd.Flags().String("name", "", "Team Name")
	TeamCreateCmd.Flags().String("display_name", "", "Team Display Name")
	TeamCreateCmd.Flags().Bool("private", false, "Create a private team.")
	TeamCreateCmd.Flags().String("email", "", "Administrator Email (anyone with this email is automatically a team admin)")

	DeleteTeamsCmd.Flags().Bool("confirm", false, "Confirm you really want to delete the team and a DB backup has been performed.")
	ArchiveTeamsCmd.Flags().Bool("confirm", false, "Confirm you really want to archive the team and a DB backup has been performed.")

	// Add flag declaration for RenameTeam
	RenameTeamCmd.Flags().String("display_name", "", "Team Display Name")
	_ = RenameTeamCmd.MarkFlagRequired("display_name")

	TeamCmd.AddCommand(
		TeamCreateCmd,
		DeleteTeamsCmd,
		ArchiveTeamsCmd,
		ListTeamsCmd,
		SearchTeamCmd,
		RenameTeamCmd,
	)

	RootCmd.AddCommand(TeamCmd)
}

func createTeamCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	printer.SetSingle(true)

	name, errn := cmd.Flags().GetString("name")
	if errn != nil || name == "" {
		return errors.New("name is required")
	}
	displayname, errdn := cmd.Flags().GetString("display_name")
	if errdn != nil || displayname == "" {
		return errors.New("display Name is required")
	}
	email, _ := cmd.Flags().GetString("email")
	useprivate, _ := cmd.Flags().GetBool("private")

	teamType := model.TEAM_OPEN
	if useprivate {
		teamType = model.TEAM_INVITE
	}

	team := &model.Team{
		Name:        name,
		DisplayName: displayname,
		Email:       email,
		Type:        teamType,
	}

	newTeam, response := c.CreateTeam(team)
	if response.Error != nil {
		return errors.New("Team creation failed: " + response.Error.Error())
	}

	printer.PrintT("New team {{.Name}} successfully created", newTeam)

	return nil
}

func deleteTeam(c client.Client, team *model.Team) (bool, *model.Response) {
	return c.PermanentDeleteTeam(team.Id)
}

func archiveTeamsCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	confirmFlag, _ := cmd.Flags().GetBool("confirm")
	if !confirmFlag {
		var confirm string
		fmt.Println("Have you performed a database backup? (YES/NO): ")
		fmt.Scanln(&confirm)

		if confirm != "YES" {
			return errors.New("ABORTED: You did not answer YES exactly, in all capitals")
		}
		fmt.Println("Are you sure you want to archive the specified teams? (YES/NO): ")
		fmt.Scanln(&confirm)
		if confirm != "YES" {
			return errors.New("ABORTED: You did not answer YES exactly, in all capitals")
		}
	}

	teams := getTeamsFromTeamArgs(c, args)
	for i, team := range teams {
		if team == nil {
			printer.PrintError("Unable to find team '" + args[i] + "'")
			continue
		}
		if _, response := c.SoftDeleteTeam(team.Id); response.Error != nil {
			printer.PrintError("Unable to archive team '" + team.Name + "' error: " + response.Error.Error())
		} else {
			printer.PrintT("Archived team '{{.Name}}'", team)
		}
	}

	return nil
}

func listTeamsCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	page := 0
	for {
		teams, response := c.GetAllTeams("", page, web.LIMIT_MAXIMUM)
		if response.Error != nil {
			return response.Error
		}

		for _, team := range teams {
			if team.DeleteAt > 0 {
				printer.PrintT("{{.Name}} (archived)", team)
			} else {
				printer.PrintT("{{.Name}}", team)
			}
		}

		if len(teams) < web.LIMIT_MAXIMUM {
			break
		}

		page++
	}

	return nil
}

func searchTeamCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	var teams []*model.Team

	for _, searchTerm := range args {
		foundTeams, response := c.SearchTeams(&model.TeamSearch{Term: searchTerm})
		if response.Error != nil {
			return response.Error
		}

		if len(foundTeams) == 0 {
			printer.PrintError("Unable to find team '" + searchTerm + "'")
			continue
		}

		teams = append(teams, foundTeams...)
	}

	sortedTeams := removeDuplicatesAndSortTeams(teams)

	for _, team := range sortedTeams {
		printer.PrintT("{{.Name}}: {{.DisplayName}} ({{.Id}})", team)
	}

	return nil
}

// Removes duplicates and sorts teams by name
func removeDuplicatesAndSortTeams(teams []*model.Team) []*model.Team {
	keys := make(map[string]bool)
	result := []*model.Team{}
	for _, team := range teams {
		if _, value := keys[team.Name]; !value {
			keys[team.Name] = true
			result = append(result, team)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

func renameTeamCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	oldTeamName := args[0]
	newDisplayName, _ := cmd.Flags().GetString("display_name")

	team := getTeamFromTeamArg(c, oldTeamName)
	if team == nil {
		return errors.New("Unable to find team '" + oldTeamName + "', to see the all teams try 'team list' command")
	}

	team.DisplayName = newDisplayName

	// Using UpdateTeam API Method to rename team
	_, response := c.UpdateTeam(team)
	if response.Error != nil {
		return errors.New("Cannot rename team '" + oldTeamName + "', error : " + response.Error.Error())
	}

	printer.Print("'" + oldTeamName + "' team renamed")
	return nil
}

func deleteTeamsCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	confirmFlag, _ := cmd.Flags().GetBool("confirm")
	if !confirmFlag {
		var confirm string
		fmt.Println("Have you performed a database backup? (YES/NO): ")
		fmt.Scanln(&confirm)

		if confirm != "YES" {
			return errors.New("ABORTED: You did not answer YES exactly, in all capitals")
		}
		fmt.Println("Are you sure you want to delete the teams specified?  All data will be permanently deleted? (YES/NO): ")
		fmt.Scanln(&confirm)
		if confirm != "YES" {
			return errors.New("ABORTED: You did not answer YES exactly, in all capitals")
		}
	}

	teams := getTeamsFromTeamArgs(c, args)
	for i, team := range teams {
		if team == nil {
			printer.PrintError("Unable to find team '" + args[i] + "'")
			continue
		}
		if _, response := deleteTeam(c, team); response.Error != nil {
			printer.PrintError("Unable to delete team '" + team.Name + "' error: " + response.Error.Error())
		} else {
			printer.PrintT("Deleted team '{{.Name}}'", team)
		}
	}

	return nil
}
