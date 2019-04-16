package commands

import (
	"fmt"

	"github.com/mattermost/mattermost-server/model"

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
	RunE: createChannelCmdF,
}

var ChannelRenameCmd = &cobra.Command{
	Use:     "rename",
	Short:   "Rename a channel",
	Long:    `Rename a channel.`,
	Example: `  channel rename myteam:mychannel newchannelname --display_name "New Display Name"`,
	RunE:    renameChannelCmdF,
}

var RemoveChannelUsersCmd = &cobra.Command{
	Use:   "remove [channel] [users]",
	Short: "Remove users from channel",
	Long:  "Remove some users from channel",
	Example: `  channel remove myteam:mychannel user@example.com username
  channel remove myteam:mychannel --all-users`,
	RunE: removeChannelUsersCmdF,
}

var AddChannelUsersCmd = &cobra.Command{
	Use:     "add [channel] [users]",
	Short:   "Add users to channel",
	Long:    "Add some users to channel",
	Example: "  channel add myteam:mychannel user@example.com username",
	RunE:    addChannelUsersCmdF,
}

var ArchiveChannelsCmd = &cobra.Command{
	Use:   "archive [channels]",
	Short: "Archive channels",
	Long: `Archive some channels.
Archive a channel along with all related information including posts from the database.
Channels can be specified by [team]:[channel]. ie. myteam:mychannel or by channel ID.`,
	Example: "  channel archive myteam:mychannel",
	RunE:    archiveChannelsCmdF,
}

var ListChannelsCmd = &cobra.Command{
	Use:   "list [teams]",
	Short: "List all channels on specified teams.",
	Long: `List all channels on specified teams.
Archived channels are appended with ' (archived)'.`,
	Example: "  channel list myteam",
	RunE:    listChannelsCmdF,
}

var RestoreChannelsCmd = &cobra.Command{
	Use:   "restore [channels]",
	Short: "Restore some channels",
	Long: `Restore a previously deleted channel
Channels can be specified by [team]:[channel]. ie. myteam:mychannel or by channel ID.`,
	Example: "  channel restore myteam:mychannel",
	RunE:    restoreChannelsCmdF,
}

var MakeChannelPrivateCmd = &cobra.Command{
	Use:   "make_private [channel]",
	Short: "Set a channel's type to private",
	Long: `Set the type of a channel from public to private.
Channel can be specified by [team]:[channel]. ie. myteam:mychannel or by channel ID.`,
	Example: "  channel make_private myteam:mychannel",
	RunE:    makeChannelPrivateCmdF,
}

var ChannelGroupConstrainedCmd = &cobra.Command{
	Use:   "group-constrained",
	Short: "Manage group-constrained status",
	Long:  "Manage channel group-constrained status and it's associated groups",
}

var ChannelGroupConstrainedEnableCmd = &cobra.Command{
	Use:     "enable [team]:[channel]",
	Short:   "Enables group-constrained restrictions in the specified channel",
	Example: "  channel group-constrained enable myteam:mychannel",
	Args:    cobra.ExactArgs(1),
	RunE:    channelGroupConstrainedEnableCmdF,
}

var ChannelGroupConstrainedDisableCmd = &cobra.Command{
	Use:     "disable [team]:[channel]",
	Short:   "Disables group-constrained restrictions in the specified channel",
	Example: "  channel group-constrained disable myteam:mychannel",
	Args:    cobra.ExactArgs(1),
	RunE:    channelGroupConstrainedDisableCmdF,
}

var ChannelGroupConstrainedStatusCmd = &cobra.Command{
	Use:     "status [team]:[channel]",
	Short:   "Show's the group-constrained status for the specified channel",
	Example: "  channel group-constrained status",
	Args:    cobra.ExactArgs(1),
	RunE:    channelGroupConstrainedStatusCmdF,
}

