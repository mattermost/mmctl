// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var BotCmd = &cobra.Command{
	Use:   "bot",
	Short: "Management of bots",
}

var CreateBotCmd = &cobra.Command{
	Use:     "create [username]",
	Short:   "Create bot",
	Long:    "Create bot.",
	Example: `  bot create testbot`,
	RunE:    withClient(botCreateCmdF),
}

func init() {
	CreateBotCmd.Flags().String("display-name", "", "Optional. The display name for the new bot.")
	CreateBotCmd.Flags().String("description", "", "Optional. The description text for the new bot.")

	BotCmd.AddCommand(
		CreateBotCmd,
	)

	RootCmd.AddCommand(BotCmd)
}

func botCreateCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("expected at least one argument. See help text for details")
	}

	username := args[0]
	displayName, _ := cmd.Flags().GetString("display-name")
	description, _ := cmd.Flags().GetString("description")

	bot, res := c.CreateBot(&model.Bot{
		Username:    username,
		DisplayName: displayName,
		Description: description,
	})
	if err := res.Error; err != nil {
		return errors.Errorf("could not create bot: %s", err)
	}

	printer.PrintT("Created bot {{.UserId}}", bot)

	return nil
}
