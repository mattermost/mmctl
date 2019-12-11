package commands

import (
	"errors"
	"fmt"
	"sort"

	"github.com/mattermost/mattermost-server/v5/model"
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

var RemoveUsersCmd = &cobra.Command{
	Use:     "remove [team] [users]",
	Short:   "Remove users from team",
	Long:    "Remove some users from team",
	Example: "  team remove myteam user@example.com username",
	RunE:    withClient(removeUsersCmdF),
}

var AddUsersCmd = &cobra.Command{
	Use:     "add [team] [users]",
	Short:   "Add users to team",
	Long:    "Add some users to team",
	Example: "  team add myteam user@example.com username",
	RunE:    withClient(addUsersCmdF),
}

var DeleteTeamsCmd = &cobra.Command{
	Use:   "delete [teams]",
	Short: "Delete teams",
	Long: `Permanently delete some teams.
Permanently deletes a team along with all related information including posts from the database.`,
	Example: "  team delete myteam",
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
	Short:   "List all teams.",
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
	Use:   "rename [team]",
	Short: "Rename team",
	Long:  "Rename an existing team",
	Example: ` team rename myoldteam newteamname --display_name 'New Team Name'
	team rename myoldteam - --display_name 'New Team Name'`,
	Args: cobra.MinimumNArgs(2),
	RunE: withClient(renameTeamCmdF),
}

func init() {
	TeamCreateCmd.Flags().String("name", "", "Team Name")
	TeamCreateCmd.Flags().String("display_name", "", "Team Display Name")
	TeamCreateCmd.Flags().Bool("private", false, "Create a private team.")
	TeamCreateCmd.Flags().String("email", "", "Administrator Email (anyone with this email is automatically a team admin)")

	DeleteTeamsCmd.Flags().Bool("confirm", false, "Confirm you really want to delete the team and a DB backup has been performed.")
	ArchiveTeamsCmd.Flags().Bool("confirm", false, "Confirm you really want to archive the team and a DB backup has been performed.")

	RenameTeamCmd.Flags().String("display_name", "", "Team Display Name")

	TeamCmd.AddCommand(
		TeamCreateCmd,
		RemoveUsersCmd,
		AddUsersCmd,
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
		return errors.New("Name is required")
	}
	displayname, errdn := cmd.Flags().GetString("display_name")
	if errdn != nil || displayname == "" {
		return errors.New("Display Name is required")
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

func removeUsersCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return errors.New("Not enough arguments.")
	}

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
	if len(args) < 2 {
		return errors.New("Not enough arguments.")
	}

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

func deleteTeamsCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("Not enough arguments.")
	}

	confirmFlag, _ := cmd.Flags().GetBool("confirm")
	if !confirmFlag {
		var confirm string
		fmt.Println("Have you performed a database backup? (YES/NO): ")
		fmt.Scanln(&confirm)

		if confirm != "YES" {
			return errors.New("ABORTED: You did not answer YES exactly, in all capitals.")
		}
		fmt.Println("Are you sure you want to delete the teams specified?  All data will be permanently deleted? (YES/NO): ")
		fmt.Scanln(&confirm)
		if confirm != "YES" {
			return errors.New("ABORTED: You did not answer YES exactly, in all capitals.")
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
			return errors.New("ABORTED: You did not answer YES exactly, in all capitals.")
		}
		fmt.Println("Are you sure you want to archive the specified teams? (YES/NO): ")
		fmt.Scanln(&confirm)
		if confirm != "YES" {
			return errors.New("ABORTED: You did not answer YES exactly, in all capitals.")
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
	teams, response := c.GetAllTeams("", 0, 10000)
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

	return nil
}

func searchTeamCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	var teams []*model.Team

	for _, searchTerm := range args {
		foundTeams, response := c.SearchTeams(&model.TeamSearch{Term: searchTerm})
		if response.Error != nil {
			return response.Error
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

// Wrapper around UpdateTeam, so it can be used by others too
func updateTeam(c client.Client, teamToUpdate *model.Team) (*model.Team, error) {
	updatedTeam, response := c.UpdateTeam(teamToUpdate)
	if response.Error != nil {
		return nil, response.Error
	}

	return updatedTeam, nil
}

func renameTeamCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	if len(args) == 0 || args[0] == "" {
		return errors.New("Error: requires at least 2 arg(s), only received 0")
	}

	if args[1] == "" {
		return errors.New("Error: required at least 2 arg(s), only received 1, If you like to change only display name; pass '-' after existing team name")
	}

	newDisplayName, err := cmd.Flags().GetString("display_name")
	if err != nil || newDisplayName == "" {
		return errors.New("Missing display name, append '--display_name' flag to your command")
	}
	oldTeamName := args[0]
	newTeamName := args[1]

	team := getTeamFromTeamArg(c, oldTeamName)
	if team == nil {
		return errors.New("Unable to find team '" + oldTeamName + "', to see the all teams try 'team list' command")
	}

	if newTeamName == team.Name {
		if newDisplayName == team.DisplayName {
			return errors.New("Failed to rename, entered display name and name are same for team")
		}
		// If new name entered is same as old, then to not update team name pass (-) to API
		newTeamName = "-"
	}

	// Update the team obj with new values
	team.Name = newTeamName
	team.DisplayName = newDisplayName

	// Using updateTeam to rename team
	updatedTeam, err := updateTeam(c, team)
	if err != nil {
		return errors.New("Cannot rename team '" + oldTeamName + "', error : " + err.Error())
	}

	// Only display name was suppose to renamed
	if newTeamName == "-" {
		if (updatedTeam.DisplayName == newDisplayName) && (oldTeamName == updatedTeam.Name) {
			printer.Print("Successfully renamed team '" + oldTeamName + "'")
			return nil
		}
		return errors.New("Failed to rename display name of '" + oldTeamName + "'")
	}
	// New name was provided to rename
	if newTeamName == updatedTeam.Name {
		if newDisplayName == updatedTeam.DisplayName {
			printer.Print("Successfully renamed team '" + oldTeamName + "'")
			return nil
		}
		return errors.New("Failed to rename display name of '" + oldTeamName + "'")
	}
	return errors.New("Failed to rename team '" + oldTeamName + "'")
}
