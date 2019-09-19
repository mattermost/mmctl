package commands

import (
	"fmt"

	"github.com/mattermost/mattermost-server/model"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var GroupCmd = &cobra.Command{
	Use:   "group",
	Short: "Management of groups",
}

var ListLdapGroupsCmd = &cobra.Command{
	Use:     "list-ldap",
	Short:   "List LDAP groups",
	Example: "  group list-ldap",
	Args:    cobra.NoArgs,
	RunE:    withClient(listLdapGroupsCmdF),
}

var ChannelGroupCmd = &cobra.Command{
	Use:   "channel",
	Short: "Management of channel groups",
}

var ChannelGroupEnableCmd = &cobra.Command{
	Use:     "enable [team]:[channel]",
	Short:   "Enables group constrains in the specified channel",
	Example: "  group channel enable myteam:mychannel",
	Args:    cobra.ExactArgs(1),
	RunE:    withClient(channelGroupEnableCmdF),
}

var ChannelGroupDisableCmd = &cobra.Command{
	Use:     "disable [team]:[channel]",
	Short:   "Disables group constrains in the specified channel",
	Example: "  group channel disable myteam:mychannel",
	Args:    cobra.ExactArgs(1),
	RunE:    withClient(channelGroupDisableCmdF),
}

var ChannelGroupStatusCmd = &cobra.Command{
	Use:     "status [team]:[channel]",
	Short:   "Show's the group constrain status for the specified channel",
	Example: "  group channel status myteam:mychannel",
	Args:    cobra.ExactArgs(1),
	RunE:    withClient(channelGroupStatusCmdF),
}

var ChannelGroupListCmd = &cobra.Command{
	Use:     "list [team]:[channel]",
	Short:   "List channel groups",
	Long:    "List the groups associated with a channel",
	Example: "  group channel list myteam:mychannel",
	Args:    cobra.ExactArgs(1),
	RunE:    withClient(channelGroupListCmdF),
}

var TeamGroupCmd = &cobra.Command{
	Use:   "team",
	Short: "Management of team groups",
}

var TeamGroupEnableCmd = &cobra.Command{
	Use:     "enable [team]",
	Short:   "Enables group constrains in the specified team",
	Example: "  group team enable myteam",
	Args:    cobra.ExactArgs(1),
	RunE:    withClient(teamGroupEnableCmdF),
}

var TeamGroupDisableCmd = &cobra.Command{
	Use:     "disable [team]",
	Short:   "Disables group constrains in the specified team",
	Example: "  group team disable myteam",
	Args:    cobra.ExactArgs(1),
	RunE:    withClient(teamGroupDisableCmdF),
}

var TeamGroupStatusCmd = &cobra.Command{
	Use:     "status [team]",
	Short:   "Show's the group constrain status for the specified team",
	Example: "  group team status myteam",
	Args:    cobra.ExactArgs(1),
	RunE:    withClient(teamGroupStatusCmdF),
}

var TeamGroupListCmd = &cobra.Command{
	Use:     "list [team]",
	Short:   "List team groups",
	Long:    "List the groups associated with a team",
	Example: "  group team list myteam",
	Args:    cobra.ExactArgs(1),
	RunE:    withClient(teamGroupListCmdF),
}

func init() {
	ChannelGroupCmd.AddCommand(
		ChannelGroupEnableCmd,
		ChannelGroupDisableCmd,
		ChannelGroupStatusCmd,
		ChannelGroupListCmd,
	)

	TeamGroupCmd.AddCommand(
		TeamGroupEnableCmd,
		TeamGroupDisableCmd,
		TeamGroupStatusCmd,
		TeamGroupListCmd,
	)

	GroupCmd.AddCommand(
		ListLdapGroupsCmd,
		ChannelGroupCmd,
		TeamGroupCmd,
	)

	RootCmd.AddCommand(GroupCmd)
}

func listLdapGroupsCmdF(c *model.Client4, cmd *cobra.Command, args []string) error {
	groups, res := c.GetLdapGroups()
	if res.Error != nil {
		return res.Error
	}

	for _, group := range groups {
		fmt.Println(group)
	}

	return nil
}

func channelGroupEnableCmdF(c *model.Client4, cmd *cobra.Command, args []string) error {
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

func channelGroupDisableCmdF(c *model.Client4, cmd *cobra.Command, args []string) error {
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

func channelGroupStatusCmdF(c *model.Client4, cmd *cobra.Command, args []string) error {
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

func channelGroupListCmdF(c *model.Client4, cmd *cobra.Command, args []string) error {
	channel := getChannelFromChannelArg(c, args[0])
	if channel == nil {
		return errors.New("Unable to find channel '" + args[0] + "'")
	}

	groups, res := c.GetGroupsByChannel(channel.Id, 0, 9999)
	if res.Error != nil {
		return res.Error
	}

	for _, group := range groups {
		fmt.Println(group.DisplayName)
	}

	return nil
}

func teamGroupEnableCmdF(c *model.Client4, cmd *cobra.Command, args []string) error {
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

func teamGroupDisableCmdF(c *model.Client4, cmd *cobra.Command, args []string) error {
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

func teamGroupStatusCmdF(c *model.Client4, cmd *cobra.Command, args []string) error {
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

func teamGroupListCmdF(c *model.Client4, cmd *cobra.Command, args []string) error {
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
