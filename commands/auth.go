package commands

import (
	"errors"
	"fmt"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

var AuthCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manages the credentials of the remote Mattermost instance",
}

var LoginCmd = &cobra.Command{
	Use:   "login [instance url] [username] [password]",
	Short: "Login into an instance instance",
	Long:  "Login into an instance and store credentials",
	Example: `  auth login https://mattermost.example.com sysadmin mysupersecret
  auth login https://mattermost.example.com sysadmin --password`,
	RunE: loginCmdF,
}

var CurrentCmd = &cobra.Command{
	Use:     "current",
	Short:   "Show current user credentials",
	Long:    "Show the currently stored user credentials",
	Example: `  auth current`,
	RunE:    currentCmdF,
}

var CleanCmd = &cobra.Command{
	Use:     "clean",
	Short:   "Clean current user credentials",
	Long:    "Clean the currently stored user credentials",
	Example: `  auth clean`,
	RunE:    cleanCmdF,
}

func init() {
	LoginCmd.Flags().Bool("password", false, "asks for the password interactively instead of getting it from the args")

	AuthCmd.AddCommand(
		LoginCmd,
		CurrentCmd,
		CleanCmd,
	)

	RootCmd.AddCommand(AuthCmd)
}

func loginCmdF(command *cobra.Command, args []string) error {
	passwordFlag, _ := command.Flags().GetBool("password")
	if passwordFlag && len(args) != 2 {
		return errors.New("instance url and username must be specified in conjunction with the --password flag")
	}

	if !passwordFlag && len(args) != 3 {
		return errors.New("Expected three arguments. See help text for details.")
	}

	var password string
	if passwordFlag {
		stdinPassword, err := getPasswordFromStdin()
		if err != nil {
			return errors.New("Couldn't read password. Error: " + err.Error())
		}
		password = stdinPassword
	} else {
		password = args[2]
	}

	credentials := Credentials{
		InstanceUrl: args[0],
		Username:    args[1],
		Password:    password,
	}

	_, err := InitClientWithCredentials(&credentials)
	if err != nil {
		CommandPrintErrorln(err.Error())
		// We don't want usage to be printed as the command was correctly built
		return nil
	}

	if err := SaveCredentials(credentials); err != nil {
		return err
	}

	fmt.Printf("\n  credentials for %v @ %v stored\n\n", args[1], args[0])
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
	credentials, err := ReadCredentials()
	if err != nil {
		return err
	}

	fmt.Printf("\n  found credentials for %v @ %v\n\n", credentials.Username, credentials.InstanceUrl)
	return nil
}

func cleanCmdF(command *cobra.Command, args []string) error {
	if err := CleanCredentials(); err != nil {
		return err
	}
	return nil
}
