package commands

import (
	"errors"
	"fmt"
	"sort"

	"github.com/mattermost/mattermost-server/model"
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
	RunE: createTeamCmdF,
}

var RemoveUsersCmd = &cobra.Command{
	Use:     "remove [team] [users]",
	Short:   "Remove users from team",
	Long:    "Remove some users from team",
	Example: "  team remove myteam user@example.com username",
	RunE:    removeUsersCmdF,
}

var AddUsersCmd = &cobra.Command{
	Use:     "add [team] [users]",
	Short:   "Add users to team",
	Long:    "Add some users to team",
	Example: "  team add myteam user@example.com username",
	RunE:    addUsersCmdF,
}

var DeleteTeamsCmd = &cobra.Command{
	Use:   "delete [teams]",
	Short: "Delete teams",
	Long: `Permanently delete some teams.
Permanently deletes a team along with all related information including posts from the database.`,
	Example: "  team delete myteam",
	RunE:    deleteTeamsCmdF,
}

var ListTeamsCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all teams.",
	Long:    `List all teams on the server.`,
	Example: "  team list",
	RunE:    listTeamsCmdF,
}

var SearchTeamCmd = &cobra.Command{
	Use:     "search [teams]",
	Short:   "Search for teams",
	Long:    "Search for teams based on name",
	Example: "  team search team1",
	Args:    cobra.MinimumNArgs(1),
	RunE:    searchTeamCmdF,
}

func init() {
	TeamCreateCmd.Flags().String("name", "", "Team Name")
	TeamCreateCmd.Flags().String("display_name", "", "Team Display Name")
	TeamCreateCmd.Flags().Bool("private", false, "Create a private team.")
	TeamCreateCmd.Flags().String("email", "", "Administrator Email (anyone with this email is automatically a team admin)")

	DeleteTeamsCmd.Flags().Bool("confirm", false, "Confirm you really want to delete the team and a DB backup has been performed.")

	TeamCmd.AddCommand(
		TeamCreateCmd,
		RemoveUsersCmd,
		AddUsersCmd,
		DeleteTeamsCmd,
		ListTeamsCmd,
		SearchTeamCmd,
	)

	RootCmd.AddCommand(TeamCmd)
}

func createTeamCmdF(command *cobra.Command, args []string) error {
	c, err := InitClient()
	if err != nil {
		return err
	}
	printer.SetSingle(true)

	name, errn := command.Flags().GetString("name")
	if errn != nil || name == "" {
		return errors.New("Name is required")
	}
	displayname, errdn := command.Flags().GetString("display_name")
	if errdn != nil || displayname == "" {
		return errors.New("Display Name is required")
	}
	email, _ := command.Flags().GetString("email")
	useprivate, _ := command.Flags().GetBool("private")

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

func removeUsersCmdF(command *cobra.Command, args []string) error {
	c, err := InitClient()
	if err != nil {
		return err
	}

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

func removeUserFromTeam(c *model.Client4, team *model.Team, user *model.User, userArg string) {
	if user == nil {
		CommandPrintErrorln("Can't find user '" + userArg + "'")
		return
	}
	if _, response := c.RemoveTeamMember(team.Id, user.Id); response.Error != nil {
		CommandPrintErrorln("Unable to remove '" + userArg + "' from " + team.Name + ". Error: " + response.Error.Error())
	}
}

func addUsersCmdF(command *cobra.Command, args []string) error {
	c, err := InitClient()
	if err != nil {
		return err
	}

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

func addUserToTeam(c *model.Client4, team *model.Team, user *model.User, userArg string) {
	if user == nil {
		CommandPrintErrorln("Can't find user '" + userArg + "'")
		return
	}

	if _, response := c.AddTeamMember(team.Id, user.Id); response.Error != nil {
		CommandPrintErrorln("Unable to add '" + userArg + "' to " + team.Name + ". Error: " + response.Error.Error())
	}
}

func deleteTeamsCmdF(command *cobra.Command, args []string) error {
	c, err := InitClient()
	if err != nil {
		return err
	}

	if len(args) < 1 {
		return errors.New("Not enough arguments.")
	}

	confirmFlag, _ := command.Flags().GetBool("confirm")
	if !confirmFlag {
		var confirm string
		CommandPrettyPrintln("Have you performed a database backup? (YES/NO): ")
		fmt.Scanln(&confirm)

		if confirm != "YES" {
			return errors.New("ABORTED: You did not answer YES exactly, in all capitals.")
		}
		CommandPrettyPrintln("Are you sure you want to delete the teams specified?  All data will be permanently deleted? (YES/NO): ")
		fmt.Scanln(&confirm)
		if confirm != "YES" {
			return errors.New("ABORTED: You did not answer YES exactly, in all capitals.")
		}
	}

	teams := getTeamsFromTeamArgs(c, args)
	for i, team := range teams {
		if team == nil {
			CommandPrintErrorln("Unable to find team '" + args[i] + "'")
			continue
		}
		if _, response := deleteTeam(c, team); response.Error != nil {
			CommandPrintErrorln("Unable to delete team '" + team.Name + "' error: " + response.Error.Error())
		} else {
			printer.PrintT("Deleted team '{{.Name}}'", team)
		}
	}

	return nil
}

func deleteTeam(c *model.Client4, team *model.Team) (bool, *model.Response) {
	return c.PermanentDeleteTeam(team.Id)
}

func listTeamsCmdF(command *cobra.Command, args []string) error {
	c, err := InitClient()
	if err != nil {
		return err
	}

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

func searchTeamCmdF(command *cobra.Command, args []string) error {
	c, err := InitClient()
	if err != nil {
		return err
	}

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
