// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mmctl/printer"
)

var AuthCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manages the credentials of the remote Mattermost instances",
}

var LoginCmd = &cobra.Command{
	Use:   "login [instance url] --name [server name] --username [username] --password [password]",
	Short: "Login into an instance",
	Long:  "Login into an instance and store credentials",
	Example: `  auth login https://mattermost.example.com
  auth login https://mattermost.example.com --name local-server --username sysadmin --password mysupersecret
  auth login https://mattermost.example.com --name local-server --username sysadmin --password mysupersecret --mfa-token 123456
  auth login https://mattermost.example.com --name local-server --access-token myaccesstoken`,
	Args: cobra.ExactArgs(1),
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

var RenewCmd = &cobra.Command{
	Use:     "renew",
	Short:   "Renews a set of credentials",
	Long:    "Renews the credentials for a given server",
	Example: `  auth renew local-server`,
	Args:    cobra.ExactArgs(1),
	RunE:    renewCmdF,
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
	LoginCmd.Flags().StringP("name", "n", "", "Name for the credentials")
	LoginCmd.Flags().StringP("username", "u", "", "Username for the credentials")
	LoginCmd.Flags().StringP("access-token", "a", "", "Access token to use instead of username/password")
	LoginCmd.Flags().StringP("mfa-token", "m", "", "MFA token for the credentials")
	LoginCmd.Flags().StringP("password", "p", "", "Password for the credentials")
	LoginCmd.Flags().Bool("no-activate", false, "If present, it won't activate the credentials after login")

	RenewCmd.Flags().StringP("password", "p", "", "Password for the credentials")
	RenewCmd.Flags().StringP("access-token", "a", "", "Access token to use instead of username/password")
	RenewCmd.Flags().StringP("mfa-token", "m", "", "MFA token for the credentials")

	AuthCmd.AddCommand(
		LoginCmd,
		CurrentCmd,
		SetCmd,
		ListCmd,
		RenewCmd,
		DeleteCmd,
		CleanCmd,
	)

	RootCmd.AddCommand(AuthCmd)
}

func loginCmdF(cmd *cobra.Command, args []string) error {
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}
	username, err := cmd.Flags().GetString("username")
	if err != nil {
		return err
	}
	password, err := cmd.Flags().GetString("password")
	if err != nil {
		return err
	}
	accessToken, err := cmd.Flags().GetString("access-token")
	if err != nil {
		return err
	}
	mfaToken, err := cmd.Flags().GetString("mfa-token")
	if err != nil {
		return err
	}

	allowInsecureSHA1 := viper.GetBool("insecure-sha1-intermediate")
	allowInsecureTLS := viper.GetBool("insecure-tls-version")

	url := strings.TrimRight(args[0], "/")
	method := MethodPassword

	if name == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("Connection name: ")
		name, err = reader.ReadString('\n')
		if err != nil {
			return err
		}
		name = strings.TrimSpace(name)
	}

	if accessToken != "" && username != "" {
		return errors.New("you must use --access-token or --username, but not both")
	}

	if accessToken == "" && username == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("Username: ")
		username, err = reader.ReadString('\n')
		if err != nil {
			return err
		}
		username = strings.TrimSpace(username)
	}

	if username != "" && password == "" {
		fmt.Printf("Password: ")
		stdinPassword, err := getPasswordFromStdin()
		if err != nil {
			return errors.WithMessage(err, "couldn't read password")
		}
		password = stdinPassword
	}

	if username != "" {
		var c *model.Client4
		var err error
		if mfaToken != "" {
			c, _, err = InitClientWithMFA(username, password, mfaToken, url, allowInsecureSHA1, allowInsecureTLS)
			method = MethodMFA
		} else {
			c, _, err = InitClientWithUsernameAndPassword(username, password, url, allowInsecureSHA1, allowInsecureTLS)
		}
		if err != nil {
			printer.PrintError(err.Error())
			// We don't want usage to be printed as the command was correctly built
			return nil
		}
		accessToken = c.AuthToken
	} else {
		username = "Personal Access Token"
		method = MethodToken
		credentials := Credentials{
			InstanceURL: url,
			AuthToken:   accessToken,
		}
		if _, _, err := InitClientWithCredentials(&credentials, allowInsecureSHA1, allowInsecureTLS); err != nil {
			printer.PrintError(err.Error())
			// We don't want usage to be printed as the command was correctly built
			return nil
		}
	}

	credentials := Credentials{
		Name:        name,
		InstanceURL: url,
		Username:    username,
		AuthToken:   accessToken,
		AuthMethod:  method,
	}

	if err := SaveCredentials(credentials); err != nil {
		return err
	}

	noActivate, _ := cmd.Flags().GetBool("no-activate")
	if !noActivate {
		if err := SetCurrent(name); err != nil {
			return err
		}
	}

	fmt.Printf("\n  credentials for %q: \"%s@%s\" stored\n\n", name, username, url)
	return nil
}

func getPasswordFromStdin() (string, error) {
	// syscall.Stdin is of type int in all architectures but in
	// windows, so we have to cast it to ensure cross compatibility
	//nolint:unconvert
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println("")
	if err != nil {
		return "", err
	}
	return string(bytePassword), nil
}

