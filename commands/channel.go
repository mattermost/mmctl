// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var ChannelCmd = &cobra.Command{
	Use:   "channel",
	Short: "Management of channels",
}

var ChannelCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a channel",
	Long:  `Create a channel.`,
	Example: `  channel create --team myteam --name mynewchannel --display_name "My New Channel"
  channel create --team myteam --name mynewprivatechannel --display_name "My New Private Channel" --private`,
	RunE: withClient(createChannelCmdF),
}

// ChannelRenameCmd is used to change name and/or display name of an existing channel.
var ChannelRenameCmd = &cobra.Command{
	Use:   "rename [channel]",
	Short: "Rename channel",
	Long:  `Rename an existing channel.`,
	Example: `  channel rename myteam:oldchannel --name 'new-channel' --display_name 'New Display Name'
  channel rename myteam:oldchannel --name 'new-channel'
  channel rename myteam:oldchannel --display_name 'New Display Name'`,
	Args: cobra.ExactArgs(1),
	RunE: withClient(renameChannelCmdF),
}

var RemoveChannelUsersCmd = &cobra.Command{
	Use:   "remove [channel] [users]",
	Short: "Remove users from channel",
	Long:  "Remove some users from channel",
	Example: `  channel remove myteam:mychannel user@example.com username
  channel remove myteam:mychannel --all-users`,
	RunE: withClient(removeChannelUsersCmdF),
}

var AddChannelUsersCmd = &cobra.Command{
	Use:     "add [channel] [users]",
	Short:   "Add users to channel",
	Long:    "Add some users to channel",
	Example: "  channel add myteam:mychannel user@example.com username",
	RunE:    withClient(addChannelUsersCmdF),
}

var ArchiveChannelsCmd = &cobra.Command{
	Use:   "archive [channels]",
	Short: "Archive channels",
	Long: `Archive some channels.
Archive a channel along with all related information including posts from the database.
Channels can be specified by [team]:[channel]. ie. myteam:mychannel or by channel ID.`,
	Example: "  channel archive myteam:mychannel",
	RunE:    withClient(archiveChannelsCmdF),
}

// ListChannelsCmd is a command which lists all the channels of team(s) in a server.
var ListChannelsCmd = &cobra.Command{
	Use:   "list [teams]",
	Short: "List all channels on specified teams.",
	Long: `List all channels on specified teams.
Archived channels are appended with ' (archived)'.`,
	Example: "  channel list myteam",
	Args:    cobra.MinimumNArgs(1),
	RunE:    withClient(listChannelsCmdF),
}

var ModifyChannelCmd = &cobra.Command{
	Use:   "modify [channel] [flags]",
	Short: "Modify a channel's public/private type",
	Long: `Change the public/private type of a channel.
Channel can be specified by [team]:[channel]. ie. myteam:mychannel or by channel ID.`,
	Example: `  channel modify myteam:mychannel --private
  channel modify channelId --public`,
	Args: cobra.ExactArgs(1),
	RunE: withClient(modifyChannelCmdF),
}

var RestoreChannelsCmd = &cobra.Command{
	Use:   "restore [channels]",
	Short: "Restore some channels",
	Long: `Restore a previously deleted channel
Channels can be specified by [team]:[channel]. ie. myteam:mychannel or by channel ID.`,
	Example: "  channel restore myteam:mychannel",
	RunE:    withClient(restoreChannelsCmdF),
}

var MakeChannelPrivateCmd = &cobra.Command{
	Use:   "make_private [channel]",
	Short: "Set a channel's type to private",
	Long: `Set the type of a channel from public to private.
Channel can be specified by [team]:[channel]. ie. myteam:mychannel or by channel ID.`,
	Example: "  channel make_private myteam:mychannel",
	RunE:    withClient(makeChannelPrivateCmdF),
}

var SearchChannelCmd = &cobra.Command{
	Use:   "search [channel]\n  mmctl search --team [team] [channel]",
	Short: "Search a channel",
	Long: `Search a channel by channel name.
Channel can be specified by team. ie. --team myTeam myChannel or by team ID.`,
	Example: `  channel search myChannel
  channel search --team myTeam myChannel`,
	Args: cobra.ExactArgs(1),
	RunE: withClient(searchChannelCmdF),
}

func init() {
	ChannelCreateCmd.Flags().String("name", "", "Channel Name")
	ChannelCreateCmd.Flags().String("display_name", "", "Channel Display Name")
	ChannelCreateCmd.Flags().String("team", "", "Team name or ID")
	ChannelCreateCmd.Flags().String("header", "", "Channel header")
	ChannelCreateCmd.Flags().String("purpose", "", "Channel purpose")
	ChannelCreateCmd.Flags().Bool("private", false, "Create a private channel.")

	ModifyChannelCmd.Flags().Bool("private", false, "Convert the channel to a private channel")
	ModifyChannelCmd.Flags().Bool("public", false, "Convert the channel to a public channel")

	ChannelRenameCmd.Flags().String("name", "", "Channel Name")
	ChannelRenameCmd.Flags().String("display_name", "", "Channel Display Name")

	RemoveChannelUsersCmd.Flags().Bool("all-users", false, "Remove all users from the indicated channel.")

	SearchChannelCmd.Flags().String("team", "", "Team name or ID")

	ChannelCmd.AddCommand(
		ChannelCreateCmd,
		RemoveChannelUsersCmd,
		AddChannelUsersCmd,
		ArchiveChannelsCmd,
		ListChannelsCmd,
		RestoreChannelsCmd,
		MakeChannelPrivateCmd,
		ModifyChannelCmd,
		ChannelRenameCmd,
		SearchChannelCmd,
	)

	RootCmd.AddCommand(ChannelCmd)
}

func createChannelCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	printer.SetSingle(true)

	name, errn := cmd.Flags().GetString("name")
	if errn != nil || name == "" {
		return errors.New("name is required")
	}
	displayname, errdn := cmd.Flags().GetString("display_name")
	if errdn != nil || displayname == "" {
		return errors.New("display Name is required")
	}
	teamArg, errteam := cmd.Flags().GetString("team")
	if errteam != nil || teamArg == "" {
		return errors.New("team is required")
	}
	header, _ := cmd.Flags().GetString("header")
	purpose, _ := cmd.Flags().GetString("purpose")
	useprivate, _ := cmd.Flags().GetBool("private")

	channelType := model.CHANNEL_OPEN
	if useprivate {
		channelType = model.CHANNEL_PRIVATE
	}

	team := getTeamFromTeamArg(c, teamArg)
	if team == nil {
		return errors.New("Unable to find team: " + teamArg)
	}

	channel := &model.Channel{
		TeamId:      team.Id,
		Name:        name,
		DisplayName: displayname,
		Header:      header,
		Purpose:     purpose,
		Type:        channelType,
		CreatorId:   "",
	}

	newChannel, response := c.CreateChannel(channel)
	if response.Error != nil {
		return response.Error
	}

	printer.PrintT("New channel {{.Name}} successfully created", newChannel)

	return nil
}

func removeChannelUsersCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	allUsers, _ := cmd.Flags().GetBool("all-users")

	if allUsers && len(args) != 1 {
		return errors.New("individual users must not be specified in conjunction with the --all-users flag")
	}

	if !allUsers && len(args) < 2 {
		return errors.New("you must specify some users to remove from the channel, or use the --all-users flag to remove them all")
	}

	channel := getChannelFromChannelArg(c, args[0])
	if channel == nil {
		return errors.New("Unable to find channel '" + args[0] + "'")
	}

	if allUsers {
		removeAllUsersFromChannel(c, channel)
	} else {
		users := getUsersFromUserArgs(c, args[1:])
		for i, user := range users {
			removeUserFromChannel(c, channel, user, args[i+1])
		}
	}

	return nil
}

func removeUserFromChannel(c client.Client, channel *model.Channel, user *model.User, userArg string) {
	if user == nil {
		printer.PrintError("Can't find user '" + userArg + "'")
		return
	}
	if _, response := c.RemoveUserFromChannel(channel.Id, user.Id); response.Error != nil {
		printer.PrintError("Unable to remove '" + userArg + "' from " + channel.Name + ". Error: " + response.Error.Error())
	}
}

func removeAllUsersFromChannel(c client.Client, channel *model.Channel) {
	members, response := c.GetChannelMembers(channel.Id, 0, 10000, "")
	if response.Error != nil {
		printer.PrintError("Unable to remove all users from " + channel.Name + ". Error: " + response.Error.Error())
	}

	for _, member := range *members {
		if _, response := c.RemoveUserFromChannel(channel.Id, member.UserId); response.Error != nil {
			printer.PrintError("Unable to remove '" + member.UserId + "' from " + channel.Name + ". Error: " + response.Error.Error())
		}
	}
}

func addChannelUsersCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return errors.New("not enough arguments")
	}

	channel := getChannelFromChannelArg(c, args[0])
	if channel == nil {
		return errors.New("Unable to find channel '" + args[0] + "'")
	}

	users := getUsersFromUserArgs(c, args[1:])
	for i, user := range users {
		addUserToChannel(c, channel, user, args[i+1])
	}

	return nil
}

func addUserToChannel(c client.Client, channel *model.Channel, user *model.User, userArg string) {
	if user == nil {
		printer.PrintError("Can't find user '" + userArg + "'")
		return
	}
	if _, response := c.AddChannelMember(channel.Id, user.Id); response.Error != nil {
		printer.PrintError("Unable to add '" + userArg + "' to " + channel.Name + ". Error: " + response.Error.Error())
	}
}

func archiveChannelsCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("enter at least one channel to archive")
	}

	channels := getChannelsFromChannelArgs(c, args)
	for i, channel := range channels {
		if channel == nil {
			printer.PrintError("Unable to find channel '" + args[i] + "'")
			continue
		}
		if _, response := c.DeleteChannel(channel.Id); response.Error != nil {
			printer.PrintError("Unable to archive channel '" + channel.Name + "' error: " + response.Error.Error())
		}
	}

	return nil
}

func listChannelsCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	teams := getTeamsFromTeamArgs(c, args)
	for i, team := range teams {
		if team == nil {
			printer.PrintError("Unable to find team '" + args[i] + "'")
			continue
		}

		publicChannels, response := c.GetPublicChannelsForTeam(team.Id, 0, 10000, "")
		if response.Error != nil {
			printer.PrintError("Unable to list public channels for '" + args[i] + "'. Error: " + response.Error.Error())
		}
		for _, channel := range publicChannels {
			printer.PrintT("{{.Name}}", channel)
		}

		deletedChannels, response := c.GetDeletedChannelsForTeam(team.Id, 0, 10000, "")
		if response.Error != nil {
			printer.PrintError("Unable to list archived channels for '" + args[i] + "'. Error: " + response.Error.Error())
		}
		for _, channel := range deletedChannels {
			printer.PrintT("{{.Name}} (archived)", channel)
		}
	}

	return nil
}

func restoreChannelsCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("enter at least one channel")
	}

	channels := getChannelsFromChannelArgs(c, args)
	for i, channel := range channels {
		if channel == nil {
			printer.PrintError("Unable to find channel '" + args[i] + "'")
			continue
		}
		if _, response := c.RestoreChannel(channel.Id); response.Error != nil {
			printer.PrintError("Unable to restore channel '" + args[i] + "'. Error: " + response.Error.Error())
		}
	}

	return nil
}

func makeChannelPrivateCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("enter one channel to modify")
	}

	channel := getChannelFromChannelArg(c, args[0])
	if channel == nil {
		return errors.New("unable to find channel '" + args[0] + "'")
	}

	if !(channel.Type == model.CHANNEL_OPEN) {
		return errors.New("you can only change the type of public channels")
	}

	if _, response := c.ConvertChannelToPrivate(channel.Id); response.Error != nil {
		return response.Error
	}

	return nil
}

func modifyChannelCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	public, _ := cmd.Flags().GetBool("public")
	private, _ := cmd.Flags().GetBool("private")

	if public == private {
		return errors.New("you must specify only one of --public or --private")
	}

	channel := getChannelFromChannelArg(c, args[0])
	if channel == nil {
		return errors.New("Unable to find channel '" + args[0] + "'")
	}

	if !(channel.Type == model.CHANNEL_OPEN || channel.Type == model.CHANNEL_PRIVATE) {
		return errors.New("you can only change the type of public/private channels")
	}

	privacy := model.CHANNEL_OPEN
	if private {
		privacy = model.CHANNEL_PRIVATE
	}

	if _, response := c.UpdateChannelPrivacy(channel.Id, privacy); response.Error != nil {
		return errors.Wrapf(response.Error, "Failed to update channel ('%s') privacy", args[0])
	}

	return nil
}

func renameChannelCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	existingTeamChannel := args[0]

	newChannelName, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	newDisplayName, err := cmd.Flags().GetString("display_name")
	if err != nil {
		return err
	}

	// At least one of display name or name flag must be present
	if newDisplayName == "" && newChannelName == "" {
		return errors.New("require at least one flag to rename channel, either 'name' or 'display_name'")
	}

	channel := getChannelFromChannelArg(c, existingTeamChannel)
	if channel == nil {
		return errors.New("unable to find channel from '" + existingTeamChannel + "'")
	}

	channelPatch := &model.ChannelPatch{}
	if newChannelName != "" {
		channelPatch.Name = &newChannelName
	}
	if newDisplayName != "" {
		channelPatch.DisplayName = &newDisplayName
	}

	// Using PatchChannel API to rename channel
	updatedChannel, response := c.PatchChannel(channel.Id, channelPatch)
	if response.Error != nil {
		return errors.New("Cannot rename channel '" + channel.Name + "', error : " + response.Error.Error())
	}

	printer.PrintT("'{{.Name}}' channel renamed", updatedChannel)
	return nil
}

func searchChannelCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	printer.SetSingle(true)

	var channel *model.Channel

	if teamArg, _ := cmd.Flags().GetString("team"); teamArg != "" {
		team := getTeamFromTeamArg(c, teamArg)
		if team == nil {
			return errors.New("Team " + teamArg + " was not found")
		}

		var response *model.Response
		channel, response = c.GetChannelByName(args[0], team.Id, "")
		if response.Error != nil {
			return response.Error
		}
		if channel == nil {
			return errors.New("Channel " + args[0] + " was not found in team " + teamArg)
		}
	} else {
		teams, response := c.GetAllTeams("", 0, 9999)
		if response.Error != nil {
			return response.Error
		}

		for _, team := range teams {
			channel, _ = c.GetChannelByName(args[0], team.Id, "")
			if channel != nil && channel.Name == args[0] {
				break
			}
		}

		if channel == nil {
			return errors.New("Channel " + args[0] + " was not found in any team")
		}
	}

	if channel.DeleteAt > 0 {
		printer.PrintT("Channel Name :{{.Name}}, Display Name :{{.DisplayName}}, Channel ID :{{.Id}} (archived)", channel)
	} else {
		printer.PrintT("Channel Name :{{.Name}}, Display Name :{{.DisplayName}}, Channel ID :{{.Id}}", channel)
	}
	return nil
}
