package commands

import (
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mmctl/printer"
	"github.com/spf13/cobra"
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

func init() {
	LdapCmd.AddCommand(
		LdapSyncCmd,
	)
	RootCmd.AddCommand(LdapCmd)
}

func ldapSyncCmdF(c *model.Client4, cmd *cobra.Command, args []string) error {
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