func init() {
	ChannelCreateCmd.Flags().String("name", "", "Channel Name")
	ChannelCreateCmd.Flags().String("display_name", "", "Channel Display Name")
	ChannelCreateCmd.Flags().String("team", "", "Team name or ID")
	ChannelCreateCmd.Flags().String("header", "", "Channel header")
	ChannelCreateCmd.Flags().String("purpose", "", "Channel purpose")
	ChannelCreateCmd.Flags().Bool("private", false, "Create a private channel.")

	ChannelRenameCmd.Flags().String("display_name", "", "Channel Display Name")

	RemoveChannelUsersCmd.Flags().Bool("all-users", false, "Remove all users from the indicated channel.")

	ChannelGroupConstrainedCmd.AddCommand(
		ChannelGroupConstrainedEnableCmd,
		ChannelGroupConstrainedDisableCmd,
		ChannelGroupConstrainedStatusCmd,
	)

	ChannelCmd.AddCommand(
		ChannelCreateCmd,
		RemoveChannelUsersCmd,
		AddChannelUsersCmd,
		ArchiveChannelsCmd,
		ListChannelsCmd,
		RestoreChannelsCmd,
		MakeChannelPrivateCmd,
		ChannelRenameCmd,
		ChannelGroupConstrainedCmd,
	)

	RootCmd.AddCommand(ChannelCmd)
}

func createChannelCmdF(command *cobra.Command, args []string) error {
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
	teamArg, errteam := command.Flags().GetString("team")
	if errteam != nil || teamArg == "" {
		return errors.New("Team is required")
	}
	header, _ := command.Flags().GetString("header")
	purpose, _ := command.Flags().GetString("purpose")
	useprivate, _ := command.Flags().GetBool("private")

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

	if _, response := c.CreateChannel(channel); response.Error != nil {
		return response.Error
	}

	return nil
}

