// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var TokenCmd = &cobra.Command{
	Use:   "token",
	Short: "manage users' access tokens",
}

var GenerateUserTokenCmd = &cobra.Command{
	Use:     "generate [user] [description]",
	Short:   "Generate token for a user",
	Long:    "Generate token for a user",
	Example: "  generate-token testuser test-token",
	RunE:    withClient(generateTokenForAUserCmdF),
	Args:    cobra.ExactArgs(2),
}

var ListUserTokensCmd = &cobra.Command{
	Use:     "list [user]",
	Short:   "List users tokens",
	Long:    "List the tokens of a user",
	Example: "  user tokens testuser",
	RunE:    withClient(listTokensOfAUserCmdF),
	Args:    cobra.ExactArgs(1),
}

func init() {
	ListUserTokensCmd.Flags().Int("page", 0, "Page number to fetch for the list of users")
	ListUserTokensCmd.Flags().Int("per-page", 200, "Number of users to be fetched")
	ListUserTokensCmd.Flags().Bool("all", false, "Fetch all tokens. --page flag will be ignore if provided")
	ListUserTokensCmd.Flags().Bool("active", true, "List only active tokens")
	ListUserTokensCmd.Flags().Bool("inactive", false, "List only inactive tokens")

	TokenCmd.AddCommand(
		GenerateUserTokenCmd,
		ListUserTokensCmd,
	)

	RootCmd.AddCommand(
		TokenCmd,
	)
}

func generateTokenForAUserCmdF(c client.Client, command *cobra.Command, args []string) error {
	userArg := args[0]
	user := getUserFromUserArg(c, userArg)
	if user == nil {
		return errors.Errorf("could not retrieve user information of %q", userArg)
	}

	token, res := c.CreateUserAccessToken(user.Id, args[1])
	if res.Error != nil {
		return errors.Errorf("could not create token for %q: %s", userArg, res.Error.Error())
	}
	printer.PrintT("{{.Token}}: {{.Description}}", token)

	return nil
}

func listTokensOfAUserCmdF(c client.Client, command *cobra.Command, args []string) error {
	page, _ := command.Flags().GetInt("page")
	perPage, _ := command.Flags().GetInt("per-page")
	showAll, _ := command.Flags().GetBool("all")
	active, _ := command.Flags().GetBool("active")
	inactive, _ := command.Flags().GetBool("inactive")

	if showAll {
		page = 0
	}

	userArg := args[0]

	user := getUserFromUserArg(c, userArg)
	if user == nil {
		return errors.Errorf("could not retrieve user information of %q", userArg)
	}

	tokens, res := c.GetUserAccessTokensForUser(user.Id, page, perPage)
	if res.Error != nil {
		return errors.Errorf("could not retrieve tokens for user %q: %s", userArg, res.Error.Error())
	}
	if len(tokens) == 0 {
		return errors.Errorf("there are no tokens for the %q", userArg)
	}

	for _, t := range tokens {
		if t.IsActive && !inactive {
			printer.PrintT("{{.Id}}: {{.Description}}", t)
		}
		if !t.IsActive && !active {
			printer.PrintT("{{.Id}}: {{.Description}}", t)
		}
	}
	return nil
}
