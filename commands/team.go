// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"errors"
	"sort"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"

	"github.com/spf13/cobra"
)

const APILimitMaximum = 200

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

var RestoreTeamsCmd = &cobra.Command{
	Use:     "restore [teams]",
	Short:   "Restore teams",
	Long:    "Restores archived teams.",
	Example: "  team restore myteam",
	Args:    cobra.MinimumNArgs(1),
	RunE:    withClient(restoreTeamsCmdF),
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

var ModifyTeamsCmd = &cobra.Command{
	Use:     "modify [teams] [flag]",
	Short:   "Modify teams",
	Long:    "Modify teams' privacy setting to public or private",
	Example: "  team modify myteam --private",
	Args:    cobra.MinimumNArgs(1),
	RunE:    withClient(modifyTeamsCmdF),
}

func init() {
	TeamCreateCmd.Flags().String("name", "", "Team Name")
	TeamCreateCmd.Flags().String("display_name", "", "Team Display Name")
	TeamCreateCmd.Flags().Bool("private", false, "Create a private team.")
	TeamCreateCmd.Flags().String("email", "", "Administrator Email (anyone with this email is automatically a team admin)")

	DeleteTeamsCmd.Flags().Bool("confirm", false, "Confirm you really want to delete the team and a DB backup has been performed.")
	ArchiveTeamsCmd.Flags().Bool("confirm", false, "Confirm you really want to archive the team and a DB backup has been performed.")

	ModifyTeamsCmd.Flags().Bool("private", false, "Modify team to be private.")
	ModifyTeamsCmd.Flags().Bool("public", false, "Modify team to be public.")

	// Add flag declaration for RenameTeam
	RenameTeamCmd.Flags().String("display_name", "", "Team Display Name")
	_ = RenameTeamCmd.MarkFlagRequired("display_name")

	TeamCmd.AddCommand(
		TeamCreateCmd,
		DeleteTeamsCmd,
		ArchiveTeamsCmd,
		RestoreTeamsCmd,
		ListTeamsCmd,
		SearchTeamCmd,
		RenameTeamCmd,
		ModifyTeamsCmd,
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

	teamType := model.TeamOpen
	if useprivate {
		teamType = model.TeamInvite
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
		if err := getConfirmation("Are you sure you want to archive the specified teams?", true); err != nil {
			return err
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
		teams, response := c.GetAllTeams("", page, APILimitMaximum)
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

		if len(teams) < APILimitMaximum {
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
		if err := getConfirmation("Are you sure you want to delete the teams specified?  All data will be permanently deleted?", true); err != nil {
			return err
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

func modifyTeamsCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	private, _ := cmd.Flags().GetBool("private")
	public, _ := cmd.Flags().GetBool("public")

	if (!private && !public) || (private && public) {
		return errors.New("must specify one of --private or --public")
	}

	// I = invite only (private)
	// O = open (public)
	privacy := model.TeamInvite
	if public {
		privacy = model.TeamOpen
	}

	teams := getTeamsFromTeamArgs(c, args)
	for i, team := range teams {
		if team == nil {
			printer.PrintError("Unable to find team '" + args[i] + "'")
			continue
		}
		if updatedTeam, response := c.UpdateTeamPrivacy(team.Id, privacy); response.Error != nil {
			printer.PrintError("Unable to modify team '" + team.Name + "' error: " + response.Error.Error())
		} else {
			printer.PrintT("Modified team '{{.Name}}'", updatedTeam)
		}
	}

	return nil
}

func restoreTeamsCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	teams := getTeamsFromTeamArgs(c, args)
	for i, team := range teams {
		if team == nil {
			printer.PrintError("Unable to find team '" + args[i] + "'")
			continue
		}
		if rteam, response := c.RestoreTeam(team.Id); response.Error != nil {
			printer.PrintError("Unable to restore team '" + team.Name + "' error: " + response.Error.Error())
		} else {
			printer.PrintT("Restored team '{{.Name}}'", rteam)
		}
	}
	return nil
}
