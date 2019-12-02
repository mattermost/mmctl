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
	"golang.org/x/crypto/ssh/terminal"

	"github.com/mattermost/mattermost-server/model"
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

	url := args[0]
	method := METHOD_PASSWORD

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
			c, err = InitClientWithMFA(username, password, mfaToken, url)
			method = METHOD_MFA
		} else {
			c, err = InitClientWithUsernameAndPassword(username, password, url)
		}
		if err != nil {
			printer.PrintError(err.Error())
			// We don't want usage to be printed as the command was correctly built
			return nil
		}
		accessToken = c.AuthToken
	} else {
		username = "Personal Access Token"
		method = METHOD_TOKEN
		credentials := Credentials{
			InstanceUrl: url,
			AuthToken:   accessToken,
		}
		if _, err := InitClientWithCredentials(&credentials); err != nil {
			printer.PrintError(err.Error())
			// We don't want usage to be printed as the command was correctly built
			return nil
		}
	}

	credentials := Credentials{
		Name:        name,
		InstanceUrl: url,
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

	fmt.Printf("\n  credentials for %s: %s@%s stored\n\n", name, username, url)
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

func currentCmdF(cmd *cobra.Command, args []string) error {
	credentials, err := GetCurrentCredentials()
	if err != nil {
		return err
	}

	fmt.Printf("\n  found credentials for %s: %s @ %s\n\n", credentials.Name, credentials.Username, credentials.InstanceUrl)
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
	var maxNameLen, maxUsernameLen, maxInstanceUrlLen int
	for _, c := range *credentialsList {
		serverNames = append(serverNames, c.Name)
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
	sort.Slice(serverNames, func(i, j int) bool {
		return serverNames[i] < serverNames[j]
	})

	fmt.Printf("\n    | Active | %*s | %*s | %*s |\n", maxNameLen, "Name", maxUsernameLen, "Username", maxInstanceUrlLen, "InstanceUrl")
	fmt.Printf("    |%s|%s|%s|%s|\n", strings.Repeat("-", 8), strings.Repeat("-", maxNameLen+2), strings.Repeat("-", maxUsernameLen+2), strings.Repeat("-", maxInstanceUrlLen+2))
	for _, name := range serverNames {
		c := (*credentialsList)[name]
		if c.Active {
			fmt.Printf("    |      * | %*s | %*s | %*s |\n", maxNameLen, c.Name, maxUsernameLen, c.Username, maxInstanceUrlLen, c.InstanceUrl)
		} else {
			fmt.Printf("    |        | %*s | %*s | %*s |\n", maxNameLen, c.Name, maxUsernameLen, c.Username, maxInstanceUrlLen, c.InstanceUrl)
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

	credentials, err := GetCredentials(args[0])
	if err != nil {
		return err
	}

	if (credentials.AuthMethod == METHOD_PASSWORD || credentials.AuthMethod == METHOD_MFA) && password == "" {
		if password == "" {
			stdinPassword, err := getPasswordFromStdin()
			if err != nil {
				return errors.WithMessage(err, "couldn't read password")
			}
			password = stdinPassword
		}
	}

	switch credentials.AuthMethod {
	case METHOD_PASSWORD:
		c, err := InitClientWithUsernameAndPassword(credentials.Username, password, credentials.InstanceUrl)
		if err != nil {
			return err
		}

		credentials.AuthToken = c.AuthToken

	case METHOD_TOKEN:
		if accessToken == "" {
			return errors.New("requires the --access-token parameter to be set")
		}

		credentials.AuthToken = accessToken
		if _, err := InitClientWithCredentials(credentials); err != nil {
			return err
		}

	case METHOD_MFA:
		if mfaToken == "" {
			return errors.New("requires the --mfa-token parameter to be set")
		}

		c, err := InitClientWithMFA(credentials.Username, password, mfaToken, credentials.InstanceUrl)
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
