package commands

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/mattermost/mattermost-server/model"
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
	name, err := command.Flags().GetString("name")
	if err != nil {
		return err
	}
	username, err := command.Flags().GetString("username")
	if err != nil {
		return err
	}
	password, err := command.Flags().GetString("password")
	if err != nil {
		return err
	}
	accessToken, err := command.Flags().GetString("access-token")
	if err != nil {
		return err
	}
	mfaToken, err := command.Flags().GetString("mfa-token")
	if err != nil {
		return err
	}

	url := args[0]

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
		return errors.New("You must use --access-token or --username, but not both.")
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
			return errors.New("Couldn't read password. Error: " + err.Error())
		}
		password = stdinPassword
	}

	if username != "" {
		var c *model.Client4
		var err error
		if mfaToken != "" {
			c, err = InitClientWithMFA(username, password, mfaToken, url)
		} else {
			c, err = InitClientWithUsernameAndPassword(username, password, url)
		}
		if err != nil {
			CommandPrintErrorln(err.Error())
			// We don't want usage to be printed as the command was correctly built
			return nil
		}
		accessToken = c.AuthToken
	} else {
		username = "Personal Access Token"
		credentials := Credentials{
			InstanceUrl: url,
			AuthToken:   accessToken,
		}
		if _, err := InitClientWithCredentials(&credentials); err != nil {
			CommandPrintErrorln(err.Error())
			// We don't want usage to be printed as the command was correctly built
			return nil
		}
	}

	credentials := Credentials{
		Name:        name,
		InstanceUrl: url,
		Username:    username,
		AuthToken:   accessToken,
	}

	if err := SaveCredentials(credentials); err != nil {
		return err
	}

	noActivate, _ := command.Flags().GetBool("no-activate")
	if !noActivate {
		if err := SetCurrent(name); err != nil {
			return err
		}
	}

	fmt.Printf("\n  credentials for %v: %v@%v stored\n\n", name, username, url)
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
	if err := SetCurrent(args[0]); err != nil {
		return err
	}

	fmt.Printf("Credentials for server \"%v\" set as active\n", args[0])

	return nil
}

func listCmdF(command *cobra.Command, args []string) error {
	credentialsList, err := ReadCredentialsList()
	if err != nil {
		return err
	}

	if len(*credentialsList) == 0 {
		return errors.New("There are no registered credentials, maybe you need to use login first")
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
