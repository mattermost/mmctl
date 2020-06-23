// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"github.com/spf13/cobra"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"
)

var LdapCmd = &cobra.Command{
	Use:   "ldap",
	Short: "LDAP related utilities",
}

var LdapSyncCmd = &cobra.Command{
	Use:     "sync",
	Short:   "Synchronize now",
	Long:    "Synchronize all LDAP users and groups now.",
	Example: "  ldap sync",
	RunE:    withClient(ldapSyncCmdF),
}

var LdapIDMigrate = &cobra.Command{
	Use:     "idmigrate",
	Short:   "Migrate LDAP IdAttribute to new value",
	Long:    "Migrate LDAP IdAttribute to new value. Run this utility then change the IdAttribute to the new value.",
	Example: " ldap idmigrate objectGUID",
	Args:    cobra.ExactArgs(1),
	RunE:    withClient(ldapIDMigrateCmdF),
}

func init() {
	LdapCmd.AddCommand(
		LdapSyncCmd,
		LdapIDMigrate,
	)
	RootCmd.AddCommand(LdapCmd)
}

func ldapSyncCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	printer.SetSingle(true)

	ok, response := c.SyncLdap()
	if response.Error != nil {
		return response.Error
	}

	if ok {
		printer.PrintT("Status: {{.status}}", map[string]interface{}{"status": "ok"})
	} else {
		printer.PrintT("Status: {{.status}}", map[string]interface{}{"status": "error"})
	}

	return nil
}

func ldapIDMigrateCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	toAttribute := args[0]
	ok, response := c.MigrateIdLdap(toAttribute)
	if response.Error != nil {
		return response.Error
	}

	if ok {
		printer.Print("AD/LDAP IdAttribute migration complete. You can now change your IdAttribute to: " + toAttribute)
	}

	return nil
}