func removeChannelUsersCmdF(command *cobra.Command, args []string) error {
	c, err := InitClient()
	if err != nil {
		return err
	}

	allUsers, _ := command.Flags().GetBool("all-users")

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

func removeUserFromChannel(c *model.Client4, channel *model.Channel, user *model.User, userArg string) {
	if user == nil {
		CommandPrintErrorln("Can't find user '" + userArg + "'")
		return
	}
	if _, response := c.RemoveUserFromChannel(channel.Id, user.Id); response.Error != nil {
		CommandPrintErrorln("Unable to remove '" + userArg + "' from " + channel.Name + ". Error: " + response.Error.Error())
	}
}

func removeAllUsersFromChannel(c *model.Client4, channel *model.Channel) {
	members, response := c.GetChannelMembers(channel.Id, 0, 10000, "")
	if response.Error != nil {
		CommandPrintErrorln("Unable to remove all users from " + channel.Name + ". Error: " + response.Error.Error())
	}

	for _, member := range *members {
		if _, response := c.RemoveUserFromChannel(channel.Id, member.UserId); response.Error != nil {
			CommandPrintErrorln("Unable to remove '" + member.UserId + "' from " + channel.Name + ". Error: " + response.Error.Error())
		}
	}
}

func addChannelUsersCmdF(command *cobra.Command, args []string) error {
	c, err := InitClient()
	if err != nil {
		return err
	}

	if len(args) < 2 {
		return errors.New("Not enough arguments.")
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

func addUserToChannel(c *model.Client4, channel *model.Channel, user *model.User, userArg string) {
	if user == nil {
		CommandPrintErrorln("Can't find user '" + userArg + "'")
		return
	}
	if _, response := c.AddChannelMember(channel.Id, user.Id); response.Error != nil {
		CommandPrintErrorln("Unable to add '" + userArg + "' from " + channel.Name + ". Error: " + response.Error.Error())
	}
}

func archiveChannelsCmdF(command *cobra.Command, args []string) error {
	c, err := InitClient()
	if err != nil {
		return err
	}

	if len(args) < 1 {
		return errors.New("Enter at least one channel to archive.")
	}

	channels := getChannelsFromChannelArgs(c, args)
	for i, channel := range channels {
		if channel == nil {
			CommandPrintErrorln("Unable to find channel '" + args[i] + "'")
			continue
		}
		if _, response := c.DeleteChannel(channel.Id); response.Error != nil {
			CommandPrintErrorln("Unable to archive channel '" + channel.Name + "' error: " + response.Error.Error())
		}
	}

	return nil
}

func listChannelsCmdF(command *cobra.Command, args []string) error {
	c, err := InitClient()
	if err != nil {
		return err
	}

	if len(args) < 1 {
		return errors.New("Enter at least one team.")
	}

	teams := getTeamsFromTeamArgs(c, args)
	for i, team := range teams {
		if team == nil {
			CommandPrintErrorln("Unable to find team '" + args[i] + "'")
			continue
		}

		publicChannels, response := c.GetPublicChannelsForTeam(team.Id, 0, 10000, "")
		if response.Error != nil {
			CommandPrintErrorln("Unable to list public channels for '" + args[i] + "'. Error: " + response.Error.Error())
		}
		for _, channel := range publicChannels {
			CommandPrettyPrintln(channel.Name)
		}

		deletedChannels, response := c.GetDeletedChannelsForTeam(team.Id, 0, 10000, "")
		if response.Error != nil {
			CommandPrintErrorln("Unable to list archived channels for '" + args[i] + "'. Error: " + response.Error.Error())
		}
		for _, channel := range deletedChannels {
			CommandPrettyPrintln(channel.Name + " (archived)")
		}
	}

	return nil
}

func restoreChannelsCmdF(command *cobra.Command, args []string) error {
	c, err := InitClient()
	if err != nil {
		return err
	}

	if len(args) < 1 {
		return errors.New("Enter at least one channel.")
	}

	channels := getChannelsFromChannelArgs(c, args)
	for i, channel := range channels {
		if channel == nil {
			CommandPrintErrorln("Unable to find channel '" + args[i] + "'")
			continue
		}
		if _, response := c.RestoreChannel(channel.Id); response.Error != nil {
			CommandPrintErrorln("Unable to restore channel '" + args[i] + "'. Error: " + response.Error.Error())
		}
	}

	return nil
}

func makeChannelPrivateCmdF(command *cobra.Command, args []string) error {
	c, err := InitClient()
	if err != nil {
		return err
	}

	if len(args) != 1 {
		return errors.New("Enter one channel to modify.")
	}

	channel := getChannelFromChannelArg(c, args[0])
	if channel == nil {
		return errors.New("Unable to find channel '" + args[0] + "'")
	}

	if !(channel.Type == model.CHANNEL_OPEN) {
		return errors.New("You can only change the type of public channels.")
	}

	if _, response := c.ConvertChannelToPrivate(channel.Id); response.Error != nil {
		return response.Error
	}

	return nil
}

func renameChannelCmdF(command *cobra.Command, args []string) error {
	c, err := InitClient()
	var newDisplayName, newChannelName string
	if err != nil {
		return err
	}

	if len(args) < 2 {
		return errors.New("Not enough arguments.")
	}

	channel := getChannelFromChannelArg(c, args[0])
	if channel == nil {
		return errors.New("Unable to find channel '" + args[0] + "'")
	}

	newChannelName = args[1]
	newDisplayName, errdn := command.Flags().GetString("display_name")
	if errdn != nil {
		return errdn
	}

	channelPatch := model.ChannelPatch{Name: &newChannelName}
	if newDisplayName != "" {
		channelPatch.DisplayName = &newDisplayName
	}

	if _, response := c.PatchChannel(channel.Id, &channelPatch); response.Error != nil {
		return response.Error
	}

	return nil
}

func channelGroupConstrainedEnableCmdF(command *cobra.Command, args []string) error {
	c, err := InitClient()
	if err != nil {
		return err
	}

	channel := getChannelFromChannelArg(c, args[0])
	if channel == nil {
		return errors.New("Unable to find channel '" + args[0] + "'")
	}

	groups, res := c.GetGroupsByChannel(channel.Id, 0, 10)
	if res.Error != nil {
		return res.Error
	}

	if len(groups) == 0 {
		return errors.New("Channel '" + args[0] + "' has no groups associated. It cannot be group-constrained")
	}

	channelPatch := model.ChannelPatch{GroupConstrained: model.NewBool(true)}
	if _, res = c.PatchChannel(channel.Id, &channelPatch); res.Error != nil {
		return res.Error
	}

	return nil
}

func channelGroupConstrainedDisableCmdF(command *cobra.Command, args []string) error {
	c, err := InitClient()
	if err != nil {
		return err
	}

	channel := getChannelFromChannelArg(c, args[0])
	if channel == nil {
		return errors.New("Unable to find channel '" + args[0] + "'")
	}

	channelPatch := model.ChannelPatch{GroupConstrained: model.NewBool(false)}
	if _, res := c.PatchChannel(channel.Id, &channelPatch); res.Error != nil {
		return res.Error
	}

	return nil
}

func channelGroupConstrainedStatusCmdF(command *cobra.Command, args []string) error {
	c, err := InitClient()
	if err != nil {
		return err
	}

	channel := getChannelFromChannelArg(c, args[0])
	if channel == nil {
		return errors.New("Unable to find channel '" + args[0] + "'")
	}

	if channel.GroupConstrained != nil && *channel.GroupConstrained {
		fmt.Println("Enabled")
	} else {
		fmt.Println("Disabled")
	}

	return nil
}
