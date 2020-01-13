// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"errors"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"
	"github.com/spf13/cobra"
)

var CommandCmd = &cobra.Command{
	Use:   "command",
	Short: "Management of slash commands",
}

var CommandCreateCmd = &cobra.Command{
	Use:     "create [team]",
	Short:   "Create a custom slash command",
	Long:    `Create a custom slash command for the specified team.`,
	Args:    cobra.MinimumNArgs(1),
	Example: `  command create myteam --title MyCommand --description "My Command Description" --trigger-word mycommand --url http://localhost:8000/my-slash-handler --creator myusername --response-username my-bot-username --icon http://localhost:8000/my-slash-handler-bot-icon.png --autocomplete --post`,
	RunE:    withClient(createCommandCmdF),
}

var CommandListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all commands on specified teams.",
	Long:    `List all commands on specified teams.`,
	Example: ` command list myteam`,
	RunE:    withClient(listCommandCmdF),
}

var CommandDeleteCmd = &cobra.Command{
	Use:        "delete",
	Short:      "Delete a slash command",
	Long:       `Delete a slash command. Commands can be specified by command ID.`,
	Example:    `  command delete commandID`,
	Deprecated: "This command is deprecated, please use archive instead",
	Args:       cobra.ExactArgs(1),
	RunE:       withClient(archiveCommandCmdF),
}

var CommandArchiveCmd = &cobra.Command{
	Use:     "archive",
	Short:   "Archive a slash command",
	Long:    `Archive a slash command. Commands can be specified by command ID.`,
	Example: `  command archive commandID`,
	Args:    cobra.ExactArgs(1),
	RunE:    withClient(archiveCommandCmdF),
}

func init() {
	CommandCreateCmd.Flags().String("title", "", "Command Title")
	CommandCreateCmd.Flags().String("description", "", "Command Description")
	CommandCreateCmd.Flags().String("trigger-word", "", "Command Trigger Word (required)")
	CommandCreateCmd.MarkFlagRequired("trigger-word")
	CommandCreateCmd.Flags().String("url", "", "Command Callback URL (required)")
	CommandCreateCmd.MarkFlagRequired("url")
	CommandCreateCmd.Flags().String("creator", "", "Command Creator's Username (required)")
	CommandCreateCmd.MarkFlagRequired("creator")
	CommandCreateCmd.Flags().String("response-username", "", "Command Response Username")
	CommandCreateCmd.Flags().String("icon", "", "Command Icon URL")
	CommandCreateCmd.Flags().Bool("autocomplete", false, "Show Command in autocomplete list")
	CommandCreateCmd.Flags().String("autocompleteDesc", "", "Short Command Description for autocomplete list")
	CommandCreateCmd.Flags().String("autocompleteHint", "", "Command Arguments displayed as help in autocomplete list")
	CommandCreateCmd.Flags().Bool("post", false, "Use POST method for Callback URL")

	CommandCmd.AddCommand(
		CommandCreateCmd,
		CommandListCmd,
		CommandDeleteCmd,
		CommandArchiveCmd,
	)
	RootCmd.AddCommand(CommandCmd)
}

func createCommandCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	printer.SetSingle(true)

	team := getTeamFromTeamArg(c, args[0])
	if team == nil {
		return errors.New("unable to find team '" + args[0] + "'")
	}

	// get the creator
	creator, _ := cmd.Flags().GetString("creator")
	user := getUserFromUserArg(c, creator)
	if user == nil {
		return errors.New("unable to find user '" + creator + "'")
	}

	title, _ := cmd.Flags().GetString("title")
	description, _ := cmd.Flags().GetString("description")
	trigger, _ := cmd.Flags().GetString("trigger-word")

	if strings.HasPrefix(trigger, "/") {
		return errors.New("a trigger word cannot begin with a /")
	}
	if strings.Contains(trigger, " ") {
		return errors.New("a trigger word must not contain spaces")
	}

	url, _ := cmd.Flags().GetString("url")
	responseUsername, _ := cmd.Flags().GetString("response-username")
	icon, _ := cmd.Flags().GetString("icon")
	autocomplete, _ := cmd.Flags().GetBool("autocomplete")
	autocompleteDesc, _ := cmd.Flags().GetString("autocompleteDesc")
	autocompleteHint, _ := cmd.Flags().GetString("autocompleteHint")
	post, errp := cmd.Flags().GetBool("post")
	method := "P"
	if errp != nil || post == false {
		method = "G"
	}

	newCommand := &model.Command{
		CreatorId:        user.Id,
		TeamId:           team.Id,
		Trigger:          trigger,
		Method:           method,
		Username:         responseUsername,
		IconURL:          icon,
		AutoComplete:     autocomplete,
		AutoCompleteDesc: autocompleteDesc,
		AutoCompleteHint: autocompleteHint,
		DisplayName:      title,
		Description:      description,
		URL:              url,
	}

	createdCommand, response := c.CreateCommand(newCommand)
	if response.Error != nil {
		return errors.New("unable to create command '" + newCommand.DisplayName + "'. " + response.Error.Error())
	}

	printer.PrintT("created command {{.DisplayName}}", createdCommand)

	return nil
}

func listCommandCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	var teams []*model.Team
	if len(args) < 1 {
		teamList, response := c.GetAllTeams("", 0, 10000)
		if response.Error != nil {
			return response.Error
		}
		teams = teamList
	} else {
		teams = getTeamsFromTeamArgs(c, args)
	}

	for i, team := range teams {
		if team == nil {
			printer.PrintError("Unable to find team '" + args[i] + "'")
			continue
		}
		commands, response := c.ListCommands(team.Id, true)
		if response.Error != nil {
			printer.PrintError("Unable to list commands for '" + args[i] + "'")
			continue
		}
		for _, command := range commands {
			printer.PrintT("{{.Id}}: {{.DisplayName}} (team: "+team.Name+")", command)
		}
	}
	return nil
}

func archiveCommandCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	ok, response := c.DeleteCommand(args[0])
	if response.Error != nil {
		return errors.New("Unable to delete command '" + args[0] + "' error: " + response.Error.Error())
	}

	if ok {
		printer.PrintT("Status: {{.status}}", map[string]interface{}{"status": "ok"})
	} else {
		printer.PrintT("Status: {{.status}}", map[string]interface{}{"status": "error"})
	}
	return nil
}