func currentCmdF(cmd *cobra.Command, args []string) error {
	credentials, err := GetCurrentCredentials()
	if err != nil {
		return err
	}

	fmt.Printf("\n  found credentials for %q: \"%s@%s\"\n\n", credentials.Name, credentials.Username, credentials.InstanceURL)
	return nil
}

func setCmdF(cmd *cobra.Command, args []string) error {
	if err := SetCurrent(args[0]); err != nil {
		return err
	}

	fmt.Printf("Credentials for server %q set as active\n", args[0])

	return nil
}

func listCmdF(cmd *cobra.Command, args []string) error {
	credentialsList, err := ReadCredentialsList()
	if err != nil {
		return err
	}

	if len(*credentialsList) == 0 {
		return errors.New("there are no registered credentials, maybe you need to use login first")
	}

	serverNames := []string{}
	var maxNameLen, maxUsernameLen, maxInstanceURLLen int
	for _, c := range *credentialsList {
		serverNames = append(serverNames, c.Name)
		if maxNameLen <= len(c.Name) {
			maxNameLen = len(c.Name)
		}
		if maxUsernameLen <= len(c.Username) {
			maxUsernameLen = len(c.Username)
		}
		if maxInstanceURLLen <= len(c.InstanceURL) {
			maxInstanceURLLen = len(c.InstanceURL)
		}
	}
	sort.Slice(serverNames, func(i, j int) bool {
		return serverNames[i] < serverNames[j]
	})

	fmt.Printf("\n    | Active | %*s | %*s | %*s |\n", maxNameLen, "Name", maxUsernameLen, "Username", maxInstanceURLLen, "InstanceURL")
	fmt.Printf("    |%s|%s|%s|%s|\n", strings.Repeat("-", 8), strings.Repeat("-", maxNameLen+2), strings.Repeat("-", maxUsernameLen+2), strings.Repeat("-", maxInstanceURLLen+2))
	for _, name := range serverNames {
		c := (*credentialsList)[name]
		if c.Active {
			fmt.Printf("    |      * | %*s | %*s | %*s |\n", maxNameLen, c.Name, maxUsernameLen, c.Username, maxInstanceURLLen, c.InstanceURL)
		} else {
			fmt.Printf("    |        | %*s | %*s | %*s |\n", maxNameLen, c.Name, maxUsernameLen, c.Username, maxInstanceURLLen, c.InstanceURL)
		}
	}
	fmt.Println("")
	return nil
}

func renewCmdF(cmd *cobra.Command, args []string) error {
	printer.SetSingle(true)
	password, _ := cmd.Flags().GetString("password")
	accessToken, _ := cmd.Flags().GetString("access-token")
	mfaToken, _ := cmd.Flags().GetString("mfa-token")
	allowInsecureSHA1 := viper.GetBool("insecure-sha1-intermediate")
	allowInsecureTLS := viper.GetBool("insecure-tls-version")

	credentials, err := GetCredentials(args[0])
	if err != nil {
		return err
	}

	if (credentials.AuthMethod == MethodPassword || credentials.AuthMethod == MethodMFA) && password == "" {
		if password == "" {
			fmt.Printf("Password: ")
			stdinPassword, err := getPasswordFromStdin()
			if err != nil {
				return errors.WithMessage(err, "couldn't read password")
			}
			password = stdinPassword
		}
	}

	switch credentials.AuthMethod {
	case MethodPassword:
		c, _, err := InitClientWithUsernameAndPassword(credentials.Username, password, credentials.InstanceURL, allowInsecureSHA1, allowInsecureTLS)
		if err != nil {
			return err
		}

		credentials.AuthToken = c.AuthToken

	case MethodToken:
		if accessToken == "" {
			return errors.New("requires the --access-token parameter to be set")
		}

		credentials.AuthToken = accessToken
		if _, _, err := InitClientWithCredentials(credentials, allowInsecureSHA1, allowInsecureTLS); err != nil {
			return err
		}

	case MethodMFA:
		if mfaToken == "" {
			return errors.New("requires the --mfa-token parameter to be set")
		}

		c, _, err := InitClientWithMFA(credentials.Username, password, mfaToken, credentials.InstanceURL, allowInsecureSHA1, allowInsecureTLS)
		if err != nil {
			return err
		}
		credentials.AuthToken = c.AuthToken

	default:
		return errors.Errorf("invalid auth method %q", credentials.AuthMethod)
	}

	if err := SaveCredentials(*credentials); err != nil {
		return err
	}

	printer.PrintT("Credentials for server \"{{.Name}}\" successfully renewed", credentials)

	return nil
}

func deleteCmdF(cmd *cobra.Command, args []string) error {
	credentialsList, err := ReadCredentialsList()
	if err != nil {
		return err
	}

	name := args[0]
	credentials := (*credentialsList)[name]
	if credentials == nil {
		return errors.Errorf("cannot find credentials for server name %q", name)
	}

	delete(*credentialsList, name)
	return SaveCredentialsList(credentialsList)
}

func cleanCmdF(cmd *cobra.Command, args []string) error {
	if err := CleanCredentials(); err != nil {
		return err
	}
	return nil
}
