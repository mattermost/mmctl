package commands

import (
	"errors"
	"fmt"
	"sort"

	"github.com/mattermost/mattermost-server/model"
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

var TeamGroupConstrainedCmd = &cobra.Command{
	Use:   "group-constrained",
	Short: "Manage group-constrained status",
	Long:  "Manage team group-constrained status and it's associated groups",
}

var TeamGroupConstrainedEnableCmd = &cobra.Command{
	Use:     "enable [team]",
	Short:   "Enables group-constrained restrictions in the specified team",
	Example: "  team group-constrained enable myteam",
	Args:    cobra.ExactArgs(1),
	RunE:    teamGroupConstrainedEnableCmdF,
}

var TeamGroupConstrainedDisableCmd = &cobra.Command{
	Use:     "disable [team]",
	Short:   "Disables group-constrained restrictions in the specified team",
	Example: "  team group-constrained disable myteam",
	Args:    cobra.ExactArgs(1),
	RunE:    teamGroupConstrainedDisableCmdF,
}

var TeamGroupConstrainedStatusCmd = &cobra.Command{
	Use:     "status [team]",
	Short:   "Show's the group-constrained status for the specified team",
	Example: "  team group-constrained status myteam",
	Args:    cobra.ExactArgs(1),
	RunE:    teamGroupConstrainedStatusCmdF,
}

var TeamGroupConstrainedGetGroupsCmd = &cobra.Command{
	Use:     "get-groups",
	Short:   "Get team's groups",
	Long:    "List the groups associated with a team",
	Example: "  team get-groups myteam",
	Args:    cobra.ExactArgs(1),
	RunE:    teamGroupConstrainedGetGroupsCmdF,
}

func init() {
	TeamCreateCmd.Flags().String("name", "", "Team Name")
	TeamCreateCmd.Flags().String("display_name", "", "Team Display Name")
	TeamCreateCmd.Flags().Bool("private", false, "Create a private team.")
	TeamCreateCmd.Flags().String("email", "", "Administrator Email (anyone with this email is automatically a team admin)")

	DeleteTeamsCmd.Flags().Bool("confirm", false, "Confirm you really want to delete the team and a DB backup has been performed.")

	TeamGroupConstrainedCmd.AddCommand(
		TeamGroupConstrainedEnableCmd,
		TeamGroupConstrainedDisableCmd,
		TeamGroupConstrainedStatusCmd,
		TeamGroupConstrainedGetGroupsCmd,
	)

	TeamCmd.AddCommand(
		TeamCreateCmd,
		RemoveUsersCmd,
		AddUsersCmd,
		DeleteTeamsCmd,
		ListTeamsCmd,
		SearchTeamCmd,
		TeamGroupConstrainedCmd,
	)
	RootCmd.AddCommand(TeamCmd)
}

func createTeamCmdF(command *cobra.Command, args []string) error {
	c, err := InitClient()
	if err != nil {
		return err
	}

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

	if _, response := c.CreateTeam(team); response.Error != nil {
		return errors.New("Team creation failed: " + response.Error.Error())
	}

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
			CommandPrettyPrintln("Deleted team '" + team.Name + "'")
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
		CommandPrettyPrintln(team.Name)
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
		CommandPrettyPrintln(team.Name + ": " + team.DisplayName + " (" + team.Id + ")")
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

func teamGroupConstrainedEnableCmdF(command *cobra.Command, args []string) error {
	c, err := InitClient()
	if err != nil {
		return err
	}

	team := getTeamFromTeamArg(c, args[0])
	if team == nil {
		return errors.New("Unable to find team '" + args[0] + "'")
	}

	groups, res := c.GetGroupsByTeam(team.Id, 0, 10)
	if res.Error != nil {
		return res.Error
	}

	if len(groups) == 0 {
		return errors.New("Team '" + args[0] + "' has no groups associated. It cannot be group-constrained")
	}

	teamPatch := model.TeamPatch{GroupConstrained: model.NewBool(true)}
	if _, res = c.PatchTeam(team.Id, &teamPatch); res.Error != nil {
		return res.Error
	}

	return nil
}

func teamGroupConstrainedDisableCmdF(command *cobra.Command, args []string) error {
	c, err := InitClient()
	if err != nil {
		return err
	}

	team := getTeamFromTeamArg(c, args[0])
	if team == nil {
		return errors.New("Unable to find team '" + args[0] + "'")
	}

	teamPatch := model.TeamPatch{GroupConstrained: model.NewBool(false)}
	if _, res := c.PatchTeam(team.Id, &teamPatch); res.Error != nil {
		return res.Error
	}

	return nil
}

func teamGroupConstrainedStatusCmdF(command *cobra.Command, args []string) error {
	c, err := InitClient()
	if err != nil {
		return err
	}

	team := getTeamFromTeamArg(c, args[0])
	if team == nil {
		return errors.New("Unable to find team '" + args[0] + "'")
	}

	if team.GroupConstrained != nil && *team.GroupConstrained {
		fmt.Println("Enabled")
	} else {
		fmt.Println("Disabled")
	}

	return nil
}

func teamGroupConstrainedGetGroupsCmdF(command *cobra.Command, args []string) error {
	c, err := InitClient()
	if err != nil {
		return err
	}

	team := getTeamFromTeamArg(c, args[0])
	if team == nil {
		return errors.New("Unable to find team '" + args[0] + "'")
	}

	groups, res := c.GetGroupsByTeam(team.Id, 0, 9999)
	if res.Error != nil {
		return res.Error
	}

	for _, group := range groups {
		fmt.Println(group.DisplayName)
	}

	return nil
}
