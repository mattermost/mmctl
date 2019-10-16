package commands

import (
	"errors"
	"reflect"
	"strings"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mmctl/printer"

	"github.com/spf13/cobra"
)

var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration",
}

var ConfigGetCmd = &cobra.Command{
	Use:     "get",
	Short:   "Get config setting",
	Long:    "Gets the value of a config setting by its name in dot notation.",
	Example: `config get SqlSettings.DriverName`,
	Args:    cobra.ExactArgs(1),
	RunE:    withClient(configGetCmdF),
}

var ConfigShowCmd = &cobra.Command{
	Use:     "show",
	Short:   "Writes the server configuration to STDOUT",
	Long:    "Prints the server configuration and writes to STDOUT in JSON format.",
	Example: "config show",
	Args:    cobra.NoArgs,
	RunE:    withClient(configShowCmdF),
}

func init() {
	ConfigCmd.AddCommand(
		ConfigGetCmd,
		ConfigShowCmd,
	)
	RootCmd.AddCommand(ConfigCmd)
}

func getValue(path []string, obj interface{}) (interface{}, bool) {
	r := reflect.ValueOf(obj)
	val := r.FieldByName(path[0])

	if val.Kind() == reflect.Invalid {
		return nil, false
	}

	if len(path) == 1 {
		return val.Interface(), true
	} else if val.Kind() == reflect.Struct {
		return getValue(path[1:], val.Interface())
	} else {
		return nil, false
	}
}

func configGetCmdF(c *model.Client4, cmd *cobra.Command, args []string) error {
	printer.SetSingle(true)
	printer.SetFormat(printer.FORMAT_JSON)

	config, response := c.GetConfig()
	if response.Error != nil {
		return response.Error
	}

	path := strings.Split(args[0], ".")
	if val, ok := getValue(path, *config); !ok {
		return errors.New("Invalid key")
	} else {
		printer.Print(val)
	}

	return nil
}

func configShowCmdF(c *model.Client4, cmd *cobra.Command, args []string) error {
	printer.SetSingle(true)
	printer.SetFormat(printer.FORMAT_JSON)
	config, response := c.GetConfig()
	if response.Error != nil {
		return response.Error
	}

	printer.Print(config)

	return nil
}
