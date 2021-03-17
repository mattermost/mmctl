// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"
)

var SamlCmd = &cobra.Command{
	Use:   "saml",
	Short: "SAML related utilities",
}

var SamlAuthDataReset = &cobra.Command{
	Use:     "authdatamigrate",
	Short:   "Reset AuthData field to Email",
	Long:    "Resets the AuthData field for SAML users to their email. Run this utility after setting the 'id' SAML attribute to an empty value.",
	Example: " saml authdatamigrate",
	Args:    cobra.ExactArgs(1),
	RunE:    withClient(samlAuthDataResetCmdF),
}

func init() {
	SamlAuthDataReset.Flags().Bool("include-deleted", false, "Include deleted users")
	SamlAuthDataReset.Flags().Bool("dry-run", false, "Dry run only")
	SamlAuthDataReset.Flags().StringSlice("users", nil, "Comma-separated list of user IDs to which the operation will be applied")
	SamlCmd.AddCommand(
		SamlAuthDataReset,
	)
	RootCmd.AddCommand(SamlCmd)
}

func samlAuthDataResetCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	includeDeleted, _ := cmd.Flags().GetBool("include-deleted")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	userIDs, _ := cmd.Flags().GetStringSlice("users")

	numAffected, response := c.ResetSamlAuthDataToEmail(includeDeleted, dryRun, userIDs)
	if response.Error != nil {
		return response.Error
	}

	if dryRun {
		printer.Print(fmt.Sprintf("%d user records would be affected.\n", numAffected))
	} else {
		printer.Print(fmt.Sprintf("%d user records were changed.\n", numAffected))
	}

	return nil
}
