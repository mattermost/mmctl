package commands

import (
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mmctl/client"
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

var ConfigSetCmd = &cobra.Command{
	Use:     "set",
	Short:   "Set config setting",
	Long:    "Adds or updates a config setting by its name in dot notation.",
	Example: `config set SqlSettings.DriverName postgresql`,
	Args:    cobra.ExactArgs(2),
	RunE:    withClient(configSetCmdF),
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
		ConfigSetCmd,
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

func setValue(path []string, obj reflect.Value, newValue string) error {
	val := obj.FieldByName(path[0])

	if val.Kind() == reflect.Invalid {
		return errors.New("Selected path object is not valid")
	}

	if val.Kind() == reflect.Struct {
		return setValue(path[1:], val, newValue)
	} else if len(path) == 1 {
		// All the updatable values should be pointers so can be modified
		// using reflection
		if val.Kind() != reflect.Ptr && val.Kind() != reflect.Slice {
			return errors.New("Value is not modifiable")
		} else if val.Kind() == reflect.Ptr {
			switch val.Elem().Kind() {
			case reflect.Int:
				v, err := strconv.ParseInt(newValue, 10, 64)
				if err != nil {
					return errors.New("Target value is of type Int and provided value is not")
				}
				val.Elem().SetInt(v)
				return nil
			case reflect.String:
				val.Elem().SetString(newValue)
				return nil
			case reflect.Bool:
				v, err := strconv.ParseBool(newValue)
				if err != nil {
					return errors.New("Target value is of type Bool and provided value is not")
				}
				val.Elem().SetBool(v)
				return nil
			default:
				return errors.New("Target value type is not supported")
			}
		} else {
			var newSlice []string
			err := json.Unmarshal([]byte(newValue), &newSlice)
			if err != nil {
				return errors.New("Target value is of type array of strings and provided value is not")
			}
			val.Set(reflect.ValueOf(newSlice))
			return nil
		}
	} else {
		return errors.New("Path object type is not supported")
	}

	return nil
}

func setConfigValue(path []string, config *model.Config, newValue string) error {
	return setValue(path, reflect.ValueOf(config).Elem(), newValue)
}

func configGetCmdF(c client.Client, cmd *cobra.Command, args []string) error {
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

func configSetCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	config, response := c.GetConfig()
	if response.Error != nil {
		return response.Error
	}

	path, err := parseConfigPath(args[0])
	if err != nil {
		return err
	}
	if err := setConfigValue(path, config, args[1]); err != nil {
		return err
	}
	if res := c.SetConfig(config); res.Error != nil {
		return res.Error
	}

	return nil
}

func configShowCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	printer.SetSingle(true)
	printer.SetFormat(printer.FORMAT_JSON)
	config, response := c.GetConfig()
	if response.Error != nil {
		return response.Error
	}

	printer.Print(config)

	return nil
}

func parseConfigPath(configPath string) ([]string, error) {
	return strings.Split(configPath, "."), nil
}
