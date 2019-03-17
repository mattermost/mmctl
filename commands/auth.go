package commands

import (
	"errors"
	"fmt"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

var AuthCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manages the credentials of the remote Mattermost instances",
}

var LoginCmd = &cobra.Command{
	Use:   "login [server name] [instance url] [username] [password]",
	Short: "Login into an instance",
	Long:  "Login into an instance and store credentials",
	Example: `  auth login local-server https://mattermost.example.com sysadmin mysupersecret
  auth login local-server https://mattermost.example.com sysadmin --password`,
	RunE: loginCmdF,
}

var CurrentCmd = &cobra.Command{
	Use:     "current",
	Short:   "Show current user credentials",
	Long:    "Show the currently stored user credentials",
	Example: `  auth current`,
	RunE:    currentCmdF,
}

var SetCmd = &cobra.Command{
	Use:     "set [server name]",
	Short:   "Set the credentials to use",
	Long:    "Set an credentials to use in the following commands",
	Example: `  auth set local-server`,
	Args:    cobra.ExactArgs(1),
	RunE:    setCmdF,
}

var ListCmd = &cobra.Command{
	Use:     "list",
	Short:   "Lists the credentials",
	Long:    "Print a list of the registered credentials",
	Example: `  auth list`,
	RunE:    listCmdF,
}

var DeleteCmd = &cobra.Command{
	Use:     "delete [server name]",
	Short:   "Delete an credentials",
	Long:    "Delete an credentials by its name",
	Example: `  auth delete local-server`,
	Args:    cobra.ExactArgs(1),
	RunE:    deleteCmdF,
}

var CleanCmd = &cobra.Command{
	Use:     "clean",
	Short:   "Clean all credentials",
	Long:    "Clean the currently stored credentials",
	Example: `  auth clean`,
	RunE:    cleanCmdF,
}

func init() {
	LoginCmd.Flags().Bool("password", false, "asks for the password interactively instead of getting it from the args")
	LoginCmd.Flags().BoolP("active", "a", false, "activates the credentials right after login")

	AuthCmd.AddCommand(
		LoginCmd,
		CurrentCmd,
		SetCmd,
		ListCmd,
		DeleteCmd,
		CleanCmd,
	)

	RootCmd.AddCommand(AuthCmd)
}

func loginCmdF(command *cobra.Command, args []string) error {
	passwordFlag, _ := command.Flags().GetBool("password")
	if passwordFlag && len(args) != 3 {
		return errors.New("name, instance url and username must be specified in conjunction with the --password flag")
	}

	if !passwordFlag && len(args) != 4 {
		return errors.New("Expected four arguments. See help text for details.")
	}

	var password string
	if passwordFlag {
		stdinPassword, err := getPasswordFromStdin()
		if err != nil {
			return errors.New("Couldn't read password. Error: " + err.Error())
		}
		password = stdinPassword
	} else {
		password = args[3]
	}

	c, err := InitClientWithUsernameAndPassword(args[2], password, args[1])
	if err != nil {
		CommandPrintErrorln(err.Error())
		// We don't want usage to be printed as the command was correctly built
		return nil
	}

	credentials := Credentials{
		Name:        args[0],
		InstanceUrl: args[1],
		Username:    args[2],
		AuthToken:   c.AuthToken,
	}

	if err := SaveCredentials(credentials); err != nil {
		return err
	}

	active, _ := command.Flags().GetBool("active")
	if active {
		if err := SetCurrent(args[0]); err != nil {
			return err
		}
	}

	fmt.Printf("\n  credentials for %v: %v@%v stored\n\n", args[0], args[2], args[1])
	return nil
}

func getPasswordFromStdin() (string, error) {
	fmt.Printf("Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println("")
	if err != nil {
		return "", nil
	}
	return string(bytePassword), nil
}

func currentCmdF(command *cobra.Command, args []string) error {
	credentials, err := GetCurrentCredentials()
	if err != nil {
		return err
	}

	fmt.Printf("\n  found credentials for %v: %v @ %v\n\n", credentials.Name, credentials.Username, credentials.InstanceUrl)
	return nil
}

func setCmdF(command *cobra.Command, args []string) error {
	return SetCurrent(args[0])
}

func listCmdF(command *cobra.Command, args []string) error {
	credentialsList, err := ReadCredentialsList()
	if err != nil {
		return err
	}

	if len(*credentialsList) == 0 {
		return errors.New("There are no registered credentials, maybe you need to use login first")
	}

	var maxNameLen, maxUsernameLen, maxInstanceUrlLen int
	for _, c := range *credentialsList {
		if maxNameLen <= len(c.Name) {
			maxNameLen = len(c.Name)
		}
		if maxUsernameLen <= len(c.Username) {
			maxUsernameLen = len(c.Username)
		}
		if maxInstanceUrlLen <= len(c.InstanceUrl) {
			maxInstanceUrlLen = len(c.InstanceUrl)
		}
	}

	fmt.Printf("\n    | Active | %*s | %*s | %*s |\n", maxNameLen, "Name", maxUsernameLen, "Username", maxInstanceUrlLen, "InstanceUrl")
	fmt.Printf("    |%s|%s|%s|%s|\n", strings.Repeat("-", 8), strings.Repeat("-", maxNameLen+2), strings.Repeat("-", maxUsernameLen+2), strings.Repeat("-", maxInstanceUrlLen+2))
	for _, c := range *credentialsList {
		if c.Active {
			fmt.Printf("    |      * | %*s | %*s | %*s |\n", maxNameLen, c.Name, maxUsernameLen, c.Username, maxInstanceUrlLen, c.InstanceUrl)
		} else {
			fmt.Printf("    |        | %*s | %*s | %*s |\n", maxNameLen, c.Name, maxUsernameLen, c.Username, maxInstanceUrlLen, c.InstanceUrl)
		}
	}
	fmt.Println("")
	return nil
}

func deleteCmdF(command *cobra.Command, args []string) error {
	credentialsList, err := ReadCredentialsList()
	if err != nil {
		return err
	}

	name := args[0]
	credentials := (*credentialsList)[name]
	if credentials == nil {
		return errors.New(fmt.Sprintf("Cannot find credentials for server name %v", name))
	}

	delete(*credentialsList, name)
	return SaveCredentialsList(credentialsList)
}

func cleanCmdF(command *cobra.Command, args []string) error {
	if err := CleanCredentials(); err != nil {
		return err
	}
	return nil
}
